package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"
)

type GameItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	ExeName      string `json:"exe"`
	WindowTitle  string `json:"windowTitle"`
	SteamFolder  string `json:"steamFolder,omitempty"`
	SteamAppID   string `json:"steamAppId,omitempty"`
	DiscordAppID string `json:"discordAppId,omitempty"`
}

const (
	AppVersion           = "v1.0.5"
	GitHubRepo           = "veykax/discord-quest-helper"
	QuestDurationSeconds = 15*60 + 15
	DefaultPort          = 39482
	RemoteGamesURL       = "https://raw.githubusercontent.com/veykax/discord-quest-helper/main/games.json"
)

var (
	mu                sync.RWMutex
	gamesList         []GameItem
	SupportedGames    []GameItem // Backwards compatibility for external callers/tests
	remoteURL         = RemoteGamesURL
	cachePathOverride string
	failedResolves    = make(map[string]time.Time)
	failedResolvesMu  sync.Mutex
)

func init() {
	ensureLoaded()
}

// SetCachePathOverride allows tests or custom environments to isolate the games cache.
func SetCachePathOverride(path string) {
	mu.Lock()
	defer mu.Unlock()
	cachePathOverride = path
}

func getLocalCachePath() string {
	if cachePathOverride != "" {
		return cachePathOverride
	}
	if custom := os.Getenv("DISCORD_GAMES_CACHE"); custom != "" {
		return custom
	}
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, "DiscordQuestHelper", "games_cache.json")
		}
	case "linux":
		if home != "" {
			return filepath.Join(home, ".local", "share", "discord-quest-helper", "games_cache.json")
		}
	case "darwin":
		if home != "" {
			return filepath.Join(home, "Library", "Application Support", "DiscordQuestHelper", "games_cache.json")
		}
	}
	return filepath.Join(os.TempDir(), "DiscordQuestHelper", "games_cache.json")
}

func ensureLoaded() {
	mu.Lock()
	defer mu.Unlock()

	if len(gamesList) > 0 {
		return
	}

	// 1. Load embedded default games
	var defaults []GameItem
	if len(defaultGamesJSON) > 0 {
		_ = json.Unmarshal(defaultGamesJSON, &defaults)
	}

	// 2. Load cached games from disk if present
	cachePath := getLocalCachePath()
	var cached []GameItem
	if data, err := os.ReadFile(cachePath); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &cached)
	}

	// Merge defaults and cached
	gamesList = mergeGameLists(defaults, cached)
	SupportedGames = append([]GameItem(nil), gamesList...)
}

func mergeGameLists(lists ...[]GameItem) []GameItem {
	seenID := make(map[string]int)
	seenAppID := make(map[string]int)
	var result []GameItem

	for _, list := range lists {
		for _, g := range list {
			if g.ID == "" {
				g.ID = slugify(g.Title)
			}
			if idx, ok := seenID[g.ID]; ok {
				result[idx] = g
				if g.DiscordAppID != "" {
					seenAppID[g.DiscordAppID] = idx
				}
				continue
			}
			if g.DiscordAppID != "" {
				if idx, ok := seenAppID[g.DiscordAppID]; ok {
					result[idx] = g
					continue
				}
			}

			idx := len(result)
			result = append(result, g)
			seenID[g.ID] = idx
			if g.DiscordAppID != "" {
				seenAppID[g.DiscordAppID] = idx
			}
		}
	}
	return result
}

// GetGames returns a thread-safe copy of all registered games.
func GetGames() []GameItem {
	ensureLoaded()
	mu.RLock()
	defer mu.RUnlock()
	res := make([]GameItem, len(gamesList))
	copy(res, gamesList)
	return res
}

// AddOrUpdateGame registers or updates a game and persists to local cache.
func AddOrUpdateGame(g GameItem) {
	mu.Lock()
	defer mu.Unlock()

	gamesList = mergeGameLists(gamesList, []GameItem{g})
	SupportedGames = append([]GameItem(nil), gamesList...)
	saveLocalCacheLocked()
}

func saveLocalCacheLocked() {
	cachePath := getLocalCachePath()
	_ = os.MkdirAll(filepath.Dir(cachePath), 0755)
	if data, err := json.MarshalIndent(gamesList, "", "  "); err == nil {
		_ = os.WriteFile(cachePath, data, 0644)
	}
}

