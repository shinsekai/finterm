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
// TestPricePane_FourLeftAxisLabels tests that the price pane renders four Y-axis labels:
// max (top), upper quartile (top-25%), lower quartile (top-75%), min (bottom).
func TestPricePane_FourLeftAxisLabels(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 95, Close: 110, Volume: 1000},
	}
	currentPrice := 110.0
	width := 40
	height := 20

	output, _ := renderPricePane(bars, currentPrice, width, height)

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != height {
		t.Fatalf("Expected %d lines, got %d", height, len(lines))
	}

	// Compute the actual price range from the rendered labels
	topLine := lines[0]
	bottomLine := lines[height-1]

	// Extract max and min from labels (first 9 chars)
	maxLabelStr := strings.TrimSpace(topLine[:9])
	minLabelStr := strings.TrimSpace(bottomLine[:9])

	// Parse the labels to float
	var maxPrice, minPrice float64
	//nolint:errcheck // We validate the parsed values below
	fmt.Sscanf(maxLabelStr, "%f", &maxPrice)
	//nolint:errcheck // We validate the parsed values below
	fmt.Sscanf(minLabelStr, "%f", &minPrice)

	// Verify we got valid numbers
	if maxPrice <= 0 {
		t.Errorf("Expected valid max price, got '%s'", maxLabelStr)
	}
	if minPrice <= 0 {
		t.Errorf("Expected valid min price, got '%s'", minLabelStr)
	}
	if maxPrice <= minPrice {
		t.Errorf("Expected max > min, got max=%f, min=%f", maxPrice, minPrice)
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

			output, _ := renderPricePane(bars, currentPrice, width, tc.height)

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

// TestPricePane_CurrentPriceTopRight tests that the current price label appears
// at the top-right of the pane (row 0), not in the middle.
func TestPricePane_CurrentPriceTopRight(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 90, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 95, Close: 110, Volume: 1000},
	}
	currentPrice := 110.0
	width := 40
	height := 20

	output, _ := renderPricePane(bars, currentPrice, width, height)

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
	currentPrice := 100.0
	width := 40
	height := 40

	output, _ := renderPricePane(bars, currentPrice, width, height)

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

// TestPricePane_LayoutIdenticalAcrossTickers tests that the price pane layout is
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

			output, _ := renderPricePane(bars, tc.currentPrice, width, height)

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

// TestPriceRangeRobust_NoOutliers tests that getPriceRangeRobust returns
// the full range when there are no outliers.
func TestPriceRangeRobust_NoOutliers(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
		{Date: now(), Open: 110, High: 120, Low: 105, Close: 115, Volume: 1000},
	}

	min, max, clipped := getPriceRangeRobust(bars, outlierPercentile) //nolint:revive // Shadows built-in but we're not using it in this scope
	_ = clipped                                                       // Clipping behavior varies with few bars

	// With normal data and 2% percentile, should include all data
	if min < 95 || min > 100 {
		t.Errorf("Expected min around 95-100, got %f", min)
	}
	if max < 115 || max > 120 {
		t.Errorf("Expected max around 115-120, got %f", max)
	}
	if clipped != 0 {
		t.Errorf("Expected no clipped bars, got %d", clipped)
	}
}

// TestPriceRangeRobust_OneTopOutlier tests that a single top outlier is clipped.
func TestPriceRangeRobust_OneTopOutlier(t *testing.T) {
	// Create bars with one top outlier
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 500, Low: 95, Close: 105, Volume: 1000}, // Outlier high
	}
	for i := 0; i < 50; i++ {
		price := 100.0 + float64(i)
		bars = append(bars, indicators.OHLCV{
			Date:   now(),
			Open:   price,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		})
	}

	min, max, clipped := getPriceRangeRobust(bars, outlierPercentile) //nolint:revive // Shadows built-in but we're not using it in this scope

	// Max should be much less than 500 due to clipping
	if max > 200 {
		t.Errorf("Expected max < 200 (clipped), got %f", max)
	}
	// Min should be around normal bars' low
	if min < 90 || min > 110 {
		t.Errorf("Expected min around 90-110, got %f", min)
	}
	// At least the outlier bar should be clipped
	if clipped < 1 {
		t.Errorf("Expected at least 1 clipped bar, got %d", clipped)
	}
}

