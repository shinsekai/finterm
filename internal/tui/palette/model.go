// Package palette provides a command palette for fuzzy-matching commands and tickers.
package palette

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// Command represents a palette command.
type Command struct {
	// ID is the unique identifier (used for matching).
	ID string
	// Label is the human-readable label.
	Label string
	// Description describes what the command does.
	Description string
	// Shortcut is the keyboard shortcut (for display).
	Shortcut string
	// Action is the command to execute when selected.
	Action func() tea.Cmd
}

// Model represents the palette model.
type Model struct {
	// commands are all available commands.
	commands []Command
	// input is the user's search input.
	input string
	// results are the filtered and sorted commands.
	results []Command
	// cursor is the index of the currently selected result.
	cursor int
	// visible indicates if the palette is currently displayed.
	visible bool
	// width and height are the terminal dimensions.
	width, height int
	// boxStyle is the style for the overlay box.
	boxStyle lipgloss.Style
	// titleStyle is the style for the title.
	titleStyle lipgloss.Style
	// cursorStyle is the style for the selected item.
	cursorStyle lipgloss.Style
	// labelStyle is the style for command labels.
	labelStyle lipgloss.Style
	// descStyle is the style for command descriptions.
	descStyle lipgloss.Style
	// inputStyle is the style for the input field.
	inputStyle lipgloss.Style
	// dimStyle is the style for non-matching characters.
	dimStyle lipgloss.Style
}

// New creates a new palette model with the given commands.
func New(commands []Command) *Model {
	// Sort commands by ID for consistent ordering
	sorted := make([]Command, len(commands))
	copy(sorted, commands)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	return &Model{
		commands: sorted,
		results:  sorted,
		input:    "",
		cursor:   0,
		visible:  false,
		width:    80,
		height:   24,
		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2),
		titleStyle: lipgloss.NewStyle().
			Bold(true),
		cursorStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#44475A")),
		labelStyle: lipgloss.NewStyle().
			Bold(true),
		descStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")),
		inputStyle: lipgloss.NewStyle(),
		dimStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")),
	}
}

// Init initializes the palette model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

// handleKeyMsg handles keyboard input messages.
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return m, func() tea.Msg { return CloseMsg{} }

	case tea.KeyEnter:
		if len(m.results) > 0 && m.cursor < len(m.results) {
			cmd := m.results[m.cursor].Action
			return m, func() tea.Msg {
				cmd := cmd()
				return ExecuteMsg{Cmd: cmd}
			}
		}
		return m, nil

	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case tea.KeyDown:
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
		return m, nil

	case tea.KeyCtrlN:
		// Move cursor down (Ctrl+N = next)
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
		return m, nil

	case tea.KeyCtrlP:
		// Move cursor up (Ctrl+P = previous) inside palette
		// Note: This is handled at app level to close the palette,
		// but if it reaches here, treat as cursor up
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case tea.KeyCtrlK:
		// Same as Ctrl+P for consistency
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case tea.KeyRunes:
		// Handle character input
		m.input += string(msg.Runes)
		m.updateResults()
		m.cursor = 0
		return m, nil

	case tea.KeyBackspace:
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
			m.updateResults()
			m.cursor = 0
		}
		return m, nil
	}

	return m, nil
}

// updateResults filters and sorts commands based on input.
func (m *Model) updateResults() {
	if m.input == "" {
		// Show all commands sorted by ID
		m.results = make([]Command, len(m.commands))
		copy(m.results, m.commands)
		return
	}

	// Check if input contains spaces for space-separated matching
	if strings.Contains(m.input, " ") {
		m.updateResultsSpaceSeparated()
		return
	}

	// Fuzzy match against command IDs
	matches := fuzzy.Find(m.input, commandIDs(m.commands))

	// Sort by score (best match first)
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			// Tie-break by ID
			return matches[i].Str < matches[j].Str
		}
		return matches[i].Score > matches[j].Score
	})

	// Build results from matches
	m.results = make([]Command, len(matches))
	for i, match := range matches {
		m.results[i] = m.commands[match.Index]
	}
}

// updateResultsSpaceSeparated handles space-separated queries.
// Each part must fuzzy match the command ID.
func (m *Model) updateResultsSpaceSeparated() {
	parts := strings.Fields(m.input)
	if len(parts) == 0 {
		m.results = make([]Command, len(m.commands))
		copy(m.results, m.commands)
		return
	}

	var matched []Command

	for _, cmd := range m.commands {
		// Check if all parts match the command ID
		allMatch := true
		for _, part := range parts {
			if part == "" {
				continue
			}
			// Use fuzzy matching for each part
			if len(fuzzy.Find(part, []string{cmd.ID})) == 0 {
				allMatch = false
				break
			}
		}
		if allMatch {
			matched = append(matched, cmd)
		}
	}

	m.results = matched
}

