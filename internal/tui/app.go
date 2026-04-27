// Package tui provides the terminal user interface for finterm.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/cache"
	"github.com/shinsekai/finterm/internal/config"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
	"github.com/shinsekai/finterm/internal/tui/chart"
	"github.com/shinsekai/finterm/internal/tui/components"
	"github.com/shinsekai/finterm/internal/tui/macro"
	"github.com/shinsekai/finterm/internal/tui/news"
	palettepkg "github.com/shinsekai/finterm/internal/tui/palette"
	"github.com/shinsekai/finterm/internal/tui/quote"
	"github.com/shinsekai/finterm/internal/tui/trend"
)

const (
	tabTrend = iota
	tabQuote
	tabMacro
	tabNews
	tabChart
	numTabs
)

// globalBindings are keyboard bindings available in all views.
var globalBindings = []components.Binding{
	{Key: "1-5", Description: "Switch tab"},
	{Key: "Tab", Description: "Cycle tabs"},
	{Key: "r", Description: "Refresh"},
	{Key: "Ctrl+P", Description: "Command palette"},
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
	cmdPalette      *palettepkg.Model
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
	avClient *alphavantage.Client,
	cacheStore cache.Cache,
	watchlist *config.WatchlistConfig,
	detector *indicators.AssetClassDetector,
	cfg *config.Config,
) Model {
	// Create and configure trend model
	trendModel := trend.NewModel()
	trendModel.Configure(context.Background(), trendEngine, watchlist, detector)

	// Create and configure quote model
	quoteModel := quote.NewModel()
	quoteModel.Configure(context.Background(), quoteClient, trendEngine, detector)

	// Create and configure macro model
	macroModel := macro.NewModel()
	macroModel.Configure(context.Background(), macroClient, cacheStore)

	// Create and configure news model
	newsModel := news.NewModel()
	newsModel.Configure(context.Background(), newsClient, watchlist.Crypto)

	// Create and configure chart model
	chartModel := chart.NewModel()
	// Create crypto fetcher for chart (will be nil if avClient is nil)
	var cryptoFetcher *cryptoFetcherAdapter
	if avClient != nil {
		cryptoFetcher = &cryptoFetcherAdapter{client: avClient}
	}
	chartModel = chartModel.Configure(context.Background(), trendEngine, avClient, cacheStore, watchlist, detector, cfg, cryptoFetcher)

	// Create command palette with default commands
	cmdPalette := palettepkg.New(palettepkg.BuildDefaultCommands(watchlist))
	// Apply theme colors to palette
	cmdPalette.SetBoxStyle(theme.Box())
	cmdPalette.SetTitleStyle(theme.BoxTitle())
	cmdPalette.SetCursorStyle(lipgloss.NewStyle().
		Background(theme.Primary()).
		Foreground(theme.Background()))
	cmdPalette.SetLabelStyle(theme.Subtitle())
	cmdPalette.SetDescStyle(theme.Muted())
	cmdPalette.SetDimStyle(theme.Muted())

	return Model{
		theme: theme,
		tabs: []tab{
			{name: "Trend", model: trendModel},
			{name: "Quote", model: quoteModel},
			{name: "Macro", model: macroModel},
			{name: "News", model: newsModel},
			{name: "Chart", model: chartModel},
		},
		activeTab:       tabTrend,
		helpOverlay:     nil,
		cmdPalette:      cmdPalette,
		lastUpdate:      time.Now(),
		errorCount:      0,
		quit:            false,
		connectionState: ConnOnline,
		rateLimitReset:  time.Time{},
		retryQueue:      []retryItem{},
	}
}

// cryptoFetcherAdapter adapts the Alpha Vantage client to chart.CryptoDataFetcher interface.
type cryptoFetcherAdapter struct {
	client *alphavantage.Client
}