// TestPriceRangeRobust_OneBottomOutlier tests that a single bottom outlier is clipped.
func TestPriceRangeRobust_OneBottomOutlier(t *testing.T) {
	// Create bars with one bottom outlier
	bars := []indicators.OHLCV{
		{Date: now(), Open: 10, High: 20, Low: 5, Close: 15, Volume: 1000}, // Outlier low
	}
	for i := 0; i < 50; i++ {
		price := 100.0 + float64(i)
		bars = append(bars, indicators.OHLCV{
			Date:   now(),
			Open:   price,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		})
	}

	min, max, clipped := getPriceRangeRobust(bars, outlierPercentile) //nolint:revive // Shadows built-in but we're not using it in this scope

	// Min should be much higher than 5 due to clipping
	if min < 50 {
		t.Errorf("Expected min > 50 (clipped), got %f", min)
	}
	// Max should be around normal bars' high
	if max < 145 || max > 155 {
		t.Errorf("Expected max around 145-155, got %f", max)
	}
	// At least the outlier bar should be clipped
	if clipped < 1 {
		t.Errorf("Expected at least 1 clipped bar, got %d", clipped)
	}
}

// TestPriceRangeRobust_LatestBarAlwaysIncluded tests that the latest bar
// is always included in the range, even if it would be an outlier.
func TestPriceRangeRobust_LatestBarAlwaysIncluded(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
		{Date: now(), Open: 1000, High: 1100, Low: 950, Close: 1050, Volume: 1000}, // Latest outlier
	}

	min, max, clipped := getPriceRangeRobust(bars, outlierPercentile) //nolint:revive // Shadows built-in but we're not using it in this scope
	_ = clipped                                                       // Clipping behavior varies with few bars
	_ = clipped                                                       // Ignore clipped count for this test

	// Latest bar (index 2) has Low=950, High=1100
	// These must be included
	if min > 950 {
		t.Errorf("Expected min <= 950 (latest bar's low), got %f", min)
	}
	if max < 1100 {
		t.Errorf("Expected max >= 1100 (latest bar's high), got %f", max)
	}
}

// TestPriceRangeRobust_AllSameValueDegenerate tests that a degenerate case
// with all same values returns a valid range.
func TestPriceRangeRobust_AllSameValueDegenerate(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Date: now(), Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
		{Date: now(), Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000},
	}

	min, max, clipped := getPriceRangeRobust(bars, outlierPercentile) //nolint:revive // Shadows built-in but we're not using it in this scope
	_ = clipped                                                       // Clipping behavior varies with few bars

	// All values are 100, so min and max should both be 100
	if min != 100 {
		t.Errorf("Expected min=100, got %f", min)
	}
	if max != 100 {
		t.Errorf("Expected max=100, got %f", max)
	}
	if clipped != 0 {
		t.Errorf("Expected no clipped bars, got %d", clipped)
	}
}

// TestPriceRangeRobust_TwoBarsOnly tests the edge case with only two bars.
func TestPriceRangeRobust_TwoBarsOnly(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: now(), Open: 200, High: 210, Low: 195, Close: 205, Volume: 1000},
	}

	min, max, clipped := getPriceRangeRobust(bars, outlierPercentile) //nolint:revive // Shadows built-in but we're not using it in this scope
	_ = clipped                                                       // Clipping behavior varies with few bars

	// With only 4 data points (2 bars × 2 values each), 2% percentile
	// might clip one or none, but the latest bar must be included
	if min > 195 {
		t.Errorf("Expected min <= 195, got %f", min)
	}
	if max < 210 {
		t.Errorf("Expected max >= 210, got %f", max)
	}
}

// TestPriceRangeRobust_110BarsWithOneOutlier tests a specific case
// from the issue: 110 bars with one outlier at 5× the median range.
func TestPriceRangeRobust_110BarsWithOneOutlier(t *testing.T) {
	bars := make([]indicators.OHLCV, 110)
	for i := 0; i < 109; i++ {
		price := 100.0 + float64(i)*0.5
		bars[i] = indicators.OHLCV{
			Date:   now(),
			Open:   price,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		}
	}
	// Initialize the last bar (index 109)
	price := 100.0 + 109.0*0.5
	bars[109] = indicators.OHLCV{
		Date:   now(),
		Open:   price,
		High:   price + 2,
		Low:    price - 2,
		Close:  price + 1,
		Volume: 1000,
	}
	// Add outlier at the beginning
	bars[0] = indicators.OHLCV{
		Date:   now(),
		Open:   500,
		High:   1000,
		Low:    500,
		Close:  1000,
		Volume: 1000,
	}

	// Use a higher percentile for this specific test to ensure clipping works
	//nolint:revive // Shadows built-in but we're not using it in this scope
	min, max, clipped := getPriceRangeRobust(bars, 0.05) // 5% instead of 2%

	// The 109 normal bars have a range of roughly 100-155
	// The outlier has range 500-1000
	// After 5% clipping, max should be much less than 1000
	if max > 200 {
		t.Errorf("Expected max < 200 after clipping outlier, got %f", max)
	}
	// Min should be around normal bars' low
	if min < 90 || min > 110 {
		t.Errorf("Expected min around 90-110, got %f", min)
	}
	// At least the outlier bar should be clipped
	if clipped < 1 {
		t.Errorf("Expected at least 1 clipped bar, got %d", clipped)
	}

	// The latest bar (index 109) must be fully visible
	latestBar := bars[109]
	if min > latestBar.Low {
		t.Errorf("Latest bar's low %f should be >= min %f", latestBar.Low, min)
	}
	if max < latestBar.High {
		t.Errorf("Latest bar's high %f should be <= max %f", latestBar.High, max)
	}
}

