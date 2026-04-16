// Package quote provides the single ticker quote TUI view.
package quote

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/shinsekai/finterm/internal/alphavantage"
	trenddomain "github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// Theme defines the interface for theme styling.
// This is defined here to avoid import cycle with the parent tui package.
type Theme interface {
	Title() lipgloss.Style
	Subtitle() lipgloss.Style
	Bullish() lipgloss.Style
	Bearish() lipgloss.Style
	Neutral() lipgloss.Style
	BullishBadge() lipgloss.Style
	BearishBadge() lipgloss.Style
	NeutralBadge() lipgloss.Style
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

func (t *defaultTheme) BullishBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("#50FA7B")).
		Bold(true).
		Padding(0, 1)
}

func (t *defaultTheme) BearishBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("#FF5555")).
		Bold(true).
		Padding(0, 1)
}

func (t *defaultTheme) NeutralBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#282A36")).
		Background(lipgloss.Color("#F1FA8C")).
		Bold(true).
		Padding(0, 1)
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

	// Recent lookups
	history := v.model.GetLookupHistory()
	if len(history) > 0 {
		content += "\n\n"
		content += v.theme.Muted().Render("Recent:")
		// Show last 5, most recent first
		limit := 5
		if len(history) < limit {
			limit = len(history)
		}
		for i := len(history) - 1; i >= len(history)-limit; i-- {
			content += "  " + v.theme.Accent().Render(history[i])
		}
	}

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
	helpText := v.theme.Accent().Render("Enter") + " " + v.theme.Muted().Render("retry") + "  ·  " + v.theme.Accent().Render("Esc") + " " + v.theme.Muted().Render("clear")
	content += helpText
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

	// Day range bar
	dayRangeBar := v.renderDayRangeBar(price, quote.Low, quote.High)
	if dayRangeBar != "" {
		content.WriteString(dayRangeBar)
		content.WriteString("\n\n")
	}

	// OHLCV data rows with muted labels
	v.writePriceField(&content, "Open", "$"+formatPrice(quote.Open), quote.Open)
	v.writePriceField(&content, "High", "$"+formatPrice(quote.High), quote.High)
	v.writePriceField(&content, "Low", "$"+formatPrice(quote.Low), quote.Low)
	v.writePriceField(&content, "Volume", formatVolume(quote.Volume, quote.Symbol, price), quote.Volume)
	v.writePriceField(&content, "Prev Close", "$"+formatPrice(quote.PreviousClose), quote.PreviousClose)

	return v.theme.Card().Render(content.String())
}

// renderDayRangeBar renders a horizontal bar showing the price position within the day's range.
func (v *View) renderDayRangeBar(price float64, lowStr, highStr string) string {
	// Parse low and high values
	low, err := strconv.ParseFloat(lowStr, 64)
	if err != nil || low == 0 {
		return ""
	}

	high, err := strconv.ParseFloat(highStr, 64)
	if err != nil || high == 0 {
		return ""
	}

	// Calculate position within range (0 to 1)
	var position float64
	if high > low {
		position = (price - low) / (high - low)
	}
	// Clamp to [0, 1]
	if position < 0 {
		position = 0
	} else if position > 1 {
		position = 1
	}

	// Build the bar: 20 characters total
	barWidth := 20
	filledWidth := int(position * float64(barWidth))
	if filledWidth > barWidth {
		filledWidth = barWidth
	}
	if filledWidth < 0 {
		filledWidth = 0
	}

	filled := strings.Repeat("▓", filledWidth)
	empty := strings.Repeat("░", barWidth-filledWidth)

	// Format low and high values with proper decimal places
	lowVal, _ := strconv.ParseFloat(lowStr, 64)
	highVal, _ := strconv.ParseFloat(highStr, 64)

	var content strings.Builder
	content.WriteString(v.theme.Muted().Render(fmt.Sprintf("Low $%.2f", lowVal)))
	content.WriteString(" ")
	content.WriteString(v.theme.Accent().Render(filled) + v.theme.Muted().Render(empty))
	content.WriteString(" ")
	content.WriteString(v.theme.Muted().Render(fmt.Sprintf("High $%.2f", highVal)))

	return content.String()
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

	// Signal Systems section
	fmt.Fprintf(&content, "%s\n", v.theme.Muted().Render("─ Signal Systems ──"))
	content.WriteString("\n")

	// FTEMA badge
	fmt.Fprintf(&content, "%s%s\n",
		v.theme.Muted().Render(fmt.Sprintf("%-14s", "FTEMA")),
		v.renderFTEMABadge(indicators.Signal))

	// BLITZ badge
	fmt.Fprintf(&content, "%s%s\n",
		v.theme.Muted().Render(fmt.Sprintf("%-14s", "BLITZ")),
		v.renderBlitzBadge(indicators.BlitzScore))

	// DESTINY badge
	fmt.Fprintf(&content, "%s%s\n",
		v.theme.Muted().Render(fmt.Sprintf("%-14s", "DESTINY")),
		v.renderDestinyBadge(indicators.DestinyScore))
	// FLOW badge
	fmt.Fprintf(&content, "%s%s\n",
		v.theme.Muted().Render(fmt.Sprintf("%-14s", "FLOW")),
		v.renderFlowBadge(indicators.FlowScore))
	// VORTEX badge
	fmt.Fprintf(&content, "%s%s\n",
		v.theme.Muted().Render(fmt.Sprintf("%-14s", "VORTEX")),
		v.renderVortexBadge(indicators.VortexScore))

	// Composite section
	fmt.Fprintf(&content, "\n%s\n", v.theme.Muted().Render("─ Composite ────"))
	content.WriteString("\n")

	// TPI value with color
	tpiStyle := v.theme.Muted()
	if indicators.TPI > 0 {
		tpiStyle = v.theme.Bullish()
	} else if indicators.TPI < 0 {
		tpiStyle = v.theme.Bearish()
	}
	fmt.Fprintf(&content, "%s%s\n",
		v.theme.Muted().Render(fmt.Sprintf("%-14s", "TPI")),
		tpiStyle.Render(fmt.Sprintf("%+.2f", indicators.TPI)))

	// TPI gauge + signal label on same line
	var tpiSignalStyle lipgloss.Style
	if indicators.TPI > 0 {
		tpiSignalStyle = v.theme.Bullish()
	} else {
		tpiSignalStyle = v.theme.Bearish()
	}
	fmt.Fprintf(&content, "%s%s %s\n",
		v.theme.Muted().Render(fmt.Sprintf("%-14s", "")),
		v.renderTPIGauge(indicators.TPI),
		tpiSignalStyle.Render(indicators.TPISignal))

	// Last updated
	if lastTradingDay != "" {
		content.WriteString("\n")
		content.WriteString(v.theme.Muted().Render("Last updated: " + lastTradingDay))
	}

	return v.theme.Card().Render(content.String())
}

