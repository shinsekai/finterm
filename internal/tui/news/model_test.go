// Package news provides tests for the news feed TUI model.
package news

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/owner/finterm/internal/alphavantage"
)

// mockClient is a mock implementation of Client for testing.
type mockClient struct {
	newsSentiment *alphavantage.NewsSentiment
	err           error
}

func (m *mockClient) GetNewsSentiment(_ context.Context, _ alphavantage.NewsOpts) (*alphavantage.NewsSentiment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.newsSentiment, nil
}

// newTestModel creates a configured model for testing and returns it as a value.
func newTestModel() Model {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{})
	return *model
}

// TestNewsModel_Init_FetchesArticles verifies that Init returns a fetch command.
func TestNewsModel_Init_FetchesArticles(t *testing.T) {
	model := newTestModel()
	cmd := model.Init()

	assert.NotNil(t, cmd, "Init should return a command")
}

// TestNewsModel_Update_Navigation verifies j/k and arrow key navigation.
func TestNewsModel_Update_Navigation(t *testing.T) {
	model := newTestModel()

	// Add test articles
	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1"}},
		{Item: &alphavantage.NewsItem{Title: "Article 2"}},
		{Item: &alphavantage.NewsItem{Title: "Article 3"}},
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.state = StateLoaded

	// Test 'j' key moves down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Equal(t, 1, newNewsModel.activeIndex, "j key should move active index down")

	// Test 'k' key moves up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = newNewsModel.Update(msg)
	newNewsModel = newModel.(Model)
	assert.Equal(t, 0, newNewsModel.activeIndex, "k key should move active index up")

	// Test arrow down
	msg = tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ = newNewsModel.Update(msg)
	newNewsModel = newModel.(Model)
	assert.Equal(t, 1, newNewsModel.activeIndex, "arrow down should move active index down")

	// Test arrow up
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = newNewsModel.Update(msg)
	newNewsModel = newModel.(Model)
	assert.Equal(t, 0, newNewsModel.activeIndex, "arrow up should move active index up")
}

// TestNewsModel_Update_NavigationBounds verifies navigation respects bounds.
func TestNewsModel_Update_NavigationBounds(t *testing.T) {
	model := newTestModel()

	// Add test articles
	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1"}},
		{Item: &alphavantage.NewsItem{Title: "Article 2"}},
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.state = StateLoaded

	// Test can't go below 0
	model.activeIndex = 0
	msg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Equal(t, 0, newNewsModel.activeIndex, "Should not go below 0")

	// Test can't go above max
	model.activeIndex = 1
	msg = tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ = model.Update(msg)
	newNewsModel = newModel.(Model)
	assert.Equal(t, 1, newNewsModel.activeIndex, "Should not go above max")
}

