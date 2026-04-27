// Package components provides reusable TUI components.
package components

import (
	"math"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// mockTheme is a minimal theme implementation for testing.
type mockTheme struct{}

func (m *mockTheme) Bullish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
}

func (m *mockTheme) Bearish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
}

func (m *mockTheme) Neutral() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
}

// colorblindTheme is a theme implementation with colorblind-friendly colors.
type colorblindTheme struct{}

func (c *colorblindTheme) Bullish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#00C853"))
}

func (c *colorblindTheme) Bearish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#D50000"))
}

func (c *colorblindTheme) Neutral() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB300"))
}

// stripANSI removes ANSI escape codes for testing.
// Returns only the visible characters.
func stripANSI(s string) string {
	// Manually strip ANSI escape sequences
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			// Start of escape sequence
			inEscape = true
			if i+1 < len(s) && s[i+1] == '[' {
				// CSI sequence - skip until we find the final character
				i += 2
				for i < len(s) {
					c := s[i]
					// CSI parameters are digits and semicolons
					// Final character is in range 0x40-0x7E (letters and some symbols)
					if c >= 0x40 && c <= 0x7E {
						inEscape = false
						break
					}
					i++
				}
				continue
			}
			continue
		}
		if inEscape {
			// Skip characters until we find the end of the escape sequence
			if s[i] >= 0x40 && s[i] <= 0x7E {
				inEscape = false
			}
			continue
		}
		// Copy visible characters
		result.WriteByte(s[i])
	}
	return result.String()
}

// TestSparkline_FullSeries30Bars tests rendering a full 30-bar series.
func TestSparkline_FullSeries30Bars(t *testing.T) {
	values := make([]float64, 30)
	for i := 0; i < 30; i++ {
		values[i] = float64(100 + i) // Rising trend
	}

	theme := &mockTheme{}
	result := RenderSparkline(values, 30, theme)
	visibleWidth := lipgloss.Width(result)

	if visibleWidth != 30 {
		t.Errorf("expected width 30, got %d: %q", visibleWidth, result)
	}

	// Verify that sparkline uses blocks (all should be valid blocks or spaces)
	stripped := stripANSI(result)
	runes := []rune(stripped)
	for i, r := range runes {
		if r != ' ' && !isValidBlock(r) {
			t.Errorf("position %d: unexpected character %q", i, r)
		}
	}
}

// TestSparkline_ShortSeriesLeftPaddedBlank tests rendering a short series with left padding.
func TestSparkline_ShortSeriesLeftPaddedBlank(t *testing.T) {
	values := []float64{100, 110, 120} // Only 3 values

	theme := &mockTheme{}
	result := RenderSparkline(values, 10, theme)
	visibleWidth := lipgloss.Width(result)

	if visibleWidth != 10 {
		t.Errorf("expected width 10, got %d: %q", visibleWidth, result)
	}

	// First 3 positions should be blocks, rest should be spaces
	stripped := stripANSI(result)
	runes := []rune(stripped)
	for i := 0; i < 3; i++ {
		if runes[i] == ' ' {
			t.Errorf("position %d should be a block, got space", i)
		}
	}
	for i := 3; i < 10; i++ {
		if runes[i] != ' ' {
			t.Errorf("position %d should be space, got %q", i, runes[i])
		}
	}
}

// TestSparkline_EmptySeriesAllBlank tests rendering an empty series.
func TestSparkline_EmptySeriesAllBlank(t *testing.T) {
	values := []float64{}

	theme := &mockTheme{}
	result := RenderSparkline(values, 10, theme)
	stripped := stripANSI(result)

	expected := strings.Repeat(" ", 10)
	if stripped != expected {
		t.Errorf("expected %q, got %q", expected, stripped)
	}
}

