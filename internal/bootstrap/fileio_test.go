package bootstrap

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAtomicWriteFileCreatesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte("content: value\n")

	if err := AtomicWriteFile(path, data, 0o600); err != nil {
		t.Fatalf("AtomicWriteFile returned error: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("unexpected file contents: %q", string(got))
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	// Windows has no Unix permission bits; Go reports any writable file
	// as 0666 there.
	if mode := info.Mode().Perm(); runtime.GOOS != "windows" && mode != 0o600 {
		t.Fatalf("expected mode 0600, got %v", mode)
	}
}

func TestAtomicWriteFileReplacesExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	if err := AtomicWriteFile(path, []byte("new"), 0o600); err != nil {
		t.Fatalf("AtomicWriteFile returned error: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("expected contents \"new\", got %q", string(got))
	}
}
