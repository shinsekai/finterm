// Package palette provides a command palette for fuzzy-matching commands and tickers.
package palette

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// updateModel is a helper that handles the type assertion for Update.
func updateModel(m *Model, msg tea.Msg) *Model {
	newM, _ := m.Update(msg)
	return newM.(*Model)
}

func TestNew_SortsCommandsByID(t *testing.T) {
	commands := []Command{
		{ID: "quote.BTC", Label: "Quote: BTC"},
		{ID: "go.trend", Label: "Go to Trend"},
		{ID: "theme.minimal", Label: "Theme: Minimal"},
		{ID: "quote.AAPL", Label: "Quote: AAPL"},
	}

	model := New(commands)

	// Commands should be sorted by ID
	expected := []string{"go.trend", "quote.AAPL", "quote.BTC", "theme.minimal"}
	for i, cmd := range model.commands {
		if cmd.ID != expected[i] {
			t.Errorf("Command %d: expected ID %q, got %q", i, expected[i], cmd.ID)
		}
	}
}

func TestShow_Hide_IsVisible(t *testing.T) {
	model := New([]Command{})
	model.Show()

	if !model.IsVisible() {
		t.Error("Expected palette to be visible after Show()")
	}

	model.Hide()

	if model.IsVisible() {
		t.Error("Expected palette to be hidden after Hide()")
	}
}

func TestShow_ResetsState(t *testing.T) {
	model := New([]Command{
		{ID: "a", Action: func() tea.Cmd { return nil }},
		{ID: "b", Action: func() tea.Cmd { return nil }},
		{ID: "c", Action: func() tea.Cmd { return nil }},
	})

	// Set some state
	model.input = "test"
	model.cursor = 2

	// Show should reset state
	model.Show()

	if model.input != "" {
		t.Errorf("Expected input to be empty after Show(), got %q", model.input)
	}
	if model.cursor != 0 {
		t.Errorf("Expected cursor to be 0 after Show(), got %d", model.cursor)
	}
}

func TestUpdate_Esc_ClosesPalette(t *testing.T) {
	model := New([]Command{})
	model.Show()

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Esc returns a CloseMsg command
	if cmd == nil {
		t.Error("Expected command to be returned for Esc key")
	}

	// The Hide() method must be called explicitly (app handles this)
	model.Hide()

	if model.IsVisible() {
		t.Error("Expected palette to be hidden after Hide(), but it's still visible")
	}
}

func TestUpdate_Enter_ExecutesAction(t *testing.T) {
	actionCalled := false
	commands := []Command{
		{ID: "test", Action: func() tea.Cmd {
			actionCalled = true
			return nil
		}},
	}

	model := New(commands)
	model.Show()

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("Expected command to be returned for Enter key")
	}

	// Execute the command to verify action is called
	if cmd != nil {
		cmd()
	}

	if !actionCalled {
		t.Error("Expected action to be called when command is executed")
	}
}

func TestUpdate_Enter_NoAction(t *testing.T) {
	model := New([]Command{})
	model.Show()

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("Expected no command when no results available")
	}
}

func TestUpdate_Up_Down_MoveCursor(t *testing.T) {
	commands := []Command{
		{ID: "a", Action: func() tea.Cmd { return nil }},
		{ID: "b", Action: func() tea.Cmd { return nil }},
		{ID: "c", Action: func() tea.Cmd { return nil }},
	}

	model := New(commands)
	model.Show()

	// Move down
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyDown})
	if model.cursor != 1 {
		t.Errorf("Expected cursor to be 1 after Down, got %d", model.cursor)
	}

	// Move down again
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyDown})
	if model.cursor != 2 {
		t.Errorf("Expected cursor to be 2 after Down, got %d", model.cursor)
	}

	// Try to move down past end
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyDown})
	if model.cursor != 2 {
		t.Errorf("Expected cursor to stay at 2, got %d", model.cursor)
	}

	// Move up
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyUp})
	if model.cursor != 1 {
		t.Errorf("Expected cursor to be 1 after Up, got %d", model.cursor)
	}

	// Try to move up past start
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyUp})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyUp})
	if model.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", model.cursor)
	}
}

