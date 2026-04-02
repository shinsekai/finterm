// Package config provides configuration loading and validation for finterm.
package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	envAPIKey = "FINTERM_AV_API_KEY"
)

// Config holds the complete application configuration.
type Config struct {
	API       APIConfig       `yaml:"api"`
	Watchlist WatchlistConfig `yaml:"watchlist"`
	Trend     TrendConfig     `yaml:"trend"`
	Valuation ValuationConfig `yaml:"valuation"`
	Cache     CacheConfig     `yaml:"cache"`
	Theme     ThemeConfig     `yaml:"theme"`
}

// APIConfig holds Alpha Vantage API configuration.
type APIConfig struct {
	Key        string        `yaml:"key"`
	BaseURL    string        `yaml:"base_url"`
	RateLimit  int           `yaml:"rate_limit"`
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
	RSIPeriod int          `yaml:"rsi_period"`
	EMAFast   int          `yaml:"ema_fast"`
	EMASlow   int          `yaml:"ema_slow"`
	Scoring   TrendScoring `yaml:"scoring"`
}

// TrendScoring holds trend scoring thresholds.
type TrendScoring struct {
	BullishRSIMin  int `yaml:"bullish_rsi_min"`
	BullishRSIMax  int `yaml:"bullish_rsi_max"`
	BearishRSILow  int `yaml:"bearish_rsi_low"`
	BearishRSIHigh int `yaml:"bearish_rsi_high"`
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

// DefaultConfig returns a Config with default values applied.
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			BaseURL:    "https://www.alphavantage.co/query",
			RateLimit:  70,
			Timeout:    10 * time.Second,
			MaxRetries: 3,
		},
		Trend: TrendConfig{
			RSIPeriod: 14,
			EMAFast:   9,
			EMASlow:   21,
			Scoring: TrendScoring{
				BullishRSIMin:  40,
				BullishRSIMax:  70,
				BearishRSILow:  40,
				BearishRSIHigh: 80,
			},
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
	}
}

// Load reads and parses the configuration file at path, applies environment variable
// overrides, and validates the configuration.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}

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
	if c.API.Key == "" {
		return fmt.Errorf("API key is required (set api.key in config.yaml or FINTERM_AV_API_KEY env var)")
	}

	if c.API.BaseURL == "" {
		return fmt.Errorf("api.base_url is required")
	}

	if c.API.RateLimit <= 0 {
		return fmt.Errorf("api.rate_limit must be positive, got %d", c.API.RateLimit)
	}

	if c.API.Timeout <= 0 {
		return fmt.Errorf("api.timeout must be positive, got %v", c.API.Timeout)
	}

	if c.API.MaxRetries < 0 {
		return fmt.Errorf("api.max_retries must be non-negative, got %d", c.API.MaxRetries)
	}

	// Validate tickers
	for _, ticker := range c.Watchlist.Equities {
		if err := validateTicker(ticker); err != nil {
			return fmt.Errorf("watchlist.equities: invalid ticker %q: %w", ticker, err)
		}
	}

	for _, ticker := range c.Watchlist.Crypto {
		if err := validateTicker(ticker); err != nil {
			return fmt.Errorf("watchlist.crypto: invalid ticker %q: %w", ticker, err)
		}
	}

	// Validate trend config
	if c.Trend.RSIPeriod <= 0 {
		return fmt.Errorf("trend.rsi_period must be positive, got %d", c.Trend.RSIPeriod)
	}

	if c.Trend.EMAFast <= 0 {
		return fmt.Errorf("trend.ema_fast must be positive, got %d", c.Trend.EMAFast)
	}

	if c.Trend.EMASlow <= 0 {
		return fmt.Errorf("trend.ema_slow must be positive, got %d", c.Trend.EMASlow)
	}

	if c.Trend.EMAFast >= c.Trend.EMASlow {
		return fmt.Errorf("trend.ema_fast (%d) must be less than ema_slow (%d)", c.Trend.EMAFast, c.Trend.EMASlow)
	}

	// Validate valuation config
	if c.Valuation.RSIPeriod <= 0 {
		return fmt.Errorf("valuation.rsi_period must be positive, got %d", c.Valuation.RSIPeriod)
	}

	if c.Valuation.Oversold < 0 || c.Valuation.Oversold > 100 {
		return fmt.Errorf("valuation.oversold must be between 0 and 100, got %d", c.Valuation.Oversold)
	}

	if c.Valuation.Undervalued < 0 || c.Valuation.Undervalued > 100 {
		return fmt.Errorf("valuation.undervalued must be between 0 and 100, got %d", c.Valuation.Undervalued)
	}

	if c.Valuation.FairLow < 0 || c.Valuation.FairLow > 100 {
		return fmt.Errorf("valuation.fair_low must be between 0 and 100, got %d", c.Valuation.FairLow)
	}

	if c.Valuation.FairHigh < 0 || c.Valuation.FairHigh > 100 {
		return fmt.Errorf("valuation.fair_high must be between 0 and 100, got %d", c.Valuation.FairHigh)
	}

	if c.Valuation.Overvalued < 0 || c.Valuation.Overvalued > 100 {
		return fmt.Errorf("valuation.overvalued must be between 0 and 100, got %d", c.Valuation.Overvalued)
	}

	if c.Valuation.Overbought < 0 || c.Valuation.Overbought > 100 {
		return fmt.Errorf("valuation.overbought must be between 0 and 100, got %d", c.Valuation.Overbought)
	}

	// Validate cache TTLs
	if c.Cache.IntradayTTL <= 0 {
		return fmt.Errorf("cache.intraday_ttl must be positive, got %v", c.Cache.IntradayTTL)
	}

	if c.Cache.DailyTTL <= 0 {
		return fmt.Errorf("cache.daily_ttl must be positive, got %v", c.Cache.DailyTTL)
	}

	if c.Cache.MacroTTL <= 0 {
		return fmt.Errorf("cache.macro_ttl must be positive, got %v", c.Cache.MacroTTL)
	}

	if c.Cache.NewsTTL <= 0 {
		return fmt.Errorf("cache.news_ttl must be positive, got %v", c.Cache.NewsTTL)
	}

	if c.Cache.CryptoTTL <= 0 {
		return fmt.Errorf("cache.crypto_ttl must be positive, got %v", c.Cache.CryptoTTL)
	}

	// Validate theme
	switch c.Theme.Style {
	case "", "default", "minimal", "colorblind":
		// Valid
	default:
		return fmt.Errorf("theme.style must be one of: default, minimal, colorblind, got %q", c.Theme.Style)
	}

	return nil
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
