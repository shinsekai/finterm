// Package chart provides tests for view rendering.
package chart

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
)

// TestTPIPane_AxisLabelsBothEdgesAreBounds tests that both left and right edges
// show axis bounds, not the current TPI value.
func TestTPIPane_AxisLabelsBothEdgesAreBounds(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
		{Date: now(), Open: 110, High: 120, Low: 105, Close: 115, Volume: 1000},
	}
	tpi := []float64{0.5, 0.6, 0.7}
	currentTPI := 0.7
	offset := 0
	width := 40
	height := 10

	output := renderTPIPane(bars, tpi, currentTPI, offset, width, height)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 1 {
		t.Fatal("Expected at least one line in output")
	}

	// Top line should have "+1.0" on both left and right edges
	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "+1.0 ") {
		t.Errorf("Expected left edge to start with '+1.0 ', got '%s'", firstLine)
	}
	if !strings.HasSuffix(firstLine, " +1.0") {
		t.Errorf("Expected right edge to end with ' +1.0', got '%s'", firstLine)
	}

	// Bottom line should have "-1.0" on left edge, nothing on right
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if !strings.HasPrefix(lastLine, "-1.0 ") {
		t.Errorf("Expected left edge to start with '-1.0 ', got '%s'", lastLine)
	}
	if strings.HasSuffix(lastLine, " +1.0") {
		t.Error("Expected right edge to not have +1.0 on bottom row")
	}
}

// TestTPIPane_BottomLabelVisibleAtAllHeights tests that the bottom "-1.0" label
// is visible at various pane heights.
func TestTPIPane_BottomLabelVisibleAtAllHeights(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
	}
	tpi := []float64{0.5, 0.6}
	currentTPI := 0.6
	offset := 0
	width := 40

	heights := []int{6, 8, 10, 12, 16}
	for _, height := range heights {
		t.Run(string(rune(height+'0')), func(t *testing.T) {
			output := renderTPIPane(bars, tpi, currentTPI, offset, width, height)

			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) != height {
				t.Errorf("Expected %d lines, got %d", height, len(lines))
			}

			// Bottom line should start with "-1.0"
			lastLine := strings.TrimSpace(lines[len(lines)-1])
			if !strings.HasPrefix(lastLine, "-1.0 ") {
				t.Errorf("Expected bottom line to start with '-1.0 ', got '%s'", lastLine)
			}
		})
	}
}

// TestTPIPane_NoCurrentValueLabel tests that the current TPI value is not rendered
// as a separate label anywhere in the pane.
func TestTPIPane_NoCurrentValueLabel(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
	}
	tpi := []float64{0.5, 0.6}
	currentTPI := 0.6
	offset := 0
	width := 40
	height := 10

	output := renderTPIPane(bars, tpi, currentTPI, offset, width, height)

	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check that no line contains the current TPI value formatted as "+0.60"
	currentLabel := "+0.60"
	for i, line := range lines {
		if strings.Contains(line, currentLabel) {
			t.Errorf("Line %d should not contain current value label '%s', got '%s'", i, currentLabel, line)
		}
	}

	// Check middle line - this is where the current value used to be
	middleLine := strings.TrimSpace(lines[height/2])
	if strings.Contains(middleLine, currentLabel) {
		t.Errorf("Middle line should not contain current value label '%s', got '%s'", currentLabel, middleLine)
	}
}

