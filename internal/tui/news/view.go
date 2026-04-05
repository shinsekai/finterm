// Package news provides the news feed TUI view.
package news

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// Theme defines the interface for theme styling.
// This is defined here to avoid import cycle with parent tui package.
type Theme interface {
	Title() lipgloss.Style
	TableRow() lipgloss.Style
	TableEmpty() lipgloss.Style
	Bullish() lipgloss.Style
	Bearish() lipgloss.Style
	Neutral() lipgloss.Style
	BullishBadge() lipgloss.Style
	BearishBadge() lipgloss.Style
	NeutralBadge() lipgloss.Style
	Accent() lipgloss.Style
	Divider() lipgloss.Style
	Help() lipgloss.Style
	Error() lipgloss.Style
	Muted() lipgloss.Style
	Spinner() lipgloss.Style
	TabActive() lipgloss.Style
	Foreground() lipgloss.Color
	Background() lipgloss.Color
}

// View represents the news view renderer.
type View struct {
	model Model
	theme Theme
}

// NewView creates a new view from a model.
func NewView(m Model) *View {
	return &View{
		model: m,
	}
}

// SetTheme sets the theme for the view.
func (v *View) SetTheme(theme Theme) *View {
	v.theme = theme
	return v
}

// Render renders the entire news view.
func (v *View) Render() string {
	if v.theme == nil {
		// Use default styles if no theme is set
		v.theme = &defaultTheme{}
	}

	width := v.model.GetWidth()

	var builder strings.Builder

	// Render title bar
	builder.WriteString(v.renderTitleBar(width))
	builder.WriteString("\n")

	// Render separator
	builder.WriteString(v.theme.Divider().Render(strings.Repeat("─", width)))
	builder.WriteString("\n")

	// Render filter/sort bar
	builder.WriteString(v.renderFilterSortBar(width))
	builder.WriteString("\n")

	// Render separator
	builder.WriteString(v.theme.Divider().Render(strings.Repeat("─", width)))
	builder.WriteString("\n")

	// Render content area
	content := v.renderContent()
	builder.WriteString(content)
	builder.WriteString("\n")

	// Render separator
	builder.WriteString(v.theme.Divider().Render(strings.Repeat("─", width)))
	builder.WriteString("\n")

	// Render help bar
	builder.WriteString(v.renderHelpBar(width))

	return builder.String()
}

// renderTitleBar renders the title bar.
func (v *View) renderTitleBar(width int) string {
	title := v.theme.Accent().Render("◉") + " " + v.theme.Title().Render("News Feed")
	refresh := v.theme.Help().Render("r refresh")
	help := v.theme.Help().Render("? help")
	rightSide := refresh + " · " + help

	// Position title on left, help items on the right
	titleWidth := lipgloss.Width(title)
	rightWidth := lipgloss.Width(rightSide)

	padding := width - titleWidth - rightWidth
	if padding < 0 {
		padding = 0
	}

	pad := strings.Repeat(" ", padding)
	return title + pad + rightSide
}

// renderFilterSortBar renders the filter and sort controls.
func (v *View) renderFilterSortBar(width int) string {
	filter := v.model.GetFilter()
	sort := v.model.GetSort()

	// Build filter options
	filterAll := renderFilterOption("all", filter, FilterAll, v.theme)
	filterEquities := renderFilterOption("equities", filter, FilterEquities, v.theme)
	filterCrypto := renderFilterOption("crypto", filter, FilterCrypto, v.theme)
	filterMacro := renderFilterOption("macro", filter, FilterMacro, v.theme)

	// Build sort options
	sortNewest := renderSortOption("newest", sort, SortNewest, v.theme)
	sortScore := renderSortOption("score", sort, SortScore, v.theme)

	// Construct the bar
	left := "Filter: " + filterAll + " " + filterEquities + " " + filterCrypto + " " + filterMacro
	right := "Sort: " + sortNewest + " " + sortScore

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := width - leftWidth - rightWidth - 2
	if padding < 0 {
		padding = 0
	}

	return left + strings.Repeat(" ", padding) + right
}

// renderFilterOption renders a single filter option.
func renderFilterOption(label string, currentFilter Filter, option Filter, theme Theme) string {
	if currentFilter == option {
		return theme.TabActive().Render("[" + label + "]")
	}
	return theme.Muted().Render(label)
}

// renderSortOption renders a single sort option.
func renderSortOption(label string, currentSort Sort, option Sort, theme Theme) string {
	if currentSort == option {
		return theme.TabActive().Render("[" + label + "]")
	}
	return theme.Muted().Render(label)
}

// renderContent renders the main content area.
func (v *View) renderContent() string {
	state := v.model.GetState()

	switch state {
	case StateLoading:
		return v.renderLoading()
	case StateError:
		return v.renderError()
	case StateLoaded:
		return v.renderArticles()
	default:
		return v.theme.Error().Render("Unknown state")
	}
}

// renderLoading renders the loading state.
func (v *View) renderLoading() string {
	spinner := components.NewSpinner().WithText("⟳ Loading news…").WithFrameStyle(v.theme.Spinner())
	return spinner.Render()
}

// renderError renders the error state.
func (v *View) renderError() string {
	err := v.model.GetError()
	if err == nil {
		return v.theme.Error().Render("✗ An unknown error occurred")
	}
	return v.theme.Error().Render("✗ Error: " + err.Error())
}

