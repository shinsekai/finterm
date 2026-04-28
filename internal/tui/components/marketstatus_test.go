// Package components provides tests for UI components.
package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/shinsekai/finterm/internal/alphavantage"
)

// marketStatusMockTheme is a mock implementation of ThemeInterface for testing.
type marketStatusMockTheme struct {
	name string
}

func (m *marketStatusMockTheme) Name() string {
	return m.name
}

func (m *marketStatusMockTheme) StatusOnline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
}

func (m *marketStatusMockTheme) StatusOffline() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("gray"))
}

func (m *marketStatusMockTheme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("gray"))
}

func marketStatusDefaultTheme() ThemeInterface {
	return &marketStatusMockTheme{name: "default"}
}

func marketStatusColorblindTheme() ThemeInterface {
	return &marketStatusMockTheme{name: "colorblind"}
}

func TestMarketStatusView_RendersOpenAndClosedDots(t *testing.T) {
	status := &alphavantage.MarketStatus{
		Markets: []alphavantage.MarketStatusEntry{
			{
				MarketType:       "Equity",
				PrimaryExchanges: "NASDAQ, NYSE",
				CurrentStatus:    "closed",
			},
			{
				MarketType:       "Forex",
				PrimaryExchanges: "Global",
				CurrentStatus:    "open",
			},
			{
				MarketType:       "Equity",
				Region:           "Japan",
				PrimaryExchanges: "JPX",
				CurrentStatus:    "open",
			},
			{
				MarketType:       "Equity",
				Region:           "United Kingdom",
				PrimaryExchanges: "London",
				CurrentStatus:    "open",
			},
		},
	}

	result := RenderMarketStatus(status, marketStatusDefaultTheme())

	assert.Contains(t, result, "NYSE", "Should contain NYSE")
	assert.Contains(t, result, "NASDAQ", "Should contain NASDAQ")
	assert.Contains(t, result, "Global", "Should contain Global")
	assert.Contains(t, result, "JPX", "Should contain JPX")
	assert.Contains(t, result, "London", "Should contain London")
	assert.Contains(t, result, "●", "Should contain open dots")
	assert.Contains(t, result, "○", "Should contain closed dots")
	assert.Contains(t, result, "·", "Should contain separator")
}

func TestMarketStatusView_ExcludesCrypto(t *testing.T) {
	status := &alphavantage.MarketStatus{
		Markets: []alphavantage.MarketStatusEntry{
			{
				MarketType:       "Cryptocurrency",
				PrimaryExchanges: "Global Crypto Exchanges",
				CurrentStatus:    "open",
			},
			{
				MarketType:       "Equity",
				PrimaryExchanges: "NYSE",
				CurrentStatus:    "open",
			},
		},
	}

	result := RenderMarketStatus(status, marketStatusDefaultTheme())

	assert.Contains(t, result, "NYSE", "Should contain NYSE")
	assert.NotContains(t, result, "Crypto", "Should not contain crypto references")
}

func TestMarketStatusView_GroupsByPrimaryExchange(t *testing.T) {
	status := &alphavantage.MarketStatus{
		Markets: []alphavantage.MarketStatusEntry{
			{
				MarketType:       "Equity",
				Region:           "United States",
				PrimaryExchanges: "NASDAQ, NYSE",
				CurrentStatus:    "open",
			},
		},
	}

	result := RenderMarketStatus(status, marketStatusDefaultTheme())

	assert.Contains(t, result, "NYSE", "Should contain NYSE")
	assert.Contains(t, result, "NASDAQ", "Should contain NASDAQ")
}

func TestMarketStatusView_NilInputRendersEmpty(t *testing.T) {
	result := RenderMarketStatus(nil, marketStatusDefaultTheme())

	assert.Equal(t, "", result, "Nil input should render empty string")
}

func TestMarketStatusView_OfflineStateRendered(t *testing.T) {
	result := RenderMarketStatusOffline(marketStatusDefaultTheme())

	assert.Contains(t, result, "markets: offline")
}

