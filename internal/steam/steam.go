package steam

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// LibraryFolder represents a detected Steam library directory containing a "steamapps" folder.
type LibraryFolder struct {
	Path         string
	SteamAppsDir string
	CommonDir    string
}

// GetAllLibraries locates all Steam libraries across all drives on the system.
func GetAllLibraries() []LibraryFolder {
	roots := GetSteamRoots()
	seen := make(map[string]bool)
	var libraries []LibraryFolder

	addLib := func(dir string) {
		dir = filepath.Clean(filepath.FromSlash(dir))
		if dir == "" || seen[strings.ToLower(dir)] {
			return
		}
		sApps := filepath.Join(dir, "steamapps")
		if fi, err := os.Stat(sApps); err == nil && fi.IsDir() {
			seen[strings.ToLower(dir)] = true
			libraries = append(libraries, LibraryFolder{
				Path:         dir,
				SteamAppsDir: sApps,
				CommonDir:    filepath.Join(sApps, "common"),
			})
		}
	}

	for _, root := range roots {
		addLib(root)

		// Parse libraryfolders.vdf for additional multi-drive libraries
		vdfPath := filepath.Join(root, "steamapps", "libraryfolders.vdf")
		if data, err := os.ReadFile(vdfPath); err == nil {
			vdfPaths := ParseLibraryFoldersVDF(string(data))
			for _, p := range vdfPaths {
				addLib(p)
			}
		}
	}

	return libraries
}

// ParseLibraryFoldersVDF extracts all library paths from a Steam libraryfolders.vdf text payload.
func ParseLibraryFoldersVDF(content string) []string {
	var paths []string
	seen := make(map[string]bool)

	addPath := func(raw string) {
		cleaned := strings.ReplaceAll(raw, `\\`, `\`)
		cleaned = filepath.Clean(filepath.FromSlash(cleaned))
		// Must look like an absolute filesystem directory (contains drive letter or root slash)
		if cleaned == "" || seen[strings.ToLower(cleaned)] {
			return
		}
		if (len(cleaned) >= 2 && cleaned[1] == ':') || strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "\\") {
			seen[strings.ToLower(cleaned)] = true
			paths = append(paths, cleaned)
		}
	}

	// Modern Steam format: "path" "D:\\SteamLibrary"
	reNew := regexp.MustCompile(`"path"\s+"([^"]+)"`)
	for _, m := range reNew.FindAllStringSubmatch(content, -1) {
		if len(m) == 2 {
			addPath(m[1])
		}
	}

	// Legacy Steam format: "1" "D:\\SteamLibrary"
	reOld := regexp.MustCompile(`"\d+"\s+"([A-Za-z]:\\[^"]+|/[^"]+)"`)
	for _, m := range reOld.FindAllStringSubmatch(content, -1) {
		if len(m) == 2 {
			addPath(m[1])
		}
	}

	return paths
}

// FindBestLibrary selects the optimal Steam library folder for the given game.
func FindBestLibrary(libraries []LibraryFolder, steamFolder, steamAppID string) LibraryFolder {
	if len(libraries) == 0 {
		return LibraryFolder{}
	}

	// 1. Look for an existing game folder
	if steamFolder != "" {
		for _, lib := range libraries {
			candidate := filepath.Join(lib.CommonDir, steamFolder)
			if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
				return lib
			}
		}
	}

	// 2. Look for an existing appmanifest
	if steamAppID != "" {
		for _, lib := range libraries {
			manifest := filepath.Join(lib.SteamAppsDir, fmt.Sprintf("appmanifest_%s.acf", steamAppID))
			if fi, err := os.Stat(manifest); err == nil && !fi.IsDir() {
				return lib
			}
		}
	}

	// 3. Fallback to the primary library
	return libraries[0]
}

// GetSteamRoots returns primary Steam installation directories based on OS and Registry.
func GetSteamRoots() []string {
	var roots []string
	seen := make(map[string]bool)

	addRoot := func(dir string) {
		dir = filepath.Clean(filepath.FromSlash(dir))
		if dir == "" || seen[strings.ToLower(dir)] {
			return
		}
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			seen[strings.ToLower(dir)] = true
			roots = append(roots, dir)
		}
	}

	switch runtime.GOOS {
	case "windows":
		// Query Windows Registry
		if regPath := getRegistrySteamPath(); regPath != "" {
			addRoot(regPath)
		}

		// Standard candidate locations on Windows
		progFiles := os.Getenv("ProgramFiles")
		progFilesX86 := os.Getenv("ProgramFiles(x86)")
		candidates := []string{
			filepath.Join(progFilesX86, "Steam"),
			filepath.Join(progFiles, "Steam"),
			`C:\Steam`,
			`D:\Steam`,
			`E:\Steam`,
			`D:\SteamLibrary`,
			`E:\SteamLibrary`,
		}
		for _, c := range candidates {
			addRoot(c)
		}

	case "linux":
		home, _ := os.UserHomeDir()
		if home != "" {
			linuxRoots := []string{
				filepath.Join(home, ".local/share/Steam"),
				filepath.Join(home, ".steam/steam"),
				filepath.Join(home, ".steam/root"),
				filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam"),
				filepath.Join(home, ".var/app/com.valvesoftware.Steam/.steam/steam"),
				filepath.Join(home, ".var/app/com.discordapp.Discord/.local/share/Steam"),
				filepath.Join(home, ".var/app/com.discordapp.Discord/.steam/steam"),
			}
			for _, r := range linuxRoots {
				addRoot(r)
			}
		}

	case "darwin":
		home, _ := os.UserHomeDir()
		if home != "" {
			addRoot(filepath.Join(home, "Library/Application Support/Steam"))
		}
	}

	return roots
}