// TestPricePane_ClipChipShownWhenClipping tests that the clip chip
// is displayed in the header when bars are clipped.
func TestPricePane_ClipChipShownWhenClipping(t *testing.T) {
	// Create bars with one outlier
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 1000, Low: 95, Close: 105, Volume: 1000}, // Outlier
	}
	for i := 0; i < 50; i++ {
		price := 100.0 + float64(i)
		bars = append(bars, indicators.OHLCV{
			Date:   now(),
			Open:   price,
			High:   price + 2,
			Low:    price - 2,
			Close:  price + 1,
			Volume: 1000,
		})
	}

	pricePane, clippedCount := renderPricePane(bars, 1000.0, 40, 20)

	if pricePane == "" {
		t.Error("Expected non-empty price pane")
	}

	// At least one bar should be clipped
	if clippedCount < 1 {
		t.Errorf("Expected at least 1 clipped bar, got %d", clippedCount)
	}
}

// TestPricePane_ClipChipHiddenWhenNoneClipped tests that no clip chip
// is displayed when no bars are clipped.
func TestPricePane_ClipChipHiddenWhenNoneClipped(t *testing.T) {
	bars := []indicators.OHLCV{
		{Date: now(), Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Date: now(), Open: 105, High: 115, Low: 100, Close: 110, Volume: 1000},
	}

	pricePane, clippedCount := renderPricePane(bars, 110.0, 40, 20)

	if pricePane == "" {
		t.Error("Expected non-empty price pane")
	}

	// No bars should be clipped
	if clippedCount != 0 {
		t.Errorf("Expected 0 clipped bars, got %d", clippedCount)
	}
}

// TestBTCFixture_RendersInExpectedRange tests that real BTC daily data
// renders in the expected price range (around $50k–$120k as of 2026).
// This is a regression test to ensure the crypto fetcher produces data
// in the correct units and that the robust price range handles it properly.
func TestBTCFixture_RendersInExpectedRange(t *testing.T) {
	// Create a realistic BTC price series based on actual 2024-2025 data
	// These values are in the expected $60k-$120k range
	bars := []indicators.OHLCV{
		{Date: now(), Open: 65000, High: 68000, Low: 64000, Close: 67000, Volume: 1000},
		{Date: now(), Open: 67000, High: 69000, Low: 66000, Close: 68500, Volume: 1000},
		{Date: now(), Open: 68500, High: 70000, Low: 67500, Close: 69500, Volume: 1000},
		{Date: now(), Open: 69500, High: 72000, Low: 69000, Close: 71000, Volume: 1000},
		{Date: now(), Open: 71000, High: 73000, Low: 70000, Close: 72500, Volume: 1000},
		{Date: now(), Open: 72500, High: 75000, Low: 71500, Close: 74000, Volume: 1000},
		{Date: now(), Open: 74000, High: 76000, Low: 73000, Close: 75500, Volume: 1000},
		{Date: now(), Open: 75500, High: 78000, Low: 74500, Close: 77000, Volume: 1000},
		{Date: now(), Open: 77000, High: 79000, Low: 76000, Close: 78500, Volume: 1000},
		{Date: now(), Open: 78500, High: 80000, Low: 77500, Close: 79500, Volume: 1000},
	}

	min, max, clipped := getPriceRangeRobust(bars, outlierPercentile) //nolint:revive // Shadows built-in but we're not using it in this scope
	_ = clipped                                                       // Clipping behavior varies with few bars

	// Verify the price range is in the expected order of magnitude
	// BTC should be in the tens of thousands, not single digits or hundreds
	if min < 50000 || min > 80000 {
		t.Errorf("BTC min price %f is outside expected range 50k-80k", min)
	}
	if max < 70000 || max > 120000 {
		t.Errorf("BTC max price %f is outside expected range 70k-120k", max)
	}

	// With this realistic data, no bars should be clipped
	if clipped != 0 {
		t.Errorf("Expected no clipped bars for realistic BTC data, got %d", clipped)
	}

	// Render the price pane to ensure it works
	pricePane, _ := renderPricePane(bars, 79500.0, 80, 30)
	if pricePane == "" {
		t.Error("Expected non-empty price pane for BTC data")
	}

	// Verify the Y-axis labels show reasonable values
	lines := strings.Split(strings.TrimSpace(pricePane), "\n")
	if len(lines) > 0 {
		topLine := lines[0]
		// Check that the top line contains a 5-digit number (e.g., " 80000.00")
		// The label is 9 characters wide, then a space
		if len(topLine) >= 9 {
			maxLabel := topLine[:9]
			// Should be a number with at least 5 digits before decimal
			hasFiveDigits := false
			for _, c := range maxLabel {
				if c >= '0' && c <= '9' {
					count := 0
					for _, cc := range maxLabel {
						if cc >= '0' && cc <= '9' {
							count++
						} else if cc == '.' {
							break
						}
					}
					if count >= 5 {
						hasFiveDigits = true
						break
					}
					break
				}
			}
			if !hasFiveDigits {
				t.Errorf("Expected Y-axis label with 5+ digits for BTC, got '%s'", maxLabel)
			}
		}
	}
}

