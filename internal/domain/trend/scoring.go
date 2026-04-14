// Package trend provides trend-following analysis and scoring.
package trend

// Signal represents the direction of the trend based on EMA crossover.
type Signal int

const (
	// Bullish indicates an uptrend - EMA fast is above EMA slow.
	Bullish Signal = iota
	// Bearish indicates a downtrend - EMA fast is below EMA slow.
	Bearish
)

// String returns the string representation of the Signal.
func (s Signal) String() string {
	switch s {
	case Bullish:
		return "Bullish"
	case Bearish:
		return "Bearish"
	default:
		return "Unknown"
	}
}

// Score computes the trend signal based on EMA crossover only.
// This is a pure function with no side effects.
//
// Rules:
//   - EMA Fast > EMA Slow → Bullish
//   - EMA Fast < EMA Slow → Bearish
//   - EMA Fast == EMA Slow → Treated as Bearish (no upward momentum)
//
// RSI is NOT used in trend scoring - it's used for valuation only.
func Score(emaFast, emaSlow float64) Signal {
	if emaFast > emaSlow {
		return Bullish
	}
	// EMA fast <= EMA slow → Bearish (includes equality)
	return Bearish
}

// TPI computes the Trend Probability Indicator.
// It averages the EMA crossover signal, BLITZ score, and DESTINY score.
// Returns a value from -1.0 to +1.0.
// TPI > 0 → LONG, TPI <= 0 → CASH.
func TPI(emaSignal Signal, blitzScore, destinyScore int) float64 {
	ema := float64(0)
	if emaSignal == Bullish {
		ema = 1.0
	} else {
		ema = -1.0
	}
	return (ema + float64(blitzScore) + float64(destinyScore)) / 3.0
}

// TPISignal returns the TPI signal label.
// TPI > 0 → "LONG", TPI <= 0 → "CASH".
func TPISignal(tpi float64) string {
	if tpi > 0 {
		return "LONG"
	}
	return "CASH"
}
