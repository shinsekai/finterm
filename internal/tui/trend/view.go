// Package trend provides the trend following TUI view.
package trend

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	trenddomain "github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/tui/components"
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

	// Render signal summary (only if rows exist)
	if len(v.model.GetRows()) > 0 {
		builder.WriteString(v.renderSummary())
		builder.WriteString("\n\n")
	}

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
	loadedCount := v.model.GetLoadedCount()

	var subtitle string
	if loadedCount > 0 && loadedCount < symbolCount {
		// Show loading progress
		subtitle = fmt.Sprintf("  Loading %d/%d…", loadedCount, symbolCount)
	} else {
		// Show normal symbol count
		subtitle = fmt.Sprintf("  %d symbols", symbolCount)
	}

	return v.theme.Accent().Render("◆") + " " + v.theme.Title().Render("Trend Analysis") + v.theme.Muted().Render(subtitle)
}

// renderSummary renders the signal summary bar.
func (v *View) renderSummary() string {
	bullish, bearish, neutral := v.model.GetSignalCounts()
	long, short, hold := v.model.GetBlitzCounts()
	destinyLong, destinyShort, destinyHold := v.model.GetDestinyCounts()

	var summary strings.Builder

	if bullish > 0 {
		summary.WriteString(v.theme.Bullish().Render(fmt.Sprintf("%d ▲", bullish)))
	}

	if bearish > 0 {
		if summary.Len() > 0 {
			summary.WriteString("  ")
		}
		summary.WriteString(v.theme.Bearish().Render(fmt.Sprintf("%d ▼", bearish)))
	}

	if neutral > 0 {
		if summary.Len() > 0 {
			summary.WriteString("  ")
		}
		summary.WriteString(v.theme.Neutral().Render(fmt.Sprintf("%d ─", neutral)))
	}

	// Show pending count if some rows are still loading
	totalRows := len(v.model.GetRows())
	loadedCount := v.model.GetLoadedCount()
	pendingCount := totalRows - loadedCount
	if pendingCount > 0 {
		if summary.Len() > 0 {
			summary.WriteString("  ")
		}
		summary.WriteString(v.theme.Muted().Render(fmt.Sprintf("· %d pending", pendingCount)))
	}

	// BLITZ summary line (on a new line)
	if long > 0 || short > 0 || hold > 0 {
		var blitz strings.Builder
		blitz.WriteString("BLITZ: ")
		if long > 0 {
			blitz.WriteString(v.theme.Bullish().Render(fmt.Sprintf("%d LONG", long)))
		}
		if short > 0 {
			if blitz.Len() > 7 { // "BLITZ: " length
				blitz.WriteString("  ")
			}
			blitz.WriteString(v.theme.Bearish().Render(fmt.Sprintf("%d SHORT", short)))
		}
		if hold > 0 {
			if blitz.Len() > 7 { // "BLITZ: " length
				blitz.WriteString("  ")
			}
			blitz.WriteString(v.theme.Muted().Render(fmt.Sprintf("%d HOLD", hold)))
		}

		if summary.Len() > 0 {
			summary.WriteString("\n")
		}
		summary.WriteString(v.theme.Muted().Render(blitz.String()))
	}

	// DESTINY summary line (on a new line after BLITZ)
	if destinyLong > 0 || destinyShort > 0 || destinyHold > 0 {
		var destiny strings.Builder
		destiny.WriteString("DESTINY: ")
		if destinyLong > 0 {
			destiny.WriteString(v.theme.Bullish().Render(fmt.Sprintf("%d LONG", destinyLong)))
		}
		if destinyShort > 0 {
			if destiny.Len() > 9 { // "DESTINY: " length
				destiny.WriteString("  ")
			}
			destiny.WriteString(v.theme.Bearish().Render(fmt.Sprintf("%d SHORT", destinyShort)))
		}
		if destinyHold > 0 {
			if destiny.Len() > 9 { // "DESTINY: " length
				destiny.WriteString("  ")
			}
			destiny.WriteString(v.theme.Muted().Render(fmt.Sprintf("%d HOLD", destinyHold)))
		}

		if summary.Len() > 0 {
			summary.WriteString("\n")
		}
		summary.WriteString(v.theme.Muted().Render(destiny.String()))
	}

	return summary.String()
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

	// Configure table with full width (reduce by 4 to account for border + padding)
	v.table.
		WithColumns(columns).
		WithRows(tableRows).
		WithEmptyMessage("No data available").
		WithMaxWidth(v.model.GetWidth() - 4)

	// Wrap table in bordered container
	borderColor := v.theme.Divider().GetForeground()
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	return containerStyle.Render(v.table.Render())
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
			Title:       "BLITZ",
			Width:       10,
			Alignment:   components.AlignCenter,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "DESTINY",
			Width:       10,
			Alignment:   components.AlignCenter,
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
			Width:       7,
			Alignment:   components.AlignRight,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "EMA FAST",
			Width:       11,
			Alignment:   components.AlignRight,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "EMA SLOW",
			Width:       11,
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
	activeRow := v.model.GetActiveRow()
	tableRows := make([]components.Row, 0, len(rows)+1)

	for i, row := range rows {
		cells := v.buildRowCells(row)

		// Add cursor marker to symbol cell
		cursor := "  "
		if i == activeRow {
			cursor = "▸ "
		}
		cells[0].Text = cursor + cells[0].Text

		// Apply alternating row style
		rowStyle := v.theme.TableRow()
		if i%2 == 1 {
			rowStyle = v.theme.TableRowAlt()
		}

		// Apply highlight style to active row (overrides alternating)
		if i == activeRow {
			rowStyle = rowStyle.Background(v.theme.Foreground()).Foreground(v.theme.Background())
		}

		tableRows = append(tableRows, components.Row{
			Cells: cells,
			Style: rowStyle,
		})
	}

	// Add section separator between equities and crypto
	cryptoStartIndex := v.model.GetCryptoStartIndex()
	if cryptoStartIndex > 0 && cryptoStartIndex < len(rows) {
		// Insert separator at the crypto start position
		separatorCells := make([]components.Cell, 9)
		separatorCells[0].Text = v.theme.Divider().Render("── Crypto ──")
		// All other cells remain empty

		separatorRow := components.Row{
			Cells: separatorCells,
			Style: v.theme.Muted(),
		}

		// Insert separator at the right position
		tableRows = append(tableRows[:cryptoStartIndex],
			append([]components.Row{separatorRow}, tableRows[cryptoStartIndex:]...)...)
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
		{Text: v.renderBlitzBadge(result.BlitzScore)},
		{Text: v.renderDestinyBadge(result.DestinyScore)},
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
		{Text: v.renderBlitzBadge(result.BlitzScore)},
		{Text: v.renderDestinyBadge(result.DestinyScore)},
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

// renderBlitzBadge renders a BLITZ score as a colored badge.
func (v *View) renderBlitzBadge(blitzScore int) string {
	switch blitzScore {
	case 1:
		return v.theme.BullishBadge().Render("▲ LONG")
	case -1:
		return v.theme.BearishBadge().Render("▼ SHORT")
	default:
		return v.theme.NeutralBadge().Render("─ HOLD")
	}
}

// renderDestinyBadge renders a DESTINY score as a colored badge.
func (v *View) renderDestinyBadge(destinyScore int) string {
	switch destinyScore {
	case 1:
		return v.theme.BullishBadge().Render("▲ LONG")
	case -1:
		return v.theme.BearishBadge().Render("▼ SHORT")
	default:
		return v.theme.Muted().Render("○ HOLD")
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
// All prices are prefixed with "$" unless zero (returns "—").
func formatPrice(price float64) string {
	if price == 0 {
		return "—"
	}

	// For very small prices (crypto fractions), show more precision
	if price < 1 {
		return "$" + fmt.Sprintf("%.6f", price)
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
	return "$" + str + fracStr[1:] // Prepend "$", remove leading "0"
}