// TestTPIPane_LabelsIdenticalAcrossTickers tests that axis labels are identical
// across different tickers with various TPI values.
func TestTPIPane_LabelsIdenticalAcrossTickers(t *testing.T) {
	testCases := []struct {
		name       string
		symbol     string
		bars       []indicators.OHLCV
		tpi        []float64
		currentTPI float64
	}{
		{
			name: "QQQ - TPI near +1.0",
			bars: []indicators.OHLCV{
				{Date: now(), Open: 400, High: 410, Low: 395, Close: 405, Volume: 10000},
				{Date: now(), Open: 405, High: 415, Low: 400, Close: 410, Volume: 10000},
			},
			tpi:        []float64{0.95, 1.00},
			currentTPI: 1.00,
		},
		{
			name: "SPY - TPI near 0.5",
			bars: []indicators.OHLCV{
				{Date: now(), Open: 500, High: 510, Low: 495, Close: 505, Volume: 10000},
				{Date: now(), Open: 505, High: 515, Low: 500, Close: 510, Volume: 10000},
			},
			tpi:        []float64{0.45, 0.50},
			currentTPI: 0.50,
		},
		{
			name: "BTC - TPI near +0.40",
			bars: []indicators.OHLCV{
				{Date: now(), Open: 60000, High: 61000, Low: 59500, Close: 60500, Volume: 100},
				{Date: now(), Open: 60500, High: 61500, Low: 60000, Close: 61000, Volume: 100},
			},
			tpi:        []float64{0.35, 0.40},
			currentTPI: 0.40,
		},
		{
			name: "ETH - TPI near 0.0",
			bars: []indicators.OHLCV{
				{Date: now(), Open: 3000, High: 3100, Low: 2950, Close: 3050, Volume: 100},
				{Date: now(), Open: 3050, High: 3150, Low: 3000, Close: 3100, Volume: 100},
			},
			tpi:        []float64{-0.05, 0.00},
			currentTPI: 0.00,
		},
		{
			name: "AAPL - TPI near -0.50",
			bars: []indicators.OHLCV{
				{Date: now(), Open: 170, High: 180, Low: 165, Close: 175, Volume: 10000},
				{Date: now(), Open: 175, High: 185, Low: 170, Close: 180, Volume: 10000},
			},
			tpi:        []float64{-0.55, -0.50},
			currentTPI: -0.50,
		},
	}

	width := 40
	height := 10
	offset := 0

	// Verify axis labels are identical across all tickers
	expectedLeftTop := "+1.0"
	expectedLeftMid := " 0.0"
	expectedLeftBottom := "-1.0"
	expectedRightTop := "+1.0"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := renderTPIPane(tc.bars, tc.tpi, tc.currentTPI, offset, width, height)
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) < 1 {
				t.Fatal("Expected at least one line")
			}

			// Check left axis labels
			if !strings.HasPrefix(lines[0], expectedLeftTop+" ") {
				t.Errorf("Expected left top edge to be '%s ', got '%s'", expectedLeftTop, lines[0])
			}
			if !strings.HasPrefix(lines[height/2], expectedLeftMid+" ") {
				t.Errorf("Expected left mid edge to be '%s ', got '%s'", expectedLeftMid, lines[height/2])
			}
			if !strings.HasPrefix(lines[height-1], expectedLeftBottom+" ") {
				t.Errorf("Expected left bottom edge to be '%s ', got '%s'", expectedLeftBottom, lines[height-1])
			}

			// Check right axis label (only on top row)
			if !strings.HasSuffix(lines[0], " "+expectedRightTop) {
				t.Errorf("Expected right top edge to be ' %s', got '%s'", expectedRightTop, lines[0])
			}
		})
	}
}

// TestTPIPane_ChartLineCountMatchesHeight tests that the chart renderer
// produces exactly `height` lines, with defensive padding if needed.
func TestTPIPane_ChartLineCountMatchesHeight(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
	}
	tpi := []float64{0.5, 0.6}
	currentTPI := 0.6
	offset := 0
	width := 40

	heights := []int{6, 8, 10, 12, 16}
	for _, height := range heights {
		t.Run(string(rune(height+'0')), func(t *testing.T) {
			output := renderTPIPane(bars, tpi, currentTPI, offset, width, height)

			// Split and trim to account for final newline
			output = strings.TrimSpace(output)
			lines := strings.Split(output, "\n")

			if len(lines) != height {
				t.Errorf("Expected exactly %d lines, got %d", height, len(lines))
			}

			// Verify each line exists and has expected structure
			for i, line := range lines {
				if line == "" {
					// Empty line might be padding, but that's okay as long as count is right
					continue
				}

				// Check for label positions
				//nolint:staticcheck // QF1002: tagged switch not applicable for variable comparisons
				switch {
				case i == 0:
					if !strings.HasPrefix(line, "+1.0 ") {
						t.Errorf("Line %d should start with '+1.0 ', got '%s'", i, line)
					}
				case i == height-1:
					if !strings.HasPrefix(line, "-1.0 ") {
						t.Errorf("Line %d should start with '-1.0 ', got '%s'", i, line)
					}
				case i == height/2:
					if !strings.HasPrefix(line, " 0.0 ") {
						t.Errorf("Line %d should start with ' 0.0 ', got '%s'", i, line)
					}
				}
			}
		})
	}
}

// now returns a fixed time for tests to avoid flakiness.
func now() time.Time {
	return time.Unix(1700000000, 0) // Fixed timestamp
}

