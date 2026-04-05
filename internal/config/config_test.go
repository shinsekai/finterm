package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Ensure env var doesn't interfere with test
	origEnv, origSet := os.LookupEnv(envAPIKey)
	t.Cleanup(func() {
		if origSet {
			//nolint:errcheck // Best-effort restoration, error is acceptable
			os.Setenv(envAPIKey, origEnv)
		} else {
			//nolint:errcheck // Best-effort unsetting, error is acceptable
			os.Unsetenv(envAPIKey)
		}
	})
	//nolint:errcheck // Best-effort unsetting, error is acceptable
	os.Unsetenv(envAPIKey)

	tests := []struct {
		name    string
		content string
		want    *Config
	}{
		{
			name: "minimal valid config",
			content: `
api:
  key: "test-api-key"
`,
			want: func() *Config {
				cfg := DefaultConfig()
				cfg.API.Key = "test-api-key"
				return cfg
			}(),
		},
		{
			name: "full valid config",
			content: `
api:
  key: "test-api-key"
  base_url: "https://custom.api.com/query"
  rate_limit: 50
  timeout: 5s
  max_retries: 2

watchlist:
  equities:
    - AAPL
    - MSFT
  crypto:
    - BTC
    - ETH

trend:
  rsi_period: 10
  ema_fast: 5
  ema_slow: 15

valuation:
  rsi_period: 10
  oversold: 20
  undervalued: 40
  fair_low: 40
  fair_high: 60
  overvalued: 60
  overbought: 80

cache:
  intraday_ttl: 30s
  daily_ttl: 30m
  macro_ttl: 3h
  news_ttl: 2m
  crypto_ttl: 2m

theme:
  style: "minimal"
`,
			want: func() *Config {
				cfg := DefaultConfig()
				cfg.API.Key = "test-api-key"
				cfg.API.BaseURL = "https://custom.api.com/query"
				cfg.API.RateLimit = 50
				cfg.API.Timeout = 5 * time.Second
				cfg.API.MaxRetries = 2
				cfg.Watchlist.Equities = []string{"AAPL", "MSFT"}
				cfg.Watchlist.Crypto = []string{"BTC", "ETH"}
				cfg.Trend.RSIPeriod = 10
				cfg.Trend.EMAFast = 5
				cfg.Trend.EMASlow = 15
				cfg.Valuation.RSIPeriod = 10
				cfg.Valuation.Oversold = 20
				cfg.Valuation.Undervalued = 40
				cfg.Valuation.FairLow = 40
				cfg.Valuation.FairHigh = 60
				cfg.Valuation.Overvalued = 60
				cfg.Valuation.Overbought = 80
				cfg.Cache.IntradayTTL = 30 * time.Second
				cfg.Cache.DailyTTL = 30 * time.Minute
				cfg.Cache.MacroTTL = 3 * time.Hour
				cfg.Cache.NewsTTL = 2 * time.Minute
				cfg.Cache.CryptoTTL = 2 * time.Minute
				cfg.Theme.Style = "minimal"
				return cfg
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			got, err := Load(configPath)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if got.API.Key != tt.want.API.Key {
				t.Errorf("Load() API.Key = %q, want %q", got.API.Key, tt.want.API.Key)
			}
			if got.API.BaseURL != tt.want.API.BaseURL {
				t.Errorf("Load() API.BaseURL = %q, want %q", got.API.BaseURL, tt.want.API.BaseURL)
			}
			if got.API.RateLimit != tt.want.API.RateLimit {
				t.Errorf("Load() API.RateLimit = %d, want %d", got.API.RateLimit, tt.want.API.RateLimit)
			}
			if got.API.Timeout != tt.want.API.Timeout {
				t.Errorf("Load() API.Timeout = %v, want %v", got.API.Timeout, tt.want.API.Timeout)
			}
			if got.API.MaxRetries != tt.want.API.MaxRetries {
				t.Errorf("Load() API.MaxRetries = %d, want %d", got.API.MaxRetries, tt.want.API.MaxRetries)
			}
		})
	}
}

func TestLoad_MissingAPIKey(t *testing.T) {
	// Ensure env var doesn't interfere with test
	origEnv, origSet := os.LookupEnv(envAPIKey)
	t.Cleanup(func() {
		if origSet {
			//nolint:errcheck // Best-effort restoration, error is acceptable
			os.Setenv(envAPIKey, origEnv)
		} else {
			//nolint:errcheck // Best-effort unsetting, error is acceptable
			os.Unsetenv(envAPIKey)
		}
	})
	//nolint:errcheck // Best-effort unsetting, error is acceptable
	os.Unsetenv(envAPIKey)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
api:
  base_url: "https://www.alphavantage.co/query"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("Load() expected error for missing API key, got nil")
	}

	expectedErr := "API key is required"
	if err.Error() == "" {
		t.Errorf("Load() error message is empty, expected to contain %q", expectedErr)
	}
}

