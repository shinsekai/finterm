// Package tui provides tests for the terminal user interface.
package tui

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/cache"
)

func TestApp_MarketStatusMessages(t *testing.T) {
	theme := NewTheme("default")
	app := Model{
		theme:               theme,
		marketStatusLoading: true,
	}

	// Test loading message
	msg := MarketStatusLoadedMsg{
		status: &alphavantage.MarketStatus{
			Markets: []alphavantage.MarketStatusEntry{
				{
					MarketType:       "Equity",
					PrimaryExchanges: "NYSE",
					CurrentStatus:    "open",
				},
			},
		},
	}
	updatedModel, _ := app.Update(msg)
	app = updatedModel.(Model)

	assert.False(t, app.marketStatusLoading, "marketStatusLoading should be false after loaded")
	assert.False(t, app.marketStatusFailed, "marketStatusFailed should be false after loaded")
	assert.NotNil(t, app.marketStatus, "marketStatus should be set")
	assert.Equal(t, "NYSE", app.marketStatus.Markets[0].PrimaryExchanges, "Should have correct market data")

	// Test failed message
	msg2 := MarketStatusFailedMsg{
		err: errors.New("API error"),
	}
	updatedModel, _ = app.Update(msg2)
	app = updatedModel.(Model)

	assert.True(t, app.marketStatusFailed, "marketStatusFailed should be true after error")
	assert.False(t, app.marketStatusLoading, "marketStatusLoading should be false after error")
}

func TestApp_MarketStatusRefreshTick(t *testing.T) {
	theme := NewTheme("default")
	avClient := &alphavantage.Client{}
	cacheStore := cache.New()

	app := Model{
		theme:      theme,
		avClient:   avClient,
		cacheStore: cacheStore,
		marketStatus: &alphavantage.MarketStatus{
			Markets: []alphavantage.MarketStatusEntry{
				{
					MarketType:       "Equity",
					PrimaryExchanges: "NASDAQ",
					CurrentStatus:    "closed",
				},
			},
		},
	}

	// Test refresh tick message
	_, cmd := app.Update(MarketStatusRefreshTickMsg{Time: time.Now()})

	// Should return a command to fetch market status
	assert.NotNil(t, cmd, "Should return a command to refresh market status")
}

func TestApp_MarketStatusCacheKey(t *testing.T) {
	assert.Equal(t, "market_status", cacheKeyMarketStatus, "Cache key should be 'market_status'")
}

func TestApp_MarketStatusShowsLoadingSpinner(t *testing.T) {
	theme := NewTheme("default")
	app := Model{
		theme:               theme,
		marketStatusLoading: true,
	}

	// Verify the status bar shows loading
	statusBar := app.renderStatusBar()

	assert.Contains(t, statusBar, "⋯", "Status bar should show loading spinner")
}

func TestApp_MarketStatusShowsOfflineOnFailure(t *testing.T) {
	theme := NewTheme("default")
	app := Model{
		theme:              theme,
		marketStatusFailed: true,
	}

	// Verify the status bar shows offline
	statusBar := app.renderStatusBar()

	assert.Contains(t, statusBar, "markets: offline", "Status bar should show offline message")
}

func TestApp_MarketStatusUpdatesStatusBar(t *testing.T) {
	theme := NewTheme("default")
	app := Model{
		theme: theme,
		marketStatus: &alphavantage.MarketStatus{
			Markets: []alphavantage.MarketStatusEntry{
				{
					MarketType:       "Equity",
					PrimaryExchanges: "NASDAQ, NYSE",
					CurrentStatus:    "open",
				},
			},
		},
	}

	// Verify the status bar renders the market status
	statusBar := app.renderStatusBar()

	assert.Contains(t, statusBar, "NASDAQ", "Status bar should contain NASDAQ")
	assert.Contains(t, statusBar, "NYSE", "Status bar should contain NYSE")
	assert.Contains(t, statusBar, "●", "Status bar should contain open indicator")
}

func TestApp_MarketStatusWithColorblindTheme(t *testing.T) {
	theme := NewTheme("colorblind")
	app := Model{
		theme: theme,
		marketStatus: &alphavantage.MarketStatus{
			Markets: []alphavantage.MarketStatusEntry{
				{
					MarketType:       "Equity",
					PrimaryExchanges: "NYSE",
					CurrentStatus:    "open",
				},
				{
					MarketType:       "Equity",
					Region:           "Japan",
					PrimaryExchanges: "JPX",
					CurrentStatus:    "closed",
				},
				{
					MarketType:       "Forex",
					PrimaryExchanges: "Global",
					CurrentStatus:    "open",
				},
				{
					MarketType:       "Equity",
					Region:           "United Kingdom",
					PrimaryExchanges: "London",
					CurrentStatus:    "open",
				},
			},
		},
	}

	// Verify the status bar renders correctly for colorblind theme
	statusBar := app.renderStatusBar()

	assert.Contains(t, statusBar, "NYSE", "Status bar should contain NYSE")
	assert.Contains(t, statusBar, "JPX", "Status bar should contain JPX")
	assert.Contains(t, statusBar, "Global", "Status bar should contain Global")
	assert.Contains(t, statusBar, "London", "Status bar should contain London")
	assert.Contains(t, statusBar, "●", "Status bar should contain filled dot for open")
	assert.Contains(t, statusBar, "○", "Status bar should contain hollow dot for closed")
}

func TestApp_MarketStatusWithOnlyCrypto(t *testing.T) {
	theme := NewTheme("default")
	app := Model{
		theme: theme,
		marketStatus: &alphavantage.MarketStatus{
			Markets: []alphavantage.MarketStatusEntry{
				{
					MarketType:       "Cryptocurrency",
					PrimaryExchanges: "Global Crypto Exchanges",
					CurrentStatus:    "open",
				},
			},
		},
	}

	// Verify the status bar does not show crypto markets
	statusBar := app.renderStatusBar()

	assert.NotContains(t, statusBar, "Crypto", "Status bar should not contain crypto references")
}

func TestApp_MarketStatusWithNilStatus(t *testing.T) {
	theme := NewTheme("default")
	app := Model{
		theme:               theme,
		marketStatus:        nil,
		marketStatusLoading: false,
		marketStatusFailed:  false,
	}

	// Verify the status bar handles nil status gracefully
	statusBar := app.renderStatusBar()

	// Should render without crashing
	assert.NotEmpty(t, statusBar, "Status bar should render even with nil status")
}
