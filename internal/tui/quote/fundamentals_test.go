// Package quote provides the single ticker quote TUI view.
package quote

import (
	"context"
	"testing"
	"time"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
)

// TestNextEarningsDate tests that NextEarningsDate returns the correct future earnings date.
func TestNextEarningsDate(t *testing.T) {
	t.Run("returns first future earnings date", func(t *testing.T) {
		tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
		nextMonth := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02")

		data := &FundamentalsData{
			Earnings: &alphavantage.Earnings{
				Quarterly: []alphavantage.QuarterlyEarnings{
					{ReportedDate: "2023-01-15"}, // Past
					{ReportedDate: tomorrow},     // Future (first)
					{ReportedDate: nextMonth},    // Future (second)
					{ReportedDate: "2022-01-15"}, // Past
				},
			},
		}

		result := data.NextEarningsDate()
		if result != tomorrow {
			t.Errorf("expected %s, got %s", tomorrow, result)
		}
	})

	t.Run("returns empty when all past", func(t *testing.T) {
		data := &FundamentalsData{
			Earnings: &alphavantage.Earnings{
				Quarterly: []alphavantage.QuarterlyEarnings{
					{ReportedDate: "2023-01-15"},
					{ReportedDate: "2022-01-15"},
				},
			},
		}

		result := data.NextEarningsDate()
		if result != "" {
			t.Errorf("expected empty, got %s", result)
		}
	})

	t.Run("returns empty when nil earnings", func(t *testing.T) {
		data := &FundamentalsData{
			Earnings: nil,
		}

		result := data.NextEarningsDate()
		if result != "" {
			t.Errorf("expected empty, got %s", result)
		}
	})

	t.Run("returns empty when empty quarterly", func(t *testing.T) {
		data := &FundamentalsData{
			Earnings: &alphavantage.Earnings{
				Quarterly: []alphavantage.QuarterlyEarnings{},
			},
		}

		result := data.NextEarningsDate()
		if result != "" {
			t.Errorf("expected empty, got %s", result)
		}
	})
}

// TestFormatNumber tests the formatNumber function.
func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "large number with commas",
			input:    "1234567890",
			expected: "1,234,567,890",
		},
		{
			name:     "medium number",
			input:    "98765",
			expected: "98,765",
		},
		{
			name:     "small number",
			input:    "123",
			expected: "123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "—",
		},
		{
			name:     "None",
			input:    "None",
			expected: "—",
		},
		{
			name:     "dash",
			input:    "-",
			expected: "—",
		},
		{
			name:     "zero",
			input:    "0",
			expected: "—",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestFormatPriceFundamentals tests the formatPriceFundamentals function.
func TestFormatPriceFundamentals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal price",
			input:    "123.456",
			expected: "$123.46",
		},
		{
			name:     "large price",
			input:    "1000.789",
			expected: "$1000.79",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "—",
		},
		{
			name:     "None",
			input:    "None",
			expected: "—",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPriceFundamentals(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestFormatMarketCap tests the formatMarketCap function.
func TestFormatMarketCap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trillions",
			input:    "2848000000000",
			expected: "$2.85T",
		},
		{
			name:     "billions",
			input:    "1500000000",
			expected: "$1.50B",
		},
		{
			name:     "millions",
			input:    "1500000",
			expected: "$1.50M",
		},
		{
			name:     "thousands",
			input:    "50000",
			expected: "$50000",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "—",
		},
		{
			name:     "None",
			input:    "None",
			expected: "—",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMarketCap(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestFormatPercentFundamentals tests the formatPercentFundamentals function.
func TestFormatPercentFundamentals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "positive percent",
			input:    "3.456",
			expected: "3.46%",
		},
		{
			name:     "with percent sign",
			input:    "5.678%",
			expected: "5.68%",
		},
		{
			name:     "zero",
			input:    "0",
			expected: "0.00%",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "—",
		},
		{
			name:     "None",
			input:    "None",
			expected: "—",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPercentFundamentals(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestQuoteModel_FToggleOnEquityFetches tests that F on equity fetches fundamentals.
func TestQuoteModel_FToggleOnEquityFetches(t *testing.T) {
	model := NewModel()
	detector := indicators.NewAssetClassDetector([]string{"BTC", "ETH"}, nil)
	model.Configure(context.Background(), nil, nil, detector)

	// Set up quote data for equity
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
	}

	// Toggle F on equity
	newModelInterface, cmd := model.handleFToggle()
	newModel := newModelInterface.(Model)

	// Check that fundamentalsVisible is set to true
	if !newModel.fundamentalsVisible {
		t.Error("expected fundamentalsVisible to be true")
	}

	// Check that a command was returned to fetch fundamentals
	if cmd == nil {
		t.Error("expected fetch command to be returned")
	}
}

// TestQuoteModel_FToggleOnCryptoNoOp tests that F on crypto shows chip but doesn't fetch.
func TestQuoteModel_FToggleOnCryptoNoOp(t *testing.T) {
	model := NewModel()
	detector := indicators.NewAssetClassDetector([]string{"BTC", "ETH"}, nil)
	model.Configure(context.Background(), nil, nil, detector)

	// Set up quote data for crypto
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "BTC"},
	}

	// Toggle F on crypto (first press)
	newModel, cmd := model.handleFToggle()

	// Check that fundamentalsVisible is still false
	if newModel.(Model).fundamentalsVisible {
		t.Error("expected fundamentalsVisible to remain false for crypto")
	}

	// Check that no fetch command was returned
	if cmd != nil {
		t.Error("expected no fetch command for crypto")
	}

	// Check that crypto chip was shown
	if !newModel.(Model).cryptoChipShown {
		t.Error("expected crypto chip to be shown")
	}
}

// TestQuoteModel_FSecondPressUsesCache tests that second F press uses cached data.
func TestQuoteModel_FSecondPressUsesCache(t *testing.T) {
	model := NewModel()
	detector := indicators.NewAssetClassDetector([]string{"BTC", "ETH"}, nil)
	model.Configure(context.Background(), nil, nil, detector)

	// Set up quote data and cached fundamentals
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
	}
	model.fundamentalsData = &FundamentalsData{
		Overview: &alphavantage.CompanyOverview{Name: "Test Company"},
	}
	model.fundamentalsVisible = false

	// Toggle F to show (should use cache)
	newModel, cmd := model.handleFToggle()

	// Check that fundamentalsVisible is now true
	if !newModel.(Model).fundamentalsVisible {
		t.Error("expected fundamentalsVisible to be true")
	}

	// Check that no fetch command was returned (data is cached)
	if cmd != nil {
		t.Error("expected no fetch command when data is cached")
	}
}

