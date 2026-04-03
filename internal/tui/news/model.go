// Package news provides the news feed TUI model.
package news

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/owner/finterm/internal/alphavantage"
	"github.com/owner/finterm/internal/tui/components"
)

// Filter represents the news filter type.
type Filter int

const (
	// FilterAll shows all articles.
	FilterAll Filter = iota
	// FilterEquities shows only equity-related articles.
	FilterEquities
	// FilterCrypto shows only crypto-related articles.
	FilterCrypto
	// FilterMacro shows only macroeconomic news.
	FilterMacro
)

// String returns the string representation of the Filter.
func (f Filter) String() string {
	switch f {
	case FilterAll:
		return "all"
	case FilterEquities:
		return "equities"
	case FilterCrypto:
		return "crypto"
	case FilterMacro:
		return "macro"
	default:
		return "all"
	}
}

// Next returns the next filter in the cycle.
func (f Filter) Next() Filter {
	return (f + 1) % 4
}

// Sort represents the sort order for news articles.
type Sort int

const (
	// SortNewest sorts by newest first.
	SortNewest Sort = iota
	// SortScore sorts by sentiment score descending.
	SortScore
)

// String returns the string representation of the Sort.
func (s Sort) String() string {
	switch s {
	case SortNewest:
		return "newest"
	case SortScore:
		return "score"
	default:
		return "newest"
	}
}

// Next returns the next sort in the cycle.
func (s Sort) Next() Sort {
	return (s + 1) % 2
}

// State represents the model state.
type State int

const (
	// StateLoading is the initial state when data is being fetched.
	StateLoading State = iota
	// StateLoaded is when data has been successfully loaded.
	StateLoaded
	// StateError is when there was an error loading data.
	StateError
)

// String returns the string representation of the State.
func (s State) String() string {
	switch s {
	case StateLoading:
		return "Loading"
	case StateLoaded:
		return "Loaded"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Client defines the interface for fetching news data.
type Client interface {
	GetNewsSentiment(ctx context.Context, opts alphavantage.NewsOpts) (*alphavantage.NewsSentiment, error)
}

// Configure sets up the model with dependencies.
func (m *Model) Configure(
	ctx context.Context,
	client Client,
) *Model {
	m.client = client
	m.ctx, m.cancel = context.WithCancel(ctx)
	return m
}

// Article represents a filtered and sorted news article for display.
type Article struct {
	// Item is the original news item from the API.
	Item *alphavantage.NewsItem
	// RelevanceScore is the calculated relevance score (0-1).
	RelevanceScore float64
	// SentimentScore is the calculated sentiment score (0-1).
	SentimentScore float64
	// Tickers contains the relevant tickers for this article.
	Tickers []string
}

// Model represents the news view model.
type Model struct {
	// client fetches news data.
	client Client
	// articles contains the current articles (after filtering and sorting).
	articles []Article
	// allArticles contains all articles before filtering.
	allArticles []Article
	// activeIndex is the index of the currently selected article.
	activeIndex int
	// scrollOffset is the scroll offset for the viewport.
	scrollOffset int
	// filter is the current filter state.
	filter Filter
	// sort is the current sort state.
	sort Sort
	// state is the current state of the model.
	state State
	// err contains any error that occurred.
	err error
	// lastRefresh is when the data was last refreshed.
	lastRefresh time.Time
	// ctx is the context for async operations.
	ctx context.Context
	// cancel cancels the context.
	cancel context.CancelFunc
	// width and height are the terminal dimensions.
	width, height int
	// viewportHeight is the height available for the article list.
	viewportHeight int
}

// NewModel creates a new news model.
func NewModel() *Model {
	return &Model{
		articles:       []Article{},
		allArticles:    []Article{},
		activeIndex:    0,
		scrollOffset:   0,
		filter:         FilterAll,
		sort:           SortNewest,
		state:          StateLoading,
		width:          80,
		height:         24,
		viewportHeight: 20,
	}
}

// Init initializes the news model and returns an initial command.
func (m Model) Init() tea.Cmd {
	return m.fetchNewsCmd()
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.calculateViewportHeight()
		return m, nil

	case RefreshMsg:
		// Refresh news data
		m.state = StateLoading
		return m, m.fetchNewsCmd()

	case DataMsg:
		// Handle news data loaded
		return m.handleNewsData(msg)

	case ErrorMsg:
		// Handle error
		m.state = StateError
		m.err = msg.Err
		return m, nil
	}

	return m, nil
}

// handleKeyMsg handles keyboard input messages.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp, tea.KeyDown:
		return m.navigate(msg.Type)

	case tea.KeyEnter:
		return m, m.openArticleCmd()

	case tea.KeyRunes:
		// Check for single key presses
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'j':
				return m.navigate(tea.KeyDown)
			case 'k':
				return m.navigate(tea.KeyUp)
			case 'f':
				return m.cycleFilter()
			case 's':
				return m.cycleSort()
			case 'r':
				m.state = StateLoading
				return m, m.fetchNewsCmd()
			}
		}
	}

	return m, nil
}

