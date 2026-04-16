// Package quote provides tests for the quote TUI view.
package quote

import (
	"context"
	"regexp"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/shinsekai/finterm/internal/alphavantage"
	trenddomain "github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
)

// viewMockEngine is a mock implementation of Engine for view tests.
type viewMockEngine struct{}

func (m *viewMockEngine) AnalyzeWithSymbolDetection(_ context.Context, _ string) (*trenddomain.Result, error) {
	return nil, nil
}

// mockQuoteClient is a mock implementation of QuoteClient for view tests.
type mockQuoteClient struct{}

func (m *mockQuoteClient) GetGlobalQuote(_ context.Context, _ string) (*alphavantage.GlobalQuote, error) {
	return nil, nil
}

// mockTheme is a mock implementation of Theme for testing.
type mockTheme struct{}

func (m *mockTheme) Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (m *mockTheme) Subtitle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
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
		Bold(true).
		Padding(0, 1)
}

func (m *mockTheme) BearishBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("#FF5555")).
		Bold(true).
		Padding(0, 1)
}

func (m *mockTheme) NeutralBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("#F1FA8C")).
		Bold(true).
		Padding(0, 1)
}

func (m *mockTheme) Help() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true)
}

func (m *mockTheme) Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("FF5555")).Bold(true)
}

func (m *mockTheme) Loading() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("8BE9FD"))
}

func (m *mockTheme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Faint(true)
}

func (m *mockTheme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("7D56F4"))
}

func (m *mockTheme) SpinnerText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("6272A4"))
}

func (m *mockTheme) Box() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#44475A")).
		Padding(1)
}

func (m *mockTheme) Card() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#44475A")).
		Padding(1, 2)
}

func (m *mockTheme) CardHeader() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BD93F9")).
		Bold(true).
		MarginBottom(1)
}

func (m *mockTheme) Accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("7D56F4")).Bold(true)
}

func (m *mockTheme) Divider() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
}

func (m *mockTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#7D56F4")
}

func (m *mockTheme) Border() lipgloss.Color {
	return lipgloss.Color("#44475A")
}

func (m *mockTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#F8F8F2")
}

func (m *mockTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282A36")
}

// newTestModelForView creates a configured model for testing.
func newTestModelForView(width, height int) *Model {
	model := NewModel()
	model.width = width
	model.height = height
	return model.Configure(
		context.Background(),
		&mockQuoteClient{},
		&viewMockEngine{},
		indicators.NewAssetClassDetector([]string{}),
	)
}

// TestQuoteView_SideBySideLayout_WideTerminal verifies that on a wide terminal,
// both cards render side-by-side and all content (including Composite section) is visible.
func TestQuoteView_SideBySideLayout_WideTerminal(t *testing.T) {
	model := newTestModelForView(120, 40)

	// Set model to loaded state
	model.state = StateLoaded

	// Set up loaded quote data with all signal systems
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:         "MSTR",
			Price:          "139.75",
			Open:           "137.50",
			High:           "141.25",
			Low:            "136.00",
			Volume:         "5432100",
			Change:         "2.25",
			ChangePercent:  "1.63%",
			PreviousClose:  "137.50",
			LastTradingDay: "2024-01-15",
		},
		Indicators: &trenddomain.Result{
			RSI:          55.5,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			FlowScore:    1,
			VortexScore:  1,
			TPI:          0.20,
			TPISignal:    "LONG",
		},
	}

	view := NewView(*model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Verify all 5 signal system labels are present
	assert.Contains(t, rendered, "FTEMA", "View should contain FTEMA label")
	assert.Contains(t, rendered, "BLITZ", "View should contain BLITZ label")
	assert.Contains(t, rendered, "DESTINY", "View should contain DESTINY label")
	assert.Contains(t, rendered, "FLOW", "View should contain FLOW label")
	assert.Contains(t, rendered, "VORTEX", "View should contain VORTEX label")

	// Verify Composite section is present
	assert.Contains(t, rendered, "Composite", "View should contain Composite section")

	// Verify TPI numeric value is present (pattern matches +0.20, -0.30, etc.)
	tpiPattern := regexp.MustCompile(`[+-]\d+\.\d{2}`)
	assert.Regexp(t, tpiPattern, rendered, "View should contain TPI numeric value")

	// Verify symbol header is present
	assert.Contains(t, rendered, "MSTR", "View should contain symbol")
	assert.Contains(t, rendered, "Technical Indicators", "View should contain Technical Indicators header")

	// Verify no double-dollar prefix
	assert.NotContains(t, rendered, "$$", "View should not contain double-dollar prefix")
}

