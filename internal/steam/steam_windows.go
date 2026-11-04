//go:build windows

package steam

import (
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func getRegistrySteamPath() string {
	cmd := exec.Command("reg", "query", `HKCU\Software\Valve\Steam`, "/v", "SteamPath")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "SteamPath") {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					rawVal := strings.Join(fields[2:], " ")
					return filepath.FromSlash(strings.TrimSpace(rawVal))
				}
			}
		}
	}
	return ""
}
