//go:build !windows

package steam

func getRegistrySteamPath() string {
	return ""
}
