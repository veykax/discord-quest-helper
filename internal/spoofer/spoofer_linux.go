//go:build linux

package spoofer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type linuxManager struct {
	mu              sync.Mutex
	cmd             *exec.Cmd
	ipcConn         net.Conn
	currentName     string
	currentExe      string
	flatpakInstance string
	currentManifest string
	currentAppIdTxt string
	tempDir         string
}

func newPlatformManager() Manager {
	tDir := filepath.Join(os.TempDir(), "DiscordQuestHelper")
	_ = os.MkdirAll(tDir, 0755)
	return &linuxManager{
		tempDir: tDir,
	}
}

// findDiscordFlatpakInstance locates the active main Discord Flatpak container.
// It filters out helper/zygote containers by ensuring the X11 display socket exists.
func findDiscordFlatpakInstance() string {
	cmd := exec.Command("flatpak", "ps", "--columns=instance,application")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(out), "\n")
	var candidates []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.Contains(fields[1], "com.discordapp.Discord") {
			candidates = append(candidates, fields[0])
		}
	}

	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	// Find the instance that has the X11 socket and the most Discord processes (main instance)
	var bestInst string
	maxProcs := -1
	for _, inst := range candidates {
		chkSocket := exec.Command("flatpak", "enter", inst, "sh", "-c", "test -e /tmp/.X11-unix/X0 && echo OK")
		if sOut, err := chkSocket.Output(); err == nil && strings.TrimSpace(string(sOut)) == "OK" {
			chkProcs := exec.Command("flatpak", "enter", inst, "pidof", "Discord")
			numProcs := 0
			if pOut, err := chkProcs.Output(); err == nil {
				numProcs = len(strings.Fields(string(pOut)))
			}
			if numProcs > maxProcs {
				maxProcs = numProcs
				bestInst = inst
			}
		}
	}

	if bestInst != "" {
		return bestInst
	}
	return candidates[0]
}

func (m *linuxManager) Start(exeName, windowTitle, steamFolder, steamAppID, discordAppID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_ = m.stopUnlocked()

	// 1. Connect to Discord IPC socket (Rich Presence support)
	if discordAppID != "" {
		if conn, err := connectDiscordIPC(discordAppID, windowTitle); err == nil {
			m.ipcConn = conn
			fmt.Printf("[Quest Helper] Connected to Discord IPC for %s (AppID: %s)\n", windowTitle, discordAppID)
		} else {
			fmt.Printf("[Quest Helper] Discord IPC warning: %v\n", err)
		}
	}

	// 2. Prepare native process emulation
	self, err := os.Executable()
	if err != nil {
		if m.ipcConn != nil {
			m.currentName = exeName
			return nil
		}
		return err
	}

	flatpakInst := findDiscordFlatpakInstance()
	m.flatpakInstance = flatpakInst

	targetDir := m.tempDir
	if flatpakInst != "" {
		home, _ := os.UserHomeDir()
		flatpakCache := filepath.Join(home, ".var/app/com.discordapp.Discord/cache")
		if fi, err := os.Stat(flatpakCache); err == nil && fi.IsDir() {
			targetDir = flatpakCache
		}
	}

	// Optional Steam common library integration
	if steamFolder != "" {
		home, _ := os.UserHomeDir()
		steamCandidates := []string{
			filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam/steamapps"),
			filepath.Join(home, ".local/share/Steam/steamapps"),
			filepath.Join(home, ".steam/steam/steamapps"),
			filepath.Join(home, ".steam/root/steamapps"),
		}

		for _, sApps := range steamCandidates {
			if fi, err := os.Stat(sApps); err == nil && fi.IsDir() {
				sCommon := filepath.Join(sApps, "common")
				gameDir := filepath.Join(sCommon, steamFolder)
				_ = os.MkdirAll(gameDir, 0755)
				if flatpakInst == "" {
					targetDir = gameDir
				}

				if steamAppID != "" {
					manifestPath := filepath.Join(sApps, fmt.Sprintf("appmanifest_%s.acf", steamAppID))
					manifestData := fmt.Sprintf("\"AppState\"\n{\n\t\"appid\"\t\t\"%s\"\n\t\"universe\"\t\t\"1\"\n\t\"name\"\t\t\"%s\"\n\t\"StateFlags\"\t\t\"4\"\n\t\"installdir\"\t\t\"%s\"\n}\n",
						steamAppID, windowTitle, steamFolder)
					_ = os.WriteFile(manifestPath, []byte(manifestData), 0644)
					m.currentManifest = manifestPath

					appIdFile := filepath.Join(gameDir, "steam_appid.txt")
					_ = os.WriteFile(appIdFile, []byte(steamAppID), 0644)
					m.currentAppIdTxt = appIdFile
				}
				break
			}
		}
	}

	targetPath := filepath.Join(targetDir, filepath.FromSlash(exeName))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		if m.ipcConn != nil {
			m.currentName = exeName
			return nil
		}
		return err
	}

	if err := copyFile(self, targetPath); err != nil {
		if m.ipcConn != nil {
			m.currentName = exeName
			return nil
		}
		return err
	}
	_ = os.Chmod(targetPath, 0755)
	m.currentExe = targetPath

	var cmd *exec.Cmd
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}
	if flatpakInst != "" {
		cmd = exec.Command("flatpak", "enter", flatpakInst,
			"env", "DISPLAY="+display, "XAUTHORITY=/run/flatpak/Xauthority",
			targetPath, "--stub", windowTitle)
	} else {
		cmd = exec.Command(targetPath, "--stub", windowTitle)
	}

	if err := cmd.Start(); err != nil {
		if m.ipcConn != nil {
			m.currentName = exeName
			return nil
		}
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
			if m.ipcConn == nil {
				m.currentName = ""
			}
		}
		m.mu.Unlock()
	}()

	return nil
}