// TestNewsModel_Update_FilterToggle verifies filter cycling.
func TestNewsModel_Update_FilterToggle(t *testing.T) {
	model := newTestModel()

	// Add test articles
	articles := []Article{
		{
			Item: &alphavantage.NewsItem{
				Title: "Article 1",
				Topic: "technology",
			},
			Tickers: []string{"AAPL"},
		},
		{
			Item: &alphavantage.NewsItem{
				Title: "Article 2",
				Topic: "blockchain",
			},
			Tickers: []string{"BTC"},
		},
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.state = StateLoaded

	// Test 'f' key cycles filter
	assert.Equal(t, FilterAll, model.filter, "Initial filter should be All")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	newModel, _ := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Equal(t, FilterEquities, newNewsModel.filter, "f key should cycle to equities")

	newModel, _ = newNewsModel.Update(msg)
	newNewsModel = newModel.(Model)
	assert.Equal(t, FilterCrypto, newNewsModel.filter, "f key should cycle to crypto")

	newModel, _ = newNewsModel.Update(msg)
	newNewsModel = newModel.(Model)
	assert.Equal(t, FilterMacro, newNewsModel.filter, "f key should cycle to macro")

	newModel, _ = newNewsModel.Update(msg)
	newNewsModel = newModel.(Model)
	assert.Equal(t, FilterAll, newNewsModel.filter, "f key should cycle back to all")
}

// TestNewsModel_Update_SortToggle verifies sort cycling.
func TestNewsModel_Update_SortToggle(t *testing.T) {
	model := newTestModel()

	// Add test articles
	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1", TimePublished: "20250101T120000"}},
		{Item: &alphavantage.NewsItem{Title: "Article 2", TimePublished: "20250102T120000"}},
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.state = StateLoaded

	// Test 's' key cycles sort
	assert.Equal(t, SortNewest, model.sort, "Initial sort should be Newest")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	newModel, _ := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Equal(t, SortScore, newNewsModel.sort, "s key should cycle to score")

	newModel, _ = newNewsModel.Update(msg)
	newNewsModel = newModel.(Model)
	assert.Equal(t, SortNewest, newNewsModel.sort, "s key should cycle back to newest")
}

// TestNewsModel_Update_OpenArticle verifies opening article returns command.
func TestNewsModel_Update_OpenArticle(t *testing.T) {
	model := newTestModel()

	// Add test articles
	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1", URL: "https://example.com/article1"}},
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.state = StateLoaded

	// Test Enter key opens article
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Equal(t, 0, newNewsModel.activeIndex, "Active index should remain unchanged")
	assert.NotNil(t, cmd, "Enter key should return a command")
}

// TestNewsModel_Update_OpenArticleNoArticles verifies no command when no articles.
func TestNewsModel_Update_OpenArticleNoArticles(t *testing.T) {
	model := newTestModel()
	model.articles = []Article{}
	model.state = StateLoaded

	// Test Enter key with no articles
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Nil(t, cmd, "Enter key with no articles should return nil command")
	// Compare specific fields since ctx/cancel are pointers and will differ
	assert.Equal(t, model.articles, newNewsModel.articles, "Articles should remain unchanged")
	assert.Equal(t, model.state, newNewsModel.state, "State should remain unchanged")
	assert.Equal(t, model.activeIndex, newNewsModel.activeIndex, "Active index should remain unchanged")
}

// TestNewsModel_Update_Refresh verifies refresh behavior.
func TestNewsModel_Update_Refresh(t *testing.T) {
	model := newTestModel()

	// Test 'r' key triggers refresh
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, cmd := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Equal(t, StateLoading, newNewsModel.state, "r key should set state to loading")
	assert.NotNil(t, cmd, "r key should return a command")

	// Test RefreshMsg
	model.state = StateLoaded
	msg2 := RefreshMsg{}
	newModel, cmd = model.Update(msg2)
	newNewsModel = newModel.(Model)
	assert.Equal(t, StateLoading, newNewsModel.state, "RefreshMsg should set state to loading")
	assert.NotNil(t, cmd, "RefreshMsg should return a command")
}

// TestNewsModel_Update_WindowSize verifies window size handling.
func TestNewsModel_Update_WindowSize(t *testing.T) {
	model := newTestModel()

	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	newModel, _ := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Equal(t, 100, newNewsModel.width, "Width should be updated")
	assert.Equal(t, 30, newNewsModel.height, "Height should be updated")
	assert.Equal(t, 26, newNewsModel.viewportHeight, "Viewport height should be calculated")
}

// TestNewsModel_Update_DataMsg verifies news data handling.
func TestNewsModel_Update_DataMsg(t *testing.T) {
	model := newTestModel()
	model.state = StateLoading

	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1"}},
		{Item: &alphavantage.NewsItem{Title: "Article 2"}},
	}
	msg := DataMsg{
		Articles:  articles,
		Timestamp: time.Now(),
	}

	newModel, _ := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Equal(t, StateLoaded, newNewsModel.state, "State should be loaded")
	assert.Len(t, newNewsModel.allArticles, 2, "All articles should be set")
	assert.Len(t, newNewsModel.articles, 2, "Articles should be set after filtering")
	assert.Equal(t, 0, newNewsModel.activeIndex, "Active index should be reset")
	assert.Equal(t, 0, newNewsModel.scrollOffset, "Scroll offset should be reset")
}

