// Package indicators provides technical indicator implementations.
package indicators

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// LocalRSI computes the Relative Strength Index locally from OHLCV data.
// It uses RMA (Wilder's Smoothing) for the moving average calculation,
// which matches TradingView's ta.rsi() function exactly.
//
// RMA alpha = 1 / period (NOT EMA's alpha = 2 / (period + 1)).
type LocalRSI struct {
	// Data is the OHLCV data to compute RSI from.
	Data []OHLCV
}

// NewLocalRSI creates a new LocalRSI instance with the provided OHLCV data.
func NewLocalRSI(data []OHLCV) *LocalRSI {
	return &LocalRSI{Data: data}
}

// SetData updates the OHLCV data for the indicator.
func (r *LocalRSI) SetData(data []OHLCV) {
	r.Data = data
}

// Compute returns indicator data points for the given symbol.
// For LocalRSI, this method is not applicable as it operates on pre-loaded OHLCV data.
// Use ComputeFromOHLCV instead.
func (r *LocalRSI) Compute(_ context.Context, _ string, _ Options) ([]DataPoint, error) {
	return nil, errors.New("LocalRSI requires OHLCV data. Use ComputeFromOHLCV method instead")
}

// ComputeFromOHLCV calculates RSI values from the provided OHLCV data.
// The period parameter determines the RSI lookback (default: 14).
// UseInProgress determines whether to include the last bar (in-progress bar) in calculations.
// When false (recommended), the last bar is excluded to prevent repainting.
//
// Returns []DataPoint with RSI values for each date after the warmup period.
// The output length will be: len(data) - period - (1 if !useInProgress else 0).
//
// Algorithm (matches TradingView ta.rsi() exactly):
// 1. Calculate price changes: change[i] = close[i] - close[i-1]
// 2. Separate gains (max(change, 0)) and losses (abs(minFloat64(change, 0)))
// 3. Smooth with RMA (alpha = 1/period), NOT EMA:
//   - Seed: SMA of first 'period' gains/losses
//   - Subsequent: avg = (prev_avg * (period - 1) + current) / period
//
// 4. Compute RS = avg_gain / avg_loss
// 5. Compute RSI = 100 - (100 / (1 + RS))
//
// Edge cases:
// - avg_loss = 0 → RSI = 100
// - avg_gain = 0 → RSI = 0
func (r *LocalRSI) ComputeFromOHLCV(period int, useInProgress bool) ([]DataPoint, error) {
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}

	// Determine the data range to use
	data := r.Data
	if !useInProgress && len(data) > 0 {
		// Exclude the last (potentially in-progress) bar
		data = data[:len(data)-1]
	}

	// We need at least period + 1 data points (first for initial close, then period for warmup)
	if len(data) < period+1 {
		return nil, fmt.Errorf("insufficient data: need at least %d points, got %d", period+1, len(data))
	}

	// Extract close prices
	closes := make([]float64, len(data))
	for i, d := range data {
		closes[i] = d.Close
	}

	// Calculate price changes, gains, and losses
	// changes[i] corresponds to change from closes[i-1] to closes[i]
	changes := make([]float64, len(closes)-1)
	gains := make([]float64, len(changes))
	losses := make([]float64, len(changes))

	for i := 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		changes[i-1] = change
		gains[i-1] = max(change, 0)
		losses[i-1] = math.Abs(minFloat64(change, 0))
	}

	// Seed with SMA of first 'period' gains/losses
	var avgGain, avgLoss float64
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// RSI for the first period bar (at index = period in closes, or period-1 in changes)
	rsiValues := make([]DataPoint, 0, len(changes)-period+1)

	// Calculate initial RSI using the seeded averages
	rsi := calculateRSI(avgGain, avgLoss)
	rsiValues = append(rsiValues, DataPoint{
		Date:  data[period].Date,
		Value: rsi,
	})

	// Calculate subsequent RSI values using RMA (Wilder's smoothing)
	// RMA formula: avg = (prev_avg * (period - 1) + current) / period
	for i := period; i < len(gains); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

		rsi := calculateRSI(avgGain, avgLoss)
		rsiValues = append(rsiValues, DataPoint{
			Date:  data[i+1].Date,
			Value: rsi,
		})
	}

	// Reverse to return newest-first (consistent with remote indicators)
	for i, j := 0, len(rsiValues)-1; i < j; i, j = i+1, j-1 {
		rsiValues[i], rsiValues[j] = rsiValues[j], rsiValues[i]
	}

	return rsiValues, nil
}

// calculateRSI computes the RSI value from average gain and average loss.
func calculateRSI(avgGain, avgLoss float64) float64 {
	// Edge cases
	if avgLoss == 0 {
		return 100
	}
	if avgGain == 0 {
		return 0
	}

	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

// minFloat64 returns the minimum of two float64 values.
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
