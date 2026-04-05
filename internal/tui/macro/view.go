// Package macro provides the macroeconomic dashboard TUI view.
package macro

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/owner/finterm/internal/tui/components"
)

// View returns the view as a string. Implements tea.Model.
func (m Model) View() string {
	return NewView(m).Render()
}

// View handles rendering the macro view.
type View struct {
	model Model
	theme Theme
}

// NewView creates a new view for the macro model.
func NewView(model Model) *View {
	return &View{
		model: model,
	}
}

// SetTheme sets the theme for the view.
func (v *View) SetTheme(theme Theme) *View {
	v.theme = theme
	return v
}

// Render renders the view as a string.
func (v *View) Render() string {
	if v.theme == nil {
		v.theme = &defaultTheme{}
	}

	if v.model.IsNarrowTerminal() {
		return v.renderNarrowLayout()
	}
	return v.renderWideLayout()
}

// renderWideLayout renders panels in a 3+2 layout.
func (v *View) renderWideLayout() string {
	// Top row: GDP, Inflation, Employment
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		v.renderGDPPanel(),
		" ",
		v.renderInflationPanel(),
		" ",
		v.renderEmploymentPanel(),
	)

	// Bottom row: Interest Rates, Treasury Yields
	bottomRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		v.renderRatesPanel(),
		" ",
		v.renderYieldsPanel(),
	)

	// Combine rows
	var builder strings.Builder

	// Title with icon
	builder.WriteString(v.theme.Accent().Render("◇") + " " + v.theme.Title().Render("Macro Dashboard"))
	builder.WriteString("\n\n")

	// Panels
	builder.WriteString(topRow)
	builder.WriteString("\n\n")
	builder.WriteString(bottomRow)

	// Footer
	builder.WriteString("\n\n")
	builder.WriteString(v.renderFooter())

	return builder.String()
}

// renderNarrowLayout renders panels in a vertical stack.
func (v *View) renderNarrowLayout() string {
	var builder strings.Builder

	// Title with icon
	builder.WriteString(v.theme.Accent().Render("◇") + " " + v.theme.Title().Render("Macro Dashboard"))
	builder.WriteString("\n\n")

	// Panels stacked vertically with single newline between them
	builder.WriteString(v.renderGDPPanel())
	builder.WriteString("\n")
	builder.WriteString(v.renderInflationPanel())
	builder.WriteString("\n")
	builder.WriteString(v.renderEmploymentPanel())
	builder.WriteString("\n")
	builder.WriteString(v.renderRatesPanel())
	builder.WriteString("\n")
	builder.WriteString(v.renderYieldsPanel())

	// Footer
	builder.WriteString("\n\n")
	builder.WriteString(v.renderFooter())

	return builder.String()
}

// renderFooter renders the footer with update info and key bindings.
func (v *View) renderFooter() string {
	staleIndicator := ""
	if v.model.IsDataStale() {
		staleIndicator = v.theme.Warning().Render("⚠ Stale") + " "
	}

	timeSince := "just now"
	if !v.model.lastUpdate.IsZero() {
		elapsed := time.Since(v.model.lastUpdate)
		switch {
		case elapsed < time.Minute:
			timeSince = "just now"
		case elapsed < time.Hour:
			timeSince = fmt.Sprintf("%d min ago", int(elapsed.Minutes()))
		default:
			timeSince = fmt.Sprintf("%d hours ago", int(elapsed.Hours()))
		}
	}

	// Build footer parts
	updateInfo := fmt.Sprintf("Updated %s  ·  TTL %s", timeSince, v.model.GetTTL())
	keyHints := v.renderKeyHints()

	return v.theme.Muted().Render(staleIndicator + updateInfo + "  ·  " + keyHints)
}

// renderKeyHints renders keyboard shortcut hints.
func (v *View) renderKeyHints() string {
	shortcuts := []struct {
		key  string
		desc string
	}{
		{"r", "refresh"},
		{"1-4", "tabs"},
		{"q", "quit"},
	}

	var parts []string
	for _, sc := range shortcuts {
		parts = append(parts, v.theme.Accent().Render(sc.key)+" "+v.theme.Muted().Render(sc.desc))
	}

	return strings.Join(parts, "  ·  ")
}

// panelIcons maps panel names to their emoji icons.
var panelIcons = map[string]string{
	"GDP":             "📊",
	"Inflation":       "📈",
	"Employment":      "👷",
	"Interest Rates":  "🏦",
	"Treasury Yields": "📋",
}

// renderGDPPanel renders the GDP panel.
func (v *View) renderGDPPanel() string {
	return v.renderPanel("GDP", v.model.gdp.State, v.model.gdp.Error, func() string {
		if v.model.gdp.Data == nil {
			return "No data available"
		}
		d := v.model.gdp.Data
		var builder strings.Builder
		v.formatRow(&builder, "Real GDP", d.RealGDP)
		v.formatRow(&builder, "QoQ", d.GDPChange)
		v.formatRow(&builder, "Per Capita", d.PerCapita)
		v.formatRowLast(&builder, "Period", d.Period)
		return builder.String()
	})
}

// renderInflationPanel renders the Inflation panel.
func (v *View) renderInflationPanel() string {
	return v.renderPanel("Inflation", v.model.inflation.State, v.model.inflation.Error, func() string {
		if v.model.inflation.Data == nil {
			return "No data available"
		}
		d := v.model.inflation.Data
		var builder strings.Builder
		v.formatRow(&builder, "CPI", d.CPI)
		v.formatRow(&builder, "YoY", d.CPIYoY)
		v.formatRow(&builder, "Inflation", d.Inflation)
		v.formatRowLast(&builder, "Period", d.Period)
		return builder.String()
	})
}

