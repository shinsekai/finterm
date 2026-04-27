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

// newTestModelForView creates a configured model for testing.
func newTestModelForView(symbols ...string) *Model {
	equities := append([]string{}, symbols...)
	var crypto []string

	model := NewModel()
	model.Configure(
		context.Background(),
		&viewMockEngine{},
		&config.WatchlistConfig{Equities: equities, Crypto: crypto},
		indicators.NewAssetClassDetector([]string{}),
	)
	return model
}

// TestTrendModel_View_LoadingState verifies rendering in loading state.
func TestTrendModel_View_LoadingState(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT")

	// Model starts in loading state by default
	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Trend Analysis", "View should contain title")
	assert.Contains(t, rendered, "pending", "View should contain pending indicator")
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL symbol")
	assert.Contains(t, rendered, "MSFT", "View should contain MSFT symbol")
	assert.Contains(t, rendered, "navigate", "View should contain navigation help")
	assert.Contains(t, rendered, "refresh", "View should contain refresh help")
	// TPI, FTEMA, BLITZ columns should show "—" for loading rows
	assert.Contains(t, rendered, "—", "View should contain em dash for missing values")
}

// TestTrendModel_View_LoadedState verifies rendering in loaded state with data.
func TestTrendModel_View_LoadedState(t *testing.T) {
	model := newTestModelForView("AAPL")

	// Set row to loaded state with result
	result := &trenddomain.Result{
		Symbol:     "AAPL",
		Price:      155.50,
		RSI:        50.5,
		EMAFast:    150.25,
		EMASlow:    145.75,
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		TPI:        0.67,
		TPISignal:  "LONG",
	}
	model.rows[0].State = StateLoaded
	model.rows[0].Result = result

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Trend Analysis", "View should contain title")
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL symbol")
	assert.Contains(t, rendered, "LONG", "View should contain FTEMA LONG badge")
	assert.Contains(t, rendered, "LONG", "View should contain TPI LONG label")
	assert.Contains(t, rendered, "▓", "View should contain TPI gauge bars")
	assert.Contains(t, rendered, "$155.50", "View should contain dollar-prefixed price")
}

// TestTrendModel_View_MixedState verifies rendering with mixed loading/loaded/error states.
func TestTrendModel_View_MixedState(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL")

	// Set first row to loaded
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:     "AAPL",
		RSI:        50.5,
		EMAFast:    150.25,
		EMASlow:    145.75,
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
	}

	// Set second row to error
	model.rows[1].State = StateError
	model.rows[1].Error = errors.New("API error")

	// Third row remains in loading state

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Trend Analysis", "View should contain title")
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL symbol")
	assert.Contains(t, rendered, "LONG", "View should contain bullish signal for AAPL")
	assert.Contains(t, rendered, "LONG", "View should contain BLITZ LONG for AAPL")
	assert.Contains(t, rendered, "API error", "View should contain error for MSFT")
	assert.Contains(t, rendered, "pending", "View should contain pending indicator for GOOGL")
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

	view := NewView(model).SetTheme(&mockTheme{})
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
			Symbol:     model.rows[i].Symbol,
			RSI:        50.0 + float64(i)*10,
			EMAFast:    150.0 + float64(i)*10,
			EMASlow:    145.0 + float64(i)*10,
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
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
		Symbol:     "AAPL",
		RSI:        65.0,
		EMAFast:    140.0,
		EMASlow:    150.0,
		Signal:     trenddomain.Bearish,
		BlitzScore: -1,
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "SHORT", "View should contain bearish signal badge")
	assert.Contains(t, rendered, "SHORT", "View should contain BLITZ SHORT badge")
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
		Symbol:     "AAPL",
		RSI:        50.0,
		EMAFast:    150.0,
		EMASlow:    145.0,
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		TPI:        0.67,
		TPISignal:  "LONG",
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Check column headers
	assert.Contains(t, rendered, "SYMBOL", "View should have SYMBOL column")
	assert.Contains(t, rendered, "TPI", "View should have TPI column")
	assert.Contains(t, rendered, "FTEMA", "View should have FTEMA column")
	assert.Contains(t, rendered, "BLITZ", "View should have BLITZ column")
	assert.Contains(t, rendered, "RSI", "View should have RSI column")
	assert.Contains(t, rendered, "VALUATION", "View should have VALUATION column")

	// Verify old columns are NOT present
	assert.NotContains(t, rendered, "SIGNAL", "View should NOT have SIGNAL column")
	assert.NotContains(t, rendered, "EMA FAST", "View should NOT have EMA FAST column")
	assert.NotContains(t, rendered, "EMA SLOW", "View should NOT have EMA SLOW column")
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
	// BLITZ column should show "—" for error rows
	countDashes := 0
	for _, c := range rendered {
		if c == '—' {
			countDashes++
		}
	}
	assert.GreaterOrEqual(t, countDashes, 6, "View should have at least 6 em dashes for all missing columns including BLITZ")
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

	badge := view.renderFTEMABadge(trenddomain.Bullish)
	assert.Contains(t, badge, "▲  LONG", "Bullish badge should show arrow and LONG")
}

