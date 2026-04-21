package main

import (
	"os"
	"path/filepath"
	"runtime"
)

// xdgDataPath returns the path to the finterm SQLite database file.
// It respects $XDG_DATA_HOME on Linux/macOS, %LOCALAPPDATA% on Windows,
// and falls back to $HOME/.local/share on Unix-like systems.
func xdgDataPath() string {
	if runtime.GOOS == "windows" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "finterm", "finterm.db")
		}
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "finterm", "finterm.db")
	}

	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "finterm", "finterm.db")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "finterm.db")
	}
	return filepath.Join(homeDir, ".local", "share", "finterm", "finterm.db")
}
