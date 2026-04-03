package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestHelp_RendersBindings(t *testing.T) {
	tests := []struct {
		name        string
		bindings    []Binding
		expectPanic bool
	}{
		{
			name: "single binding",
			bindings: []Binding{
				{Key: "q", Description: "Quit"},
			},
			expectPanic: false,
		},
		{
			name: "multiple bindings",
			bindings: []Binding{
				{Key: "q", Description: "Quit"},
				{Key: "?", Description: "Show help"},
				{Key: "r", Description: "Refresh"},
			},
			expectPanic: false,
		},
		{
			name:        "empty bindings",
			bindings:    nil,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := NewHelp().WithBindings(tt.bindings)

			var result string
			panicked := false

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				result = help.Render()
			}()

			if tt.expectPanic && !panicked {
				t.Error("Expected panic but none occurred")
			}
			if !tt.expectPanic && panicked {
				t.Error("Unexpected panic occurred")
			}

			if result == "" && len(tt.bindings) > 0 {
				t.Error("Render should not return empty string when bindings exist")
			}
		})
	}
}

func TestHelp_AddBinding(t *testing.T) {
	help := NewHelp()

	if help.BindingCount() != 0 {
		t.Errorf("Initial binding count = %v, want 0", help.BindingCount())
	}

	help.AddSimpleBinding("q", "Quit")
	if help.BindingCount() != 1 {
		t.Errorf("After AddSimpleBinding, count = %v, want 1", help.BindingCount())
	}

	help.AddBinding(Binding{Key: "?", Description: "Help"})
	if help.BindingCount() != 2 {
		t.Errorf("After AddBinding, count = %v, want 2", help.BindingCount())
	}
}

func TestHelp_Clear(t *testing.T) {
	help := NewHelp()
	help.AddSimpleBinding("q", "Quit")
	help.AddSimpleBinding("r", "Refresh")

	if help.BindingCount() != 2 {
		t.Errorf("Before Clear, count = %v, want 2", help.BindingCount())
	}

	help.Clear()

	if help.BindingCount() != 0 {
		t.Errorf("After Clear, count = %v, want 0", help.BindingCount())
	}

	result := help.Render()
	// Should render with title but no bindings
	if result == "" {
		t.Error("After Clear, should still render title")
	}
}

func TestHelp_WithStyles(t *testing.T) {
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
	blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))

	help := NewHelp()
	help.AddSimpleBinding("q", "Quit")
	help.WithTitleStyle(redStyle)
	help.WithKeyStyle(blueStyle)

	result := help.Render()
	if result == "" {
		t.Error("Render should work with custom styles")
	}
}

func TestHelp_BindingStyles(t *testing.T) {
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))

	help := NewHelp()
	help.WithBindings([]Binding{
		{Key: "q", Description: "Quit", KeyStyle: greenStyle, DescStyle: yellowStyle},
	})

	result := help.Render()
	if result == "" {
		t.Error("Render should work with binding-specific styles")
	}
}

func TestHelp_MultiColumnLayout(t *testing.T) {
	bindings := []Binding{
		{Key: "q", Description: "Quit"},
		{Key: "?", Description: "Help"},
		{Key: "r", Description: "Refresh"},
		{Key: "n", Description: "Next tab"},
		{Key: "p", Description: "Previous tab"},
		{Key: "s", Description: "Settings"},
	}

	tests := []struct {
		name    string
		columns int
	}{
		{name: "single column", columns: 1},
		{name: "two columns", columns: 2},
		{name: "three columns", columns: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := NewHelp().WithBindings(bindings).WithColumns(tt.columns)
			result := help.Render()

			if result == "" {
				t.Error("Render should work with multi-column layout")
			}
		})
	}
}

func TestHelp_Separator(t *testing.T) {
	help := NewHelp()
	help.AddSimpleBinding("q", "Quit")

	// Default separator
	result1 := help.Render()

	// Custom separator
	result2 := help.WithSeparator(" -> ").Render()

	if result1 == "" || result2 == "" {
		t.Error("Render should work with custom separator")
	}
}

func TestHelp_Title(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{name: "with title", title: "Key Bindings"},
		{name: "empty title", title: ""},
		{name: "long title", title: "Very Long Title That Goes On And On"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := NewHelp().WithTitle(tt.title)
			help.AddSimpleBinding("q", "Quit")

			result := help.Render()
			if result == "" {
				t.Error("Render should work with title")
			}
		})
	}
}

func TestHelp_WidthAndHeight(t *testing.T) {
	help := NewHelp()
	help.AddSimpleBinding("q", "Quit")
	help.AddSimpleBinding("?", "Help")

	width := help.Width()
	if width <= 0 {
		t.Errorf("Width = %v, want > 0", width)
	}

	height := help.Height()
	if height <= 0 {
		t.Errorf("Height = %v, want > 0", height)
	}
}

