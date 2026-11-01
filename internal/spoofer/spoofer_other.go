//go:build !windows && !linux

package spoofer

import "errors"

type fallbackManager struct{}

func newPlatformManager() Manager {
	return &fallbackManager{}
}

func (m *fallbackManager) Start(exeName, windowTitle, steamFolder, steamAppID, discordAppID string) error {
	return errors.New("platform not supported")
}

func (m *fallbackManager) Stop() error {
	return nil
}

func (m *fallbackManager) IsRunning() bool {
	return false
}

func (m *fallbackManager) CurrentGame() string {
	return ""
}