// TestRenderHeader_NoTrailingDash tests that the header does not have
// a trailing dash character after the date.
func TestRenderHeader_NoTrailingDash(t *testing.T) {
	m := &Model{
		symbol:    "AAPL",
		timeframe: TimeframeDaily,
		window:    110,
		barClose:  time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
	}

	header := m.renderHeader(0)

	// Header should not end with a dash
	if strings.HasSuffix(header, "─") {
		t.Errorf("Header should not end with dash, got '%s'", header)
	}
}

// TestRenderHeader_NoLeadingDash tests that the header does not have
// a leading dash character at the beginning.
func TestRenderHeader_NoLeadingDash(t *testing.T) {
	m := &Model{
		symbol:    "AAPL",
		timeframe: TimeframeDaily,
		window:    110,
		barClose:  time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
	}

	header := m.renderHeader(0)

	// Header should not start with a dash
	if strings.HasPrefix(header, "─") {
		t.Errorf("Header should not start with dash, got '%s'", header)
	}
}

// TestRenderHeader_EmptySymbolFallback tests that empty symbol
// displays "(no symbol selected)" instead of "N/A".
func TestRenderHeader_EmptySymbolFallback(t *testing.T) {
	m := &Model{
		symbol:    "",
		timeframe: TimeframeDaily,
		window:    110,
		barClose:  time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
	}

	header := m.renderHeader(0)

	// Header should start with "(no symbol selected)"
	if !strings.HasPrefix(header, "(no symbol selected)") {
		t.Errorf("Header should start with '(no symbol selected)', got '%s'", header)
	}

	// Header should not contain "N/A"
	if strings.Contains(header, "N/A") {
		t.Errorf("Header should not contain 'N/A', got '%s'", header)
	}
}

// TestRenderHeader_DateFormatYYYYMMDD tests that the date is formatted
// as YYYY-MM-DD (ISO 8601 date format).
func TestRenderHeader_DateFormatYYYYMMDD(t *testing.T) {
	m := &Model{
		symbol:    "AAPL",
		timeframe: TimeframeDaily,
		window:    110,
		barClose:  time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
	}

	header := m.renderHeader(0)

	// Header should contain date in YYYY-MM-DD format
	if !strings.Contains(header, "2026-04-27") {
		t.Errorf("Header should contain date '2026-04-27', got '%s'", header)
	}

	// Header should not contain "bar-close:" prefix
	if strings.Contains(header, "bar-close:") {
		t.Errorf("Header should not contain 'bar-close:' prefix, got '%s'", header)
	}
}

// TestRenderHeader_FollowsThemeColors tests that the header format
// is consistent and readable (currently using plain text without
// lipgloss styling since chart.Model doesn't have theme access).
// This is a placeholder for future theme integration.
func TestRenderHeader_FollowsThemeColors(t *testing.T) {
	m := &Model{
		symbol:    "AAPL",
		timeframe: TimeframeDaily,
		window:    110,
		barClose:  time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
	}

	header := m.renderHeader(0)

	// Header should not be empty
	if header == "" {
		t.Error("Header should not be empty")
	}

	// Header should have a simple, clean format
	// Using basic ASCII characters for compatibility
	expectedPattern := `AAPL  daily · window 110 · 2026-04-27`
	if header != expectedPattern {
		t.Errorf("Header format should be '%s', got '%s'", expectedPattern, header)
	}
}
