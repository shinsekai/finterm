package alphavantage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetNewsSentiment_NoFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifyNoFilterParams(t, r)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, sampleNewsResponse)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	result, err := client.GetNewsSentiment(context.Background(), NewsOpts{})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 news items, got %d", len(result.Items))
	}

	// Verify first article (should be sorted by time, most recent first)
	item := result.Items[0]
	verifyFirstArticleFields(t, item)
}

func verifyNoFilterParams(t *testing.T, r *http.Request) {
	t.Helper()
	if r.URL.Query().Get("function") != functionNewsSentiment {
		t.Errorf("expected function %s, got %s", functionNewsSentiment, r.URL.Query().Get("function"))
	}
	if r.URL.Query().Get("limit") != fmt.Sprintf("%d", DefaultNewsLimit) {
		t.Errorf("expected limit %d, got %s", DefaultNewsLimit, r.URL.Query().Get("limit"))
	}
	if r.URL.Query().Get("tickers") != "" {
		t.Errorf("expected no tickers param, got %s", r.URL.Query().Get("tickers"))
	}
	if r.URL.Query().Get("topics") != "" {
		t.Errorf("expected no topics param, got %s", r.URL.Query().Get("topics"))
	}
	if r.URL.Query().Get("sort") != "" {
		t.Errorf("expected no sort param, got %s", r.URL.Query().Get("sort"))
	}
}

func verifyFirstArticleFields(t *testing.T, item NewsItem) {
	t.Helper()
	if item.Title != "Apple reports strong quarterly earnings" {
		t.Errorf("expected title 'Apple reports strong quarterly earnings', got %s", item.Title)
	}
	if item.URL != "https://example.com/apple-earnings" {
		t.Errorf("expected URL 'https://example.com/apple-earnings', got %s", item.URL)
	}
	if item.Source != "Reuters" {
		t.Errorf("expected source 'Reuters', got %s", item.Source)
	}
	if item.OverallSentimentScore != "0.234567" {
		t.Errorf("expected sentiment score '0.234567', got %s", item.OverallSentimentScore)
	}
	if len(item.Tickers) != 1 {
		t.Fatalf("expected 1 ticker sentiment, got %d", len(item.Tickers))
	}
	if item.Tickers[0].Ticker != "AAPL" {
		t.Errorf("expected ticker 'AAPL', got %s", item.Tickers[0].Ticker)
	}
	if item.Tickers[0].RelevanceScore != "0.876543" {
		t.Errorf("expected relevance score '0.876543', got %s", item.Tickers[0].RelevanceScore)
	}
	if item.Tickers[0].TickerSentimentScore != "0.345678" {
		t.Errorf("expected ticker sentiment score '0.345678', got %s", item.Tickers[0].TickerSentimentScore)
	}
}

const sampleNewsResponse = `{
	"feed": [
		{
			"title": "Apple reports strong quarterly earnings",
			"url": "https://example.com/apple-earnings",
			"time_published": "20240402T153000",
			"authors": ["John Doe", "Jane Smith"],
			"summary": "Apple Inc. reported better-than-expected earnings for Q1 2024.",
			"banner_image": "https://example.com/images/apple.jpg",
			"source": "Reuters",
			"category_within": "technology",
			"topic": "technology",
			"overall_sentiment_score": "0.234567",
			"sentiment_label": "Bullish",
			"ticker_sentiment": [
				{
					"ticker": "AAPL",
					"relevance_score": "0.876543",
					"ticker_sentiment_score": "0.345678",
					"ticker_sentiment_label": "Bullish"
				}
			]
		},
		{
			"title": "Microsoft launches new AI features",
			"url": "https://example.com/microsoft-ai",
			"time_published": "20240402T140000",
			"authors": ["Tech Reporter"],
			"summary": "Microsoft announced new AI-powered features for Office.",
			"banner_image": "https://example.com/images/msft.jpg",
			"source": "TechCrunch",
			"category_within": "technology",
			"topic": "technology",
			"overall_sentiment_score": "0.345678",
			"sentiment_label": "Bullish",
			"ticker_sentiment": [
				{
					"ticker": "MSFT",
					"relevance_score": "0.987654",
					"ticker_sentiment_score": "0.456789",
					"ticker_sentiment_label": "Bullish"
				}
			]
		}
	]
}`

