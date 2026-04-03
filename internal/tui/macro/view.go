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

	// Title
	builder.WriteString(v.theme.Title().Render("MACRO DASHBOARD"))
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

	// Title
	builder.WriteString(v.theme.Title().Render("MACRO DASHBOARD"))
	builder.WriteString("\n\n")

	// Panels stacked vertically
	builder.WriteString(v.renderGDPPanel())
	builder.WriteString("\n\n")
	builder.WriteString(v.renderInflationPanel())
	builder.WriteString("\n\n")
	builder.WriteString(v.renderEmploymentPanel())
	builder.WriteString("\n\n")
	builder.WriteString(v.renderRatesPanel())
	builder.WriteString("\n\n")
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
		staleIndicator = v.theme.Warning().Render(" ⚠ Data may be stale") + " "
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

	left := fmt.Sprintf("Last updated: %s", timeSince)
	right := fmt.Sprintf("TTL: %s", v.model.GetTTL())
	help := "r: refresh  |  q: quit  |  1-4: tabs"

	footer := lipgloss.NewStyle().
		Foreground(v.theme.Muted().GetForeground()).
		Render(fmt.Sprintf("%s%s    %s    %s", staleIndicator, left, right, help))

	return footer
}

// renderGDPPanel renders the GDP panel.
func (v *View) renderGDPPanel() string {
	return v.renderPanel("GDP", v.model.gdp.State, v.model.gdp.Error, func() string {
		if v.model.gdp.Data == nil {
			return "No data available"
		}
		d := v.model.gdp.Data
		return fmt.Sprintf(
			" Real GDP:     %s\n"+
				" QoQ:          %s\n"+
				" Per Capita:   %s\n"+
				" Period:       %s",
			d.RealGDP, d.GDPChange, d.PerCapita, d.Period,
		)
	})
}

// renderInflationPanel renders the Inflation panel.
func (v *View) renderInflationPanel() string {
	return v.renderPanel("Inflation", v.model.inflation.State, v.model.inflation.Error, func() string {
		if v.model.inflation.Data == nil {
			return "No data available"
		}
		d := v.model.inflation.Data
		return fmt.Sprintf(
			" CPI:          %s\n"+
				" YoY:          %s\n"+
				" Inflation:    %s\n"+
				" Period:       %s",
			d.CPI, d.CPIYoY, d.Inflation, d.Period,
		)
	})
}

// renderEmploymentPanel renders the Employment panel.
func (v *View) renderEmploymentPanel() string {
	return v.renderPanel("Employment", v.model.employment.State, v.model.employment.Error, func() string {
		if v.model.employment.Data == nil {
			return "No data available"
		}
		d := v.model.employment.Data
		return fmt.Sprintf(
			" Unemployment:  %s\n"+
				" Nonfarm:      %s\n"+
				" Trend:        %s\n"+
				" Period:       %s",
			d.Unemployment, d.Nonfarm, d.Trend, d.Period,
		)
	})
}

// renderRatesPanel renders the Interest Rates panel.
func (v *View) renderRatesPanel() string {
	return v.renderPanel("Interest Rates", v.model.rates.State, v.model.rates.Error, func() string {
		if v.model.rates.Data == nil {
			return "No data available"
		}
		d := v.model.rates.Data
		return fmt.Sprintf(
			" Fed Funds:    %s\n"+
				" Previous:     %s\n"+
				" Last Change:  %s\n"+
				" Period:       %s",
			d.FedFundsRate, d.Previous, d.LastChange, d.Period,
		)
	})
}

// renderYieldsPanel renders the Treasury Yields panel.
func (v *View) renderYieldsPanel() string {
	return v.renderPanel("Treasury Yields", v.model.yields.State, v.model.yields.Error, func() string {
		if v.model.yields.Data == nil {
			return "No data available"
		}
		d := v.model.yields.Data
		return fmt.Sprintf(
			" 2Y:           %s\n"+
				" 5Y:           %s\n"+
				" 10Y:          %s\n"+
				" 30Y:          %s",
			d.Yield2Y, d.Yield5Y, d.Yield10Y, d.Yield30Y,
		)
	})
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
		content = v.theme.Error().Render("Error loading data")
		if err != nil {
			errorText := err.Error()
			if len(errorText) > 30 {
				errorText = errorText[:27] + "..."
			}
			content += "\n" + v.theme.Muted().Render(errorText)
		}
	case PanelLoaded:
		content = contentFunc()
	}

	// Create box with title
	box := v.theme.Box().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.theme.BoxBorder().GetForeground()).
		Render(
			lipgloss.NewStyle().
				Foreground(v.theme.BoxTitle().GetForeground()).
				Bold(true).
				Render(title) + "\n\n" + content,
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