func TestUpdate_CtrlNCtrlP_MoveCursor(t *testing.T) {
	commands := []Command{
		{ID: "a", Action: func() tea.Cmd { return nil }},
		{ID: "b", Action: func() tea.Cmd { return nil }},
		{ID: "c", Action: func() tea.Cmd { return nil }},
	}

	model := New(commands)
	model.Show()

	// Ctrl+P should move up
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyCtrlP})
	if model.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0 after Ctrl+P (at start), got %d", model.cursor)
	}

	// Move down first
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyDown})

	// Now Ctrl+P should move up
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyCtrlP})
	if model.cursor != 0 {
		t.Errorf("Expected cursor to be 0 after Ctrl+P, got %d", model.cursor)
	}

	// Ctrl+N should move down
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyCtrlN})
	if model.cursor != 1 {
		t.Errorf("Expected cursor to be 1 after Ctrl+N, got %d", model.cursor)
	}

	// Ctrl+K should also move up (alias)
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyCtrlK})
	if model.cursor != 0 {
		t.Errorf("Expected cursor to be 0 after Ctrl+K, got %d", model.cursor)
	}
}

func TestUpdate_Characters_FilterResults(t *testing.T) {
	commands := []Command{
		{ID: "quote.AAPL", Label: "Quote: AAPL"},
		{ID: "quote.AMZN", Label: "Quote: AMZN"},
		{ID: "quote.BTC", Label: "Quote: BTC"},
		{ID: "go.trend", Label: "Go to Trend"},
	}

	model := New(commands)
	model.Show()

	// Type "aap"
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})

	if model.input != "aap" {
		t.Errorf("Expected input to be 'aap', got %q", model.input)
	}

	if len(model.results) != 1 {
		t.Errorf("Expected 1 result for 'aap', got %d", len(model.results))
	}

	if len(model.results) > 0 && model.results[0].ID != "quote.AAPL" {
		t.Errorf("Expected first result to be quote.AAPL, got %q", model.results[0].ID)
	}
}

func TestUpdate_SpaceSeparated_Matches(t *testing.T) {
	commands := []Command{
		{ID: "theme.minimal", Label: "Theme: Minimal"},
		{ID: "theme.default", Label: "Theme: Default"},
		{ID: "theme.colorblind", Label: "Theme: Colorblind"},
		{ID: "go.trend", Label: "Go to Trend"},
	}

	model := New(commands)
	model.Show()

	// Type "thm min" (should match theme.minimal)
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// "thm min" should fuzzy match "theme.minimal" (th -> theme, m -> minimal)
	if len(model.results) == 0 {
		t.Error("Expected at least one result for 'thm min'")
	}
}

func TestUpdate_Backspace(t *testing.T) {
	commands := []Command{
		{ID: "quote.AAPL", Label: "Quote: AAPL"},
		{ID: "quote.BTC", Label: "Quote: BTC"},
	}

	model := New(commands)
	model.Show()

	// Type "aapl"
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})

	// Backspace
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyBackspace})

	if model.input != "aap" {
		t.Errorf("Expected input to be 'aap' after backspace, got %q", model.input)
	}

	// Cursor should reset
	if model.cursor != 0 {
		t.Errorf("Expected cursor to be 0 after backspace, got %d", model.cursor)
	}
}

func TestUpdate_EmptyInputShowsAllSorted(t *testing.T) {
	commands := []Command{
		{ID: "quote.BTC", Label: "Quote: BTC"},
		{ID: "go.trend", Label: "Go to Trend"},
		{ID: "theme.minimal", Label: "Theme: Minimal"},
		{ID: "quote.AAPL", Label: "Quote: AAPL"},
	}

	model := New(commands)
	model.Show()

	// Empty input should show all commands sorted by ID
	if len(model.results) != 4 {
		t.Errorf("Expected 4 results for empty input, got %d", len(model.results))
	}

	expectedOrder := []string{"go.trend", "quote.AAPL", "quote.BTC", "theme.minimal"}
	for i, result := range model.results {
		if result.ID != expectedOrder[i] {
			t.Errorf("Result %d: expected ID %q, got %q", i, expectedOrder[i], result.ID)
		}
	}
}

