package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func createBackup(path string, now time.Time) (string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	timestamp := now.UTC().Format("20060102-150405")
	backupPath := filepath.Join(dir, fmt.Sprintf("%s.bak-%s", base, timestamp))
	// Backups are never deleted, so never let one replace another.
	for n := 1; fileExists(backupPath); n++ {
		backupPath = filepath.Join(dir, fmt.Sprintf("%s.bak-%s-%d", base, timestamp, n))
	}

	if err := os.Rename(path, backupPath); err != nil {
		return "", fmt.Errorf("create backup: %w", err)
	}

	return backupPath, nil
}

func fileExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
