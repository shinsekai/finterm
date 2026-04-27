// Package chart provides view rendering for the chart tab.
package chart

import (
	"fmt"
	"strings"

	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// renderView renders the chart view.
func renderView(m *Model) string {
	var builder strings.Builder

	// Render header with status chip
	builder.WriteString(m.renderHeader())
	builder.WriteString("\n")

	// Render chart content based on state
	switch m.state {
	case StateLoading:
		builder.WriteString(m.renderLoading())
	case StateLoaded:
		if len(m.data.Bars) > 0 {
			builder.WriteString(m.renderChart())
		} else {
			builder.WriteString(m.renderEmpty())
		}
	case StateError:
		builder.WriteString(m.renderError())
	}

	return builder.String()
}

// renderHeader renders the status chip at the top.
func (m *Model) renderHeader() string {
	symbol := m.symbol
	if symbol == "" {
		symbol = "N/A"
	}

	timeframeStr := m.timeframe.String()
	barCloseStr := "N/A"
	if !m.barClose.IsZero() {
		barCloseStr = m.barClose.Format("2006-01-02")
	}

	chip := fmt.Sprintf("─ %s · %s · window %d · bar-close: %s ─",
		symbol, timeframeStr, m.window, barCloseStr)

	return chip
}

// renderLoading renders a loading indicator.
func (m *Model) renderLoading() string {
	spinner := components.NewSpinner()
	return spinner.Render() + " Loading chart data..."
}

// renderEmpty renders a message when no data is available.
func (m *Model) renderEmpty() string {
	return "No data available. Press 'r' to refresh."
}

// renderError renders an error message.
func (m *Model) renderError() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	return "An unknown error occurred."
}

// renderChart renders the candlestick chart with TPI overlay.
func (m *Model) renderChart() string {
	// Calculate pane heights (price ~70%, TPI ~30%)
	totalHeight := m.height
	priceHeight := (totalHeight * 7) / 10
	tpiHeight := totalHeight - priceHeight

	// Get visible bars based on offset and window
	visibleBars := m.getVisibleBars()
	if len(visibleBars) == 0 {
		return "No data to display in current view."
	}

	// Calculate price range
	minPrice, maxPrice := getPriceRange(visibleBars)

	// Get current price (last visible bar's close)
	var currentPrice float64
	if len(visibleBars) > 0 {
		currentPrice = visibleBars[len(visibleBars)-1].Close
	}

	// Get current TPI (last visible TPI value)
	var currentTPI float64
	tpiIndex := m.offset + len(visibleBars) - 1
	if tpiIndex >= 0 && tpiIndex < len(m.data.TPI) {
		currentTPI = m.data.TPI[tpiIndex]
	}

	// Render price pane
	pricePane := renderPricePane(visibleBars, minPrice, maxPrice, currentPrice, m.width, priceHeight)

	// Render TPI pane
	tpiPane := renderTPIPane(visibleBars, m.data.TPI, currentTPI, m.offset, m.width, tpiHeight)

	// Join panes vertically
	return pricePane + "\n" + strings.Repeat("─", m.width) + "\n" + tpiPane
}

// getVisibleBars returns the bars visible in the current view.
func (m *Model) getVisibleBars() []indicators.OHLCV {
	if len(m.data.Bars) == 0 {
		return nil
	}

	start := m.offset
	end := start + m.window
	if end > len(m.data.Bars) {
		end = len(m.data.Bars)
	}
	if start < 0 {
		start = 0
	}
	if start >= len(m.data.Bars) {
		start = len(m.data.Bars) - 1
		if start < 0 {
			start = 0
		}
		end = start
	}

	return m.data.Bars[start:end]
}

// getPriceRange returns the min and max price for the given bars.
func getPriceRange(bars []indicators.OHLCV) (float64, float64) {
	if len(bars) == 0 {
		return 0, 0
	}

	minPrice := bars[0].Low
	maxPrice := bars[0].High

	for _, bar := range bars {
		if bar.Low < minPrice {
			minPrice = bar.Low
		}
		if bar.High > maxPrice {
			maxPrice = bar.High
		}
	}

	return minPrice, maxPrice
}