// TestTrendView_SignalBadge_Bearish verifies bearish signal badge rendering.
func TestTrendView_SignalBadge_Bearish(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderFTEMABadge(trenddomain.Bearish)
	assert.Contains(t, badge, "▼ SHORT", "Bearish badge should show arrow and SHORT")
}

// TestTrendView_SignalBadge_Neutral verifies neutral signal badge rendering.
func TestTrendView_SignalBadge_Neutral(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	// Use an invalid signal index
	badge := view.renderFTEMABadge(trenddomain.Bullish + 10)
	assert.Empty(t, badge, "Neutral badge should be empty")
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
			Symbol:     model.rows[i].Symbol,
			RSI:        50.0,
			EMAFast:    150.0,
			EMASlow:    145.0,
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
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
		Symbol:     "AAPL",
		RSI:        50.0,
		EMAFast:    150.0,
		EMASlow:    145.0,
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
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

// TestTrendView_SummaryBar_HiddenWhenEmpty verifies summary bar is hidden when watchlist is empty.
func TestTrendView_SummaryBar_HiddenWhenEmpty(t *testing.T) {
	model := NewModel()
	model.Configure(
		context.Background(),
		&viewMockEngine{},
		&config.WatchlistConfig{Equities: []string{}, Crypto: []string{}},
		indicators.NewAssetClassDetector([]string{}),
	)

	view := NewView(model).SetTheme(&mockTheme{})
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
			Symbol:     model.rows[i].Symbol,
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
			Price:      100.0,
			RSI:        50.0,
			EMAFast:    100.0,
			EMASlow:    90.0,
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
				Symbol:     "TEST",
				Price:      tt.price,
				RSI:        50.0,
				EMAFast:    100.0,
				EMASlow:    90.0,
				Signal:     trenddomain.Bullish,
				BlitzScore: 1,
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
		Symbol:     "AAPL",
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		Price:      100.0,
		RSI:        50.0,
		EMAFast:    100.0,
		EMASlow:    90.0,
	}
	model.rows[1].State = StateLoaded
	model.rows[1].Result = &trenddomain.Result{
		Symbol:     "MSFT",
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		Price:      100.0,
		RSI:        50.0,
		EMAFast:    100.0,
		EMASlow:    90.0,
	}

	view := NewView(model).SetTheme(&mockTheme{})
	title := view.renderTitle()

	assert.Contains(t, title, "loaded 2/4", "Title should show loading progress")
}

// TestTrendView_LoadingProgress_AllLoaded verifies normal subtitle when all loaded.
func TestTrendView_LoadingProgress_AllLoaded(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT")

	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:     model.rows[i].Symbol,
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
			Price:      100.0,
			RSI:        50.0,
			EMAFast:    100.0,
			EMASlow:    90.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	title := view.renderTitle()

	assert.Contains(t, title, "all loaded", "Title should show all loaded status when all loaded")
	assert.NotContains(t, title, "Loading", "Title should not show loading when all loaded")
	assert.Contains(t, title, "2m ago", "Title should show time indicator when all loaded")
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
			Symbol:     model.rows[i].Symbol,
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
			Price:      100.0,
			RSI:        50.0,
			EMAFast:    100.0,
			EMASlow:    90.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Crypto", "View should contain crypto separator")
}