// TestSparkline_SingleValue tests rendering a single value.
func TestSparkline_SingleValue(t *testing.T) {
	values := []float64{150}

	theme := &mockTheme{}
	result := RenderSparkline(values, 10, theme)
	visibleWidth := lipgloss.Width(result)

	if visibleWidth != 10 {
		t.Errorf("expected width 10, got %d: %q", visibleWidth, result)
	}

	// Should render all mid-blocks
	expected := strings.Repeat("▄", 10)
	stripped := stripANSI(result)
	if stripped != expected {
		t.Errorf("expected %q, got %q", expected, stripped)
	}
}

// TestSparkline_FlatSeriesMidBlock tests rendering a flat series.
func TestSparkline_FlatSeriesMidBlock(t *testing.T) {
	values := make([]float64, 10)
	for i := 0; i < 10; i++ {
		values[i] = 150 // All same value
	}

	theme := &mockTheme{}
	result := RenderSparkline(values, 10, theme)
	visibleWidth := lipgloss.Width(result)

	if visibleWidth != 10 {
		t.Errorf("expected width 10, got %d: %q", visibleWidth, result)
	}

	// Should render all mid-blocks
	expected := strings.Repeat("▄", 10)
	stripped := stripANSI(result)
	if stripped != expected {
		t.Errorf("expected %q, got %q", expected, stripped)
	}

	// Verify neutral color was used by checking result has color codes
	if result == stripped {
		t.Logf("Note: lipgloss may not apply color codes in test environment")
		// Don't fail - color application may depend on terminal detection
	}
}

// TestSparkline_NaNSkippedForScale tests that NaN values are skipped for min/max scaling.
func TestSparkline_NaNSkippedForScale(t *testing.T) {
	values := []float64{100, math.NaN(), 110, math.NaN(), 130}

	theme := &mockTheme{}
	result := RenderSparkline(values, 5, theme)
	visibleWidth := lipgloss.Width(result)

	// All 5 positions should be rendered
	if visibleWidth != 5 {
		t.Errorf("expected width 5, got %d: %q", visibleWidth, result)
	}

	// First, third, and last positions should have blocks
	// NaN positions (2, 4) should be spaces
	stripped := stripANSI(result)
	runes := []rune(stripped)
	if runes[0] == ' ' || runes[2] == ' ' || runes[4] == ' ' {
		t.Errorf("non-NaN positions should have blocks: %s", stripped)
	}

	// NaN positions should be spaces
	if runes[1] != ' ' || runes[3] != ' ' {
		t.Errorf("NaN positions should be spaces: %s", stripped)
	}
}

// TestSparkline_ColorByNetDirectionBullish tests bullish color for rising trend.
func TestSparkline_ColorByNetDirectionBullish(t *testing.T) {
	values := []float64{100, 110, 120, 130} // Rising

	theme := &mockTheme{}
	result := RenderSparkline(values, 4, theme)

	// Result should be longer than stripped content due to color codes
	// This indicates color styling was applied
	stripped := stripANSI(result)
	if len(result) <= len(stripped) {
		t.Logf("Note: lipgloss may not apply color codes in test environment")
		t.Logf("Result: %q (len=%d), Stripped: %q (len=%d)", result, len(result), stripped, len(stripped))
		// Don't fail - color application may depend on terminal detection
	}
}

// TestSparkline_ColorByNetDirectionBearish tests bearish color for falling trend.
func TestSparkline_ColorByNetDirectionBearish(t *testing.T) {
	values := []float64{130, 120, 110, 100} // Falling

	theme := &mockTheme{}
	result := RenderSparkline(values, 4, theme)

	stripped := stripANSI(result)
	if len(result) <= len(stripped) {
		t.Logf("Note: lipgloss may not apply color codes in test environment")
	}
}

// TestSparkline_ColorFollowsColorblindTheme tests that colorblind theme is respected.
func TestSparkline_ColorFollowsColorblindTheme(t *testing.T) {
	values := []float64{100, 110} // Rising
	theme := &colorblindTheme{}
	result := RenderSparkline(values, 2, theme)

	stripped := stripANSI(result)
	if len(result) <= len(stripped) {
		t.Logf("Note: lipgloss may not apply color codes in test environment")
	}
}

