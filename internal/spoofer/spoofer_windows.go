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
)

type windowsManager struct {
	mu          sync.Mutex
	cmd         *exec.Cmd
	currentName string
	currentExe  string
	altExe      string
	tempDir     string
	baseStub    string
}

func newPlatformManager() Manager {
	tDir := filepath.Join(os.TempDir(), "DiscordQuestHelper")
	os.MkdirAll(tDir, 0755)
	return &windowsManager{
		tempDir:  tDir,
		baseStub: filepath.Join(tDir, "stub_base_v2.exe"),
	}
}

func (m *windowsManager) ensureStub() error {
	if _, err := os.Stat(m.baseStub); err == nil {
		return nil
	}

	// Исходник C# заглушки с видимым окном, запускающимся сразу свёрнутым
	csSource := `using System;
using System.Drawing;
using System.Windows.Forms;
namespace DiscordGameStub {
    static class Program {
        [STAThread]
        static void Main(string[] args) {
            string title = args.Length > 0 ? args[0] : "Game";
            Form form = new Form {
                Text = title,
                Width = 360,
                Height = 180,
                StartPosition = FormStartPosition.CenterScreen,
                FormBorderStyle = FormBorderStyle.FixedSingle,
                MaximizeBox = false,
                ShowInTaskbar = true,
                WindowState = FormWindowState.Minimized
            };
            Label lbl = new Label {
                Text = title + "\n\n(Эмуляция для квеста Discord)",
                Dock = DockStyle.Fill,
                TextAlign = ContentAlignment.MiddleCenter,
                Font = new Font("Segoe UI", 11, FontStyle.Regular)
            };
            form.Controls.Add(lbl);
            Application.Run(form);
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
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	// Fallback: копируем сам текущий исполняемый файл
	self, err := os.Executable()
	if err == nil {
		return copyFile(self, m.baseStub)
	}

	return fmt.Errorf("failed to build or copy stub")
}

func (m *windowsManager) Start(exeName string, windowTitle string, steamFolder string, steamAppID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_ = m.stopUnlocked()

	if err := m.ensureStub(); err != nil {
		return err
	}

	// Нормализуем путь для Windows
	exeName = filepath.FromSlash(exeName)

	if !strings.HasSuffix(strings.ToLower(exeName), ".exe") {
		exeName += ".exe"
	}

	targetDir := m.tempDir

	// Проверяем Steam-директорию
	steamCandidates := []string{
		`C:\Program Files (x86)\Steam\steamapps\common`,
		`C:\Program Files\Steam\steamapps\common`,
	}

	for _, sCommon := range steamCandidates {
		if steamFolder != "" {
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

					// Создаем steam_appid.txt прямо в папке игры
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

	// Копия с альтернативным регистром
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

	if m.currentExe != "" {
		_ = os.Remove(m.currentExe)
		parent := filepath.Dir(m.currentExe)
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