// FetchRemoteGames downloads the latest games list from GitHub, merges, and caches it.
func FetchRemoteGames() ([]GameItem, error) {
	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequest(http.MethodGet, remoteURL, nil)
	if err != nil {
		return GetGames(), err
	}
	req.Header.Set("User-Agent", "DiscordQuestHelper")

	resp, err := client.Do(req)
	if err != nil {
		return GetGames(), err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GetGames(), fmt.Errorf("remote games returned HTTP %d", resp.StatusCode)
	}

	var remote []GameItem
	if err := json.NewDecoder(resp.Body).Decode(&remote); err != nil {
		return GetGames(), fmt.Errorf("decode remote games: %w", err)
	}

	mu.Lock()
	gamesList = mergeGameLists(gamesList, remote)
	SupportedGames = append([]GameItem(nil), gamesList...)
	saveLocalCacheLocked()
	mu.Unlock()

	return GetGames(), nil
}

type discordRPCResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Executables    []struct {
		OS         string `json:"os"`
		Name       string `json:"name"`
		IsLauncher bool   `json:"is_launcher"`
	} `json:"executables"`
	ThirdPartySKUs []struct {
		Distributor string `json:"distributor"`
		ID          string `json:"id"`
	} `json:"third_party_skus"`
}

// ResolveDiscordApp queries Discord's public RPC endpoint by App ID, extracts the executable,
// creates a GameItem, and adds it to the active games registry.
func ResolveDiscordApp(appID string, questTitle string) (*GameItem, error) {
	if appID == "" || appID == "545364944258990091" { // Skip Discord internal app ID
		return nil, fmt.Errorf("invalid app id")
	}

	// Negative cache check to prevent repeated failing requests
	failedResolvesMu.Lock()
	if failTime, ok := failedResolves[appID]; ok {
		if time.Since(failTime) < 15*time.Minute {
			failedResolvesMu.Unlock()
			return nil, fmt.Errorf("app id %s recently failed to resolve (cooldown)", appID)
		}
		delete(failedResolves, appID)
	}
	failedResolvesMu.Unlock()

	recordFailure := func() {
		failedResolvesMu.Lock()
		failedResolves[appID] = time.Now()
		failedResolvesMu.Unlock()
	}

	// 1. Check if already known
	for _, g := range GetGames() {
		if g.DiscordAppID == appID {
			return &g, nil
		}
	}

	// 2. Fetch application info from Discord RPC endpoint
	url := fmt.Sprintf("https://discord.com/api/v9/applications/%s/rpc", appID)
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		recordFailure()
		return nil, err
	}
	req.Header.Set("User-Agent", "DiscordQuestHelper/1.1")

	resp, err := client.Do(req)
	if err != nil {
		recordFailure()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		recordFailure()
		return nil, fmt.Errorf("discord rpc returned HTTP %d", resp.StatusCode)
	}

	var rpcResp discordRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		recordFailure()
		return nil, err
	}

	title := rpcResp.Name
	if title == "" {
		title = questTitle
	}
	if title == "" {
		title = "Game " + appID
	}

	// Select best executable
	exeName := ""
	for _, e := range rpcResp.Executables {
		if e.OS == "win32" && !e.IsLauncher && strings.HasSuffix(strings.ToLower(e.Name), ".exe") {
			exeName = e.Name
			break
		}
	}
	if exeName == "" {
		for _, e := range rpcResp.Executables {
			if strings.HasSuffix(strings.ToLower(e.Name), ".exe") {
				exeName = e.Name
				break
			}
		}
	}
	if exeName == "" {
		for _, e := range rpcResp.Executables {
			if e.OS == "win32" && e.Name != "" {
				exeName = e.Name
				break
			}
		}
	}
	if exeName == "" {
		// Fallback: derive executable from title
		exeName = slugify(title) + ".exe"
	}

	// Check Steam App ID
	steamAppID := ""
	for _, sku := range rpcResp.ThirdPartySKUs {
		if sku.Distributor == "steam" && sku.ID != "" {
			steamAppID = sku.ID
			break
		}
	}

	itemID := slugify(title)
	if itemID == "game" && appID != "" {
		itemID = "game_" + appID
	}

	item := GameItem{
		ID:           itemID,
		Title:        title,
		ExeName:      exeName,
		WindowTitle:  title,
		SteamAppID:   steamAppID,
		DiscordAppID: appID,
	}

	AddOrUpdateGame(item)
	return &item, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteRune('_')
		}
	}
	res := strings.Trim(b.String(), "_")
	for strings.Contains(res, "__") {
		res = strings.ReplaceAll(res, "__", "_")
	}
	if res == "" {
		res = "game"
	}
	return res
}
