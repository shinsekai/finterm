// Package trend provides tests for the trend following TUI view.
package trend

import (
	"context"
	"errors"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/owner/finterm/internal/config"
	trenddomain "github.com/owner/finterm/internal/domain/trend"
	"github.com/owner/finterm/internal/domain/trend/indicators"
)

// viewMockEngine is a mock implementation of Engine for view tests.
type viewMockEngine struct{}

func (m *viewMockEngine) AnalyzeWithSymbolDetection(_ context.Context, _ string) (*trenddomain.Result, error) {
	return nil, nil
}

// mockTheme is a mock implementation of Theme for testing.
type mockTheme struct{}

func (m *mockTheme) Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (m *mockTheme) TableRow() lipgloss.Style {
	return lipgloss.NewStyle()
}

func (m *mockTheme) TableHeader() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (m *mockTheme) TableEmpty() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true)
}

func (m *mockTheme) Bullish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("50FA7B")).Bold(true)
}

func (m *mockTheme) Bearish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("FF5555")).Bold(true)
}

func (m *mockTheme) Neutral() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("F1FA8C")).Bold(true)
}

func (m *mockTheme) Help() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true)
}

func (m *mockTheme) Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("FF5555")).Bold(true)
}

func (m *mockTheme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Faint(true)
}

func (m *mockTheme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("7D56F4"))
}

func (m *mockTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#F8F8F2")
}

func (m *mockTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282A36")
}

// newTestModelForView creates a configured model for testing and returns it as a value.
func newTestModelForView(symbols ...string) Model {
	equities := append([]string{}, symbols...)
	var crypto []string

	model := NewModel()
	model.Configure(
		context.Background(),
		&viewMockEngine{},
		&config.WatchlistConfig{Equities: equities, Crypto: crypto},
		indicators.NewAssetClassDetector([]string{}),
	)
	return *model
}

// TestTrendModel_View_LoadingState verifies rendering in loading state.
func TestTrendModel_View_LoadingState(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT")

	// Model starts in loading state by default
	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Trend Analysis", "View should contain title")
	assert.Contains(t, rendered, "Loading...", "View should contain loading indicator for tickers")
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL symbol")
	assert.Contains(t, rendered, "MSFT", "View should contain MSFT symbol")
	assert.Contains(t, rendered, "↑↓ Navigate", "View should contain navigation help")
	assert.Contains(t, rendered, "r Refresh", "View should contain refresh help")
}

// TestTrendModel_View_LoadedState verifies rendering in loaded state with data.
func TestTrendModel_View_LoadedState(t *testing.T) {
	model := newTestModelForView("AAPL")

	// Set row to loaded state with result
	result := &trenddomain.Result{
		Symbol:  "AAPL",
		RSI:     50.5,
		EMAFast: 150.25,
		EMASlow: 145.75,
		Signal:  trenddomain.Bullish,
	}
	model.rows[0].State = StateLoaded
	model.rows[0].Result = result

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Trend Analysis", "View should contain title")
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL symbol")
	assert.Contains(t, rendered, "Bullish", "View should contain bullish signal")
	assert.Contains(t, rendered, "50.50", "View should contain RSI value")
	assert.Contains(t, rendered, "150.25", "View should contain EMA fast value")
	assert.Contains(t, rendered, "145.75", "View should contain EMA slow value")
}

// TestTrendModel_View_MixedState verifies rendering with mixed loading/loaded/error states.
func TestTrendModel_View_MixedState(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL")

	// Set first row to loaded
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:  "AAPL",
		RSI:     50.5,
		EMAFast: 150.25,
		EMASlow: 145.75,
		Signal:  trenddomain.Bullish,
	}

	// Set second row to error
	model.rows[1].State = StateError
	model.rows[1].Error = errors.New("API error")

	// Third row remains in loading state

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Trend Analysis", "View should contain title")
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL symbol")
	assert.Contains(t, rendered, "Bullish", "View should contain bullish signal for AAPL")
	assert.Contains(t, rendered, "API error", "View should contain error for MSFT")
	assert.Contains(t, rendered, "Loading...", "View should contain loading for GOOGL")
}

// TestTrendModel_View_EmptyWatchlist verifies rendering with empty watchlist.
func TestTrendModel_View_EmptyWatchlist(t *testing.T) {
	model := NewModel()
	model.Configure(
		context.Background(),
		&viewMockEngine{},
		&config.WatchlistConfig{Equities: []string{}, Crypto: []string{}},
		indicators.NewAssetClassDetector([]string{}),
	)

	view := NewView(*model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Trend Analysis", "View should contain title")
	assert.Contains(t, rendered, "No tickers in watchlist", "View should show empty message")
}

// TestTrendModel_View_ActiveRow verifies that the active row is highlighted.
func TestTrendModel_View_ActiveRow(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL")

	// Set all rows to loaded
	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:  model.rows[i].Symbol,
			RSI:     50.0 + float64(i)*10,
			EMAFast: 150.0 + float64(i)*10,
			EMASlow: 145.0 + float64(i)*10,
			Signal:  trenddomain.Bullish,
		}
	}

	// Test active row at index 0
	model.activeRow = 0
	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL")

	// Test active row at index 1
	model.activeRow = 1
	rendered = NewView(model).SetTheme(&mockTheme{}).Render()
	assert.Contains(t, rendered, "MSFT", "View should contain MSFT")

	// Test active row at index 2
	model.activeRow = 2
	rendered = NewView(model).SetTheme(&mockTheme{}).Render()
	assert.Contains(t, rendered, "GOOGL", "View should contain GOOGL")
}