// TestNewsModel_Update_ErrorMsg verifies error handling.
func TestNewsModel_Update_ErrorMsg(t *testing.T) {
	model := newTestModel()
	model.state = StateLoading

	err := errors.New("API error")
	msg := ErrorMsg{Err: err}

	newModel, _ := model.Update(msg)
	newNewsModel := newModel.(Model)
	assert.Equal(t, StateError, newNewsModel.state, "State should be error")
	assert.Equal(t, err, newNewsModel.err, "Error should be set")
}

// TestNewsModel_filterArticles_FilterAll verifies all filter returns all articles.
func TestNewsModel_filterArticles_FilterAll(t *testing.T) {
	model := newTestModel()
	model.filter = FilterAll

	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1"}, Tickers: []string{"AAPL"}},
		{Item: &alphavantage.NewsItem{Title: "Article 2"}, Tickers: []string{"BTC"}},
	}

	result := model.filterArticles(articles)
	assert.Len(t, result, 2, "All filter should return all articles")
}

// TestNewsModel_filterArticles_FilterEquities verifies equities filter.
func TestNewsModel_filterArticles_FilterEquities(t *testing.T) {
	model := newTestModel()
	model.filter = FilterEquities

	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1"}, Tickers: []string{"AAPL"}},
		{Item: &alphavantage.NewsItem{Title: "Article 2"}, Tickers: []string{"BTC"}},
		{Item: &alphavantage.NewsItem{Title: "Article 3"}, Tickers: []string{"MSFT"}},
	}

	result := model.filterArticles(articles)
	assert.Len(t, result, 2, "Equities filter should return only equity articles")
}

// TestNewsModel_filterArticles_FilterCrypto verifies crypto filter.
func TestNewsModel_filterArticles_FilterCrypto(t *testing.T) {
	model := newTestModel()
	model.filter = FilterCrypto

	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1"}, Tickers: []string{"AAPL"}},
		{Item: &alphavantage.NewsItem{Title: "Article 2"}, Tickers: []string{"BTC"}},
		{Item: &alphavantage.NewsItem{Title: "Article 3"}, Tickers: []string{"ETH"}},
	}

	result := model.filterArticles(articles)
	assert.Len(t, result, 2, "Crypto filter should return only crypto articles")
}

// TestNewsModel_filterArticles_FilterMacro verifies macro filter.
func TestNewsModel_filterArticles_FilterMacro(t *testing.T) {
	model := newTestModel()
	model.filter = FilterMacro

	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1", Topic: "technology"}},
		{Item: &alphavantage.NewsItem{Title: "Article 2", Topic: "macro_economy"}},
		{Item: &alphavantage.NewsItem{Title: "Article 3", Topic: "employment_report"}},
	}

	result := model.filterArticles(articles)
	assert.Len(t, result, 2, "Macro filter should return only macro articles")
}

// TestNewsModel_sortArticles_SortNewest verifies newest sort.
func TestNewsModel_sortArticles_SortNewest(t *testing.T) {
	model := newTestModel()
	model.sort = SortNewest

	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1", TimePublished: "20250101T120000"}},
		{Item: &alphavantage.NewsItem{Title: "Article 2", TimePublished: "20250103T120000"}},
		{Item: &alphavantage.NewsItem{Title: "Article 3", TimePublished: "20250102T120000"}},
	}

	model.sortArticles(articles)
	assert.Equal(t, "Article 2", articles[0].Item.Title, "First should be newest")
	assert.Equal(t, "Article 3", articles[1].Item.Title, "Second should be middle")
	assert.Equal(t, "Article 1", articles[2].Item.Title, "Third should be oldest")
}

