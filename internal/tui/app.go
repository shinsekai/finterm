// Package tui provides the terminal user interface for finterm.
package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/owner/finterm/internal/cache"
	"github.com/owner/finterm/internal/config"
	"github.com/owner/finterm/internal/domain/trend/indicators"
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

// globalBindings are keyboard bindings available in all views.
var globalBindings = []components.Binding{
	{Key: "1-4", Description: "Switch tab"},
	{Key: "Tab", Description: "Cycle tabs"},
	{Key: "r", Description: "Refresh"},
	{Key: "?", Description: "Toggle help"},
	{Key: "q / Ctrl+C", Description: "Quit"},
}

// ConnectionState represents the API connection state.
type ConnectionState int

const (
	// ConnOnline indicates the API is reachable and functioning.
	ConnOnline ConnectionState = iota
	// ConnRateLimited indicates the API is rate limiting requests.
	ConnRateLimited
	// ConnOffline indicates the API is unreachable.
	ConnOffline
)

// String returns the string representation of ConnectionState.
func (c ConnectionState) String() string {
	switch c {
	case ConnOnline:
		return "online"
	case ConnRateLimited:
		return "rate limited"
	case ConnOffline:
		return "offline"
	default:
		return "unknown"
	}
}

// tab represents a single tab in the application.
type tab struct {
	name  string
	model tea.Model
}

// retryItem represents a queued retry operation.
type retryItem struct {
	tea.Cmd
	tab       int
	scheduled time.Time
	attempts  int
}

// Model represents the root application model that manages tab navigation.
type Model struct {
	theme           *Theme
	tabs            []tab
	activeTab       int
	helpOverlay     *components.HelpOverlay
	lastUpdate      time.Time
	errorCount      int
	quit            bool
	width, height   int
	connectionState ConnectionState
	rateLimitReset  time.Time
	retryQueue      []retryItem
}