// navigate handles article navigation with j/k or arrow keys.
func (m Model) navigate(keyType tea.KeyType) (Model, tea.Cmd) {
	if len(m.articles) == 0 {
		return m, nil
	}

	switch keyType {
	case tea.KeyUp:
		if m.activeIndex > 0 {
			m.activeIndex--
			// Scroll if the active index is above the viewport
			if m.activeIndex < m.scrollOffset {
				m.scrollOffset = m.activeIndex
			}
		}
	case tea.KeyDown:
		if m.activeIndex < len(m.articles)-1 {
			m.activeIndex++
			// Scroll if the active index is below the viewport
			if m.activeIndex >= m.scrollOffset+m.viewportHeight {
				m.scrollOffset = m.activeIndex - m.viewportHeight + 1
			}
		}
	}

	return m, nil
}

// cycleFilter cycles through the filter options.
func (m Model) cycleFilter() (Model, tea.Cmd) {
	m.filter = m.filter.Next()
	m.applyFilterAndSort()
	return m, nil
}

// cycleSort cycles through the sort options.
func (m Model) cycleSort() (Model, tea.Cmd) {
	m.sort = m.sort.Next()
	m.applyFilterAndSort()
	return m, nil
}

// applyFilterAndSort applies the current filter and sort to the articles.
func (m *Model) applyFilterAndSort() {
	// Filter articles
	m.articles = m.filterArticles(m.allArticles)

	// Sort articles
	m.sortArticles(m.articles)

	// Reset navigation
	m.activeIndex = 0
	m.scrollOffset = 0
}

// isCryptoTicker checks if a ticker is a cryptocurrency.

// filterArticles filters articles based on current filter.
func (m *Model) filterArticles(articles []Article) []Article {
	if m.filter == FilterAll {
		return articles
	}

	filtered := make([]Article, 0)
	for _, article := range articles {
		if m.matchesFilter(article) {
			filtered = append(filtered, article)
		}
	}
	return filtered
}

// matchesFilter checks if an article matches the current filter.
func (m *Model) matchesFilter(article Article) bool {
	switch m.filter {
	case FilterEquities:
		return m.hasEquityTicker(article)
	case FilterCrypto:
		return m.hasCryptoTicker(article)
	case FilterMacro:
		return m.isMacroTopic(article)
	}
	return false
}

// hasEquityTicker checks if an article has any equity tickers.
func (m *Model) hasEquityTicker(article Article) bool {
	for _, ticker := range article.Tickers {
		if !isCryptoTicker(ticker) {
			return true
		}
	}
	return false
}

// hasCryptoTicker checks if an article has any crypto tickers.
func (m *Model) hasCryptoTicker(article Article) bool {
	for _, ticker := range article.Tickers {
		if isCryptoTicker(ticker) {
			return true
		}
	}
	return false
}

// isMacroTopic checks if an article topic is macro-related.
func (m *Model) isMacroTopic(article Article) bool {
	topic := strings.ToLower(article.Item.Topic)
	macroKeywords := []string{"macro", "economy", "inflation", "gdp", "employment", "fed"}
	for _, keyword := range macroKeywords {
		if strings.Contains(topic, keyword) {
			return true
		}
	}
	return false
}
func isCryptoTicker(ticker string) bool {
	cryptoTickers := map[string]bool{
		"BTC":   true,
		"ETH":   true,
		"SOL":   true,
		"DOGE":  true,
		"ADA":   true,
		"XRP":   true,
		"DOT":   true,
		"AVAX":  true,
		"MATIC": true,
		"LINK":  true,
	}
	return cryptoTickers[strings.ToUpper(ticker)]
}

// sortArticles sorts articles based on the current sort option.
func (m *Model) sortArticles(articles []Article) {
	switch m.sort {
	case SortNewest:
		// Sort by time published (most recent first)
		sort.Slice(articles, func(i, j int) bool {
			return strings.Compare(articles[i].Item.TimePublished, articles[j].Item.TimePublished) > 0
		})
	case SortScore:
		// Sort by sentiment score (highest first)
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].SentimentScore > articles[j].SentimentScore
		})
	}
}

// handleNewsData processes the loaded news data.
func (m Model) handleNewsData(msg DataMsg) (Model, tea.Cmd) {
	m.allArticles = msg.Articles
	m.lastRefresh = msg.Timestamp
	m.applyFilterAndSort()
	m.state = StateLoaded
	m.err = nil
	return m, nil
}

// fetchNewsCmd returns a command to fetch news data.
func (m Model) fetchNewsCmd() tea.Cmd {
	m.state = StateLoading
	return func() tea.Msg {
		opts := alphavantage.NewsOpts{
			Limit: alphavantage.DefaultNewsLimit,
			Sort:  "LATEST",
		}
		response, err := m.client.GetNewsSentiment(m.ctx, opts)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("fetching news: %w", err)}
		}

		// Convert to Article format with scores
		articles := make([]Article, 0, len(response.Items))
		for _, item := range response.Items {
			article := convertToArticle(item)
			articles = append(articles, article)
		}

		return DataMsg{
			Articles:  articles,
			Timestamp: time.Now(),
		}
	}
}

