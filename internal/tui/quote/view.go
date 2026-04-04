// Package quote provides the single ticker quote TUI view.
package quote

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	trenddomain "github.com/owner/finterm/internal/domain/trend"
	"github.com/owner/finterm/internal/tui/components"
)

// Theme defines the interface for theme styling.
// This is defined here to avoid import cycle with the parent tui package.
type Theme interface {
	Title() lipgloss.Style
	Subtitle() lipgloss.Style
	Bullish() lipgloss.Style
	Bearish() lipgloss.Style
	Neutral() lipgloss.Style
	Help() lipgloss.Style
	Error() lipgloss.Style
	Loading() lipgloss.Style
	Muted() lipgloss.Style
	Spinner() lipgloss.Style
	SpinnerText() lipgloss.Style
	Box() lipgloss.Style
	Foreground() lipgloss.Color
	Background() lipgloss.Color
}

// defaultTheme provides a fallback theme when none is set.
type defaultTheme struct{}

func (t *defaultTheme) Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (t *defaultTheme) Subtitle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (t *defaultTheme) Bullish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
}

func (t *defaultTheme) Bearish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
}

func (t *defaultTheme) Neutral() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")).Bold(true)
}

func (t *defaultTheme) Help() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#6272A4"))
}

func (t *defaultTheme) Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
}

func (t *defaultTheme) Loading() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
}

func (t *defaultTheme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
}

func (t *defaultTheme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
}

func (t *defaultTheme) SpinnerText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
}

func (t *defaultTheme) Box() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#44475A")).
		Padding(1)
}

func (t *defaultTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#F8F8F2")
}

func (t *defaultTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282A36")
}

// View handles rendering the quote view.
type View struct {
	model Model
	theme Theme
}

// NewView creates a new quote view.
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
		// Use default styles if no theme is set
		v.theme = &defaultTheme{}
	}

	switch v.model.state {
	case StateIdle:
		return v.renderIdle()
	case StateLoading:
		return v.renderLoading()
	case StateLoaded:
		return v.renderLoaded()
	case StateError:
		return v.renderError()
	default:
		return "Unknown state"
	}
}

// renderIdle renders the idle state with text input.
func (v *View) renderIdle() string {
	var content string

	// Title
	content += v.theme.Title().Render("QUOTE LOOKUP")
	content += "\n\n"

	// Help text
	helpText := "Enter ticker symbol (e.g., AAPL, BTC, ETH)"
	content += v.theme.Help().Render(helpText)
	content += "\n\n"

	// Text input
	inputStyle := lipgloss.NewStyle().
		Foreground(v.theme.Foreground()).
		Background(v.theme.Background()).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5C3FD6"))

	content += inputStyle.Render(v.model.textInput.View())
	content += "\n"

	// Keyboard shortcuts
	content += v.theme.Muted().Render(v.renderShortcuts())

	return content
}

// renderLoading renders the loading state with spinner.
func (v *View) renderLoading() string {
	var content string

	// Title
	content += v.theme.Title().Render("QUOTE LOOKUP")
	content += "\n\n"

	// Spinner
	spinner := components.NewSpinner().
		WithFrameStyle(v.theme.Spinner()).
		WithTextStyle(v.theme.SpinnerText())

	content += spinner.Render()
	content += "\n\n"

	// Loading message
	message := fmt.Sprintf("Fetching quote for %s...", strings.ToUpper(v.model.textInput.Value()))
	content += v.theme.Loading().Render(message)

	return content
}

// renderLoaded renders the loaded quote data.
func (v *View) renderLoaded() string {
	var content string

	// Title
	content += v.theme.Title().Render("QUOTE LOOKUP")
	content += "\n\n"

	// Help text
	helpText := "Enter another ticker symbol (e.g., AAPL, BTC, ETH)"
	content += v.theme.Help().Render(helpText)
	content += "\n\n"

	// Text input - show interactive input for next query
	inputStyle := lipgloss.NewStyle().
		Foreground(v.theme.Foreground()).
		Background(v.theme.Background()).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5C3FD6"))

	content += inputStyle.Render(v.model.textInput.View())
	content += "\n"

	// Keyboard shortcuts
	content += v.theme.Muted().Render(v.renderShortcuts())
	content += "\n\n"

	// Quote data box
	content += v.renderQuoteBox()

	return content
}

// renderError renders the error state.
func (v *View) renderError() string {
	var content string

	// Title
	content += v.theme.Title().Render("QUOTE LOOKUP")
	content += "\n\n"

	// Error message
	content += v.theme.Error().Render("Error loading quote")
	content += "\n\n"

	// Error details
	errorText := fmt.Sprintf("%v", v.model.err)
	content += v.theme.Muted().Render(errorText)
	content += "\n\n"

	// Help text
	helpText := "Press Enter to try again or Esc to clear input"
	content += v.theme.Help().Render(helpText)
	content += "\n\n"

	// Text input - show interactive input for retry
	inputStyle := lipgloss.NewStyle().
		Foreground(v.theme.Foreground()).
		Background(v.theme.Background()).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5C3FD6"))

	content += inputStyle.Render(v.model.textInput.View())
	content += "\n"

	// Keyboard shortcuts
	content += v.theme.Muted().Render(v.renderShortcuts())

	return content
}