func TestGetNewsSentiment_TickerFilter(t *testing.T) {
	tests := []struct {
		name    string
		tickers []string
		expect  string
	}{
		{"single ticker", []string{"AAPL"}, "AAPL"},
		{"multiple tickers", []string{"AAPL", "MSFT", "GOOGL"}, "AAPL,MSFT,GOOGL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tickers := r.URL.Query().Get("tickers")
				if tickers != tt.expect {
					t.Errorf("expected tickers param %q, got %q", tt.expect, tickers)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{
					"feed": [
						{
							"title": "Stock market update",
							"url": "https://example.com/stock-update",
							"time_published": "20240402T120000",
							"authors": ["Finance Desk"],
							"summary": "Market analysis for today.",
							"banner_image": "https://example.com/images/stock.jpg",
							"source": "Bloomberg",
							"category_within": "finance",
							"topic": "financial_markets",
							"overall_sentiment_score": "0.123456",
							"sentiment_label": "Neutral",
							"ticker_sentiment": [
								{
									"ticker": "AAPL",
									"relevance_score": "0.9",
									"ticker_sentiment_score": "0.2",
									"ticker_sentiment_label": "Neutral"
								}
							]
						}
					]
				}`)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			})

			_, err := client.GetNewsSentiment(context.Background(), NewsOpts{
				Tickers: tt.tickers,
			})
			if err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
		})
	}
}

func TestGetNewsSentiment_TopicFilter(t *testing.T) {
	tests := []struct {
		name   string
		topics []string
		expect string
	}{
		{"single topic", []string{"technology"}, "technology"},
		{"multiple topics", []string{"technology", "finance", "healthcare"}, "technology,finance,healthcare"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				topics := r.URL.Query().Get("topics")
				if topics != tt.expect {
					t.Errorf("expected topics param %q, got %q", tt.expect, topics)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{
					"feed": [
						{
							"title": "Tech industry news",
							"url": "https://example.com/tech-news",
							"time_published": "20240402T100000",
							"authors": ["Tech Writer"],
							"summary": "Latest developments in technology.",
							"banner_image": "https://example.com/images/tech.jpg",
							"source": "Wired",
							"category_within": "technology",
							"topic": "technology",
							"overall_sentiment_score": "0.456789",
							"sentiment_label": "Bullish",
							"ticker_sentiment": []
						}
					]
				}`)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			})

			_, err := client.GetNewsSentiment(context.Background(), NewsOpts{
				Topics: tt.topics,
			})
			if err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
		})
	}
}

