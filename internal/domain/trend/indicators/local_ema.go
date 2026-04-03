// Package indicators provides technical indicator implementations.
package indicators

import (
	"context"
	"errors"
	"fmt"
)

// LocalEMA computes the Exponential Moving Average locally from OHLCV data.
// It matches TradingView's ta.ema() function exactly.
//
// Critical differences from other EMA implementations:
// - Multiplier: alpha = 2 / (period + 1) (NOT 1/period like RMA)
// - Seed: EMA[0] = first close price (NOT SMA of first period values)
type LocalEMA struct {
	// Data is the OHLCV data to compute EMA from.
	Data []OHLCV
}

// NewLocalEMA creates a new LocalEMA instance with the provided OHLCV data.
func NewLocalEMA(data []OHLCV) *LocalEMA {
	return &LocalEMA{Data: data}
}

// SetData updates the OHLCV data for the indicator.
func (e *LocalEMA) SetData(data []OHLCV) {
	e.Data = data
}

// Compute returns indicator data points for the given symbol.
// For LocalEMA, this method is not applicable as it operates on pre-loaded OHLCV data.
// Use ComputeFromOHLCV instead.
func (e *LocalEMA) Compute(_ context.Context, _ string, _ Options) ([]DataPoint, error) {
	return nil, errors.New("LocalEMA requires OHLCV data. Use ComputeFromOHLCV method instead")
}

// ComputeFromOHLCV calculates EMA values from the provided OHLCV data.
// The period parameter determines the EMA lookback (default: 10 for fast, 20 for slow).
// UseInProgress determines whether to include the last bar (in-progress bar) in calculations.
// When false (recommended), the last bar is excluded to prevent repainting.
//
// Returns []DataPoint with EMA values for each date.
// The output length will be: len(data) - (1 if !useInProgress else 0).
//
// Algorithm (matches TradingView ta.ema() exactly):
// 1. alpha = 2 / (period + 1)
// 2. Seed: EMA[0] = close[0] (first close price, NOT SMA)
// 3. Subsequent: EMA[i] = alpha * close[i] + (1 - alpha) * EMA[i-1]
//
// Edge cases:
// - period = 1 → EMA equals close price at every bar (alpha = 1, EMA[i] = close[i])
// - insufficient data (< 1 point) → error
func (e *LocalEMA) ComputeFromOHLCV(period int, useInProgress bool) ([]DataPoint, error) {
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}

	// Determine the data range to use
	data := e.Data
	if !useInProgress && len(data) > 0 {
		// Exclude the last (potentially in-progress) bar
		data = data[:len(data)-1]
	}

	// We need at least 1 data point (EMA can start immediately)
	if len(data) < 1 {
		return nil, fmt.Errorf("insufficient data: need at least 1 point, got %d", len(data))
	}

	// Calculate alpha (smoothing multiplier)
	// For TradingView ta.ema(): alpha = 2 / (period + 1)
	alpha := 2.0 / float64(period+1)

	// Extract close prices
	closes := make([]float64, len(data))
	for i, d := range data {
		closes[i] = d.Close
	}

	// TradingView seeds EMA with the first source value, NOT SMA
	// This is a critical difference from some other implementations
	emaValues := make([]DataPoint, len(data))

	// Seed: EMA[0] = first close price
	prevEMA := closes[0]
	emaValues[0] = DataPoint{
		Date:  data[0].Date,
		Value: prevEMA,
	}

	// Calculate subsequent EMA values
	// EMA[i] = alpha * close[i] + (1 - alpha) * EMA[i-1]
	for i := 1; i < len(closes); i++ {
		ema := alpha*closes[i] + (1-alpha)*prevEMA
		prevEMA = ema
		emaValues[i] = DataPoint{
			Date:  data[i].Date,
			Value: ema,
		}
	}

	return emaValues, nil
}
