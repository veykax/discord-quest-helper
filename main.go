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
	"runtime"
	"syscall"
	"time"

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
		launchNativeAppWindow(url)
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

func launchNativeAppWindow(url string) {
	width := "880"
	height := "640"

	switch runtime.GOOS {
	case "windows":
		// Пробуем запустить Edge в режиме Standalone App (без рамок браузера)
		edgePaths := []string{
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			filepathJoin(os.Getenv("LOCALAPPDATA"), `Microsoft\Edge\Application\msedge.exe`),
		}
		for _, p := range edgePaths {
			if _, err := os.Stat(p); err == nil {
				appArg := fmt.Sprintf("--app=%s", url)
				sizeArg := fmt.Sprintf("--window-size=%s,%s", width, height)
				cmd := exec.Command(p, appArg, sizeArg)
				if err := cmd.Start(); err == nil {
					return
				}
			}
		}
		// Fallback: запуск в браузере по умолчанию
		_ = exec.Command("cmd", "/c", "start", url).Start()

	case "linux":
		// Пробуем Chromium / Chrome в режиме App
		browsers := []string{"google-chrome", "chromium", "chromium-browser", "brave"}
		for _, b := range browsers {
			if p, err := exec.LookPath(b); err == nil {
				appArg := fmt.Sprintf("--app=%s", url)
				sizeArg := fmt.Sprintf("--window-size=%s,%s", width, height)
				cmd := exec.Command(p, appArg, sizeArg)
				if err := cmd.Start(); err == nil {
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

func filepathJoin(elem ...string) string {
	res := ""
	for i, e := range elem {
		if i == 0 {
			res = e
		} else {
			res = res + "\\" + e
		}
	}
	return res
}
