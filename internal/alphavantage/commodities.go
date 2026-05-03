// Package alphavantage provides a client for the Alpha Vantage API.
// This file implements commodity data endpoints.
package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
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
