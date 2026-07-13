//go:build !windows

package main

import "os"

// enableVirtualTerminal is a no-op outside Windows; every supported Unix
// terminal understands ANSI escapes natively.
func enableVirtualTerminal(_ *os.File) bool {
	return true
}