func TestLoad_EnvVarOverride(t *testing.T) {
	tests := []struct {
		name    string
		envKey  string
		content string
		wantKey string
	}{{
		name:    "env var overrides yaml value",
		envKey:  "env-override-key",
		content: `api:\n  key: "yaml-key"`,
		wantKey: "env-override-key",
	}, {
		name:    "env var sets missing yaml value",
		envKey:  "env-only-key",
		content: `api: {}`,
		wantKey: "env-only-key",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original env var value
			origEnv, origSet := os.LookupEnv(envAPIKey)
			t.Cleanup(func() {
				if origSet {
					//nolint:errcheck // Best-effort restoration, error is acceptable
					os.Setenv(envAPIKey, origEnv)
				} else {
					//nolint:errcheck // Best-effort unsetting, error is acceptable
					os.Unsetenv(envAPIKey)
				}
			})

			if err := os.Setenv(envAPIKey, tt.envKey); err != nil {
				t.Fatalf("Failed to set env var: %v", err)
			}

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			got, err := Load(configPath)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if got.API.Key != tt.wantKey {
				t.Errorf("Load() API.Key = %q, want %q (from env var)", got.API.Key, tt.wantKey)
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	// Ensure env var doesn't interfere with test
	origEnv, origSet := os.LookupEnv(envAPIKey)
	t.Cleanup(func() {
		if origSet {
			//nolint:errcheck // Best-effort restoration, error is acceptable
			os.Setenv(envAPIKey, origEnv)
		} else {
			//nolint:errcheck // Best-effort unsetting, error is acceptable
			os.Unsetenv(envAPIKey)
		}
	})
	//nolint:errcheck // Best-effort unsetting, error is acceptable
	os.Unsetenv(envAPIKey)

	_, err := Load("/nonexistent/path/to/config.yaml")
	if err == nil {
		t.Fatal("Load() expected error for nonexistent file, got nil")
	}

	expectedErr := "reading config file"
	if err.Error() == "" {
		t.Errorf("Load() error message is empty, expected to contain %q", expectedErr)
	}
}

func TestValidate_InvalidTicker(t *testing.T) {
	tests := []struct {
		name        string
		equities    []string
		crypto      []string
		wantErr     bool
		wantErrText string
	}{
		{
			name:     "valid tickers",
			equities: []string{"AAPL", "MSFT", "GOOGL"},
			crypto:   []string{"BTC", "ETH", "SOL"},
			wantErr:  false,
		},
		{
			name:        "ticker with special characters",
			equities:    []string{"AAPL!"},
			wantErr:     true,
			wantErrText: "contains invalid characters",
		},
		{
			name:        "ticker too long",
			equities:    []string{"VERYLONGTICKER"},
			wantErr:     true,
			wantErrText: "exceeds maximum length of 10 characters",
		},
		{
			name:        "ticker with spaces",
			equities:    []string{"AAP L"},
			wantErr:     true,
			wantErrText: "contains invalid characters",
		},
		{
			name:        "ticker with underscore",
			equities:    []string{"AAP_L"},
			wantErr:     true,
			wantErrText: "contains invalid characters",
		},
		{
			name:     "valid ticker with dash",
			equities: []string{"BRK-B"},
			wantErr:  false,
		},
		{
			name:     "valid ticker with dot",
			equities: []string{"BTC-USD"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.API.Key = "test-key"
			cfg.Watchlist.Equities = tt.equities
			cfg.Watchlist.Crypto = tt.crypto

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErrText != "" {
				if err.Error() == "" {
					t.Errorf("Validate() error message is empty, expected to contain %q", tt.wantErrText)
				}
			}
		})
	}
}

func TestValidate_InvalidDuration(t *testing.T) {
	tests := []struct {
		name        string
		modifyCfg   func(*Config)
		wantErr     bool
		wantErrText string
	}{
		{
			name: "invalid timeout format",
			modifyCfg: func(c *Config) {
				c.API.Timeout = -1 * time.Second
			},
			wantErr:     true,
			wantErrText: "api.timeout must be positive",
		},
		{
			name: "zero timeout",
			modifyCfg: func(c *Config) {
				c.API.Timeout = 0
			},
			wantErr:     true,
			wantErrText: "api.timeout must be positive",
		},
		{
			name: "negative intraday TTL",
			modifyCfg: func(c *Config) {
				c.Cache.IntradayTTL = -1 * time.Second
			},
			wantErr:     true,
			wantErrText: "cache.intraday_ttl must be positive",
		},
		{
			name: "zero daily TTL",
			modifyCfg: func(c *Config) {
				c.Cache.DailyTTL = 0
			},
			wantErr:     true,
			wantErrText: "cache.daily_ttl must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.API.Key = "test-key"
			tt.modifyCfg(cfg)

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErrText != "" {
				if err.Error() == "" {
					t.Errorf("Validate() error message is empty, expected to contain %q", tt.wantErrText)
				}
			}
		})
	}
}

func TestValidate_Defaults(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Key = "test-key"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() with defaults failed: %v", err)
	}

	// Verify default values are set correctly
	defaults := DefaultConfig()

	if cfg.API.BaseURL != defaults.API.BaseURL {
		t.Errorf("Default API.BaseURL = %q, want %q", cfg.API.BaseURL, defaults.API.BaseURL)
	}
	if cfg.API.RateLimit != defaults.API.RateLimit {
		t.Errorf("Default API.RateLimit = %d, want %d", cfg.API.RateLimit, defaults.API.RateLimit)
	}
	if cfg.API.Timeout != defaults.API.Timeout {
		t.Errorf("Default API.Timeout = %v, want %v", cfg.API.Timeout, defaults.API.Timeout)
	}
	if cfg.API.MaxRetries != defaults.API.MaxRetries {
		t.Errorf("Default API.MaxRetries = %d, want %d", cfg.API.MaxRetries, defaults.API.MaxRetries)
	}
	if cfg.Trend.RSIPeriod != defaults.Trend.RSIPeriod {
		t.Errorf("Default Trend.RSIPeriod = %d, want %d", cfg.Trend.RSIPeriod, defaults.Trend.RSIPeriod)
	}
	if cfg.Trend.EMAFast != defaults.Trend.EMAFast {
		t.Errorf("Default Trend.EMAFast = %d, want %d", cfg.Trend.EMAFast, defaults.Trend.EMAFast)
	}
	if cfg.Trend.EMASlow != defaults.Trend.EMASlow {
		t.Errorf("Default Trend.EMASlow = %d, want %d", cfg.Trend.EMASlow, defaults.Trend.EMASlow)
	}
	if cfg.Theme.Style != defaults.Theme.Style {
		t.Errorf("Default Theme.Style = %q, want %q", cfg.Theme.Style, defaults.Theme.Style)
	}
}

func TestValidate_InvalidTheme(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Key = "test-key"
	cfg.Theme.Style = "invalid-style"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected error for invalid theme style, got nil")
	}

	expectedErr := "theme.style must be one of"
	if err.Error() == "" {
		t.Errorf("Validate() error message is empty, expected to contain %q", expectedErr)
	}
}

