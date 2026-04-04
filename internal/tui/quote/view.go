// Package quote provides the single ticker quote TUI view.
package quote

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/owner/finterm/internal/alphavantage"
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

	// Build box content
	boxContent := v.buildQuoteContent(quote, indicators)

	// Calculate and render box with width
	box := v.renderBoxWithWidth(boxContent)

	// Center the box within terminal width
	return v.centerBox(box)
}

// buildQuoteContent builds the content string for the quote box.
func (v *View) buildQuoteContent(quote *alphavantage.GlobalQuote, indicators *trenddomain.Result) string {
	var content strings.Builder

	// Parse quote values
	price := parsePrice(quote.Price)
	change, changePercent := parseChange(quote.Change, quote.ChangePercent)

	// Determine change color
	changeStyle := v.getChangeStyle(change)

	// Symbol row
	content.WriteString(" " + quote.Symbol)
	if quote.Symbol != "" {
		content.WriteString("\n\n")
	}

	// Basic quote fields
	v.writeQuoteField(&content, "  Price:       ", "$"+price, "")
	v.writeChangeField(&content, change, changePercent, changeStyle)
	v.writeQuoteField(&content, "  Open:        ", "$"+quote.Open, quote.Open)
	v.writeQuoteField(&content, "  High:        ", "$"+quote.High, quote.High)
	v.writeQuoteField(&content, "  Low:         ", "$"+quote.Low, quote.Low)
	v.writeQuoteField(&content, "  Volume:      ", formatVolume(quote.Volume), quote.Volume)
	v.writeQuoteField(&content, "  Prev Close:  ", "$"+quote.PreviousClose, quote.PreviousClose)

	// Indicators section
	if indicators != nil {
		content.WriteString("\n\n")
		v.writeIndicators(&content, indicators, quote.LastTradingDay)
	}

	return content.String()
}

// getChangeStyle returns the appropriate style based on price change direction.
func (v *View) getChangeStyle(change float64) lipgloss.Style {
	switch {
	case change > 0:
		return v.theme.Bullish()
	case change < 0:
		return v.theme.Bearish()
	default:
		return v.theme.Neutral()
	}
}

// writeQuoteField writes a quote field line if the value is non-empty.
func (v *View) writeQuoteField(content *strings.Builder, label, value, condition string) {
	if condition != "" {
		content.WriteString(label + value + "\n")
	}
}

// writeChangeField writes the change field with appropriate styling.
func (v *View) writeChangeField(content *strings.Builder, change float64, changePercent string, style lipgloss.Style) {
	changeText := fmt.Sprintf("%s (%s)", formatFloat(change, 2), changePercent)
	fmt.Fprintf(content, "  Change:      %s\n", style.Render(changeText))
}

// writeIndicators writes the technical indicators section.
func (v *View) writeIndicators(content *strings.Builder, indicators *trenddomain.Result, lastTradingDay string) {
	rsiValuation := getRSIValuation(indicators.RSI)
	valuationStyle := v.getValuationStyle(rsiValuation)
	fmt.Fprintf(content, "  RSI(%d):     %s  — %s\n", 14, formatFloat(indicators.RSI, 1), valuationStyle.Render(rsiValuation))
	fmt.Fprintf(content, "  EMA(10):     %s\n", formatFloat(indicators.EMAFast, 2))
	fmt.Fprintf(content, "  EMA(20):     %s\n", formatFloat(indicators.EMASlow, 2))
	v.writeTrendSignal(content, indicators)

	if lastTradingDay != "" {
		content.WriteString("\n")
		fmt.Fprintf(content, "  Last updated: %s", lastTradingDay)
	}
}

// writeTrendSignal writes the trend signal with appropriate styling.
func (v *View) writeTrendSignal(content *strings.Builder, indicators *trenddomain.Result) {
	trendStyle, trendSymbol := v.getTrendStyleAndSymbol(indicators.Signal)
	fmt.Fprintf(content, "  Trend:       %s %s\n", trendStyle.Render(trendSymbol), trendStyle.Render(indicators.Signal.String()))
}

// getTrendStyleAndSymbol returns the style and symbol for a trend signal.
func (v *View) getTrendStyleAndSymbol(signal trenddomain.Signal) (lipgloss.Style, string) {
	switch signal {
	case trenddomain.Bullish:
		return v.theme.Bullish(), "▲"
	case trenddomain.Bearish:
		return v.theme.Bearish(), "▼"
	default:
		return v.theme.Neutral(), "─"
	}
}

// getValuationStyle returns the style for an RSI valuation label.
func (v *View) getValuationStyle(valuation string) lipgloss.Style {
	switch valuation {
	case "Oversold", "Undervalued":
		return v.theme.Bullish()
	case "Overbought", "Overvalued":
		return v.theme.Bearish()
	default:
		return v.theme.Neutral()
	}
}

// renderBoxWithWidth renders the box with calculated width based on terminal size.
func (v *View) renderBoxWithWidth(content string) string {
	boxWidth := v.calculateBoxWidth()

	return v.theme.Box().
		Width(boxWidth).
		Render(content)
}

// calculateBoxWidth calculates the appropriate box width based on terminal size.
func (v *View) calculateBoxWidth() int {
	maxBoxWidth := 70
	terminalWidth := v.model.GetWidth()
	boxWidth := maxBoxWidth

	if terminalWidth < maxBoxWidth+4 {
		boxWidth = terminalWidth - 4
	}
	if boxWidth < 30 {
		boxWidth = 30
	}
	return boxWidth
}

// centerBox centers the rendered box within the terminal width.
func (v *View) centerBox(box string) string {
	boxLines := strings.Split(box, "\n")
	terminalWidth := v.model.GetWidth()

	var centeredBox strings.Builder
	for _, line := range boxLines {
		lineWidth := lipgloss.Width(line)
		if lineWidth < terminalWidth {
			leftPadding := (terminalWidth - lineWidth) / 2
			centeredBox.WriteString(strings.Repeat(" ", leftPadding) + line)
		} else {
			centeredBox.WriteString(line)
		}
		centeredBox.WriteString("\n")
	}

	return strings.TrimSuffix(centeredBox.String(), "\n")
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

	// Keep % sign for display
	return change, changePercentStr
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