// FetchCryptoOHLCV fetches and converts crypto daily OHLCV data to domain types.
func (a *cryptoFetcherAdapter) FetchCryptoOHLCV(ctx context.Context, symbol string) ([]indicators.OHLCV, error) {
	data, err := a.client.GetCryptoDaily(ctx, symbol, "USD")
	if err != nil {
		return nil, err
	}

	// Get today's date in UTC to skip the in-progress bar
	today := time.Now().UTC().Format("2006-01-02")

	// Convert to OHLCV slice
	ohlcvSlice := make([]indicators.OHLCV, 0, len(data.TimeSeries))
	for dateStr, entry := range data.TimeSeries {
		// Skip today's in-progress bar (bar-close-only rule)
		if dateStr >= today {
			continue
		}

		date, err := alphavantage.ParseDate(dateStr)
		if err != nil {
			continue
		}
		open, _ := alphavantage.ParseFloat(entry.Open)
		high, _ := alphavantage.ParseFloat(entry.High)
		low, _ := alphavantage.ParseFloat(entry.Low)
		closeVal, _ := alphavantage.ParseFloat(entry.Close)
		volume, _ := alphavantage.ParseFloat(entry.Volume)

		ohlcvSlice = append(ohlcvSlice, indicators.OHLCV{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closeVal,
			Volume: volume,
		})
	}

	// Sort oldest-first
	for i := 0; i < len(ohlcvSlice)-1; i++ {
		for j := i + 1; j < len(ohlcvSlice); j++ {
			if ohlcvSlice[i].Date.After(ohlcvSlice[j].Date) {
				ohlcvSlice[i], ohlcvSlice[j] = ohlcvSlice[j], ohlcvSlice[i]
			}
		}
	}

	return ohlcvSlice, nil
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

	case palettepkg.CloseMsg:
		// Dismiss palette
		m.cmdPalette.Hide()
		return m, nil

	case palettepkg.ExecuteMsg:
		// Execute palette command and close palette
		m.cmdPalette.Hide()
		return m, msg.Cmd

	case palettepkg.SwitchTabMsg:
		// Switch to specific tab
		if msg.Tab >= 0 && msg.Tab < numTabs {
			m.activeTab = msg.Tab
			return m, m.refreshActiveTab()
		}
		return m, nil

	case palettepkg.RefreshCurrentTabMsg:
		// Refresh current tab
		return m, m.refreshActiveTab()

	case palettepkg.ChangeThemeMsg:
		// Theme change would require rebuilding the app, not implemented in current architecture
		// For now, just log it
		return m, nil

	case palettepkg.ShowHelpMsg:
		// Show help overlay
		return m.showHelpOverlay()

	case palettepkg.OpenQuoteWithTickerMsg:
		// Switch to quote tab and preload ticker
		m.setActiveTab(tabQuote)
		// TODO: Send message to quote model to prepopulate ticker
		// This would require quote model to accept a prepopulate message
		return m, nil

	case tea.KeyMsg:
		// Intercept Ctrl+P and Ctrl+K to open palette (if palette is not already visible)
		if msg.Type == tea.KeyCtrlP || msg.Type == tea.KeyCtrlK {
			if !m.cmdPalette.IsVisible() {
				m.cmdPalette.Show()
				return m, nil
			}
		}

		// If palette is visible, route all keys to palette
		if m.cmdPalette.IsVisible() {
			overlay, cmd := m.cmdPalette.Update(msg)
			m.cmdPalette = overlay.(*palettepkg.Model)
			return m, cmd
		}

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

		// Propagate adjusted size to all child models
		childMsg := tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: msg.Height - 4, // Reserve for tab bar + status bar
		}
		for i := range m.tabs {
			m.tabs[i].model, _ = m.tabs[i].model.Update(childMsg)
		}

		if m.helpOverlay != nil {
			overlay, cmd := m.helpOverlay.Update(msg)
			m.helpOverlay = overlay.(*components.HelpOverlay)
			return m, cmd
		}

		if m.cmdPalette.IsVisible() {
			overlay, cmd := m.cmdPalette.Update(msg)
			m.cmdPalette = overlay.(*palettepkg.Model)
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

	// Calculate available heights
	tabBarHeight := 3    // tab row + accent line + newline
	statusBarHeight := 3 // separator + status bar + newline
	contentHeight := m.height - tabBarHeight - statusBarHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Render tab bar
	tabBar := m.renderTabBar()

	// Render content (help, palette, or active child)
	var content string
	switch {
	case m.cmdPalette.IsVisible():
		// Render active child in background, then overlay palette
		content = m.tabs[m.activeTab].model.View()
	case m.helpOverlay != nil:
		content = m.helpOverlay.View()
	default:
		content = m.tabs[m.activeTab].model.View()
	}

	// Constrain content height using lipgloss
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		MaxHeight(contentHeight)
	content = contentStyle.Render(content)

	// Render status bar
	statusBar := m.renderStatusBar()

	// Join vertically
	baseView := lipgloss.JoinVertical(lipgloss.Left, tabBar, content, statusBar)

	// Overlay palette if visible
	if m.cmdPalette.IsVisible() {
		return m.cmdPalette.View()
	}

	return baseView
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
			return m, m.refreshActiveTab()

		case "2":
			m.setActiveTab(tabQuote)
			return m, m.refreshActiveTab()

		case "3":
			m.setActiveTab(tabMacro)
			return m, m.refreshActiveTab()

		case "4":
			m.setActiveTab(tabNews)
			return m, m.refreshActiveTab()

		case "5":
			m.setActiveTab(tabChart)
			return m, m.refreshActiveTab()
		}

	case tea.KeyTab:
		m.cycleTab()
		return m, m.refreshActiveTab()
	}

	// Delegate to active child model
	return m.delegateToChild(msg)
}

