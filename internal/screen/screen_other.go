//go:build !windows

package screen

import (
	"os/exec"
	"regexp"
	"strconv"
)

// GetCenter returns the centered (x, y) coordinates for a window of size (w, h).
func GetCenter(w, h int) (int, int) {
	out, err := exec.Command("xrandr", "--current").Output()
	if err == nil {
		re := regexp.MustCompile(`current\s+(\d+)\s+x\s+(\d+)`)
		matches := re.FindSubmatch(out)
		if len(matches) == 3 {
			sw, errW := strconv.Atoi(string(matches[1]))
			sh, errH := strconv.Atoi(string(matches[2]))
			if errW == nil && errH == nil && sw > 0 && sh > 0 {
				x := (sw - w) / 2
				y := (sh - h) / 2
				if x < 0 {
					x = 0
				}
				if y < 0 {
					y = 0
				}
				return x, y
			}
		}
	}
	return 100, 100
}
