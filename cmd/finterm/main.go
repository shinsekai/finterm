// Package main provides the finterm application entry point.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/cache"
	"github.com/shinsekai/finterm/internal/config"
	"github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
	"github.com/shinsekai/finterm/internal/tui"
)

// Build-time variables set via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Handle version flag early, before config loading
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("finterm %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}
	// Load configuration
	cfgPath := getConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config from %s: %v\n", cfgPath, err)
		fmt.Fprintf(os.Stderr, "Please create a config file or set FINTERM_AV_API_KEY environment variable.\n")
		os.Exit(1)
	}

	// Validate that we have a usable config with API key
	if cfg.API.Key == "" {
		fmt.Fprintln(os.Stderr, "Error: API key is required for data features.")
		fmt.Fprintln(os.Stderr, "Set FINTERM_AV_API_KEY environment variable or configure api.key in config.yaml")
		//nolint:gocritic // defer won't run after os.Exit, but that's fine for startup error
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

	// Create cache store — prefer persistent SQLite, fall back to in-memory.
	var cacheStore cache.Cache
	cacheStore, err = cache.NewSQLite(xdgDataPath())
	if err != nil {
		log.Printf("Warning: failed to open persistent cache (%v), falling back to in-memory cache\n", err)
		cacheStore = cache.New()
	}

	// Create asset class detector with crypto symbols
	detector := indicators.NewAssetClassDetector(cfg.Watchlist.Crypto)

	// Create remote indicators (for equities)
	remoteRSI := indicators.NewRemoteRSI(avClient)
	remoteEMA := indicators.NewRemoteEMA(avClient)

	// Create local indicators (for crypto) - initialized with empty data
	// Data will be fetched when crypto symbols are queried via the fetcher
	localRSI := indicators.NewLocalRSI(nil)
	localEMA := indicators.NewLocalEMA(nil)

	// Create crypto data fetcher adapter
	cryptoFetcher := &cryptoFetcherAdapter{client: avClient}

	// Create trend engine
	trendEngine := trend.New(
		remoteRSI,
		remoteEMA,
		localRSI,
		localEMA,
		cfg,
		detector,
		cryptoFetcher,
		avClient,   // TimeSeriesClient for equity BLITZ computation
		cacheStore, // Cache for time series data
	)

	// Create theme
	theme := tui.NewTheme(cfg.Theme.Style)

	// Create the application with all dependencies wired
	// The avClient implements all required interfaces: quote.QuoteClient, macro.Client, news.Client
	app := tui.NewApp(
		theme,
		avClient, // quote.QuoteClient
		avClient, // macro.Client
		avClient, // news.Client
		trendEngine,
		cacheStore,
		&cfg.Watchlist,
		detector,
	)

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the TUI with alt screen and mouse support
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithContext(ctx),
	)

	// Run the program and handle signals
	programDone := make(chan error, 1)
	go func() {
		_, err := p.Run()
		programDone <- err
	}()

	// Wait for either program completion or signal
	exitCode := 0
	select {
	case err := <-programDone:
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			exitCode = 1
		}
		// Normal exit - cancel context to clean up any background goroutines
		cancel()

	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down gracefully...\n", sig)
		p.Quit()
		// Wait for program to finish cleanup
		<-programDone
		cancel()
	}

	if err := cacheStore.Close(); err != nil {
		log.Printf("Warning: failed to close cache: %v\n", err)
	}
	os.Exit(exitCode)
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

// cryptoFetcherAdapter adapts the Alpha Vantage client to CryptoDataFetcher interface.
// It fetches crypto daily OHLCV data and converts it to the domain OHLCV format.
type cryptoFetcherAdapter struct {
	client *alphavantage.Client
}

// FetchCryptoOHLCV fetches and converts crypto daily OHLCV data to domain types.
// The returned slice is sorted oldest-first as required by local indicators.
// Excludes today's in-progress bar (bar-close-only rule).
func (a *cryptoFetcherAdapter) FetchCryptoOHLCV(ctx context.Context, symbol string) ([]indicators.OHLCV, error) {
	data, err := a.client.GetCryptoDaily(ctx, symbol, "USD")
	if err != nil {
		return nil, err
	}

	// Get today's date in UTC to skip the in-progress bar
	today := time.Now().UTC().Format("2006-01-02")

	// Convert to OHLCV slice
	ohlcvSlice := make([]indicators.OHLCV, 0, len(data.TimeSeries))
	for dateStr, entry := range data.TimeSeries {
		// Skip today's in-progress bar (bar-close-only rule)
		if dateStr >= today {
			continue
		}

		date, err := alphavantage.ParseDate(dateStr)
		if err != nil {
			continue
		}
		open, _ := alphavantage.ParseFloat(entry.Open)
		high, _ := alphavantage.ParseFloat(entry.High)
		low, _ := alphavantage.ParseFloat(entry.Low)
		closeVal, _ := alphavantage.ParseFloat(entry.Close)
		volume, _ := alphavantage.ParseFloat(entry.Volume)

		ohlcvSlice = append(ohlcvSlice, indicators.OHLCV{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closeVal,
			Volume: volume,
		})
	}

	// Sort oldest-first (local indicators expect chronological order)
	sort.Slice(ohlcvSlice, func(i, j int) bool {
		return ohlcvSlice[i].Date.Before(ohlcvSlice[j].Date)
	})

	return ohlcvSlice, nil
}