// TestQuoteView_VerticalFallback_NarrowTerminal verifies that on a narrow terminal,
// the view falls back to vertical stacking without losing content.
func TestQuoteView_VerticalFallback_NarrowTerminal(t *testing.T) {
	model := newTestModelForView(80, 40)

	// Set model to loaded state
	model.state = StateLoaded

	// Set up loaded quote data with all signal systems
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:         "MSTR",
			Price:          "139.75",
			Open:           "137.50",
			High:           "141.25",
			Low:            "136.00",
			Volume:         "5432100",
			Change:         "2.25",
			ChangePercent:  "1.63%",
			PreviousClose:  "137.50",
			LastTradingDay: "2024-01-15",
		},
		Indicators: &trenddomain.Result{
			RSI:          55.5,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			FlowScore:    1,
			VortexScore:  1,
			TPI:          0.20,
			TPISignal:    "LONG",
		},
	}

	view := NewView(*model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Verify all 5 signal system labels are still present
	assert.Contains(t, rendered, "FTEMA", "View should contain FTEMA label")
	assert.Contains(t, rendered, "BLITZ", "View should contain BLITZ label")
	assert.Contains(t, rendered, "DESTINY", "View should contain DESTINY label")
	assert.Contains(t, rendered, "FLOW", "View should contain FLOW label")
	assert.Contains(t, rendered, "VORTEX", "View should contain VORTEX label")

	// Verify Composite section is still present
	assert.Contains(t, rendered, "Composite", "View should contain Composite section")

	// Verify TPI numeric value is still present
	tpiPattern := regexp.MustCompile(`[+-]\d+\.\d{2}`)
	assert.Regexp(t, tpiPattern, rendered, "View should contain TPI numeric value")
}

// TestQuoteView_NoDoubleDollarSign verifies that price fields show single $ prefix.
func TestQuoteView_NoDoubleDollarSign(t *testing.T) {
	model := newTestModelForView(120, 40)

	// Set model to loaded state
	model.state = StateLoaded

	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:         "AAPL",
			Price:          "155.50",
			Open:           "153.00",
			High:           "157.00",
			Low:            "152.00",
			Volume:         "50000000",
			Change:         "2.50",
			ChangePercent:  "1.63%",
			PreviousClose:  "153.00",
			LastTradingDay: "2024-01-15",
		},
		Indicators: &trenddomain.Result{
			RSI:       50.0,
			Signal:    trenddomain.Bullish,
			TPI:       0.50,
			TPISignal: "LONG",
		},
	}

	view := NewView(*model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Verify single $ prefix is present
	assert.Contains(t, rendered, "$155", "View should contain single-dollar prefix for price")
	assert.Contains(t, rendered, "$153", "View should contain single-dollar prefix for price fields")

	// Verify NO double-dollar prefix
	assert.NotContains(t, rendered, "$$", "View should not contain double-dollar prefix")
}

// TestQuoteView_CompositeSectionRendered verifies that the Composite section
// with TPI value is rendered when indicators are present.
func TestQuoteView_CompositeSectionRendered(t *testing.T) {
	model := newTestModelForView(120, 40)

	// Set model to loaded state
	model.state = StateLoaded

	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:         "AAPL",
			Price:          "155.50",
			Open:           "153.00",
			High:           "157.00",
			Low:            "152.00",
			Volume:         "50000000",
			Change:         "2.50",
			ChangePercent:  "1.63%",
			PreviousClose:  "153.00",
			LastTradingDay: "2024-01-15",
		},
		Indicators: &trenddomain.Result{
			RSI:       50.0,
			Signal:    trenddomain.Bullish,
			TPI:       0.75,
			TPISignal: "LONG",
		},
	}

	view := NewView(*model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Verify Composite section is present
	assert.Contains(t, rendered, "Composite", "View should contain Composite section")

	// Verify TPI value is present
	assert.Contains(t, rendered, "+0.75", "View should contain TPI value +0.75")
}

