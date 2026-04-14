// Package blitz provides BLITZ signal scoring combining TSI, Dynamic RSI, and smoothed RSI.
package blitz

import (
	"fmt"

	"github.com/shinsekai/finterm/internal/domain/dynamo"
)

// Config holds BLITZ system parameters.
type Config struct {
	RSILength int     // Default: 12. Dynamic RSI lookback period.
	TSIPeriod int     // Default: 14. Pearson correlation lookback period.
	Threshold float64 // Default: 48. RSI smooth threshold for long/short.
}

// DefaultConfig returns the default BLITZ configuration.
func DefaultConfig() Config {
	return Config{RSILength: 12, TSIPeriod: 14, Threshold: 48}
}

// Score represents the BLITZ signal for a single bar.
type Score int

// Score constants for BLITZ signals.
const (
	Hold  Score = 0  // No signal / hold previous
	Long  Score = 1  // Bullish signal
	Short Score = -1 // Bearish signal
)

// String returns the display string for a Score.
func (s Score) String() string {
	switch s {
	case Long:
		return "LONG"
	case Short:
		return "SHORT"
	default:
		return "HOLD"
	}
}

// Result contains the BLITZ analysis output for a symbol.
type Result struct {
	Scores    []Score // Score for each bar
	Current   Score   // Latest score
	TSI       float64 // Latest TSI value
	RSISmooth float64 // Latest smoothed RSI value
}

// Compute performs BLITZ signal scoring on price data.
// It combines TSI (trend strength), Dynamic RSI (relative strength), and smoothed RSI
// to generate +1 (Long), -1 (Short), or 0 (Hold) signals for each bar.
//
// Validation:
//   - len(closes) >= 2 (need at least 2 bars for RSI change detection)
//   - cfg.RSILength >= 1
//   - cfg.TSIPeriod >= 1
//   - cfg.Threshold > 0
//
// Long signal fires when: TSI > 0 AND smoothed RSI is rising AND smoothed RSI > threshold
// Short signal fires when: TSI < 0 AND smoothed RSI is falling AND smoothed RSI < threshold
// Score holds previous value when neither long nor short triggers.
func Compute(closes []float64, cfg Config) (*Result, error) {
	// Validate inputs
	if len(closes) < 2 {
		return nil, fmt.Errorf("insufficient data: need at least 2 bars, got %d", len(closes))
	}
	if cfg.RSILength < 1 {
		return nil, fmt.Errorf("invalid RSILength: must be >= 1, got %d", cfg.RSILength)
	}
	if cfg.TSIPeriod < 1 {
		return nil, fmt.Errorf("invalid TSIPeriod: must be >= 1, got %d", cfg.TSIPeriod)
	}
	if cfg.Threshold <= 0 {
		return nil, fmt.Errorf("invalid Threshold: must be > 0, got %f", cfg.Threshold)
	}

	// Step 1: Compute TSI (Trend Strength Indicator)
	tsi := dynamo.TSI(closes, cfg.TSIPeriod)

	// Step 2: Compute Dynamic RSI
	rsi := dynamo.DynamicRSI(closes, cfg.RSILength)

	// Step 3: Smooth the RSI using DynamicEMA
	rsiSmooth := dynamo.DynamicEMA(rsi, cfg.RSILength)

	// Step 4: Generate scores for each bar
	scores := make([]Score, len(closes))
	prevScore := Hold // Initial score is Hold

	for i := 0; i < len(closes); i++ {
		// First bar: cannot compute change, use Hold
		if i == 0 {
			scores[i] = Hold
			continue
		}

		// Evaluate signal conditions
		longSignal := tsi[i] > 0 &&
			rsiSmooth[i] > rsiSmooth[i-1] &&
			rsiSmooth[i] > cfg.Threshold

		shortSignal := tsi[i] < 0 &&
			rsiSmooth[i] < rsiSmooth[i-1] &&
			rsiSmooth[i] < cfg.Threshold

		// Score assignment (maintains state like Pine's var score)
		// Short overrides long when both conditions are met
		switch {
		case shortSignal:
			scores[i] = Short
		case longSignal:
			scores[i] = Long
		default:
			scores[i] = prevScore // Hold previous score
		}

		prevScore = scores[i]
	}

	// Build and return result
	return &Result{
		Scores:    scores,
		Current:   scores[len(scores)-1],
		TSI:       tsi[len(tsi)-1],
		RSISmooth: rsiSmooth[len(rsiSmooth)-1],
	}, nil
}

// ComputeSingle is a convenience wrapper using DefaultConfig().
func ComputeSingle(closes []float64) (*Result, error) {
	return Compute(closes, DefaultConfig())
}
