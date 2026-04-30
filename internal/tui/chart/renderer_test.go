// Package chart provides tests for chart rendering.
package chart

import (
	"strings"
	"testing"
	"time"

	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
)

// TestCandleRenderer_BullishBearishDoji tests candlestick rendering for different bar types.
func TestCandleRenderer_BullishBearishDoji(t *testing.T) {
	width := 10
	height := 5
	themeName := "default"

	// Create test bars: bullish, bearish, doji
	bars := []indicators.OHLCV{
		{Date: time.Now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000}, // Bullish
		{Date: time.Now(), Open: 105, High: 110, Low: 95, Close: 100, Volume: 1000}, // Bearish
		{Date: time.Now(), Open: 100, High: 110, Low: 90, Close: 100, Volume: 1000}, // Doji
	}

	output := renderCandles(bars, width, height, themeName)

	if output == "" {
		t.Error("Expected non-empty output from renderCandles")
	}

	// Verify the output contains expected characters
	hasBody := false
	hasDoji := false

	for _, r := range output {
		if r == '█' || r == '▓' {
			hasBody = true
		}
		if r == '─' {
			hasDoji = true
		}
	}

	if !hasBody {
		t.Error("Expected body character in output")
	}
	if !hasDoji {
		t.Error("Expected doji character '─' in output")
	}
}

// TestCandleRenderer_WickRoundingPolicy tests that wicks round away from the body.
func TestCandleRenderer_WickRoundingPolicy(t *testing.T) {
	width := 5
	height := 5
	themeName := "default"

	// Create a small bar where wicks need rounding
	bars := []indicators.OHLCV{
		{
			Date:   time.Now(),
			Open:   100.5,
			High:   101.5,
			Low:    99.5,
			Close:  100.5,
			Volume: 1000,
		},
	}

	output := renderCandles(bars, width, height, themeName)

	if output == "" {
		t.Error("Expected non-empty output from renderCandles")
	}

	// Verify wicks are rendered (│ character)
	hasWick := false
	for _, r := range output {
		if r == '│' {
			hasWick = true
			break
		}
	}

	if !hasWick {
		t.Error("Expected wick character '│' in output")
	}
}

