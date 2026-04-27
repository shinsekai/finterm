package components

import (
	"github.com/charmbracelet/lipgloss"
)

const brailleBase = 0x2800

// Canvas provides 2×4 subpixel resolution per character cell using Unicode braille.
// Each cell can display 8 dots (2 columns × 4 rows) at positions (0,0) to (1,3).
type Canvas struct {
	widthCells  int
	heightCells int
	pixels      [][]bool         // 2D array of pixels, indexed by (x, y)
	colors      []lipgloss.Color // cell-level colors, indexed by cell index
}

// NewCanvas creates a new braille canvas with the specified dimensions in character cells.
// The actual pixel grid is (widthCells*2) pixels wide and (heightCells*4) pixels tall.
func NewCanvas(widthCells, heightCells int) *Canvas {
	pixelWidth := widthCells * 2
	pixelHeight := heightCells * 4
	pixels := make([][]bool, pixelWidth)
	for i := range pixels {
		pixels[i] = make([]bool, pixelHeight)
	}
	colors := make([]lipgloss.Color, widthCells*heightCells)
	return &Canvas{
		widthCells:  widthCells,
		heightCells: heightCells,
		pixels:      pixels,
		colors:      colors,
	}
}

// Set lights the pixel at the given coordinates with the specified color.
// Coordinates are in pixel space, with (0,0) at the top-left.
// Out-of-range coordinates are silently ignored.
func (c *Canvas) Set(px, py int, color lipgloss.Color) {
	if px < 0 || py < 0 || px >= c.widthCells*2 || py >= c.heightCells*4 {
		return
	}
	c.pixels[px][py] = true
	cellX, cellY := px/2, py/4
	c.colors[cellY*c.widthCells+cellX] = color
}

// Line draws a line from (x0, y0) to (x1, y1) using Bresenham's algorithm.
// All pixels along the line are set to the specified color.
func (c *Canvas) Line(x0, y0, x1, y1 int, color lipgloss.Color) {
	if x0 == x1 {
		for y := y0; y <= y1; y++ {
			c.Set(x0, y, color)
		}
		for y := y1; y <= y0; y++ {
			c.Set(x0, y, color)
		}
		return
	}
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		c.Set(x0, y0, color)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// Render converts the canvas to a string of braille characters with ANSI color escapes.
// Each character cell renders as a single braille character with the last-set color.
// Rows are terminated by newlines except the last row.
func (c *Canvas) Render() string {
	// Pre-allocate buffer: up to 16 bytes per cell (ANSI codes + reset + braille + newline)
	// Fixed allocation pattern: exactly (widthCells * heightCells) cells processed
	const maxBytesPerCell = 16
	result := make([]byte, 0, c.widthCells*c.heightCells*maxBytesPerCell)
	for cy := 0; cy < c.heightCells; cy++ {
		for cx := 0; cx < c.widthCells; cx++ {
			var code uint8
			if c.pixels[cx*2][cy*4] {
				code |= 1 << 0
			}
			if c.pixels[cx*2][cy*4+1] {
				code |= 1 << 1
			}
			if c.pixels[cx*2][cy*4+2] {
				code |= 1 << 2
			}
			if c.pixels[cx*2+1][cy*4] {
				code |= 1 << 3
			}
			if c.pixels[cx*2+1][cy*4+1] {
				code |= 1 << 4
			}
			if c.pixels[cx*2+1][cy*4+2] {
				code |= 1 << 5
			}
			if c.pixels[cx*2][cy*4+3] {
				code |= 1 << 6
			}
			if c.pixels[cx*2+1][cy*4+3] {
				code |= 1 << 7
			}
			color := c.colors[cy*c.widthCells+cx]
			// Use fixed foreground color to minimize allocations
			styled := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(string(rune(brailleBase) + rune(code)))
			result = append(result, styled...)
		}
		if cy < c.heightCells-1 {
			result = append(result, '\n')
		}
	}
	return string(result)
}

// countSetBits returns the number of set bits in a byte.
func countSetBits(b byte) int {
	count := 0
	for b != 0 {
		count += int(b & 1)
		b >>= 1
	}
	return count
}
