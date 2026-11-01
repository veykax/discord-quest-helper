package quests

import (
	"testing"

	"discord-quest-helper/internal/config"
)

func TestReadQuestsLiveCache(t *testing.T) {
	resp := ReadQuests(config.SupportedGames)
	if !resp.FoundCache {
		t.Skip("Discord cache not found on host, skipping live cache test")
	}

	t.Logf("Found cache at %s (Last updated: %s)", resp.CachePath, resp.LastUpdated)
	t.Logf("Detected %d enrolled quests", len(resp.Quests))

	for _, q := range resp.Quests {
		t.Logf("- Quest: %q (AppID: %s, AppName: %q) -> Matched: %q | Progress: %d/%ds | Completed: %v, Claimed: %v",
			q.QuestName, q.AppID, q.AppName, q.MatchedGameID, q.ProgressSeconds, q.TargetSeconds, q.Completed, q.Claimed)
	}
}

func TestNormalizeText(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"EA SPORTS FC™ 27", "ea sports fc 27"},
		{"Call of Duty: Modern Warfare 4", "call of duty modern warfare 4"},
		{"Once Human Free Expansion", "once human free expansion"},
		{"   Resident   Evil: Requiem!  ", "resident evil requiem"},
	}

	for _, c := range cases {
		actual := normalizeText(c.input)
		if actual != c.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", c.input, actual, c.expected)
		}
	}
}