// TestNewsModel_sortArticles_SortScore verifies score sort.
func TestNewsModel_sortArticles_SortScore(t *testing.T) {
	model := newTestModel()
	model.sort = SortScore

	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1"}, SentimentScore: 0.3},
		{Item: &alphavantage.NewsItem{Title: "Article 2"}, SentimentScore: 0.9},
		{Item: &alphavantage.NewsItem{Title: "Article 3"}, SentimentScore: 0.6},
	}

	model.sortArticles(articles)
	assert.Equal(t, 0.9, articles[0].SentimentScore, "First should be highest score")
	assert.Equal(t, 0.6, articles[1].SentimentScore, "Second should be middle score")
	assert.Equal(t, 0.3, articles[2].SentimentScore, "Third should be lowest score")
}

// TestNewsModel_scrollWithViewport verifies scrolling behavior.
func TestNewsModel_scrollWithViewport(t *testing.T) {
	model := newTestModel()
	model.width = 80
	model.height = 20
	model.calculateViewportHeight()

	// Add more articles than viewport
	articles := make([]Article, 20)
	for i := range articles {
		articles[i] = Article{
			Item: &alphavantage.NewsItem{Title: string(rune('A' + i))},
		}
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.state = StateLoaded

	// Test scrolling down to trigger scroll offset
	// With viewportHeight = 16 (20 - 4), scroll triggers when activeIndex >= scrollOffset + viewportHeight
	// So at index 16 with scrollOffset=0: 16 >= 16, scroll triggers and scrollOffset = 16 - 16 + 1 = 1
	model.activeIndex = 0
	msg := tea.KeyMsg{Type: tea.KeyDown}
	for i := 0; i < 15; i++ {
		newModel, _ := model.Update(msg)
		model = newModel.(Model)
	}
	assert.Equal(t, 15, model.activeIndex, "Should move to index 15")
	assert.Equal(t, 0, model.scrollOffset, "Scroll offset should remain 0 at index 15")

	newModel, _ := model.Update(msg)
	model = newModel.(Model)
	assert.Equal(t, 16, model.activeIndex, "Should move to index 16")
	assert.Equal(t, 1, model.scrollOffset, "Scroll offset should move to 1 at index 16")

	// Test scrolling up
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)
	assert.Equal(t, 15, model.activeIndex, "Should move up to index 15")
	assert.Equal(t, 1, model.scrollOffset, "Scroll offset should remain 1 at index 15")

	// Continue scrolling up until scroll offset changes
	for i := 0; i < 14; i++ {
		newModel, _ = model.Update(msg)
		model = newModel.(Model)
	}
	assert.Equal(t, 1, model.activeIndex, "Should move up to index 1")
	assert.Equal(t, 1, model.scrollOffset, "Scroll offset should remain 1 at index 1")

	newModel, _ = model.Update(msg)
	model = newModel.(Model)
	assert.Equal(t, 0, model.activeIndex, "Should move up to index 0")
	assert.Equal(t, 0, model.scrollOffset, "Scroll offset should move to 0 at index 0")
}

// TestNewsModel_convertToArticle verifies article conversion.
func TestNewsModel_convertToArticle(t *testing.T) {
	item := alphavantage.NewsItem{
		Title:                 "Test Article",
		URL:                   "https://example.com/test",
		TimePublished:         "20250101T120000",
		OverallSentimentScore: "0.5",
		Tickers: []alphavantage.TickerSentiment{
			{Ticker: "AAPL", RelevanceScore: "0.8", TickerSentimentScore: "0.7"},
			{Ticker: "MSFT", RelevanceScore: "0.6", TickerSentimentScore: "0.5"},
		},
	}

	article := convertToArticle(item)
	assert.Equal(t, "Test Article", article.Item.Title)
	assert.Equal(t, "https://example.com/test", article.Item.URL)
	assert.Contains(t, article.Tickers, "AAPL")
	assert.Contains(t, article.Tickers, "MSFT")
	assert.InDelta(t, 0.7, article.RelevanceScore, 0.01)
	assert.InDelta(t, 0.6, article.SentimentScore, 0.01)
}