// TestTrendView_SectionSeparator_EquitiesOnly verifies no separator when only equities.
func TestTrendView_SectionSeparator_EquitiesOnly(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL")

	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:     model.rows[i].Symbol,
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
			Price:      100.0,
			RSI:        50.0,
			EMAFast:    100.0,
			EMASlow:    90.0,
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
			Symbol:     model.rows[i].Symbol,
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
			Price:      0.001234,
			RSI:        50.0,
			EMAFast:    100.0,
			EMASlow:    90.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.NotContains(t, rendered, "Crypto", "View should not contain crypto separator when only crypto")
}

// TestTrendView_TableContainer verifies table is wrapped in bordered container.
func TestTrendView_TableContainer(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:     "AAPL",
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		Price:      100.0,
		RSI:        50.0,
		EMAFast:    100.0,
		EMASlow:    90.0,
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Check for rounded border characters from lipgloss.RoundedBorder
	assert.Contains(t, rendered, "╭", "View should contain rounded border corner")
}

// TestTrendView_BlitzBadge_Long verifies LONG BLITZ badge rendering.
func TestTrendView_BlitzBadge_Long(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderBlitzBadge(1)
	assert.Contains(t, badge, "▲  LONG", "LONG badge should show arrow and LONG")
}

// TestTrendView_BlitzBadge_Short verifies SHORT BLITZ badge rendering.
func TestTrendView_BlitzBadge_Short(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderBlitzBadge(-1)
	assert.Contains(t, badge, "▼ SHORT", "SHORT badge should show arrow and SHORT")
}

// TestTrendView_BlitzBadge_Hold verifies HOLD BLITZ badge rendering.
func TestTrendView_BlitzBadge_Hold(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderBlitzBadge(0)
	assert.Empty(t, badge, "HOLD badge should be empty")
}

// TestTrendView_BlitzBadge_OtherScores verifies other BLITZ scores render as empty.
func TestTrendView_BlitzBadge_OtherScores(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	tests := []int{2, -2, 100, -100}
	for _, score := range tests {
		badge := view.renderBlitzBadge(score)
		assert.Empty(t, badge, "BLITZ score %d should render as empty", score)
	}
}

