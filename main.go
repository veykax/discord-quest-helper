package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"discord-quest-helper/internal/config"
	"discord-quest-helper/internal/screen"
	"discord-quest-helper/internal/server"
	"discord-quest-helper/internal/spoofer"
	"discord-quest-helper/internal/stub"
)

//go:embed web/*
var embeddedWebFS embed.FS

func main() {
	// Если процесс запущен как заглушка на Linux:
	if len(os.Args) > 1 && os.Args[1] == "--stub" {
		runStubMode()
		return
	}

	sp := spoofer.New()
	exitChan := make(chan struct{})

	// Фоновое обновление базы игр с GitHub
	go func() {
		_, _ = config.FetchRemoteGames()
	}()

	// Корневая поддиректория для встраивания web/
	webSubFS, err := fs.Sub(embeddedWebFS, "web")
	if err != nil {
		fmt.Printf("Error loading web files: %v\n", err)
		os.Exit(1)
	}

	appServer := server.New(sp, webSubFS, exitChan)
	mux := http.NewServeMux()
	appServer.RegisterRoutes(mux)

	// Пробуем порт по умолчанию или свободный порт на localhost
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", config.DefaultPort))
	if err != nil {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Printf("Error starting listener: %v\n", err)
			os.Exit(1)
		}
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Printf("Discord Quest Helper running at %s\n", url)

	// Запуск HTTP сервера в фоне
	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Открываем нативное окно приложения (App Mode)
	go func() {
		time.Sleep(150 * time.Millisecond)
		launchNativeAppWindow(url, exitChan)
	}()

	// Обработка прерывания (Ctrl+C, закрытие)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-exitChan:
	case <-sigChan:
	}

	// Корректная остановка эмулятора при выходе
	_ = sp.Stop()
	_ = srv.Close()
}

func runStubMode() {
	title := "Game"
	if len(os.Args) > 2 {
		title = strings.Join(os.Args[2:], " ")
	}
	stub.Run(title)
}