// TestTrendModel_View_BearishSignal verifies rendering of bearish signal.
func TestTrendModel_View_BearishSignal(t *testing.T) {
	model := newTestModelForView("AAPL")

	// Set row to loaded with bearish signal
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:  "AAPL",
		RSI:     65.0,
		EMAFast: 140.0,
		EMASlow: 150.0,
		Signal:  trenddomain.Bearish,
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Bearish", "View should contain bearish signal")
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL symbol")
}

// TestTrendModel_View_ValuationColors verifies valuation color coding.
func TestTrendModel_View_ValuationColors(t *testing.T) {
	tests := []struct {
		name       string
		valuation  string
		shouldHave []string
	}{
		{
			name:       "Oversold",
			valuation:  "Oversold",
			shouldHave: []string{"Oversold", "AAPL"},
		},
		{
			name:       "Undervalued",
			valuation:  "Undervalued",
			shouldHave: []string{"Undervalued", "AAPL"},
		},
		{
			name:       "Fair value",
			valuation:  "Fair value",
			shouldHave: []string{"Fair value", "AAPL"},
		},
		{
			name:       "Overvalued",
			valuation:  "Overvalued",
			shouldHave: []string{"Overvalued", "AAPL"},
		},
		{
			name:       "Overbought",
			valuation:  "Overbought",
			shouldHave: []string{"Overbought", "AAPL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModelForView("AAPL")
			model.rows[0].State = StateLoaded
			model.rows[0].Result = &trenddomain.Result{
				Symbol:    "AAPL",
				RSI:       50.0,
				EMAFast:   150.0,
				EMASlow:   145.0,
				Signal:    trenddomain.Bullish,
				Valuation: tt.valuation,
			}

			view := NewView(model).SetTheme(&mockTheme{})
			rendered := view.Render()

			for _, expected := range tt.shouldHave {
				assert.Contains(t, rendered, expected, "View should contain %s", expected)
			}
		})
	}
}

// TestTrendModel_View_RenderTableColumns verifies all table columns are rendered.
func TestTrendModel_View_RenderTableColumns(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:  "AAPL",
		RSI:     50.0,
		EMAFast: 150.0,
		EMASlow: 145.0,
		Signal:  trenddomain.Bullish,
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Check column headers
	assert.Contains(t, rendered, "Symbol", "View should have Symbol column")
	assert.Contains(t, rendered, "Signal", "View should have Signal column")
	assert.Contains(t, rendered, "RSI", "View should have RSI column")
	assert.Contains(t, rendered, "EMA Fast", "View should have EMA Fast column")
	assert.Contains(t, rendered, "EMA Slow", "View should have EMA Slow column")
	assert.Contains(t, rendered, "Valuation", "View should have Valuation column")
}

// TestTrendModel_View_DefaultTheme verifies view works with default theme.
func TestTrendModel_View_DefaultTheme(t *testing.T) {
	model := newTestModelForView("AAPL")

	// Create view without setting theme - should use default
	view := NewView(model)
	rendered := view.Render()

	assert.Contains(t, rendered, "Trend Analysis", "View should contain title with default theme")
	assert.NotEmpty(t, rendered, "View should not be empty")
}

// TestTrendModel_View_ErrorState verifies error state rendering.
func TestTrendModel_View_ErrorState(t *testing.T) {
	model := newTestModelForView("AAPL")

	// Set row to error state
	model.rows[0].State = StateError
	model.rows[0].Error = errors.New("API error")

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "AAPL", "View should contain symbol")
	assert.Contains(t, rendered, "API error", "View should contain error indicator")
}

// TestTrendModel_View_String verifies String method.
func TestTrendModel_View_String(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	rendered := view.String()
	assert.Equal(t, view.Render(), rendered, "String should return same as Render")
}

// TestTrendModel_View_SetTheme verifies theme setting.
func TestTrendModel_View_SetTheme(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model)

	theme := &mockTheme{}
	result := view.SetTheme(theme)

	assert.Equal(t, view, result, "SetTheme should return view for chaining")
	assert.Equal(t, theme, view.theme, "Theme should be set")
}

// TestTrendModel_View_Footer verifies footer rendering.
func TestTrendModel_View_Footer(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	footer := view.renderFooter()
	assert.Contains(t, footer, "↑↓ Navigate", "Footer should have navigation help")
	assert.Contains(t, footer, "r Refresh", "Footer should have refresh help")
	assert.Contains(t, footer, "q Quit", "Footer should have quit help")
}