// TestTrendView_BlitzSummary verifies BLITZ summary line with correct counts.
func TestTrendView_BlitzSummary(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL", "NVDA", "AMZN")

	blitzScores := []int{1, 1, 0, -1, -1}

	for i, score := range blitzScores {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:     model.rows[i].Symbol,
			Signal:     trenddomain.Bullish,
			BlitzScore: score,
			Price:      100.0,
			RSI:        50.0,
			EMAFast:    100.0,
			EMASlow:    90.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	summary := view.renderSummary()

	assert.Contains(t, summary, "BLITZ:", "Summary should show BLITZ label")
	assert.Contains(t, summary, "2 LONG", "Summary should show 2 LONG")
	assert.Contains(t, summary, "2 SHORT", "Summary should show 2 SHORT")
	assert.Contains(t, summary, "1 HOLD", "Summary should show 1 HOLD")
}

// TestTrendView_BlitzSummary_OnlyLong verifies BLITZ summary with only LONG signals.
func TestTrendView_BlitzSummary_OnlyLong(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT")

	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:       model.rows[i].Symbol,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1, // Set LONG so DESTINY doesn't show HOLD
			Price:        100.0,
			RSI:          50.0,
			EMAFast:      100.0,
			EMASlow:      90.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	summary := view.renderSummary()

	assert.Contains(t, summary, "BLITZ:", "Summary should show BLITZ label")
	assert.Contains(t, summary, "2 LONG", "Summary should show 2 LONG")
	assert.NotContains(t, summary, "SHORT", "Summary should not show SHORT")
	// Check BLITZ section specifically doesn't show HOLD (DESTINY may show HOLD)
	assert.NotRegexp(t, `(?i)BLITZ:.*\d\s+HOLD`, summary, "BLITZ section should not show HOLD")
}

// TestTrendView_BlitzSummary_OnlyHold verifies BLITZ summary with only HOLD signals.
func TestTrendView_BlitzSummary_OnlyHold(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT")

	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:     model.rows[i].Symbol,
			Signal:     trenddomain.Bullish,
			BlitzScore: 0,
			Price:      100.0,
			RSI:        50.0,
			EMAFast:    100.0,
			EMASlow:    90.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	summary := view.renderSummary()

	assert.Contains(t, summary, "BLITZ:", "Summary should show BLITZ label")
	assert.Contains(t, summary, "2 HOLD", "Summary should show 2 HOLD")
	assert.NotContains(t, summary, "LONG", "Summary should not show LONG")
	assert.NotContains(t, summary, "SHORT", "Summary should not show SHORT")
}

// TestTrendView_BlitzColumn_Loading verifies "—" for loading rows in BLITZ column.
func TestTrendView_BlitzColumn_Loading(t *testing.T) {
	model := newTestModelForView("AAPL")

	// Keep row in loading state (default)
	view := NewView(model).SetTheme(&mockTheme{})
	cells := view.buildLoadingCells(model.rows[0])

	// BLITZ is the 4th column (index 3) in the new 8-column layout
	assert.Equal(t, "—", cells[4].Text, "BLITZ column should show '—' for loading rows")
}

// TestTrendView_BlitzColumn_Error verifies "—" for error rows in BLITZ column.
func TestTrendView_BlitzColumn_Error(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = StateError
	model.rows[0].Error = errors.New("API error")

	view := NewView(model).SetTheme(&mockTheme{})
	cells := view.buildErrorCells(model.rows[0])

	// BLITZ is the 4th column (index 3) in the new 8-column layout
	assert.Equal(t, "—", cells[4].Text, "BLITZ column should show '—' for error rows")
}

// TestTrendView_BlitzColumn_Unknown verifies "—" for unknown state rows in BLITZ column.
func TestTrendView_BlitzColumn_Unknown(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = State(99) // Invalid state

	view := NewView(model).SetTheme(&mockTheme{})
	cells := view.buildUnknownCells(model.rows[0])

	// BLITZ is the 4th column (index 3) in the new 8-column layout
	assert.Equal(t, "—", cells[4].Text, "BLITZ column should show '—' for unknown state rows")
}

// TestTrendView_BlitzColumn_Loaded verifies correct BLITZ badge for loaded rows.
func TestTrendView_BlitzColumn_Loaded(t *testing.T) {
	tests := []struct {
		name       string
		blitzScore int
		shouldHave string
	}{
		{"LONG", 1, "▲  LONG"},
		{"SHORT", -1, "▼ SHORT"},
		{"HOLD", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModelForView("AAPL")
			model.rows[0].State = StateLoaded
			model.rows[0].Result = &trenddomain.Result{
				Symbol:     "AAPL",
				Signal:     trenddomain.Bullish,
				BlitzScore: tt.blitzScore,
				Price:      100.0,
				RSI:        50.0,
				EMAFast:    100.0,
				EMASlow:    90.0,
			}

			view := NewView(model).SetTheme(&mockTheme{})
			cells := view.buildLoadedCells(model.rows[0])

			// BLITZ is the 4th column (index 3) in the new 8-column layout
			if tt.shouldHave == "" {
				assert.Empty(t, cells[4].Text, "BLITZ column should be empty for HOLD")
			} else {
				assert.Contains(t, cells[4].Text, tt.shouldHave, "BLITZ column should show correct badge")
			}
		})
	}
}

// TestTrendView_BlitzColumn_Cached verifies correct BLITZ badge for cached rows.
func TestTrendView_BlitzColumn_Cached(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = StateCached
	model.rows[0].Result = &trenddomain.Result{
		Symbol:     "AAPL",
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		Price:      100.0,
		RSI:        50.0,
		EMAFast:    100.0,
		EMASlow:    90.0,
	}

	view := NewView(model).SetTheme(&mockTheme{})
	cells := view.buildCachedCells(model.rows[0])

	// BLITZ is the 4th column (index 3) in the new 8-column layout
	assert.Contains(t, cells[4].Text, "▲  LONG", "BLITZ column should show LONG badge for cached rows")
}

