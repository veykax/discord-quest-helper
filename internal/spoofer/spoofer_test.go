package spoofer

import (
	"testing"
	"time"
)

func TestSpooferLifecycle(t *testing.T) {
	sp := New()
	if sp.IsRunning() {
		t.Fatalf("expected spoofer to be idle initially")
	}

	// Start Roblox (using real AppID)
	err := sp.Start("RobloxPlayerBeta.exe", "Roblox", "", "", "363445589247131668")
	if err != nil {
		t.Fatalf("spoofer Start failed: %v", err)
	}

	if !sp.IsRunning() {
		t.Fatalf("expected spoofer to be running")
	}

	if sp.CurrentGame() != "RobloxPlayerBeta.exe" {
		t.Fatalf("expected current game to be RobloxPlayerBeta.exe, got %s", sp.CurrentGame())
	}

	time.Sleep(300 * time.Millisecond)

	// Stop
	if err := sp.Stop(); err != nil {
		t.Fatalf("spoofer Stop failed: %v", err)
	}

	if sp.IsRunning() {
		t.Fatalf("expected spoofer to be stopped")
	}
}
