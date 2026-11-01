//go:build windows

package screen

import "syscall"

// GetCenter returns the centered (x, y) coordinates for a window of size (w, h).
func GetCenter(w, h int) (int, int) {
	user32 := syscall.NewLazyDLL("user32.dll")
	procMetrics := user32.NewProc("GetSystemMetrics")
	screenWidth, _, _ := procMetrics.Call(0)  // SM_CXSCREEN
	screenHeight, _, _ := procMetrics.Call(1) // SM_CYSCREEN
	if screenWidth > 0 && screenHeight > 0 {
		x := (int(screenWidth) - w) / 2
		y := (int(screenHeight) - h) / 2
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
		return x, y
	}
	return 100, 100
}
