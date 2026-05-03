// Package alphavantage provides a client for the Alpha Vantage API.
// This file implements commodity data endpoints.
package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// CommodityFunction represents an Alpha Vantage commodity data function.
type CommodityFunction string

const (
	// CommodityFunctionWTI is West Texas Intermediate crude oil.
	CommodityFunctionWTI CommodityFunction = "WTI"
	// CommodityFunctionBrent is Brent crude oil.
	CommodityFunctionBrent CommodityFunction = "BRENT"
	// CommodityFunctionNaturalGas is natural gas.
	CommodityFunctionNaturalGas CommodityFunction = "NATURAL_GAS"
	// CommodityFunctionCopper is copper.
	CommodityFunctionCopper CommodityFunction = "COPPER"
	// CommodityFunctionAluminum is aluminum.
	CommodityFunctionAluminum CommodityFunction = "ALUMINUM"
	// CommodityFunctionWheat is wheat.
	CommodityFunctionWheat CommodityFunction = "WHEAT"
	// CommodityFunctionCorn is corn.
	CommodityFunctionCorn CommodityFunction = "CORN"
	// CommodityFunctionCotton is cotton.
	CommodityFunctionCotton CommodityFunction = "COTTON"
	// CommodityFunctionSugar is sugar.
	CommodityFunctionSugar CommodityFunction = "SUGAR"
	// CommodityFunctionCoffee is coffee.
	CommodityFunctionCoffee CommodityFunction = "COFFEE"
	// CommodityFunctionAllCommodities is a composite of all commodities.
	CommodityFunctionAllCommodities CommodityFunction = "ALL_COMMODITIES"
)

// validIntervals maps each commodity function to its supported intervals.
// Based on Alpha Vantage API documentation for commodity data.
var validIntervals = map[CommodityFunction][]string{
	CommodityFunctionWTI:            {"daily", "weekly", "monthly", "quarterly"},
	CommodityFunctionBrent:          {"daily", "weekly", "monthly", "quarterly"},
	CommodityFunctionNaturalGas:     {"daily", "weekly", "monthly", "quarterly"},
	CommodityFunctionCopper:         {"monthly", "quarterly", "annual"},
	CommodityFunctionAluminum:       {"monthly", "quarterly", "annual"},
	CommodityFunctionWheat:          {"daily", "weekly", "monthly", "quarterly", "annual"},
	CommodityFunctionCorn:           {"daily", "weekly", "monthly", "quarterly", "annual"},
	CommodityFunctionCotton:         {"daily", "weekly", "monthly", "quarterly", "annual"},
	CommodityFunctionSugar:          {"daily", "weekly", "monthly", "quarterly", "annual"},
	CommodityFunctionCoffee:         {"daily", "weekly", "monthly", "quarterly", "annual"},
	CommodityFunctionAllCommodities: {"monthly", "quarterly", "annual"},
}

// ErrUnsupportedInterval is returned when an interval is not supported for a given commodity function.
var ErrUnsupportedInterval = fmt.Errorf("unsupported interval for this commodity")

// symbolToFunction maps commodity symbols to their corresponding CommodityFunction values.
// This provides a case-insensitive lookup for commodity symbols.
var symbolToFunction = map[string]CommodityFunction{
	"WTI":             CommodityFunctionWTI,
	"BRENT":           CommodityFunctionBrent,
	"NATURAL_GAS":     CommodityFunctionNaturalGas,
	"COPPER":          CommodityFunctionCopper,
	"ALUMINUM":        CommodityFunctionAluminum,
	"WHEAT":           CommodityFunctionWheat,
	"CORN":            CommodityFunctionCorn,
	"COTTON":          CommodityFunctionCotton,
	"SUGAR":           CommodityFunctionSugar,
	"COFFEE":          CommodityFunctionCoffee,
	"ALL_COMMODITIES": CommodityFunctionAllCommodities,
}