// renderTabBar renders the tab navigation bar.
func (m Model) renderTabBar() string {
	tabIcons := []string{"◆", "◈", "◇", "◉", "⬡"}
	var tabs []string
	for i, tab := range m.tabs {
		style := m.theme.Tab()
		if i == m.activeTab {
			style = m.theme.TabActive()
		}
		tabText := fmt.Sprintf(" %s %s ", tabIcons[i], tab.name)
		tabs = append(tabs, style.Render(tabText))
	}

	// Join tabs with subtle separator
	divider := m.theme.Divider().Render("│")
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs[0], divider, tabs[1], divider, tabs[2], divider, tabs[3], divider, tabs[4])

	// Add full-width bottom accent line
	accentLine := m.theme.Divider().Render(strings.Repeat("━", m.width))

	return lipgloss.JoinVertical(lipgloss.Left, tabRow, accentLine) + "\n"
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
	// Top separator line
	separator := m.theme.Divider().Render(strings.Repeat("─", m.width))

	// Left side: connection state + last update time
	left := fmt.Sprintf("%s | %s", m.renderConnectionState(), m.theme.Muted().Render(m.formatLastUpdate()))

	// Right side: error count with icon + key hints
	var errorText string
	if m.errorCount > 0 {
		errorText = fmt.Sprintf("✗ %d errors", m.errorCount)
	} else {
		errorText = "✓ no errors"
	}
	right := fmt.Sprintf("%s  %s", errorText, m.theme.Muted().Render("?:help  q:quit"))

	statusBar := m.theme.StatusBar().Render(lipgloss.JoinHorizontal(lipgloss.Bottom,
		left,
		lipgloss.PlaceHorizontal(m.width, lipgloss.Right, right)))

	return lipgloss.JoinVertical(lipgloss.Left, separator, statusBar) + "\n"
}

// renderConnectionState renders the connection state with appropriate styling.
func (m Model) renderConnectionState() string {
	var style lipgloss.Style
	var text string

	switch m.connectionState {
	case ConnOnline:
		style = m.theme.StatusOnline()
		text = "● online"
	case ConnRateLimited:
		style = m.theme.StatusRateLimited()
		text = "● rate limited"
	case ConnOffline:
		style = m.theme.StatusOffline()
		text = "● offline"
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
	case tabChart:
		return func() tea.Msg { return chart.RefreshMsg{} }
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
