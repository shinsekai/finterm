// Package news provides tests for the news feed TUI view.
package news

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/owner/finterm/internal/alphavantage"
)

// viewMockTheme is a mock implementation of Theme for view tests.
type viewMockTheme struct{}

func (m *viewMockTheme) Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (m *viewMockTheme) TableRow() lipgloss.Style {
	return lipgloss.NewStyle()
}

func (m *viewMockTheme) TableEmpty() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true)
}

func (m *viewMockTheme) Bullish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("50FA7B")).Bold(true)
}

func (m *viewMockTheme) Bearish() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("FF5555")).Bold(true)
}

func (m *viewMockTheme) Neutral() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("F1FA8C")).Bold(true)
}

func (m *viewMockTheme) Help() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true)
}

func (m *viewMockTheme) Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("FF5555")).Bold(true)
}

func (m *viewMockTheme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Faint(true)
}

func (m *viewMockTheme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("7D56F4"))
}

func (m *viewMockTheme) TabActive() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Background(lipgloss.Color("#7D56F4")).Bold(true)
}

func (m *viewMockTheme) Foreground() lipgloss.Color {
	return lipgloss.Color("#F8F8F2")
}

func (m *viewMockTheme) Background() lipgloss.Color {
	return lipgloss.Color("#282A36")
}

// newTestModelForView creates a model with test articles for view testing.
func newTestModelForView() *Model {
	model := NewModel()
	model.width = 100
	model.height = 30
	model.calculateViewportHeight()

	// Add test articles
	articles := []Article{
		{
			Item: &alphavantage.NewsItem{
				Title:         "Apple Reports Record Revenue",
				URL:           "https://example.com/aapl",
				Source:        "Reuters",
				TimePublished: "20250101T120000",
			},
			Tickers:        []string{"AAPL"},
			SentimentScore: 0.72,
			RelevanceScore: 0.95,
		},
		{
			Item: &alphavantage.NewsItem{
				Title:         "Bitcoin Drops Below Support",
				URL:           "https://example.com/btc",
				Source:        "CoinDesk",
				TimePublished: "20250102T130000",
			},
			Tickers:        []string{"BTC"},
			SentimentScore: 0.28,
			RelevanceScore: 0.88,
		},
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.state = StateLoaded

	return model
}

// TestNewsModel_View_ArticleRendering verifies article rendering.
func TestNewsModel_View_ArticleRendering(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "NEWS FEED", "View should contain title")
	assert.Contains(t, rendered, "Apple Reports Record Revenue", "View should contain first article title")
	assert.Contains(t, rendered, "Bitcoin Drops Below Support", "View should contain second article title")
	assert.Contains(t, rendered, "AAPL", "View should contain AAPL ticker")
	assert.Contains(t, rendered, "BTC", "View should contain BTC ticker")
	assert.Contains(t, rendered, "Reuters", "View should contain Reuters source")
	assert.Contains(t, rendered, "CoinDesk", "View should contain CoinDesk source")
}

// TestNewsModel_View_SentimentColors verifies sentiment color coding.
func TestNewsModel_View_SentimentColors(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	// Check for sentiment indicators
	assert.Contains(t, rendered, "▲", "View should contain bullish indicator")
	assert.Contains(t, rendered, "▼", "View should contain bearish indicator")
	assert.Contains(t, rendered, "0.72", "View should contain bullish score")
	assert.Contains(t, rendered, "0.28", "View should contain bearish score")
}

// TestNewsModel_View_EmptyState verifies empty state rendering.
func TestNewsModel_View_EmptyState(t *testing.T) {
	model := NewModel()
	model.width = 100
	model.height = 30
	model.calculateViewportHeight()
	model.state = StateLoaded

	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "NEWS FEED", "View should contain title")
	assert.Contains(t, rendered, "No articles found", "View should show empty message")
}

