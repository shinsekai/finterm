// Package trend provides tests for the trend following TUI model.
package trend

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/owner/finterm/internal/config"
	trenddomain "github.com/owner/finterm/internal/domain/trend"
	"github.com/owner/finterm/internal/domain/trend/indicators"
)

// mockEngine is a mock implementation of Engine for testing.
type mockEngine struct {
	result *trenddomain.Result
	err    error
}

func (m *mockEngine) AnalyzeWithSymbolDetection(_ context.Context, _ string) (*trenddomain.Result, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// asModel converts a tea.Model to *Model, handling both value and pointer types.
func asModel(t *testing.T, m tea.Model) *Model {
	t.Helper()
	if ptrModel, ok := m.(*Model); ok {
		return ptrModel
	}
	if model, ok := m.(Model); ok {
		// Return a pointer to a copy of the value
		return &model
	}
	require.Fail(t, "expected Model or *Model, got %T", m)
	return nil
}

// newTestModel creates a configured model for testing.
func newTestModel(t *testing.T, engine Engine, symbols ...string) *Model {
	t.Helper()

	equities := append([]string{}, symbols...)
	var crypto []string

	model := NewModel()
	model.Configure(
		context.Background(),
		engine,
		&config.WatchlistConfig{Equities: equities, Crypto: crypto},
		indicators.NewAssetClassDetector([]string{}),
	)
	return model
}

// TestTrendModel_Init_DispatchesCommands verifies that Init returns a command.
func TestTrendModel_Init_DispatchesCommands(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	cmd := model.Init()
	assert.NotNil(t, cmd, "Init should return a command")
}

// TestTrendModel_Update_DataMsg verifies handling of TrendDataMsg.
func TestTrendModel_Update_DataMsg(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	// Send a TrendDataMsg
	result := &trenddomain.Result{
		Symbol:  "AAPL",
		RSI:     50.5,
		EMAFast: 150.25,
		EMASlow: 145.75,
		Signal:  trenddomain.Bullish,
	}
	msg := TrendDataMsg{Symbol: "AAPL", Result: result}

	updatedM, cmd := model.Update(msg)
	updatedModel := asModel(t, updatedM)
	assert.Nil(t, cmd)

	// Verify the row was updated
	rows := updatedModel.GetRows()
	require.Len(t, rows, 1, "Should have 1 row")
	assert.Equal(t, StateLoaded, rows[0].State, "Row state should be Loaded")
	assert.Equal(t, result, rows[0].Result, "Row result should match")
	assert.Nil(t, rows[0].Error, "Row error should be nil")
}

// TestTrendModel_Update_ErrorMsg verifies handling of TrendErrorMsg.
func TestTrendModel_Update_ErrorMsg(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	// Send a TrendErrorMsg
	err := errors.New("API error")
	msg := TrendErrorMsg{Symbol: "AAPL", Err: err}

	updatedM, cmd := model.Update(msg)
	updatedModel := asModel(t, updatedM)
	assert.Nil(t, cmd)

	// Verify the row was updated with error
	rows := updatedModel.GetRows()
	require.Len(t, rows, 1, "Should have 1 row")
	assert.Equal(t, StateError, rows[0].State, "Row state should be Error")
	assert.Equal(t, err, rows[0].Error, "Row error should match")
	assert.Nil(t, rows[0].Result, "Row result should be nil")
}

// TestTrendModel_Update_ArrowNavigation verifies arrow key navigation.
func TestTrendModel_Update_ArrowNavigation(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")

	t.Run("ArrowDown moves to next row", func(t *testing.T) {
		// Start at row 0
		assert.Equal(t, 0, model.GetActiveRow(), "Initial active row should be 0")

		// Press down
		msg := tea.KeyMsg{Type: tea.KeyDown}
		updatedM, _ := model.Update(msg)
		updatedModel := asModel(t, updatedM)
		assert.Equal(t, 1, updatedModel.GetActiveRow(), "Active row should move to 1")

		// Press down again
		updatedM, _ = updatedModel.Update(msg)
		updatedModel = asModel(t, updatedM)
		assert.Equal(t, 2, updatedModel.GetActiveRow(), "Active row should move to 2")
	})

	t.Run("ArrowUp moves to previous row", func(t *testing.T) {
		// Start at row 2 (navigate there first)
		model2 := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")
		msg := tea.KeyMsg{Type: tea.KeyDown}
		updatedM, _ := model2.Update(msg)
		model2 = asModel(t, updatedM)
		updatedM, _ = model2.Update(msg)
		model2 = asModel(t, updatedM)
		assert.Equal(t, 2, model2.GetActiveRow(), "Should be at row 2")

		// Press up
		updatedM, _ = model2.Update(tea.KeyMsg{Type: tea.KeyUp})
		updatedModel := asModel(t, updatedM)
		assert.Equal(t, 1, updatedModel.GetActiveRow(), "Active row should move to 1")

		// Press up again
		updatedM, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyUp})
		updatedModel = asModel(t, updatedM)
		assert.Equal(t, 0, updatedModel.GetActiveRow(), "Active row should move to 0")
	})

	t.Run("ArrowDown does not go past last row", func(t *testing.T) {
		// Navigate to last row
		model2 := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")
		msg := tea.KeyMsg{Type: tea.KeyDown}
		updatedM, _ := model2.Update(msg)
		model2 = asModel(t, updatedM)
		updatedM, _ = model2.Update(msg)
		model2 = asModel(t, updatedM)
		assert.Equal(t, 2, model2.GetActiveRow(), "Should be at last row")

		// Try to go past
		updatedM, _ = model2.Update(msg)
		updatedModel := asModel(t, updatedM)
		assert.Equal(t, 2, updatedModel.GetActiveRow(), "Active row should stay at 2")
	})

	t.Run("ArrowUp does not go past first row", func(t *testing.T) {
		// Start at row 0
		model2 := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")
		assert.Equal(t, 0, model2.GetActiveRow(), "Should be at row 0")

		// Try to go past
		updatedM, _ := model2.Update(tea.KeyMsg{Type: tea.KeyUp})
		updatedModel := asModel(t, updatedM)
		assert.Equal(t, 0, updatedModel.GetActiveRow(), "Active row should stay at 0")
	})
}

