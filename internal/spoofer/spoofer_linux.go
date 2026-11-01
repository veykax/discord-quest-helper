//go:build !windows

package spoofer

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type linuxManager struct {
	mu          sync.Mutex
	cmd         *exec.Cmd
	currentName string
	currentExe  string
	tempDir     string
}

func newPlatformManager() Manager {
	tDir := filepath.Join(os.TempDir(), "DiscordQuestHelper")
	os.MkdirAll(tDir, 0755)
	return &linuxManager{
		tempDir: tDir,
	}
}

func (m *linuxManager) Start(exeName string, windowTitle string, steamFolder string, steamAppID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_ = m.stopUnlocked()

	self, err := os.Executable()
	if err != nil {
		return err
	}

	targetPath := filepath.Join(m.tempDir, exeName)
	if err := copyFile(self, targetPath); err != nil {
		return err
	}
	os.Chmod(targetPath, 0755)
	m.currentExe = targetPath

	// Запуск процесса в фоновом режиме
	cmd := exec.Command(targetPath, "--stub")
	if err := cmd.Start(); err != nil {
		return err
	}

	m.cmd = cmd
	m.currentName = exeName
	return nil
}

func (m *linuxManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopUnlocked()
}

func (m *linuxManager) stopUnlocked() error {
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		_ = m.cmd.Wait()
		m.cmd = nil
	}

	if m.currentExe != "" {
		_ = os.Remove(m.currentExe)
		m.currentExe = ""
	}
	m.currentName = ""
	return nil
}

func (m *linuxManager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cmd != nil && m.cmd.Process != nil
}

func (m *linuxManager) CurrentGame() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentName
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
