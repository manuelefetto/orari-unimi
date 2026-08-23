//go:build !windows

package tui

import (
	"os"

	"golang.org/x/sys/unix"
)

func dimensioniTerminale() (int, int, error) {
	dimensioni, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}
	return int(dimensioni.Col), int(dimensioni.Row), nil
}
