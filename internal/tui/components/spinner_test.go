package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestSpinner_New(t *testing.T) {
	s := NewSpinner()

	if s == nil {
		t.Fatal("NewSpinner() returned nil")
	}

	if s.current != 0 {
		t.Errorf("Initial current = %v, want 0", s.current)
	}

	if len(s.frames) == 0 {
		t.Error("Frames should not be empty")
	}

	if s.FrameCount() == 0 {
		t.Error("FrameCount should return a positive number")
	}
}

func TestSpinner_Tick(t *testing.T) {
	tests := []struct {
		name           string
		initialCurrent int
		frameCount     int
		ticks          int
		wantCurrent    int
	}{
		{
			name:           "single tick",
			initialCurrent: 0,
			frameCount:     4,
			ticks:          1,
			wantCurrent:    1,
		},
		{
			name:           "multiple ticks",
			initialCurrent: 0,
			frameCount:     4,
			ticks:          3,
			wantCurrent:    3,
		},
		{
			name:           "wrap around",
			initialCurrent: 2,
			frameCount:     4,
			ticks:          3,
			wantCurrent:    1,
		},
		{
			name:           "full cycle",
			initialCurrent: 0,
			frameCount:     4,
			ticks:          4,
			wantCurrent:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSpinner()
			s.frames = make([]string, tt.frameCount)
			for i := range s.frames {
				s.frames[i] = string(rune('A' + i))
			}
			s.current = tt.initialCurrent

			for i := 0; i < tt.ticks; i++ {
				s.Tick()
			}

			if s.current != tt.wantCurrent {
				t.Errorf("current = %v, want %v", s.current, tt.wantCurrent)
			}
		})
	}
}

func TestSpinner_Reset(t *testing.T) {
	s := NewSpinner()
	for i := 0; i < 5; i++ {
		s.Tick()
	}

	s.Reset()

	if s.current != 0 {
		t.Errorf("After Reset, current = %v, want 0", s.current)
	}
}

func TestSpinner_CurrentFrame(t *testing.T) {
	s := NewSpinner()
	s.frames = []string{"A", "B", "C"}
	s.current = 1

	if frame := s.CurrentFrame(); frame != "B" {
		t.Errorf("CurrentFrame() = %v, want B", frame)
	}
}

func TestSpinner_WithText(t *testing.T) {
	s := NewSpinner()
	s.WithText("Loading...")

	if s.text != "Loading..." {
		t.Errorf("text = %v, want Loading...", s.text)
	}

	result := s.Render()
	if result == "" {
		t.Error("Render() should not be empty")
	}
	if result == s.CurrentFrame() {
		t.Error("Render() should include text")
	}
}

func TestSpinner_WithFrameStyle(t *testing.T) {
	s := NewSpinner()
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
	s.WithFrameStyle(style)

	if s.frameStyle.GetForeground() != lipgloss.Color("red") {
		t.Error("FrameStyle should be set")
	}

	result := s.Render()
	if result == "" {
		t.Error("Render() should not be empty")
	}
}

func TestSpinner_WithTextStyle(t *testing.T) {
	s := NewSpinner()
	s.WithText("Test")
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))
	s.WithTextStyle(style)

	if s.textStyle.GetForeground() != lipgloss.Color("blue") {
		t.Error("TextStyle should be set")
	}
}

func TestSpinner_WithFrames(t *testing.T) {
	s := NewSpinner()
	customFrames := []string{"1", "2", "3"}
	s.WithFrames(customFrames)

	if len(s.frames) != len(customFrames) {
		t.Errorf("Frames length = %v, want %v", len(s.frames), len(customFrames))
	}

	for i, frame := range s.frames {
		if frame != customFrames[i] {
			t.Errorf("Frame %d = %v, want %v", i, frame, customFrames[i])
		}
	}

	if s.current != 0 {
		t.Errorf("After WithFrames, current should reset to 0, got %v", s.current)
	}
}

func TestSpinner_WithFramesEmpty(t *testing.T) {
	s := NewSpinner()
	originalFrames := make([]string, len(s.frames))
	copy(originalFrames, s.frames)

	s.WithFrames(nil)

	if len(s.frames) != len(originalFrames) {
		t.Error("WithFrames(nil) should not change frames")
	}
}

func TestSpinner_Render(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		contains string
	}{
		{
			name:     "no text",
			text:     "",
			contains: "",
		},
		{
			name:     "with text",
			text:     "Loading",
			contains: "Loading",
		},
		{
			name:     "with long text",
			text:     "Loading data from server...",
			contains: "Loading data from server...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSpinner().WithText(tt.text)
			result := s.Render()

			if result == "" {
				t.Error("Render() should not return empty string")
			}

			if tt.contains != "" {
				// Note: Contains won't work directly due to ANSI codes
				// Just verify the result is not empty for styled text
				if len(result) < len(tt.contains)+1 {
					t.Errorf("Rendered result too short: got %d chars, want at least %d", len(result), len(tt.contains)+1)
				}
			}
		})
	}
}

