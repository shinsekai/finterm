// Package components provides reusable UI components for finterm.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/shinsekai/finterm/internal/alphavantage"
)

// ThemeInterface provides the theme methods needed by market status rendering.
// This interface breaks the import cycle between tui and components packages.
type ThemeInterface interface {
	Name() string
	StatusOnline() lipgloss.Style
	StatusOffline() lipgloss.Style
	Muted() lipgloss.Style
}

// RenderMarketStatus renders a single-line strip of market open/closed dots.
// Filters out cryptocurrency entries (trade 24/7 — dot conveys no information).
// Groups duplicate entries by primary exchange.
// Returns empty string on nil input (loading state).
// Layout: "NYSE ● · NASDAQ ● · LSE ○ · TSE ○ · FX ●" with ● green for open, ○ gray for closed.
func RenderMarketStatus(status *alphavantage.MarketStatus, theme ThemeInterface) string {
	if status == nil {
		return ""
	}

	// Filter out cryptocurrency and group by primary exchange
	exchanges := make(map[string]bool)
	for _, market := range status.Markets {
		// Skip cryptocurrency - trade 24/7 so dot conveys no information
		if market.MarketType == "Cryptocurrency" {
			continue
		}

		// Parse primary exchanges (comma-separated)
		if market.PrimaryExchanges != "" {
			parts := strings.Split(market.PrimaryExchanges, ",")
			for _, part := range parts {
				exchange := strings.TrimSpace(part)
				if exchange != "" {
					exchanges[exchange] = market.CurrentStatus == "open"
				}
			}
		}
	}

	// If no exchanges after filtering (or only crypto), return empty
	if len(exchanges) == 0 {
		return ""
	}

	// Order exchanges for consistent display
	orderedExchanges := orderedExchangeList(exchanges)

	var parts []string
	for _, exchange := range orderedExchanges {
		isOpen := exchanges[exchange]
		var dot string
		if theme.Name() == "colorblind" {
			// Use glyph differentiation for colorblind theme
			// Filled circle for open, hollow circle for closed
			if isOpen {
				dot = "●"
			} else {
				dot = "○"
			}
		} else {
			// For default/minimal themes, use same glyphs with color
			if isOpen {
				dot = theme.StatusOnline().Render("●")
			} else {
				dot = theme.StatusOffline().Render("○")
			}
		}

		parts = append(parts, exchange+" "+dot)
	}

	// Join with separator
	separator := theme.Muted().Render(" · ")
	return strings.Join(parts, separator)
}

// orderedExchangeList returns only the selected exchanges in fixed order:
// NYSE, NASDAQ, London, Tokyo (JPX), Global. All other exchanges are filtered out.
func orderedExchangeList(exchanges map[string]bool) []string {
	// Only display these specific exchanges
	targetExchanges := []string{
		"NYSE",
		"NASDAQ",
		"London", // London Stock Exchange
		"JPX",    // Tokyo Stock Exchange
		"Global", // Global/FX
	}

	// Build result with exchanges that exist in the input
	result := []string{}
	for _, ex := range targetExchanges {
		if _, exists := exchanges[ex]; exists {
			result = append(result, ex)
		}
	}

	return result
}

// RenderMarketStatusLoading renders a spinner for loading state.
func RenderMarketStatusLoading(theme ThemeInterface) string {
	return theme.Muted().Render("⋯")
}

// RenderMarketStatusOffline renders offline state for API failure.
func RenderMarketStatusOffline(theme ThemeInterface) string {
	return theme.Muted().Render("markets: offline")
}
