// Package trend provides tests for the trend following TUI model.
package trend

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shinsekai/finterm/internal/config"
	trenddomain "github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
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

// asModel converts a tea.Model to *Model.
func asModel(t *testing.T, m tea.Model) *Model {
	t.Helper()
	if ptrModel, ok := m.(*Model); ok {
		return ptrModel
	}
	require.Fail(t, "expected *Model, got %T", m)
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

// TestTrendModel_GetCryptoStartIndex_MixedWatchlist verifies GetCryptoStartIndex
// when the watchlist contains both equities and crypto.
func TestTrendModel_GetCryptoStartIndex_MixedWatchlist(t *testing.T) {
	model := NewModel()

	watchlist := &config.WatchlistConfig{
		Equities: []string{"AAPL", "MSFT", "GOOGL"},
		Crypto:   []string{"BTC", "ETH"},
	}

	model.Configure(context.Background(), &mockEngine{}, watchlist, indicators.NewAssetClassDetector([]string{}))

	assert.Equal(t, 3, model.GetCryptoStartIndex(), "Crypto should start at index 3 (after 3 equities)")
}

// TestTrendModel_GetCryptoStartIndex_EquitiesOnly verifies GetCryptoStartIndex
// when the watchlist contains only equities.
func TestTrendModel_GetCryptoStartIndex_EquitiesOnly(t *testing.T) {
	model := NewModel()

	watchlist := &config.WatchlistConfig{
		Equities: []string{"AAPL", "MSFT"},
		Crypto:   []string{},
	}

	model.Configure(context.Background(), &mockEngine{}, watchlist, indicators.NewAssetClassDetector([]string{}))

	assert.Equal(t, 2, model.GetCryptoStartIndex(), "Crypto start index should equal number of equities")
}

// TestTrendModel_GetCryptoStartIndex_CryptoOnly verifies GetCryptoStartIndex
// when the watchlist contains only crypto.
func TestTrendModel_GetCryptoStartIndex_CryptoOnly(t *testing.T) {
	model := NewModel()

	watchlist := &config.WatchlistConfig{
		Equities: []string{},
		Crypto:   []string{"BTC", "ETH", "SOL"},
	}

	model.Configure(context.Background(), &mockEngine{}, watchlist, indicators.NewAssetClassDetector([]string{}))

	assert.Equal(t, 0, model.GetCryptoStartIndex(), "Crypto should start at index 0 (no equities)")
}

// TestTrendModel_GetCryptoStartIndex_EmptyWatchlist verifies GetCryptoStartIndex
// when the watchlist is empty.
func TestTrendModel_GetCryptoStartIndex_EmptyWatchlist(t *testing.T) {
	model := NewModel()

	watchlist := &config.WatchlistConfig{
		Equities: []string{},
		Crypto:   []string{},
	}

	model.Configure(context.Background(), &mockEngine{}, watchlist, indicators.NewAssetClassDetector([]string{}))

	assert.Equal(t, 0, model.GetCryptoStartIndex(), "Crypto start index should be 0 for empty watchlist")
}

// TestTrendModel_GetLoadedCount_AllLoading verifies GetLoadedCount returns 0
// when all rows are in loading state.
func TestTrendModel_GetLoadedCount_AllLoading(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")

	// All rows start in loading state
	assert.Equal(t, 0, model.GetLoadedCount(), "Loaded count should be 0 when all are loading")
}

// TestTrendModel_GetLoadedCount_MixedStates verifies GetLoadedCount correctly
// counts loaded and cached rows but ignores loading and error states.
func TestTrendModel_GetLoadedCount_MixedStates(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL", "BTC")

	// Set some rows to loaded state
	result := &trenddomain.Result{Symbol: "AAPL", RSI: 50.5, Signal: trenddomain.Bullish}
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: result})
	model = asModel(t, updatedM)

	result2 := &trenddomain.Result{Symbol: "GOOGL", RSI: 45.0, Signal: trenddomain.Bearish}
	updatedM, _ = model.Update(TrendDataMsg{Symbol: "GOOGL", Result: result2})
	model = asModel(t, updatedM)

	// Set one row to error state
	err := errors.New("network error")
	updatedM, _ = model.Update(TrendErrorMsg{Symbol: "MSFT", Err: err})
	model = asModel(t, updatedM)

	// BTC remains in loading state
	// 2 loaded (AAPL, GOOGL) + 0 cached + 1 error (MSFT) + 1 loading (BTC) = 2 loaded
	assert.Equal(t, 2, model.GetLoadedCount(), "Loaded count should be 2")
}