// commandIDs extracts IDs from commands for fuzzy matching.
func commandIDs(commands []Command) []string {
	ids := make([]string, len(commands))
	for i, cmd := range commands {
		ids[i] = cmd.ID
	}
	return ids
}

// View renders the palette.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	// Calculate dimensions (60% width, 40% height, min 50x10)
	paletteWidth := maxInt(50, m.width*3/5)
	paletteHeight := maxInt(10, m.height*2/5)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(m.titleStyle.Render("Command Palette"))
	content.WriteString("\n\n")

	// Input field with prompt
	content.WriteString("> ")
	content.WriteString(m.inputStyle.Render(m.input))

	// Cursor indicator
	if len(m.input) > 0 {
		content.WriteString("█")
	}

	content.WriteString("\n\n")

	// Results
	if len(m.results) == 0 {
		content.WriteString(m.dimStyle.Render("No matches found"))
	} else {
		// Calculate visible rows (leave room for title, input, and padding)
		maxVisibleRows := paletteHeight - 6
		startRow := 0
		endRow := len(m.results)

		// Scroll if cursor is beyond visible area
		if len(m.results) > maxVisibleRows {
			if m.cursor >= maxVisibleRows {
				startRow = m.cursor - maxVisibleRows + 1
				endRow = m.cursor + 1
			} else {
				endRow = maxVisibleRows
			}
		}

		for i := startRow; i < endRow; i++ {
			cmd := m.results[i]
			line := m.renderResultLine(cmd, i == m.cursor)
			content.WriteString(line)
			content.WriteString("\n")
		}

		// Scroll indicator if needed
		if len(m.results) > maxVisibleRows {
			content.WriteString("\n")
			content.WriteString(m.dimStyle.Render(
				fmt.Sprintf("Showing %d-%d of %d", startRow+1, min(endRow, len(m.results)), len(m.results)),
			))
		}
	}

	// Apply box style
	overlay := m.boxStyle.
		Width(paletteWidth - 4).
		Height(paletteHeight - 2).
		Render(content.String())

	// Center the overlay
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		overlay,
	)
}

// renderResultLine renders a single result line.
func (m Model) renderResultLine(cmd Command, isSelected bool) string {
	var line strings.Builder

	// Apply cursor style if selected
	if isSelected {
		line.WriteString(m.cursorStyle.Render(" "))
	} else {
		line.WriteString(" ")
	}

	// Render label
	label := cmd.Label
	if cmd.Shortcut != "" {
		label = fmt.Sprintf("%s (%s)", label, cmd.Shortcut)
	}
	line.WriteString(m.labelStyle.Render(label))

	// Add padding and description
	if cmd.Description != "" {
		padding := 40 - lipgloss.Width(label)
		if padding < 1 {
			padding = 1
		}
		line.WriteString(strings.Repeat(" ", padding))
		line.WriteString(m.descStyle.Render(cmd.Description))
	}

	return line.String()
}

// Show displays the palette.
func (m *Model) Show() {
	m.visible = true
	m.input = ""
	m.cursor = 0
	m.updateResults()
}

// Hide hides the palette.
func (m *Model) Hide() {
	m.visible = false
}

// IsVisible returns whether the palette is visible.
func (m Model) IsVisible() bool {
	return m.visible
}

// GetInput returns the current input.
func (m Model) GetInput() string {
	return m.input
}

// GetResults returns the current results.
func (m Model) GetResults() []Command {
	return m.results
}

// GetCursor returns the current cursor position.
func (m Model) GetCursor() int {
	return m.cursor
}

// SetCommands sets the available commands and updates results.
func (m *Model) SetCommands(commands []Command) {
	// Sort commands by ID for consistent ordering
	sorted := make([]Command, len(commands))
	copy(sorted, commands)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	m.commands = sorted
	m.updateResults()
}

// SetBoxStyle sets the box style.
func (m *Model) SetBoxStyle(style lipgloss.Style) {
	m.boxStyle = style
}

// SetTitleStyle sets the title style.
func (m *Model) SetTitleStyle(style lipgloss.Style) {
	m.titleStyle = style
}

// SetCursorStyle sets the cursor style.
func (m *Model) SetCursorStyle(style lipgloss.Style) {
	m.cursorStyle = style
}

// SetLabelStyle sets the label style.
func (m *Model) SetLabelStyle(style lipgloss.Style) {
	m.labelStyle = style
}

// SetDescStyle sets the description style.
func (m *Model) SetDescStyle(style lipgloss.Style) {
	m.descStyle = style
}

// SetInputStyle sets the input style.
func (m *Model) SetInputStyle(style lipgloss.Style) {
	m.inputStyle = style
}

// SetDimStyle sets the dim style for non-matching text.
func (m *Model) SetDimStyle(style lipgloss.Style) {
	m.dimStyle = style
}

// CloseMsg is sent when the palette is closed.
type CloseMsg struct{}

// ExecuteMsg is sent when a command is executed.
type ExecuteMsg struct {
	Cmd tea.Cmd
}

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
