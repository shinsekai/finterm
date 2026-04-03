// Package components provides reusable TUI components.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Spinner represents an animated loading indicator.
type Spinner struct {
	// Current frame index (0-based)
	frames  []string
	current int

	// Text to display alongside the spinner
	text string

	// Style for the spinner frames
	frameStyle lipgloss.Style

	// Style for the accompanying text
	textStyle lipgloss.Style
}

// NewSpinner creates a new spinner with default frames.
func NewSpinner() *Spinner {
	return &Spinner{
		frames:     defaultFrames,
		current:    0,
		text:       "",
		frameStyle: lipgloss.NewStyle(),
		textStyle:  lipgloss.NewStyle(),
	}
}

// defaultFrames provides the default spinner animation frames.
var defaultFrames = []string{
	"⠋",
	"⠙",
	"⠹",
	"⠸",
	"⠼",
	"⠴",
	"⠦",
	"⠧",
	"⠇",
	"⠏",
}

// dotFrames provides an alternative dot-based spinner.
var dotFrames = []string{
	"⠈⠁",
	"⠈⠃",
	"⠈⠕",
	"⠸⠁",
	"⠸⠃",
	"⠸⠕",
	"⠹⠁",
	"⠹⠃",
	"⠹⠕",
}

// barFrames provides a bar-based spinner.
var barFrames = []string{
	"|",
	"/",
	"-",
	"\\",
}

// WithText sets the text to display alongside the spinner.
func (s *Spinner) WithText(text string) *Spinner {
	s.text = text
	return s
}

// WithFrameStyle sets the style for the spinner frames.
func (s *Spinner) WithFrameStyle(style lipgloss.Style) *Spinner {
	s.frameStyle = style
	return s
}

// WithTextStyle sets the style for the accompanying text.
func (s *Spinner) WithTextStyle(style lipgloss.Style) *Spinner {
	s.textStyle = style
	return s
}

// WithFrames sets custom animation frames.
func (s *Spinner) WithFrames(frames []string) *Spinner {
	if len(frames) > 0 {
		s.frames = frames
		s.current = 0
	}
	return s
}

// Tick advances the spinner to the next frame.
func (s *Spinner) Tick() {
	s.current = (s.current + 1) % len(s.frames)
}

// Reset resets the spinner to the first frame.
func (s *Spinner) Reset() {
	s.current = 0
}

// CurrentFrame returns the current frame character.
func (s *Spinner) CurrentFrame() string {
	if len(s.frames) == 0 {
		return ""
	}
	return s.frames[s.current]
}

// Render returns the rendered spinner with optional text.
func (s *Spinner) Render() string {
	frame := s.frameStyle.Render(s.CurrentFrame())

	if s.text == "" {
		return frame
	}

	text := s.textStyle.Render(s.text)
	return frame + " " + text
}

// FrameCount returns the number of animation frames.
func (s *Spinner) FrameCount() int {
	return len(s.frames)
}

// String returns the current string representation of the spinner.
func (s *Spinner) String() string {
	return s.Render()
}

// Width returns the visual width of the rendered spinner.
func (s *Spinner) Width() int {
	frameWidth := lipgloss.Width(s.CurrentFrame())
	if s.text == "" {
		return frameWidth
	}
	return frameWidth + 1 + lipgloss.Width(s.text)
}

// SpinnerConfig provides configuration for creating a spinner.
type SpinnerConfig struct {
	Text       string
	Frames     []string
	FrameStyle lipgloss.Style
	TextStyle  lipgloss.Style
}

// NewSpinnerFromConfig creates a spinner from a configuration.
func NewSpinnerFromConfig(cfg SpinnerConfig) *Spinner {
	s := NewSpinner()
	s.WithText(cfg.Text)
	s.WithFrameStyle(cfg.FrameStyle)
	s.WithTextStyle(cfg.TextStyle)
	if len(cfg.Frames) > 0 {
		s.WithFrames(cfg.Frames)
	}
	return s
}

// CreateSpinnerType is a helper function to create different spinner types.
type CreateSpinnerType int

const (
	// SpinnerDefault uses the default moon/arc frames.
	SpinnerDefault CreateSpinnerType = iota
	// SpinnerDot uses dot-based frames.
	SpinnerDot
	// SpinnerBar uses bar-based frames.
	SpinnerBar
)

// NewSpinnerOfType creates a new spinner of the specified type.
func NewSpinnerOfType(spinType CreateSpinnerType) *Spinner {
	s := NewSpinner()

	switch spinType {
	case SpinnerDot:
		s.WithFrames(dotFrames)
	case SpinnerBar:
		s.WithFrames(barFrames)
	default:
		// Already uses defaultFrames
	}

	return s
}

// TruncateText truncates text to fit within a maximum width.
// Returns the truncated text with an ellipsis if needed.
func TruncateText(text string, maxWidth int) string {
	width := lipgloss.Width(text)
	if width <= maxWidth {
		return text
	}

	// Use ellipsis and trim to fit
	ellipsis := "…"
	availableWidth := maxWidth - lipgloss.Width(ellipsis)
	if availableWidth < 1 {
		return ellipsis[:maxWidth]
	}

	// Find the rune boundary that fits
	var result strings.Builder
	currentWidth := 0
	for _, r := range text {
		rWidth := lipgloss.Width(string(r))
		if currentWidth+rWidth > availableWidth {
			break
		}
		result.WriteRune(r)
		currentWidth += rWidth
	}

	return result.String() + ellipsis
}