// renderPricePane renders the price candlestick pane with Y-axis labels.
func renderPricePane(bars []indicators.OHLCV, minPrice, maxPrice, currentPrice float64, width, height int) string {
	if height <= 0 {
		return ""
	}

	// Calculate chart width (subtract space for Y-axis labels)
	gutterWidth := 10  // 9 for price labels + 1 space
	currentWidth := 10 // 9 for current price + 1 space
	chartWidth := width - gutterWidth - currentWidth
	if chartWidth <= 0 {
		return fmt.Sprintf("Width %d too small for price pane", width)
	}

	// Render the candlestick chart
	chart := renderCandles(bars, chartWidth, height, "default")

	// Split chart into lines
	chartLines := strings.Split(chart, "\n")

	// Prepare Y-axis labels (max, min)
	maxLabel := fmt.Sprintf("%9.2f", maxPrice)
	minLabel := fmt.Sprintf("%9.2f", minPrice)

	// Prepare current price label
	currentLabel := fmt.Sprintf("%9.2f", currentPrice)

	// Build result with Y-axis gutter, chart, and current value
	var result strings.Builder
	for i := 0; i < height; i++ {
		// Y-axis label on left
		//nolint:staticcheck // QF1002: tagged switch not applicable for variable comparisons
		switch {
		case i == 0:
			result.WriteString(maxLabel)
		case i == height-1:
			result.WriteString(minLabel)
		default:
			result.WriteString("         ") // 9 spaces
		}
		result.WriteString(" ")

		// Chart content
		if i < len(chartLines) {
			result.WriteString(chartLines[i])
		} else {
			result.WriteString(strings.Repeat(" ", chartWidth))
		}

		// Current value label on right
		if i == height/2 {
			result.WriteString(" ")
			result.WriteString(currentLabel)
		}

		result.WriteString("\n")
	}

	return result.String()
}

// renderTPIPane renders the TPI line pane with Y-axis labels.
func renderTPIPane(bars []indicators.OHLCV, tpi []float64, currentTPI float64, offset int, width, height int) string {
	if height <= 0 {
		return ""
	}

	// Calculate chart width (subtract space for Y-axis labels)
	gutterWidth := 5  // 4 for TPI labels + 1 space
	currentWidth := 6 // 5 for current TPI + 1 space
	chartWidth := width - gutterWidth - currentWidth
	if chartWidth <= 0 {
		return fmt.Sprintf("Width %d too small for TPI pane", width)
	}

	// Get TPI slice for visible bars
	startIdx := offset
	endIdx := offset + len(bars)
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(tpi) {
		endIdx = len(tpi)
	}
	visibleTPI := tpi[startIdx:endIdx]

	// Render the TPI chart
	chart := renderTPI(visibleTPI, chartWidth, height, "default")

	// Split chart into lines
	chartLines := strings.Split(chart, "\n")

	// Prepare Y-axis labels (+1.0, 0.0, -1.0)
	topLabel := "+1.0"
	midLabel := " 0.0"
	bottomLabel := "-1.0"

	// Prepare current TPI label
	currentLabel := fmt.Sprintf("%+5.2f", currentTPI)

	// Build result with Y-axis gutter, chart, and current value
	var result strings.Builder
	for i := 0; i < height; i++ {
		// Y-axis label on left
		//nolint:staticcheck // QF1002: tagged switch not applicable for variable comparisons
		switch {
		case i == 0:
			result.WriteString(topLabel)
		case i == height-1:
			result.WriteString(bottomLabel)
		case i == height/2:
			result.WriteString(midLabel)
		default:
			result.WriteString("    ") // 4 spaces
		}
		result.WriteString(" ")

		// Chart content
		if i < len(chartLines) {
			result.WriteString(chartLines[i])
		} else {
			result.WriteString(strings.Repeat(" ", chartWidth))
		}

		// Current value label on right
		if i == height/2 {
			result.WriteString(" ")
			result.WriteString(currentLabel)
		}

		result.WriteString("\n")
	}

	return result.String()
}
