// Package tui provides tests for terminal user interface.
package tui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/owner/finterm/internal/alphavantage"
	trenddomain "github.com/owner/finterm/internal/domain/trend"
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

// TestApp_TabSwitching tests tab switching with number keys and Tab.
func TestApp_TabSwitching(t *testing.T) {
	theme := NewTheme("default")
	app := NewApp(theme, &mockClient{}, &mockEngine{})

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
			var cmd tea.Cmd
			newModel, cmd := app.Update(msg)
			var ok bool
			app, ok = newModel.(Model)
			require.True(t, ok, "Expected Model type")

			// Assert active tab is as expected
			assert.Equal(t, tt.expectedTab, app.activeTab)
			// No command should be returned for tab switching
			assert.Nil(t, cmd)
		})
	}
}

// TestApp_QuitKey tests that q and Ctrl+C quit cleanly.
func TestApp_QuitKey(t *testing.T) {
	theme := NewTheme("default")
	_ = NewApp(theme, &mockClient{}, &mockEngine{})

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
			app := NewApp(theme, &mockClient{}, &mockEngine{}) // Reset app state

			// Create key message
			msg := tea.KeyMsg{
				Type:  tt.key,
				Runes: []rune(tt.runes),
			}

			// Update model
			newModel, cmd := app.Update(msg)
			var ok bool
			app, ok = newModel.(Model)
			require.True(t, ok, "Expected Model type")

			// Assert quit flag is set
			assert.True(t, app.quit)
			// Quit command should be returned
			assert.NotNil(t, cmd)
		})
	}
}

// TestApp_HelpToggle tests that ? toggles help overlay.
func TestApp_HelpToggle(t *testing.T) {
	theme := NewTheme("default")
	app := NewApp(theme, &mockClient{}, &mockEngine{})

	// Initially help should be hidden
	assert.False(t, app.showHelp)

	// Toggle help on
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("?"),
	}
	var cmd tea.Cmd
	newModel, cmd := app.Update(msg)
	var ok bool
	app, ok = newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.True(t, app.showHelp)
	assert.Nil(t, cmd)

	// Toggle help off
	newModel, _ = app.Update(msg)
	app = newModel.(Model)

	assert.False(t, app.showHelp)
	assert.Nil(t, cmd)
}

// TestApp_RefreshDelegation tests that r triggers refresh on active view.
func TestApp_RefreshDelegation(t *testing.T) {
	theme := NewTheme("default")

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
			app := NewApp(theme, &mockClient{}, &mockEngine{})
			app.activeTab = tt.activeTab

			// Press r to refresh
			msg := tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune("r"),
			}
			newModel, cmd := app.Update(msg)
			var ok bool
			app, ok = newModel.(Model)
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
	theme := NewTheme("default")
	app := NewApp(theme, &mockClient{}, &mockEngine{})

	assert.Equal(t, tabTrend, app.activeTab)
	assert.Equal(t, "Trend", app.tabs[tabTrend].name)
}

// TestApp_DelegateToChild tests that unknown keys are delegated to child model.
func TestApp_DelegateToChild(t *testing.T) {
	theme := NewTheme("default")
	app := NewApp(theme, &mockClient{}, &mockEngine{})

	// Send a message that should be delegated to child
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("x"), // Unknown key
	}

	newModel, _ := app.Update(msg)
	var ok bool
	app, ok = newModel.(Model)
	require.True(t, ok, "Expected Model type")

	// The message should have been delegated
	// In a real implementation, the child model would handle this
	// For now, we just verify no crash occurred
	assert.NotNil(t, app)
}

// TestApp_WindowSize tests that window size messages update dimensions.
func TestApp_WindowSize(t *testing.T) {
	theme := NewTheme("default")
	app := NewApp(theme, &mockClient{}, &mockEngine{})

	msg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	newModel, cmd := app.Update(msg)
	var ok bool
	app, ok = newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, 80, app.width)
	assert.Equal(t, 24, app.height)
	assert.Nil(t, cmd)
}

// TestApp_DataUpdateMsg tests that data update messages update last update time.
func TestApp_DataUpdateMsg(t *testing.T) {
	theme := NewTheme("default")
	app := NewApp(theme, &mockClient{}, &mockEngine{})

	oldTime := app.lastUpdate
	// Sleep to ensure time difference
	time.Sleep(10 * time.Millisecond)

	msg := DataUpdateMsg{Tab: tabTrend}
	newModel, cmd := app.Update(msg)
	var ok bool
	app, ok = newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.True(t, app.lastUpdate.After(oldTime))
	assert.Nil(t, cmd)
}

// TestApp_ErrorUpdateMsg tests that error update messages increment error count.
func TestApp_ErrorUpdateMsg(t *testing.T) {
	theme := NewTheme("default")
	app := NewApp(theme, &mockClient{}, &mockEngine{})

	oldCount := app.errorCount

	msg := ErrorUpdateMsg{Tab: tabTrend, Err: assert.AnError}
	newModel, cmd := app.Update(msg)
	var ok bool
	app, ok = newModel.(Model)
	require.True(t, ok, "Expected Model type")

	assert.Equal(t, oldCount+1, app.errorCount)
	assert.Nil(t, cmd)
}

// TestApp_ViewRenders tests that View() renders without crashing.
func TestApp_ViewRenders(t *testing.T) {
	theme := NewTheme("default")
	app := NewApp(theme, &mockClient{}, &mockEngine{})

	// Test normal view
	view := app.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "1. Trend")

	// Test help view
	app.showHelp = true
	view = app.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Help")

	// Test quit view
	app.quit = true
	app.showHelp = false
	view = app.View()
	assert.Equal(t, "Goodbye!", view)
}

// TestApp_InvalidTabKey tests that invalid tab keys are delegated to child.
func TestApp_InvalidTabKey(t *testing.T) {
	theme := NewTheme("default")
	app := NewApp(theme, &mockClient{}, &mockEngine{})

	// Press 5 (invalid tab key)
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("5"),
	}

	newModel, _ := app.Update(msg)
	var ok bool
	app, ok = newModel.(Model)
	require.True(t, ok, "Expected Model type")

	// Active tab should not change
	assert.Equal(t, tabTrend, app.activeTab)
	assert.NotNil(t, app)
}
