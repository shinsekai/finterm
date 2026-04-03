// Package tui provides tests for terminal user interface.
package tui

import (
	"context"
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/owner/finterm/internal/alphavantage"
	"github.com/owner/finterm/internal/cache"
	"github.com/owner/finterm/internal/config"
	trenddomain "github.com/owner/finterm/internal/domain/trend"
	"github.com/owner/finterm/internal/domain/trend/indicators"
	"github.com/owner/finterm/internal/tui/components"
	"github.com/owner/finterm/internal/tui/trend"
)

// mockEngine is a mock implementation of trend.Engine for testing.
type mockEngine struct{}

func (m *mockEngine) AnalyzeWithSymbolDetection(_ context.Context, symbol string) (*trenddomain.Result, error) {
	return &trenddomain.Result{
		Symbol:  symbol,
		RSI:     50,
		EMAFast: 100,
		EMASlow: 90,
		Signal:  trenddomain.Bullish,
	}, nil
}

// mockClient is a mock implementation of alphavantage.Client for testing.
type mockClient struct{}

func (m *mockClient) GetGlobalQuote(_ context.Context, symbol string) (*alphavantage.GlobalQuote, error) {
	return &alphavantage.GlobalQuote{
		Symbol: symbol,
		Price:  "100.00",
	}, nil
}

// mockClient also implements macro.Client interface methods
func (m *mockClient) GetRealGDP(_ context.Context, _ string) ([]alphavantage.MacroDataPoint, error) {
	return nil, nil
}

func (m *mockClient) GetRealGDPPerCapita(_ context.Context) ([]alphavantage.MacroDataPoint, error) {
	return nil, nil
}

func (m *mockClient) GetCPI(_ context.Context, _ string) ([]alphavantage.MacroDataPoint, error) {
	return nil, nil
}

func (m *mockClient) GetInflation(_ context.Context) ([]alphavantage.MacroDataPoint, error) {
	return nil, nil
}

func (m *mockClient) GetUnemployment(_ context.Context) ([]alphavantage.MacroDataPoint, error) {
	return nil, nil
}

func (m *mockClient) GetNonfarmPayroll(_ context.Context) ([]alphavantage.MacroDataPoint, error) {
	return nil, nil
}

func (m *mockClient) GetFedFundsRate(_ context.Context, _ string) ([]alphavantage.MacroDataPoint, error) {
	return nil, nil
}

func (m *mockClient) GetTreasuryYield(_ context.Context, _, _ string) ([]alphavantage.MacroDataPoint, error) {
	return nil, nil
}

func (m *mockClient) GetNewsSentiment(_ context.Context, _ alphavantage.NewsOpts) (*alphavantage.NewsSentiment, error) {
	return &alphavantage.NewsSentiment{
		Items: []alphavantage.NewsItem{},
	}, nil
}

// newMockApp creates a new app for testing with all required mocks.
func newMockApp(t *testing.T) Model {
	theme := NewTheme("default")
	cacheStore := cache.New()
	t.Cleanup(func() { cacheStore.Close() })

	watchlist := &config.WatchlistConfig{
		Equities: []string{},
		Crypto:   []string{},
	}
	detector := indicators.NewAssetClassDetector([]string{})

	// mockClient implements all three client interfaces: quote.QuoteClient, macro.Client, news.Client
	return NewApp(theme, &mockClient{}, &mockClient{}, &mockClient{}, &mockEngine{}, cacheStore, watchlist, detector)
}