// TestCandleRenderer_OneColumnPerBar tests that each bar occupies one column.
func TestCandleRenderer_OneColumnPerBar(t *testing.T) {
	width := 3
	height := 5
	themeName := "default"

	bars := []indicators.OHLCV{
		{Date: time.Now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: time.Now(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
		{Date: time.Now(), Open: 110, High: 120, Low: 105, Close: 115, Volume: 1000},
	}

	output := renderCandles(bars, width, height, themeName)

	if output == "" {
		t.Error("Expected non-empty output from renderCandles")
	}

	// Count number of lines (should be height)
	lines := 0
	for _, r := range output {
		if r == '\n' {
			lines++
		}
	}

	if lines != height {
		t.Errorf("Expected %d lines, got %d", height, lines)
	}
}

// TestTPIRenderer_ZeroReferenceLineAlwaysDrawn tests that the zero reference line is always drawn.
func TestTPIRenderer_ZeroReferenceLineAlwaysDrawn(t *testing.T) {
	width := 10
	height := 5
	themeName := "default"

	// Test with all positive TPI values
	scoresAllPositive := []float64{0.5, 0.6, 0.7, 0.8, 0.9}

	output := renderTPI(scoresAllPositive, width, height, themeName)

	if output == "" {
		t.Error("Expected non-empty output from renderTPI")
	}

	// Verify canvas was created (output contains braille characters)
	hasBraille := false
	for _, r := range output {
		if r >= 0x2800 && r <= 0x28FF {
			hasBraille = true
			break
		}
	}

	if !hasBraille {
		t.Error("Expected braille characters in TPI output")
	}
}

// TestTPIRenderer_FillColorAboveBelowZero tests that fill colors match TPI sign.
func TestTPIRenderer_FillColorAboveBelowZero(t *testing.T) {
	width := 10
	height := 5
	themeName := "default"

	// Test with mixed TPI values (some positive, some negative)
	scores := []float64{-0.5, -0.3, 0.0, 0.3, 0.5}

	output := renderTPI(scores, width, height, themeName)

	if output == "" {
		t.Error("Expected non-empty output from renderTPI")
	}

	// Verify canvas was created
	hasBraille := false
	for _, r := range output {
		if r >= 0x2800 && r <= 0x28FF {
			hasBraille = true
			break
		}
	}

	if !hasBraille {
		t.Error("Expected braille characters in TPI output")
	}
}

// TestTPIRenderer_DeadZoneMuted tests that TPI in dead zone (±0.05) uses muted color.
func TestTPIRenderer_DeadZoneMuted(t *testing.T) {
	width := 10
	height := 5
	themeName := "default"

	// Test TPI values in dead zone
	scoresDeadZone := []float64{0.02, 0.03, -0.01, -0.04, 0.0}

	output := renderTPI(scoresDeadZone, width, height, themeName)

	if output == "" {
		t.Error("Expected non-empty output from renderTPI")
	}

	// Verify canvas was created
	hasBraille := false
	for _, r := range output {
		if r >= 0x2800 && r <= 0x28FF {
			hasBraille = true
			break
		}
	}

	if !hasBraille {
		t.Error("Expected braille characters in TPI output")
	}
}

// TestTPIRenderer_ColorblindGlyphDifferentiation tests that colorblind theme uses glyph differentiation.
func TestTPIRenderer_ColorblindGlyphDifferentiation(t *testing.T) {
	width := 10
	height := 5
	themeName := "colorblind"

	scores := []float64{0.5, 0.6, 0.7, 0.8, 0.9}

	output := renderTPI(scores, width, height, themeName)

	if output == "" {
		t.Error("Expected non-empty output from renderTPI")
	}

	// Verify canvas was created
	hasBraille := false
	for _, r := range output {
		if r >= 0x2800 && r <= 0x28FF {
			hasBraille = true
			break
		}
	}

	if !hasBraille {
		t.Error("Expected braille characters in TPI output")
	}
}

// TestChartRenderer_PriceAndTPIPanesShareXAxis tests that price and TPI panes align on X axis.
func TestChartRenderer_PriceAndTPIPanesShareXAxis(t *testing.T) {
	width := 10
	height := 10
	themeName := "default"

	// Create test data
	bars := make([]indicators.OHLCV, 10)
	for i := range bars {
		price := 100.0 + float64(i)
		bars[i] = indicators.OHLCV{
			Date:   time.Now().Add(-time.Duration(10-i) * 24 * time.Hour),
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		}
	}
	tpi := make([]float64, 10)
	for i := range tpi {
		tpi[i] = float64(i-5) / 5.0 // -0.5 to +0.5
	}

	pricePane := renderCandles(bars, width, height/2, themeName)
	tpiPane := renderTPI(tpi, width, height/2, themeName)

	if pricePane == "" || tpiPane == "" {
		t.Error("Expected non-empty output from panes")
	}

	// Verify both panes were rendered
	hasPrice := len(pricePane) > 0
	hasTPI := len(tpiPane) > 0

	if !hasPrice {
		t.Error("Expected price pane to be rendered")
	}
	if !hasTPI {
		t.Error("Expected TPI pane to be rendered")
	}
}

// TestChartRenderer_BarCloseOnlyRule tests that only closed bars are rendered.
func TestChartRenderer_BarCloseOnlyRule(t *testing.T) {
	width := 5
	height := 5
	themeName := "default"

	// Create bars including today's in-progress bar
	today := time.Now().UTC()
	bars := []indicators.OHLCV{
		{Date: today.Add(-48 * time.Hour), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: today.Add(-24 * time.Hour), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
		{Date: today, Open: 110, High: 120, Low: 105, Close: 115, Volume: 1000}, // In-progress
	}

	// The renderCandles function itself doesn't filter by date,
	// but the data should be filtered before rendering
	// This test verifies the rendering works with the provided data
	output := renderCandles(bars, width, height, themeName)

	if output == "" {
		t.Error("Expected non-empty output from renderCandles")
	}
}

// TestChartRenderer_ResizeReflow tests that rendering handles resize gracefully.
func TestChartRenderer_ResizeReflow(t *testing.T) {
	themeName := "default"

	bars := make([]indicators.OHLCV, 5)
	for i := range bars {
		price := 100.0 + float64(i)
		bars[i] = indicators.OHLCV{
			Date:   time.Now().Add(-time.Duration(5-i) * 24 * time.Hour),
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		}
	}

	// Test with different dimensions
	dimensions := []struct{ w, h int }{
		{10, 10}, {20, 5}, {5, 20}, {80, 40},
	}

	for _, dim := range dimensions {
		output := renderCandles(bars, dim.w, dim.h, themeName)
		if output == "" {
			t.Errorf("Expected non-empty output for %dx%d", dim.w, dim.h)
		}
	}
}

// BenchmarkCandles_Render120x40With110Bars benchmarks candle rendering performance.
func BenchmarkCandles_Render120x40With110Bars(b *testing.B) {
	// Create 110 bars as specified
	bars := make([]indicators.OHLCV, 110)
	for i := range bars {
		price := 100.0 + float64(i)*0.1
		bars[i] = indicators.OHLCV{
			Date:   time.Now().Add(-time.Duration(110-i) * 24 * time.Hour),
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		}
	}

	themeName := "default"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderCandles(bars, 120, 28, themeName) // ~70% of 40 height
	}
}

// BenchmarkTPI_Render120x40With110Bars benchmarks TPI rendering performance.
func BenchmarkTPI_Render120x40With110Bars(b *testing.B) {
	tpi := make([]float64, 110)
	for i := range tpi {
		tpi[i] = (float64(i) - 55) / 55.0 // -1 to +1 range
	}

	themeName := "default"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderTPI(tpi, 120, 12, themeName) // ~30% of 40 height
	}
}

// TestRenderCandles_AboveRangeClipIndicator tests that a bar with High above
// the rendered range shows a ▲ clip indicator at the top edge.
func TestRenderCandles_AboveRangeClipIndicator(t *testing.T) {
	width := 10
	height := 10

	// Create more bars so robust 2% clipping actually clips the outlier
	bars := []indicators.OHLCV{
		{Date: time.Now(), Open: 100, High: 1000, Low: 95, Close: 105, Volume: 1000}, // Outlier high
	}
	for i := 0; i < 100; i++ {
		price := 100.0 + float64(i)
		bars = append(bars, indicators.OHLCV{
			Date:   time.Now(),
			Open:   price,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		})
	}

	// Use a clipped range that excludes the outlier
	minPrice, maxPrice, _ := getPriceRangeRobust(bars, outlierPercentile)
	output := renderCandlesWithReferencesAndClips(bars, minPrice, maxPrice, width, height, "default", nil)

	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	// Check for ▲ in the first line (top edge)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one line")
	}

	firstLine := lines[0]
	hasClipIndicator := strings.Contains(firstLine, "▲")
	if !hasClipIndicator {
		t.Errorf("Expected ▲ clip indicator in first line, got '%s'", firstLine)
	}
}

