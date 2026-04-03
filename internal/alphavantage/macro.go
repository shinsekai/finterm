// Package alphavantage provides a client for the Alpha Vantage API.
// This file implements macroeconomic data endpoints.
package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	// function names for Alpha Vantage macroeconomic endpoints
	functionRealGDP          = "REAL_GDP"
	functionRealGDPPerCapita = "REAL_GDP_PER_CAPITA"
	functionCPI              = "CPI"
	functionInflation        = "INFLATION"
	functionFedFundsRate     = "FEDERAL_FUNDS_RATE"
	functionTreasuryYield    = "TREASURY_YIELD"
	functionUnemployment     = "UNEMPLOYMENT"
	functionNonfarmPayroll   = "NONFARM_PAYROLL"
)

// GetRealGDP fetches real GDP data for the US economy.
// The interval parameter can be "annual" or "quarterly".
// Context is propagated to the underlying HTTP call.
func (c *Client) GetRealGDP(ctx context.Context, interval string) ([]MacroDataPoint, error) {
	if interval == "" {
		interval = "quarterly"
	}

	// Validate interval
	validIntervals := map[string]bool{
		"annual":    true,
		"quarterly": true,
	}
	if !validIntervals[interval] {
		return nil, fmt.Errorf("invalid interval %s, must be one of: annual, quarterly", interval)
	}

	params := map[string]string{
		"function": functionRealGDP,
		"interval": interval,
		"datatype": "json",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching real GDP data: %w", err)
	}

	return parseMacroResponse(body)
}

// GetRealGDPPerCapita fetches real GDP per capita data for the US economy.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetRealGDPPerCapita(ctx context.Context) ([]MacroDataPoint, error) {
	params := map[string]string{
		"function": functionRealGDPPerCapita,
		"datatype": "json",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching real GDP per capita data: %w", err)
	}

	return parseMacroResponse(body)
}

// GetCPI fetches Consumer Price Index data for the US economy.
// The interval parameter can be "monthly" or "semiannual".
// Context is propagated to the underlying HTTP call.
func (c *Client) GetCPI(ctx context.Context, interval string) ([]MacroDataPoint, error) {
	if interval == "" {
		interval = "monthly"
	}

	// Validate interval
	validIntervals := map[string]bool{
		"monthly":    true,
		"semiannual": true,
	}
	if !validIntervals[interval] {
		return nil, fmt.Errorf("invalid interval %s, must be one of: monthly, semiannual", interval)
	}

	params := map[string]string{
		"function": functionCPI,
		"interval": interval,
		"datatype": "json",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching CPI data: %w", err)
	}

	return parseMacroResponse(body)
}

// GetInflation fetches inflation data (CPI rate of change) for the US economy.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetInflation(ctx context.Context) ([]MacroDataPoint, error) {
	params := map[string]string{
		"function": functionInflation,
		"datatype": "json",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching inflation data: %w", err)
	}

	return parseMacroResponse(body)
}

// GetFedFundsRate fetches the Federal Funds Rate data.
// The interval parameter can be "monthly", "quarterly", or "annual".
// Context is propagated to the underlying HTTP call.
func (c *Client) GetFedFundsRate(ctx context.Context, interval string) ([]MacroDataPoint, error) {
	if interval == "" {
		interval = "monthly"
	}

	// Validate interval
	validIntervals := map[string]bool{
		"daily":     true,
		"weekly":    true,
		"monthly":   true,
		"quarterly": true,
		"annual":    true,
	}
	if !validIntervals[interval] {
		return nil, fmt.Errorf("invalid interval %s, must be one of: daily, weekly, monthly, quarterly, annual", interval)
	}

	params := map[string]string{
		"function": functionFedFundsRate,
		"interval": interval,
		"datatype": "json",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching Fed Funds Rate data: %w", err)
	}

	return parseMacroResponse(body)
}

// GetTreasuryYield fetches US Treasury yield data.
// The interval parameter can be "daily", "weekly", or "monthly".
// The maturity parameter can be "2year", "5year", "10year", or "30year".
// Context is propagated to the underlying HTTP call.
func (c *Client) GetTreasuryYield(ctx context.Context, interval, maturity string) ([]MacroDataPoint, error) {
	if interval == "" {
		interval = "daily"
	}
	if maturity == "" {
		maturity = "10year"
	}

	// Validate interval
	validIntervals := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	if !validIntervals[interval] {
		return nil, fmt.Errorf("invalid interval %s, must be one of: daily, weekly, monthly", interval)
	}

	// Validate maturity
	validMaturities := map[string]bool{
		"2year":  true,
		"5year":  true,
		"10year": true,
		"30year": true,
	}
	if !validMaturities[maturity] {
		return nil, fmt.Errorf("invalid maturity %s, must be one of: 2year, 5year, 10year, 30year", maturity)
	}

	params := map[string]string{
		"function": functionTreasuryYield,
		"interval": interval,
		"maturity": maturity,
		"datatype": "json",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching Treasury Yield data: %w", err)
	}

	return parseMacroResponse(body)
}

// GetUnemployment fetches unemployment rate data for the US economy.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetUnemployment(ctx context.Context) ([]MacroDataPoint, error) {
	params := map[string]string{
		"function": functionUnemployment,
		"datatype": "json",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching unemployment data: %w", err)
	}

	return parseMacroResponse(body)
}

// GetNonfarmPayroll fetches nonfarm payroll (employment) data for the US economy.
// Context is propagated to the underlying HTTP call.
func (c *Client) GetNonfarmPayroll(ctx context.Context) ([]MacroDataPoint, error) {
	params := map[string]string{
		"function": functionNonfarmPayroll,
		"datatype": "json",
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetching nonfarm payroll data: %w", err)
	}

	return parseMacroResponse(body)
}

// parseMacroResponse parses a macroeconomic API response and returns sorted MacroDataPoint.
func parseMacroResponse(body []byte) ([]MacroDataPoint, error) {
	var response MacroResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing macro response: %w", err)
	}

	// Handle empty data
	if len(response.Data) == 0 {
		return []MacroDataPoint{}, nil
	}

	// Sort by date descending (most recent first)
	sort.Slice(response.Data, func(i, j int) bool {
		return strings.Compare(response.Data[i].Date, response.Data[j].Date) > 0
	})

	return response.Data, nil
}