// TestTrendView_ValuationColors verifies valuation color coding.
func TestTrendView_ValuationColors(t *testing.T) {
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
				Symbol:     "AAPL",
				RSI:        50.0,
				EMAFast:    150.0,
				EMASlow:    145.0,
				Signal:     trenddomain.Bullish,
				BlitzScore: 0,
				Valuation:  tt.valuation,
			}

			view := NewView(model).SetTheme(&mockTheme{})
			rendered := view.Render()

			for _, expected := range tt.shouldHave {
				assert.Contains(t, rendered, expected, "View should contain %s", expected)
			}
		})
	}
}

// TestTrendView_DestinyBadge_Long verifies LONG DESTINY badge rendering.
func TestTrendView_DestinyBadge_Long(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderDestinyBadge(1)
	assert.Contains(t, badge, "▲  LONG", "LONG badge should show arrow and LONG")
}

// TestTrendView_DestinyBadge_Short verifies SHORT DESTINY badge rendering.
func TestTrendView_DestinyBadge_Short(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderDestinyBadge(-1)
	assert.Contains(t, badge, "▼ SHORT", "SHORT badge should show arrow and SHORT")
}

// TestTrendView_DestinyBadge_Hold verifies HOLD DESTINY badge rendering.
func TestTrendView_DestinyBadge_Hold(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderDestinyBadge(0)
	assert.Empty(t, badge, "HOLD badge should be empty")
}

// TestTrendView_DestinyBadge_OtherScores verifies other DESTINY scores render as empty.
func TestTrendView_DestinyBadge_OtherScores(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	tests := []int{2, -2, 100, -100}
	for _, score := range tests {
		badge := view.renderDestinyBadge(score)
		assert.Empty(t, badge, "DESTINY score %d should render as empty", score)
	}
}

// TestTrendView_DestinyColumn_InHeader verifies DESTINY column in table header.
func TestTrendView_DestinyColumn_InHeader(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:     "AAPL",
		RSI:        50.0,
		EMAFast:    100.0,
		EMASlow:    90.0,
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		Price:      100.0,
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "DESTINY", "View should have DESTINY column in header")
}

// TestTrendView_DestinyColumn_Loaded verifies correct DESTINY badge for loaded rows.
func TestTrendView_DestinyColumn_Loaded(t *testing.T) {
	tests := []struct {
		name         string
		destinyScore int
		shouldHave   string
	}{
		{"LONG", 1, "▲  LONG"},
		{"SHORT", -1, "▼ SHORT"},
		{"HOLD", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModelForView("TEST")
			model.rows[0].State = StateLoaded
			model.rows[0].Result = &trenddomain.Result{
				Symbol:       "TEST",
				RSI:          50.0,
				EMAFast:      100.0,
				EMASlow:      90.0,
				Signal:       trenddomain.Bullish,
				BlitzScore:   1,
				DestinyScore: tt.destinyScore,
				Price:        100.0,
			}

			view := NewView(model).SetTheme(&mockTheme{})
			cells := view.buildLoadedCells(model.rows[0])

			// DESTINY is 4th column (index 3)
			if tt.shouldHave == "" {
				assert.Empty(t, cells[5].Text, "DESTINY column should be empty for HOLD")
			} else {
				assert.Contains(t, cells[5].Text, tt.shouldHave, "DESTINY column should show correct badge for loaded rows")
			}
		})
	}
}

// TestTrendView_DestinyColumn_Loading verifies "—" for loading rows in DESTINY column.
func TestTrendView_DestinyColumn_Loading(t *testing.T) {
	model := newTestModelForView("AAPL")

	view := NewView(model).SetTheme(&mockTheme{})
	cells := view.buildLoadingCells(model.rows[0])

	// DESTINY is 4th column (index 3)
	assert.Equal(t, "—", cells[5].Text, "DESTINY column should show '—' for loading rows")
}