// TestQuoteView_PriceCardOnly_NoIndicators verifies that when indicators are nil,
// only the price card is rendered.
func TestQuoteView_PriceCardOnly_NoIndicators(t *testing.T) {
	model := newTestModelForView(120, 40)

	// Set model to loaded state
	model.state = StateLoaded

	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:         "AAPL",
			Price:          "155.50",
			Open:           "153.00",
			High:           "157.00",
			Low:            "152.00",
			Volume:         "50000000",
			Change:         "2.50",
			ChangePercent:  "1.63%",
			PreviousClose:  "153.00",
			LastTradingDay: "2024-01-15",
		},
		Indicators: nil, // No indicators
	}

	view := NewView(*model).SetTheme(&mockTheme{})
	rendered := view.Render()

	// Verify price card content is present
	assert.Contains(t, rendered, "AAPL", "View should contain symbol")
	assert.Contains(t, rendered, "$155", "View should contain price")

	// Verify indicators content is NOT present
	assert.NotContains(t, rendered, "Technical Indicators", "View should not contain Technical Indicators header")
	assert.NotContains(t, rendered, "RSI", "View should not contain RSI label")
	assert.NotContains(t, rendered, "FTEMA", "View should not contain FTEMA label")
	assert.NotContains(t, rendered, "Composite", "View should not contain Composite section")
}

// TestCalculateBoxWidth_WideTerminal verifies box width calculation on a wide terminal.
func TestCalculateBoxWidth_WideTerminal(t *testing.T) {
	model := *newTestModelForView(120, 40)
	view := NewView(model).SetTheme(&mockTheme{})

	width := view.calculateBoxWidth()
	assert.Equal(t, 48, width, "Wide terminal (120) should return max width 48")
}

// TestCalculateBoxWidth_MediumTerminal verifies box width calculation on a medium terminal.
func TestCalculateBoxWidth_MediumTerminal(t *testing.T) {
	model := *newTestModelForView(100, 40)
	view := NewView(model).SetTheme(&mockTheme{})

	width := view.calculateBoxWidth()
	assert.Equal(t, 45, width, "Medium terminal (100) should return (100-10)/2 = 45")
}

// TestCalculateBoxWidth_NarrowTerminal verifies box width is clamped to minimum on a narrow terminal.
func TestCalculateBoxWidth_NarrowTerminal(t *testing.T) {
	model := *newTestModelForView(60, 40)
	view := NewView(model).SetTheme(&mockTheme{})

	width := view.calculateBoxWidth()
	assert.Equal(t, 30, width, "Narrow terminal (60) should return minimum width 30")
}

// TestCalculateBoxWidth_VeryNarrowTerminal verifies box width is clamped to minimum on a very narrow terminal.
func TestCalculateBoxWidth_VeryNarrowTerminal(t *testing.T) {
	model := *newTestModelForView(40, 40)
	view := NewView(model).SetTheme(&mockTheme{})

	width := view.calculateBoxWidth()
	assert.Equal(t, 30, width, "Very narrow terminal (40) should return minimum width 30")
}

// TestQuoteView_DefaultTheme verifies view works with default theme.
func TestQuoteView_DefaultTheme(t *testing.T) {
	model := newTestModelForView(120, 40)

	// Set model to loaded state
	model.state = StateLoaded

	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:         "AAPL",
			Price:          "155.50",
			Open:           "153.00",
			High:           "157.00",
			Low:            "152.00",
			Volume:         "50000000",
			Change:         "2.50",
			ChangePercent:  "1.63%",
			PreviousClose:  "153.00",
			LastTradingDay: "2024-01-15",
		},
		Indicators: &trenddomain.Result{
			RSI:       50.0,
			Signal:    trenddomain.Bullish,
			TPI:       0.50,
			TPISignal: "LONG",
		},
	}

	// Create view without setting theme - should use default
	view := NewView(*model)
	rendered := view.Render()

	assert.Contains(t, rendered, "Quote Lookup", "View should contain title with default theme")
	assert.NotEmpty(t, rendered, "View should not be empty")
}
