package config

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestGetGames(t *testing.T) {
	games := GetGames()
	if len(games) == 0 {
		t.Fatalf("expected non-empty games list from embedded default_games.json")
	}

	foundRoblox := false
	foundMaple := false
	for _, g := range games {
		if g.ID == "roblox" {
			foundRoblox = true
		}
		if g.ID == "maplestory_cw" {
			foundMaple = true
		}
	}

	if !foundRoblox {
		t.Errorf("expected to find roblox in games list")
	}
	if !foundMaple {
		t.Errorf("expected to find maplestory_cw in games list")
	}
}

func TestResolveDiscordApp(t *testing.T) {
	// Test resolving MapleStory Classic World App ID
	item, err := ResolveDiscordApp("1565580212664795249", "MapleStory Classic World")
	if err != nil {
		t.Fatalf("ResolveDiscordApp failed: %v", err)
	}

	if item == nil {
		t.Fatalf("expected resolved item, got nil")
	}

	if item.DiscordAppID != "1565580212664795249" {
		t.Errorf("expected DiscordAppID to be 1565580212664795249, got %s", item.DiscordAppID)
	}

	if item.ExeName == "" {
		t.Errorf("expected ExeName to be resolved")
	}

	t.Logf("Resolved game: %+v", item)
}

func TestFetchRemoteGames(t *testing.T) {
	// Isolate cache path to temp directory so tests never pollute real host cache
	tempCache := filepath.Join(t.TempDir(), "test_games_cache.json")
	SetCachePathOverride(tempCache)
	defer SetCachePathOverride("")

	// Mock remote games.json endpoint
	mockPayload := `[
		{
			"id": "mock_game_test",
			"title": "Mock Remote Game",
			"exe": "mockgame.exe",
			"windowTitle": "Mock Remote Game",
			"discordAppId": "999999999999"
		}
	]`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockPayload))
	}))
	defer ts.Close()

	oldURL := remoteURL
	remoteURL = ts.URL
	defer func() { remoteURL = oldURL }()

	updated, err := FetchRemoteGames()
	if err != nil {
		t.Fatalf("FetchRemoteGames failed: %v", err)
	}

	found := false
	for _, g := range updated {
		if g.ID == "mock_game_test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find mock_game_test in updated games list")
	}
}

func TestSlugify(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Roblox Player Beta", "roblox_player_beta"},
		{"Call of Duty: Modern Warfare 4", "call_of_duty_modern_warfare_4"},
		{"Мир Танков", "мир_танков"},
		{"---Special__Game---", "special_game"},
		{"", "game"},
	}

	for _, c := range cases {
		actual := slugify(c.input)
		if actual != c.expected {
			t.Errorf("slugify(%q) = %q, expected %q", c.input, actual, c.expected)
		}
	}
}