// renderQuoteBox renders the quote data in a styled box.
func (v *View) renderQuoteBox() string {
	if v.model.quoteData == nil || v.model.quoteData.Quote == nil {
		return ""
	}

	quote := v.model.quoteData.Quote
	indicators := v.model.quoteData.Indicators

	// Parse quote values
	price := parsePrice(quote.Price)
	change, changePercent := parseChange(quote.Change, quote.ChangePercent)

	// Determine change color
	changeStyle := v.theme.Neutral()
	if change > 0 {
		changeStyle = v.theme.Bullish()
	} else if change < 0 {
		changeStyle = v.theme.Bearish()
	}

	// Build box content
	var boxContent string

	// Symbol row
	boxContent += fmt.Sprintf(" %s", quote.Symbol)
	if quote.Symbol != "" {
		boxContent += "\n\n"
	}

	// Price row
	boxContent += fmt.Sprintf("  Price:       $%s", price)
	boxContent += "\n"

	// Change row
	changeText := fmt.Sprintf("%s (%s)", formatFloat(change, 2), changePercent)
	boxContent += fmt.Sprintf("  Change:      %s", changeStyle.Render(changeText))
	boxContent += "\n"

	// Open
	if quote.Open != "" {
		boxContent += fmt.Sprintf("  Open:        $%s", quote.Open)
		boxContent += "\n"
	}

	// High
	if quote.High != "" {
		boxContent += fmt.Sprintf("  High:        $%s", quote.High)
		boxContent += "\n"
	}

	// Low
	if quote.Low != "" {
		boxContent += fmt.Sprintf("  Low:         $%s", quote.Low)
		boxContent += "\n"
	}

	// Volume
	if quote.Volume != "" {
		vol := formatVolume(quote.Volume)
		boxContent += fmt.Sprintf("  Volume:      %s", vol)
		boxContent += "\n"
	}

	// Previous close
	if quote.PreviousClose != "" {
		boxContent += fmt.Sprintf("  Prev Close:  $%s", quote.PreviousClose)
		boxContent += "\n"
	}

	boxContent += "\n"

	// Indicators section
	if indicators != nil {
		// RSI
		rsiValuation := getRSIValuation(indicators.RSI)
		boxContent += fmt.Sprintf("  RSI(%d):     %s  — %s",
			14, formatFloat(indicators.RSI, 1), rsiValuation)
		boxContent += "\n"

		// EMA Fast
		boxContent += fmt.Sprintf("  EMA(10):     %s", formatFloat(indicators.EMAFast, 2))
		boxContent += "\n"

		// EMA Slow
		boxContent += fmt.Sprintf("  EMA(20):     %s", formatFloat(indicators.EMASlow, 2))
		boxContent += "\n"

		// Trend signal
		trendStyle := v.theme.Neutral()
		trendSymbol := "─"
		switch indicators.Signal {
		case trenddomain.Bullish:
			trendStyle = v.theme.Bullish()
			trendSymbol = "▲"
		case trenddomain.Bearish:
			trendStyle = v.theme.Bearish()
			trendSymbol = "▼"
		}
		boxContent += fmt.Sprintf("  Trend:       %s %s", trendStyle.Render(trendSymbol), trendStyle.Render(indicators.Signal.String()))
		boxContent += "\n"

		// Last trading day
		if quote.LastTradingDay != "" {
			boxContent += "\n"
			boxContent += fmt.Sprintf("  Last updated: %s", quote.LastTradingDay)
		}
	}

	// Render box with border
	box := v.theme.Box().Render(boxContent)

	return box
}

// renderShortcuts renders keyboard shortcut hints.
func (v *View) renderShortcuts() string {
	shortcuts := []string{
		"Enter: Submit",
		"Esc: Clear",
		"↑/↓: History",
		"r: Refresh",
		"1-4: Tabs",
		"q: Quit",
	}

	var result string
	for i, shortcut := range shortcuts {
		if i > 0 {
			result += "  │  "
		}
		result += shortcut
	}

	return result
}

// parsePrice parses a price string, returns "--" if empty.
func parsePrice(s string) string {
	if s == "" {
		return "--"
	}
	return s
}

// parseChange parses the change and change percent, returns values.
func parseChange(changeStr, changePercentStr string) (float64, string) {
	if changeStr == "" {
		return 0, "--"
	}

	change, err := strconv.ParseFloat(changeStr, 64)
	if err != nil {
		change = 0
	}

	if changePercentStr == "" {
		return change, "--"
	}

	// Remove % sign if present
	changePercent := strings.TrimSuffix(changePercentStr, "%")

	return change, changePercent
}

// formatVolume formats a volume number with thousands separators.
func formatVolume(s string) string {
	if s == "" {
		return "--"
	}

	vol, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}

	switch {
	case vol >= 1_000_000_000:
		return fmt.Sprintf("%.2fB", vol/1_000_000_000)
	case vol >= 1_000_000:
		return fmt.Sprintf("%.2fM", vol/1_000_000)
	case vol >= 1_000:
		return fmt.Sprintf("%.2fK", vol/1_000)
	default:
		return fmt.Sprintf("%.0f", vol)
	}
}

// formatFloat formats a float with specified decimal places.
func formatFloat(f float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, f)
}

// getRSIValuation returns the valuation label for an RSI value.
func getRSIValuation(rsi float64) string {
	switch {
	case rsi < 30:
		return "Oversold"
	case rsi < 45:
		return "Undervalued"
	case rsi <= 55:
		return "Fair value"
	case rsi < 70:
		return "Overvalued"
	default:
		return "Overbought"
	}
}

// View returns the view as a string.
func (m Model) View() string {
	return NewView(m).Render()
}
