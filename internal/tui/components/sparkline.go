// Package components provides reusable TUI components.
package components

import (
	"math"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines the interface for theme styling.
// This is defined here to avoid import cycle with the parent tui package.
type Theme interface {
	Bullish() lipgloss.Style
	Bearish() lipgloss.Style
	Neutral() lipgloss.Style
}

// sparklineBlocks are Unicode block characters used for sparkline rendering.
// Indexed from 0 (lowest) to 7 (highest).
var sparklineBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderSparkline renders a sparkline from a slice of float64 values.
//
// Parameters:
//   - values: price history data, oldest first
//   - width: exact visible width of the output (number of runes)
//   - theme: theme for color selection
//
// Returns a string of exactly `width` visible runes (excluding ANSI color codes).
//
// Color rules:
//   - values[len-1] > values[0] → bullish color
//   - values[len-1] < values[0] → bearish color
//   - values[len-1] == values[0] → neutral color
//
// Edge cases:
//   - len(values) == 0 → returns width spaces
//   - len(values) == 1 → renders a single mid-block padded to width
//   - all values are NaN → renders blanks
func RenderSparkline(values []float64, width int, theme Theme) string {
	if width <= 0 {
		return ""
	}

	if len(values) == 0 {
		return repeatSpaces(width)
	}

	// Filter out NaN values for min/max calculation
	filtered := filterNaN(values)
	if len(filtered) == 0 {
		return repeatSpaces(width)
	}

	minVal, maxVal := minMax(filtered)

	// If all values are equal, render as mid-blocks
	if minVal == maxVal {
		block := string(sparklineBlocks[3]) // mid-block
		style := colorForNetDirection(values, theme)
		return style.Render(repeatString(block, width))
	}

	// Map each value to a block character
	blocks := make([]rune, 0, len(values))
	for _, v := range values {
		blocks = append(blocks, valueToBlock(v, minVal, maxVal))
	}

	// Ensure exact width:
	// - If values is shorter than width: right-pad with spaces (blocks on left)
	// - If values is longer than width: truncate from the left (keep newest data)
	sparkline := toExactWidth(string(blocks), width)

	// Apply foreground-only color for direction
	style := colorForNetDirection(values, theme)
	return style.Render(sparkline)
}

// valueToBlock maps a value to a sparkline block character.
// Linearly maps the value from [minVal, maxVal] to [0, 7].
// NaN values return space.
func valueToBlock(value, minVal, maxVal float64) rune {
	if math.IsNaN(value) {
		return ' '
	}

	if minVal == maxVal {
		return sparklineBlocks[3] // mid-block
	}

	// Normalize to [0, 1] and map to [0, 7]
	ratio := (value - minVal) / (maxVal - minVal)
	index := int(ratio * 7)

	// Clamp to valid range
	if index < 0 {
		index = 0
	} else if index > 7 {
		index = 7
	}

	return sparklineBlocks[index]
}

// colorForNetDirection returns the appropriate style based on net price direction.
// Foreground-only styling to preserve table row background striping.
func colorForNetDirection(values []float64, theme Theme) lipgloss.Style {
	if len(values) < 2 {
		return theme.Neutral()
	}

	first := values[0]
	last := values[len(values)-1]

	// Compare first and last values (skipping NaN)
	if math.IsNaN(first) || math.IsNaN(last) {
		return theme.Neutral()
	}

	if last > first {
		return theme.Bullish()
	} else if last < first {
		return theme.Bearish()
	}
	return theme.Neutral()
}

// filterNaN returns a new slice excluding NaN values.
func filterNaN(values []float64) []float64 {
	filtered := make([]float64, 0, len(values))
	for _, v := range values {
		if !math.IsNaN(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// minMax returns the minimum and maximum values in a slice.
// Panics if slice is empty.
func minMax(values []float64) (minVal, maxVal float64) {
	if len(values) == 0 {
		panic("minMax called on empty slice")
	}

	minVal = values[0]
	maxVal = values[0]

	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	return minVal, maxVal
}

// repeatSpaces returns a string of n space characters.
func repeatSpaces(n int) string {
	return repeatString(" ", n)
}

// repeatString returns s repeated n times.
func repeatString(s string, n int) string {
	result := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		result = append(result, s...)
	}
	return string(result)
}

// toExactWidth ensures the sparkline string has exactly width visible runes.
// If shorter, right-pads with spaces (blocks on left). If longer, truncates from the left.
func toExactWidth(s string, width int) string {
	runes := []rune(s)
	runesLen := len(runes)

	switch {
	case runesLen == width:
		return s
	case runesLen < width:
		// Right-pad with spaces (blocks on left)
		padding := width - runesLen
		return s + repeatSpaces(padding)
	default:
		// Truncate from the left (keep newest data on the right)
		return string(runes[runesLen-width:])
	}
}