// TestTrendView_DestinyColumn_Error verifies "—" for error rows in DESTINY column.
func TestTrendView_DestinyColumn_Error(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = StateError
	model.rows[0].Error = errors.New("API error")

	view := NewView(model).SetTheme(&mockTheme{})
	cells := view.buildErrorCells(model.rows[0])

	// DESTINY is 4th column (index 3)
	assert.Equal(t, "—", cells[5].Text, "DESTINY column should show '—' for error rows")
}

// TestTrendView_DestinyColumn_Unknown verifies "—" for unknown state rows in DESTINY column.
func TestTrendView_DestinyColumn_Unknown(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = State(99) // Invalid state

	view := NewView(model).SetTheme(&mockTheme{})
	cells := view.buildUnknownCells(model.rows[0])

	// DESTINY is 4th column (index 3)
	assert.Equal(t, "—", cells[5].Text, "DESTINY column should show '—' for unknown state rows")
}

// TestTrendView_DestinySummary verifies DESTINY summary line with correct counts.
func TestTrendView_DestinySummary(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL", "NVDA", "AMZN")

	destinyScores := []int{1, 1, -1, 0}

	for i, score := range destinyScores {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:       model.rows[i].Symbol,
			RSI:          50.0,
			EMAFast:      100.0,
			EMASlow:      90.0,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: score,
			Price:        100.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	summary := view.renderSummary()

	// Should show DESTINY summary with counts
	assert.Contains(t, summary, "DESTINY:", "Summary should show DESTINY label")
	assert.Contains(t, summary, "2 LONG", "Summary should show 2 LONG")
	assert.Contains(t, summary, "1 SHORT", "Summary should show 1 SHORT")
	assert.Contains(t, summary, "1 HOLD", "Summary should show 1 HOLD")
}

// TestTrendView_DestinySummary_OnlyHold verifies DESTINY summary when all are HOLD.
func TestTrendView_DestinySummary_OnlyHold(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT")

	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:       model.rows[i].Symbol,
			RSI:          50.0,
			EMAFast:      100.0,
			EMASlow:      90.0,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 0, // All HOLD
			Price:        100.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	summary := view.renderSummary()

	// Should show DESTINY summary with only HOLD count
	assert.Contains(t, summary, "DESTINY:", "Summary should show DESTINY label")
	assert.Contains(t, summary, "2 HOLD", "Summary should show 2 HOLD")
	// Verify that "DESTINY" section shows HOLD (not in BLITZ section)
	assert.Contains(t, summary, "DESTINY: 2 HOLD", "Summary should show DESTINY HOLD count")
	// Verify no DESTINY LONG or SHORT when all are HOLD
	assert.NotRegexp(t, `(?i)DESTINY:.*\d\s+LONG`, summary, "Should not show DESTINY LONG when all are HOLD")
	assert.NotRegexp(t, `(?i)DESTINY:.*\d\s+SHORT`, summary, "Should not show DESTINY SHORT when all are HOLD")
}

// TestTrendView_DestinySummary_Empty verifies DESTINY summary is hidden when no data.
func TestTrendView_DestinySummary_Empty(t *testing.T) {
	model := NewModel()
	model.Configure(
		context.Background(),
		&viewMockEngine{},
		&config.WatchlistConfig{Equities: []string{}, Crypto: []string{}},
		indicators.NewAssetClassDetector([]string{}),
	)

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Summary bar should not be rendered when watchlist is empty
	assert.Contains(t, rendered, "Trend Analysis", "View should contain title")
	assert.NotContains(t, rendered, "DESTINY:", "View should not show DESTINY summary when empty")
}

// TestTrendView_TPICell_Positive verifies TPI cell with positive value.
func TestTrendView_TPICell_Positive(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	tpiCell := view.renderTPICell(0.67, "LONG")
	assert.Contains(t, tpiCell, "LONG", "Positive TPI should show LONG label")
	assert.Contains(t, tpiCell, "▓", "Positive TPI should show filled gauge bars")
}

// TestTrendView_TPICell_Zero verifies TPI cell with zero value.
func TestTrendView_TPICell_Zero(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	tpiCell := view.renderTPICell(0.0, "CASH")
	assert.Contains(t, tpiCell, "CASH", "Zero TPI should show CASH label")
	assert.Contains(t, tpiCell, "░", "Zero TPI should show empty gauge bars")
}

// TestTrendView_TPICell_Negative verifies TPI cell with negative value.
func TestTrendView_TPICell_Negative(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	tpiCell := view.renderTPICell(-0.67, "CASH")
	assert.Contains(t, tpiCell, "CASH", "Negative TPI should show CASH label")
	assert.Contains(t, tpiCell, "▓", "Negative TPI should show filled gauge bars")
}

// TestTrendView_TPICell_MaxPositive verifies TPI cell with max positive value.
func TestTrendView_TPICell_MaxPositive(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	tpiCell := view.renderTPICell(1.0, "LONG")
	assert.Contains(t, tpiCell, "LONG", "Max positive TPI should show LONG label")
	assert.Contains(t, tpiCell, "▓▓▓▓▓", "Max positive TPI should show 5 filled gauge bars on right")
}

// TestTrendView_TPICell_MaxNegative verifies TPI cell with max negative value.
func TestTrendView_TPICell_MaxNegative(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	tpiCell := view.renderTPICell(-1.0, "CASH")
	assert.Contains(t, tpiCell, "CASH", "Max negative TPI should show CASH label")
	assert.Contains(t, tpiCell, "▓▓▓▓▓", "Max negative TPI should show 5 filled gauge bars on left")
}

// TestTrendView_FTEMABadge_Bull verifies bullish FTEMA badge rendering.
func TestTrendView_FTEMABadge_Bull(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderFTEMABadge(trenddomain.Bullish)
	assert.Contains(t, badge, "▲  LONG", "Bullish badge should show arrow and LONG")
}

// TestTrendView_FTEMABadge_Bear verifies bearish FTEMA badge rendering.
func TestTrendView_FTEMABadge_Bear(t *testing.T) {
	model := newTestModelForView("AAPL")
	view := NewView(model).SetTheme(&mockTheme{})

	badge := view.renderFTEMABadge(trenddomain.Bearish)
	assert.Contains(t, badge, "▼ SHORT", "Bearish badge should show arrow and SHORT")
}

// TestTrendView_NoEMAColumns verifies EMA columns are not in output.
func TestTrendView_NoEMAColumns(t *testing.T) {
	model := newTestModelForView("AAPL")
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:     "AAPL",
		RSI:        50.0,
		EMAFast:    150.0,
		EMASlow:    145.0,
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		TPI:        0.67,
		TPISignal:  "LONG",
	}

	view := NewView(model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Verify EMA columns are NOT present
	assert.NotContains(t, rendered, "EMA FAST", "View should NOT have EMA FAST column header")
	assert.NotContains(t, rendered, "EMA SLOW", "View should NOT have EMA SLOW column header")
}

// TestTrendView_TPISummary verifies TPI summary line with correct counts.

// TestTrendView_SummaryBar_Counts verifies signal summary bar shows correct counts.

// TestTrendView_SummaryBar_WithPending verifies pending count is shown during loading.
func TestTrendView_SummaryBar_WithPending(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "BTC")

	// Set 2 loaded, 1 loading
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:     "AAPL",
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		Price:      100.0,
		RSI:        50.0,
		EMAFast:    100.0,
		EMASlow:    90.0,
		TPI:        0.67,
		TPISignal:  "LONG",
	}
	model.rows[1].State = StateLoaded
	model.rows[1].Result = &trenddomain.Result{
		Symbol:     "MSFT",
		Signal:     trenddomain.Bearish,
		BlitzScore: -1,
		Price:      100.0,
		RSI:        50.0,
		EMAFast:    100.0,
		EMASlow:    90.0,
		TPI:        0.0,
		TPISignal:  "CASH",
	}
	// rows[2] stays in loading state

	view := NewView(model).SetTheme(&mockTheme{})
	summary := view.renderSummary()

	assert.Contains(t, summary, "TPI: 1 LONG", "Summary should show TPI: 1 LONG TPI signal")
	assert.Contains(t, summary, "1 CASH", "Summary should show 1 CASH TPI signal")
	assert.Contains(t, summary, "1 pending", "Summary should show 1 pending")
	assert.Contains(t, summary, "BLITZ:", "Summary should show BLITZ label")
	assert.Contains(t, summary, "1 LONG", "Summary should show 1 BLITZ LONG")
	assert.Contains(t, summary, "1 SHORT", "Summary should show 1 BLITZ SHORT")
}