// TestQuoteModel_FundamentalsHiddenByDefault tests that fundamentals is hidden by default.
func TestQuoteModel_FundamentalsHiddenByDefault(t *testing.T) {
	model := NewModel()

	// Check that fundamentalsVisible is false by default
	if model.fundamentalsVisible {
		t.Error("expected fundamentalsVisible to be false by default")
	}

	// Check that fundamentalsData is nil by default
	if model.fundamentalsData != nil {
		t.Error("expected fundamentalsData to be nil by default")
	}
}

// TestQuoteModel_CryptoChipShownOnce tests that crypto chip is only shown once.
func TestQuoteModel_CryptoChipShownOnce(t *testing.T) {
	model := NewModel()
	detector := indicators.NewAssetClassDetector([]string{"BTC", "ETH"}, nil)
	model.Configure(context.Background(), nil, nil, detector)

	// Set up quote data for crypto
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "BTC"},
	}

	// First F press
	modelInterface, _ := model.handleFToggle()
	model = modelInterface.(Model)
	if !model.cryptoChipShown {
		t.Error("expected crypto chip to be shown on first press")
	}

	// Second F press
	modelInterface2, _ := model.handleFToggle()
	model = modelInterface2.(Model)
	// Crypto chip should still be true (not toggled off)
	if !model.cryptoChipShown {
		t.Error("expected crypto chip to remain true on second press")
	}
}

// TestQuoteModel_NextEarningsFromQuarterly tests that next earnings is found.
func TestQuoteModel_NextEarningsFromQuarterly(t *testing.T) {
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	nextMonth := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02")

	data := &FundamentalsData{
		Earnings: &alphavantage.Earnings{
			Quarterly: []alphavantage.QuarterlyEarnings{
				{ReportedDate: "2023-01-15"},
				{ReportedDate: tomorrow},
				{ReportedDate: nextMonth},
				{ReportedDate: "2022-01-15"},
			},
		},
	}

	result := data.NextEarningsDate()
	if result != tomorrow {
		t.Errorf("expected %s, got %s", tomorrow, result)
	}
}

// TestQuoteModel_NextEarningsEmptyWhenAllPast tests that empty is returned when all past.
func TestQuoteModel_NextEarningsEmptyWhenAllPast(t *testing.T) {
	data := &FundamentalsData{
		Earnings: &alphavantage.Earnings{
			Quarterly: []alphavantage.QuarterlyEarnings{
				{ReportedDate: "2023-01-15"},
				{ReportedDate: "2022-01-15"},
				{ReportedDate: "2021-01-15"},
			},
		},
	}

	result := data.NextEarningsDate()
	if result != "" {
		t.Errorf("expected empty, got %s", result)
	}
}

// TestQuoteModel_KeyBindingsIncludesF tests that F is in key bindings.
func TestQuoteModel_KeyBindingsIncludesF(t *testing.T) {
	model := NewModel()
	bindings := model.KeyBindings()

	found := false
	for _, b := range bindings {
		if b.Key == "F" {
			found = true
			if b.Description != "Toggle fundamentals" {
				t.Errorf("expected description 'Toggle fundamentals', got %s", b.Description)
			}
			break
		}
	}

	if !found {
		t.Error("expected F key to be in bindings")
	}
}

// BenchmarkTestQuoteModel_FundamentalsFromCacheUnder50ms benchmarks cache access.
func BenchmarkTestQuoteModel_FundamentalsFromCacheUnder50ms(b *testing.B) {
	model := NewModel()

	// Pre-cache fundamentals data
	model.fundamentalsData = &FundamentalsData{
		Overview: &alphavantage.CompanyOverview{
			Name:    "Apple Inc.",
			Symbol:  "AAPL",
			Sector:  "Technology",
			PERatio: "29.42",
			EPS:     "6.16",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		modelInterface, _ := model.handleFToggle()
		model = modelInterface.(Model)
		_ = model.fundamentalsData
	}
}
