// Package trend provides tests for the trend following TUI view.
package trend

import (
	"context"
	"errors"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/shinsekai/finterm/internal/config"
	trenddomain "github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
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

func (m *mockTheme) TableRowAlt() lipgloss.Style {
	return lipgloss.NewStyle().Background(lipgloss.Color("#2D2F3D"))
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

func (m *mockTheme) BullishBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("50FA7B")).
		Bold(true)
}

func (m *mockTheme) BearishBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Background(lipgloss.Color("FF5555")).
		Bold(true)
}

func (m *mockTheme) NeutralBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("F1FA8C")).
		Bold(true)
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

func (m *mockTheme) Accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("7D56F4")).Bold(true)
}

func (m *mockTheme) Divider() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
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
	assert.Contains(t, rendered, "Loading…", "View should contain loading indicator for tickers")
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL symbol")
	assert.Contains(t, rendered, "MSFT", "View should contain MSFT symbol")
	assert.Contains(t, rendered, "navigate", "View should contain navigation help")
	assert.Contains(t, rendered, "refresh", "View should contain refresh help")
}

// TestTrendModel_View_LoadedState verifies rendering in loaded state with data.
func TestTrendModel_View_LoadedState(t *testing.T) {
	model := newTestModelForView("AAPL")

	// Set row to loaded state with result
	result := &trenddomain.Result{
		Symbol:  "AAPL",
		Price:   155.50,
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
	assert.Contains(t, rendered, "BULL", "View should contain bullish signal badge")
	assert.Contains(t, rendered, "50.50", "View should contain RSI value")
	assert.Contains(t, rendered, "150.25", "View should contain EMA fast value")
	assert.Contains(t, rendered, "145.75", "View should contain EMA slow value")
	assert.Contains(t, rendered, "$155.50", "View should contain dollar-prefixed price")
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
	assert.Contains(t, rendered, "BULL", "View should contain bullish signal for AAPL")
	assert.Contains(t, rendered, "API error", "View should contain error for MSFT")
	assert.Contains(t, rendered, "Loading…", "View should contain loading for GOOGL")
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
	assert.Contains(t, rendered, "▸ AAPL", "View should contain cursor marker before active symbol")
	assert.NotContains(t, rendered, "▸ MSFT", "View should not contain cursor before non-active MSFT")

	// Test active row at index 1
	model.activeRow = 1
	rendered = NewView(model).SetTheme(&mockTheme{}).Render()
	assert.Contains(t, rendered, "▸ MSFT", "View should contain cursor marker before active MSFT")
	assert.NotContains(t, rendered, "▸ AAPL", "View should not contain cursor before non-active AAPL")

	// Test active row at index 2
	model.activeRow = 2
	rendered = NewView(model).SetTheme(&mockTheme{}).Render()
	assert.Contains(t, rendered, "▸ GOOGL", "View should contain cursor marker before active GOOGL")
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

	assert.Contains(t, rendered, "BEAR", "View should contain bearish signal badge")
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
	assert.Contains(t, rendered, "SYMBOL", "View should have SYMBOL column")
	assert.Contains(t, rendered, "SIGNAL", "View should have SIGNAL column")
	assert.Contains(t, rendered, "RSI", "View should have RSI column")
	assert.Contains(t, rendered, "EMA FAST", "View should have EMA FAST column")
	assert.Contains(t, rendered, "EMA SLOW", "View should have EMA SLOW column")
	assert.Contains(t, rendered, "VALUATION", "View should have VALUATION column")
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
	assert.Contains(t, rendered, "✗ API error", "View should contain error indicator")
	assert.Contains(t, rendered, "—", "View should contain em dash for missing values")
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
	assert.Contains(t, footer, "navigate", "Footer should have navigation help")
	assert.Contains(t, footer, "refresh", "Footer should have refresh help")
	assert.Contains(t, footer, "tabs", "Footer should have tabs help")
	assert.Contains(t, footer, "help", "Footer should have help key")
	assert.Contains(t, footer, "·", "Footer should have separator")
}

// TestTrendView_SignalBadge_Bullish verifies bullish signal badge rendering.
func TestTrendView_SignalBadge_Bullish(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderSignalBadge(trenddomain.Bullish)
	assert.Contains(t, badge, "▲ BULL", "Bullish badge should show arrow and BULL")
}

// TestTrendView_SignalBadge_Bearish verifies bearish signal badge rendering.
func TestTrendView_SignalBadge_Bearish(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderSignalBadge(trenddomain.Bearish)
	assert.Contains(t, badge, "▼ BEAR", "Bearish badge should show arrow and BEAR")
}

// TestTrendView_SignalBadge_Neutral verifies neutral signal badge rendering.
func TestTrendView_SignalBadge_Neutral(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	// Use Bullish as fallback since Neutral doesn't exist in domain
	// The renderSignalBadge handles unknown signals as neutral
	badge := view.renderSignalBadge(trenddomain.Bullish + 10) // Use an invalid signal index
	assert.Contains(t, badge, "─ HOLD", "Neutral badge should show dash and HOLD")
}

// TestTrendView_RSIColorCoding verifies RSI value color coding.
func TestTrendView_RSIColorCoding(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	tests := []struct {
		name          string
		rsi           float64
		shouldHave    string
		shouldNotHave string
	}{
		{"RSI < 30 (oversold) - green", 25.0, "25.00", ""},
		{"RSI > 70 (overbought) - red", 75.0, "75.00", ""},
		{"RSI 30-45 (undervalued) - green", 40.0, "40.00", ""},
		{"RSI 55-70 (overvalued) - red", 60.0, "60.00", ""},
		{"RSI 45-55 (fair) - default", 50.0, "50.00", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsiRendered := view.renderRSIValue(tt.rsi)
			assert.Contains(t, rsiRendered, tt.shouldHave, "RSI should contain value")
		})
	}
}

