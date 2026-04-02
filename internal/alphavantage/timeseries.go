// Package alphavantage provides a client for the Alpha Vantage API.
// This file implements time series endpoints for equities and crypto.
package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// function names for Alpha Vantage API
	functionDailyTimeSeries    = "TIME_SERIES_DAILY"
	functionCryptoDaily        = "DIGITAL_CURRENCY_DAILY"
	functionCryptoIntraday     = "CRYPTO_INTRADAY"
	functionGlobalQuote        = "GLOBAL_QUOTE"
	functionRealtimeBulkQuotes = "REALTIME_BULK_QUOTES"
)

// GetDailyTimeSeries fetches daily time series data for an equity symbol.
// The outputsize parameter can be "compact" (100 data points) or "full" (20+ years of data).
// Context is propagated to the underlying HTTP call.
func (c *Client) GetDailyTimeSeries(ctx context.Context, symbol, outputsize string) (*TimeSeriesDaily, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	params := map[string]string{
		"function":   functionDailyTimeSeries,
		"symbol":     symbol,
		"outputsize": outputsize,
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching daily time series for %s: %w", symbol, err)
	}

	var response TimeSeriesDaily
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing daily time series response for %s: %w", symbol, err)
	}

	// Validate that we got data
	if len(response.TimeSeries) == 0 {
		return nil, fmt.Errorf("empty time series data for symbol %s", symbol)
	}

	return &response, nil
}

// GetCryptoDaily fetches daily time series data for a cryptocurrency.
// The symbol parameter is the digital currency code (e.g., "BTC").
// The market parameter is the physical market code (e.g., "USD").
// Context is propagated to the underlying HTTP call.
func (c *Client) GetCryptoDaily(ctx context.Context, symbol, market string) (*CryptoDaily, error) {
	if symbol == "" {
		return nil, fmt.Errorf("crypto symbol cannot be empty")
	}
	if market == "" {
		return nil, fmt.Errorf("market cannot be empty")
	}

	params := map[string]string{
		"function": functionCryptoDaily,
		"symbol":   symbol,
		"market":   market,
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching crypto daily data for %s/%s: %w", symbol, market, err)
	}

	var response CryptoDaily
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing crypto daily response for %s/%s: %w", symbol, market, err)
	}

	// Validate that we got data
	if len(response.TimeSeries) == 0 {
		return nil, fmt.Errorf("empty crypto time series data for %s/%s", symbol, market)
	}

	return &response, nil
}

// GetCryptoIntraday fetches intraday time series data for a cryptocurrency.
// The symbol parameter is the digital currency code (e.g., "BTC").
// The market parameter is the physical market code (e.g., "USD").
// The interval parameter can be "1min", "5min", "15min", "30min", "60min".
// Context is propagated to the underlying HTTP call.
func (c *Client) GetCryptoIntraday(ctx context.Context, symbol, market, interval string) (*CryptoIntraday, error) {
	if symbol == "" {
		return nil, fmt.Errorf("crypto symbol cannot be empty")
	}
	if market == "" {
		return nil, fmt.Errorf("market cannot be empty")
	}
	if interval == "" {
		return nil, fmt.Errorf("interval cannot be empty")
	}

	// Validate interval
	validIntervals := map[string]bool{
		"1min":  true,
		"5min":  true,
		"15min": true,
		"30min": true,
		"60min": true,
	}
	if !validIntervals[interval] {
		return nil, fmt.Errorf("invalid interval %s, must be one of: 1min, 5min, 15min, 30min, 60min", interval)
	}

	params := map[string]string{
		"function": functionCryptoIntraday,
		"symbol":   symbol,
		"market":   market,
		"interval": interval,
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching crypto intraday data for %s/%s: %w", symbol, market, err)
	}

	// Parse the response dynamically since the time series key varies by interval
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("parsing crypto intraday response for %s/%s: %w", symbol, market, err)
	}

	// Build the response structure
	response := &CryptoIntraday{}

	// Extract metadata
	if metaDataRaw, ok := rawResponse["Meta Data"].(map[string]interface{}); ok {
		response.MetaData = CryptoMetadata{
			Information:   getStringField(metaDataRaw, "1. Information"),
			DigitalCode:   getStringField(metaDataRaw, "2. Digital Currency Code"),
			DigitalName:   getStringField(metaDataRaw, "3. Digital Currency Name"),
			MarketCode:    getStringField(metaDataRaw, "4. Market Code"),
			MarketName:    getStringField(metaDataRaw, "5. Market Name"),
			LastRefreshed: getStringField(metaDataRaw, "6. Last Refreshed"),
			TimeZone:      getStringField(metaDataRaw, "7. Time Zone"),
		}
	}

	// Extract time series - the key is dynamic based on interval
	// Try both possible key formats
	expectedKey := fmt.Sprintf("Time Series Crypto (%s)", interval)
	timeSeriesFound := false

	// First try the exact key
	if timeSeriesRaw, ok := rawResponse[expectedKey]; ok {
		response.TimeSeries = parseCryptoTimeSeries(timeSeriesRaw)
		timeSeriesFound = true
	}

	// Validate that we got data
	if !timeSeriesFound || len(response.TimeSeries) == 0 {
		return nil, fmt.Errorf("empty crypto intraday data for %s/%s", symbol, market)
	}

	return response, nil
}

