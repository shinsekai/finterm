// Package commodities provides the commodities dashboard TUI view.
package commodities

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// Theme defines the interface for theme styling.
// Defined here to avoid import cycle with parent tui package.
type Theme interface {
	Title() lipgloss.Style
	TableRow() lipgloss.Style
	TableRowAlt() lipgloss.Style
	TableHeader() lipgloss.Style
	TableEmpty() lipgloss.Style
	Bullish() lipgloss.Style
	Bearish() lipgloss.Style
	Neutral() lipgloss.Style
	Error() lipgloss.Style
	Muted() lipgloss.Style
	Spinner() lipgloss.Style
	Accent() lipgloss.Style
	Divider() lipgloss.Style
	Foreground() lipgloss.Color
	Background() lipgloss.Color
}

// ThemeAdapter adapts our Theme to components.Theme.
type ThemeAdapter struct {
	theme Theme
}

// Bullish returns the bullish style.
func (ta ThemeAdapter) Bullish() lipgloss.Style {
	return ta.theme.Bullish()
}

// Bearish returns the bearish style.
func (ta ThemeAdapter) Bearish() lipgloss.Style {
	return ta.theme.Bearish()
}

// Neutral returns the neutral style.
func (ta ThemeAdapter) Neutral() lipgloss.Style {
	return ta.theme.Neutral()
}

// View handles rendering of the commodities dashboard.
type View struct {
	model *Model
	theme Theme
	table *components.Table
}

// NewView creates a new view for the commodities model.
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

// Render renders the commodities view as a string.
func (v *View) Render() string {
	if v.theme == nil {
		v.theme = &defaultTheme{}
	}

	var builder strings.Builder

	// Render title
	builder.WriteString(v.renderTitle())
	builder.WriteString("\n\n")

	// Render empty state or table
	if v.model.IsEmpty() {
		builder.WriteString(v.renderEmptyState())
		builder.WriteString("\n\n")
		builder.WriteString(v.renderFooter())
		return builder.String()
	}

	// Render table
	builder.WriteString(v.renderTable())
	builder.WriteString("\n\n")

	// Render footer
	builder.WriteString(v.renderFooter())

	return builder.String()
}

// renderTitle renders the view title with loading progress.
func (v *View) renderTitle() string {
	symbolCount := len(v.model.GetRows())
	loadedCount := v.model.GetLoadedCount()

	// Show progress chip when watchlist is non-empty
	var chip string
	if symbolCount > 0 {
		if loadedCount < symbolCount {
			chip = v.theme.Muted().Render(fmt.Sprintf("  loaded %d/%d", loadedCount, symbolCount))
		} else {
			elapsed := time.Since(v.model.lastUpdate)
			chip = v.theme.Muted().Render(fmt.Sprintf("  all loaded · %s ago", formatDuration(elapsed)))
		}
	}

	return v.theme.Accent().Render("◇") + " " + v.theme.Title().Render("Commodities") + chip
}

// renderEmptyState renders a placeholder when no commodities are configured.
func (v *View) renderEmptyState() string {
	emptyStyle := v.theme.TableEmpty().
		Width(v.model.GetWidth()).
		MaxHeight(v.model.GetHeight()-6).
		Align(lipgloss.Center, lipgloss.Center)

	return emptyStyle.Render("no commodities configured")
}

// renderTable renders the commodities data table.
func (v *View) renderTable() string {
	if v.model.IsEmpty() {
		return ""
	}

	// Configure columns
	columns := []components.Column{
		{Title: "SYMBOL", Width: 12, Alignment: components.AlignLeft, HeaderStyle: v.theme.TableHeader()},
		{Title: "NAME", Width: 16, Alignment: components.AlignLeft, HeaderStyle: v.theme.TableHeader()},
		{Title: "SPARKLINE", Width: 30, Alignment: components.AlignLeft, HeaderStyle: v.theme.TableHeader()},
		{Title: "VALUE", Width: 16, Alignment: components.AlignRight, HeaderStyle: v.theme.TableHeader()},
		{Title: "CHANGE", Width: 10, Alignment: components.AlignRight, HeaderStyle: v.theme.TableHeader()},
		{Title: "PERIOD", Width: 12, Alignment: components.AlignRight, HeaderStyle: v.theme.TableHeader()},
	}

	// Build rows
	var tableRows []components.Row
	for i, rowData := range v.model.GetRows() {
		row := components.NewRow(v.buildRow(rowData))
		// Apply alternating row styles
		if i%2 == 0 {
			row.Style = v.theme.TableRow()
		} else {
			row.Style = v.theme.TableRowAlt()
		}
		tableRows = append(tableRows, row)
	}

	// Configure table
	v.table.WithColumns(columns)
	v.table.WithRows(tableRows)
	v.table.WithMaxWidth(v.model.GetWidth())

	return v.table.Render()
}