// TestTrendView_ValuationBadge verifies valuation badge rendering.
func TestTrendView_ValuationBadge(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	tests := []struct {
		name      string
		valuation string
		expected  string
	}{
		{"Oversold", "Oversold", "◆ Oversold"},
		{"Undervalued", "Undervalued", "◇ Undervalued"},
		{"Overbought", "Overbought", "◆ Overbought"},
		{"Overvalued", "Overvalued", "◇ Overvalued"},
		{"Fair value", "Fair value", "○ Fair value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := view.renderValuationBadge(tt.valuation)
			assert.Contains(t, badge, tt.expected, "Valuation badge should show icon and text")
		})
	}
}

// TestTrendView_AlternatingRows verifies that odd rows use TableRowAlt style.
func TestTrendView_AlternatingRows(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL")

	// Set all rows to loaded
	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:  model.rows[i].Symbol,
			RSI:     50.0,
			EMAFast: 150.0,
			EMASlow: 145.0,
			Signal:  trenddomain.Bullish,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Just verify rendering doesn't crash
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL")
	assert.Contains(t, rendered, "MSFT", "View should contain MSFT")
	assert.Contains(t, rendered, "GOOGL", "View should contain GOOGL")
}

// TestTrendView_SymbolAccentStyle verifies symbols are rendered with accent style.
func TestTrendView_SymbolAccentStyle(t *testing.T) {
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

	assert.Contains(t, rendered, "AAPL", "View should contain symbol with accent style")
}

// TestTrendView_TitleWithIcon verifies title has diamond icon.
func TestTrendView_TitleWithIcon(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	title := view.renderTitle()
	assert.Contains(t, title, "◆", "Title should contain diamond icon")
	assert.Contains(t, title, "Trend Analysis", "Title should contain text")
}

