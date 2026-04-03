// Package alphavantage provides a client for the Alpha Vantage API.
// This file implements the news and sentiment endpoint.
package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	// function names for Alpha Vantage news endpoint
	functionNewsSentiment = "NEWS_SENTIMENT"
	// DefaultNewsLimit is the default number of articles to return.
	DefaultNewsLimit = 50
	// MaxNewsLimit is the maximum number of articles the API allows per request.
	MaxNewsLimit = 1000
)

// NewsOpts contains options for filtering news sentiment requests.
type NewsOpts struct {
	// Tickers is a list of ticker symbols to filter news for (e.g., "AAPL", "MSFT").
	Tickers []string
	// Topics is a list of topics to filter news for (e.g., "technology", "finance").
	Topics []string
	// Sort specifies the sort order: "LATEST", "EARLIEST", or "RELEVANCE".
	Sort string
	// Limit is the maximum number of articles to return (default: 50, max: 1000).
	Limit int
}

// GetNewsSentiment fetches news articles with sentiment analysis.
// The opts parameter allows filtering by tickers, topics, sort order, and result limit.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetNewsSentiment(ctx context.Context, opts NewsOpts) (*NewsSentiment, error) {
	// Set default limit
	if opts.Limit <= 0 {
		opts.Limit = DefaultNewsLimit
	}
	// Validate limit
	if opts.Limit > MaxNewsLimit {
		return nil, fmt.Errorf("limit cannot exceed %d, got %d", MaxNewsLimit, opts.Limit)
	}

	params := map[string]string{
		"function": functionNewsSentiment,
		"limit":    strconv.Itoa(opts.Limit),
	}

	// Add tickers filter if provided
	if len(opts.Tickers) > 0 {
		params["tickers"] = strings.Join(opts.Tickers, ",")
	}

	// Add topics filter if provided
	if len(opts.Topics) > 0 {
		params["topics"] = strings.Join(opts.Topics, ",")
	}

	// Add sort order if provided
	if opts.Sort != "" {
		// Validate sort order
		validSorts := map[string]bool{
			"LATEST":       true,
			"EARLIEST":     true,
			"RELEVANCE":    true,
			"most_recent":  true, // Alternative naming supported by API
			"least_recent": true,
		}
		if !validSorts[opts.Sort] {
			return nil, fmt.Errorf("invalid sort order %s, must be one of: LATEST, EARLIEST, RELEVANCE", opts.Sort)
		}
		params["sort"] = opts.Sort
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching news sentiment: %w", err)
	}

	var response NewsSentiment
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing news sentiment response: %w", err)
	}

	// Sort articles by published time (most recent first) for consistent ordering
	sort.Slice(response.Items, func(i, j int) bool {
		return strings.Compare(response.Items[i].TimePublished, response.Items[j].TimePublished) > 0
	})

	return &response, nil
}
