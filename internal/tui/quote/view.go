// Package quote provides the single ticker quote TUI view.
package quote

import (
	"fmt"
	"math"
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
	Card() lipgloss.Style
	CardHeader() lipgloss.Style
	Accent() lipgloss.Style
	Divider() lipgloss.Style
	Primary() lipgloss.Color
	Border() lipgloss.Color
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

func (t *defaultTheme) Card() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#44475A")).
		Padding(1, 2)
}

func (t *defaultTheme) CardHeader() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BD93F9")).
		Bold(true).
		MarginBottom(1)
}

func (t *defaultTheme) Accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
}

func (t *defaultTheme) Divider() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A"))
}

func (t *defaultTheme) Primary() lipgloss.Color {
	return lipgloss.Color("#7D56F4")
}

func (t *defaultTheme) Border() lipgloss.Color {
	return lipgloss.Color("#44475A")
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

	// Title with icon
	content += v.theme.Accent().Render("◈") + " " + v.theme.Title().Render("Quote Lookup")
	content += "\n\n"

	// Input label
	label := v.theme.Accent().Render("Symbol ") + v.theme.Muted().Render("(e.g. AAPL, BTC, ETH)")
	content += label + "\n"

	// Text input with rounded border using primary color
	inputStyle := lipgloss.NewStyle().
		Foreground(v.theme.Foreground()).
		Background(v.theme.Background()).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.theme.Primary())

	content += inputStyle.Render(v.model.textInput.View())
	content += "\n\n"

	// Keyboard shortcuts
	content += v.renderShortcuts()

	return content
}

// renderLoading renders the loading state with spinner.
func (v *View) renderLoading() string {
	var content string

	// Title with icon
	content += v.theme.Accent().Render("◈") + " " + v.theme.Title().Render("Quote Lookup")
	content += "\n\n"

	// Spinner
	spinner := components.NewSpinner().
		WithFrameStyle(v.theme.Spinner()).
		WithTextStyle(v.theme.SpinnerText())

	content += spinner.Render()
	content += "\n\n"

	// Loading message
	message := fmt.Sprintf("⟳ Fetching %s…", strings.ToUpper(v.model.textInput.Value()))
	content += v.theme.Loading().Render(message)

	return content
}

// renderLoaded renders the loaded quote data.
func (v *View) renderLoaded() string {
	var content string

	// Title with icon
	content += v.theme.Accent().Render("◈") + " " + v.theme.Title().Render("Quote Lookup")
	content += "\n\n"

	// Input label
	label := v.theme.Accent().Render("Symbol ") + v.theme.Muted().Render("(e.g. AAPL, BTC, ETH)")
	content += label + "\n"

	// Text input with rounded border using primary color
	inputStyle := lipgloss.NewStyle().
		Foreground(v.theme.Foreground()).
		Background(v.theme.Background()).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.theme.Primary())

	content += inputStyle.Render(v.model.textInput.View())
	content += "\n\n"

	// Keyboard shortcuts
	content += v.renderShortcuts()
	content += "\n\n"

	// Quote data cards
	content += v.renderQuoteCards()

	return content
}

// renderError renders the error state.
func (v *View) renderError() string {
	var content string

	// Title with icon
	content += v.theme.Accent().Render("◈") + " " + v.theme.Title().Render("Quote Lookup")
	content += "\n\n"

	// Error card with red border
	errorCardStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF5555")).
		Padding(1, 2)

	var cardContent strings.Builder
	cardContent.WriteString(v.theme.Error().Render("✗ Error loading quote") + "\n\n")
	cardContent.WriteString(v.theme.Muted().Render(fmt.Sprintf("%v", v.model.err)))

	content += errorCardStyle.Render(cardContent.String())
	content += "\n\n"

	// Help text
	helpText := "Press Enter to try again or Esc to clear input"
	content += v.theme.Help().Render(helpText)
	content += "\n\n"

	// Input label
	label := v.theme.Accent().Render("Symbol ") + v.theme.Muted().Render("(e.g. AAPL, BTC, ETH)")
	content += label + "\n"

	// Text input with rounded border using primary color
	inputStyle := lipgloss.NewStyle().
		Foreground(v.theme.Foreground()).
		Background(v.theme.Background()).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.theme.Primary())

	content += inputStyle.Render(v.model.textInput.View())
	content += "\n\n"

	// Keyboard shortcuts
	content += v.renderShortcuts()

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

// renderQuoteCards renders the quote data in two card sections.
func (v *View) renderQuoteCards() string {
	if v.model.quoteData == nil || v.model.quoteData.Quote == nil {
		return ""
	}

	quote := v.model.quoteData.Quote
	indicators := v.model.quoteData.Indicators

	var cards []string

	// Price Card
	priceCard := v.renderPriceCard(quote)
	cards = append(cards, priceCard)

	// Indicators Card
	if indicators != nil {
		indicatorsCard := v.renderIndicatorsCard(indicators, quote.LastTradingDay)
		cards = append(cards, indicatorsCard)
	}

	return strings.Join(cards, "\n\n")
}