func TestMarketStatusView_ColorblindGlyphs(t *testing.T) {
	status := &alphavantage.MarketStatus{
		Markets: []alphavantage.MarketStatusEntry{
			{
				MarketType:       "Equity",
				PrimaryExchanges: "NASDAQ, NYSE",
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
				CurrentStatus:    "closed",
			},
		},
	}

	result := RenderMarketStatus(status, marketStatusColorblindTheme())

	assert.Contains(t, result, "NASDAQ", "Should contain NASDAQ")
	assert.Contains(t, result, "NYSE", "Should contain NYSE")
	assert.Contains(t, result, "JPX", "Should contain JPX")
	assert.Contains(t, result, "Global", "Should contain Global")
	assert.Contains(t, result, "London", "Should contain London")
	assert.Contains(t, result, "●", "Should contain filled dot for open")
	assert.Contains(t, result, "○", "Should contain hollow dot for closed")
}

func TestMarketStatusView_LoadingState(t *testing.T) {
	result := RenderMarketStatusLoading(marketStatusDefaultTheme())

	assert.Contains(t, result, "⋯")
}

func TestMarketStatusView_OnlyCryptoReturnsEmpty(t *testing.T) {
	status := &alphavantage.MarketStatus{
		Markets: []alphavantage.MarketStatusEntry{
			{
				MarketType:       "Cryptocurrency",
				PrimaryExchanges: "Global Crypto Exchanges",
				CurrentStatus:    "open",
			},
		},
	}

	result := RenderMarketStatus(status, marketStatusDefaultTheme())

	assert.Equal(t, "", result, "Only crypto entries should render empty")
}

func TestMarketStatusView_PriorityOrder(t *testing.T) {
	status := &alphavantage.MarketStatus{
		Markets: []alphavantage.MarketStatusEntry{
			{
				MarketType:       "Equity",
				Region:           "United States",
				PrimaryExchanges: "NASDAQ, NYSE",
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
			// These should be filtered out
			{
				MarketType:       "Equity",
				Region:           "Canada",
				PrimaryExchanges: "Toronto, Toronto Ventures",
				CurrentStatus:    "closed",
			},
		},
	}

	result := RenderMarketStatus(status, marketStatusDefaultTheme())

	// Verify only target exchanges are shown: NYSE, NASDAQ, London, JPX, Global
	assert.Contains(t, result, "NYSE", "Should find NYSE")
	assert.Contains(t, result, "NASDAQ", "Should find NASDAQ")
	assert.Contains(t, result, "JPX", "Should find JPX")
	assert.Contains(t, result, "Global", "Should find Global")
	assert.Contains(t, result, "London", "Should find London")
	assert.NotContains(t, result, "Toronto", "Should not filter in Toronto")

	// Verify fixed order: NYSE, NASDAQ, London, JPX, Global
	nyseIdx := strings.Index(result, "NYSE")
	nasdaqIdx := strings.Index(result, "NASDAQ")
	londonIdx := strings.Index(result, "London")
	jpxIdx := strings.Index(result, "JPX")
	globalIdx := strings.Index(result, "Global")

	assert.True(t, nyseIdx < nasdaqIdx, "NYSE should come before NASDAQ")
	assert.True(t, nasdaqIdx < londonIdx, "NASDAQ should come before London")
	assert.True(t, londonIdx < jpxIdx, "London should come before JPX")
	assert.True(t, jpxIdx < globalIdx, "JPX should come before Global")
}

func TestMarketStatusView_Forex(t *testing.T) {
	status := &alphavantage.MarketStatus{
		Markets: []alphavantage.MarketStatusEntry{
			{
				MarketType:       "Forex",
				PrimaryExchanges: "Global",
				CurrentStatus:    "open",
			},
		},
	}

	result := RenderMarketStatus(status, marketStatusDefaultTheme())

	assert.Contains(t, result, "Global", "Should contain Global exchange")
}