// TestPricePane_FourLeftAxisLabels tests that the price pane renders four Y-axis labels:
// max (top), upper quartile (top-25%), lower quartile (top-75%), min (bottom).
func TestPricePane_FourLeftAxisLabels(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 95, Close: 110, Volume: 1000},
	}
	minPrice := 90.0
	maxPrice := 115.0
	currentPrice := 110.0
	width := 40
	height := 20

	output := renderPricePane(bars, minPrice, maxPrice, currentPrice, width, height)

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != height {
		t.Fatalf("Expected %d lines, got %d", height, len(lines))
	}

	// Verify max label at top (row 0) - first 9 chars are label
	topLine := lines[0]
	actualMaxLabel := topLine[:9]
	expectedMax := fmt.Sprintf("%9.2f", maxPrice)
	if actualMaxLabel != expectedMax {
		t.Errorf("Expected max label '%s', got '%s'", expectedMax, actualMaxLabel)
	}

	// Verify min label at bottom (row height-1)
	bottomLine := lines[height-1]
	actualMinLabel := bottomLine[:9]
	expectedMin := fmt.Sprintf("%9.2f", minPrice)
	if actualMinLabel != expectedMin {
		t.Errorf("Expected min label '%s', got '%s'", expectedMin, actualMinLabel)
	}

	// Verify upper quartile label at row height/4
	upperQRow := height / 4
	upperQLine := lines[upperQRow]
	actualUpperQLabel := upperQLine[:9]
	expectedUpperQ := fmt.Sprintf("%9.2f", maxPrice-(maxPrice-minPrice)*0.25)
	if actualUpperQLabel != expectedUpperQ {
		t.Errorf("Expected upper quartile label '%s', got '%s'", expectedUpperQ, actualUpperQLabel)
	}

	// Verify lower quartile label at row (height*3)/4
	lowerQRow := (height * 3) / 4
	lowerQLine := lines[lowerQRow]
	actualLowerQLabel := lowerQLine[:9]
	expectedLowerQ := fmt.Sprintf("%9.2f", maxPrice-(maxPrice-minPrice)*0.75)
	if actualLowerQLabel != expectedLowerQ {
		t.Errorf("Expected lower quartile label '%s', got '%s'", expectedLowerQ, actualLowerQLabel)
	}
}

// TestPricePane_QuartileLabelsCorrectlyComputed tests that quartile labels are computed
// correctly for various price ranges.
func TestPricePane_QuartileLabelsCorrectlyComputed(t *testing.T) {
	testCases := []struct {
		name           string
		minPrice       float64
		maxPrice       float64
		height         int
		expectedUpperQ float64
		expectedLowerQ float64
	}{
		{
			name:           "range_100_to_200_height_20",
			minPrice:       100.0,
			maxPrice:       200.0,
			height:         20,
			expectedUpperQ: 175.0,
			expectedLowerQ: 125.0,
		},
		{
			name:           "range_0_to_100_height_10",
			minPrice:       0.0,
			maxPrice:       100.0,
			height:         10,
			expectedUpperQ: 75.0,
			expectedLowerQ: 25.0,
		},
		{
			name:           "range_50_to_150_height_16",
			minPrice:       50.0,
			maxPrice:       150.0,
			height:         16,
			expectedUpperQ: 125.0,
			expectedLowerQ: 75.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bars := []indicators.OHLCV{
				{Date: now(), Open: tc.minPrice, High: tc.maxPrice, Low: tc.minPrice, Close: tc.maxPrice, Volume: 1000},
			}
			currentPrice := (tc.minPrice + tc.maxPrice) / 2
			width := 40

			output := renderPricePane(bars, tc.minPrice, tc.maxPrice, currentPrice, width, tc.height)

			lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
			if len(lines) != tc.height {
				t.Fatalf("Expected %d lines, got %d", tc.height, len(lines))
			}

			upperQRow := tc.height / 4
			upperQLine := lines[upperQRow]
			actualUpperQLabel := upperQLine[:9]
			expectedUpperQStr := fmt.Sprintf("%9.2f", tc.expectedUpperQ)
			if actualUpperQLabel != expectedUpperQStr {
				t.Errorf("Expected upper quartile label '%s', got '%s'", expectedUpperQStr, actualUpperQLabel)
			}

			lowerQRow := (tc.height * 3) / 4
			lowerQLine := lines[lowerQRow]
			actualLowerQLabel := lowerQLine[:9]
			expectedLowerQStr := fmt.Sprintf("%9.2f", tc.expectedLowerQ)
			if actualLowerQLabel != expectedLowerQStr {
				t.Errorf("Expected lower quartile label '%s', got '%s'", expectedLowerQStr, actualLowerQLabel)
			}
		})
	}
}