// CommodityFunctionFromSymbol maps a commodity symbol to its CommodityFunction.
// Returns the function and true if found, false otherwise.
// Lookup is case-insensitive (e.g., "wti", "WTI", "Wti" all return CommodityFunctionWTI).
func CommodityFunctionFromSymbol(symbol string) (CommodityFunction, bool) {
	upper := strings.ToUpper(symbol)
	fn, ok := symbolToFunction[upper]
	return fn, ok
}

// CommoditySupportedIntervals returns the supported intervals for a given CommodityFunction.
// Returns the slice of intervals and true if the function is known, false otherwise.
func CommoditySupportedIntervals(fn CommodityFunction) ([]string, bool) {
	intervals, ok := validIntervals[fn]
	return intervals, ok
}

// BestSupportedInterval returns the best supported interval for a given commodity function.
// If the requested interval is supported, it is returned unchanged.
// Otherwise, returns the next-coarser supported interval following the precedence:
// daily → weekly → monthly → quarterly → annual.
// If the requested interval is unknown but valid intervals exist, returns the coarsest supported.
// If no overlap exists (defensive), returns an empty string.
func BestSupportedInterval(fn CommodityFunction, requested string) string {
	valid, ok := validIntervals[fn]
	if !ok {
		return ""
	}

	// Check exact match first
	for _, iv := range valid {
		if iv == requested {
			return requested
		}
	}

	// Fallback precedence: daily → weekly → monthly → quarterly → annual
	precedence := []string{"daily", "weekly", "monthly", "quarterly", "annual"}

	// Find the requested interval in precedence
	requestedIndex := -1
	for i, iv := range precedence {
		if iv == requested {
			requestedIndex = i
			break
		}
	}

	// If requested interval is unknown (not in precedence), return the coarsest supported interval
	if requestedIndex == -1 {
		for i := len(precedence) - 1; i >= 0; i-- {
			if contains(valid, precedence[i]) {
				return precedence[i]
			}
		}
		return ""
	}

	// Find next coarser supported interval (after requested in precedence)
	for i := requestedIndex + 1; i < len(precedence); i++ {
		if contains(valid, precedence[i]) {
			return precedence[i]
		}
	}

	// No coarser interval found - requested is finer than any supported
	// Return the coarsest supported interval as fallback
	for i := len(precedence) - 1; i >= 0; i-- {
		if contains(valid, precedence[i]) {
			return precedence[i]
		}
	}

	// Should not reach here
	return ""
}

// contains checks if a string slice contains a specific string.
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// GetCommodity fetches commodity price data for the specified function and interval.
// The interval parameter can be "daily", "weekly", "monthly", "quarterly", or "annual",
// depending on the commodity function. Valid intervals per function are documented
// in the CommodityFunction constants.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetCommodity(ctx context.Context, fn CommodityFunction, interval string) (*CommoditySeries, error) {
	// Validate function and interval combination
	valid, ok := validIntervals[fn]
	if !ok {
		return nil, fmt.Errorf("unknown commodity function: %s", fn)
	}

	intervalSupported := false
	for _, iv := range valid {
		if iv == interval {
			intervalSupported = true
			break
		}
	}

	if !intervalSupported {
		return nil, fmt.Errorf("%w: interval %q not supported for %s (valid: %v)", ErrUnsupportedInterval, interval, fn, valid)
	}

	params := map[string]string{
		"function": string(fn),
		"interval": interval,
		"datatype": "json",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching commodity data: %w", err)
	}

	return parseCommodityResponse(body)
}

// parseCommodityResponse parses a commodity API response and returns CommoditySeries.
// Data is returned in descending order (newest first) as provided by Alpha Vantage.
func parseCommodityResponse(body []byte) (*CommoditySeries, error) {
	var response CommoditySeries
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing commodity response: %w", err)
	}

	// Handle empty data
	if len(response.Data) == 0 {
		response.Data = []CommodityDataPoint{}
	}

	return &response, nil
}