// convertToArticle converts a NewsItem to an Article with calculated scores.
func convertToArticle(item alphavantage.NewsItem) Article {
	article := Article{
		Item:    &item,
		Tickers: make([]string, 0, len(item.Tickers)),
	}

	var totalRelevance, totalSentiment float64
	relevantTickers := 0

	for _, ticker := range item.Tickers {
		relevance, _ := alphavantage.ParseFloat(ticker.RelevanceScore)
		sentiment, _ := alphavantage.ParseFloat(ticker.TickerSentimentScore)

		// Only count relevant tickers (relevance > 0.5)
		if relevance > 0.5 {
			article.Tickers = append(article.Tickers, ticker.Ticker)
			totalRelevance += relevance
			totalSentiment += sentiment
			relevantTickers++
		}
	}

	// Calculate average scores
	if relevantTickers > 0 {
		article.RelevanceScore = totalRelevance / float64(relevantTickers)
		article.SentimentScore = totalSentiment / float64(relevantTickers)
	} else {
		// Fallback to overall sentiment if no relevant tickers
		sentiment, _ := alphavantage.ParseFloat(item.OverallSentimentScore)
		article.RelevanceScore = 0.5
		article.SentimentScore = sentiment
	}

	return article
}

// openArticleCmd returns a command to open the article URL.
func (m Model) openArticleCmd() tea.Cmd {
	if len(m.articles) == 0 {
		return nil
	}

	article := m.articles[m.activeIndex]
	url := article.Item.URL

	return func() tea.Msg {
		// Try to open in browser based on OS
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			// Fallback: copy to clipboard
			return ArticleOpenedMsg{
				URL:    url,
				Copied: true,
			}
		}

		err := cmd.Start()
		if err != nil {
			return ArticleOpenedMsg{
				URL:   url,
				Error: fmt.Errorf("opening browser: %w", err),
			}
		}

		return ArticleOpenedMsg{
			URL:    url,
			Opened: true,
		}
	}
}

// calculateViewportHeight calculates the available height for the article list.
func (m *Model) calculateViewportHeight() {
	// Reserve space for header (3 lines) and footer (1 line)
	m.viewportHeight = m.height - 4
	if m.viewportHeight < 1 {
		m.viewportHeight = 1
	}
}

// GetArticles returns the current articles.
func (m Model) GetArticles() []Article {
	return m.articles
}

// GetActiveArticle returns the active article.
func (m Model) GetActiveArticle() *Article {
	if len(m.articles) == 0 || m.activeIndex < 0 || m.activeIndex >= len(m.articles) {
		return nil
	}
	return &m.articles[m.activeIndex]
}

// GetActiveIndex returns the active index.
func (m Model) GetActiveIndex() int {
	return m.activeIndex
}

// GetFilter returns the current filter.
func (m Model) GetFilter() Filter {
	return m.filter
}

// GetSort returns the current sort.
func (m Model) GetSort() Sort {
	return m.sort
}

// GetState returns the current state.
func (m Model) GetState() State {
	return m.state
}

// GetError returns the error.
func (m Model) GetError() error {
	return m.err
}

// GetLastRefresh returns the last refresh time.
func (m Model) GetLastRefresh() time.Time {
	return m.lastRefresh
}

// GetWidth returns the current width.
func (m Model) GetWidth() int {
	return m.width
}

// GetHeight returns the current height.
func (m Model) GetHeight() int {
	return m.height
}

// GetViewportHeight returns the viewport height.
func (m Model) GetViewportHeight() int {
	return m.viewportHeight
}

// GetScrollOffset returns the scroll offset.
func (m Model) GetScrollOffset() int {
	return m.scrollOffset
}

// KeyBindings returns the keyboard bindings for the news view.
func (m Model) KeyBindings() []components.KeyBinding {
	return []components.KeyBinding{
		{Key: "↑/k", Description: "Previous article"},
		{Key: "↓/j", Description: "Next article"},
		{Key: "Enter", Description: "Open article URL"},
		{Key: "f", Description: "Cycle filter"},
		{Key: "s", Description: "Cycle sort"},
		{Key: "r", Description: "Refresh news"},
	}
}

// View renders the news view.
func (m Model) View() string {
	return NewView(m).Render()
}

// RefreshMsg is a message to refresh news data.
type RefreshMsg struct{}

// DataMsg is a message when news data is loaded.
type DataMsg struct {
	Articles  []Article
	Timestamp time.Time
}

// ErrorMsg is a message when an error occurs fetching news.
type ErrorMsg struct {
	Err error
}

// ArticleOpenedMsg is a message when an article URL is opened or copied.
type ArticleOpenedMsg struct {
	URL    string
	Opened bool
	Copied bool
	Error  error
}