// TestTrendModel_GetLoadedCount_IncludesCached verifies GetLoadedCount includes
// rows in StateCached.
func TestTrendModel_GetLoadedCount_IncludesCached(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT")

	// Set first row to loaded state
	result := &trenddomain.Result{Symbol: "AAPL", RSI: 50.5}
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: result})
	model = asModel(t, updatedM)

	// Manually set second row to cached state (simulating cached data)
	model.rows[1].State = StateCached

	assert.Equal(t, 2, model.GetLoadedCount(), "Loaded count should include cached rows")
}

// TestTrendModel_GetSignalCounts_MixedSignals verifies GetSignalCounts correctly
// counts bullish, bearish, and neutral signals.
func TestTrendModel_GetSignalCounts_MixedSignals(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL", "TSLA", "NVDA")

	// Set rows with various signals
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: &trenddomain.Result{Symbol: "AAPL", Signal: trenddomain.Bullish}})
	model = asModel(t, updatedM)

	updatedM, _ = model.Update(TrendDataMsg{Symbol: "MSFT", Result: &trenddomain.Result{Symbol: "MSFT", Signal: trenddomain.Bearish}})
	model = asModel(t, updatedM)

	updatedM, _ = model.Update(TrendDataMsg{Symbol: "GOOGL", Result: &trenddomain.Result{Symbol: "GOOGL", Signal: trenddomain.Bullish}})
	model = asModel(t, updatedM)

	updatedM, _ = model.Update(TrendDataMsg{Symbol: "TSLA", Result: &trenddomain.Result{Symbol: "TSLA", Signal: trenddomain.Bearish}})
	model = asModel(t, updatedM)

	updatedM, _ = model.Update(TrendDataMsg{Symbol: "NVDA", Result: &trenddomain.Result{Symbol: "NVDA", Signal: trenddomain.Bearish}})
	model = asModel(t, updatedM)

	bullish, bearish, neutral := model.GetSignalCounts()

	assert.Equal(t, 2, bullish, "Should count 2 bullish signals")
	assert.Equal(t, 3, bearish, "Should count 3 bearish signals")
	assert.Equal(t, 0, neutral, "Should count 0 neutral signals")
}

// TestTrendModel_GetSignalCounts_IgnoresLoadingRows verifies GetSignalCounts
// ignores rows that are still loading or in error state.
func TestTrendModel_GetSignalCounts_IgnoresLoadingRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")

	// Set only one row to loaded
	result := &trenddomain.Result{Symbol: "AAPL", Signal: trenddomain.Bullish}
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: result})
	model = asModel(t, updatedM)

	// Set one row to error
	err := errors.New("network error")
	updatedM, _ = model.Update(TrendErrorMsg{Symbol: "MSFT", Err: err})
	model = asModel(t, updatedM)

	// GOOGL remains in loading state

	bullish, bearish, neutral := model.GetSignalCounts()

	assert.Equal(t, 1, bullish, "Should count 1 bullish signal")
	assert.Equal(t, 0, bearish, "Should have 0 bearish signals")
	assert.Equal(t, 0, neutral, "Should have 0 neutral signals")
}

// TestTrendModel_GetSignalCounts_EmptyRows verifies GetSignalCounts returns
// all zeros when there are no loaded rows.
func TestTrendModel_GetSignalCounts_EmptyRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{})

	// No rows at all
	bullish, bearish, neutral := model.GetSignalCounts()

	assert.Equal(t, 0, bullish, "Should have 0 bullish signals")
	assert.Equal(t, 0, bearish, "Should have 0 bearish signals")
	assert.Equal(t, 0, neutral, "Should have 0 neutral signals")
}

// TestTrendModel_GetBlitzCounts_MixedSignals verifies GetBlitzCounts correctly
// counts LONG, SHORT, and HOLD BLITZ signals.
func TestTrendModel_GetBlitzCounts_MixedSignals(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL", "NVDA", "AMZN")

	// Set up 3 LONG, 2 SHORT
	blitzScores := []int{1, 1, 0, -1, -1}
	signals := []trenddomain.Signal{
		trenddomain.Bullish,
		trenddomain.Bullish,
		trenddomain.Bullish,
		trenddomain.Bearish,
		trenddomain.Bearish,
	}

	for i, score := range blitzScores {
		result := &trenddomain.Result{
			Symbol:     model.rows[i].Symbol,
			Signal:     signals[i],
			BlitzScore: score,
		}
		updatedM, _ := model.Update(TrendDataMsg{Symbol: model.rows[i].Symbol, Result: result})
		model = asModel(t, updatedM)
	}

	long, short, hold := model.GetBlitzCounts()

	assert.Equal(t, 2, long, "Should count 2 LONG signals")
	assert.Equal(t, 2, short, "Should count 2 SHORT signals")
	assert.Equal(t, 1, hold, "Should count 1 HOLD signal")
}