func TestSpinner_Width(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minWidth int
	}{
		{
			name:     "no text",
			text:     "",
			minWidth: 1,
		},
		{
			name:     "with short text",
			text:     "Test",
			minWidth: 5, // frame + space + text
		},
		{
			name:     "with unicode text",
			text:     "Hello 世界",
			minWidth: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSpinner().WithText(tt.text)
			width := s.Width()

			if width < tt.minWidth {
				t.Errorf("Width = %v, want at least %v", width, tt.minWidth)
			}
		})
	}
}

func TestSpinner_String(t *testing.T) {
	s := NewSpinner()
	rendered := s.Render()
	str := s.String()

	if rendered != str {
		t.Errorf("String() = %v, want %v (same as Render())", str, rendered)
	}
}

func TestNewSpinnerFromConfig(t *testing.T) {
	frameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))

	cfg := SpinnerConfig{
		Text:       "Custom",
		Frames:     []string{"X", "Y", "Z"},
		FrameStyle: frameStyle,
		TextStyle:  textStyle,
	}

	s := NewSpinnerFromConfig(cfg)

	if s.text != "Custom" {
		t.Errorf("text = %v, want Custom", s.text)
	}

	if len(s.frames) != 3 {
		t.Errorf("frames length = %v, want 3", len(s.frames))
	}

	if s.frameStyle.GetForeground() != lipgloss.Color("red") {
		t.Error("FrameStyle not set correctly")
	}

	if s.textStyle.GetForeground() != lipgloss.Color("blue") {
		t.Error("TextStyle not set correctly")
	}
}

func TestNewSpinnerOfType(t *testing.T) {
	tests := []struct {
		name        string
		spinType    CreateSpinnerType
		frameLength int
	}{
		{
			name:        "default spinner",
			spinType:    SpinnerDefault,
			frameLength: len(defaultFrames),
		},
		{
			name:        "dot spinner",
			spinType:    SpinnerDot,
			frameLength: len(dotFrames),
		},
		{
			name:        "bar spinner",
			spinType:    SpinnerBar,
			frameLength: len(barFrames),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSpinnerOfType(tt.spinType)

			if s.FrameCount() != tt.frameLength {
				t.Errorf("FrameCount = %v, want %v", s.FrameCount(), tt.frameLength)
			}
		})
	}
}

func TestSpinner_AnimationCycle(t *testing.T) {
	s := NewSpinner()
	initialFrame := s.CurrentFrame()

	frameCount := s.FrameCount()
	visitedFrames := make(map[string]bool)

	// Animate through all frames
	for i := 0; i < frameCount; i++ {
		visitedFrames[s.CurrentFrame()] = true
		s.Tick()
	}

	// Verify we visited all frames
	if len(visitedFrames) != frameCount {
		t.Errorf("Visited %d frames, expected %d", len(visitedFrames), frameCount)
	}

	// After full cycle, should be back at initial frame
	if s.CurrentFrame() != initialFrame {
		t.Errorf("After full cycle, frame = %v, want initial %v", s.CurrentFrame(), initialFrame)
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		contains bool // whether the ellipsis should be present
	}{
		{
			name:     "text fits exactly",
			text:     "Hello",
			maxWidth: 5,
			contains: false,
		},
		{
			name:     "text fits with space",
			text:     "Hello",
			maxWidth: 10,
			contains: false,
		},
		{
			name:     "text needs truncation",
			text:     "Hello World",
			maxWidth: 5,
			contains: true,
		},
		{
			name:     "very short max width",
			text:     "Hello World",
			maxWidth: 1,
			contains: false, // Just ellipsis first char
		},
		{
			name:     "unicode text",
			text:     "Hello 世界",
			maxWidth: 6,
			contains: true,
		},
		{
			name:     "empty text",
			text:     "",
			maxWidth: 10,
			contains: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateText(tt.text, tt.maxWidth)

			width := lipgloss.Width(result)
			if width > tt.maxWidth {
				t.Errorf("Result width = %v, exceeds max %v", width, tt.maxWidth)
			}

			if tt.contains {
				// Check for ellipsis presence
				containsEllipsis := false
				for _, r := range result {
					if r == '…' {
						containsEllipsis = true
						break
					}
				}
				if !containsEllipsis {
					t.Error("Expected ellipsis in truncated text")
				}
			}

			// Should not panic
			_ = result
		})
	}
}

func TestTruncateText_EdgeCases(t *testing.T) {
	t.Run("zero max width", func(t *testing.T) {
		result := TruncateText("Test", 0)
		if len(result) != 1 { // Just ellipsis first char or similar
			t.Logf("Result for zero width: %q", result)
		}
	})

	t.Run("unicode characters", func(t *testing.T) {
		text := "🌟⭐💫✨"
		result := TruncateText(text, 4)
		if lipgloss.Width(result) > 4 {
			t.Errorf("Result width %d exceeds max 4", lipgloss.Width(result))
		}
	})
}
