// Package trend provides the trend following TUI view.
package trend

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	trenddomain "github.com/owner/finterm/internal/domain/trend"
	"github.com/owner/finterm/internal/tui/components"
)

// Theme defines the interface for theme styling.
// This is defined here to avoid import cycle with the parent tui package.
type Theme interface {
	Title() lipgloss.Style
	TableRow() lipgloss.Style
	TableRowAlt() lipgloss.Style
	TableHeader() lipgloss.Style
	TableEmpty() lipgloss.Style
	Bullish() lipgloss.Style
	Bearish() lipgloss.Style
	BullishBadge() lipgloss.Style
	BearishBadge() lipgloss.Style
	Neutral() lipgloss.Style
	NeutralBadge() lipgloss.Style
	Help() lipgloss.Style
	Error() lipgloss.Style
	Muted() lipgloss.Style
	Spinner() lipgloss.Style
	Accent() lipgloss.Style
	Divider() lipgloss.Style
	Foreground() lipgloss.Color
	Background() lipgloss.Color
}

// View handles rendering of the trend view.
type View struct {
	model Model
	theme Theme
	table *components.Table
}

// NewView creates a new view for the trend model.
func NewView(model Model) *View {
	return &View{
		model: model,
		table: components.NewTable(),
	}
}

// SetTheme sets the theme for the view.
func (v *View) SetTheme(theme Theme) *View {
	v.theme = theme
	return v
}

// Render renders the trend view as a string.
func (v *View) Render() string {
	if v.theme == nil {
		// Use default styles if no theme is set
		v.theme = &defaultTheme{}
	}

	var builder strings.Builder

	// Render title
	builder.WriteString(v.renderTitle())
	builder.WriteString("\n\n")

	// Render table
	builder.WriteString(v.renderTable())
	builder.WriteString("\n\n")

	// Render footer
	builder.WriteString(v.renderFooter())

	return builder.String()
}

// renderTitle renders the view title.
func (v *View) renderTitle() string {
	symbolCount := len(v.model.GetRows())
	subtitle := fmt.Sprintf("  %d symbols", symbolCount)
	return v.theme.Accent().Render("◆") + " " + v.theme.Title().Render("Trend Analysis") + v.theme.Muted().Render(subtitle)
}

// renderTable renders the data table.
func (v *View) renderTable() string {
	rows := v.model.GetRows()
	if len(rows) == 0 {
		return v.theme.TableEmpty().Render("No tickers in watchlist")
	}

	// Build table columns
	columns := v.buildColumns()

	// Build table rows
	tableRows := v.buildTableRows(rows)

	// Configure table with full width
	v.table.
		WithColumns(columns).
		WithRows(tableRows).
		WithEmptyMessage("No data available").
		WithMaxWidth(v.model.GetWidth())

	return v.table.Render()
}