// renderFTEMABadge renders FTEMA (EMA crossover) as a colored badge.
func (v *View) renderFTEMABadge(signal trenddomain.Signal) string {
	switch signal {
	case trenddomain.Bullish:
		return v.theme.Bullish().Render("▲  LONG")
	case trenddomain.Bearish:
		return v.theme.Bearish().Render("▼ SHORT")
	default:
		return ""
	}
}

// renderBlitzBadge renders a BLITZ score as a colored badge.
func (v *View) renderBlitzBadge(blitzScore int) string {
	switch blitzScore {
	case 1:
		return v.theme.Bullish().Render("▲  LONG")
	case -1:
		return v.theme.Bearish().Render("▼ SHORT")
	default:
		return ""
	}
}

// renderDestinyBadge renders a DESTINY score as a colored badge.
func (v *View) renderDestinyBadge(destinyScore int) string {
	switch destinyScore {
	case 1:
		return v.theme.Bullish().Render("▲  LONG")
	case -1:
		return v.theme.Bearish().Render("▼ SHORT")
	default:
		return ""
	}
}

// renderFlowBadge renders a FLOW score as a colored badge.
func (v *View) renderFlowBadge(flowScore int) string {
	switch flowScore {
	case 1:
		return v.theme.Bullish().Render("▲  LONG")
	case -1:
		return v.theme.Bearish().Render("▼ SHORT")
	default:
		return ""
	}
}

// renderVortexBadge renders a VORTEX score as a colored badge.
func (v *View) renderVortexBadge(vortexScore int) string {
	switch vortexScore {
	case 1:
		return v.theme.Bullish().Render("▲  LONG")
	case -1:
		return v.theme.Bearish().Render("▼ SHORT")
	default:
		return ""
	}
}

// renderTPIGauge renders TPI as a horizontal gauge bar.
// Gauge is 10 chars wide: left half (0-4) is bearish, center (5) is neutral, right half (6-9) is bullish.
// For TPI > 0: fill from center rightward with "▓" in bullish color, "░" in muted.
// For TPI <= 0: fill from center leftward with "▓" in bearish color, "░" in muted.
func (v *View) renderTPIGauge(tpi float64) string {
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

	return gaugeStr
}

// renderRSIBar renders an RSI progress bar with zone markers.
func (v *View) renderRSIBar(rsi float64) string {
	var content strings.Builder

	// RSI label
	rsiValuation := getRSIValuation(rsi)
	valuationStyle := v.getValuationStyle(rsiValuation)
	fmt.Fprintf(&content, "%-14s%.1f  — %s\n",
		v.theme.Muted().Render("RSI(14):"),
		rsi,
		valuationStyle.Render(rsiValuation))

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
		fmt.Fprintf(content, "%-14s%s\n", v.theme.Muted().Render(label+":"), value)
	}
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