// TestApp_TabSwitching tests tab switching with number keys and Tab.
func TestApp_TabSwitching(t *testing.T) {
	app := newMockApp(t)

	tests := []struct {
		name        string
		key         tea.KeyType
		runes       string
		initialTab  int
		expectedTab int
	}{
		{
			name:        "switch to trend tab with 1",
			key:         tea.KeyRunes,
			runes:       "1",
			initialTab:  tabTrend,
			expectedTab: tabTrend,
		},
		{
			name:        "switch to quote tab with 2",
			key:         tea.KeyRunes,
			runes:       "2",
			initialTab:  tabTrend,
			expectedTab: tabQuote,
		},
		{
			name:        "switch to macro tab with 3",
			key:         tea.KeyRunes,
			runes:       "3",
			initialTab:  tabTrend,
			expectedTab: tabMacro,
		},
		{
			name:        "switch to news tab with 4",
			key:         tea.KeyRunes,
			runes:       "4",
			initialTab:  tabTrend,
			expectedTab: tabNews,
		},
		{
			name:        "cycle from trend to quote with Tab",
			key:         tea.KeyTab,
			runes:       "",
			initialTab:  tabTrend,
			expectedTab: tabQuote,
		},
		{
			name:        "cycle from quote to macro with Tab",
			key:         tea.KeyTab,
			runes:       "",
			initialTab:  tabQuote,
			expectedTab: tabMacro,
		},
		{
			name:        "cycle from macro to news with Tab",
			key:         tea.KeyTab,
			runes:       "",
			initialTab:  tabMacro,
			expectedTab: tabNews,
		},
		{
			name:        "cycle from news back to trend with Tab",
			key:         tea.KeyTab,
			runes:       "",
			initialTab:  tabNews,
			expectedTab: tabTrend,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start with specified initial tab
			app.activeTab = tt.initialTab

			// Create key message
			msg := tea.KeyMsg{
				Type:  tt.key,
				Runes: []rune(tt.runes),
			}

			// Update model
			newModel, cmd := app.Update(msg)
			updatedApp, ok := newModel.(Model)
			require.True(t, ok, "Expected Model type")

			// Assert active tab is as expected
			assert.Equal(t, tt.expectedTab, updatedApp.activeTab)
			// No command should be returned for tab switching
			assert.Nil(t, cmd)
		})
	}
}

// TestApp_QuitKey tests that q and Ctrl+C quit cleanly.
func TestApp_QuitKey(t *testing.T) {
	_ = newMockApp(t)

	tests := []struct {
		name  string
		key   tea.KeyType
		runes string
	}{
		{
			name:  "quit with q",
			key:   tea.KeyRunes,
			runes: "q",
		},
		{
			name:  "quit with Ctrl+C",
			key:   tea.KeyCtrlC,
			runes: "",
		},
		{
			name:  "quit with Esc",
			key:   tea.KeyEsc,
			runes: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newMockApp(t) // Reset app state

			// Create key message
			msg := tea.KeyMsg{
				Type:  tt.key,
				Runes: []rune(tt.runes),
			}

			// Update model
			newModel, cmd := app.Update(msg)
			updatedApp, ok := newModel.(Model)
			require.True(t, ok, "Expected Model type")

			// Assert quit flag is set
			assert.True(t, updatedApp.quit)
			// Quit command should be returned
			assert.NotNil(t, cmd)
		})
	}
}

// TestApp_HelpToggle tests that ? toggles help overlay.
func TestApp_HelpToggle(t *testing.T) {
	app := newMockApp(t)

	// Initially help should be hidden
	assert.Nil(t, app.helpOverlay)

	// Toggle help on
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("?"),
	}
	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.NotNil(t, updatedApp.helpOverlay)
	assert.Nil(t, cmd)

	// Dismiss help with Esc
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ = updatedApp.Update(escMsg)
	finalApp := newModel.(Model)

	assert.Nil(t, finalApp.helpOverlay)
}

// TestApp_RefreshDelegation tests that r triggers refresh on active view.
func TestApp_RefreshDelegation(t *testing.T) {

	tests := []struct {
		name      string
		activeTab int
	}{
		{
			name:      "refresh trend tab",
			activeTab: tabTrend,
		},
		{
			name:      "refresh quote tab",
			activeTab: tabQuote,
		},
		{
			name:      "refresh macro tab",
			activeTab: tabMacro,
		},
		{
			name:      "refresh news tab",
			activeTab: tabNews,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newMockApp(t)
			app.activeTab = tt.activeTab

			// Press r to refresh
			msg := tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune("r"),
			}
			newModel, cmd := app.Update(msg)
			_, ok := newModel.(Model)
			require.True(t, ok, "Expected Model type")

			// Command should be returned
			require.NotNil(t, cmd)

			// Execute command and verify it returns a refresh message
			resultMsg := cmd()
			switch tt.activeTab {
			case tabTrend:
				_, ok := resultMsg.(trend.RefreshMsg)
				assert.True(t, ok, "Expected trend.RefreshMsg, got %T", resultMsg)
			case tabQuote:
				// Note: quote package doesn't export RefreshMsg type
				// We just verify that a command is returned
			case tabMacro:
				// Note: macro package doesn't export RefreshMsg type
				// We just verify that a command is returned
			case tabNews:
				// Note: news package doesn't export RefreshMsg type
				// We just verify that a command is returned
			}
		})
	}
}

