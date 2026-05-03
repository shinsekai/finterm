// Package config provides configuration loading and validation for finterm.
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	envAPIKey = "FINTERM_AV_API_KEY"
)

// Config holds the complete application configuration.
type Config struct {
	API         APIConfig         `yaml:"api"`
	Watchlist   WatchlistConfig   `yaml:"watchlist"`
	Trend       TrendConfig       `yaml:"trend"`
	Valuation   ValuationConfig   `yaml:"valuation"`
	Commodities CommoditiesConfig `yaml:"commodities"`
	Cache       CacheConfig       `yaml:"cache"`
	Theme       ThemeConfig       `yaml:"theme"`
}

// APIConfig holds Alpha Vantage API configuration.
type APIConfig struct {
	Key        string        `yaml:"key"`
	BaseURL    string        `yaml:"base_url"`
	RateLimit  int           `yaml:"rate_limit"`
	BurstLimit int           `yaml:"burst_limit"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
}

// WatchlistConfig holds ticker watchlist configuration.
type WatchlistConfig struct {
	Equities []string `yaml:"equities"`
	Crypto   []string `yaml:"crypto"`
}

// TrendConfig holds trend-following configuration.
type TrendConfig struct {
	RSIPeriod int `yaml:"rsi_period"`
	EMAFast   int `yaml:"ema_fast"`
	EMASlow   int `yaml:"ema_slow"`
}

// ValuationConfig holds valuation configuration.
type ValuationConfig struct {
	RSIPeriod   int `yaml:"rsi_period"`
	Oversold    int `yaml:"oversold"`
	Undervalued int `yaml:"undervalued"`
	FairLow     int `yaml:"fair_low"`
	FairHigh    int `yaml:"fair_high"`
	Overvalued  int `yaml:"overvalued"`
	Overbought  int `yaml:"overbought"`
}

// CacheConfig holds cache TTL configuration.
type CacheConfig struct {
	IntradayTTL time.Duration `yaml:"intraday_ttl"`
	DailyTTL    time.Duration `yaml:"daily_ttl"`
	MacroTTL    time.Duration `yaml:"macro_ttl"`
	NewsTTL     time.Duration `yaml:"news_ttl"`
	CryptoTTL   time.Duration `yaml:"crypto_ttl"`
}

// ThemeConfig holds theme configuration.
type ThemeConfig struct {
	Style string `yaml:"style"`
}

// CommoditiesConfig holds commodities dashboard configuration.
type CommoditiesConfig struct {
	Watchlist []string `yaml:"watchlist"`
	Interval  string   `yaml:"interval"`
}

// DefaultConfig returns a Config with default values applied.
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			BaseURL:    "https://www.alphavantage.co/query",
			RateLimit:  70,
			BurstLimit: 5,
			Timeout:    10 * time.Second,
			MaxRetries: 3,
		},
		Watchlist: WatchlistConfig{
			Equities: []string{"QQQ", "SPY"},
			Crypto:   []string{"BTC", "ETH", "SOL"},
		},
		Trend: TrendConfig{
			RSIPeriod: 14,
			EMAFast:   10,
			EMASlow:   20,
		},
		Valuation: ValuationConfig{
			RSIPeriod:   14,
			Oversold:    30,
			Undervalued: 45,
			FairLow:     45,
			FairHigh:    55,
			Overvalued:  70,
			Overbought:  70,
		},
		Cache: CacheConfig{
			IntradayTTL: 60 * time.Second,
			DailyTTL:    1 * time.Hour,
			MacroTTL:    6 * time.Hour,
			NewsTTL:     5 * time.Minute,
			CryptoTTL:   5 * time.Minute,
		},
		Theme: ThemeConfig{
			Style: "default",
		},

		Commodities: CommoditiesConfig{
			Watchlist: []string{
				"WTI", "BRENT", "NATURAL_GAS", "COPPER", "ALUMINUM",
				"WHEAT", "CORN", "COFFEE", "SUGAR", "COTTON",
			},
			Interval: "daily",
		},
	}
}

// Load reads and parses configuration file at path, applies environment variable
// overrides, and validates configuration. The config file is optional; if it doesn't
// exist, defaults are used. Validation will fail only if required values (like API key)
// are missing from both the config file and environment variables.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// Try to read config file, but don't fail if it doesn't exist
	data, err := os.ReadFile(path)
	if err == nil {
		// File exists, parse it
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config YAML: %w", err)
		}
	}
	// If file doesn't exist, continue with defaults (os.IsNotExist(err))

	// Apply environment variable overrides
	if key := os.Getenv(envAPIKey); key != "" {
		cfg.API.Key = key
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that all required fields are set and that all values are valid.
func (c *Config) Validate() error {
	if err := validateAPIConfig(c.API); err != nil {
		return err
	}
	if err := validateWatchlistConfig(c.Watchlist); err != nil {
		return err
	}
	if err := validateTrendConfig(c.Trend); err != nil {
		return err
	}
	if err := validateValuationConfig(c.Valuation); err != nil {
		return err
	}
	if err := validateCommoditiesConfig(c.Commodities); err != nil {
		return err
	}
	if err := validateCacheConfig(c.Cache); err != nil {
		return err
	}
	if err := validateThemeConfig(c.Theme); err != nil {
		return err
	}
	return nil
}

// validateAPIConfig validates the API configuration section.
func validateAPIConfig(api APIConfig) error {
	if api.Key == "" {
		return fmt.Errorf("API key is required (set api.key in config.yaml or FINTERM_AV_API_KEY env var)")
	}
	if api.BaseURL == "" {
		return fmt.Errorf("api.base_url is required")
	}
	if api.RateLimit <= 0 {
		return fmt.Errorf("api.rate_limit must be positive, got %d", api.RateLimit)
	}
	if api.BurstLimit <= 0 {
		return fmt.Errorf("api.burst_limit must be positive, got %d", api.BurstLimit)
	}
	if api.Timeout <= 0 {
		return fmt.Errorf("api.timeout must be positive, got %v", api.Timeout)
	}
	if api.MaxRetries < 0 {
		return fmt.Errorf("api.max_retries must be non-negative, got %d", api.MaxRetries)
	}
	return nil
}

// validateWatchlistConfig validates the watchlist configuration section.
func validateWatchlistConfig(watchlist WatchlistConfig) error {
	for _, ticker := range watchlist.Equities {
		if err := validateTicker(ticker); err != nil {
			return fmt.Errorf("watchlist.equities: invalid ticker %q: %w", ticker, err)
		}
	}
	for _, ticker := range watchlist.Crypto {
		if err := validateTicker(ticker); err != nil {
			return fmt.Errorf("watchlist.crypto: invalid ticker %q: %w", ticker, err)
		}
	}
	return nil
}

// validateTrendConfig validates the trend configuration section.
func validateTrendConfig(trend TrendConfig) error {
	if trend.RSIPeriod <= 0 {
		return fmt.Errorf("trend.rsi_period must be positive, got %d", trend.RSIPeriod)
	}
	if trend.EMAFast <= 0 {
		return fmt.Errorf("trend.ema_fast must be positive, got %d", trend.EMAFast)
	}
	if trend.EMASlow <= 0 {
		return fmt.Errorf("trend.ema_slow must be positive, got %d", trend.EMASlow)
	}
	if trend.EMAFast >= trend.EMASlow {
		return fmt.Errorf("trend.ema_fast (%d) must be less than ema_slow (%d)", trend.EMAFast, trend.EMASlow)
	}
	return nil
}

// validateValuationConfig validates the valuation configuration section.
func validateValuationConfig(valuation ValuationConfig) error {
	if valuation.RSIPeriod <= 0 {
		return fmt.Errorf("valuation.rsi_period must be positive, got %d", valuation.RSIPeriod)
	}

	// Validate RSI thresholds (must be 0-100)
	rsiThresholds := []struct {
		name  string
		value int
	}{
		{"valuation.oversold", valuation.Oversold},
		{"valuation.undervalued", valuation.Undervalued},
		{"valuation.fair_low", valuation.FairLow},
		{"valuation.fair_high", valuation.FairHigh},
		{"valuation.overvalued", valuation.Overvalued},
		{"valuation.overbought", valuation.Overbought},
	}

	for _, threshold := range rsiThresholds {
		if threshold.value < 0 || threshold.value > 100 {
			return fmt.Errorf("%s must be between 0 and 100, got %d", threshold.name, threshold.value)
		}
	}

	return nil
}

// validateCommoditiesConfig validates the commodities configuration section.
func validateCommoditiesConfig(commodities CommoditiesConfig) error {
	// Interval validation (optional field, defaults to "daily" if empty)
	if commodities.Interval != "" {
		normalizedInterval := strings.ToLower(commodities.Interval)
		validIntervals := map[string]bool{
			"daily":     true,
			"weekly":    true,
			"monthly":   true,
			"quarterly": true,
			"annual":    true,
		}
		if !validIntervals[normalizedInterval] {
			return fmt.Errorf("commodities.interval must be one of: daily, weekly, monthly, quarterly, annual, got %q", commodities.Interval)
		}
	}

	return nil
}

// validateCacheConfig validates the cache configuration section.
func validateCacheConfig(cache CacheConfig) error {
	ttlFields := []struct {
		name  string
		value time.Duration
	}{
		{"cache.intraday_ttl", cache.IntradayTTL},
		{"cache.daily_ttl", cache.DailyTTL},
		{"cache.macro_ttl", cache.MacroTTL},
		{"cache.news_ttl", cache.NewsTTL},
		{"cache.crypto_ttl", cache.CryptoTTL},
	}

	for _, field := range ttlFields {
		if field.value <= 0 {
			return fmt.Errorf("%s must be positive, got %v", field.name, field.value)
		}
	}

	return nil
}

// validateThemeConfig validates the theme configuration section.
func validateThemeConfig(theme ThemeConfig) error {
	switch theme.Style {
	case "", "default", "minimal", "colorblind":
		// Valid
		return nil
	default:
		return fmt.Errorf("theme.style must be one of: default, minimal, colorblind, got %q", theme.Style)
	}
}

// validateTicker checks that a ticker symbol is valid (alphanumeric, dots, dashes, max 10 chars).
func validateTicker(ticker string) error {
	if ticker == "" {
		return fmt.Errorf("ticker cannot be empty")
	}

	if len(ticker) > 10 {
		return fmt.Errorf("ticker exceeds maximum length of 10 characters (got %d)", len(ticker))
	}

	// Allow: A-Z, a-z, 0-9, dots, dashes
	matched, err := regexp.MatchString(`^[A-Za-z0-9.-]+$`, ticker)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !matched {
		return fmt.Errorf("ticker contains invalid characters (only alphanumeric, dot, and dash allowed)")
	}

	return nil
}