// renderPriceCard renders the price data in a card.
func (v *View) renderPriceCard(quote *alphavantage.GlobalQuote) string {
	var content strings.Builder

	// Parse quote values
	price, _ := strconv.ParseFloat(quote.Price, 64)
	change, changePercent := parseChange(quote.Change, quote.ChangePercent)

	// Symbol header
	content.WriteString(v.theme.Title().Render(quote.Symbol))
	content.WriteString("\n")

	// Divider
	boxWidth := v.calculateBoxWidth()
	content.WriteString(v.theme.Divider().Render(strings.Repeat("─", boxWidth)))
	content.WriteString("\n\n")

	// Price prominently displayed
	priceStyle := lipgloss.NewStyle().
		Foreground(v.theme.Foreground()).
		Bold(true).
		MarginBottom(1)
	content.WriteString(priceStyle.Render("$" + humanizeFloat(price, 2)))
	content.WriteString("\n")

	// Change with directional icon
	changeStyle := v.getChangeStyle(change)
	icon := "─"
	if change > 0 {
		icon = "▲"
	} else if change < 0 {
		icon = "▼"
	}
	changeText := fmt.Sprintf("%s %s$%s (%s)", icon,
		func() string {
			if change < 0 {
				return ""
			}
			return "+"
		}(),
		humanizeFloat(math.Abs(change), 2),
		formatPercent(changePercent))
	content.WriteString(changeStyle.Render(changeText))
	content.WriteString("\n\n")

	// OHLCV data rows with muted labels
	v.writePriceField(&content, "Open", "$"+formatPrice(quote.Open), quote.Open)
	v.writePriceField(&content, "High", "$"+formatPrice(quote.High), quote.High)
	v.writePriceField(&content, "Low", "$"+formatPrice(quote.Low), quote.Low)
	v.writePriceField(&content, "Volume", formatVolume(quote.Volume, quote.Symbol, price), quote.Volume)
	v.writePriceField(&content, "Prev Close", "$"+formatPrice(quote.PreviousClose), quote.PreviousClose)

	return v.theme.Card().Render(content.String())
}

// renderIndicatorsCard renders the technical indicators in a card.
func (v *View) renderIndicatorsCard(indicators *trenddomain.Result, lastTradingDay string) string {
	var content strings.Builder

	// Header
	content.WriteString(v.theme.CardHeader().Render("Technical Indicators"))
	content.WriteString("\n")

	// Divider
	boxWidth := v.calculateBoxWidth()
	content.WriteString(v.theme.Divider().Render(strings.Repeat("─", boxWidth)))
	content.WriteString("\n\n")

	// RSI with progress bar
	content.WriteString(v.renderRSIBar(indicators.RSI))
	content.WriteString("\n\n")

	// EMA values
	content.WriteString(fmt.Sprintf("%-14s$%s\n", v.theme.Muted().Render("EMA(10):"), humanizeFloat(indicators.EMAFast, 2)))
	content.WriteString(fmt.Sprintf("%-14s$%s\n", v.theme.Muted().Render("EMA(20):"), humanizeFloat(indicators.EMASlow, 2)))

	// Trend signal with icon
	trendStyle, trendSymbol := v.getTrendStyleAndSymbol(indicators.Signal)
	content.WriteString(fmt.Sprintf("%-14s%s %s\n",
		v.theme.Muted().Render("Trend:"),
		trendStyle.Render(trendSymbol),
		trendStyle.Render(indicators.Signal.String())))

	// Last updated
	if lastTradingDay != "" {
		content.WriteString("\n")
		content.WriteString(v.theme.Muted().Render("Last updated: " + lastTradingDay))
	}

	return v.theme.Card().Render(content.String())
}

// renderRSIBar renders an RSI progress bar with zone markers.
func (v *View) renderRSIBar(rsi float64) string {
	var content strings.Builder

	// RSI label
	rsiValuation := getRSIValuation(rsi)
	valuationStyle := v.getValuationStyle(rsiValuation)
	content.WriteString(fmt.Sprintf("%-14s%.1f  — %s\n",
		v.theme.Muted().Render("RSI(14):"),
		rsi,
		valuationStyle.Render(rsiValuation)))

	// Progress bar
	barWidth := 40
	filledWidth := int(rsi / 100 * float64(barWidth))
	if filledWidth > barWidth {
		filledWidth = barWidth
	}

	// Determine bar color
	var barStyle lipgloss.Style
	switch {
	case rsi < 30:
		barStyle = v.theme.Bullish()
	case rsi > 70:
		barStyle = v.theme.Bearish()
	default:
		barStyle = v.theme.Accent()
	}

	// Build bar: filled portion + empty portion
	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", barWidth-filledWidth)
	bar := barStyle.Render(filled) + v.theme.Muted().Render(empty)

	content.WriteString("  " + bar + "\n")

	// Zone markers
	content.WriteString("  " + v.theme.Muted().Render("0       30       70      100"))

	return content.String()
}

