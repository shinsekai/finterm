// Package alphavantage provides a client for the Alpha Vantage API.
// This file implements technical indicator endpoints for equities.
package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	// function names for technical indicators
	functionRSI = "RSI"
	functionEMA = "EMA"
)

// GetRSI fetches the Relative Strength Index (RSI) technical indicator for a symbol.
// The interval parameter specifies the time interval (e.g., "1min", "5min", "15min", "30min", "60min", "daily", "weekly", "monthly").
// The period parameter specifies the time period for the indicator calculation (typically 14 for RSI).
// The series_type is always set to "close" as per project requirements.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetRSI(ctx context.Context, symbol, interval string, period int) (*RSIResponse, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	if interval == "" {
		return nil, fmt.Errorf("interval cannot be empty")
	}
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}

	params := map[string]string{
		"function":    functionRSI,
		"symbol":      symbol,
		"interval":    interval,
		"time_period": strconv.Itoa(period),
		"series_type": "close",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching RSI data for %s: %w", symbol, err)
	}

	var response RSIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing RSI response for %s: %w", symbol, err)
	}

	// Validate that we got data
	if len(response.TechnicalAnalysis) == 0 {
		return nil, fmt.Errorf("empty RSI data for symbol %s", symbol)
	}

	return &response, nil
}

// GetEMA fetches the Exponential Moving Average (EMA) technical indicator for a symbol.
// The interval parameter specifies the time interval (e.g., "1min", "5min", "15min", "30min", "60min", "daily", "weekly", "monthly").
// The period parameter specifies the time period for the indicator calculation (e.g., 9 for fast EMA, 21 for slow EMA).
// The series_type is always set to "close" as per project requirements.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetEMA(ctx context.Context, symbol, interval string, period int) (*EMAResponse, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	if interval == "" {
		return nil, fmt.Errorf("interval cannot be empty")
	}
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}

	params := map[string]string{
		"function":    functionEMA,
		"symbol":      symbol,
		"interval":    interval,
		"time_period": strconv.Itoa(period),
		"series_type": "close",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching EMA data for %s: %w", symbol, err)
	}

	var response EMAResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing EMA response for %s: %w", symbol, err)
	}

	// Validate that we got data
	if len(response.TechnicalAnalysis) == 0 {
		return nil, fmt.Errorf("empty EMA data for symbol %s", symbol)
	}

	return &response, nil
}