// TestTrendModel_GetBlitzCounts_IgnoresLoadingRows verifies GetBlitzCounts
// ignores rows that are still loading or in error state.
func TestTrendModel_GetBlitzCounts_IgnoresLoadingRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")

	// Set only one row to loaded
	result := &trenddomain.Result{Symbol: "AAPL", Signal: trenddomain.Bullish, BlitzScore: 1}
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: result})
	model = asModel(t, updatedM)

	// Set one row to error
	err := errors.New("network error")
	updatedM, _ = model.Update(TrendErrorMsg{Symbol: "MSFT", Err: err})
	model = asModel(t, updatedM)

	// GOOGL remains in loading state

	long, short, hold := model.GetBlitzCounts()

	assert.Equal(t, 1, long, "Should count 1 LONG signal")
	assert.Equal(t, 0, short, "Should have 0 SHORT signals")
	assert.Equal(t, 0, hold, "Should have 0 HOLD signals")
}

// TestTrendModel_GetBlitzCounts_EmptyRows verifies GetBlitzCounts returns
// all zeros when there are no loaded rows.
func TestTrendModel_GetBlitzCounts_EmptyRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{})

	// No rows at all
	long, short, hold := model.GetBlitzCounts()

	assert.Equal(t, 0, long, "Should have 0 LONG signals")
	assert.Equal(t, 0, short, "Should have 0 SHORT signals")
	assert.Equal(t, 0, hold, "Should have 0 HOLD signals")
}

// TestTrendModel_GetBlitzCounts_CachedRows verifies GetBlitzCounts
// includes cached rows in the count.
func TestTrendModel_GetBlitzCounts_CachedRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	// Set row to loaded
	result := &trenddomain.Result{Symbol: "AAPL", Signal: trenddomain.Bullish, BlitzScore: 1}
	updatedM, _ := model.Update(TrendDataMsg{Symbol: "AAPL", Result: result})
	model = asModel(t, updatedM)

	// Manually set to cached state
	model.rows[0].State = StateCached

	long, short, hold := model.GetBlitzCounts()

	assert.Equal(t, 1, long, "Should count cached row as LONG")
	assert.Equal(t, 0, short, "Should have 0 SHORT signals")
	assert.Equal(t, 0, hold, "Should have 0 HOLD signals")
}

// TestTrendModel_GetDestinyCounts_MixedSignals verifies GetDestinyCounts correctly
// counts LONG, SHORT, and HOLD DESTINY signals.
func TestTrendModel_GetDestinyCounts_MixedSignals(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL", "NVDA", "AMZN")

	destinyScores := []int{1, 1, -1, 0, -1}

	for i, score := range destinyScores {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:       model.rows[i].Symbol,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: score,
			Price:        100.0,
		}
	}

	long, short, hold := model.GetDestinyCounts()

	assert.Equal(t, 2, long, "Should count 2 LONG signals")
	assert.Equal(t, 2, short, "Should count 2 SHORT signals")
	assert.Equal(t, 1, hold, "Should count 1 HOLD signal")
}

// TestTrendModel_GetDestinyCounts_IgnoresLoadingRows verifies GetDestinyCounts
// ignores rows that are still loading or in error state.
func TestTrendModel_GetDestinyCounts_IgnoresLoadingRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")

	// Set only one row to loaded
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:       "AAPL",
		Signal:       trenddomain.Bullish,
		BlitzScore:   1,
		DestinyScore: 1,
		Price:        100.0,
	}

	// Set second row to error
	model.rows[1].State = StateError
	model.rows[1].Error = errors.New("network error")

	// Third row remains in loading state
	// rows[2] stays in StateLoading

	long, short, hold := model.GetDestinyCounts()

	assert.Equal(t, 1, long, "Should count 1 LONG signal (only loaded row)")
	assert.Equal(t, 0, short, "Should have 0 SHORT signals")
	assert.Equal(t, 0, hold, "Should have 0 HOLD signals")
}

