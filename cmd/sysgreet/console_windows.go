//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// enableVirtualTerminal switches the Windows console into VT processing so
// ANSI escapes render as colors instead of literal text. Returns false when
// the console refuses (legacy conhost), in which case output stays plain.
func enableVirtualTerminal(out *os.File) bool {
	handle := windows.Handle(out.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		// Not a console (piped/redirected); nothing to enable.
		return true
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		return true
	}
	return windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
}