// TestTrendModel_Update_RefreshMsg verifies handling of RefreshMsg.
func TestTrendModel_Update_RefreshMsg(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT")

	// Set one row to loaded state
	result := &trenddomain.Result{Symbol: "AAPL", RSI: 50.5}
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: result})
	model = asModel(t, updatedM)

	// Verify row is loaded
	rows := model.GetRows()
	assert.Equal(t, StateLoaded, rows[0].State, "First row should be loaded")

	// Send refresh
	refreshMsg := RefreshMsg{}
	updatedM, _ = model.Update(refreshMsg)
	updatedModel := asModel(t, updatedM)

	// Verify all rows are reset to loading
	rows = updatedModel.GetRows()
	for i, row := range rows {
		assert.Equal(t, StateLoading, row.State, "Row %d should be in loading state after refresh", i)
		assert.Nil(t, row.Result, "Row %d result should be nil after refresh", i)
		assert.Nil(t, row.Error, "Row %d error should be nil after refresh", i)
	}
}

// TestTrendModel_Update_WindowSizeMsg verifies handling of WindowSizeMsg.
func TestTrendModel_Update_WindowSizeMsg(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedM, _ := model.Update(msg)
	updatedModel := asModel(t, updatedM)
	assert.Equal(t, 100, updatedModel.GetWidth(), "Width should be updated")
	assert.Equal(t, 50, updatedModel.GetHeight(), "Height should be updated")
}

// TestTrendModel_Update_FetchCompleteMsg verifies handling of FetchCompleteMsg.
func TestTrendModel_Update_FetchCompleteMsg(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	// Load data for the ticker
	result := &trenddomain.Result{Symbol: "AAPL", RSI: 50.5}
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: result})
	model = asModel(t, updatedM)

	// Send fetch complete message
	msg := FetchCompleteMsg{}
	updatedM, _ = model.Update(msg)
	updatedModel := asModel(t, updatedM)
	assert.Equal(t, StateLoaded, updatedModel.GetOverallState(), "Overall state should be loaded")
}