func launchNativeAppWindow(url string, exitChan chan struct{}) {
	width := 960
	height := 680
	posX, posY := screen.GetCenter(width, height)

	switch runtime.GOOS {
	case "windows":
		browserPath := findWindowsChromium()
		if browserPath != "" {
			profileDir := filepath.Join(os.TempDir(), "DiscordQuestHelper_AppProfile")
			appArg := fmt.Sprintf("--app=%s", url)
			sizeArg := fmt.Sprintf("--window-size=%d,%d", width, height)
			posArg := fmt.Sprintf("--window-position=%d,%d", posX, posY)
			profileArg := fmt.Sprintf("--user-data-dir=%s", profileDir)

			cmd := exec.Command(browserPath,
				appArg,
				sizeArg,
				posArg,
				profileArg,
				"--no-first-run",
				"--no-default-browser-check",
				"--disable-extensions",
				"--disable-sync",
				"--disable-background-networking",
				"--disable-component-update",
				"--disable-default-apps",
				"--renderer-process-limit=1",
				"--disable-features=Translate,OptimizationHints,MediaRouter",
			)

			if err := cmd.Start(); err == nil {
				// При закрытии окна пользователем - плавно закрываем приложение
				go func() {
					_ = cmd.Wait()
					select {
					case <-exitChan:
					default:
						close(exitChan)
					}
				}()
				return
			}
		}

		// Fallback for non-Chromium environments (e.g. Zen Browser, Firefox)
		localAppData := os.Getenv("LOCALAPPDATA")
		progFiles := os.Getenv("ProgramFiles")
		progFilesX86 := os.Getenv("ProgramFiles(x86)")
		geckoCandidates := []string{
			filepath.Join(localAppData, `Zen Browser\zen.exe`),
			filepath.Join(progFiles, `Zen Browser\zen.exe`),
			filepath.Join(progFilesX86, `Zen Browser\zen.exe`),
			filepath.Join(progFiles, `Mozilla Firefox\firefox.exe`),
			filepath.Join(progFilesX86, `Mozilla Firefox\firefox.exe`),
			filepath.Join(localAppData, `Mozilla Firefox\firefox.exe`),
		}
		for _, gp := range geckoCandidates {
			if fi, err := os.Stat(gp); err == nil && !fi.IsDir() {
				cmd := exec.Command(gp, "--new-window", url)
				if err := cmd.Start(); err == nil {
					return
				}
			}
		}

		// Fallback: запуск в браузере по умолчанию
		_ = exec.Command("cmd", "/c", "start", url).Start()

	case "linux":
		ensureLinuxDesktopEntry()

		home, _ := os.UserHomeDir()
		profileDir := filepath.Join(home, ".local", "share", "discord-quest-helper", "browser-profile")
		_ = os.MkdirAll(profileDir, 0755)

		browsers := []string{
			"google-chrome",
			"google-chrome-stable",
			"chromium",
			"chromium-browser",
			"brave-browser",
			"microsoft-edge",
			"vivaldi",
		}
		appArg := fmt.Sprintf("--app=%s", url)
		sizeArg := fmt.Sprintf("--window-size=%d,%d", width, height)
		posArg := fmt.Sprintf("--window-position=%d,%d", posX, posY)
		profileArg := fmt.Sprintf("--user-data-dir=%s", profileDir)

		for _, b := range browsers {
			if p, err := exec.LookPath(b); err == nil {
				cmd := exec.Command(p,
					appArg,
					sizeArg,
					posArg,
					profileArg,
					"--class=discord-quest-helper",
					"--name=discord-quest-helper",
					"--no-first-run",
					"--no-default-browser-check",
					"--disable-extensions",
					"--disable-sync",
					"--disable-background-networking",
					"--disable-component-update",
					"--disable-default-apps",
					"--renderer-process-limit=1",
					"--disable-features=Translate,OptimizationHints,MediaRouter",
				)
				if err := cmd.Start(); err == nil {
					go func() {
						startTime := time.Now()
						_ = cmd.Wait()
						if time.Since(startTime) > 3*time.Second {
							select {
							case <-exitChan:
							default:
								close(exitChan)
							}
						}
					}()
					return
				}
			}
		}

		// Try flatpak browser if native binary not in PATH
		if _, err := exec.LookPath("flatpak"); err == nil {
			flatpakBrowsers := []string{
				"com.vivaldi.Vivaldi",
				"com.google.Chrome",
				"org.chromium.Chromium",
				"com.brave.Browser",
				"com.microsoft.Edge",
			}
			for _, fb := range flatpakBrowsers {
				if err := exec.Command("flatpak", "info", fb).Run(); err == nil {
					cmd := exec.Command("flatpak", "run", fb,
						appArg,
						sizeArg,
						posArg,
						profileArg,
						"--class=discord-quest-helper",
						"--name=discord-quest-helper",
						"--no-first-run",
						"--no-default-browser-check",
						"--disable-extensions",
					)
					if err := cmd.Start(); err == nil {
						go func() {
							startTime := time.Now()
							_ = cmd.Wait()
							if time.Since(startTime) > 3*time.Second {
								select {
								case <-exitChan:
								default:
									close(exitChan)
								}
							}
						}()
						return
					}
				}
			}
		}

		// Fallback: xdg-open
		_ = exec.Command("xdg-open", url).Start()

	case "darwin":
		_ = exec.Command("open", url).Start()
	}
}