// TestTrendView_SummaryBar_Counts verifies signal summary bar shows correct counts.
func TestTrendView_SummaryBar_Counts(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL", "NVDA", "AMZN")

	// Set up 3 bullish, 2 bearish (no neutral signal exists in domain)
	signals := []trenddomain.Signal{
		trenddomain.Bullish,
		trenddomain.Bullish,
		trenddomain.Bullish,
		trenddomain.Bearish,
		trenddomain.Bearish,
	}

	for i, signal := range signals {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:  model.rows[i].Symbol,
			Signal:  signal,
			Price:   100.0,
			RSI:     50.0,
			EMAFast: 100.0,
			EMASlow: 90.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	summary := view.renderSummary()

	assert.Contains(t, summary, "3 ▲", "Summary should show 3 bullish signals")
	assert.Contains(t, summary, "2 ▼", "Summary should show 2 bearish signals")
}

// TestTrendView_SummaryBar_WithPending verifies pending count is shown during loading.
func TestTrendView_SummaryBar_WithPending(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "BTC")

	// Set 2 loaded, 1 loading
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:  "AAPL",
		Signal:  trenddomain.Bullish,
		Price:   100.0,
		RSI:     50.0,
		EMAFast: 100.0,
		EMASlow: 90.0,
	}
	model.rows[1].State = StateLoaded
	model.rows[1].Result = &trenddomain.Result{
		Symbol:  "MSFT",
		Signal:  trenddomain.Bearish,
		Price:   100.0,
		RSI:     50.0,
		EMAFast: 100.0,
		EMASlow: 90.0,
	}
	// rows[2] stays in loading state

	view := NewView(model).SetTheme(&mockTheme{})
	summary := view.renderSummary()

	assert.Contains(t, summary, "1 ▲", "Summary should show 1 bullish signal")
	assert.Contains(t, summary, "1 ▼", "Summary should show 1 bearish signal")
	assert.Contains(t, summary, "1 pending", "Summary should show 1 pending")
}

// TestTrendView_SummaryBar_HiddenWhenEmpty verifies summary bar is hidden when watchlist is empty.
func TestTrendView_SummaryBar_HiddenWhenEmpty(t *testing.T) {
	model := NewModel()
	model.Configure(
		context.Background(),
		&viewMockEngine{},
		&config.WatchlistConfig{Equities: []string{}, Crypto: []string{}},
		indicators.NewAssetClassDetector([]string{}),
	)

	view := NewView(*model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Summary bar should not be rendered (no row data)
	// Verify we still get a valid render, just without summary
	assert.Contains(t, rendered, "Trend Analysis", "View should contain title")
	assert.NotContains(t, rendered, "▲", "View should not contain signal markers when empty")
}

// TestTrendView_ActiveRowCursor verifies cursor marker is only on active row.
func TestTrendView_ActiveRowCursor(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL")

	// Set all rows to loaded
	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:  model.rows[i].Symbol,
			Signal:  trenddomain.Bullish,
			Price:   100.0,
			RSI:     50.0,
			EMAFast: 100.0,
			EMASlow: 90.0,
		}
	}

	model.activeRow = 1
	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Active row should have cursor marker
	assert.Contains(t, rendered, "▸ MSFT", "Active row should have cursor marker")
	// Non-active rows should have spacer
	assert.Contains(t, rendered, "  AAPL", "Non-active row should have spacer")
	assert.Contains(t, rendered, "  GOOGL", "Non-active row should have spacer")
}

// TestTrendView_PriceDollarPrefix verifies prices show dollar sign prefix.
func TestTrendView_PriceDollarPrefix(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		expected string
	}{
		{"Normal price", 155.50, "$155.50"},
		{"Small price (crypto)", 0.001234, "$0.001234"},
		{"Zero price", 0.0, "—"},
		{"Large price", 1234.56, "$1,234.56"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModelForView("TEST")
			model.rows[0].State = StateLoaded
			model.rows[0].Result = &trenddomain.Result{
				Symbol:  "TEST",
				Price:   tt.price,
				RSI:     50.0,
				EMAFast: 100.0,
				EMASlow: 90.0,
				Signal:  trenddomain.Bullish,
			}

			view := NewView(model).SetTheme(&mockTheme{})
			rendered := view.Render()
			assert.Contains(t, rendered, tt.expected, "View should contain dollar-prefixed price")
		})
	}
}

