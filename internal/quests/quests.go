package quests

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode"

	"discord-quest-helper/internal/config"
)

// QuestStatus represents a Discord quest detected from the local Discord cache.
type QuestStatus struct {
	ID               string `json:"id"`
	QuestName        string `json:"questName"`
	GameTitle        string `json:"gameTitle"`
	AppID            string `json:"appId"`
	AppName          string `json:"appName"`
	TargetSeconds    int    `json:"targetSeconds"`
	ProgressSeconds  int    `json:"progressSeconds"`
	RemainingSeconds int    `json:"remainingSeconds"`
	Enrolled         bool   `json:"enrolled"`
	Completed        bool   `json:"completed"`
	Claimed          bool   `json:"claimed"`
	MatchedGameID    string `json:"matchedGameId,omitempty"`
}

// QuestsResponse is the JSON response returned by the /api/quests endpoint.
type QuestsResponse struct {
	Quests      []QuestStatus `json:"quests"`
	FoundCache  bool          `json:"foundCache"`
	CachePath   string        `json:"cachePath,omitempty"`
	LastUpdated string        `json:"lastUpdated,omitempty"`
	ModUnix     int64         `json:"modUnix,omitempty"`
	Error       string        `json:"error,omitempty"`
}

// Internal JSON structure for Discord /api/v9/quests/@me response stored in Chromium cache.
type rawDiscordCache struct {
	Quests []struct {
		ID     string `json:"id"`
		Config struct {
			ExpiresAt   string `json:"expires_at"`
			Application struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"application"`
			Messages struct {
				QuestName string `json:"quest_name"`
				GameTitle string `json:"game_title"`
			} `json:"messages"`
			TaskConfigV2 struct {
				Tasks map[string]struct {
					Type         string `json:"type"`
					Target       int    `json:"target"`
					Applications []struct {
						ID string `json:"id"`
					} `json:"applications"`
				} `json:"tasks"`
			} `json:"task_config_v2"`
			TaskConfig struct {
				Tasks map[string]struct {
					Type         string `json:"type"`
					Target       int    `json:"target"`
					Applications []struct {
						ID string `json:"id"`
					} `json:"applications"`
				} `json:"tasks"`
			} `json:"task_config"`
		} `json:"config"`
		UserStatus *struct {
			EnrolledAt  *string `json:"enrolled_at"`
			CompletedAt *string `json:"completed_at"`
			ClaimedAt   *string `json:"claimed_at"`
			Progress    map[string]struct {
				Value float64 `json:"value"`
			} `json:"progress"`
		} `json:"user_status"`
	} `json:"quests"`
}

const (
	// Deterministic Chromium Simple Cache key for "1/0/https://discord.com/api/v9/quests/@me"
	cacheFileName = "6a03c8f27701ed48_0"
)

// candidateDirs returns a list of candidate cache directories based on OS and common Discord installs.
func candidateDirs() []string {
	var dirs []string

	// Custom environment variable override if provided
	if custom := os.Getenv("DISCORD_CACHE_DIR"); custom != "" {
		dirs = append(dirs, custom)
	}

	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		localAppData := os.Getenv("LOCALAPPDATA")
		if appData == "" && home != "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		if localAppData == "" && home != "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		discordFlavors := []string{"discord", "discordcanary", "discordptb", "discorddevelopment"}

		if appData != "" {
			for _, flavor := range discordFlavors {
				dirs = append(dirs, filepath.Join(appData, flavor, "Cache", "Cache_Data"))
			}
		}
		if localAppData != "" {
			for _, flavor := range []string{"Discord", "DiscordCanary", "DiscordPTB"} {
				dirs = append(dirs, filepath.Join(localAppData, flavor, "Cache", "Cache_Data"))
			}
		}

	case "linux":
		if home != "" {
			// Flatpak Discord locations
			flatpaks := []struct {
				appID  string
				config string
			}{
				{"com.discordapp.Discord", "discord"},
				{"com.discordapp.DiscordCanary", "discordcanary"},
				{"com.discordapp.DiscordPTB", "discordptb"},
			}
			for _, fp := range flatpaks {
				dirs = append(dirs, filepath.Join(home, ".var", "app", fp.appID, "config", fp.config, "Cache", "Cache_Data"))
			}

			// Native Linux ~/.config locations
			flavors := []string{"discord", "discordcanary", "discordptb"}
			for _, flavor := range flavors {
				dirs = append(dirs, filepath.Join(home, ".config", flavor, "Cache", "Cache_Data"))
			}

			// Snap Discord
			dirs = append(dirs, filepath.Join(home, "snap", "discord", "current", ".config", "discord", "Cache", "Cache_Data"))
		}

		// Also check WSL Windows mounts if running inside WSL environment
		if _, err := os.Stat("/mnt/c/Users"); err == nil {
			if wslMatches, err := filepath.Glob("/mnt/c/Users/*/AppData/Roaming/discord/Cache/Cache_Data"); err == nil {
				dirs = append(dirs, wslMatches...)
			}
			if wslMatches, err := filepath.Glob("/mnt/c/Users/*/AppData/Roaming/discordcanary/Cache/Cache_Data"); err == nil {
				dirs = append(dirs, wslMatches...)
			}
			if wslMatches, err := filepath.Glob("/mnt/c/Users/*/AppData/Roaming/discordptb/Cache/Cache_Data"); err == nil {
				dirs = append(dirs, wslMatches...)
			}
		}

	case "darwin":
		if home != "" {
			dirs = append(dirs, filepath.Join(home, "Library", "Application Support", "discord", "Cache", "Cache_Data"))
			dirs = append(dirs, filepath.Join(home, "Library", "Application Support", "discordcanary", "Cache", "Cache_Data"))
		}
	}

	return dirs
}

// isQuestsCacheFile checks if a file is a valid Discord quests response payload
// (either Chromium Simple Cache member or Blockfile external stream f_XXXXXX).
func isQuestsCacheFile(filePath string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 32768)
	n, err := f.Read(buf)
	if n < 10 {
		return false
	}
	buf = buf[:n]

	// Check if uncompressed JSON
	if bytes.Contains(buf, []byte("\"quests\"")) {
		return true
	}

	// Check for gzip stream header
	gzipHeader := []byte{0x1f, 0x8b, 0x08}
	offset := bytes.Index(buf, gzipHeader)
	if offset == -1 {
		return false
	}

	zr, err := gzip.NewReader(bytes.NewReader(buf[offset:]))
	if err != nil {
		return false
	}
	defer zr.Close()
	zr.Multistream(false)

	dec := make([]byte, 2048)
	dn, _ := zr.Read(dec)
	return bytes.Contains(dec[:dn], []byte("\"quests\""))
}

// FindFreshestCacheFile locates the most recently updated quests cache file on disk.
func FindFreshestCacheFile() (string, time.Time, error) {
	if customFile := os.Getenv("DISCORD_CACHE_FILE"); customFile != "" {
		if fi, err := os.Stat(customFile); err == nil && !fi.IsDir() {
			return customFile, fi.ModTime(), nil
		}
	}

	var bestPath string
	var bestMod time.Time

	dirs := candidateDirs()
	for _, dir := range dirs {
		// 1. Direct match: Simple Cache (Linux Flatpak / native ~/.config / macOS)
		p := filepath.Join(dir, cacheFileName)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			if fi.ModTime().After(bestMod) {
				bestMod = fi.ModTime()
				bestPath = p
			}
			continue
		}

		// 2. Scan for Blockfile Cache external files (f_XXXXXX), standard on Windows Electron
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		type fileEntry struct {
			path    string
			modTime time.Time
		}
		var fFiles []fileEntry
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasPrefix(entry.Name(), "f_") {
				continue
			}
			if fi, err := entry.Info(); err == nil {
				fFiles = append(fFiles, fileEntry{
					path:    filepath.Join(dir, entry.Name()),
					modTime: fi.ModTime(),
				})
			}
		}

		// Sort newest first: the first valid quest file found is guaranteed to be the freshest in this dir
		sort.Slice(fFiles, func(i, j int) bool {
			return fFiles[i].modTime.After(fFiles[j].modTime)
		})

		for _, item := range fFiles {
			if bestPath != "" && !item.modTime.After(bestMod) {
				break
			}
			if isQuestsCacheFile(item.path) {
				bestMod = item.modTime
				bestPath = item.path
				break
			}
		}
	}

	if bestPath != "" {
		return bestPath, bestMod, nil
	}

	return "", time.Time{}, fmt.Errorf("no discord quests cache found in candidate paths")
}

// ReadQuests parses the Discord quests cache and matches quests against SupportedGames.
func ReadQuests(supportedGames []config.GameItem) QuestsResponse {
	filePath, modTime, err := FindFreshestCacheFile()
	if err != nil {
		return QuestsResponse{
			Quests:     []QuestStatus{},
			FoundCache: false,
			Error:      err.Error(),
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return QuestsResponse{
			Quests:     []QuestStatus{},
			FoundCache: true,
			CachePath:  filePath,
			Error:      fmt.Sprintf("read cache file: %v", err),
		}
	}

	var raw rawDiscordCache

	// Chromium Simple Cache Stream 1 starts with gzip header: 0x1F, 0x8B, 0x08
	gzipHeader := []byte{0x1f, 0x8b, 0x08}
	offset := bytes.Index(data, gzipHeader)
	if offset != -1 {
		zr, err := gzip.NewReader(bytes.NewReader(data[offset:]))
		if err != nil {
			return QuestsResponse{
				Quests:     []QuestStatus{},
				FoundCache: true,
				CachePath:  filePath,
				Error:      fmt.Sprintf("create gzip reader: %v", err),
			}
		}
		defer zr.Close()

		// Prevent reading beyond the single gzip member into Chromium cache trailer
		zr.Multistream(false)

		decompressed, err := io.ReadAll(zr)
		if err != nil {
			return QuestsResponse{
				Quests:     []QuestStatus{},
				FoundCache: true,
				CachePath:  filePath,
				Error:      fmt.Sprintf("decompress stream: %v", err),
			}
		}

		if err := json.Unmarshal(decompressed, &raw); err != nil {
			return QuestsResponse{
				Quests:     []QuestStatus{},
				FoundCache: true,
				CachePath:  filePath,
				Error:      fmt.Sprintf("unmarshal json: %v", err),
			}
		}
	} else if jsonStart := bytes.IndexByte(data, '{'); jsonStart != -1 && bytes.Contains(data, []byte("\"quests\"")) {
		// Fallback for uncompressed JSON payload in cache
		unmarshalErr := json.Unmarshal(data[jsonStart:], &raw)
		if unmarshalErr != nil {
			if lastBrace := bytes.LastIndexByte(data, '}'); lastBrace > jsonStart {
				unmarshalErr = json.Unmarshal(data[jsonStart:lastBrace+1], &raw)
			}
		}
		if unmarshalErr != nil {
			return QuestsResponse{
				Quests:     []QuestStatus{},
				FoundCache: true,
				CachePath:  filePath,
				Error:      fmt.Sprintf("unmarshal uncompressed json: %v", unmarshalErr),
			}
		}
	} else {
		return QuestsResponse{
			Quests:     []QuestStatus{},
			FoundCache: true,
			CachePath:  filePath,
			Error:      "neither gzip header nor quests JSON found in cache entry",
		}
	}

	var result []QuestStatus
	for _, q := range raw.Quests {
		if q.UserStatus == nil || q.UserStatus.EnrolledAt == nil {
			// Skip quests the user hasn't enrolled in
			continue
		}

		// Skip expired quests that were not completed
		if q.Config.ExpiresAt != "" {
			var expTime time.Time
			var parseErr error
			expTime, parseErr = time.Parse(time.RFC3339Nano, q.Config.ExpiresAt)
			if parseErr != nil {
				expTime, parseErr = time.Parse(time.RFC3339, q.Config.ExpiresAt)
			}
			if parseErr == nil && expTime.Before(time.Now()) && q.UserStatus.CompletedAt == nil {
				continue
			}
		}

		target := 900
		var taskAppIDs []string
		isDesktopPlay := false

		if t, ok := q.Config.TaskConfigV2.Tasks["PLAY_ON_DESKTOP"]; ok {
			isDesktopPlay = true
			if t.Target > 0 {
				target = t.Target
			}
			for _, app := range t.Applications {
				if app.ID != "" {
					taskAppIDs = append(taskAppIDs, app.ID)
				}
			}
		} else if t, ok := q.Config.TaskConfig.Tasks["PLAY_ON_DESKTOP"]; ok {
			isDesktopPlay = true
			if t.Target > 0 {
				target = t.Target
			}
			for _, app := range t.Applications {
				if app.ID != "" {
					taskAppIDs = append(taskAppIDs, app.ID)
				}
			}
		}

		if !isDesktopPlay {
			// Skip video, mobile, and activity quests — only desktop game quests can be spoofed
			continue
		}

		progress := 0
		if p, ok := q.UserStatus.Progress["PLAY_ON_DESKTOP"]; ok {
			progress = int(p.Value)
		}

		remaining := target - progress
		if remaining < 0 {
			remaining = 0
		}

		qName := q.Config.Messages.QuestName
		if qName == "" {
			qName = q.Config.Messages.GameTitle
		}

		gTitle := q.Config.Messages.GameTitle
		if gTitle == "" {
			gTitle = qName
		}

		appID := q.Config.Application.ID
		appName := q.Config.Application.Name

		qs := QuestStatus{
			ID:               q.ID,
			QuestName:        qName,
			GameTitle:        gTitle,
			AppID:            appID,
			AppName:          appName,
			TargetSeconds:    target,
			ProgressSeconds:  progress,
			RemainingSeconds: remaining,
			Enrolled:         q.UserStatus.EnrolledAt != nil,
			Completed:        q.UserStatus.CompletedAt != nil,
			Claimed:          q.UserStatus.ClaimedAt != nil,
		}

		// Match quest to supported game
		qs.MatchedGameID = matchGameID(qs, taskAppIDs, supportedGames)
		if qs.MatchedGameID == "" {
			// Auto-resolver for desktop play quests
			candidateAppIDs := append([]string(nil), taskAppIDs...)
			if qs.AppID != "" && qs.AppID != "545364944258990091" {
				candidateAppIDs = append(candidateAppIDs, qs.AppID)
			}
			for _, tid := range candidateAppIDs {
				if item, err := config.ResolveDiscordApp(tid, qs.GameTitle); err == nil && item != nil {
					qs.MatchedGameID = item.ID
					break
				}
			}
		}
		result = append(result, qs)
	}

	return QuestsResponse{
		Quests:      result,
		FoundCache:  true,
		CachePath:   filePath,
		LastUpdated: modTime.UTC().Format(time.RFC3339),
		ModUnix:     modTime.Unix(),
	}
}

// normalizeText cleans a string to lowercase alphanumeric words for fuzzy matching.
func normalizeText(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// matchGameID finds the corresponding GameItem.ID in SupportedGames for this quest.
func matchGameID(qs QuestStatus, taskAppIDs []string, games []config.GameItem) string {
	// 1. Direct AppID match
	for _, g := range games {
		if g.DiscordAppID != "" {
			if qs.AppID != "" && g.DiscordAppID == qs.AppID {
				return g.ID
			}
			for _, tid := range taskAppIDs {
				if g.DiscordAppID == tid {
					return g.ID
				}
			}
		}
	}

	// 2. Normalized Title / AppName matching
	normQuestName := normalizeText(qs.QuestName)
	normGameTitle := normalizeText(qs.GameTitle)
	normAppName := normalizeText(qs.AppName)

	for _, g := range games {
		normGTitle := normalizeText(g.Title)
		normGWin := normalizeText(g.WindowTitle)
		normGID := normalizeText(strings.ReplaceAll(g.ID, "_", " "))

		// Check if normalized title is contained in quest name or app name
		if normGTitle != "" {
			if strings.Contains(normQuestName, normGTitle) ||
				strings.Contains(normGameTitle, normGTitle) ||
				strings.Contains(normAppName, normGTitle) ||
				strings.Contains(normGTitle, normAppName) {
				return g.ID
			}
		}

		if normGWin != "" && (strings.Contains(normQuestName, normGWin) || strings.Contains(normAppName, normGWin)) {
			return g.ID
		}

		if normGID != "" && (strings.Contains(normQuestName, normGID) || strings.Contains(normAppName, normGID)) {
			return g.ID
		}
	}

	return ""
}