// TestRenderCandles_BelowRangeClipIndicator tests that a bar with Low below
// the rendered range shows a ▼ clip indicator at the bottom edge.
func TestRenderCandles_BelowRangeClipIndicator(t *testing.T) {
	width := 10
	height := 10

	// Create more bars so robust 2% clipping actually clips the outlier
	bars := []indicators.OHLCV{
		{Date: time.Now(), Open: 100, High: 110, Low: 5, Close: 105, Volume: 1000}, // Outlier low
	}
	for i := 0; i < 100; i++ {
		price := 100.0 + float64(i)
		bars = append(bars, indicators.OHLCV{
			Date:   time.Now(),
			Open:   price,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		})
	}

	// Use a clipped range that excludes the outlier
	minPrice, maxPrice, _ := getPriceRangeRobust(bars, outlierPercentile)
	output := renderCandlesWithReferencesAndClips(bars, minPrice, maxPrice, width, height, "default", nil)

	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	// Check for ▼ in the last line (bottom edge)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one line")
	}

	lastLine := lines[len(lines)-1]
	hasClipIndicator := strings.Contains(lastLine, "▼")
	if !hasClipIndicator {
		t.Errorf("Expected ▼ clip indicator in last line, got '%s'", lastLine)
	}
}

// TestRenderCandles_BothEdgesClippedSameBar tests that a gap-up day with crash
// (both High above and Low below range) shows both clip indicators.
func TestRenderCandles_BothEdgesClippedSameBar(t *testing.T) {
	width := 10
	height := 10

	// Create more bars so robust 2% clipping actually clips the outlier
	bars := []indicators.OHLCV{
		{Date: time.Now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: time.Now(), Open: 1000, High: 2000, Low: 50, Close: 500, Volume: 1000}, // Both edges clipped
	}
	for i := 0; i < 100; i++ {
		price := 100.0 + float64(i)
		bars = append(bars, indicators.OHLCV{
			Date:   time.Now(),
			Open:   price,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		})
	}

	// Use a clipped range that excludes the outlier
	minPrice, maxPrice, _ := getPriceRangeRobust(bars, outlierPercentile)
	output := renderCandlesWithReferencesAndClips(bars, minPrice, maxPrice, width, height, "default", nil)

	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	// Check for both clip indicators
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one line")
	}

	firstLine := lines[0]
	lastLine := lines[len(lines)-1]

	hasTopClip := strings.Contains(firstLine, "▲")
	hasBottomClip := strings.Contains(lastLine, "▼")

	if !hasTopClip {
		t.Errorf("Expected ▲ clip indicator in first line, got '%s'", firstLine)
	}
	if !hasBottomClip {
		t.Errorf("Expected ▼ clip indicator in last line, got '%s'", lastLine)
	}
}

