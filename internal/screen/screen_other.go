//go:build !windows

package screen

// GetCenter returns fallback (x, y) coordinates for non-Windows platforms.
func GetCenter(w, h int) (int, int) {
	return 100, 100
}