// buildColumns returns the table column definitions.
func (v *View) buildColumns() []components.Column {
	return []components.Column{
		{
			Title:       "SYMBOL",
			Width:       10,
			Alignment:   components.AlignLeft,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "SIGNAL",
			Width:       15,
			Alignment:   components.AlignLeft,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "PRICE",
			Width:       14,
			Alignment:   components.AlignRight,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "RSI",
			Width:       8,
			Alignment:   components.AlignRight,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "EMA FAST",
			Width:       12,
			Alignment:   components.AlignRight,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "EMA SLOW",
			Width:       12,
			Alignment:   components.AlignRight,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "VALUATION",
			Width:       0, // 0 means auto-calculate to fill remaining space
			Alignment:   components.AlignLeft,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
	}
}

// buildTableRows converts model row data to table rows.
func (v *View) buildTableRows(rows []RowData) []components.Row {
	tableRows := make([]components.Row, len(rows))
	activeRow := v.model.GetActiveRow()

	for i, row := range rows {
		cells := v.buildRowCells(row)

		// Apply alternating row style
		rowStyle := v.theme.TableRow()
		if i%2 == 1 {
			rowStyle = v.theme.TableRowAlt()
		}

		// Apply highlight style to active row (overrides alternating)
		if i == activeRow {
			rowStyle = rowStyle.Background(v.theme.Foreground()).Foreground(v.theme.Background())
		}

		tableRows[i] = components.Row{
			Cells: cells,
			Style: rowStyle,
		}
	}

	return tableRows
}

// buildRowCells builds the cells for a single row.
func (v *View) buildRowCells(row RowData) []components.Cell {
	switch row.State {
	case StateLoading:
		return v.buildLoadingCells(row)
	case StateLoaded:
		return v.buildLoadedCells(row)
	case StateCached:
		return v.buildCachedCells(row)
	case StateError:
		return v.buildErrorCells(row)
	default:
		return v.buildUnknownCells(row)
	}
}

// buildLoadingCells builds cells for a loading row.
func (v *View) buildLoadingCells(row RowData) []components.Cell {
	return []components.Cell{
		{Text: row.Symbol},
		{Text: v.theme.Spinner().Render("⟳ Loading…")},
		{Text: "—"},
		{Text: "—"},
		{Text: "—"},
		{Text: "—"},
		{Text: "—"},
	}
}

// buildLoadedCells builds cells for a loaded row with color-coded signals.
func (v *View) buildLoadedCells(row RowData) []components.Cell {
	result := row.Result

	return []components.Cell{
		{Text: v.theme.Accent().Render(result.Symbol)},
		{Text: v.renderSignalBadge(result.Signal)},
		{Text: formatPrice(result.Price)},
		{Text: v.renderRSIValue(result.RSI)},
		{Text: FormatValue(result.EMAFast, 2)},
		{Text: FormatValue(result.EMASlow, 2)},
		{Text: v.renderValuationBadge(result.Valuation)},
	}
}

// buildCachedCells builds cells for a cached row with offline indicator.
func (v *View) buildCachedCells(row RowData) []components.Cell {
	result := row.Result

	// Add ○ offline indicator to symbol
	symbolWithBadge := v.theme.Accent().Render(result.Symbol) + " " + v.theme.Muted().Render("○")

	return []components.Cell{
		{Text: symbolWithBadge},
		{Text: v.renderSignalBadge(result.Signal)},
		{Text: formatPrice(result.Price)},
		{Text: v.renderRSIValue(result.RSI)},
		{Text: FormatValue(result.EMAFast, 2)},
		{Text: FormatValue(result.EMASlow, 2)},
		{Text: v.renderValuationBadge(result.Valuation)},
	}
}

// buildErrorCells builds cells for an error row.
func (v *View) buildErrorCells(row RowData) []components.Cell {
	errorText := "✗ Error"
	if row.Error != nil {
		errorText = "✗ " + row.Error.Error()
		// Show full error text (no truncation)
	}

	return []components.Cell{
		{Text: row.Symbol},
		{Text: v.theme.Error().Render(errorText)},
		{Text: "—"},
		{Text: "—"},
		{Text: "—"},
		{Text: "—"},
		{Text: "—"},
	}
}

// buildUnknownCells builds cells for an unknown state row.
func (v *View) buildUnknownCells(row RowData) []components.Cell {
	return []components.Cell{
		{Text: row.Symbol},
		{Text: v.theme.Muted().Render("Unknown")},
		{Text: "—"},
		{Text: "—"},
		{Text: "—"},
		{Text: "—"},
		{Text: "—"},
	}
}

// renderFooter renders the footer with key bindings.
func (v *View) renderFooter() string {
	return v.theme.Accent().Render("↑↓") + " " + v.theme.Muted().Render("navigate") +
		"  ·  " +
		v.theme.Accent().Render("r") + " " + v.theme.Muted().Render("refresh") +
		"  ·  " +
		v.theme.Accent().Render("1-4") + " " + v.theme.Muted().Render("tabs") +
		"  ·  " +
		v.theme.Accent().Render("?") + " " + v.theme.Muted().Render("help")
}

// renderSignalBadge renders a signal as a colored badge.
func (v *View) renderSignalBadge(signal trenddomain.Signal) string {
	switch signal {
	case trenddomain.Bullish:
		return v.theme.BullishBadge().Render("▲ BULL")
	case trenddomain.Bearish:
		return v.theme.BearishBadge().Render("▼ BEAR")
	default:
		return v.theme.NeutralBadge().Render("─ HOLD")
	}
}

// renderRSIValue renders an RSI value with appropriate color coding.
func (v *View) renderRSIValue(rsi float64) string {
	var style lipgloss.Style
	switch {
	case rsi < 30:
		// Oversold - bullish style (green)
		style = v.theme.Bullish()
	case rsi > 70:
		// Overbought - bearish style (red)
		style = v.theme.Bearish()
	case rsi >= 30 && rsi < 45:
		// Undervalued - lighter green
		style = v.theme.Bullish()
	case rsi > 55 && rsi <= 70:
		// Overvalued - lighter red
		style = v.theme.Bearish()
	default:
		// Fair value (45-55) - default text
		style = lipgloss.NewStyle()
	}
	return style.Render(FormatValue(rsi, 2))
}

// renderValuationBadge renders a valuation as a styled badge with icon.
func (v *View) renderValuationBadge(valuation string) string {
	switch valuation {
	case "Oversold":
		return v.theme.Bullish().Render("◆ Oversold")
	case "Undervalued":
		return v.theme.Bullish().Render("◇ Undervalued")
	case "Overbought":
		return v.theme.Bearish().Render("◆ Overbought")
	case "Overvalued":
		return v.theme.Bearish().Render("◇ Overvalued")
	case "Fair value":
		return v.theme.Muted().Render("○ Fair value")
	default:
		return valuation
	}
}

// String returns the rendered view.
func (v *View) String() string {
	return v.Render()
}

// defaultTheme provides a fallback theme when no theme is set.
type defaultTheme struct{}

func (d *defaultTheme) Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (d *defaultTheme) TableRow() lipgloss.Style {
	return lipgloss.NewStyle()
}

func (d *defaultTheme) TableRowAlt() lipgloss.Style {
	return lipgloss.NewStyle().Background(lipgloss.Color("#2D2F3D"))
}

func (d *defaultTheme) TableHeader() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (d *defaultTheme) TableEmpty() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true)
}

