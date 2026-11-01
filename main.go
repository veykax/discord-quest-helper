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
	"syscall"
	"time"

	"discord-quest-helper/internal/screen"
	"discord-quest-helper/internal/server"
	"discord-quest-helper/internal/spoofer"
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

	// Корневая поддиректория для встраивания web/
	webSubFS, err := fs.Sub(embeddedWebFS, "web")
	if err != nil {
		fmt.Printf("Error loading web files: %v\n", err)
		os.Exit(1)
	}

	appServer := server.New(sp, webSubFS, exitChan)
	mux := http.NewServeMux()
	appServer.RegisterRoutes(mux)

	// Находим свободный порт на localhost
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("Error starting listener: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

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
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
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

		// Fallback: запуск в браузере по умолчанию
		_ = exec.Command("cmd", "/c", "start", url).Start()

	case "linux":
		browsers := []string{
			"google-chrome",
			"google-chrome-stable",
			"chromium",
			"chromium-browser",
			"brave-browser",
			"microsoft-edge",
			"vivaldi",
		}
		profileDir := filepath.Join(os.TempDir(), "DiscordQuestHelper_AppProfile")
		appArg := fmt.Sprintf("--app=%s", url)
		sizeArg := fmt.Sprintf("--window-size=%s,%s", width, height)
		profileArg := fmt.Sprintf("--user-data-dir=%s", profileDir)

		for _, b := range browsers {
			if p, err := exec.LookPath(b); err == nil {
				cmd := exec.Command(p,
					appArg,
					sizeArg,
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

	candidates := []string{
		// 1. Google Chrome (наиболее надежный standalone app режим)
		filepath.Join(localAppData, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(progFiles, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(progFilesX86, `Google\Chrome\Application\chrome.exe`),

		// 2. Microsoft Edge
		filepath.Join(progFilesX86, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(progFiles, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(localAppData, `Microsoft\Edge\Application\msedge.exe`),

		// 3. Vivaldi
		filepath.Join(localAppData, `Vivaldi\Application\vivaldi.exe`),
		filepath.Join(progFiles, `Vivaldi\Application\vivaldi.exe`),
		filepath.Join(progFilesX86, `Vivaldi\Application\vivaldi.exe`),

		// 4. Brave
		filepath.Join(localAppData, `BraveSoftware\Brave-Browser\Application\brave.exe`),
		filepath.Join(progFiles, `BraveSoftware\Brave-Browser\Application\brave.exe`),

		// 5. Opera
		filepath.Join(localAppData, `Programs\Opera GX\opera.exe`),
		filepath.Join(localAppData, `Programs\Opera\launcher.exe`),
		filepath.Join(progFiles, `Opera\launcher.exe`),
	}

	for _, p := range candidates {
		if p == "" {
			continue
		}
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}

	// Также проверяем в PATH
	for _, name := range []string{"chrome.exe", "msedge.exe", "vivaldi.exe", "brave.exe"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}

	return ""
}
