// Package chart provides rendering functions for candlestick charts and TPI overlays.
package chart

import (
	"math"

	"github.com/charmbracelet/lipgloss"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// renderCandles renders cell-based candlesticks for the price pane.
func renderCandles(bars []indicators.OHLCV, width, height int, _ string) string {
	return renderCandlesWithReferences(bars, width, height, "", nil)
}

// renderCandlesWithReferences renders cell-based candlesticks with optional reference lines.
// Reference lines are drawn first, then candles overwrite them where they overlap.
func renderCandlesWithReferences(bars []indicators.OHLCV, width, height int, _ string, references []float64) string {
	if len(bars) == 0 || height <= 0 {
		return ""
	}

	// Calculate price range
	minPrice, maxPrice := getPriceRange(bars)
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}

	// Create grid for rendering (height rows, width columns)
	grid := make([][]string, height)
	for i := range grid {
		grid[i] = make([]string, width)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// Render reference lines first (so candles overwrite them where they overlap)
	mutedColor := lipgloss.Color("#6272A4")
	for _, refPrice := range references {
		refPos := int(float64(height-1) * (1 - (refPrice-minPrice)/priceRange))
		refPos = clamp(refPos, 0, height-1)
		// Draw dashed reference line: · · · · · pattern
		for x := 0; x < width; x++ {
			if x%2 == 0 {
				grid[refPos][x] = lipgloss.NewStyle().Foreground(mutedColor).Render("·")
			}
		}
	}

	// Calculate bar width (number of columns per bar)
	barWidth := width / len(bars)
	if barWidth < 1 {
		barWidth = 1
	}

	// Get colors - use predefined colors
	bullishColor := lipgloss.Color("#50FA7B")
	bearishColor := lipgloss.Color("#FF5555")

	// Render each bar
	for i, bar := range bars {
		x := i * barWidth
		if x >= width {
			break
		}

		// Calculate positions (0 at top, height-1 at bottom)
		highPos := int(float64(height-1) * (1 - (bar.High-minPrice)/priceRange))
		lowPos := int(float64(height-1) * (1 - (bar.Low-minPrice)/priceRange))
		openPos := int(float64(height-1) * (1 - (bar.Open-minPrice)/priceRange))
		closePos := int(float64(height-1) * (1 - (bar.Close-minPrice)/priceRange))

		// Clamp positions
		highPos = clamp(highPos, 0, height-1)
		lowPos = clamp(lowPos, 0, height-1)
		openPos = clamp(openPos, 0, height-1)
		closePos = clamp(closePos, 0, height-1)

		// Determine bar direction
		isBullish := bar.Close >= bar.Open

		// Determine body character
		bodyChar := "█"
		if math.Abs(bar.Close-bar.Open) < priceRange/float64(height) {
			bodyChar = "─" // Doji
		}

		// Select color
		var barColor lipgloss.Color
		if isBullish {
			barColor = bullishColor
		} else {
			barColor = bearishColor
		}

		// Render wicks
		for y := minInt(highPos, lowPos); y <= maxInt(highPos, lowPos); y++ {
			if x < width {
				grid[y][x] = lipgloss.NewStyle().Foreground(barColor).Render("│")
			}
		}

		// Render body
		bodyTop := minInt(openPos, closePos)
		bodyBottom := maxInt(openPos, closePos)
		for y := bodyTop; y <= bodyBottom; y++ {
			if x < width {
				grid[y][x] = lipgloss.NewStyle().Foreground(barColor).Render(bodyChar)
			}
		}
	}

	// Build result string
	result := ""
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			result += grid[y][x]
		}
		result += "\n"
	}

	return result
}

// renderTPI renders the TPI line using braille canvas.
func renderTPI(scores []float64, width, height int, themeName string) string {
	if len(scores) == 0 || height <= 0 {
		return ""
	}

	// Get colors - use predefined colors
	mutedColor := lipgloss.Color("#6272A4")
	bullishColor := lipgloss.Color("#50FA7B")
	bearishColor := lipgloss.Color("#FF5555")

	// Normalize TPI scores to -1 to +1 range
	minTPI := -1.0
	maxTPI := 1.0

	// Create canvas
	canvas := components.NewCanvas(width, height)

	// Calculate zero line position (middle of canvas)
	zeroY := (height * 4) / 2 // Convert to pixel space (4 pixels per cell)

	// Draw zero reference line (dashed)
	for x := 0; x < width*2; x += 4 {
		canvas.Set(x, zeroY, mutedColor)
		if x+1 < width*2 {
			canvas.Set(x+1, zeroY, mutedColor)
		}
	}

	// Calculate pixel positions for each TPI value
	if len(scores) > 1 {
		prevX := 0
		prevY := tpiToPixelY(scores[0], minTPI, maxTPI, height)

		for i := 1; i < len(scores); i++ {
			x := int(float64(i) / float64(len(scores)-1) * float64(width*2))
			y := tpiToPixelY(scores[i], minTPI, maxTPI, height)

			// Draw line segment
			color := getTPIColor(scores[i], bullishColor, bearishColor, mutedColor)
			canvas.Line(prevX, prevY, x, y, color)

			// Fill below the line
			fillBelowZero(canvas, prevX, prevY, x, y, zeroY, scores[i], bullishColor, bearishColor, mutedColor, themeName)

			prevX = x
			prevY = y
		}
	}

	return canvas.Render()
}

// tpiToPixelY converts a TPI value to a pixel Y coordinate.
func tpiToPixelY(tpi, minTPI, maxTPI float64, height int) int {
	// Normalize TPI to 0-1 range
	normalized := (tpi - minTPI) / (maxTPI - minTPI)
	if normalized < 0 {
		normalized = 0
	} else if normalized > 1 {
		normalized = 1
	}

	// Convert to pixel space (inverted because Y=0 is at top)
	pixelY := int(float64(height*4) * (1 - normalized))
	return clamp(pixelY, 0, height*4-1)
}

// getTPIColor returns the color for a TPI value.
func getTPIColor(tpi float64, bullish, bearish, muted lipgloss.Color) lipgloss.Color {
	// Dead zone: muted gray between ±0.05
	if math.Abs(tpi) < 0.05 {
		return muted
	}

	// Green for positive, red for negative
	if tpi > 0 {
		return bullish
	}
	return bearish
}

// fillBelowZero fills the area below the TPI line down to the zero line.
// Uses solid fill for colorblind theme, otherwise uses color.
func fillBelowZero(canvas *components.Canvas, x0, y0, x1, y1, zeroY int, tpi float64, bullish, bearish, muted lipgloss.Color, themeName string) {
	// Determine fill color
	fillColor := getTPIColor(tpi, bullish, bearish, muted)

	// For each X in the line segment, fill from the line to zero
	for x := x0; x <= x1; x++ {
		// Interpolate Y at this X
		var y int
		if x1 == x0 {
			y = y0
		} else {
			t := float64(x-x0) / float64(x1-x0)
			y = int(float64(y0) + t*float64(y1-y0))
		}

		// Fill from y to zeroY
		startY := minInt(y, zeroY)
		endY := maxInt(y, zeroY)

		for py := startY; py <= endY; py++ {
			if themeName == "colorblind" {
				// Use dashed fill pattern for colorblind theme
				if (x+py)%2 == 0 {
					canvas.Set(x, py, fillColor)
				}
			} else {
				// Use solid fill for other themes
				canvas.Set(x, py, fillColor)
			}
		}
	}
}

// clamp returns value clamped to [minVal, maxVal].
func clamp(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