func (d *defaultTheme) Bullish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("50FA7B")).Bold(true)
}

func (d *defaultTheme) Bearish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("FF5555")).Bold(true)
}

func (d *defaultTheme) Neutral() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("F1FA8C")).Bold(true)
}

func (d *defaultTheme) BullishBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("50FA7B")).
		Bold(true)
}

func (d *defaultTheme) BearishBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Background(lipgloss.Color("FF5555")).
		Bold(true)
}

func (d *defaultTheme) NeutralBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("F1FA8C")).
		Bold(true)
}

func (d *defaultTheme) Help() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true)
}

func (d *defaultTheme) Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("FF5555")).Bold(true)
}

func (d *defaultTheme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Faint(true)
}

func (d *defaultTheme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("7D56F4"))
}

func (d *defaultTheme) Accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("7D56F4")).Bold(true)
}

func (d *defaultTheme) Divider() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
}

func (d *defaultTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#F8F8F2")
}

func (d *defaultTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282A36")
}

// formatPrice formats a price value with comma separators and 2 decimal places.
// For prices < 1, shows more decimals; for prices >= 1000, adds comma separators.
func formatPrice(price float64) string {
	if price == 0 {
		return "--"
	}

	// For very small prices (crypto fractions), show more precision
	if price < 1 {
		return fmt.Sprintf("%.6f", price)
	}

	// For normal prices, format with 2 decimal places
	intPart := int64(price)
	fracPart := price - float64(intPart)

	// Format integer part with commas
	str := fmt.Sprintf("%d", intPart)
	n := len(str)
	if n > 3 {
		var result []byte
		for i, c := range str {
			if i > 0 && (n-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		str = string(result)
	}

	// Add fractional part
	fracStr := fmt.Sprintf("%.2f", fracPart)
	return str + fracStr[1:] // Remove leading "0"
}