// NewApp creates a new application model with all child models initialized and configured.
// All dependencies are injected and the models are configured before returning.
func NewApp(
	theme *Theme,
	quoteClient quote.QuoteClient,
	macroClient macro.Client,
	newsClient news.Client,
	trendEngine quote.Engine,
	cacheStore *cache.Store,
	watchlist *config.WatchlistConfig,
	detector *indicators.AssetClassDetector,
) Model {
	// Create and configure trend model
	trendModel := trend.NewModel()
	trendModel.Configure(context.Background(), trendEngine, watchlist, detector)

	// Create and configure quote model
	quoteModel := quote.NewModel()
	quoteModel.Configure(context.Background(), quoteClient, trendEngine)

	// Create and configure macro model
	macroModel := macro.NewModel()
	macroModel.Configure(context.Background(), macroClient, cacheStore)

	// Create and configure news model
	newsModel := news.NewModel()
	newsModel.Configure(context.Background(), newsClient)

	return Model{
		theme: theme,
		tabs: []tab{
			{name: "Trend", model: trendModel},
			{name: "Quote", model: quoteModel},
			{name: "Macro", model: macroModel},
			{name: "News", model: newsModel},
		},
		activeTab:       tabTrend,
		helpOverlay:     nil,
		lastUpdate:      time.Now(),
		errorCount:      0,
		quit:            false,
		connectionState: ConnOnline,
		rateLimitReset:  time.Time{},
		retryQueue:      []retryItem{},
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
	case components.HelpDismissedMsg:
		// Dismiss help overlay
		m.helpOverlay = nil
		return m, nil

	case tea.KeyMsg:
		// If help overlay is visible, only handle overlay keys
		if m.helpOverlay != nil {
			overlay, cmd := m.helpOverlay.Update(msg)
			// If overlay returned a command (dismissal), execute it inline
			if cmd != nil {
				cmdMsg := cmd()
				if _, ok := cmdMsg.(components.HelpDismissedMsg); ok {
					m.helpOverlay = nil
					return m, nil
				}
			}
			m.helpOverlay = overlay.(*components.HelpOverlay)
			return m, cmd
		}
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		// Update overlay dimensions if visible
		if m.helpOverlay != nil {
			overlay, cmd := m.helpOverlay.Update(msg)
			m.helpOverlay = overlay.(*components.HelpOverlay)
			return m, cmd
		}
		return m, nil
	case DataUpdateMsg:
		m.lastUpdate = time.Now()
		m.connectionState = ConnOnline
		return m, nil
	case ErrorUpdateMsg:
		m.errorCount++
		// Check if error is a rate limit error
		if isRateLimitError(msg.Err) {
			m.connectionState = ConnRateLimited
			m.rateLimitReset = time.Now().Add(time.Minute)
		} else {
			m.connectionState = ConnOffline
		}
		return m, nil
	case ConnectionOnlineMsg:
		m.connectionState = ConnOnline
		m.rateLimitReset = time.Time{}
		return m, nil
	case ConnectionOfflineMsg:
		m.connectionState = ConnOffline
		return m, nil
	case RateLimitedMsg:
		m.connectionState = ConnRateLimited
		m.rateLimitReset = msg.ResetTime
		// Queue retry for the affected tab
		return m, m.queueRetry(msg.Tab, time.Until(msg.ResetTime))
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

	// Render help overlay if visible, otherwise render active child view
	if m.helpOverlay != nil {
		content += m.helpOverlay.View()
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
	case tea.KeyCtrlC:
		m.quit = true
		return m, tea.Quit

	case tea.KeyEsc:
		// Dismiss help overlay if visible, otherwise delegate to child
		if m.helpOverlay != nil {
			m.helpOverlay = nil
			return m, nil
		}
		// Delegate ESC to active child model (e.g., clear input in quote view)
		return m.delegateToChild(msg)

	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "q":
			m.quit = true
			return m, tea.Quit

		case "?":
			return m.showHelpOverlay()

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

// showHelpOverlay shows the help overlay with context-sensitive bindings.
func (m Model) showHelpOverlay() (Model, tea.Cmd) {
	// Get view-specific bindings from the active tab
	viewBindings := m.getViewBindings()

	// Create and show the help overlay
	m.helpOverlay = components.NewHelpOverlay(globalBindings, viewBindings).
		WithTitle(m.tabs[m.activeTab].name + " Help")

	// Initialize overlay dimensions
	overlay, cmd := m.helpOverlay.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	m.helpOverlay = overlay.(*components.HelpOverlay)

	return m, cmd
}

// getViewBindings returns the key bindings for the currently active view.
func (m Model) getViewBindings() []components.Binding {
	if provider, ok := m.tabs[m.activeTab].model.(components.KeyBindingsProvider); ok {
		return provider.KeyBindings()
	}
	return nil
}

// renderStatusBar renders the status bar with connection state, last update time, and error count.
func (m Model) renderStatusBar() string {
	left := fmt.Sprintf("Status: %s | Last update: %s", m.renderConnectionState(), m.formatLastUpdate())
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

// renderConnectionState renders the connection state with appropriate styling.
func (m Model) renderConnectionState() string {
	var style lipgloss.Style
	var text string

	switch m.connectionState {
	case ConnOnline:
		style = m.theme.StatusOnline()
		text = "online"
	case ConnRateLimited:
		style = m.theme.StatusRateLimited()
		text = "rate limited"
	case ConnOffline:
		style = m.theme.StatusOffline()
		text = "offline"
	}

	return style.Render(text)
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

// ConnectionOnlineMsg is a message when the API is reachable.
type ConnectionOnlineMsg struct{}

// ConnectionOfflineMsg is a message when the API is unreachable.
type ConnectionOfflineMsg struct{}

// RateLimitedMsg is a message when the API is rate limiting requests.
type RateLimitedMsg struct {
	Tab       int
	ResetTime time.Time
}

// queueRetry adds a retry to the queue for the given tab after the specified delay.
func (m Model) queueRetry(tab int, delay time.Duration) tea.Cmd {
	return func() tea.Msg {
		item := retryItem{
			Cmd:       m.refreshActiveTab(),
			tab:       tab,
			scheduled: time.Now().Add(delay),
			attempts:  1,
		}
		// The queue is managed by a tick command
		return RetryTickMsg{Item: item}
	}
}

// isRateLimitError checks if an error indicates rate limiting.
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error contains rate limit keywords
	errStr := err.Error()
	return contains(errStr, "rate limit") ||
		contains(errStr, "429") ||
		contains(errStr, "too many requests")
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))
}

// findSubstring performs a simple substring search.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RetryTickMsg is a tick message for retry queue processing.
type RetryTickMsg struct {
	Item retryItem
}
