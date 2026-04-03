// Package main provides the finterm application entry point.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/owner/finterm/internal/alphavantage"
	"github.com/owner/finterm/internal/config"
	"github.com/owner/finterm/internal/domain/trend"
	"github.com/owner/finterm/internal/domain/trend/indicators"
	"github.com/owner/finterm/internal/tui"
)

func main() {
	// Load configuration
	cfgPath := getConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Using default config. Set FINTERM_AV_API_KEY environment variable or create config file.\n")
		cfg = config.DefaultConfig()
	}

	// Validate that we have at least some usable config
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Config validation error: %v\n", err)
		os.Exit(1)
	}

	// Create Alpha Vantage client
	avClient := alphavantage.New(alphavantage.Config{
		Key:        cfg.API.Key,
		BaseURL:    cfg.API.BaseURL,
		RateLimit:  cfg.API.RateLimit,
		Timeout:    cfg.API.Timeout,
		MaxRetries: cfg.API.MaxRetries,
	})

	// Create asset class detector with crypto symbols
	detector := indicators.NewAssetClassDetector(cfg.Watchlist.Crypto)

	// Create remote indicators (for equities)
	remoteRSI := indicators.NewRemoteRSI(avClient)
	remoteEMA := indicators.NewRemoteEMA(avClient)

	// Create local indicators (for crypto) - initialized with empty data
	// Data will be loaded when crypto symbols are queried
	localRSI := indicators.NewLocalRSI(nil)
	localEMA := indicators.NewLocalEMA(nil)

	// Create trend engine
	trendEngine := trend.New(
		remoteRSI,
		remoteEMA,
		localRSI,
		localEMA,
		cfg,
		detector,
	)

	// Create theme
	theme := tui.NewTheme(cfg.Theme.Style)

	// Create and start the application
	app := tui.NewApp(theme, avClient, trendEngine)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting TUI: %v\n", err)
		os.Exit(1)
	}
}

// getConfigPath returns the path to the config file.
func getConfigPath() string {
	// Check if FINTERM_CONFIG environment variable is set
	if envPath := os.Getenv("FINTERM_CONFIG"); envPath != "" {
		return envPath
	}

	// Use default path in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(homeDir, ".config", "finterm", "config.yaml")
}

// validateConfig performs basic validation that we have enough config to run.
func validateConfig(cfg *config.Config) error {
	// In demo mode (without API key), we can still run the TUI
	// Just warn that data won't be available
	if cfg.API.Key == "" {
		fmt.Fprintln(os.Stderr, "Warning: No API key configured. Data features will be unavailable.")
		fmt.Fprintln(os.Stderr, "Set FINTERM_AV_API_KEY environment variable or configure api.key in config.yaml")
	}
	return nil
}