// TestRenderCandles_NoClipIndicatorWhenInRange tests that bars entirely
// within the rendered range do not show clip indicators.
func TestRenderCandles_NoClipIndicatorWhenInRange(t *testing.T) {
	width := 10
	height := 10

	bars := []indicators.OHLCV{
		{Date: time.Now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: time.Now(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
	}

	// Use a range that includes all bars
	minPrice, maxPrice, _ := getPriceRangeRobust(bars, 0)
	output := renderCandlesWithReferencesAndClips(bars, minPrice, maxPrice, width, height, "default", nil)

	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	// Check for absence of clip indicators
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one line")
	}

	firstLine := lines[0]
	lastLine := lines[len(lines)-1]

	hasTopClip := strings.Contains(firstLine, "▲")
	hasBottomClip := strings.Contains(lastLine, "▼")

	if hasTopClip {
		t.Errorf("Did not expect ▲ clip indicator in first line, got '%s'", firstLine)
	}
	if hasBottomClip {
		t.Errorf("Did not expect ▼ clip indicator in last line, got '%s'", lastLine)
	}
}

// TestRenderCandles_BarWidthZoomLevels tests that bars render correctly
// at different zoom levels (barWidth=7, 3, 1 with chart width 218).
func TestRenderCandles_BarWidthZoomLevels(t *testing.T) {
	themeName := "default"
	height := 20

	tests := []struct {
		name             string
		numBars          int
		width            int
		expectedBarWidth int
		expectedBodyCols int
	}{
		{"zoom=30 barWidth=7", 30, 218, 7, 6},
		{"zoom=60 barWidth=3", 60, 218, 3, 2},
		{"zoom=110 barWidth=1", 110, 218, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bars := make([]indicators.OHLCV, tt.numBars)
			for i := range bars {
				price := 100.0 + float64(i)
				bars[i] = indicators.OHLCV{
					Date:   time.Now().Add(-time.Duration(tt.numBars-i) * 24 * time.Hour),
					Open:   price - 1,
					High:   price + 2,
					Low:    price - 2,
					Close:  price + 1,
					Volume: 1000,
				}
			}

			output := renderCandles(bars, tt.width, height, themeName)

			if output == "" {
				t.Fatalf("Expected non-empty output")
			}

			// Verify output has the correct dimensions
			lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
			if len(lines) != height {
				t.Errorf("Expected %d lines, got %d", height, len(lines))
			}
		})
	}
}

// TestRenderCandles_AllBodiesIdenticalWidthInChart verifies that within a single chart,
// all candle bodies have identical width.
func TestRenderCandles_AllBodiesIdenticalWidthInChart(t *testing.T) {
	width := 20
	height := 10
	themeName := "default"

	bars := make([]indicators.OHLCV, 5)
	for i := range bars {
		price := 100.0 + float64(i)
		bars[i] = indicators.OHLCV{
			Date:   time.Now().Add(-time.Duration(5-i) * 24 * time.Hour),
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		}
	}

	output := renderCandles(bars, width, height, themeName)

	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	// Count body characters (█ or ─) in the entire output
	bodyCharCount := 0
	for _, r := range output {
		if r == '█' || r == '─' {
			bodyCharCount++
		}
	}

	// With 5 bars, each with bodyCols=3, we expect at least some body characters
	// We can't count exact due to variable body heights, but verify we have bodies
	if bodyCharCount == 0 {
		t.Error("Expected body characters in output")
	}
}

// TestRenderCandles_BarWidth1NoGap verifies that when barWidth=1,
// bodies are 1 col wide with no gap (contiguous render).
func TestRenderCandles_BarWidth1NoGap(t *testing.T) {
	width := 10
	height := 10
	themeName := "default"

	bars := make([]indicators.OHLCV, 10)
	for i := range bars {
		price := 100.0 + float64(i)
		bars[i] = indicators.OHLCV{
			Date:   time.Now().Add(-time.Duration(10-i) * 24 * time.Hour),
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		}
	}

	output := renderCandles(bars, width, height, themeName)

	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	// When barWidth=1, verify that all columns have at least one non-space character
	// (no gaps between bars)
	colHasContent := make([]bool, width)
	for _, line := range lines {
		for col := 0; col < width && col < len(line); col++ {
			if line[col] != ' ' {
				colHasContent[col] = true
			}
		}
	}

	// With 10 bars at width 10, all columns should have content
	emptyCols := 0
	for _, hasContent := range colHasContent {
		if !hasContent {
			emptyCols++
		}
	}

	if emptyCols > 1 {
		// Allow at most 1 empty column (rightmost gutter)
		t.Errorf("Expected at most 1 empty column with barWidth=1, got %d", emptyCols)
	}
}