func TestGetNewsSentiment_CombinedFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all filters are applied
		if r.URL.Query().Get("tickers") != "AAPL,MSFT" {
			t.Errorf("expected tickers 'AAPL,MSFT', got %s", r.URL.Query().Get("tickers"))
		}
		if r.URL.Query().Get("topics") != "technology" {
			t.Errorf("expected topics 'technology', got %s", r.URL.Query().Get("topics"))
		}
		if r.URL.Query().Get("sort") != "LATEST" {
			t.Errorf("expected sort 'LATEST', got %s", r.URL.Query().Get("sort"))
		}
		if r.URL.Query().Get("limit") != "100" {
			t.Errorf("expected limit '100', got %s", r.URL.Query().Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"feed": [
				{
					"title": "Apple and Microsoft AI collaboration",
					"url": "https://example.com/ai-collab",
					"time_published": "20240402T160000",
					"authors": ["AI Reporter"],
					"summary": "Tech giants partner on AI initiatives.",
					"banner_image": "https://example.com/images/ai.jpg",
					"source": "The Verge",
					"category_within": "technology",
					"topic": "technology",
					"overall_sentiment_score": "0.567890",
					"sentiment_label": "Bullish",
					"ticker_sentiment": [
						{
							"ticker": "AAPL",
							"relevance_score": "0.95",
							"ticker_sentiment_score": "0.5",
							"ticker_sentiment_label": "Bullish"
						},
						{
							"ticker": "MSFT",
							"relevance_score": "0.92",
							"ticker_sentiment_score": "0.55",
							"ticker_sentiment_label": "Bullish"
						}
					]
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	result, err := client.GetNewsSentiment(context.Background(), NewsOpts{
		Tickers: []string{"AAPL", "MSFT"},
		Topics:  []string{"technology"},
		Sort:    "LATEST",
		Limit:   100,
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(result.Items) != 1 {
		t.Errorf("expected 1 news item, got %d", len(result.Items))
	}

	// Verify multiple ticker sentiments
	if len(result.Items[0].Tickers) != 2 {
		t.Fatalf("expected 2 ticker sentiments, got %d", len(result.Items[0].Tickers))
	}
}

func TestGetNewsSentiment_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"feed": []}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	result, err := client.GetNewsSentiment(context.Background(), NewsOpts{})
	if err != nil {
		t.Fatalf("expected success for empty result, got error: %v", err)
	}

	if result.Items == nil {
		t.Error("expected Items to be an empty slice, got nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("expected 0 news items, got %d", len(result.Items))
	}
}

func TestGetNewsSentiment_ArticleParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"feed": [
				{
					"title": "Comprehensive financial analysis: Markets in flux",
					"url": "https://finance.example.com/markets-flux",
					"time_published": "20240402T123045",
					"authors": ["Sarah Johnson", "Michael Chen", "Emma Davis"],
					"summary": "A detailed look at current market conditions, sector performance, and investor sentiment across global markets.",
					"banner_image": "https://images.example.com/markets-banner.jpg",
					"source": "Financial Times",
					"category_within": "markets",
					"topic": "financial_markets",
					"overall_sentiment_score": "-0.123456",
					"sentiment_label": "Bearish",
					"ticker_sentiment": [
						{
							"ticker": "AAPL",
							"relevance_score": "0.789012",
							"ticker_sentiment_score": "-0.234567",
							"ticker_sentiment_label": "Bearish"
						},
						{
							"ticker": "MSFT",
							"relevance_score": "0.654321",
							"ticker_sentiment_score": "0.123456",
							"ticker_sentiment_label": "Bullish"
						},
						{
							"ticker": "GOOGL",
							"relevance_score": "0.567890",
							"ticker_sentiment_score": "-0.098765",
							"ticker_sentiment_label": "Neutral"
						}
					]
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	result, err := client.GetNewsSentiment(context.Background(), NewsOpts{})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 news item, got %d", len(result.Items))
	}

	item := result.Items[0]

	// Test all article fields are parsed correctly
	tests := []struct {
		field string
		got   string
		want  string
	}{
		{"title", item.Title, "Comprehensive financial analysis: Markets in flux"},
		{"url", item.URL, "https://finance.example.com/markets-flux"},
		{"time_published", item.TimePublished, "20240402T123045"},
		{"source", item.Source, "Financial Times"},
		{"topic", item.Topic, "financial_markets"},
		{"category_within", item.CategoryWithin, "markets"},
		{"overall_sentiment_score", item.OverallSentimentScore, "-0.123456"},
		{"sentiment_label", item.SentimentLabel, "Bearish"},
		{"banner_image", item.BannerImage, "https://images.example.com/markets-banner.jpg"},
		{"summary", item.Summary, "A detailed look at current market conditions, sector performance, and investor sentiment across global markets."},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %s %q, want %q", tt.field, tt.got, tt.want)
			}
		})
	}

	// Test authors array
	if len(item.Authors) != 3 {
		t.Fatalf("expected 3 authors, got %d", len(item.Authors))
	}
	expectedAuthors := []string{"Sarah Johnson", "Michael Chen", "Emma Davis"}
	for i, author := range item.Authors {
		if author != expectedAuthors[i] {
			t.Errorf("author[%d]: got %q, want %q", i, author, expectedAuthors[i])
		}
	}

	// Test ticker sentiments array
	if len(item.Tickers) != 3 {
		t.Fatalf("expected 3 ticker sentiments, got %d", len(item.Tickers))
	}

	expectedTickers := []struct {
		ticker               string
		relevanceScore       string
		tickerSentimentScore string
		tickerSentimentLabel string
	}{
		{"AAPL", "0.789012", "-0.234567", "Bearish"},
		{"MSFT", "0.654321", "0.123456", "Bullish"},
		{"GOOGL", "0.567890", "-0.098765", "Neutral"},
	}

	for i, tt := range expectedTickers {
		t.Run(fmt.Sprintf("ticker_%s", tt.ticker), func(t *testing.T) {
			ts := item.Tickers[i]
			if ts.Ticker != tt.ticker {
				t.Errorf("ticker: got %q, want %q", ts.Ticker, tt.ticker)
			}
			if ts.RelevanceScore != tt.relevanceScore {
				t.Errorf("relevance_score: got %q, want %q", ts.RelevanceScore, tt.relevanceScore)
			}
			if ts.TickerSentimentScore != tt.tickerSentimentScore {
				t.Errorf("ticker_sentiment_score: got %q, want %q", ts.TickerSentimentScore, tt.tickerSentimentScore)
			}
			if ts.TickerSentimentLabel != tt.tickerSentimentLabel {
				t.Errorf("ticker_sentiment_label: got %q, want %q", ts.TickerSentimentLabel, tt.tickerSentimentLabel)
			}
		})
	}
}

func TestGetNewsSentiment_InvalidSort(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetNewsSentiment(context.Background(), NewsOpts{
		Sort: "INVALID_SORT",
	})
	if err == nil {
		t.Fatal("expected error for invalid sort order, got nil")
	}

	if !strings.Contains(err.Error(), "invalid sort order") {
		t.Errorf("expected invalid sort order error, got: %v", err)
	}
}

func TestGetNewsSentiment_LimitExceeded(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetNewsSentiment(context.Background(), NewsOpts{
		Limit: 2000,
	})
	if err == nil {
		t.Fatal("expected error for limit exceeded, got nil")
	}

	if !strings.Contains(err.Error(), "limit cannot exceed") {
		t.Errorf("expected limit exceeded error, got: %v", err)
	}
}

func TestGetNewsSentiment_DifferentSortOrders(t *testing.T) {
	tests := []struct {
		name string
		sort string
	}{
		{"LATEST", "LATEST"},
		{"EARLIEST", "EARLIEST"},
		{"RELEVANCE", "RELEVANCE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("sort") != tt.sort {
					t.Errorf("expected sort %s, got %s", tt.sort, r.URL.Query().Get("sort"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{"feed": []}`)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			})

			_, err := client.GetNewsSentiment(context.Background(), NewsOpts{
				Sort: tt.sort,
			})
			if err != nil {
				t.Fatalf("expected success for sort %s, got error: %v", tt.sort, err)
			}
		})
	}
}

func TestGetNewsSentiment_DifferentLimits(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{"default limit", 0},
		{"custom limit 10", 10},
		{"custom limit 100", 100},
		{"custom limit 1000", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedLimit := fmt.Sprintf("%d", tt.limit)
				if tt.limit == 0 {
					expectedLimit = fmt.Sprintf("%d", DefaultNewsLimit)
				}

				if r.URL.Query().Get("limit") != expectedLimit {
					t.Errorf("expected limit %s, got %s", expectedLimit, r.URL.Query().Get("limit"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{"feed": []}`)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			})

			_, err := client.GetNewsSentiment(context.Background(), NewsOpts{
				Limit: tt.limit,
			})
			if err != nil {
				t.Fatalf("expected success for limit %d, got error: %v", tt.limit, err)
			}
		})
	}
}