// TestPricePane_CurrentPriceTopRight tests that current price label appears
// at the top-right of the pane (row 0), not in the middle.
func TestPricePane_CurrentPriceTopRight(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 95, Close: 110, Volume: 1000},
	}
	minPrice := 90.0
	maxPrice := 115.0
	currentPrice := 110.0
	width := 40
	height := 20

	output := renderPricePane(bars, minPrice, maxPrice, currentPrice, width, height)

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	topLine := lines[0]
	expectedCurrent := fmt.Sprintf("%9.2f", currentPrice)
	if !strings.HasSuffix(topLine, " "+expectedCurrent) {
		t.Errorf("Expected top line to end with current price ' %s'", expectedCurrent)
	}

	middleLine := lines[height/2]
	currentStr := fmt.Sprintf("%.2f", currentPrice)
	if strings.Contains(middleLine, currentStr) {
		t.Errorf("Middle line should not contain current price value '%s'", currentStr)
	}
}

// TestPricePane_CandlesWinOverReferenceLines tests that when a candle body overlaps
// a reference line at a given cell, the candle pixel is rendered instead of the reference dot.
func TestPricePane_CandlesWinOverReferenceLines(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 120, Low: 80, Close: 100, Volume: 1000},
	}
	minPrice := 80.0
	maxPrice := 120.0
	currentPrice := 100.0
	width := 40
	height := 40

	output := renderPricePane(bars, minPrice, maxPrice, currentPrice, width, height)

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	upperQRow := height / 4
	lowerQRow := (height * 3) / 4
	chartStart := 10

	upperQLine := lines[upperQRow]
	cellChar := upperQLine[chartStart]
	if cellChar == '·' || cellChar == ' ' {
		t.Errorf("Expected candle body at upper quartile row %d, got '%c'", upperQRow, cellChar)
	}

	lowerQLine := lines[lowerQRow]
	cellChar = lowerQLine[chartStart]
	if cellChar == '·' || cellChar == ' ' {
		t.Errorf("Expected candle body at lower quartile row %d, got '%c'", lowerQRow, cellChar)
	}
}

// TestPricePane_LayoutIdenticalAcrossTickers tests that price pane layout is
// identical across all five test tickers (QQQ, SPY, BTC, ETH, AAPL).
func TestPricePane_LayoutIdenticalAcrossTickers(t *testing.T) {
	testCases := []struct {
		name         string
		minPrice     float64
		maxPrice     float64
		currentPrice float64
	}{
		{"QQQ", 400.0, 450.0, 425.0},
		{"SPY", 500.0, 550.0, 525.0},
		{"BTC", 60000.0, 65000.0, 62500.0},
		{"ETH", 3000.0, 3500.0, 3250.0},
		{"AAPL", 170.0, 190.0, 180.0},
	}

	width := 40
	height := 20

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bars := []indicators.OHLCV{
				{Date: now(), Open: tc.minPrice, High: tc.maxPrice, Low: tc.minPrice, Close: tc.currentPrice, Volume: 1000},
			}

			output := renderPricePane(bars, tc.minPrice, tc.maxPrice, tc.currentPrice, width, height)

			lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
			if len(lines) != height {
				t.Fatalf("Expected %d lines, got %d", height, len(lines))
			}

			topLine := lines[0]
			if len(topLine) < 10 {
				t.Errorf("Top line too short")
			}
			maxLabel := topLine[:9]
			if maxLabel[0] == ' ' && maxLabel[8] == ' ' {
				t.Errorf("Expected non-space max label")
			}

			upperQLine := lines[height/4]
			if len(upperQLine) >= 10 {
				upperQLabel := upperQLine[:9]
				if upperQLabel[0] == ' ' && upperQLabel[8] == ' ' {
					t.Errorf("Expected non-space upper quartile label")
				}
			}

			lowerQLine := lines[(height*3)/4]
			if len(lowerQLine) >= 10 {
				lowerQLabel := lowerQLine[:9]
				if lowerQLabel[0] == ' ' && lowerQLabel[8] == ' ' {
					t.Errorf("Expected non-space lower quartile label")
				}
			}

			bottomLine := lines[height-1]
			if len(bottomLine) >= 10 {
				minLabel := bottomLine[:9]
				if minLabel[0] == ' ' && minLabel[8] == ' ' {
					t.Errorf("Expected non-space min label")
				}
			}
		})
	}
}