// TestNewsModel_isCryptoTicker verifies crypto ticker detection.
func TestNewsModel_isCryptoTicker(t *testing.T) {
	tests := []struct {
		ticker   string
		expected bool
	}{
		{"BTC", true},
		{"ETH", true},
		{"SOL", true},
		{"DOGE", true},
		{"AAPL", false},
		{"MSFT", false},
		{"GOOGL", false},
		{"btc", true}, // case insensitive
		{"eth", true}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.ticker, func(t *testing.T) {
			result := isCryptoTicker(tt.ticker)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilter_String verifies filter string representation.
func TestFilter_String(t *testing.T) {
	tests := []struct {
		filter Filter
		want   string
	}{
		{FilterAll, "all"},
		{FilterEquities, "equities"},
		{FilterCrypto, "crypto"},
		{FilterMacro, "macro"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filter.String())
		})
	}
}

// TestFilter_Next verifies filter cycling.
func TestFilter_Next(t *testing.T) {
	tests := []struct {
		current Filter
		next    Filter
	}{
		{FilterAll, FilterEquities},
		{FilterEquities, FilterCrypto},
		{FilterCrypto, FilterMacro},
		{FilterMacro, FilterAll},
	}

	for _, tt := range tests {
		t.Run(tt.current.String(), func(t *testing.T) {
			assert.Equal(t, tt.next, tt.current.Next())
		})
	}
}

// TestSort_String verifies sort string representation.
func TestSort_String(t *testing.T) {
	tests := []struct {
		sort Sort
		want string
	}{
		{SortNewest, "newest"},
		{SortScore, "score"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sort.String())
		})
	}
}

// TestSort_Next verifies sort cycling.
func TestSort_Next(t *testing.T) {
	tests := []struct {
		current Sort
		next    Sort
	}{
		{SortNewest, SortScore},
		{SortScore, SortNewest},
	}

	for _, tt := range tests {
		t.Run(tt.current.String(), func(t *testing.T) {
			assert.Equal(t, tt.next, tt.current.Next())
		})
	}
}

// TestState_String verifies state string representation.
func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateLoading, "Loading"},
		{StateLoaded, "Loaded"},
		{StateError, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.state.String())
		})
	}
}

// TestNewsModel_Getters verifies all getter methods.
func TestNewsModel_Getters(t *testing.T) {
	model := newTestModel()

	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1"}},
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.activeIndex = 0
	model.filter = FilterCrypto
	model.sort = SortScore
	model.state = StateLoaded
	model.err = errors.New("test error")
	model.lastRefresh = time.Now()
	model.width = 100
	model.height = 30

	assert.Equal(t, articles, model.GetArticles())
	assert.NotNil(t, model.GetActiveArticle())
	assert.Equal(t, 0, model.GetActiveIndex())
	assert.Equal(t, FilterCrypto, model.GetFilter())
	assert.Equal(t, SortScore, model.GetSort())
	assert.Equal(t, StateLoaded, model.GetState())
	assert.Equal(t, errors.New("test error"), model.GetError())
	assert.False(t, model.GetLastRefresh().IsZero())
	assert.Equal(t, 100, model.GetWidth())
	assert.Equal(t, 30, model.GetHeight())
}

// TestNewsModel_GetActiveArticle_NoArticles verifies behavior with no articles.
func TestNewsModel_GetActiveArticle_NoArticles(t *testing.T) {
	model := newTestModel()
	model.articles = []Article{}

	assert.Nil(t, model.GetActiveArticle(), "Should return nil when no articles")
}

// TestNewsModel_GetActiveArticle_InvalidIndex verifies behavior with invalid index.
func TestNewsModel_GetActiveArticle_InvalidIndex(t *testing.T) {
	model := newTestModel()
	articles := []Article{
		{Item: &alphavantage.NewsItem{Title: "Article 1"}},
	}
	model.allArticles = articles
	model.applyFilterAndSort()
	model.activeIndex = 10 // Invalid index

	assert.Nil(t, model.GetActiveArticle(), "Should return nil when index is out of bounds")
}
