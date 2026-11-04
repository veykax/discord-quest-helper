package steam

import (
	"path/filepath"
	"testing"
)

func TestParseLibraryFoldersVDF(t *testing.T) {
	vdfData := `"libraryfolders"
{
	"0"
	{
		"path"		"C:\\Program Files (x86)\\Steam"
		"label"		""
		"contentid"		"3256074335729261489"
		"totalsize"		"0"
		"update_clean_bytes_tally"		"102931551"
		"time_last_update_verified"		"1792674972"
		"apps"
		{
			"228980"		"234737537"
			"2139460"		"0"
		}
	}
	"1"
	{
		"path"		"D:\\SteamLibrary"
		"label"		"FastSSD"
		"contentid"		"4849800525158520319"
		"totalsize"		"960178941952"
		"update_clean_bytes_tally"		"2148216562"
		"time_last_update_verified"		"1793475691"
		"apps"
		{
			"108600"		"10335862594"
			"289070"		"28049031644"
		}
	}
}`

	paths := ParseLibraryFoldersVDF(vdfData)
	if len(paths) != 2 {
		t.Fatalf("expected 2 library paths, got %d: %v", len(paths), paths)
	}

	expected0 := filepath.Clean(`C:\Program Files (x86)\Steam`)
	expected1 := filepath.Clean(`D:\SteamLibrary`)

	if paths[0] != expected0 {
		t.Errorf("expected paths[0] to be %q, got %q", expected0, paths[0])
	}
	if paths[1] != expected1 {
		t.Errorf("expected paths[1] to be %q, got %q", expected1, paths[1])
	}
}

func TestLegacyLibraryFoldersVDF(t *testing.T) {
	legacyVDF := `"LibraryFolders"
{
	"TimeNextStatsReport"	"1580000000"
	"ContentStatsID"		"12345"
	"1"						"E:\\Games\\Steam"
}`

	paths := ParseLibraryFoldersVDF(legacyVDF)
	if len(paths) != 1 {
		t.Fatalf("expected 1 library path, got %d: %v", len(paths), paths)
	}

	expected := filepath.Clean(`E:\Games\Steam`)
	if paths[0] != expected {
		t.Errorf("expected %q, got %q", expected, paths[0])
	}
}

func TestFindBestLibraryFallback(t *testing.T) {
	libs := []LibraryFolder{
		{
			Path:         `C:\Steam`,
			SteamAppsDir: `C:\Steam\steamapps`,
			CommonDir:    `C:\Steam\steamapps\common`,
		},
		{
			Path:         `D:\SteamLibrary`,
			SteamAppsDir: `D:\SteamLibrary\steamapps`,
			CommonDir:    `D:\SteamLibrary\steamapps\common`,
		},
	}

	// Unknown game falls back to primary library
	best := FindBestLibrary(libs, "NonExistentGameFolder", "9999999")
	if best.Path != `C:\Steam` {
		t.Errorf("expected fallback to primary library C:\\Steam, got %s", best.Path)
	}
}

func TestGetAllLibrariesLive(t *testing.T) {
	libs := GetAllLibraries()
	t.Logf("Detected %d Steam libraries on this host:", len(libs))
	for i, l := range libs {
		t.Logf("  [%d] Path: %s | Apps: %s | Common: %s", i, l.Path, l.SteamAppsDir, l.CommonDir)
	}
}