// getStringField extracts a string field from a map, returning empty string if not found.
func getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// parseCryptoTimeSeries parses the crypto time series from the raw JSON data.
func parseCryptoTimeSeries(raw interface{}) map[string]CryptoEntry {
	result := make(map[string]CryptoEntry)

	timeSeriesMap, ok := raw.(map[string]interface{})
	if !ok {
		return result
	}

	for timestamp, entryRaw := range timeSeriesMap {
		entryMap, ok := entryRaw.(map[string]interface{})
		if !ok {
			continue
		}

		result[timestamp] = CryptoEntry{
			OpenMarket:  getStringField(entryMap, "1a. open (USD)"),
			HighMarket:  getStringField(entryMap, "2a. high (USD)"),
			LowMarket:   getStringField(entryMap, "3a. low (USD)"),
			CloseMarket: getStringField(entryMap, "4a. close (USD)"),
			Volume:      getStringField(entryMap, "5. volume"),
			MarketCap:   getStringField(entryMap, "6. market cap (USD)"),
		}
	}

	return result
}

// GetGlobalQuote fetches the global quote for a single ticker symbol.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetGlobalQuote(ctx context.Context, symbol string) (*GlobalQuote, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	params := map[string]string{
		"function": functionGlobalQuote,
		"symbol":   symbol,
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching global quote for %s: %w", symbol, err)
	}

	var response GlobalQuoteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing global quote response for %s: %w", symbol, err)
	}

	// Validate that we got a quote
	if response.GlobalQuote == nil {
		return nil, fmt.Errorf("empty global quote data for symbol %s", symbol)
	}

	return response.GlobalQuote, nil
}

// GetBulkQuotes fetches real-time bulk quotes for multiple symbols.
// The API accepts up to 100 symbols per request, so this method batches
// symbols in groups of 100 and makes multiple requests if necessary.
// Context is propagated to the underlying HTTP calls.
//
// Partial failures are handled: if one batch fails, the error is returned
// and no quotes from that or subsequent batches are included.
func (c *Client) GetBulkQuotes(ctx context.Context, symbols []string) ([]GlobalQuote, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("symbols list cannot be empty")
	}

	// Remove any empty symbols
	var filteredSymbols []string
	for _, s := range symbols {
		if s != "" {
			filteredSymbols = append(filteredSymbols, s)
		}
	}
	if len(filteredSymbols) == 0 {
		return nil, fmt.Errorf("symbols list contains only empty strings")
	}

	const batchSize = 100
	var allQuotes []GlobalQuote

	// Process symbols in batches of 100
	for i := 0; i < len(filteredSymbols); i += batchSize {
		end := i + batchSize
		if end > len(filteredSymbols) {
			end = len(filteredSymbols)
		}
		batch := filteredSymbols[i:end]

		quotes, err := c.getBatchQuotes(ctx, batch)
		if err != nil {
			return allQuotes, fmt.Errorf("fetching bulk quotes for batch %d-%d: %w", i+1, end, err)
		}

		allQuotes = append(allQuotes, quotes...)
	}

	return allQuotes, nil
}

// getBatchQuotes fetches quotes for a single batch of symbols (max 100).
func (c *Client) getBatchQuotes(ctx context.Context, symbols []string) ([]GlobalQuote, error) {
	symbolsStr := strings.Join(symbols, ",")

	params := map[string]string{
		"function": functionRealtimeBulkQuotes,
		"symbols":  symbolsStr,
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, err
	}

	// Parse the response - bulk quotes returns an array
	var bulkResponse struct {
		Data []GlobalQuote `json:"data"`
	}
	if err := json.Unmarshal(body, &bulkResponse); err != nil {
		return nil, fmt.Errorf("parsing bulk quotes response: %w", err)
	}

	if len(bulkResponse.Data) == 0 {
		return nil, fmt.Errorf("empty bulk quotes data for symbols: %s", symbolsStr)
	}

	return bulkResponse.Data, nil
}