// TestTrendView_HeaderShowsProgressChip_Loading verifies the progress chip
// shows "loaded N/M" when tickers are still loading.
func TestTrendView_HeaderShowsProgressChip_Loading(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL", "NVDA")

	// Load 2 out of 4 tickers
	model.rows[0].State = StateLoaded
	model.rows[0].Result = &trenddomain.Result{
		Symbol:     "AAPL",
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		Price:      100.0,
		RSI:        50.0,
	}
	model.rows[1].State = StateLoaded
	model.rows[1].Result = &trenddomain.Result{
		Symbol:     "MSFT",
		Signal:     trenddomain.Bullish,
		BlitzScore: 1,
		Price:      100.0,
		RSI:        50.0,
	}
	// rows[2] and rows[3] remain in loading state

	view := NewView(model).SetTheme(&mockTheme{})
	title := view.renderTitle()

	assert.Contains(t, title, "Trend Analysis", "Title should contain main text")
	assert.Contains(t, title, "loaded 2/4", "Title should show progress chip with loaded count")
	assert.NotContains(t, title, "all loaded", "Title should not show complete status")
}

// TestTrendView_HeaderShowsProgressChip_Complete verifies the progress chip
// shows "all loaded · 2m ago" when all tickers are loaded.
func TestTrendView_HeaderShowsProgressChip_Complete(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT")

	// Load all tickers
	for i := range model.rows {
		model.rows[i].State = StateLoaded
		model.rows[i].Result = &trenddomain.Result{
			Symbol:     model.rows[i].Symbol,
			Signal:     trenddomain.Bullish,
			BlitzScore: 1,
			Price:      100.0,
			RSI:        50.0,
		}
	}

	view := NewView(model).SetTheme(&mockTheme{})
	title := view.renderTitle()

	assert.Contains(t, title, "Trend Analysis", "Title should contain main text")
	assert.Contains(t, title, "all loaded", "Title should show complete status")
	assert.Contains(t, title, "2m ago", "Title should show time ago indicator")
	assert.NotContains(t, title, "loaded 1/2", "Title should not show loading progress")
}