// TestNewsModel_View_FilterSortBar verifies filter and sort bar rendering.
func TestNewsModel_View_FilterSortBar(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	// Check filter options
	assert.Contains(t, rendered, "Filter:", "View should contain filter label")
	assert.Contains(t, rendered, "all", "View should contain all filter option")
	assert.Contains(t, rendered, "equities", "View should contain equities filter option")
	assert.Contains(t, rendered, "crypto", "View should contain crypto filter option")
	assert.Contains(t, rendered, "macro", "View should contain macro filter option")

	// Check sort options
	assert.Contains(t, rendered, "Sort:", "View should contain sort label")
	assert.Contains(t, rendered, "newest", "View should contain newest sort option")
	assert.Contains(t, rendered, "score", "View should contain score sort option")
}

// TestNewsModel_View_FilterState verifies filter state in rendering.
func TestNewsModel_View_FilterState(t *testing.T) {
	tests := []struct {
		name       string
		filter     Filter
		shouldHave string
	}{
		{"FilterAll", FilterAll, "[all]"},
		{"FilterEquities", FilterEquities, "[equities]"},
		{"FilterCrypto", FilterCrypto, "[crypto]"},
		{"FilterMacro", FilterMacro, "[macro]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModelForView()
			model.filter = tt.filter
			model.applyFilterAndSort()

			view := NewView(*model).SetTheme(&viewMockTheme{})
			rendered := view.Render()

			assert.Contains(t, rendered, tt.shouldHave, "View should show active filter")
		})
	}
}

// TestNewsModel_View_SortState verifies sort state in rendering.
func TestNewsModel_View_SortState(t *testing.T) {
	tests := []struct {
		name       string
		sort       Sort
		shouldHave string
	}{
		{"SortNewest", SortNewest, "[newest]"},
		{"SortScore", SortScore, "[score]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newTestModelForView()
			model.sort = tt.sort
			model.sortArticles(model.articles)

			view := NewView(*model).SetTheme(&viewMockTheme{})
			rendered := view.Render()

			assert.Contains(t, rendered, tt.shouldHave, "View should show active sort")
		})
	}
}

// TestNewsModel_View_HelpBar verifies help bar rendering.
func TestNewsModel_View_HelpBar(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "[j/k] navigate", "View should contain navigation help")
	assert.Contains(t, rendered, "[Enter] open", "View should contain open help")
	assert.Contains(t, rendered, "[f] filter", "View should contain filter help")
	assert.Contains(t, rendered, "[s] sort", "View should contain sort help")
	assert.Contains(t, rendered, "[r] refresh", "View should contain refresh help")
}

// TestNewsModel_View_LoadingState verifies loading state rendering.
func TestNewsModel_View_LoadingState(t *testing.T) {
	model := NewModel()
	model.width = 100
	model.height = 30
	model.calculateViewportHeight()
	model.state = StateLoading

	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "NEWS FEED", "View should contain title")
	assert.Contains(t, rendered, "Loading news...", "View should contain loading message")
}

// TestNewsModel_View_ErrorState verifies error state rendering.
func TestNewsModel_View_ErrorState(t *testing.T) {
	model := NewModel()
	model.width = 100
	model.height = 30
	model.calculateViewportHeight()
	model.state = StateError
	model.err = newTestError("test error")

	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "NEWS FEED", "View should contain title")
	assert.Contains(t, rendered, "Error:", "View should contain error indicator")
	assert.Contains(t, rendered, "test error", "View should contain error message")
}

// TestNewsModel_View_ActiveArticleHighlight verifies active article highlighting.
func TestNewsModel_View_ActiveArticleHighlight(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	// Default active index is 0, so first article should be highlighted
	assert.Contains(t, rendered, "Apple Reports Record Revenue", "View should contain active article title")
}

// TestNewsModel_View_RelevanceScore verifies relevance score rendering.
func TestNewsModel_View_RelevanceScore(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Relevance:", "View should contain relevance label")
	assert.Contains(t, rendered, "0.95", "View should contain relevance score")
	assert.Contains(t, rendered, "0.88", "View should contain relevance score")
}

// TestNewsModel_View_SentimentLabel verifies sentiment label rendering.
func TestNewsModel_View_SentimentLabel(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "Sentiment:", "View should contain sentiment label")
	assert.Contains(t, rendered, "Bullish", "View should contain bullish label")
	assert.Contains(t, rendered, "Bearish", "View should contain bearish label")
}

