// Package chart provides tests for the chart model.
package chart

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/cache"
	"github.com/shinsekai/finterm/internal/config"
	trenddomain "github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
)

// mockCryptoFetcher is a mock crypto fetcher.
type mockCryptoFetcher struct {
	data []indicators.OHLCV
	err  error
}

func (m *mockCryptoFetcher) FetchCryptoOHLCV(_ context.Context, _ string) ([]indicators.OHLCV, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// TestChartModel_InitLoadsDefaultSymbol tests that the model initializes with the first symbol.
func TestChartModel_InitLoadsDefaultSymbol(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.Watchlist.Equities = []string{"AAPL", "MSFT"}
	cfg.Watchlist.Crypto = []string{"BTC"}
	detector := indicators.NewAssetClassDetector(cfg.Watchlist.Crypto)
	cacheStore := cache.New()

	engine := trenddomain.New(nil, nil, nil, nil, cfg, detector, nil, nil, cacheStore)

	avClient := &alphavantage.Client{} // Not actually used
	cryptoFetcher := &mockCryptoFetcher{
		data: []indicators.OHLCV{
			{Date: time.Now().Add(-24 * time.Hour), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		},
	}

	model := NewModel()
	model.Configure(ctx, engine, avClient, cacheStore, &cfg.Watchlist, detector, cfg, cryptoFetcher)

	// Verify initial symbol
	if model.symbol != "AAPL" {
		t.Errorf("Expected initial symbol AAPL, got %s", model.symbol)
	}
}

// TestChartModel_SwitchTickerJK tests cycling through tickers with j/k keys.
func TestChartModel_SwitchTickerJK(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.Watchlist.Equities = []string{"AAPL", "MSFT"}
	cfg.Watchlist.Crypto = []string{"BTC"}
	detector := indicators.NewAssetClassDetector(cfg.Watchlist.Crypto)
	cacheStore := cache.New()

	engine := trenddomain.New(nil, nil, nil, nil, cfg, detector, nil, nil, cacheStore)
	avClient := &alphavantage.Client{}
	cryptoFetcher := &mockCryptoFetcher{
		data: []indicators.OHLCV{
			{Date: time.Now().Add(-24 * time.Hour), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		},
	}

	m := NewModel()
	m.Configure(ctx, engine, avClient, cacheStore, &cfg.Watchlist, detector, cfg, cryptoFetcher)

	// Test k (previous) - should wrap to BTC
	initialSymbol := m.symbol
	newModel, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	result := newModel.(Model)
	m = &result

	// Execute the command to get ChangeSymbolMsg
	if cmd != nil {
		msg := cmd()
		newModel, _ = m.Update(msg)
		result = newModel.(Model)
		m = &result
	}

	if m.symbol == initialSymbol {
		t.Error("Expected symbol to change with 'k' key")
	}

	// Test j (next) - should wrap back to AAPL
	newModel, cmd = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	result = newModel.(Model)
	m = &result

	// Execute the command to get ChangeSymbolMsg
	if cmd != nil {
		msg := cmd()
		newModel, _ = m.Update(msg)
		result = newModel.(Model)
		m = &result
	}

	if m.symbol == "BTC" {
		t.Errorf("Expected symbol to be AAPL, got %s", m.symbol)
	}
}

// TestChartModel_SwitchTickerWrapsAtEnds tests that ticker selection wraps around the watchlist.
func TestChartModel_SwitchTickerWrapsAtEnds(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.Watchlist.Equities = []string{"AAPL"}
	cfg.Watchlist.Crypto = []string{"BTC"}
	detector := indicators.NewAssetClassDetector(cfg.Watchlist.Crypto)
	cacheStore := cache.New()

	engine := trenddomain.New(nil, nil, nil, nil, cfg, detector, nil, nil, cacheStore)
	avClient := &alphavantage.Client{}
	cryptoFetcher := &mockCryptoFetcher{
		data: []indicators.OHLCV{
			{Date: time.Now().Add(-24 * time.Hour), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		},
	}

	m := NewModel()
	m.Configure(ctx, engine, avClient, cacheStore, &cfg.Watchlist, detector, cfg, cryptoFetcher)

	// Start with AAPL
	m.symbol = "AAPL"

	// Move to next (j) -> should be BTC
	newModel, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	result := newModel.(Model)
	m = &result

	// Execute the command to get ChangeSymbolMsg
	if cmd != nil {
		msg := cmd()
		newModel, _ = m.Update(msg)
		result = newModel.(Model)
		m = &result
	}

	if m.symbol != "BTC" {
		t.Errorf("Expected BTC after j from AAPL, got %s", m.symbol)
	}

	// Move to next (j) -> should wrap to AAPL
	newModel, cmd = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	result = newModel.(Model)
	m = &result

	// Execute the command to get ChangeSymbolMsg
	if cmd != nil {
		msg := cmd()
		newModel, _ = m.Update(msg)
		result = newModel.(Model)
		m = &result
	}

	if m.symbol != "AAPL" {
		t.Errorf("Expected AAPL after j from BTC (wrap), got %s", m.symbol)
	}
}

// TestChartModel_TimeframeCycle1234 tests switching between timeframes with 1-4 keys.
func TestChartModel_TimeframeCycle1234(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.Watchlist.Equities = []string{"AAPL"}
	cfg.Watchlist.Crypto = []string{}
	detector := indicators.NewAssetClassDetector(cfg.Watchlist.Crypto)
	cacheStore := cache.New()

	engine := trenddomain.New(nil, nil, nil, nil, cfg, detector, nil, nil, cacheStore)
	avClient := &alphavantage.Client{}
	cryptoFetcher := &mockCryptoFetcher{
		data: []indicators.OHLCV{
			{Date: time.Now().Add(-24 * time.Hour), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		},
	}

	m := NewModel()
	m.Configure(ctx, engine, avClient, cacheStore, &cfg.Watchlist, detector, cfg, cryptoFetcher)

	initialTimeframe := m.timeframe

	// Test 1 (intraday)
	newModel, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	result := newModel.(Model)
	m = &result
	if m.timeframe != TimeframeIntraday {
		t.Errorf("Expected timeframe Intraday, got %v", m.timeframe)
	}

	// Test 2 (daily)
	newModel, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	result = newModel.(Model)
	m = &result
	if m.timeframe != TimeframeDaily {
		t.Errorf("Expected timeframe Daily, got %v", m.timeframe)
	}

	// Test 3 (weekly)
	newModel, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	result = newModel.(Model)
	m = &result
	if m.timeframe != TimeframeWeekly {
		t.Errorf("Expected timeframe Weekly, got %v", m.timeframe)
	}

	// Test 4 (monthly)
	newModel, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	result = newModel.(Model)
	m = &result
	if m.timeframe != TimeframeMonthly {
		t.Errorf("Expected timeframe Monthly, got %v", m.timeframe)
	}

	// Reset and test intraday rejection for crypto
	m.timeframe = initialTimeframe
	m.symbol = "BTC" // Switch to crypto
	cfg.Watchlist.Equities = []string{}
	cfg.Watchlist.Crypto = []string{"BTC"}
	detector = indicators.NewAssetClassDetector(cfg.Watchlist.Crypto)
	m.Configure(ctx, engine, avClient, cacheStore, &cfg.Watchlist, detector, cfg, cryptoFetcher)

	// Test 1 (intraday) on crypto - should be rejected
	newModel, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	result = newModel.(Model)
	m = &result
	if cmd != nil {
		t.Error("Expected no command for intraday on crypto, got command")
	}
	if m.timeframe == TimeframeIntraday {
		t.Error("Expected timeframe to not change to Intraday for crypto")
	}
}

// TestChartModel_ZoomPlusMinus tests zooming in/out with +/- keys.
func TestChartModel_ZoomPlusMinus(t *testing.T) {
	m := NewModel()

	// Start with default window (110)
	if m.window != 110 {
		t.Errorf("Expected initial window 110, got %d", m.window)
	}

	// Test + (zoom in) from 110 - should stay at 110 (max)
	m.zoomIn()
	if m.window != 110 {
		t.Errorf("Expected window 110 at max, got %d", m.window)
	}

	// Test - (zoom out) from 110 -> should go to 60
	m.zoomOut()
	if m.window != 60 {
		t.Errorf("Expected window 60 after zoom out from 110, got %d", m.window)
	}

	// Test - (zoom out) from 60 -> should go to 30
	m.zoomOut()
	if m.window != 30 {
		t.Errorf("Expected window 30 after zoom out from 60, got %d", m.window)
	}

	// Test - (zoom out) from 30 - should stay at 30 (min)
	m.zoomOut()
	if m.window != 30 {
		t.Errorf("Expected window 30 at min, got %d", m.window)
	}

	// Test + (zoom in) from 30 -> should go to 60
	m.zoomIn()
	if m.window != 60 {
		t.Errorf("Expected window 60 after zoom in from 30, got %d", m.window)
	}

	// Test + (zoom in) from 60 -> should go to 110
	m.zoomIn()
	if m.window != 110 {
		t.Errorf("Expected window 110 after zoom in from 60, got %d", m.window)
	}
}

// TestChartModel_ZoomAtBoundsClamped tests that zoom is clamped at bounds.
func TestChartModel_ZoomAtBoundsClamped(t *testing.T) {
	m := NewModel()

	// Test upper bound
	m.window = 110
	m.zoomIn()
	if m.window != 110 {
		t.Errorf("Expected window to be clamped at 110 (max), got %d", m.window)
	}

	// Test lower bound
	m.window = 30
	m.zoomOut()
	if m.window != 30 {
		t.Errorf("Expected window to be clamped at 30 (min), got %d", m.window)
	}
}

// TestChartModel_PanHL tests panning left/right with h/l keys.
func TestChartModel_PanHL(t *testing.T) {
	m := NewModel()
	m.data.Bars = make([]indicators.OHLCV, 20)
	for i := range m.data.Bars {
		m.data.Bars[i] = indicators.OHLCV{
			Date:   time.Now().Add(-time.Duration(20-i) * 24 * time.Hour),
			Open:   100 + float64(i),
			High:   110 + float64(i),
			Low:    90 + float64(i),
			Close:  105 + float64(i),
			Volume: 1000,
		}
	}
	m.window = 10
	m.offset = 5

	// Test h (pan left) by 10% of window (1 bar)
	initialOffset := m.offset
	m.panLeft()
	expectedOffset := initialOffset - 1
	if m.offset != expectedOffset {
		t.Errorf("Expected offset %d after pan left, got %d", expectedOffset, m.offset)
	}

	// Test l (pan right) by 10% of window (1 bar)
	m.panRight()
	expectedOffset = initialOffset
	if m.offset != expectedOffset {
		t.Errorf("Expected offset %d after pan right, got %d", expectedOffset, m.offset)
	}
}

// TestChartModel_RefreshReissuesFetch tests that r triggers a new fetch.
func TestChartModel_RefreshReissuesFetch(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.Watchlist.Equities = []string{"AAPL"}
	detector := indicators.NewAssetClassDetector(cfg.Watchlist.Crypto)
	cacheStore := cache.New()

	engine := trenddomain.New(nil, nil, nil, nil, cfg, detector, nil, nil, cacheStore)
	avClient := &alphavantage.Client{}
	cryptoFetcher := &mockCryptoFetcher{
		data: []indicators.OHLCV{
			{Date: time.Now().Add(-24 * time.Hour), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		},
	}

	m := NewModel()
	m.Configure(ctx, engine, avClient, cacheStore, &cfg.Watchlist, detector, cfg, cryptoFetcher)
	m.symbol = "AAPL"

	// Test r key triggers refresh
	_, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Error("Expected command on refresh, got nil")
	}

	// Verify the command type (should return a message, but we just check it's not nil)
}

// TestChartModel_KeyBindings returns the correct keybindings.
func TestChartModel_KeyBindings(t *testing.T) {
	m := NewModel()
	bindings := m.KeyBindings()

	expectedKeys := []string{"j/k", "1", "2", "3", "4", "+/-", "h/l", "r"}
	if len(bindings) != len(expectedKeys) {
		t.Errorf("Expected %d bindings, got %d", len(expectedKeys), len(bindings))
	}

	for i, expectedKey := range expectedKeys {
		if bindings[i].Key != expectedKey {
			t.Errorf("Expected binding %d key %s, got %s", i, expectedKey, bindings[i].Key)
		}
	}
}
