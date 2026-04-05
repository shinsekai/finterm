// Package indicators provides technical indicator implementations.
package indicators

import (
	"context"
	"fmt"
	"sort"

	"github.com/shinsekai/finterm/internal/alphavantage"
)

// RSIResponse defines the interface for RSI responses from Alpha Vantage.
type RSIResponse interface {
	GetTechnicalAnalysis() map[string]alphavantage.RSIEntry
}

// EMAResponse defines the interface for EMA responses from Alpha Vantage.
type EMAResponse interface {
	GetTechnicalAnalysis() map[string]alphavantage.EMAEntry
}

// AVClient defines the interface for Alpha Vantage API client methods.
// Using an interface allows for testable mocks and follows the dependency inversion principle.
type AVClient interface {
	GetRSI(ctx context.Context, symbol, interval string, period int) (*alphavantage.RSIResponse, error)
	GetEMA(ctx context.Context, symbol, interval string, period int) (*alphavantage.EMAResponse, error)
}

// RemoteRSI wraps the Alpha Vantage server-side RSI endpoint.
// It implements the Indicator interface for equities that use remote API computation.
type RemoteRSI struct {
	client AVClient
}

// NewRemoteRSI creates a new RemoteRSI with the provided Alpha Vantage client.
// The client is passed as an interface to enable testable mocks.
func NewRemoteRSI(client AVClient) *RemoteRSI {
	return &RemoteRSI{
		client: client,
	}
}

// Compute returns RSI data points for the given symbol using Alpha Vantage's server-side computation.
// The opts.Interval parameter specifies the time interval (e.g., "daily", "weekly").
// The opts.Period parameter specifies the RSI lookback period (typically 14).
// Context is propagated to the underlying HTTP call.
//
// Returns []DataPoint sorted by date descending (newest first).
// AV client errors are propagated with additional context.
func (r *RemoteRSI) Compute(ctx context.Context, symbol string, opts Options) ([]DataPoint, error) {
	if opts.Period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", opts.Period)
	}
	if opts.Interval == "" {
		return nil, fmt.Errorf("interval cannot be empty")
	}

	resp, err := r.client.GetRSI(ctx, symbol, opts.Interval, opts.Period)
	if err != nil {
		return nil, fmt.Errorf("fetching RSI for %s: %w", symbol, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("nil RSI response for symbol %s", symbol)
	}

	// Convert AV response format to []DataPoint
	dataPoints, err := convertRSIResponseToDataPoints(resp)
	if err != nil {
		return nil, fmt.Errorf("converting RSI response for %s: %w", symbol, err)
	}

	// Sort by date descending (newest first)
	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Date.After(dataPoints[j].Date)
	})

	return dataPoints, nil
}

// RemoteEMA wraps the Alpha Vantage server-side EMA endpoint.
// It implements the Indicator interface for equities that use remote API computation.
type RemoteEMA struct {
	client AVClient
}

// NewRemoteEMA creates a new RemoteEMA with the provided Alpha Vantage client.
// The client is passed as an interface to enable testable mocks.
func NewRemoteEMA(client AVClient) *RemoteEMA {
	return &RemoteEMA{
		client: client,
	}
}

// Compute returns EMA data points for the given symbol using Alpha Vantage's server-side computation.
// The opts.Interval parameter specifies the time interval (e.g., "daily", "weekly").
// The opts.Period parameter specifies the EMA lookback period (e.g., 10 for fast, 20 for slow).
// Context is propagated to the underlying HTTP call.
//
// Returns []DataPoint sorted by date descending (newest first).
// AV client errors are propagated with additional context.
func (e *RemoteEMA) Compute(ctx context.Context, symbol string, opts Options) ([]DataPoint, error) {
	if opts.Period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", opts.Period)
	}
	if opts.Interval == "" {
		return nil, fmt.Errorf("interval cannot be empty")
	}

	resp, err := e.client.GetEMA(ctx, symbol, opts.Interval, opts.Period)
	if err != nil {
		return nil, fmt.Errorf("fetching EMA for %s: %w", symbol, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("nil EMA response for symbol %s", symbol)
	}

	// Convert AV response format to []DataPoint
	dataPoints, err := convertEMAResponseToDataPoints(resp)
	if err != nil {
		return nil, fmt.Errorf("converting EMA response for %s: %w", symbol, err)
	}

	// Sort by date descending (newest first)
	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Date.After(dataPoints[j].Date)
	})

	return dataPoints, nil
}

// convertRSIResponseToDataPoints converts an Alpha Vantage RSI response to []DataPoint.
// It parses the TechnicalAnalysis map (keyed by date string) and converts each entry.
func convertRSIResponseToDataPoints(resp *alphavantage.RSIResponse) ([]DataPoint, error) {
	if resp == nil || len(resp.TechnicalAnalysis) == 0 {
		return nil, fmt.Errorf("empty RSI technical analysis data")
	}

	dataPoints := make([]DataPoint, 0, len(resp.TechnicalAnalysis))

	for dateStr, entry := range resp.TechnicalAnalysis {
		// Parse date string
		date, err := alphavantage.ParseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("parsing RSI date %q: %w", dateStr, err)
		}

		// Parse RSI value
		value, err := alphavantage.ParseFloat(entry.RSI)
		if err != nil {
			return nil, fmt.Errorf("parsing RSI value %q for date %s: %w", entry.RSI, dateStr, err)
		}

		dataPoints = append(dataPoints, DataPoint{
			Date:  date,
			Value: value,
		})
	}

	return dataPoints, nil
}

// convertEMAResponseToDataPoints converts an Alpha Vantage EMA response to []DataPoint.
// It parses the TechnicalAnalysis map (keyed by date string) and converts each entry.
func convertEMAResponseToDataPoints(resp *alphavantage.EMAResponse) ([]DataPoint, error) {
	if resp == nil || len(resp.TechnicalAnalysis) == 0 {
		return nil, fmt.Errorf("empty EMA technical analysis data")
	}

	dataPoints := make([]DataPoint, 0, len(resp.TechnicalAnalysis))

	for dateStr, entry := range resp.TechnicalAnalysis {
		// Parse date string
		date, err := alphavantage.ParseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("parsing EMA date %q: %w", dateStr, err)
		}

		// Parse EMA value
		value, err := alphavantage.ParseFloat(entry.EMA)
		if err != nil {
			return nil, fmt.Errorf("parsing EMA value %q for date %s: %w", entry.EMA, dateStr, err)
		}

		dataPoints = append(dataPoints, DataPoint{
			Date:  date,
			Value: value,
		})
	}

	return dataPoints, nil
}
