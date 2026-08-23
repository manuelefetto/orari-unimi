//go:build windows

package tui

import (
	"os"

	"golang.org/x/sys/windows"
)

func dimensioniTerminale() (int, int, error) {
	var informazioni windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(windows.Handle(os.Stdout.Fd()), &informazioni)
	if err != nil {
		return 0, 0, err
	}
	larghezza := int(informazioni.Window.Right-informazioni.Window.Left) + 1
	altezza := int(informazioni.Window.Bottom-informazioni.Window.Top) + 1
	return larghezza, altezza, nil
}