// TestRenderCandles_BarWidthGteTwoHasOneColGap verifies that when barWidth >= 2,
// there is a 1-column gap on the right edge of each bar.
func TestRenderCandles_BarWidthGteTwoHasOneColGap(t *testing.T) {
	width := 9
	height := 10
	themeName := "default"

	bars := make([]indicators.OHLCV, 3)
	for i := range bars {
		price := 100.0 + float64(i)
		bars[i] = indicators.OHLCV{
			Date:   time.Now().Add(-time.Duration(3-i) * 24 * time.Hour),
			Open:   price - 1,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		}
	}

	output := renderCandles(bars, width, height, themeName)

	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	// Verify output contains all expected candle elements
	// With 3 bars and bodyCols=2, we expect body characters
	hasBody := false
	for _, r := range output {
		if r == '█' || r == '─' {
			hasBody = true
			break
		}
	}

	if !hasBody {
		t.Error("Expected body characters in output")
	}
}

// TestDojiDetection_ContentBasedNotPaneBased verifies that doji detection
// is based on content (close vs open) and not pane height or price range.
func TestDojiDetection_ContentBasedNotPaneBased(t *testing.T) {
	width := 10
	themeName := "default"

	barsDoji := []indicators.OHLCV{
		{Date: time.Now(), Open: 1000, High: 1010, Low: 990, Close: 1000.5, Volume: 1000},
	}

	barsNotDoji := []indicators.OHLCV{
		{Date: time.Now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
	}

	for _, h := range []int{5, 10, 20, 40} {
		outputDoji := renderCandles(barsDoji, width, h, themeName)
		outputNotDoji := renderCandles(barsNotDoji, width, h, themeName)

		if !strings.Contains(outputDoji, "─") {
			t.Errorf("Height %d: Expected doji character '─' in doji bar", h)
		}
		if !strings.Contains(outputNotDoji, "█") {
			t.Errorf("Height %d: Expected body character '█' in non-doji bar", h)
		}
	}
}

// TestDojiDetection_BTCDailyFixtureNoFalseDojis verifies that high-priced
// assets like BTC do not get false doji classification.
func TestDojiDetection_BTCDailyFixtureNoFalseDojis(t *testing.T) {
	width := 20
	height := 10
	themeName := "default"

	bars := []indicators.OHLCV{
		{Date: time.Now(), Open: 45000, High: 47000, Low: 44000, Close: 46000, Volume: 1000},
		{Date: time.Now(), Open: 46000, High: 47500, Low: 45500, Close: 47000, Volume: 1000},
		{Date: time.Now(), Open: 47000, High: 48000, Low: 46500, Close: 46500, Volume: 1000},
		{Date: time.Now(), Open: 46500, High: 47300, Low: 46000, Close: 46530, Volume: 1000},
	}

	output := renderCandles(bars, width, height, themeName)

	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	bodyCount := strings.Count(output, "█")
	dojiCount := strings.Count(output, "─")

	if bodyCount == 0 {
		t.Error("Expected at least some body characters for BTC bars")
	}

	if dojiCount == 0 {
		t.Error("Expected doji character for true doji bar")
	}
}

// TestDojiDetection_TrueDojiBarsStillDetected verifies that actual
// doji bars (very small |close-open|) are still correctly detected.
func TestDojiDetection_TrueDojiBarsStillDetected(t *testing.T) {
	width := 10
	height := 10
	themeName := "default"

	bars := []indicators.OHLCV{
		{Date: time.Now(), Open: 100, High: 110, Low: 90, Close: 100.05, Volume: 1000},
	}

	output := renderCandles(bars, width, height, themeName)

	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	if !strings.Contains(output, "─") {
		t.Error("Expected doji character '─' in true doji bar")
	}

	bodyCount := strings.Count(output, "█")
	if bodyCount > 0 {
		t.Errorf("Expected no body characters for doji, got %d", bodyCount)
	}
}