// TestTrendModel_GetDestinyCounts_EmptyRows verifies GetDestinyCounts returns
// all zeros when there are no loaded rows.
func TestTrendModel_GetDestinyCounts_EmptyRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{})

	// No rows at all
	long, short, hold := model.GetDestinyCounts()

	assert.Equal(t, 0, long, "Should have 0 LONG signals when empty")
	assert.Equal(t, 0, short, "Should have 0 SHORT signals when empty")
	assert.Equal(t, 0, hold, "Should have 0 HOLD signals when empty")
}

// TestTrendModel_GetDestinyCounts_CachedRows verifies GetDestinyCounts
// includes cached rows in count.
func TestTrendModel_GetDestinyCounts_CachedRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	// Set row to cached state
	model.rows[0].State = StateCached
	model.rows[0].Result = &trenddomain.Result{
		Symbol:       "AAPL",
		Signal:       trenddomain.Bullish,
		BlitzScore:   1,
		DestinyScore: 1,
		Price:        100.0,
	}

	long, short, hold := model.GetDestinyCounts()

	assert.Equal(t, 1, long, "Should count cached row as LONG")
	assert.Equal(t, 0, short, "Should have 0 SHORT signals")
	assert.Equal(t, 0, hold, "Should have 0 HOLD signals")
}

// TestTrendModel_GetTPICounts verifies GetTPICounts correctly
// counts LONG and CASH TPI signals.
func TestTrendModel_GetTPICounts(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL", "NVDA", "AMZN")

	// Set up 3 LONG, 2 CASH
	for i, symbol := range []string{"AAPL", "MSFT", "GOOGL", "NVDA", "AMZN"} {
		tpiSignal := "CASH"
		tpiValue := -0.33
		if i < 3 {
			tpiSignal = "LONG"
			tpiValue = 0.67
		}
		updatedM, _ := model.Update(TrendDataMsg{
			Symbol: symbol,
			Result: &trenddomain.Result{
				Symbol:     symbol,
				Signal:     trenddomain.Bullish,
				BlitzScore: 1,
				TPI:        tpiValue,
				TPISignal:  tpiSignal,
			},
		})
		model = asModel(t, updatedM)
	}

	long, cash := model.GetTPICounts()

	assert.Equal(t, 3, long, "Should count 3 LONG TPI signals")
	assert.Equal(t, 2, cash, "Should count 2 CASH TPI signals")
}

// TestTrendModel_GetTPICounts_IgnoresLoadingRows verifies GetTPICounts
// ignores rows that are still loading or in error state.
func TestTrendModel_GetTPICounts_IgnoresLoadingRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL", "MSFT", "GOOGL")

	// Set only one row to loaded
	updatedM, _ := model.Update(TrendDataMsg{
		Symbol: "AAPL",
		Result: &trenddomain.Result{
			Symbol:     "AAPL",
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
			TPI:        0.67,
			TPISignal:  "LONG",
		},
	})
	model = asModel(t, updatedM)

	// Set one row to error
	updatedM, _ = model.Update(TrendErrorMsg{Symbol: "MSFT", Err: errors.New("network error")})
	model = asModel(t, updatedM)

	// GOOGL remains in loading state

	long, cash := model.GetTPICounts()

	assert.Equal(t, 1, long, "Should count 1 LONG signal")
	assert.Equal(t, 0, cash, "Should have 0 CASH signals")
}

// TestTrendModel_GetTPICounts_EmptyRows verifies GetTPICounts returns
// zeros when there are no loaded rows.
func TestTrendModel_GetTPICounts_EmptyRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{})

	// No rows at all
	long, cash := model.GetTPICounts()

	assert.Equal(t, 0, long, "Should have 0 LONG signals")
	assert.Equal(t, 0, cash, "Should have 0 CASH signals")
}

// TestTrendModel_GetTPICounts_CachedRows verifies GetTPICounts
// includes cached rows in count.
func TestTrendModel_GetTPICounts_CachedRows(t *testing.T) {
	model := newTestModel(t, &mockEngine{}, "AAPL")

	// Set row to loaded
	updatedM, _ := model.Update(TrendDataMsg{
		Symbol: "AAPL",
		Result: &trenddomain.Result{
			Symbol:     "AAPL",
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
			TPI:        0.67,
			TPISignal:  "LONG",
		},
	})
	model = asModel(t, updatedM)

	// Manually set to cached state
	model.rows[0].State = StateCached

	long, cash := model.GetTPICounts()

	assert.Equal(t, 1, long, "Should count cached row as LONG")
	assert.Equal(t, 0, cash, "Should have 0 CASH signals")
}