func TestUpdate_NoMatches(t *testing.T) {
	commands := []Command{
		{ID: "quote.AAPL", Label: "Quote: AAPL"},
		{ID: "quote.BTC", Label: "Quote: BTC"},
	}

	model := New(commands)
	model.Show()

	// Type something that won't match
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})

	if len(model.results) != 0 {
		t.Errorf("Expected 0 results for 'xyz', got %d", len(model.results))
	}
}

func TestUpdate_WindowSize(t *testing.T) {
	model := New([]Command{})
	model.width = 80
	model.height = 24

	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	model = updateModel(model, msg)

	if model.width != 100 {
		t.Errorf("Expected width to be 100, got %d", model.width)
	}
	if model.height != 30 {
		t.Errorf("Expected height to be 30, got %d", model.height)
	}
}

func TestView_NotVisibleReturnsEmpty(t *testing.T) {
	model := New([]Command{})
	view := model.View()

	if view != "" {
		t.Error("Expected empty view when palette is not visible")
	}
}

func TestView_VisibleReturnsContent(t *testing.T) {
	model := New([]Command{
		{ID: "test", Label: "Test Command", Description: "Test description"},
	})
	model.Show()
	model.width = 80
	model.height = 24

	view := model.View()

	if view == "" {
		t.Error("Expected non-empty view when palette is visible")
	}

	// Should contain the title
	if !containsString(view, "Command Palette") {
		t.Error("Expected view to contain 'Command Palette'")
	}

	// Should contain the command label
	if !containsString(view, "Test Command") {
		t.Error("Expected view to contain 'Test Command'")
	}
}

func TestView_RendersAtMinimumTerminalSize(t *testing.T) {
	model := New([]Command{
		{ID: "test", Label: "Test Command"},
	})
	model.Show()
	model.width = 50
	model.height = 10

	view := model.View()

	if view == "" {
		t.Error("Expected view to render at minimum terminal size (50x10)")
	}
}

func TestGetters(t *testing.T) {
	commands := []Command{
		{ID: "a", Label: "A"},
		{ID: "b", Label: "B"},
		{ID: "c", Label: "C"},
	}

	model := New(commands)
	model.Show()
	model.input = "test"
	model.cursor = 1

	if model.GetInput() != "test" {
		t.Errorf("Expected GetInput() to return 'test', got %q", model.GetInput())
	}

	if len(model.GetResults()) != 3 {
		t.Errorf("Expected GetResults() to return 3 commands, got %d", len(model.GetResults()))
	}

	if model.GetCursor() != 1 {
		t.Errorf("Expected GetCursor() to return 1, got %d", model.GetCursor())
	}
}

func TestSetCommands_UpdatesResults(t *testing.T) {
	model := New([]Command{
		{ID: "old", Label: "Old Command"},
	})

	if len(model.GetResults()) != 1 {
		t.Errorf("Expected 1 result initially, got %d", len(model.GetResults()))
	}

	newCommands := []Command{
		{ID: "new1", Label: "New 1"},
		{ID: "new2", Label: "New 2"},
	}
	model.SetCommands(newCommands)

	if len(model.GetResults()) != 2 {
		t.Errorf("Expected 2 result after SetCommands, got %d", len(model.GetResults()))
	}
}

func TestStyleSetters(_ *testing.T) {
	model := New([]Command{})

	// Just ensure these don't panic
	model.SetBoxStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("red")))
	model.SetTitleStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("blue")))
	model.SetCursorStyle(lipgloss.NewStyle().Background(lipgloss.Color("yellow")))
	model.SetLabelStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("green")))
	model.SetDescStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("gray")))
	model.SetInputStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("white")))
	model.SetDimStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("dim")))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
