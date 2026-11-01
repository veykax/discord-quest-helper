package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"discord-quest-helper/internal/config"
)

type mockSpoofer struct {
	running     bool
	currentGame string
}

func (m *mockSpoofer) Start(exeName, windowTitle, steamFolder, steamAppID, discordAppID string) error {
	m.running = true
	m.currentGame = exeName
	return nil
}

func (m *mockSpoofer) Stop() error {
	m.running = false
	m.currentGame = ""
	return nil
}

func (m *mockSpoofer) IsRunning() bool {
	return m.running
}

func (m *mockSpoofer) CurrentGame() string {
	return m.currentGame
}

func TestServerRoutes(t *testing.T) {
	mockSp := &mockSpoofer{}
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>hello</html>")},
	}
	exitChan := make(chan struct{})

	srv := New(mockSp, mockFS, exitChan)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	// 1. GET /api/games
	req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var games []config.GameItem
	if err := json.NewDecoder(rec.Body).Decode(&games); err != nil {
		t.Fatalf("failed to decode games: %v", err)
	}
	if len(games) != len(config.SupportedGames) {
		t.Fatalf("expected %d games, got %d", len(config.SupportedGames), len(games))
	}

	// 2. GET /api/status (idle)
	req = httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var st StatusResponse
	_ = json.NewDecoder(rec.Body).Decode(&st)
	if st.Running {
		t.Fatalf("expected running to be false")
	}

	// 3. POST /api/start
	startBody, _ := json.Marshal(StartRequest{
		ID:           "roblox",
		Title:        "Roblox",
		ExeName:      "RobloxPlayerBeta.exe",
		WindowTitle:  "Roblox",
		DiscordAppID: "363445589247131668",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/start", bytes.NewReader(startBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on start, got %d", rec.Code)
	}
	if !mockSp.IsRunning() {
		t.Fatalf("expected mock spoofer to be running")
	}

	// 4. POST /api/stop
	req = httptest.NewRequest(http.MethodPost, "/api/stop", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on stop, got %d", rec.Code)
	}
	if mockSp.IsRunning() {
		t.Fatalf("expected mock spoofer to be stopped")
	}

	// 5. GET /api/quests
	req = httptest.NewRequest(http.MethodGet, "/api/quests", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on /api/quests, got %d", rec.Code)
	}
}
