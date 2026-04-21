package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestXDGDataPath_HonorsXDGDataHome verifies that $XDG_DATA_HOME is respected.
func TestXDGDataPath_HonorsXDGDataHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}

	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	got := xdgDataPath()
	want := filepath.Join(tmpDir, "finterm", "finterm.db")
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// TestXDGDataPath_FallsBackToHome verifies fallback to $HOME/.local/share.
func TestXDGDataPath_FallsBackToHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}

	t.Setenv("XDG_DATA_HOME", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	got := xdgDataPath()
	want := filepath.Join(home, ".local", "share", "finterm", "finterm.db")
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// TestXDGDataPath_WindowsLocalAppData verifies Windows %LOCALAPPDATA% path.
func TestXDGDataPath_WindowsLocalAppData(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("skipping on non-Windows")
	}

	tmpDir := t.TempDir()
	t.Setenv("LOCALAPPDATA", tmpDir)

	got := xdgDataPath()
	want := filepath.Join(tmpDir, "finterm", "finterm.db")
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// TestXDGDataPath_CreatesParentDirs verifies that the returned path can be used for DB creation.
func TestXDGDataPath_CreatesParentDirs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}

	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	path := xdgDataPath()
	if !strings.HasSuffix(path, "finterm.db") {
		t.Errorf("expected path to end with finterm.db, got %q", path)
	}
}
