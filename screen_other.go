//go:build !windows

package main

func getScreenCenter(w, h int) (int, int) {
	return 100, 100
}