// writePriceField writes a price field line with muted label.
func (v *View) writePriceField(content *strings.Builder, label, value, condition string) {
	if condition != "" {
		content.WriteString(fmt.Sprintf("%-14s%s\n", v.theme.Muted().Render(label+":"), value))
	}
}

// buildQuoteContent builds the content string for the quote box.
func (v *View) buildQuoteContent(quote *alphavantage.GlobalQuote, indicators *trenddomain.Result) string {
	var content strings.Builder

	// Parse quote values
	price, _ := strconv.ParseFloat(quote.Price, 64)
	change, changePercent := parseChange(quote.Change, quote.ChangePercent)

	// Determine change color
	changeStyle := v.getChangeStyle(change)

	// Symbol row
	content.WriteString(" " + quote.Symbol)
	if quote.Symbol != "" {
		content.WriteString("\n\n")
	}

	// Basic quote fields
	v.writeQuoteField(&content, "  Price:       ", "$"+formatPrice(quote.Price), "")
	v.writeChangeField(&content, change, changePercent, changeStyle)
	v.writeQuoteField(&content, "  Open:        ", "$"+formatPrice(quote.Open), quote.Open)
	v.writeQuoteField(&content, "  High:        ", "$"+formatPrice(quote.High), quote.High)
	v.writeQuoteField(&content, "  Low:         ", "$"+formatPrice(quote.Low), quote.Low)
	v.writeQuoteField(&content, "  Volume:      ", formatVolume(quote.Volume, quote.Symbol, price), quote.Volume)
	v.writeQuoteField(&content, "  Prev Close:  ", "$"+formatPrice(quote.PreviousClose), quote.PreviousClose)

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
	// Format: -$14.26 (-0.02%)
	sign := ""
	if change < 0 {
		sign = "-"
	}
	changeText := fmt.Sprintf("%s$%s (%s)", sign, humanizeFloat(math.Abs(change), 2), formatPercent(changePercent))
	fmt.Fprintf(content, "  Change:      %s\n", style.Render(changeText))
}

// writeIndicators writes the technical indicators section.
func (v *View) writeIndicators(content *strings.Builder, indicators *trenddomain.Result, lastTradingDay string) {
	rsiValuation := getRSIValuation(indicators.RSI)
	valuationStyle := v.getValuationStyle(rsiValuation)
	fmt.Fprintf(content, "  RSI(%d):     %s  — %s\n", 14, formatFloat(indicators.RSI, 1), valuationStyle.Render(rsiValuation))
	fmt.Fprintf(content, "  EMA(10):     $%s\n", humanizeFloat(indicators.EMAFast, 2))
	fmt.Fprintf(content, "  EMA(20):     $%s\n", humanizeFloat(indicators.EMASlow, 2))
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
	shortcuts := []struct {
		key  string
		desc string
	}{
		{"Enter", "submit"},
		{"Esc", "clear"},
		{"↑/↓", "history"},
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

// parseChange parses the change and change percent, returns values.
func parseChange(changeStr, changePercentStr string) (float64, string) {
	if changeStr == "" {
		return 0, "—"
	}

	change, err := strconv.ParseFloat(changeStr, 64)
	if err != nil {
		change = 0
	}

	if changePercentStr == "" {
		return change, "—"
	}

	// Keep % sign for display
	return change, changePercentStr
}

// formatVolume formats a volume number with thousands separators.
// For crypto, also displays USD volume.
func formatVolume(volumeStr, symbol string, price float64) string {
	if volumeStr == "" {
		return "—"
	}

	vol, err := strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return volumeStr
	}

	// Check if this is crypto (common crypto symbols)
	isCrypto := symbol == "BTC" || symbol == "ETH" || symbol == "SOL" ||
		symbol == "XRP" || symbol == "ADA" || symbol == "DOGE" || symbol == "AVAX"

	if isCrypto {
		// Display crypto volume with native units and USD value
		usdVol := vol * price
		return fmt.Sprintf("%.2f %s ($%s)", vol, symbol, humanizeFloat(usdVol, 0))
	}

	// For stocks, format with traditional volume notation
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

// formatPrice formats a price string with 2 decimal places and comma separators.
func formatPrice(s string) string {
	if s == "" {
		return "—"
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}
	return "$" + humanizeFloat(val, 2)
}

// humanizeFloat formats a float with comma separators and specified decimal places.
func humanizeFloat(val float64, decimals int) string {
	// Format with commas: 66959.98 → "66,959.98"
	intPart := int64(val)
	fracPart := val - float64(intPart)

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

	if decimals > 0 {
		fracStr := fmt.Sprintf("%.*f", decimals, fracPart)
		return str + fracStr[1:] // Remove leading "0"
	}
	return str
}

// formatPercent formats a percentage string with 2 decimal places.
func formatPercent(s string) string {
	if s == "" {
		return "—"
	}
	// Remove % sign, parse, reformat with 2 decimals
	clean := strings.TrimSuffix(s, "%")
	pct, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%.2f%%", pct)
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
