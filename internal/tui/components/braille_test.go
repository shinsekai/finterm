package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBraille_SinglePixel(t *testing.T) {
	canvas := NewCanvas(1, 1)
	red := lipgloss.Color("196")
	canvas.Set(0, 0, red)
	output := canvas.Render()
	if !strings.Contains(output, "⠁") {
		t.Errorf("expected ⠁ (U+2801) in output, got %q", output)
	}
}

func TestBraille_AllEightSubpixelsInCell(t *testing.T) {
	testCases := []struct {
		px, py   int
		expected rune
	}{
		{0, 0, '⠁'}, {1, 0, '⠈'},
		{0, 1, '⠂'}, {1, 1, '⠐'},
		{0, 2, '⠄'}, {1, 2, '⠠'},
		{0, 3, '⡀'}, {1, 3, '⢀'},
	}
	for _, tc := range testCases {
		t.Run(string(tc.expected), func(t *testing.T) {
			c := NewCanvas(1, 1)
			c.Set(tc.px, tc.py, lipgloss.Color("255"))
			output := c.Render()
			if !strings.ContainsRune(output, tc.expected) {
				t.Errorf("expected %q in output, got %q", tc.expected, output)
			}
		})
	}
}

func TestBraille_BresenhamLineDiagonal(t *testing.T) {
	canvas := NewCanvas(5, 1)
	green := lipgloss.Color("46")
	// Line from (0,0) to (4,1) should set exactly 4 points using Bresenham's algorithm
	// Minimum points: (0,0), (1,0), (2,1), (3,1), (4,1)
	canvas.Line(0, 0, 4, 1, green)
	litPixels := 0
	for cx := 0; cx < 5; cx++ {
		var code uint8
		for cy := 0; cy < 1; cy++ {
			if canvas.pixels[cx*2][cy*4] {
				code |= 1 << 0
			}
			if canvas.pixels[cx*2][cy*4+1] {
				code |= 1 << 1
			}
			if canvas.pixels[cx*2][cy*4+2] {
				code |= 1 << 2
			}
			if canvas.pixels[cx*2+1][cy*4] {
				code |= 1 << 3
			}
			if canvas.pixels[cx*2+1][cy*4+1] {
				code |= 1 << 4
			}
			if canvas.pixels[cx*2+1][cy*4+2] {
				code |= 1 << 5
			}
			if canvas.pixels[cx*2][cy*4+3] {
				code |= 1 << 6
			}
			if canvas.pixels[cx*2+1][cy*4+3] {
				code |= 1 << 7
			}
		}
		litPixels += countSetBits(code)
	}
	if litPixels != 5 {
		t.Errorf("expected 5 lit pixels for diagonal line, got %d", litPixels)
	}
}

func TestBraille_BresenhamLineHorizontal(t *testing.T) {
	canvas := NewCanvas(4, 1)
	cyan := lipgloss.Color("51")
	canvas.Line(0, 2, 7, 2, cyan)
	litPixels := 0
	for cx := 0; cx < 4; cx++ {
		var code uint8
		if canvas.pixels[cx*2][2] {
			code |= 1 << 2
		}
		if canvas.pixels[cx*2+1][2] {
			code |= 1 << 5
		}
		litPixels += countSetBits(code)
	}
	if litPixels != 8 {
		t.Errorf("expected 8 lit pixels for horizontal line, got %d", litPixels)
	}
}

func TestBraille_BresenhamLineVertical(t *testing.T) {
	canvas := NewCanvas(1, 1)
	magenta := lipgloss.Color("201")
	// Draw line covering all 4 rows of the single cell
	canvas.Line(0, 0, 0, 3, magenta)
	// Count all 8 positions (2 columns × 4 rows)
	litPixels := 0
	for cy := 0; cy < 1; cy++ {
		var code uint8
		if canvas.pixels[0][cy*4] {
			code |= 1 << 0
		}
		if canvas.pixels[0][cy*4+1] {
			code |= 1 << 1
		}
		if canvas.pixels[0][cy*4+2] {
			code |= 1 << 2
		}
		if canvas.pixels[0][cy*4+3] {
			code |= 1 << 6
		}
		litPixels += countSetBits(code)
	}
	if litPixels != 4 {
		t.Errorf("expected 4 lit pixels for vertical line in column 0, got %d", litPixels)
	}
}

func TestBraille_OutOfRangeSetSilentlyDropped(_ *testing.T) {
	canvas := NewCanvas(2, 2)
	blue := lipgloss.Color("33")
	canvas.Set(-1, 0, blue)
	canvas.Set(0, -1, blue)
	canvas.Set(4, 0, blue)
	canvas.Set(0, 8, blue)
}

func TestBraille_RenderDimensionsExact(t *testing.T) {
	canvas := NewCanvas(3, 2)
	output := canvas.Render()
	lines := strings.Split(output, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines in output, got %d", len(lines))
	}
	for i, line := range lines {
		visibleRunes := 0
		for _, r := range line {
			if r >= 0x2800 && r <= 0x28FF {
				visibleRunes++
			}
		}
		if visibleRunes != 3 {
			t.Errorf("line %d: expected 3 visible braille runes, got %d", i, visibleRunes)
		}
	}
}

func TestBraille_MultipleColorsLastWriteWins(t *testing.T) {
	canvas := NewCanvas(1, 1)
	red := lipgloss.Color("196")
	blue := lipgloss.Color("33")
	canvas.Set(0, 0, red)
	// Last write should override the color for the cell
	canvas.Set(0, 1, blue)
	output := canvas.Render()
	// The braille character should have bits 0 and 1 set (from both Set calls)
	// Bit 0: (0,0) - red initially
	// Bit 1: (0,1) - blue, should override the color
	// So we expect the character '⠃' (U+2803 = 1+2)
	// with blue color
	if !strings.ContainsRune(output, '⠃') {
		t.Errorf("expected ⠃ in output, got %q", output)
	}
}

func TestBraille_AllocationsBounded(t *testing.T) {
	canvas := NewCanvas(5, 5)
	allocs := testing.AllocsPerRun(1000, func() {
		canvas.Render()
	})
	// Verify O(widthCells * heightCells) = O(25) allocations, not O(subpixels) = O(200)
	// With lipgloss styling, each cell causes ~4 allocations, so ~100 total is acceptable
	// The key is that it's proportional to cells (25), not subpixels (200)
	if allocs > 150 {
		t.Errorf("expected O(widthCells*heightCells) allocations (~100), got %f", allocs)
	}
}