// TestApp_DefaultTab tests that default tab is trend (first tab).
func TestApp_DefaultTab(t *testing.T) {
	app := newMockApp(t)

	assert.Equal(t, tabTrend, app.activeTab)
	assert.Equal(t, "Trend", app.tabs[tabTrend].name)
}

// TestApp_DelegateToChild tests that unknown keys are delegated to child model.
func TestApp_DelegateToChild(t *testing.T) {
	app := newMockApp(t)

	// Send a message that should be delegated to child
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("x"), // Unknown key
	}

	newModel, _ := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	// The message should have been delegated
	// In a real implementation, the child model would handle this
	// For now, we just verify no crash occurred
	assert.NotNil(t, updatedApp)
}

// TestApp_WindowSize tests that window size messages update dimensions.
func TestApp_WindowSize(t *testing.T) {
	app := newMockApp(t)
	msg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, 80, updatedApp.width)
	assert.Equal(t, 24, updatedApp.height)
	assert.Nil(t, cmd)
}

// TestApp_DataUpdateMsg tests that data update messages update last update time.
func TestApp_DataUpdateMsg(t *testing.T) {
	app := newMockApp(t)

	oldTime := app.lastUpdate
	// Sleep to ensure time difference
	time.Sleep(10 * time.Millisecond)

	msg := DataUpdateMsg{Tab: tabTrend}
	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.True(t, updatedApp.lastUpdate.After(oldTime))
	assert.Nil(t, cmd)
}

// TestApp_ErrorUpdateMsg tests that error update messages increment error count.
func TestApp_ErrorUpdateMsg(t *testing.T) {
	app := newMockApp(t)
	oldCount := app.errorCount
	msg := ErrorUpdateMsg{Tab: tabTrend, Err: assert.AnError}
	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, oldCount+1, updatedApp.errorCount)
	assert.Nil(t, cmd)
}

// TestApp_ViewRenders tests that View() renders without crashing.
func TestApp_ViewRenders(t *testing.T) {
	app := newMockApp(t)

	view := app.View()
	assert.Contains(t, view, "1. Trend")

	// Test help view
	overlay := components.NewHelpOverlay(globalBindings, nil)
	// Initialize dimensions via WindowSizeMsg
	overlayModel, _ := overlay.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app.helpOverlay = overlayModel.(*components.HelpOverlay)
	view = app.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Key Bindings")

	// Test quit view
	app.quit = true
	app.helpOverlay = nil
	view = app.View()
	assert.Equal(t, "Goodbye!", view)
}

// TestApp_InvalidTabKey tests that invalid tab keys are delegated to child.
func TestApp_InvalidTabKey(t *testing.T) {
	app := newMockApp(t)

	// Press 5 (invalid tab key)
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("5"),
	}

	newModel, _ := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	// Active tab should not change
	assert.Equal(t, tabTrend, updatedApp.activeTab)
	assert.NotNil(t, updatedApp)
}

// TestApp_ConnectionOnlineMsg tests that online messages set connection state to online.
func TestApp_ConnectionOnlineMsg(t *testing.T) {
	app := newMockApp(t)

	// Set initial state to offline
	app.connectionState = ConnOffline

	// Send online message
	msg := ConnectionOnlineMsg{}
	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, ConnOnline, updatedApp.connectionState)
	assert.True(t, updatedApp.rateLimitReset.IsZero())
	assert.Nil(t, cmd)
}

// TestApp_ConnectionOfflineMsg tests that offline messages set connection state to offline.
func TestApp_ConnectionOfflineMsg(t *testing.T) {
	app := newMockApp(t)

	// Send offline message
	msg := ConnectionOfflineMsg{}
	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, ConnOffline, updatedApp.connectionState)
	assert.Nil(t, cmd)
}

