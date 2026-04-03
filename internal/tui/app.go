// Package tui provides the terminal user interface for finterm.
package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/owner/finterm/internal/tui/components"
	"github.com/owner/finterm/internal/tui/macro"
	"github.com/owner/finterm/internal/tui/news"
	"github.com/owner/finterm/internal/tui/quote"
	"github.com/owner/finterm/internal/tui/trend"
)

const (
	tabTrend = iota
	tabQuote
	tabMacro
	tabNews
	numTabs
)

// tab represents a single tab in the application.
type tab struct {
	name  string
	model tea.Model
}

// Model represents the root application model that manages tab navigation.
type Model struct {
	theme         *Theme
	tabs          []tab
	activeTab     int
	showHelp      bool
	lastUpdate    time.Time
	errorCount    int
	quit          bool
	width, height int
}

// NewApp creates a new application model with all child models initialized.
func NewApp(theme *Theme) Model {
	return Model{
		theme: theme,
		tabs: []tab{
			{name: "Trend", model: trend.NewModel()},
			{name: "Quote", model: quote.NewModel()},
			{name: "Macro", model: macro.NewModel()},
			{name: "News", model: news.NewModel()},
		},
		activeTab:  tabTrend,
		showHelp:   false,
		lastUpdate: time.Now(),
		errorCount: 0,
		quit:       false,
	}
}

// Init initializes the application and returns an initial command.
// Triggers data load for the default (first) tab.
func (m Model) Init() tea.Cmd {
	// Trigger data load for the default tab
	return m.refreshActiveTab()
}

// Update handles messages and updates the application state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case DataUpdateMsg:
		m.lastUpdate = time.Now()
		return m, nil
	case ErrorUpdateMsg:
		m.errorCount++
		return m, nil
	}

	// Delegate all other messages to the active child model
	return m.delegateToChild(msg)
}

// View renders the entire application view.
func (m Model) View() string {
	if m.quit {
		return "Goodbye!"
	}

	var content string

	// Render tab bar
	content += m.renderTabBar()
	content += "\n"

	// Render active child view
	if m.showHelp {
		content += m.renderHelp()
	} else {
		content += m.tabs[m.activeTab].model.View()
	}

	// Render status bar
	content += "\n" + m.renderStatusBar()

	return content
}

// handleKeyMsg handles keyboard input messages.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.quit = true
		return m, tea.Quit

	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "q":
			m.quit = true
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "r":
			return m, m.refreshActiveTab()

		case "1":
			m.setActiveTab(tabTrend)
			return m, nil

		case "2":
			m.setActiveTab(tabQuote)
			return m, nil

		case "3":
			m.setActiveTab(tabMacro)
			return m, nil

		case "4":
			m.setActiveTab(tabNews)
			return m, nil
		}

	case tea.KeyTab:
		m.cycleTab()
		return m, nil
	}

	// Delegate to active child model
	return m.delegateToChild(msg)
}

// renderTabBar renders the tab navigation bar.
func (m Model) renderTabBar() string {
	var tabs []string
	for i, tab := range m.tabs {
		style := m.theme.Tab()
		if i == m.activeTab {
			style = m.theme.TabActive()
		}
		tabs = append(tabs, style.Render(fmt.Sprintf("%d. %s", i+1, tab.name)))
	}
	return m.theme.TabBar().Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
}

// renderHelp renders the help overlay.
func (m Model) renderHelp() string {
	help := components.NewHelp().
		WithTitle("Help").
		WithBindings([]components.Binding{
			{Key: "1-4", Description: "Switch tab"},
			{Key: "Tab", Description: "Cycle tabs"},
			{Key: "r", Description: "Refresh"},
			{Key: "?", Description: "Toggle help"},
			{Key: "q / Ctrl+C", Description: "Quit"},
		}).
		WithColumns(2)

	return m.theme.Box().Render(help.Render())
}

// renderStatusBar renders the status bar with last update time, data source, and error count.
func (m Model) renderStatusBar() string {
	left := fmt.Sprintf("Last update: %s", m.formatLastUpdate())
	right := fmt.Sprintf("Errors: %d", m.errorCount)

	statusBar := lipgloss.NewStyle().
		Foreground(m.theme.Foreground()).
		Background(m.theme.Background()).
		Width(m.width).
		Padding(0, 1)

	return statusBar.Render(lipgloss.JoinHorizontal(lipgloss.Bottom,
		left,
		lipgloss.PlaceHorizontal(m.width, lipgloss.Right, right)))
}

// formatLastUpdate formats the last update time for display.
func (m Model) formatLastUpdate() string {
	if m.lastUpdate.IsZero() {
		return "Never"
	}

	elapsed := time.Since(m.lastUpdate)
	switch {
	case elapsed < time.Minute:
		return "Just now"
	case elapsed < time.Hour:
		return fmt.Sprintf("%d min ago", int(elapsed.Minutes()))
	case elapsed < 24*time.Hour:
		return fmt.Sprintf("%d hr ago", int(elapsed.Hours()))
	default:
		return m.lastUpdate.Format("2006-01-02 15:04")
	}
}

// setActiveTab sets the active tab index.
func (m *Model) setActiveTab(index int) {
	if index >= 0 && index < numTabs {
		m.activeTab = index
	}
}

// cycleTab cycles to the next tab.
func (m *Model) cycleTab() {
	m.activeTab = (m.activeTab + 1) % numTabs
}

// refreshActiveTab sends a refresh command to the active child model.
func (m Model) refreshActiveTab() tea.Cmd {
	switch m.activeTab {
	case tabTrend:
		return func() tea.Msg { return trend.RefreshMsg{} }
	case tabQuote:
		return func() tea.Msg { return quote.RefreshMsg{} }
	case tabMacro:
		return func() tea.Msg { return macro.RefreshMsg{} }
	case tabNews:
		return func() tea.Msg { return news.RefreshMsg{} }
	default:
		return nil
	}
}

// delegateToChild delegates a message to the active child model.
func (m Model) delegateToChild(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		var cmd tea.Cmd
		m.tabs[m.activeTab].model, cmd = m.tabs[m.activeTab].model.Update(msg)
		return m, cmd
	}
	return m, nil
}

// DataUpdateMsg is a message indicating data was updated.
type DataUpdateMsg struct {
	Tab int
}

// ErrorUpdateMsg is a message indicating an error occurred.
type ErrorUpdateMsg struct {
	Tab int
	Err error
}
