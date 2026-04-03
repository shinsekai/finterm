// Package trend provides the trend following TUI view.
package trend

import (
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
	TableHeader() lipgloss.Style
	TableEmpty() lipgloss.Style
	Bullish() lipgloss.Style
	Bearish() lipgloss.Style
	Neutral() lipgloss.Style
	Help() lipgloss.Style
	Error() lipgloss.Style
	Muted() lipgloss.Style
	Spinner() lipgloss.Style
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
	return v.theme.Title().Render("Trend Analysis")
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

	// Configure table
	v.table.
		WithColumns(columns).
		WithRows(tableRows).
		WithEmptyMessage("No data available")

	return v.table.Render()
}

// buildColumns returns the table column definitions.
func (v *View) buildColumns() []components.Column {
	return []components.Column{
		{
			Title:       "Symbol",
			Width:       10,
			Alignment:   components.AlignLeft,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "Signal",
			Width:       10,
			Alignment:   components.AlignLeft,
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
			Title:       "EMA Fast",
			Width:       12,
			Alignment:   components.AlignRight,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "EMA Slow",
			Width:       12,
			Alignment:   components.AlignRight,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "Valuation",
			Width:       14,
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

		// Apply highlight style to active row
		rowStyle := v.theme.TableRow()
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
		{Text: v.theme.Spinner().Render("Loading...")},
		{Text: "-"},
		{Text: "-"},
		{Text: "-"},
		{Text: "-"},
	}
}

// buildLoadedCells builds cells for a loaded row with color-coded signals.
func (v *View) buildLoadedCells(row RowData) []components.Cell {
	result := row.Result

	// Color-code signal
	var signalStyle lipgloss.Style
	switch result.Signal {
	case trenddomain.Bullish:
		signalStyle = v.theme.Bullish()
	case trenddomain.Bearish:
		signalStyle = v.theme.Bearish()
	default:
		signalStyle = v.theme.Neutral()
	}

	// Color-code valuation
	var valuationStyle lipgloss.Style
	switch result.Valuation {
	case "Oversold", "Undervalued":
		valuationStyle = v.theme.Bullish()
	case "Overvalued", "Overbought":
		valuationStyle = v.theme.Bearish()
	default:
		valuationStyle = v.theme.TableRow()
	}

	return []components.Cell{
		{Text: result.Symbol},
		{Text: signalStyle.Render(result.Signal.String())},
		{Text: FormatValue(result.RSI, 2)},
		{Text: FormatValue(result.EMAFast, 4)},
		{Text: FormatValue(result.EMASlow, 4)},
		{Text: valuationStyle.Render(result.Valuation)},
	}
}

// buildCachedCells builds cells for a cached row with offline indicator.
func (v *View) buildCachedCells(row RowData) []components.Cell {
	result := row.Result

	// Color-code signal
	var signalStyle lipgloss.Style
	switch result.Signal {
	case trenddomain.Bullish:
		signalStyle = v.theme.Bullish()
	case trenddomain.Bearish:
		signalStyle = v.theme.Bearish()
	default:
		signalStyle = v.theme.Neutral()
	}

	// Color-code valuation
	var valuationStyle lipgloss.Style
	switch result.Valuation {
	case "Oversold", "Undervalued":
		valuationStyle = v.theme.Bullish()
	case "Overvalued", "Overbought":
		valuationStyle = v.theme.Bearish()
	default:
		valuationStyle = v.theme.TableRow()
	}

	// Add "offline" badge to symbol
	symbolWithBadge := result.Symbol + " " + v.theme.Muted().Render("[offline]")

	return []components.Cell{
		{Text: symbolWithBadge},
		{Text: signalStyle.Render(result.Signal.String())},
		{Text: FormatValue(result.RSI, 2)},
		{Text: FormatValue(result.EMAFast, 4)},
		{Text: FormatValue(result.EMASlow, 4)},
		{Text: valuationStyle.Render(result.Valuation)},
	}
}

// buildErrorCells builds cells for an error row.
func (v *View) buildErrorCells(row RowData) []components.Cell {
	errorText := "Error"
	if row.Error != nil {
		errorText = row.Error.Error()
		// Truncate long error messages
		if len(errorText) > 20 {
			errorText = errorText[:17] + "..."
		}
	}

	return []components.Cell{
		{Text: row.Symbol},
		{Text: v.theme.Error().Render(errorText)},
		{Text: "-"},
		{Text: "-"},
		{Text: "-"},
		{Text: "-"},
	}
}

// buildUnknownCells builds cells for an unknown state row.
func (v *View) buildUnknownCells(row RowData) []components.Cell {
	return []components.Cell{
		{Text: row.Symbol},
		{Text: v.theme.Muted().Render("Unknown")},
		{Text: "-"},
		{Text: "-"},
		{Text: "-"},
		{Text: "-"},
	}
}

// renderFooter renders the footer with key bindings.
func (v *View) renderFooter() string {
	return v.theme.Help().Render("↑↓ Navigate  |  r Refresh  |  q Quit")
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

func (d *defaultTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#F8F8F2")
}

func (d *defaultTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282A36")
}
