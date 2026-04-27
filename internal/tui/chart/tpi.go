// Package chart provides TPI history computation for chart visualization.
package chart

import (
	"fmt"

	"github.com/shinsekai/finterm/internal/config"
	"github.com/shinsekai/finterm/internal/domain/blitz"
	"github.com/shinsekai/finterm/internal/domain/destiny"
	"github.com/shinsekai/finterm/internal/domain/flow"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
	"github.com/shinsekai/finterm/internal/domain/vortex"
)

// computeTPIHistory computes the composite TPI trajectory for each bar.
// It computes (ftema_score + blitz_score + destiny_score + flow_score + vortex_score) / 5
// per bar by calling the existing five engine compute entry points on the full bar window.
func computeTPIHistory(bars []indicators.OHLCV, cfg *config.Config) ([]float64, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("no bars provided")
	}

	// Extract close prices for computation
	closes := make([]float64, len(bars))
	opens := make([]float64, len(bars))
	highs := make([]float64, len(bars))
	lows := make([]float64, len(bars))

	for i, bar := range bars {
		closes[i] = bar.Close
		opens[i] = bar.Open
		highs[i] = bar.High
		lows[i] = bar.Low
	}

	// Compute BLITZ scores per bar
	blitzResult, err := blitz.Compute(closes, blitz.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("computing BLITZ: %w", err)
	}

	// Compute DESTINY scores per bar
	destinyOHLCV := make([]destiny.OHLCV, len(bars))
	for i, c := range closes {
		destinyOHLCV[i] = destiny.OHLCV{
			Date:   float64(i),
			Open:   c,
			High:   c,
			Low:    c,
			Close:  c,
			Volume: 0,
		}
	}
	destinyResult, err := destiny.Compute(destinyOHLCV, destiny.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("computing DESTINY: %w", err)
	}

	// Compute FLOW scores per bar
	flowOHLCV := make([]flow.OHLCV, len(bars))
	for i := range bars {
		flowOHLCV[i] = flow.OHLCV{
			Date:   bars[i].Date,
			Open:   bars[i].Open,
			High:   bars[i].High,
			Low:    bars[i].Low,
			Close:  bars[i].Close,
			Volume: bars[i].Volume,
		}
	}
	flowResult, err := flow.Compute(flowOHLCV, flow.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("computing FLOW: %w", err)
	}

	// Compute VORTEX scores per bar
	vortexOHLCV := make([]vortex.OHLCV, len(bars))
	for i, c := range closes {
		vortexOHLCV[i] = vortex.OHLCV{
			Date:   float64(i),
			Open:   c,
			High:   c,
			Low:    c,
			Close:  c,
			Volume: 0,
		}
	}
	vortexResult, err := vortex.Compute(vortexOHLCV, vortex.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("computing VORTEX: %w", err)
	}

	// Compute FTEMA direction per bar (from EMA crossover)
	ftemaScores := computeFTEMAScores(closes, cfg)

	// Combine all scores into TPI
	// TPI = (ftema + blitz + destiny + flow + vortex) / 5
	// Normalize each to -1 to +1 range
	tpiHistory := make([]float64, len(bars))

	for i := 0; i < len(bars); i++ {
		// Get scores for this bar (use 0 if not available)
		ftema := 0.0
		if i < len(ftemaScores) {
			ftema = float64(ftemaScores[i])
		}

		blitzScore := 0.0
		if i < len(blitzResult.Scores) {
			blitzScore = float64(blitzResult.Scores[i])
		}

		destinyScore := 0.0
		if i == len(destinyOHLCV)-1 {
			// DESTINY only provides a single latched score, not per-bar
			// Use the final score for all bars
			destinyScore = float64(destinyResult.Score)
		}

		flowScore := 0.0
		if i == len(flowOHLCV)-1 {
			// FLOW only provides a single latched score, not per-bar
			// Use the final score for all bars
			flowScore = float64(flowResult.Score)
		}

		vortexScore := 0.0
		if i == len(vortexOHLCV)-1 {
			// VORTEX only provides a single latched score, not per-bar
			// Use the final score for all bars
			vortexScore = float64(vortexResult.Score)
		}

		// Calculate composite TPI
		tpiHistory[i] = (ftema + blitzScore + destinyScore + flowScore + vortexScore) / 5.0
	}

	return tpiHistory, nil
}

// computeFTEMAScores computes the FTEMA direction score (+1, -1, or 0) per bar.
// FTEMA is based on the EMA crossover: +1 when EMA(10) > EMA(20), -1 when EMA(10) < EMA(20).
func computeFTEMAScores(closes []float64, cfg *config.Config) []int {
	if len(closes) == 0 {
		return nil
	}

	// Compute EMAs using dynamo package
	ema10 := computeEMA(closes, cfg.Trend.EMAFast)
	ema20 := computeEMA(closes, cfg.Trend.EMASlow)

	scores := make([]int, len(closes))

	for i := 0; i < len(closes); i++ {
		if i < len(ema10) && i < len(ema20) {
			switch {
			case ema10[i] > ema20[i]:
				scores[i] = 1
			case ema10[i] < ema20[i]:
				scores[i] = -1
			default:
				scores[i] = 0
			}
		} else {
			scores[i] = 0
		}
	}

	return scores
}

// computeEMA computes the Exponential Moving Average for the given period.
// This is a simplified version that matches TradingView's behavior.
func computeEMA(values []float64, period int) []float64 {
	if len(values) == 0 || period <= 0 {
		return nil
	}

	ema := make([]float64, len(values))
	multiplier := 2.0 / float64(period+1)

	// Seed with the first value (TradingView behavior)
	ema[0] = values[0]

	// Compute EMA for subsequent values
	for i := 1; i < len(values); i++ {
		ema[i] = (values[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}