// TestTrendModel_Configure verifies the Configure method properly sets up the model.
func TestTrendModel_Configure(t *testing.T) {
	model := NewModel()

	watchlist := &config.WatchlistConfig{
		Equities: []string{"AAPL", "MSFT"},
		Crypto:   []string{"BTC", "ETH"},
	}
	detector := indicators.NewAssetClassDetector([]string{"BTC", "ETH"})
	engine := &mockEngine{}

	model.Configure(context.Background(), engine, watchlist, detector)

	assert.Equal(t, engine, model.engine, "Engine should be set")
	assert.Equal(t, detector, model.detector, "Detector should be set")
	assert.NotNil(t, model.ctx, "Context should be set")
	assert.NotNil(t, model.cancel, "Cancel function should be set")

	// Verify watchlist is combined
	assert.Len(t, model.watchlist, 4, "Watchlist should have 4 items")
	assert.Contains(t, model.watchlist, "AAPL", "Watchlist should contain AAPL")
	assert.Contains(t, model.watchlist, "MSFT", "Watchlist should contain MSFT")
	assert.Contains(t, model.watchlist, "BTC", "Watchlist should contain BTC")
	assert.Contains(t, model.watchlist, "ETH", "Watchlist should contain ETH")

	// Verify rows are initialized
	rows := model.GetRows()
	assert.Len(t, rows, 4, "Should have 4 rows")
	for _, row := range rows {
		assert.Equal(t, StateLoading, row.State, "All rows should be in loading state initially")
	}
}

// TestTrendModel_Update_MixedState verifies handling of mixed loading/loaded/error states.
func TestTrendModel_Update_MixedState(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")

	// Set first row to loaded
	result := &trenddomain.Result{Symbol: "AAPL", RSI: 50.5}
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: result})
	model = asModel(t, updatedM)

	// Set second row to error
	err := errors.New("network error")
	updatedM, _ = model.Update(TrendErrorMsg{Symbol: "MSFT", Err: err})
	model = asModel(t, updatedM)

	// Third row remains in loading state

	// Send fetch complete to update overall state
	updatedM, _ = model.Update(FetchCompleteMsg{})
	model = asModel(t, updatedM)

	// Verify mixed states
	rows := model.GetRows()
	assert.Equal(t, StateLoaded, rows[0].State, "First row should be loaded")
	assert.Equal(t, StateError, rows[1].State, "Second row should be error")
	assert.Equal(t, StateLoading, rows[2].State, "Third row should be loading")

	// Overall state should still be loading since not all are done
	assert.Equal(t, StateLoading, model.GetOverallState(), "Overall state should be loading")
}

// TestTrendModel_ContextCancellation verifies that the context is properly set up.
func TestTrendModel_ContextCancellation(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	assert.NotNil(t, model.cancel, "Cancel function should be set")

	// Call cancel to verify it doesn't panic
	model.cancel()
}

// TestTrendModel_RefreshKey verifies pressing 'r' triggers refresh.
func TestTrendModel_RefreshKey(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	// Set row to loaded
	result := &trenddomain.Result{Symbol: "AAPL", RSI: 50.5}
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: result})
	model = asModel(t, updatedM)
	assert.Equal(t, StateLoaded, model.GetRows()[0].State)

	// Press 'r'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedM, _ = model.Update(msg)
	updatedModel := asModel(t, updatedM)
	assert.Equal(t, StateLoading, updatedModel.GetRows()[0].State, "Row should be in loading state after refresh")
}

// TestFormatValue verifies the FormatValue helper function.
func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		decimals int
		expected string
	}{
		{"RSI value", 50.5678, 2, "50.57"},
		{"EMA value", 150.1234, 4, "150.1234"},
		{"No decimals", 100.999, 0, "101"},
		{"Negative value", -5.25, 2, "-5.25"},
		{"Zero", 0.0, 2, "0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValue(tt.value, tt.decimals)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestState_String verifies the String method for State.
func TestState_String(t *testing.T) {
	tests := []struct {
		name     string
		state    State
		expected string
	}{
		{"Loading", StateLoading, "Loading"},
		{"Loaded", StateLoaded, "Loaded"},
		{"Error", StateError, "Error"},
		{"Unknown", State(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
