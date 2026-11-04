package updater

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"discord-quest-helper/internal/config"
)

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"v1.0.0", "v1.0.1", true},
		{"v1.0.1", "v1.1.0", true},
		{"v1.1.0", "v2.0.0", true},
		{"v1.1.0", "v1.1.0", false},
		{"v1.1.0", "v1.0.1", false},
		{"v1.1.0", "v1.0.0", false},
		{"1.1.0", "1.1.1", true},
		{"v1.1", "v1.1.1", true},
		{"v1.1.0", "", false},
	}

	for _, c := range cases {
		actual := IsNewerVersion(c.current, c.latest)
		if actual != c.expected {
			t.Errorf("IsNewerVersion(%q, %q) = %v, expected %v", c.current, c.latest, actual, c.expected)
		}
	}
}

func TestCheckUpdateWithMock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tag_name": "v9.9.9",
			"html_url": "https://github.com/veykax/discord-quest-helper/releases/tag/v9.9.9"
		}`))
	}))
	defer ts.Close()

	oldURL := apiURL
	apiURL = ts.URL
	defer func() {
		apiURL = oldURL
		cachedInfo = nil
		lastCheck = time.Time{}
	}()

	cachedInfo = nil
	lastCheck = time.Time{}

	info := CheckUpdate()
	if !info.HasUpdate {
		t.Errorf("expected HasUpdate to be true for v9.9.9 vs %s", config.AppVersion)
	}
	if info.LatestVersion != "v9.9.9" {
		t.Errorf("expected LatestVersion to be v9.9.9, got %s", info.LatestVersion)
	}
	if info.ReleaseURL != "https://github.com/veykax/discord-quest-helper/releases/tag/v9.9.9" {
		t.Errorf("expected ReleaseURL to match mock release URL")
	}
}
