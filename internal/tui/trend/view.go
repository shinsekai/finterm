// Package trend provides the trend following TUI view.
package trend

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	trenddomain "github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// Hardcoded color styles for table cell rendering.
// These direct foreground colors survive table row styling because they're applied to inner text content.
var (
	greenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")) // Dracula green
	redStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")) // Dracula red
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
	model *Model
	theme Theme
	table *components.Table
}

// NewView creates a new view for the trend model.
func NewView(model *Model) *View {
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
// nolint:gocyclo // High complexity due to conditional rendering of multiple summary sections
func (v *View) renderSummary() string {
	long, cash := v.model.GetTPICounts()
	blitzLong, blitzShort, blitzHold := v.model.GetBlitzCounts()
	destinyLong, destinyShort, destinyHold := v.model.GetDestinyCounts()

	var summary strings.Builder

	// TPI summary
	if long > 0 || cash > 0 {
		summary.WriteString(v.theme.Muted().Render("TPI: "))
		if long > 0 {
			summary.WriteString(v.theme.Bullish().Render(fmt.Sprintf("%d LONG", long)))
		}

		if cash > 0 {
			if long > 0 {
				summary.WriteString("  ")
			}
			summary.WriteString(v.theme.Muted().Render(fmt.Sprintf("%d CASH", cash)))
		}
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

	flowLong, flowShort, flowHold := v.model.GetFlowCounts()

	// BLITZ summary line (on a new line)
	if blitzLong > 0 || blitzShort > 0 || blitzHold > 0 {
		var blitz strings.Builder
		blitz.WriteString("BLITZ: ")
		if blitzLong > 0 {
			blitz.WriteString(v.theme.Bullish().Render(fmt.Sprintf("%d LONG", blitzLong)))
		}
		if blitzShort > 0 {
			if blitz.Len() > 7 { // "BLITZ: " length
				blitz.WriteString("  ")
			}
			blitz.WriteString(v.theme.Bearish().Render(fmt.Sprintf("%d SHORT", blitzShort)))
		}
		if blitzHold > 0 {
			if blitz.Len() > 7 { // "BLITZ: " length
				blitz.WriteString("  ")
			}
			blitz.WriteString(v.theme.Muted().Render(fmt.Sprintf("%d HOLD", blitzHold)))
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

	// FLOW summary line (on a new line after DESTINY)
	if flowLong > 0 || flowShort > 0 || flowHold > 0 {
		var flow strings.Builder
		flow.WriteString("FLOW: ")
		if flowLong > 0 {
			flow.WriteString(v.theme.Bullish().Render(fmt.Sprintf("%d LONG", flowLong)))
		}
		if flowShort > 0 {
			if flow.Len() > 6 { // "FLOW: " length
				flow.WriteString("  ")
			}
			flow.WriteString(v.theme.Bearish().Render(fmt.Sprintf("%d SHORT", flowShort)))
		}
		if flowHold > 0 {
			if flow.Len() > 6 { // "FLOW: " length
				flow.WriteString("  ")
			}
			flow.WriteString(v.theme.Muted().Render(fmt.Sprintf("%d HOLD", flowHold)))
		}

		if summary.Len() > 0 {
			summary.WriteString("\n")
		}
		summary.WriteString(v.theme.Muted().Render(flow.String()))
	}

	// VORTEX summary line (on a new line after FLOW)
	vortexLong, vortexShort, vortexHold := v.model.GetVortexCounts()
	if vortexLong > 0 || vortexShort > 0 || vortexHold > 0 {
		var vortex strings.Builder
		vortex.WriteString("VORTEX: ")
		if vortexLong > 0 {
			vortex.WriteString(v.theme.Bullish().Render(fmt.Sprintf("%d LONG", vortexLong)))
		}
		if vortexShort > 0 {
			if vortex.Len() > 8 {
				vortex.WriteString("  ")
			}
			vortex.WriteString(v.theme.Bearish().Render(fmt.Sprintf("%d SHORT", vortexShort)))
		}
		if vortexHold > 0 {
			if vortex.Len() > 8 {
				vortex.WriteString("  ")
			}
			vortex.WriteString(v.theme.Muted().Render(fmt.Sprintf("%d HOLD", vortexHold)))
		}

		if summary.Len() > 0 {
			summary.WriteString("\n")
		}
		summary.WriteString(v.theme.Muted().Render(vortex.String()))
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
			Title:       "TPI",
			Width:       24,
			Alignment:   components.AlignLeft,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "FTEMA",
			Width:       10,
			Alignment:   components.AlignCenter,
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
			Title:       "FLOW",
			Width:       10,
			Alignment:   components.AlignCenter,
			Style:       v.theme.TableRow(),
			HeaderStyle: v.theme.TableHeader(),
		},
		{
			Title:       "VORTEX",
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
		separatorCells := make([]components.Cell, 10)
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
		{Text: "—"},
		{Text: "—"},
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
		{Text: v.renderTPICell(result.TPI, result.TPISignal)},
		{Text: v.renderFTEMABadge(result.Signal)},
		{Text: v.renderBlitzBadge(result.BlitzScore)},
		{Text: v.renderDestinyBadge(result.DestinyScore)},
		{Text: v.renderFlowBadge(result.FlowScore)},
		{Text: v.renderVortexBadge(result.VortexScore)},
		{Text: formatPrice(result.Price)},
		{Text: v.renderRSIValue(result.RSI)},
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
		{Text: v.renderTPICell(result.TPI, result.TPISignal)},
		{Text: v.renderFTEMABadge(result.Signal)},
		{Text: v.renderBlitzBadge(result.BlitzScore)},
		{Text: v.renderDestinyBadge(result.DestinyScore)},
		{Text: v.renderFlowBadge(result.FlowScore)},
		{Text: v.renderVortexBadge(result.VortexScore)},
		{Text: formatPrice(result.Price)},
		{Text: v.renderRSIValue(result.RSI)},
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

// renderTPICell renders the TPI gauge, numeric value, and label.
// Gauge is 10 chars wide: left half (0-4) is bearish, center (5) is neutral, right half (6-9) is bullish.
// For TPI > 0: fill from center rightward with "▓" in bullish color, "░" in muted.
// For TPI <= 0: fill from center leftward with "▓" in bearish color, "░" in muted.
// The numeric value is shown with sign (+/-) and color-coded based on TPI sign.
func (v *View) renderTPICell(tpi float64, tpiSignal string) string {
	// Clamp TPI to [-1, 1]
	switch {
	case tpi > 1:
		tpi = 1
	case tpi < -1:
		tpi = -1
	}

	// Build gauge: 10 characters
	// Positions: 0 1 2 3 4 5 6 7 8 9
	// Center is at position 5
	gauge := make([]rune, 10)

	// Fill with empty blocks first
	for i := range gauge {
		gauge[i] = '░'
	}

	// Calculate filled positions
	// TPI = 1 -> fill all positions 0-9 (full bar)
	// TPI = 0 -> fill position 5 only (center)
	// TPI = -1 -> fill all positions 0-9 (full bar, but bearish color)
	var filledCount int
	switch {
	case tpi >= 0:
		// Positive: fill from center (5) rightward
		// TPI = 0.5 -> 5 filled positions (5, 6, 7, 8, 9)
		// TPI = 1 -> 10 filled positions (0-9)
		filledCount = 5 + int(tpi*5) // 5 to 10
		for i := 5; i < filledCount && i < 10; i++ {
			gauge[i] = '▓'
		}
	default:
		// Negative: fill from center (5) leftward
		// TPI = -0.5 -> 5 filled positions (4, 3, 2, 1, 0)
		// TPI = -1 -> 10 filled positions (0-9)
		filledCount = 5 + int(-tpi*5) // 5 to 10
		lowerBound := 5 - filledCount
		if lowerBound < 0 {
			lowerBound = 0
		}
		for i := 4; i >= lowerBound; i-- {
			gauge[i] = '▓'
		}
	}

	// Render each character with positional color.
	// Position 0 = far left (red), position 5 = center (yellow), position 9 = far right (green).
	var gaugeStr string
	for i, ch := range gauge {
		// Interpolate color based on position:
		// Position 0-2: red tones
		// Position 3-4: red-to-yellow transition
		// Position 5:   yellow (neutral center)
		// Position 6-7: yellow-to-green transition
		// Position 8-9: green tones
		var charStyle lipgloss.Style
		if ch == '░' {
			charStyle = v.theme.Muted()
		} else {
			// Filled character — color by position
			switch {
			case i <= 1:
				charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")) // deep red
			case i <= 3:
				charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8844")) // orange-red
			case i == 4:
				charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA33")) // orange
			case i == 5:
				charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")) // yellow
			case i == 6:
				charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AADD44")) // yellow-green
			case i <= 8:
				charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#66DD55")) // green
			default:
				charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")) // bright green
			}
		}
		gaugeStr += charStyle.Render(string(ch))
	}

	// Render label
	var label string
	if tpi > 0 {
		label = greenStyle.Render(tpiSignal)
	} else {
		label = redStyle.Render(tpiSignal)
	}

	// Format TPI value with sign and color
	var tpiValue string
	switch {
	case tpi > 0:
		tpiValue = greenStyle.Render(fmt.Sprintf("%+.2f", tpi))
	case tpi < 0:
		tpiValue = redStyle.Render(fmt.Sprintf("%+.2f", tpi))
	default:
		tpiValue = fmt.Sprintf("%+.2f", tpi)
	}

	return gaugeStr + " " + tpiValue + " " + label
}

// renderFTEMABadge renders the FTEMA (EMA crossover) as a colored badge.
func (v *View) renderFTEMABadge(signal trenddomain.Signal) string {
	switch signal {
	case trenddomain.Bullish:
		return greenStyle.Render("▲  LONG")
	case trenddomain.Bearish:
		return redStyle.Render("▼ SHORT")
	default:
		return ""
	}
}

// renderBlitzBadge renders a BLITZ score as a colored badge.
func (v *View) renderBlitzBadge(blitzScore int) string {
	switch blitzScore {
	case 1:
		return greenStyle.Render("▲  LONG")
	case -1:
		return redStyle.Render("▼ SHORT")
	default:
		return ""
	}
}

// renderDestinyBadge renders a DESTINY score as a colored badge.
func (v *View) renderDestinyBadge(destinyScore int) string {
	switch destinyScore {
	case 1:
		return greenStyle.Render("▲  LONG")
	case -1:
		return redStyle.Render("▼ SHORT")
	default:
		return ""
	}
}

// renderFlowBadge renders a FLOW score as a colored badge.
func (v *View) renderFlowBadge(flowScore int) string {
	switch flowScore {
	case 1:
		return greenStyle.Render("▲  LONG")
	case -1:
		return redStyle.Render("▼ SHORT")
	default:
		return ""
	}
}

// renderVortexBadge renders a VORTEX score as a colored badge.
func (v *View) renderVortexBadge(vortexScore int) string {
	switch vortexScore {
	case 1:
		return greenStyle.Render("▲  LONG")
	case -1:
		return redStyle.Render("▼ SHORT")
	default:
		return ""
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

// FormatValue formats a float value with specified decimal places.
func FormatValue(value float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, value)
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