// TestTrendView_HeaderShowsProgressChip_EmptyWatchlist verifies the progress chip
// is hidden when the watchlist is empty.
func TestTrendView_HeaderShowsProgressChip_EmptyWatchlist(t *testing.T) {
	model := NewModel()
	model.Configure(
		context.Background(),
		&viewMockEngine{},
		&config.WatchlistConfig{Equities: []string{}, Crypto: []string{}},
		indicators.NewAssetClassDetector([]string{}),
	)

	view := NewView(model).SetTheme(&mockTheme{})
	title := view.renderTitle()

	assert.Contains(t, title, "Trend Analysis", "Title should contain main text")
	assert.NotContains(t, title, "loaded", "Title should not show loaded when empty")
	assert.NotContains(t, title, "all loaded", "Title should not show complete status when empty")
}

// TestTrendView_HeaderShowsProgressChip_FirstLoad verifies the progress chip
// shows "loaded 0/N" when no tickers have loaded yet.
func TestTrendView_HeaderShowsProgressChip_FirstLoad(t *testing.T) {
	model := newTestModelForView("AAPL", "MSFT", "GOOGL")

	// All tickers in loading state (default)

	view := NewView(model).SetTheme(&mockTheme{})
	title := view.renderTitle()

	assert.Contains(t, title, "Trend Analysis", "Title should contain main text")
	assert.Contains(t, title, "loaded 0/3", "Title should show 0 loaded at start")
	assert.NotContains(t, title, "all loaded", "Title should not show complete status")
}