func findWindowsChromium() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	progFiles := os.Getenv("ProgramFiles")
	progFilesX86 := os.Getenv("ProgramFiles(x86)")

	// 1. Direct candidates
	candidates := []string{
		// Google Chrome (наиболее распространенный standalone app режим)
		filepath.Join(localAppData, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(progFiles, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(progFilesX86, `Google\Chrome\Application\chrome.exe`),

		// Standard Edge
		filepath.Join(progFilesX86, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(progFiles, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(localAppData, `Microsoft\Edge\Application\msedge.exe`),

		// Brave
		filepath.Join(localAppData, `BraveSoftware\Brave-Browser\Application\brave.exe`),
		filepath.Join(progFiles, `BraveSoftware\Brave-Browser\Application\brave.exe`),

		// Vivaldi
		filepath.Join(localAppData, `Vivaldi\Application\vivaldi.exe`),
		filepath.Join(progFiles, `Vivaldi\Application\vivaldi.exe`),
		filepath.Join(progFilesX86, `Vivaldi\Application\vivaldi.exe`),

		// Opera / Opera GX
		filepath.Join(localAppData, `Programs\Opera GX\opera.exe`),
		filepath.Join(localAppData, `Programs\Opera\launcher.exe`),
		filepath.Join(progFiles, `Opera\launcher.exe`),

		// Yandex
		filepath.Join(localAppData, `Yandex\YandexBrowser\Application\browser.exe`),
		filepath.Join(progFilesX86, `Yandex\YandexBrowser\Application\browser.exe`),

		// Thorium & Chromium
		filepath.Join(localAppData, `Thorium\Application\thorium.exe`),
		filepath.Join(progFiles, `Thorium\Application\thorium.exe`),
		filepath.Join(localAppData, `Chromium\Application\chrome.exe`),
	}

	for _, p := range candidates {
		if p == "" {
			continue
		}
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}

	// 2. Modern Windows 11 EdgeCore & EdgeWebView (side-by-side versioned directories)
	globPatterns := []string{
		filepath.Join(progFilesX86, `Microsoft\EdgeCore\*\msedge.exe`),
		filepath.Join(progFiles, `Microsoft\EdgeCore\*\msedge.exe`),
		filepath.Join(localAppData, `Microsoft\EdgeCore\*\msedge.exe`),
		filepath.Join(progFilesX86, `Microsoft\EdgeWebView\Application\*\msedgewebview2.exe`),
		filepath.Join(progFiles, `Microsoft\EdgeWebView\Application\*\msedgewebview2.exe`),
		filepath.Join(localAppData, `Microsoft\EdgeWebView\Application\*\msedgewebview2.exe`),
		filepath.Join(progFilesX86, `Microsoft\Edge\Application\*\msedge.exe`),
		filepath.Join(progFiles, `Microsoft\Edge\Application\*\msedge.exe`),
	}

	for _, pattern := range globPatterns {
		if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
			for i := len(matches) - 1; i >= 0; i-- {
				if fi, err := os.Stat(matches[i]); err == nil && !fi.IsDir() {
					return matches[i]
				}
			}
		}
	}

	// 3. Также проверяем в PATH
	for _, name := range []string{"chrome.exe", "msedge.exe", "vivaldi.exe", "brave.exe", "opera.exe"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}

	return ""
}

func ensureLinuxDesktopEntry() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// 1. Install hicolor icons in multiple standard resolutions
	sizes := []struct {
		folder string
		file   string
	}{
		{"64x64", "web/favicon.png"},
		{"192x192", "web/icon-192.png"},
		{"512x512", "web/icon-512.png"},
	}
	for _, s := range sizes {
		dir := filepath.Join(home, ".local", "share", "icons", "hicolor", s.folder, "apps")
		_ = os.MkdirAll(dir, 0755)
		if data, err := embeddedWebFS.ReadFile(s.file); err == nil {
			_ = os.WriteFile(filepath.Join(dir, "discord-quest-helper.png"), data, 0644)
		}
	}

	// Fallback in pixmaps
	pixmapsDir := filepath.Join(home, ".local", "share", "pixmaps")
	_ = os.MkdirAll(pixmapsDir, 0755)
	if data, err := embeddedWebFS.ReadFile("web/icon-192.png"); err == nil {
		_ = os.WriteFile(filepath.Join(pixmapsDir, "discord-quest-helper.png"), data, 0644)
	}

	appDir := filepath.Join(home, ".local", "share", "applications")
	_ = os.MkdirAll(appDir, 0755)
	desktopPath := filepath.Join(appDir, "discord-quest-helper.desktop")
	desktopContent := `[Desktop Entry]
Type=Application
Name=Discord Quest Helper
Comment=Game process emulator for Discord Quest rewards
Icon=discord-quest-helper
StartupWMClass=discord-quest-helper
Terminal=false
Categories=Utility;
`
	_ = os.WriteFile(desktopPath, []byte(desktopContent), 0644)

	// Refresh system cache for KDE Plasma and GNOME
	_ = exec.Command("update-desktop-database", appDir).Run()
	_ = exec.Command("kbuildsycoca6").Run()
}
