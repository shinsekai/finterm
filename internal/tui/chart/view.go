// Package chart provides view rendering for the chart tab.
package chart

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// renderView renders the chart view.
func renderView(m *Model) string {
	var builder strings.Builder

	// Calculate clipped count for header chip
	clippedCount := 0
	if m.state == StateLoaded && len(m.data.Bars) > 0 {
		visibleBars := m.getVisibleBars()
		if len(visibleBars) > 0 {
			_, _, clippedCount = getPriceRangeRobust(visibleBars, outlierPercentile)
		}
	}

	// Render header with status chip
	builder.WriteString(m.renderHeader(clippedCount))
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
// clippedCount is the number of bars clipped from the visible range.
//
// Chosen format (Option A): Simple cleanup with plain text.
// Format: "SYMBOL  timeframe · window N · YYYY-MM-DD"
// Example: "AAPL  daily · window 110 · 2026-04-27"
//
// This format:
// - Removes leading/trailing dashes that were visually noisy
// - Drops the "bar-close:" prefix (the date format is self-explanatory)
// - Uses "(no symbol selected)" instead of "N/A" for empty symbols
// - Keeps the clean two-space separator between symbol and timeframe
// - Maintains consistent visual weight across all theme modes
// Note: Theme styling support can be added in the future by passing
// *tui.Theme and using lipgloss.Style() for colors.
func (m *Model) renderHeader(clippedCount int) string {
	symbol := m.symbol
	if symbol == "" {
		symbol = "(no symbol selected)"
	}

	timeframeStr := m.timeframe.String()
	barCloseStr := ""
	if !m.barClose.IsZero() {
		barCloseStr = m.barClose.Format("2006-01-02")
	}

	var builder strings.Builder
	builder.WriteString(symbol)
	builder.WriteString("  ")
	builder.WriteString(timeframeStr)
	fmt.Fprintf(&builder, " · window %d", m.window)

	if barCloseStr != "" {
		builder.WriteString(" · ")
		builder.WriteString(barCloseStr)
	}

	// Add clip chip if any bars were clipped
	if clippedCount > 0 {
		fmt.Fprintf(&builder, " · (%d bars clipped)", clippedCount)
	}

	return builder.String()
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

	// Render price pane (includes robust price range with outlier clipping)
	pricePane, _ := renderPricePane(visibleBars, currentPrice, m.width, priceHeight)

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
//
// Deprecated: Use getPriceRangeRobust for outlier handling.
//
//nolint:unused // Kept for potential future use
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

// outlierPercentile is the default percentile to clip from both ends.
// Clip top and bottom 2% of values to handle outliers.
const outlierPercentile = 0.02

// getPriceRangeRobust computes a robust price range by clipping outliers.
// It computes the percentile and 1-percentile quantiles of all Low and High values,
// and always includes the most recent bar's High and Low in the returned range.
// Returns (clippedMin, clippedMax, clippedCount) where clippedCount is the number
// of bars that had High or Low outside the clipped range.
func getPriceRangeRobust(bars []indicators.OHLCV, percentile float64) (float64, float64, int) {
	if len(bars) == 0 {
		return 0, 0, 0
	}

	// Collect all Low and High values
	allPrices := make([]float64, 0, len(bars)*2)
	for _, bar := range bars {
		allPrices = append(allPrices, bar.Low, bar.High)
	}

	// Sort for percentile computation
	sort.Float64s(allPrices)

	// Compute quantile indices
	lowerIdx := int(float64(len(allPrices)) * percentile)
	upperIdx := int(float64(len(allPrices)) * (1 - percentile))

	// Clamp indices to valid range
	if lowerIdx < 0 {
		lowerIdx = 0
	}
	if upperIdx >= len(allPrices) {
		upperIdx = len(allPrices) - 1
	}
	if lowerIdx >= upperIdx {
		lowerIdx = 0
		upperIdx = len(allPrices) - 1
	}

	clippedMin := allPrices[lowerIdx]
	clippedMax := allPrices[upperIdx]

	// Always include the latest bar (the most recent closed bar must always be visible)
	latestBar := bars[len(bars)-1]
	if latestBar.Low < clippedMin {
		clippedMin = latestBar.Low
	}
	if latestBar.High > clippedMax {
		clippedMax = latestBar.High
	}

	// Count how many bars are clipped (have High or Low outside the range)
	clippedCount := 0
	for _, bar := range bars {
		if bar.Low < clippedMin || bar.High > clippedMax {
			clippedCount++
		}
	}

	return clippedMin, clippedMax, clippedCount
}

// renderPricePane renders the price candlestick pane with Y-axis labels.
// Uses robust price range with outlier clipping and returns the rendered string
// along with the count of clipped bars for the chip display.
func renderPricePane(bars []indicators.OHLCV, currentPrice float64, width, height int) (string, int) {
	if height <= 0 {
		return "", 0
	}

	// Calculate robust price range with outlier clipping
	minPrice, maxPrice, clippedCount := getPriceRangeRobust(bars, outlierPercentile)

	// Calculate chart width (subtract space for Y-axis labels)
	gutterWidth := 10  // 9 for price labels + 1 space
	currentWidth := 10 // 9 for current price + 1 space
	chartWidth := width - gutterWidth - currentWidth
	if chartWidth <= 0 {
		return fmt.Sprintf("Width %d too small for price pane", width), clippedCount
	}

	// Calculate quartile prices for reference lines
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}
	upperQuartilePrice := maxPrice - (priceRange * 0.25)
	lowerQuartilePrice := maxPrice - (priceRange * 0.75)

	// Render the candlestick chart with reference lines and clip indicators
	references := []float64{upperQuartilePrice, lowerQuartilePrice}
	chart := renderCandlesWithReferencesAndClips(bars, minPrice, maxPrice, chartWidth, height, "default", references)

	// Split chart into lines
	chartLines := strings.Split(chart, "\n")

	// Prepare Y-axis labels (max, upper Q, lower Q, min)
	maxLabel := fmt.Sprintf("%9.2f", maxPrice)
	upperQuartileLabel := fmt.Sprintf("%9.2f", upperQuartilePrice)
	lowerQuartileLabel := fmt.Sprintf("%9.2f", lowerQuartilePrice)
	minLabel := fmt.Sprintf("%9.2f", minPrice)

	// Prepare current price label (top-right)
	currentLabel := fmt.Sprintf("%9.2f", currentPrice)

	// Build result with Y-axis gutter, chart, and current value
	var result strings.Builder
	for i := 0; i < height; i++ {
		// Y-axis label on left (max, upper Q, lower Q, min)
		//nolint:staticcheck // QF1002: tagged switch not applicable for variable comparisons
		switch {
		case i == 0:
			result.WriteString(maxLabel)
		case i == height/4:
			result.WriteString(upperQuartileLabel)
		case i == (height*3)/4:
			result.WriteString(lowerQuartileLabel)
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

		// Current value label on right (top row only)
		if i == 0 {
			result.WriteString(" ")
			result.WriteString(currentLabel)
		}

		result.WriteString("\n")
	}

	return result.String(), clippedCount
}

// renderTPIPane renders the TPI line pane with Y-axis labels.
func renderTPIPane(bars []indicators.OHLCV, tpi []float64, _ float64, offset int, width, height int) string {
	if height <= 0 {
		return ""
	}

	// Calculate chart width (subtract space for Y-axis labels)
	gutterWidth := 5 // 4 for TPI labels + 1 space
	rightWidth := 5  // 4 for axis bound + 1 space
	chartWidth := width - gutterWidth - rightWidth
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

	// Split chart into lines and validate count
	chartLines := strings.Split(chart, "\n")
	// Defensive: ensure chartLines has exactly height lines
	if len(chartLines) < height {
		// Pad with empty lines if too few
		for len(chartLines) < height {
			chartLines = append(chartLines, "")
		}
	} else if len(chartLines) > height {
		// Truncate if too many (shouldn't happen, but defensive)
		chartLines = chartLines[:height]
	}

	// Prepare Y-axis labels (+1.0, 0.0, -1.0)
	topLabel := "+1.0"
	midLabel := " 0.0"
	bottomLabel := "-1.0"

	// Build result with Y-axis gutter, chart, and right-axis bound
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
		result.WriteString(chartLines[i])

		// Right-axis bound label (symmetric with left edge)
		if i == 0 {
			result.WriteString(" ")
			result.WriteString(topLabel)
		}

		result.WriteString("\n")
	}

	return result.String()
}
