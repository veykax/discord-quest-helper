//go:build windows

package spoofer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

type windowsManager struct {
	mu              sync.Mutex
	cmd             *exec.Cmd
	currentName     string
	currentExe      string
	altExe          string
	currentManifest string
	tempDir         string
	baseStub        string
}

func newPlatformManager() Manager {
	tDir := filepath.Join(os.TempDir(), "DiscordQuestHelper")
	os.MkdirAll(tDir, 0755)
	return &windowsManager{
		tempDir:  tDir,
		baseStub: filepath.Join(tDir, "stub_base_v5.exe"),
	}
}

func (m *windowsManager) ensureStub() error {
	if _, err := os.Stat(m.baseStub); err == nil {
		return nil
	}

	// C# source: registers an off-screen, non-activating window with title
	// for Discord EnumWindows scanner. 100% invisible to user, zero taskbar presence.
	csSource := `using System;
using System.Drawing;
using System.Windows.Forms;
namespace DiscordGameStub {
    class HiddenForm : Form {
        protected override bool ShowWithoutActivation { get { return true; } }
        public HiddenForm(string title) {
            Text = title;
            Width = 320;
            Height = 200;
            StartPosition = FormStartPosition.Manual;
            Location = new Point(-32000, -32000);
            ShowInTaskbar = false;
            Opacity = 0.0;
        }
    }
    static class Program {
        [STAThread]
        static void Main(string[] args) {
            string title = args.Length > 0 ? args[0] : "Game";
            Application.Run(new HiddenForm(title));
        }
    }
}`
	csPath := filepath.Join(m.tempDir, "stub.cs")
	if err := os.WriteFile(csPath, []byte(csSource), 0644); err != nil {
		return err
	}

	cscCandidates := []string{
		`C:\Windows\Microsoft.NET\Framework64\v4.0.30319\csc.exe`,
		`C:\Windows\Microsoft.NET\Framework\v4.0.30319\csc.exe`,
	}

	for _, csc := range cscCandidates {
		if _, err := os.Stat(csc); err == nil {
			cmd := exec.Command(csc, "/nologo", "/target:winexe",
				"/r:System.Windows.Forms.dll", "/r:System.Drawing.dll",
				fmt.Sprintf("/out:%s", m.baseStub), csPath)
			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow:    true,
				CreationFlags: 0x08000000, // CREATE_NO_WINDOW
			}
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	// Fallback: copy current executable
	self, err := os.Executable()
	if err == nil {
		return copyFile(self, m.baseStub)
	}

	return fmt.Errorf("failed to build or copy stub")
}

func (m *windowsManager) Start(exeName string, windowTitle string, steamFolder string, steamAppID string, discordAppID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_ = m.stopUnlocked()

	if err := m.ensureStub(); err != nil {
		return err
	}

	// Normalize Windows path
	exeName = filepath.FromSlash(exeName)

	if !strings.HasSuffix(strings.ToLower(exeName), ".exe") {
		exeName += ".exe"
	}

	targetDir := m.tempDir

	// Optional Steam common library check
	if steamFolder != "" {
		steamCandidates := []string{
			`C:\Program Files (x86)\Steam\steamapps\common`,
			`C:\Program Files\Steam\steamapps\common`,
		}

		for _, sCommon := range steamCandidates {
			if _, err := os.Stat(sCommon); err == nil {
				gameDir := filepath.Join(sCommon, steamFolder)
				os.MkdirAll(gameDir, 0755)
				targetDir = gameDir

				if steamAppID != "" {
					steamApps := filepath.Dir(sCommon)
					manifestPath := filepath.Join(steamApps, fmt.Sprintf("appmanifest_%s.acf", steamAppID))
					manifestData := fmt.Sprintf("\"AppState\"\n{\n\t\"appid\"\t\t\"%s\"\n\t\"universe\"\t\t\"1\"\n\t\"name\"\t\t\"%s\"\n\t\"StateFlags\"\t\t\"4\"\n\t\"installdir\"\t\t\"%s\"\n}\n",
						steamAppID, windowTitle, steamFolder)
					_ = os.WriteFile(manifestPath, []byte(manifestData), 0644)
					m.currentManifest = manifestPath

					appIdFile := filepath.Join(targetDir, "steam_appid.txt")
					_ = os.WriteFile(appIdFile, []byte(steamAppID), 0644)
				}
				break
			}
		}
	}

	targetPath := filepath.Join(targetDir, exeName)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	if err := copyFile(m.baseStub, targetPath); err != nil {
		return err
	}
	m.currentExe = targetPath

	// Alternate case copy
	baseName := filepath.Base(exeName)
	var altBaseName string
	if baseName == strings.ToLower(baseName) {
		altBaseName = strings.ToUpper(baseName)
	} else {
		altBaseName = strings.ToLower(baseName)
	}
	if altBaseName != baseName {
		altPath := filepath.Join(filepath.Dir(targetPath), altBaseName)
		_ = copyFile(m.baseStub, altPath)
		m.altExe = altPath
	}

	cmd := exec.Command(targetPath, windowTitle)
	if err := cmd.Start(); err != nil {
		return err
	}

	m.cmd = cmd
	m.currentName = exeName

	// Track process lifecycle in background
	go func() {
		_ = cmd.Wait()
		m.mu.Lock()
		if m.cmd == cmd {
			m.cmd = nil
			m.currentName = ""
		}
		m.mu.Unlock()
	}()

	return nil
}

func (m *windowsManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopUnlocked()
}

func (m *windowsManager) stopUnlocked() error {
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		_ = m.cmd.Wait()
		m.cmd = nil
	}

	if m.currentManifest != "" {
		_ = os.Remove(m.currentManifest)
		m.currentManifest = ""
	}

	if m.currentExe != "" {
		_ = os.Remove(m.currentExe)
		parent := filepath.Dir(m.currentExe)
		_ = os.Remove(filepath.Join(parent, "steam_appid.txt"))
		if parent != m.tempDir && !strings.Contains(strings.ToLower(parent), "steamapps") {
			_ = os.Remove(parent)
		}
		m.currentExe = ""
	}
	if m.altExe != "" {
		_ = os.Remove(m.altExe)
		m.altExe = ""
	}
	m.currentName = ""
	return nil
}

func (m *windowsManager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cmd != nil && m.cmd.Process != nil
}

func (m *windowsManager) CurrentGame() string {
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