// buildRow builds a table row for a single commodity.
func (v *View) buildRow(row RowData) []string {
	switch row.State {
	case StateLoading:
		return []string{
			row.Symbol,
			row.Name,
			v.theme.Spinner().Render("…"),
			v.theme.Muted().Render("loading"),
			v.theme.Muted().Render("-"),
			v.theme.Muted().Render("-"),
		}

	case StateError:
		return []string{
			row.Symbol,
			row.Name,
			v.theme.Error().Render("✗"),
			v.theme.Error().Render("error"),
			v.theme.Muted().Render("-"),
			v.theme.Muted().Render("-"),
		}

	case StateLoaded, StateCached:
		if row.Series == nil || len(row.Series.Data) == 0 {
			return []string{
				row.Symbol,
				row.Name,
				v.theme.Muted().Render("…"),
				v.theme.Muted().Render("no data"),
				v.theme.Muted().Render("-"),
				v.theme.Muted().Render("-"),
			}
		}

		data := row.Series.Data
		latest := data[len(data)-1]

		// Sparkline: 30 most recent data points
		sparklineWidth := 30
		values := extractValuesForSparkline(data, sparklineWidth)
		sparkline := components.RenderSparkline(values, sparklineWidth, ThemeAdapter{theme: v.theme})

		// Latest value with unit
		valueStr := formatValue(latest.Value, row.Series.Unit)

		// Change calculation (compare with previous period)
		changeStr, changeStyle := v.calculateChange(data)

		// Period (last date in series) with fallback chip if applicable
		periodStr := latest.Date.Format("2006-01-02")
		if row.ActualInterval != "" && row.ActualInterval != v.model.GetInterval() {
			fallbackChip := v.theme.Muted().Render(fmt.Sprintf(" → %s", row.ActualInterval))
			periodStr += fallbackChip
		}

		return []string{
			row.Symbol,
			row.Name,
			sparkline,
			valueStr,
			changeStyle.Render(changeStr),
			periodStr,
		}

	default:
		return []string{
			row.Symbol,
			row.Name,
			v.theme.Muted().Render("?"),
			v.theme.Muted().Render("unknown"),
			v.theme.Muted().Render("-"),
			v.theme.Muted().Render("-"),
		}
	}
}

// calculateChange calculates the change from the previous period.
// Returns the formatted change string and appropriate style.
func (v *View) calculateChange(data []alphavantage.CommodityDataPoint) (string, lipgloss.Style) {
	if len(data) < 2 {
		return "-", v.theme.Muted()
	}

	latest := data[len(data)-1]
	previous := data[len(data)-2]

	if previous.Value == 0 {
		return "-", v.theme.Muted()
	}

	absChange := latest.Value - previous.Value
	percentChange := (absChange / previous.Value) * 100

	var style lipgloss.Style
	switch {
	case absChange > 0:
		style = v.theme.Bullish()
	case absChange < 0:
		style = v.theme.Bearish()
	default:
		style = v.theme.Neutral()
	}

	if percentChange == 0 {
		return "0.00%", style
	}

	sign := "+"
	if absChange < 0 {
		sign = ""
	}

	return fmt.Sprintf("%s%.2f%%", sign, percentChange), style
}

// formatValue formats a price value with appropriate unit.
func formatValue(value float64, unit string) string {
	return fmt.Sprintf("%.2f %s", value, unit)
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// extractValuesForSparkline extracts the last N values for sparkline rendering.
// Returns values in chronological order (oldest first) for proper sparkline direction.
func extractValuesForSparkline(data []alphavantage.CommodityDataPoint, n int) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	// Take last n points or all if fewer
	start := len(data) - n
	if start < 0 {
		start = 0
	}

	values := make([]float64, 0, len(data)-start)
	for i := start; i < len(data); i++ {
		values = append(values, data[i].Value)
	}

	return values
}

// renderFooter renders the footer with keyboard hints.
func (v *View) renderFooter() string {
	hints := []string{
		"↑/k: up",
		"↓/j: down",
		"r: refresh",
		"q: quit",
	}
	return v.theme.Muted().Render(strings.Join(hints, "  ·  "))
}

// defaultTheme provides fallback styling when no theme is set.
type defaultTheme struct{}

func (dt *defaultTheme) Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (dt *defaultTheme) TableRow() lipgloss.Style {
	return lipgloss.NewStyle()
}

func (dt *defaultTheme) TableRowAlt() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
}

func (dt *defaultTheme) TableHeader() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8BE9FD"))
}

func (dt *defaultTheme) TableEmpty() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
}

func (dt *defaultTheme) Bullish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
}

func (dt *defaultTheme) Bearish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
}

func (dt *defaultTheme) Neutral() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
}

func (dt *defaultTheme) Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
}

func (dt *defaultTheme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
}

func (dt *defaultTheme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))
}

func (dt *defaultTheme) Accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9"))
}

func (dt *defaultTheme) Divider() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
}

func (dt *defaultTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#F8F8F2")
}

func (dt *defaultTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282A36")
}
