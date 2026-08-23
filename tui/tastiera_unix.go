//go:build !windows

package tui

import "github.com/eiannone/keyboard"

// ChiudiTastiera ripristina la modalità del terminale sui sistemi Unix.
func ChiudiTastiera() error {
	return keyboard.Close()
}
