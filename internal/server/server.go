package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"discord-quest-helper/internal/config"
	"discord-quest-helper/internal/quests"
	"discord-quest-helper/internal/spoofer"
)

type AppServer struct {
	spoofer          spoofer.Manager
	webFS            fs.FS
	mu               sync.Mutex
	isRunning        bool
	currentGame      config.GameItem
	remainingSeconds int
	totalSeconds     int
	exitChan         chan struct{}
}

func New(sp spoofer.Manager, webFS fs.FS, exitChan chan struct{}) *AppServer {
	return &AppServer{
		spoofer:          sp,
		webFS:            webFS,
		totalSeconds:     config.QuestDurationSeconds,
		remainingSeconds: config.QuestDurationSeconds,
		exitChan:         exitChan,
	}
}

func (s *AppServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/games", s.handleGames)
	mux.HandleFunc("/api/quests", s.handleQuests)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/start", s.handleStart)
	mux.HandleFunc("/api/stop", s.handleStop)
	mux.HandleFunc("/api/exit", s.handleExit)
	mux.HandleFunc("/api/open-discord", s.handleOpenDiscord)

	fileServer := http.FileServer(http.FS(s.webFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" || strings.HasSuffix(r.URL.Path, ".png") || r.URL.Path == "/manifest.json" {
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		}
		fileServer.ServeHTTP(w, r)
	})
}

func (s *AppServer) handleGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config.SupportedGames)
}

func (s *AppServer) handleQuests(w http.ResponseWriter, r *http.Request) {
	resp := quests.ReadQuests(config.SupportedGames)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type StatusResponse struct {
	Running   bool            `json:"running"`
	Game      config.GameItem `json:"game"`
	Remaining int             `json:"remaining"`
	Total     int             `json:"total"`
}

func (s *AppServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp := StatusResponse{
		Running:   s.isRunning,
		Game:      s.currentGame,
		Remaining: s.remainingSeconds,
		Total:     s.totalSeconds,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type StartRequest struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	ExeName         string      `json:"exe"`
	WindowTitle     string      `json:"windowTitle"`
	SteamFolder     string      `json:"steamFolder"`
	SteamAppID      interface{} `json:"steamAppId"`
	DiscordAppID    interface{} `json:"discordAppId"`
	DurationSeconds int         `json:"durationSeconds,omitempty"`
}

func (s *AppServer) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wTitle := req.WindowTitle
	if wTitle == "" {
		wTitle = req.Title
	}

	steamAppID := ""
	if req.SteamAppID != nil {
		switch v := req.SteamAppID.(type) {
		case string:
			steamAppID = v
		case float64:
			if v > 0 {
				steamAppID = fmt.Sprintf("%.0f", v)
			}
		}
	}

	discordAppID := ""
	if req.DiscordAppID != nil {
		switch v := req.DiscordAppID.(type) {
		case string:
			discordAppID = v
		case float64:
			if v > 0 {
				discordAppID = fmt.Sprintf("%.0f", v)
			}
		}
	}

	if err := s.spoofer.Start(req.ExeName, wTitle, req.SteamFolder, steamAppID, discordAppID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	duration := config.QuestDurationSeconds
	if req.DurationSeconds > 0 {
		duration = req.DurationSeconds
	}

	s.mu.Lock()
	s.isRunning = true
	s.currentGame = config.GameItem{
		ID:           req.ID,
		Title:        req.Title,
		ExeName:      req.ExeName,
		WindowTitle:  wTitle,
		SteamFolder:  req.SteamFolder,
		SteamAppID:   steamAppID,
		DiscordAppID: discordAppID,
	}
	s.totalSeconds = duration
	s.remainingSeconds = duration
	s.mu.Unlock()

	go s.runTimer()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *AppServer) handleStop(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.isRunning = false
	s.spoofer.Stop()
	s.remainingSeconds = s.totalSeconds
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *AppServer) handleExit(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.isRunning = false
	s.spoofer.Stop()
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	close(s.exitChan)
}

func (s *AppServer) handleOpenDiscord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uri := "discord://-/quest-home"

	switch runtime.GOOS {
	case "windows":
		// rundll32 не создает консольного окна cmd и передает url зарегистрированному обработчику
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", uri).Start()
	case "linux":
		// На Linux gio open обращается напрямую к зарегистрированным desktop-файлам (Flatpak / native)
		// без бага KDE/xdg-open с открытием пустого окна браузера
		opened := false
		if _, err := exec.LookPath("gio"); err == nil {
			if err := exec.Command("gio", "open", uri).Start(); err == nil {
				opened = true
			}
		}
		if !opened {
			if _, err := exec.LookPath("flatpak"); err == nil {
				if err := exec.Command("flatpak", "run", "com.discordapp.Discord", uri).Start(); err == nil {
					opened = true
				}
			}
		}
		if !opened {
			_ = exec.Command("xdg-open", uri).Start()
		}
	case "darwin":
		_ = exec.Command("open", uri).Start()
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *AppServer) runTimer() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		if !s.isRunning {
			s.mu.Unlock()
			return
		}

		// Если окно симуляции игры было закрыто пользователем — останавливаем квест
		if !s.spoofer.IsRunning() {
			s.isRunning = false
			s.mu.Unlock()
			return
		}

		if s.remainingSeconds > 0 {
			s.remainingSeconds--
			s.mu.Unlock()
		} else {
			s.isRunning = false
			s.spoofer.Stop()
			s.mu.Unlock()
			return
		}
	}
}
