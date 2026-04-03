// Package trend provides the trend following TUI model.
package trend

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the trend following view model.
type Model struct {
	loading bool
	data    string // Placeholder for trend data
	err     error
}

// NewModel creates a new trend model.
func NewModel() Model {
	return Model{
		loading: false,
		data:    "",
		err:     nil,
	}
}

// Init initializes the trend model and returns an initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case RefreshMsg:
		// Handle refresh request
		m.loading = true
		return m, m.loadDataCmd()
	case DataLoadedMsg:
		// Handle data loaded
		m.loading = false
		m.data = msg.Data
		return m, nil
	case ErrorMsg:
		// Handle error
		m.loading = false
		m.err = msg.Err
		return m, nil
	default:
		// Delegate to default handling
		return m, nil
	}
}

// View renders the trend view.
func (m Model) View() string {
	if m.loading {
		return "Loading trend data...\n"
	}
	if m.err != nil {
		return "Error loading trend data\n"
	}
	if m.data == "" {
		return "Trend view - Press 'r' to refresh\n"
	}
	return m.data
}

// loadDataCmd returns a command to load trend data.
func (m Model) loadDataCmd() tea.Cmd {
	return func() tea.Msg {
		// Placeholder: in real implementation, fetch data from domain layer
		return DataLoadedMsg{Data: "Trend data loaded"}
	}
}

// RefreshMsg is a message to refresh the trend data.
type RefreshMsg struct{}

// DataLoadedMsg is a message when trend data is loaded.
type DataLoadedMsg struct {
	Data string
}

// ErrorMsg is a message when an error occurs.
type ErrorMsg struct {
	Err error
}