// TestNewsModel_View_TimeFormatting verifies time formatting.
func TestNewsModel_View_TimeFormatting(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	// Check that time is formatted (dates are from Jan 2025, which is over a year ago)
	assert.Contains(t, rendered, "Jan", "View should contain formatted time")
	assert.Contains(t, rendered, "2025", "View should contain formatted date")
}

// TestNewsModel_View_NeutralSentiment verifies neutral sentiment rendering.
func TestNewsModel_View_NeutralSentiment(t *testing.T) {
	model := newTestModelForView()

	// Add a neutral sentiment article
	article := Article{
		Item: &alphavantage.NewsItem{
			Title:         "Microsoft Expands Azure",
			URL:           "https://example.com/msft",
			Source:        "Bloomberg",
			TimePublished: "20250103T140000",
		},
		Tickers:        []string{"MSFT"},
		SentimentScore: 0.51,
		RelevanceScore: 0.76,
	}
	model.allArticles = append(model.allArticles, article)
	model.applyFilterAndSort()

	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.Contains(t, rendered, "─", "View should contain neutral indicator")
	assert.Contains(t, rendered, "0.51", "View should contain neutral score")
	assert.Contains(t, rendered, "Neutral", "View should contain neutral label")
}

// TestNewsModel_View_DefaultTheme verifies view works with default theme.
func TestNewsModel_View_DefaultTheme(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model)

	// Should not panic with default theme
	rendered := view.Render()

	assert.NotEmpty(t, rendered, "View should not be empty with default theme")
	assert.Contains(t, rendered, "NEWS FEED", "View should contain title with default theme")
}

// TestNewsModel_View_String verifies String method.
func TestNewsModel_View_String(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model).SetTheme(&viewMockTheme{})

	rendered := view.String()
	assert.Equal(t, view.Render(), rendered, "String should return same as Render")
}

// TestNewsModel_View_SetTheme verifies theme setting.
func TestNewsModel_View_SetTheme(t *testing.T) {
	model := newTestModelForView()
	view := NewView(*model)

	theme := &viewMockTheme{}
	result := view.SetTheme(theme)

	assert.Equal(t, view, result, "SetTheme should return view for chaining")
	assert.Equal(t, theme, view.theme, "Theme should be set")
}

// TestNewsModel_View_ViewportHandling verifies viewport rendering with many articles.
func TestNewsModel_View_ViewportHandling(t *testing.T) {
	model := newTestModelForView()

	// Add more articles than viewport height
	articles := make([]Article, 20)
	for i := range articles {
		articles[i] = Article{
			Item: &alphavantage.NewsItem{
				Title:         string(rune('A' + i)),
				URL:           "https://example.com",
				Source:        "Test Source",
				TimePublished: "20250101T120000",
			},
			Tickers:        []string{"TEST"},
			SentimentScore: 0.5,
			RelevanceScore: 0.5,
		}
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.state = StateLoaded

	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.NotEmpty(t, rendered, "View should render with many articles")
	assert.Contains(t, rendered, "NEWS FEED", "View should contain title")
}

// TestNewsModel_View_ScrollOffset verifies scroll offset in rendering.
func TestNewsModel_View_ScrollOffset(t *testing.T) {
	model := newTestModelForView()

	// Add many articles
	articles := make([]Article, 15)
	for i := range articles {
		articles[i] = Article{
			Item: &alphavantage.NewsItem{
				Title:         string(rune('A' + i)),
				URL:           "https://example.com",
				Source:        "Test Source",
				TimePublished: "20250101T120000",
			},
			Tickers:        []string{"TEST"},
			SentimentScore: 0.5,
			RelevanceScore: 0.5,
		}
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.state = StateLoaded

	// Scroll down
	model.activeIndex = 10
	model.calculateViewportHeight()
	if model.activeIndex >= model.viewportHeight {
		model.scrollOffset = model.activeIndex - model.viewportHeight + 1
	}

	view := NewView(*model).SetTheme(&viewMockTheme{})
	rendered := view.Render()

	assert.NotEmpty(t, rendered, "View should render with scroll offset")
}

// testError is a simple error type for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// newTestError creates a test error.
func newTestError(msg string) *testError {
	return &testError{msg: msg}
}
