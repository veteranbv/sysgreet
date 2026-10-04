//go:build !linux

package collectors

// Only Linux keeps a wtmp history this collector can read; elsewhere the
// banner omits the line rather than guess.
func previousLogin(string, string) (*LastLoginInfo, error) { return nil, nil }

func currentTTY() string { return "" }