// renderEmploymentPanel renders the Employment panel.
func (v *View) renderEmploymentPanel() string {
	return v.renderPanel("Employment", v.model.employment.State, v.model.employment.Error, func() string {
		if v.model.employment.Data == nil {
			return "No data available"
		}
		d := v.model.employment.Data
		var builder strings.Builder
		v.formatRow(&builder, "Unemployment", d.Unemployment)
		v.formatRow(&builder, "Nonfarm", d.Nonfarm)
		v.formatRow(&builder, "Trend", d.Trend)
		v.formatRowLast(&builder, "Period", d.Period)
		return builder.String()
	})
}

// renderRatesPanel renders the Interest Rates panel.
func (v *View) renderRatesPanel() string {
	return v.renderPanel("Interest Rates", v.model.rates.State, v.model.rates.Error, func() string {
		if v.model.rates.Data == nil {
			return "No data available"
		}
		d := v.model.rates.Data
		var builder strings.Builder
		v.formatRow(&builder, "Fed Funds", d.FedFundsRate)
		v.formatRow(&builder, "Previous", d.Previous)
		v.formatRow(&builder, "Last Change", d.LastChange)
		v.formatRowLast(&builder, "Period", d.Period)
		return builder.String()
	})
}

// renderYieldsPanel renders the Treasury Yields panel.
func (v *View) renderYieldsPanel() string {
	return v.renderPanel("Treasury Yields", v.model.yields.State, v.model.yields.Error, func() string {
		if v.model.yields.Data == nil {
			return "No data available"
		}
		d := v.model.yields.Data
		var builder strings.Builder
		v.formatRow(&builder, "2Y", d.Yield2Y)
		v.formatRow(&builder, "5Y", d.Yield5Y)
		v.formatRow(&builder, "10Y", d.Yield10Y)
		v.formatRowLast(&builder, "30Y", d.Yield30Y)
		return builder.String()
	})
}

// formatRow formats a label-value pair with consistent formatting.
func (v *View) formatRow(builder *strings.Builder, label, value string) {
	if value != "" {
		builder.WriteString(v.theme.Muted().Render(fmt.Sprintf("  %-14s", label+":")) + value + "\n")
	}
}

// formatRowLast formats the last label-value pair (no trailing newline).
func (v *View) formatRowLast(builder *strings.Builder, label, value string) {
	if value != "" {
		builder.WriteString(v.theme.Muted().Render(fmt.Sprintf("  %-14s", label+":")) + value)
	}
}

// renderPanel renders a single panel with title, content, and border.
func (v *View) renderPanel(title string, state PanelState, err error, contentFunc func() string) string {
	var content string

	switch state {
	case PanelLoading:
		spinner := components.NewSpinner().
			WithFrameStyle(v.theme.Spinner()).
			WithTextStyle(v.theme.SpinnerText())
		content = spinner.Render()
	case PanelError:
		content = v.theme.Error().Render("✗ Error loading data")
		if err != nil {
			errorText := err.Error()
			if len(errorText) > 30 {
				errorText = errorText[:27] + "…"
			}
			content += "\n" + v.theme.Muted().Render(errorText)
		}
	case PanelLoaded:
		content = contentFunc()
	}

	// Calculate panel width based on terminal width
	// For wide layout (3 columns): divide width by 3 with gap
	// For narrow layout: use full width
	panelWidth := v.model.GetWidth() - 4 // Account for borders and padding
	if !v.model.IsNarrowTerminal() {
		panelWidth = (panelWidth - 4) / 3 // 3 columns with 2 gaps
	}
	if panelWidth < 25 {
		panelWidth = 25 // Minimum width
	}

	// Get panel icon
	icon := panelIcons[title]

	// Create box with title and calculated width
	box := v.theme.Box().
		Width(panelWidth).
		Padding(1, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.theme.BoxBorder().GetForeground()).
		Render(
			lipgloss.NewStyle().
				Foreground(v.theme.BoxTitle().GetForeground()).
				Bold(true).
				Render(icon+" "+title) + "\n" +
				v.theme.Divider().Render(strings.Repeat("─", panelWidth-4)) + "\n\n" + content,
		)

	return box
}

// Theme defines the interface for theme styling.
type Theme interface {
	Title() lipgloss.Style
	Muted() lipgloss.Style
	Warning() lipgloss.Style
	Error() lipgloss.Style
	Box() lipgloss.Style
	BoxBorder() lipgloss.Style
	BoxTitle() lipgloss.Style
	Spinner() lipgloss.Style
	SpinnerText() lipgloss.Style
	Accent() lipgloss.Style
	Divider() lipgloss.Style
}

// defaultTheme provides a fallback theme when no theme is set.
type defaultTheme struct{}

func (t *defaultTheme) Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (t *defaultTheme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
}

func (t *defaultTheme) Warning() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")).Bold(true)
}

func (t *defaultTheme) Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
}

func (t *defaultTheme) Box() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#44475A"))
}

func (t *defaultTheme) BoxBorder() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
}

func (t *defaultTheme) BoxTitle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true)
}

func (t *defaultTheme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
}

func (t *defaultTheme) SpinnerText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
}

func (t *defaultTheme) Accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
}

func (t *defaultTheme) Divider() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
}
