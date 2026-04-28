// Package alphavantage provides a client for the Alpha Vantage API.
// This file implements the market status endpoint.
package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	// functionMarketStatus is the function name for the market status endpoint.
	functionMarketStatus = "MARKET_STATUS"
)

// GetMarketStatus fetches the current market status for all supported markets.
// The response includes market type, region, trading hours, and current status
// (open/closed) for equity, forex, commodity, and cryptocurrency markets.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetMarketStatus(ctx context.Context) (*MarketStatus, error) {
	params := map[string]string{
		"function": functionMarketStatus,
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching market status: %w", err)
	}

	var response MarketStatus
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing market status response: %w", err)
	}

	// Validate that we got data
	if response.Endpoint == "" {
		return nil, fmt.Errorf("empty market status response")
	}

	return &response, nil
}