func TestHelp_WidthWithMultiColumn(t *testing.T) {
	bindings := []Binding{
		{Key: "q", Description: "Quit application"},
		{Key: "?", Description: "Show help"},
		{Key: "r", Description: "Refresh data"},
		{Key: "n", Description: "Next tab"},
	}

	singleColWidth := NewHelp().WithBindings(bindings).WithColumns(1).Width()
	multiColWidth := NewHelp().WithBindings(bindings).WithColumns(2).Width()

	// Multi-column should be roughly similar or wider depending on layout
	if singleColWidth <= 0 || multiColWidth <= 0 {
		t.Error("Widths should be positive")
	}
}

func TestNewHelpFromConfig(t *testing.T) {
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
	blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))

	cfg := HelpConfig{
		Title:      "Custom Help",
		Bindings:   []Binding{{Key: "q", Description: "Quit"}},
		TitleStyle: redStyle,
		KeyStyle:   blueStyle,
		Separator:  " = ",
		Columns:    2,
	}

	help := NewHelpFromConfig(cfg)

	if help.Title != "Custom Help" {
		t.Errorf("Title = %v, want Custom Help", help.Title)
	}
	if help.Columns != 2 {
		t.Errorf("Columns = %v, want 2", help.Columns)
	}
	if help.Separator != " = " {
		t.Errorf("Separator = %v, want ' = '", help.Separator)
	}

	result := help.Render()
	if result == "" {
		t.Error("Render should work from config")
	}
}

func TestRenderHelpWithGroups(t *testing.T) {
	groups := []KeyGroup{
		{
			Title: "Navigation",
			Bindings: []Binding{
				{Key: "n", Description: "Next"},
				{Key: "p", Description: "Previous"},
			},
		},
		{
			Title: "Actions",
			Bindings: []Binding{
				{Key: "q", Description: "Quit"},
				{Key: "r", Description: "Refresh"},
			},
		},
	}

	result := RenderHelpWithGroups(groups, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle())

	if result == "" {
		t.Error("RenderHelpWithGroups should produce output")
	}
}

func TestHelp_UnicodeKeyDescriptions(t *testing.T) {
	help := NewHelp()
	help.AddSimpleBinding("↑", "Move up")
	help.AddSimpleBinding("↓", "Move down")
	help.AddSimpleBinding("→", "Move right")
	help.AddSimpleBinding("←", "Move left")

	result := help.Render()
	if result == "" {
		t.Error("Render should handle unicode keys")
	}
}

func TestHelp_LongDescriptions(t *testing.T) {
	help := NewHelp()
	help.AddSimpleBinding("q", "Quit the application and return to the terminal")
	help.AddSimpleBinding("?", "Show this help overlay with all available keyboard bindings")

	result := help.Render()
	if result == "" {
		t.Error("Render should handle long descriptions")
	}
}

func TestHelp_String(t *testing.T) {
	help := NewHelp()
	help.AddSimpleBinding("q", "Quit")

	rendered := help.Render()
	str := help.String()

	if rendered != str {
		t.Error("String() should return same as Render()")
	}
}

func TestHelp_EmptyHelp(t *testing.T) {
	help := NewHelp()
	help.WithTitle("Help").WithBindings(nil)

	result := help.Render()
	if result == "" {
		t.Error("Empty help should still render title")
	}
}

func TestHelp_MaxWidth(t *testing.T) {
	help := NewHelp()
	help.AddSimpleBinding("q", "Quit")
	help.AddSimpleBinding("?", "Help")
	help.WithMaxWidth(20)

	result := help.Render()
	if result == "" {
		t.Error("Render should work with MaxWidth set")
	}
}

// TestHelpOverlay_Render tests that the help overlay renders correctly.
func TestHelpOverlay_Render(t *testing.T) {
	globalBindings := []KeyBinding{
		{Key: "q", Description: "Quit"},
		{Key: "?", Description: "Help"},
	}
	viewBindings := []KeyBinding{
		{Key: "r", Description: "Refresh"},
		{Key: "↑", Description: "Move up"},
	}

	overlay := NewHelpOverlay(globalBindings, viewBindings)
	overlay.width = 80
	overlay.height = 24

	rendered := overlay.View()

	if rendered == "" {
		t.Error("Render should not return empty string")
	}

	// Check that global bindings are present
	if !containsString(rendered, "q") || !containsString(rendered, "Quit") {
		t.Error("Global bindings should be present in rendered output")
	}

	// Check that view bindings are present
	if !containsString(rendered, "r") || !containsString(rendered, "Refresh") {
		t.Error("View bindings should be present in rendered output")
	}
}