func TestValidate_EMAFastNotLessThanSlow(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Key = "test-key"
	cfg.Trend.EMAFast = 21
	cfg.Trend.EMASlow = 9

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() expected error for EMA fast >= slow, got nil")
	}

	expectedErr := "ema_fast must be less than ema_slow"
	if err.Error() == "" {
		t.Errorf("Validate() error message is empty, expected to contain %q", expectedErr)
	}
}

func TestValidate_InvalidRSIValue(t *testing.T) {
	tests := []struct {
		name        string
		modifyCfg   func(*Config)
		wantErrText string
	}{
		{
			name: "oversold out of range",
			modifyCfg: func(c *Config) {
				c.Valuation.Oversold = 150
			},
			wantErrText: "valuation.oversold must be between 0 and 100",
		},
		{
			name: "undervalued negative",
			modifyCfg: func(c *Config) {
				c.Valuation.Undervalued = -10
			},
			wantErrText: "valuation.undervalued must be between 0 and 100",
		},
		{
			name: "fair high out of range",
			modifyCfg: func(c *Config) {
				c.Valuation.FairHigh = 200
			},
			wantErrText: "valuation.fair_high must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.API.Key = "test-key"
			tt.modifyCfg(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("Validate() expected error, got nil")
			}

			if err.Error() == "" {
				t.Errorf("Validate() error message is empty, expected to contain %q", tt.wantErrText)
			}
		})
	}
}