func TestGetNewsSentiment_ErrorInBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Error Message": "Invalid API call. Please retry or visit the documentation for NEWS_SENTIMENT."
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	_, err := client.GetNewsSentiment(context.Background(), NewsOpts{})
	if err == nil {
		t.Fatal("expected error for API error in body, got nil")
	}

	if !strings.Contains(err.Error(), "Invalid API call") {
		t.Errorf("expected API error message, got: %v", err)
	}
}

func TestGetNewsSentiment_RateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Note": "Thank you for using Alpha Vantage! Our standard API rate limit is 25 requests per day. Please subscribe to any of the premium plans at https://www.alphavantage.co/premium/ to instantly remove all daily rate limits."
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	_, err := client.GetNewsSentiment(context.Background(), NewsOpts{})
	if err == nil {
		t.Fatal("expected error for rate limit, got nil")
	}

	if !strings.Contains(err.Error(), "API note") {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}

func TestGetNewsSentiment_ContextPropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// This shouldn't be reached due to canceled context
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"feed": []}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	_, err := client.GetNewsSentiment(ctx, NewsOpts{})
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestGetNewsSentiment_ArticleSorting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"feed": [
				{
					"title": "Oldest article",
					"url": "https://example.com/oldest",
					"time_published": "20240401T100000",
					"authors": ["Writer"],
					"summary": "Oldest article summary",
					"banner_image": "https://example.com/oldest.jpg",
					"source": "Source",
					"category_within": "tech",
					"topic": "technology",
					"overall_sentiment_score": "0.1",
					"sentiment_label": "Neutral",
					"ticker_sentiment": []
				},
				{
					"title": "Newest article",
					"url": "https://example.com/newest",
					"time_published": "20240403T100000",
					"authors": ["Writer"],
					"summary": "Newest article summary",
					"banner_image": "https://example.com/newest.jpg",
					"source": "Source",
					"category_within": "tech",
					"topic": "technology",
					"overall_sentiment_score": "0.1",
					"sentiment_label": "Neutral",
					"ticker_sentiment": []
				},
				{
					"title": "Middle article",
					"url": "https://example.com/middle",
					"time_published": "20240402T100000",
					"authors": ["Writer"],
					"summary": "Middle article summary",
					"banner_image": "https://example.com/middle.jpg",
					"source": "Source",
					"category_within": "tech",
					"topic": "technology",
					"overall_sentiment_score": "0.1",
					"sentiment_label": "Neutral",
					"ticker_sentiment": []
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	result, err := client.GetNewsSentiment(context.Background(), NewsOpts{})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(result.Items) != 3 {
		t.Fatalf("expected 3 news items, got %d", len(result.Items))
	}

	// Articles should be sorted by time_published descending (most recent first)
	if result.Items[0].Title != "Newest article" {
		t.Errorf("expected first item to be 'Newest article', got %s", result.Items[0].Title)
	}
	if result.Items[1].Title != "Middle article" {
		t.Errorf("expected second item to be 'Middle article', got %s", result.Items[1].Title)
	}
	if result.Items[2].Title != "Oldest article" {
		t.Errorf("expected third item to be 'Oldest article', got %s", result.Items[2].Title)
	}
}
