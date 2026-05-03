package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// OverviewFunction is the Alpha Vantage function name for company overview.
	OverviewFunction = "OVERVIEW"
	// EarningsFunction is the Alpha Vantage function name for earnings data.
	EarningsFunction = "EARNINGS"
)

// GetCompanyOverview fetches company fundamentals data for a given symbol.
// The OVERVIEW endpoint provides comprehensive company information including
// valuation metrics, financial ratios, and other fundamental data.
// Returns an error if the symbol is empty or if the API request fails.
func (c *Client) GetCompanyOverview(ctx context.Context, symbol string) (*CompanyOverview, error) {
	if err := validateSymbol(symbol); err != nil {
		return nil, err
	}

	body, err := c.get(ctx, map[string]string{
		"function": OverviewFunction,
		"symbol":   strings.ToUpper(symbol),
	})
	if err != nil {
		return nil, fmt.Errorf("fetching company overview for %s: %w", symbol, err)
	}

	var overview CompanyOverview
	if err := json.Unmarshal(body, &overview); err != nil {
		return nil, fmt.Errorf("unmarshaling company overview for %s: %w", symbol, err)
	}

	return &overview, nil
}

// GetEarnings fetches earnings history for a given symbol.
// The EARNINGS endpoint provides both annual and quarterly earnings data
// including reported and estimated EPS, surprise metrics, and reporting dates.
// Returns an error if the symbol is empty or if the API request fails.
func (c *Client) GetEarnings(ctx context.Context, symbol string) (*Earnings, error) {
	if err := validateSymbol(symbol); err != nil {
		return nil, err
	}

	body, err := c.get(ctx, map[string]string{
		"function": EarningsFunction,
		"symbol":   strings.ToUpper(symbol),
	})
	if err != nil {
		return nil, fmt.Errorf("fetching earnings for %s: %w", symbol, err)
	}

	var earnings Earnings
	if err := json.Unmarshal(body, &earnings); err != nil {
		return nil, fmt.Errorf("unmarshaling earnings for %s: %w", symbol, err)
	}

	return &earnings, nil
}

// validateSymbol checks if a ticker symbol is valid.
// Returns an error for empty strings or symbols that don't match the expected format.
func validateSymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	// Allow alphanumeric, dots, and dashes (common in ticker symbols)
	// Max length of 10 characters
	symbol = strings.TrimSpace(symbol)
	if len(symbol) == 0 || len(symbol) > 10 {
		return fmt.Errorf("symbol must be 1-10 characters")
	}
	for _, r := range symbol {
		isAlpha := (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
		isDigit := r >= '0' && r <= '9'
		isSpecial := r == '.' || r == '-'
		if !isAlpha && !isDigit && !isSpecial {
			return fmt.Errorf("symbol contains invalid characters: %s", symbol)
		}
	}
	return nil
}