// renderArticles renders the article list.
func (v *View) renderArticles() string {
	articles := v.model.GetArticles()
	activeIndex := v.model.GetActiveIndex()
	scrollOffset := v.model.GetScrollOffset()
	viewportHeight := v.model.GetViewportHeight()

	if len(articles) == 0 {
		return v.renderEmptyState()
	}

	// Determine visible range
	endIndex := scrollOffset + viewportHeight
	if endIndex > len(articles) {
		endIndex = len(articles)
	}

	var builder strings.Builder
	for i := scrollOffset; i < endIndex; i++ {
		if i >= len(articles) {
			break
		}

		article := articles[i]
		isActive := (i == activeIndex)
		builder.WriteString(v.renderArticle(article, isActive))
		builder.WriteString("\n")

		// Add spacing between articles
		if i < endIndex-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// renderArticle renders a single article.
func (v *View) renderArticle(article Article, isActive bool) string {
	item := article.Item
	sentiment := article.SentimentScore
	relevance := article.RelevanceScore

	// Determine sentiment indicator and style
	indicator, sentimentStyle := v.getSentimentIndicator(sentiment)
	sentimentLabel := v.getSentimentLabel(sentiment)

	// Build ticker string with accent style
	tickerStr := strings.Join(article.Tickers, ", ")
	if len(tickerStr) > 10 {
		tickerStr = tickerStr[:7] + "..."
	}
	tickerDisplay := v.theme.Accent().Render(tickerStr)

	// Parse time
	timeStr := v.formatTimePublished(item.TimePublished)

	// Render headline (truncated if needed)
	headline := item.Title
	maxHeadlineWidth := v.model.GetWidth() - 30
	if lipgloss.Width(headline) > maxHeadlineWidth {
		headline = components.TruncateText(headline, maxHeadlineWidth)
	}

	// Build article lines
	var lines []string

	// Line 1: sentiment + ticker + headline
	line1Style := v.theme.TableRow()
	if isActive {
		line1Style = line1Style.Background(v.theme.Foreground()).Foreground(v.theme.Background())
	}

	line1 := indicator + "  " +
		sentimentStyle.Render(fmt.Sprintf("%.2f", sentiment)) + "  " +
		tickerDisplay + " │ " +
		line1Style.Render(headline)
	lines = append(lines, line1)

	// Line 2: source + time
	line2 := "              │ " +
		v.theme.Muted().Render(item.Source) + " · " +
		v.theme.Muted().Render(timeStr)
	lines = append(lines, line2)

	// Line 3: sentiment label + relevance
	line3 := "              │ " +
		v.theme.Muted().Render("Sentiment: ") +
		sentimentStyle.Render(sentimentLabel) + " · " +
		v.theme.Muted().Render("Relevance: ") +
		v.theme.Muted().Render(fmt.Sprintf("%.2f", relevance))
	lines = append(lines, line3)

	return strings.Join(lines, "\n")
}

// getSentimentIndicator returns the sentiment indicator and style.
func (v *View) getSentimentIndicator(score float64) (string, lipgloss.Style) {
	switch {
	case score > 0.6:
		return v.theme.BullishBadge().Render("▲"), v.theme.Bullish()
	case score < 0.4:
		return v.theme.BearishBadge().Render("▼"), v.theme.Bearish()
	default:
		return v.theme.NeutralBadge().Render("─"), v.theme.Neutral()
	}
}

// getSentimentLabel returns the sentiment label.
func (v *View) getSentimentLabel(score float64) string {
	switch {
	case score > 0.6:
		return "Bullish"
	case score < 0.4:
		return "Bearish"
	default:
		return "Neutral"
	}
}

// formatTimePublished formats the published time as a relative time.
func (v *View) formatTimePublished(timeStr string) string {
	// Alpha Vantage format: YYYYMMDDTHHMMSS
	if len(timeStr) < 10 {
		return timeStr
	}

	// Parse the time
	published, err := time.Parse("20060102T150405", timeStr)
	if err != nil {
		return timeStr
	}

	// Calculate relative time
	duration := time.Since(published)
	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	return published.Format("Jan 2, 2006")
}

// renderEmptyState renders the empty state.
func (v *View) renderEmptyState() string {
	return v.theme.TableEmpty().Render("○ No articles found")
}

// renderHelpBar renders the help bar.
func (v *View) renderHelpBar(_ int) string {
	helpParts := []string{
		v.theme.Accent().Render("j/k") + " " + v.theme.Muted().Render("navigate"),
		v.theme.Accent().Render("Enter") + " " + v.theme.Muted().Render("open"),
		v.theme.Accent().Render("f") + " " + v.theme.Muted().Render("filter"),
		v.theme.Accent().Render("s") + " " + v.theme.Muted().Render("sort"),
		v.theme.Accent().Render("r") + " " + v.theme.Muted().Render("refresh"),
	}
	return strings.Join(helpParts, "  ·  ")
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

func (d *defaultTheme) TableEmpty() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true).Faint(true)
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
	return lipgloss.NewStyle().Italic(true).Faint(true)
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

func (d *defaultTheme) TabActive() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Background(lipgloss.Color("#7D56F4")).Bold(true)
}

func (d *defaultTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#F8F8F2")
}

func (d *defaultTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282A36")
}

func (d *defaultTheme) BullishBadge() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#282A36")).Background(lipgloss.Color("50FA7B")).Bold(true)
}

func (d *defaultTheme) BearishBadge() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#282A36")).Background(lipgloss.Color("FF5555")).Bold(true)
}

func (d *defaultTheme) NeutralBadge() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#282A36")).Background(lipgloss.Color("F1FA8C")).Bold(true)
}

func (d *defaultTheme) Accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("BD93F9")).Bold(true)
}

func (d *defaultTheme) Divider() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#44475A")).Faint(true)
}
