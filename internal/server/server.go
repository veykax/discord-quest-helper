package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"discord-quest-helper/internal/config"
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
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/start", s.handleStart)
	mux.HandleFunc("/api/stop", s.handleStop)
	mux.HandleFunc("/api/exit", s.handleExit)
	mux.HandleFunc("/api/open-discord", s.handleOpenDiscord)

	fileServer := http.FileServer(http.FS(s.webFS))
	mux.Handle("/", fileServer)
}

func (s *AppServer) handleGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config.SupportedGames)
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
	ID          string `json:"id"`
	Title       string `json:"title"`
	ExeName     string `json:"exe"`
	WindowTitle string `json:"windowTitle"`
	SteamFolder string `json:"steamFolder"`
	SteamAppID  string `json:"steamAppId"`
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

	if err := s.spoofer.Start(req.ExeName, wTitle, req.SteamFolder, req.SteamAppID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.isRunning = true
	s.currentGame = config.GameItem{
		ID:          req.ID,
		Title:       req.Title,
		ExeName:     req.ExeName,
		WindowTitle: wTitle,
	}
	s.remainingSeconds = s.totalSeconds
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

	switch runtime.GOOS {
	case "windows":
		// rundll32 не создает консольного окна cmd
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", "discord://").Start()
	case "linux":
		_ = exec.Command("xdg-open", "discord://").Start()
	case "darwin":
		_ = exec.Command("open", "discord://").Start()
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