// TestTrendView_LoadingProgress verifies loading progress is shown in subtitle.
func TestTrendView_LoadingProgress(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL", "NVDA")

	// Load first 2, keep 2 in loading state
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:  "AAPL",
		Signal:  trenddomain.Bullish,
		Price:   100.0,
		RSI:     50.0,
		EMAFast: 100.0,
		EMASlow: 90.0,
	}
	model.rows[1].State = StateLoaded
	model.rows[1].Result = &trenddomain.Result{
		Symbol:  "MSFT",
		Signal:  trenddomain.Bullish,
		Price:   100.0,
		RSI:     50.0,
		EMAFast: 100.0,
		EMASlow: 90.0,
	}

	view := NewView(model).SetTheme(&mockTheme{})
	title := view.renderTitle()

	assert.Contains(t, title, "Loading 2/4…", "Title should show loading progress")
}

// TestTrendView_LoadingProgress_AllLoaded verifies normal subtitle when all loaded.
func TestTrendView_LoadingProgress_AllLoaded(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT")

	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:  model.rows[i].Symbol,
			Signal:  trenddomain.Bullish,
			Price:   100.0,
			RSI:     50.0,
			EMAFast: 100.0,
			EMASlow: 90.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	title := view.renderTitle()

	assert.Contains(t, title, "2 symbols", "Title should show symbol count when all loaded")
	assert.NotContains(t, title, "Loading", "Title should not show loading when all loaded")
}

// TestTrendView_SectionSeparator_MixedWatchlist verifies separator appears between equities and crypto.
func TestTrendView_SectionSeparator_MixedWatchlist(t *testing.T) {
	model := NewModel()
	model.Configure(
		context.Background(),
		&viewMockEngine{},
		&config.WatchlistConfig{Equities: []string{"AAPL", "MSFT"}, Crypto: []string{"BTC", "ETH"}},
		indicators.NewAssetClassDetector([]string{}),
	)

	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:  model.rows[i].Symbol,
			Signal:  trenddomain.Bullish,
			Price:   100.0,
			RSI:     50.0,
			EMAFast: 100.0,
			EMASlow: 90.0,
		}
	}

	view := NewView(*model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Crypto", "View should contain crypto separator")
}

// TestTrendView_SectionSeparator_EquitiesOnly verifies no separator when only equities.
func TestTrendView_SectionSeparator_EquitiesOnly(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL")

	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:  model.rows[i].Symbol,
			Signal:  trenddomain.Bullish,
			Price:   100.0,
			RSI:     50.0,
			EMAFast: 100.0,
			EMASlow: 90.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.NotContains(t, rendered, "Crypto", "View should not contain crypto separator when only equities")
}

// TestTrendView_SectionSeparator_CryptoOnly verifies no separator when only crypto.
func TestTrendView_SectionSeparator_CryptoOnly(t *testing.T) {
	model := NewModel()
	model.Configure(
		context.Background(),
		&viewMockEngine{},
		&config.WatchlistConfig{Equities: []string{}, Crypto: []string{"BTC", "ETH"}},
		indicators.NewAssetClassDetector([]string{}),
	)

	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:  model.rows[i].Symbol,
			Signal:  trenddomain.Bullish,
			Price:   0.001234,
			RSI:     50.0,
			EMAFast: 100.0,
			EMASlow: 90.0,
		}
	}

	view := NewView(*model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.NotContains(t, rendered, "Crypto", "View should not contain crypto separator when only crypto")
}

// TestTrendView_TableContainer verifies table is wrapped in bordered container.
func TestTrendView_TableContainer(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:  "AAPL",
		Signal:  trenddomain.Bullish,
		Price:   100.0,
		RSI:     50.0,
		EMAFast: 100.0,
		EMASlow: 90.0,
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Check for rounded border characters from lipgloss.RoundedBorder
	assert.Contains(t, rendered, "╭", "View should contain rounded border corner")
}