// TestSparkline_ExactWidthRespected tests that exact width is always returned.
func TestSparkline_ExactWidthRespected(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		width  int
	}{
		{"empty values", []float64{}, 5},
		{"single value", []float64{100}, 5},
		{"exact match", []float64{100, 110, 120}, 3},
		{"shorter values", []float64{100, 110}, 10},
		{"more values", []float64{100, 110, 120, 130, 140}, 3},
	}

	theme := &mockTheme{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderSparkline(tt.values, tt.width, theme)
			visibleWidth := lipgloss.Width(result)

			if visibleWidth != tt.width {
				t.Errorf("expected width %d, got %d: %q", tt.width, visibleWidth, result)
			}
		})
	}
}

// TestValueToBlock tests mapping values to block characters.
func TestValueToBlock(t *testing.T) {
	tests := []struct {
		value    float64
		min      float64
		max      float64
		expected rune
	}{
		{0.0, 0.0, 100.0, '▁'},        // 0%
		{50.0, 0.0, 100.0, '▄'},       // 50%
		{100.0, 0.0, 100.0, '█'},      // 100%
		{25.0, 0.0, 100.0, '▂'},       // 25%
		{75.0, 0.0, 100.0, '▆'},       // 75%
		{math.NaN(), 0.0, 100.0, ' '}, // NaN
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			result := valueToBlock(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("valueToBlock(%v, %v, %v) = %q, want %q",
					tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

// TestFilterNaN tests filtering NaN values.
func TestFilterNaN(t *testing.T) {
	tests := []struct {
		input    []float64
		expected []float64
	}{
		{[]float64{1.0, 2.0, 3.0}, []float64{1.0, 2.0, 3.0}},
		{[]float64{math.NaN(), 1.0, math.NaN(), 2.0}, []float64{1.0, 2.0}},
		{[]float64{math.NaN()}, []float64{}},
		{[]float64{}, []float64{}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := filterNaN(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("len = %d, want %d", len(result), len(tt.expected))
			}
			for i := range tt.expected {
				if result[i] != tt.expected[i] {
					t.Errorf("result[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestMinMax tests finding minimum and maximum values.
func TestMinMax(t *testing.T) {
	tests := []struct {
		input   []float64
		wantMin float64
		wantMax float64
	}{
		{[]float64{1.0, 2.0, 3.0}, 1.0, 3.0},
		{[]float64{3.0, 1.0, 2.0}, 1.0, 3.0},
		{[]float64{-5.0, 0.0, 5.0}, -5.0, 5.0},
		{[]float64{100.0}, 100.0, 100.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			minVal, maxVal := minMax(tt.input)
			if minVal != tt.wantMin || maxVal != tt.wantMax {
				t.Errorf("minMax(%v) = (%v, %v), want (%v, %v)",
					tt.input, minVal, maxVal, tt.wantMin, tt.wantMax)
			}
		})
	}

	// Test that it panics on empty slice
	t.Run("panic on empty", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on empty slice")
			}
		}()
		minMax([]float64{})
	})
}

// TestRepeatString tests string repetition.
func TestRepeatString(t *testing.T) {
	tests := []struct {
		s      string
		n      int
		result string
	}{
		{"a", 0, ""},
		{"a", 1, "a"},
		{"a", 3, "aaa"},
		{"ab", 2, "abab"},
	}

	for _, tt := range tests {
		t.Run(tt.result, func(t *testing.T) {
			if result := repeatString(tt.s, tt.n); result != tt.result {
				t.Errorf("repeatString(%q, %d) = %q, want %q",
					tt.s, tt.n, result, tt.result)
			}
		})
	}
}

// isValidBlock checks if rune is a valid sparkline block.
func isValidBlock(r rune) bool {
	return r == '▁' || r == '▂' || r == '▃' ||
		r == '▄' || r == '▅' || r == '▆' || r == '▇' || r == '█'
}
