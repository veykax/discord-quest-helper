package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"discord-quest-helper/internal/config"
)

type UpdateInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	ReleaseURL     string `json:"releaseUrl"`
}

var (
	mu         sync.RWMutex
	cachedInfo *UpdateInfo
	lastCheck  time.Time
	apiURL     = "https://api.github.com/repos/" + config.GitHubRepo + "/releases/latest"
)

// CheckUpdate checks GitHub Releases for a newer version of the application.
func CheckUpdate() UpdateInfo {
	mu.RLock()
	if cachedInfo != nil && time.Since(lastCheck) < 15*time.Minute {
		info := *cachedInfo
		mu.RUnlock()
		return info
	}
	mu.RUnlock()

	info := UpdateInfo{
		CurrentVersion: config.AppVersion,
		LatestVersion:  config.AppVersion,
		HasUpdate:      false,
		ReleaseURL:     "https://github.com/" + config.GitHubRepo + "/releases/latest",
	}

	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return info
	}
	req.Header.Set("User-Agent", "DiscordQuestHelper/"+config.AppVersion)

	resp, err := client.Do(req)
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return info
	}

	var ghRelease struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghRelease); err != nil {
		return info
	}

	latest := strings.TrimSpace(ghRelease.TagName)
	if latest != "" {
		info.LatestVersion = latest
		if ghRelease.HTMLURL != "" {
			info.ReleaseURL = ghRelease.HTMLURL
		}
		if IsNewerVersion(config.AppVersion, latest) {
			info.HasUpdate = true
		}
	}

	mu.Lock()
	cachedInfo = &info
	lastCheck = time.Now()
	mu.Unlock()

	return info
}

// IsNewerVersion compares semver version strings (e.g. "v1.1.0" vs "v1.2.0").
func IsNewerVersion(current, latest string) bool {
	c := strings.TrimPrefix(strings.TrimSpace(current), "v")
	l := strings.TrimPrefix(strings.TrimSpace(latest), "v")
	if c == l || l == "" {
		return false
	}

	cParts := strings.Split(c, ".")
	lParts := strings.Split(l, ".")

	for i := 0; i < len(cParts) && i < len(lParts); i++ {
		var cNum, lNum int
		fmt.Sscanf(cParts[i], "%d", &cNum)
		fmt.Sscanf(lParts[i], "%d", &lNum)
		if lNum > cNum {
			return true
		} else if lNum < cNum {
			return false
		}
	}
	return len(lParts) > len(cParts)
}