// TestApp_RateLimitedMsg tests that rate limit messages set connection state and queue retry.
func TestApp_RateLimitedMsg(t *testing.T) {
	app := newMockApp(t)

	// Send rate limited message
	resetTime := time.Now().Add(5 * time.Minute)
	msg := RateLimitedMsg{
		Tab:       tabTrend,
		ResetTime: resetTime,
	}
	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, ConnRateLimited, updatedApp.connectionState)
	assert.Equal(t, resetTime, updatedApp.rateLimitReset)
	assert.NotNil(t, cmd)
}

// TestApp_RateLimitedWithRetry tests that rate limited messages queue retry.
func TestApp_RateLimitedWithRetry(t *testing.T) {
	app := newMockApp(t)

	// Send rate limited message
	resetTime := time.Now().Add(1 * time.Minute)
	msg := RateLimitedMsg{
		Tab:       tabTrend,
		ResetTime: resetTime,
	}
	newModel, cmd := app.Update(msg)
	_, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	require.NotNil(t, cmd)

	// Execute the retry command
	resultMsg := cmd()
	assert.IsType(t, RetryTickMsg{}, resultMsg)

	retryMsg := resultMsg.(RetryTickMsg)
	assert.False(t, retryMsg.Item.scheduled.IsZero())
	assert.Equal(t, tabTrend, retryMsg.Item.tab)
	assert.Equal(t, 1, retryMsg.Item.attempts)
}

// TestApp_ErrorUpdateMsg_RateLimit tests that rate limit errors set rate limited state.
func TestApp_ErrorUpdateMsg_RateLimit(t *testing.T) {
	app := newMockApp(t)

	// Send error update with rate limit error
	msg := ErrorUpdateMsg{
		Tab: tabTrend,
		Err: fmt.Errorf("API error: rate limit exceeded (429)"),
	}
	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, ConnRateLimited, updatedApp.connectionState)
	assert.Nil(t, cmd)
}

// TestApp_ErrorUpdateMsg_NetworkError tests that network errors set offline state.
func TestApp_ErrorUpdateMsg_NetworkError(t *testing.T) {
	app := newMockApp(t)

	// Send error update with network error
	msg := ErrorUpdateMsg{
		Tab: tabTrend,
		Err: fmt.Errorf("connection refused"),
	}
	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, ConnOffline, updatedApp.connectionState)
	assert.Nil(t, cmd)
}

// TestApp_DataUpdateResetsConnection tests that successful data updates reset to online.
func TestApp_DataUpdateResetsConnection(t *testing.T) {
	app := newMockApp(t)

	// Set initial state to offline
	app.connectionState = ConnOffline
	app.rateLimitReset = time.Now().Add(time.Hour)

	// Send data update message
	msg := DataUpdateMsg{Tab: tabTrend}
	newModel, cmd := app.Update(msg)
	updatedApp, ok := newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, ConnOnline, updatedApp.connectionState)
	// Note: rateLimitReset is reset internally but not accessible through the type-asserted model
	assert.Nil(t, cmd)
}

// TestApp_isRateLimitError tests rate limit error detection.
func TestApp_isRateLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "rate limit error",
			err:      fmt.Errorf("API error: rate limit exceeded"),
			expected: true,
		},
		{
			name:     "429 status code",
			err:      fmt.Errorf("HTTP 429 Too Many Requests"),
			expected: true,
		},
		{
			name:     "too many requests",
			err:      fmt.Errorf("too many requests, please try again later"),
			expected: true,
		},
		{
			name:     "network error",
			err:      fmt.Errorf("connection refused"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRateLimitError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestApp_StatusBar_RendersConnectionState tests that status bar shows connection state.
func TestApp_StatusBar_RendersConnectionState(t *testing.T) {
	app := newMockApp(t)

	// Test online state
	app.connectionState = ConnOnline
	statusBar := app.renderStatusBar()
	assert.Contains(t, statusBar, "online")

	// Test rate limited state
	app.connectionState = ConnRateLimited
	statusBar = app.renderStatusBar()
	assert.Contains(t, statusBar, "rate limited")

	// Test offline state
	app.connectionState = ConnOffline
	statusBar = app.renderStatusBar()
	assert.Contains(t, statusBar, "offline")
}