func (m *linuxManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopUnlocked()
}

func (m *linuxManager) stopUnlocked() error {
	if m.ipcConn != nil {
		_ = m.ipcConn.Close()
		m.ipcConn = nil
	}

	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		_ = m.cmd.Wait()
		m.cmd = nil
	}

	if m.flatpakInstance != "" && m.currentExe != "" {
		_ = exec.Command("flatpak", "enter", m.flatpakInstance, "pkill", "-f", filepath.Base(m.currentExe)).Run()
		m.flatpakInstance = ""
	}

	if m.currentManifest != "" {
		_ = os.Remove(m.currentManifest)
		m.currentManifest = ""
	}

	if m.currentAppIdTxt != "" {
		_ = os.Remove(m.currentAppIdTxt)
		m.currentAppIdTxt = ""
	}

	if m.currentExe != "" {
		_ = os.Remove(m.currentExe)
		parent := filepath.Dir(m.currentExe)
		if parent != m.tempDir && !strings.Contains(parent, "steamapps") && !strings.Contains(parent, ".var/app/com.discordapp.Discord/cache") {
			_ = os.Remove(parent)
		}
		m.currentExe = ""
	}

	m.currentName = ""
	return nil
}

func (m *linuxManager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ipcConn != nil || (m.cmd != nil && m.cmd.Process != nil)
}

func (m *linuxManager) CurrentGame() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentName
}

func findDiscordIPCSocket() string {
	uid := os.Getuid()
	candidates := []string{
		fmt.Sprintf("/run/user/%d/discord-ipc-0", uid),
		fmt.Sprintf("/run/user/%d/app/com.discordapp.Discord/discord-ipc-0", uid),
		fmt.Sprintf("/run/user/%d/.flatpak/com.discordapp.Discord/xdg-run/discord-ipc-0", uid),
		filepath.Join(os.TempDir(), "discord-ipc-0"),
	}
	for i := 0; i <= 9; i++ {
		candidates = append(candidates,
			fmt.Sprintf("/run/user/%d/discord-ipc-%d", uid, i),
			fmt.Sprintf("/run/user/%d/app/com.discordapp.Discord/discord-ipc-%d", uid, i),
		)
	}

	for _, p := range candidates {
		if fi, err := os.Stat(p); err == nil && (fi.Mode()&os.ModeSocket != 0) {
			return p
		}
	}
	return ""
}

func connectDiscordIPC(appID, windowTitle string) (net.Conn, error) {
	sock := findDiscordIPCSocket()
	if sock == "" {
		return nil, fmt.Errorf("discord socket not found")
	}

	conn, err := net.DialTimeout("unix", sock, 2*time.Second)
	if err != nil {
		return nil, err
	}

	// 1. Handshake
	hs, _ := json.Marshal(map[string]interface{}{
		"v":         1,
		"client_id": appID,
	})
	if err := sendIPCFrame(conn, 0, hs); err != nil {
		_ = conn.Close()
		return nil, err
	}

	_, _, err = readIPCFrame(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	// 2. SET_ACTIVITY with full game metadata
	act, _ := json.Marshal(map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid": os.Getpid(),
			"activity": map[string]interface{}{
				"type":    0,
				"name":    windowTitle,
				"details": windowTitle,
				"state":   "Playing (Discord Quest)",
				"timestamps": map[string]interface{}{
					"start": time.Now().Unix(),
				},
			},
		},
		"nonce": fmt.Sprintf("dqh-%d", time.Now().UnixNano()),
	})
	if err := sendIPCFrame(conn, 1, act); err != nil {
		_ = conn.Close()
		return nil, err
	}

	_, _, err = readIPCFrame(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	// Read remaining incoming frames in background so socket buffer does not block
	go func() {
		for {
			_, _, err := readIPCFrame(conn)
			if err != nil {
				return
			}
		}
	}()

	return conn, nil
}

func sendIPCFrame(conn net.Conn, opcode uint32, payload []byte) error {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, opcode)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(payload)))
	buf.Write(payload)
	_, err := conn.Write(buf.Bytes())
	return err
}

func readIPCFrame(conn net.Conn) (uint32, []byte, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(conn, header); err != nil {
		return 0, nil, err
	}
	opcode := binary.LittleEndian.Uint32(header[0:4])
	length := binary.LittleEndian.Uint32(header[4:8])
	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return opcode, nil, err
	}
	return opcode, payload, nil
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