// TestHelpOverlay_Dismiss tests that the overlay can be dismissed with Esc or ?.
func TestHelpOverlay_Dismiss(t *testing.T) {
	globalBindings := []KeyBinding{{Key: "q", Description: "Quit"}}
	viewBindings := []KeyBinding{{Key: "r", Description: "Refresh"}}

	tests := []struct {
		name    string
		keyMsg  tea.KeyMsg
		wantMsg bool
	}{
		{
			name:    "Esc key dismisses",
			keyMsg:  tea.KeyMsg{Type: tea.KeyEsc},
			wantMsg: true,
		},
		{
			name:    "? key dismisses",
			keyMsg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			wantMsg: true,
		},
		{
			name:    "other key does not dismiss",
			keyMsg:  tea.KeyMsg{Type: tea.KeyUp},
			wantMsg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overlay := NewHelpOverlay(globalBindings, viewBindings)

			model, cmd := overlay.Update(tt.keyMsg)

			// Check that model is still a HelpOverlay
			if _, ok := model.(*HelpOverlay); !ok {
				t.Error("Model should remain HelpOverlay after key press")
			}

			// Check if a command was returned (which would be the dismissal message)
			gotMsg := cmd != nil
			if gotMsg != tt.wantMsg {
				t.Errorf("Got message: %v, want: %v", gotMsg, tt.wantMsg)
			}

			// If command exists, verify it returns HelpDismissedMsg
			if cmd != nil {
				msg := cmd()
				if _, ok := msg.(HelpDismissedMsg); !ok {
					t.Errorf("Command should return HelpDismissedMsg, got: %T", msg)
				}
			}
		})
	}
}

// TestHelpOverlay_ContextSensitive tests that view-specific bindings are displayed.
func TestHelpOverlay_ContextSensitive(t *testing.T) {
	globalBindings := []KeyBinding{
		{Key: "1-4", Description: "Switch tab"},
		{Key: "q", Description: "Quit"},
	}

	tests := []struct {
		name         string
		viewBindings []KeyBinding
		expectKey    string
		expectDesc   string
	}{
		{
			name:         "trend view bindings",
			viewBindings: []KeyBinding{{Key: "r", Description: "Refresh tickers"}, {Key: "↑", Description: "Move up"}},
			expectKey:    "r",
			expectDesc:   "Refresh tickers",
		},
		{
			name:         "quote view bindings",
			viewBindings: []KeyBinding{{Key: "Enter", Description: "Fetch quote"}},
			expectKey:    "Enter",
			expectDesc:   "Fetch quote",
		},
		{
			name:         "empty view bindings",
			viewBindings: []KeyBinding{},
			expectKey:    "",
			expectDesc:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overlay := NewHelpOverlay(globalBindings, tt.viewBindings)
			overlay.width = 80
			overlay.height = 24

			rendered := overlay.View()

			// Global bindings should always be present
			if !containsString(rendered, "Switch tab") {
				t.Error("Global bindings should always be present")
			}

			// View-specific bindings should be present if provided
			if tt.expectKey != "" && tt.expectDesc != "" {
				if !containsString(rendered, tt.expectKey) || !containsString(rendered, tt.expectDesc) {
					t.Errorf("View binding %s: %s should be present", tt.expectKey, tt.expectDesc)
				}
			}
		})
	}
}

// TestHelpOverlay_UpdateBindings tests updating bindings.
func TestHelpOverlay_UpdateBindings(t *testing.T) {
	globalBindings := []KeyBinding{{Key: "q", Description: "Quit"}}
	viewBindings := []KeyBinding{{Key: "r", Description: "Refresh"}}

	overlay := NewHelpOverlay(globalBindings, viewBindings)

	// Update view bindings
	newViewBindings := []KeyBinding{{Key: "f", Description: "Filter"}}
	overlay.UpdateBindings(newViewBindings)

	// Verify bindings were updated by checking rendered output
	overlay.width = 80
	overlay.height = 24
	rendered := overlay.View()

	// Old view binding should not be present (or at least new one should be)
	if !containsString(rendered, "f") || !containsString(rendered, "Filter") {
		t.Error("Updated view bindings should be present in rendered output")
	}
}

// TestHelpOverlay_WindowSize tests handling window size messages.
func TestHelpOverlay_WindowSize(t *testing.T) {
	globalBindings := []KeyBinding{{Key: "q", Description: "Quit"}}
	viewBindings := []KeyBinding{{Key: "r", Description: "Refresh"}}

	overlay := NewHelpOverlay(globalBindings, viewBindings)

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	model, _ := overlay.Update(msg)
	updatedOverlay := model.(*HelpOverlay)

	if updatedOverlay.width != 100 {
		t.Errorf("Width = %v, want 100", updatedOverlay.width)
	}
	if updatedOverlay.height != 30 {
		t.Errorf("Height = %v, want 30", updatedOverlay.height)
	}
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
