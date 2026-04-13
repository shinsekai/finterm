// Package blitz provides dynamic-length indicators for BLITZ analysis.
package blitz

// DynamicRSI computes the Relative Strength Index with adaptive length using DynamicRMA.
// For each bar i (starting from 1):
//
//	u = max(closes[i] - closes[i-1], 0)  — upward movement
//	d = max(closes[i-1] - closes[i], 0)  — downward movement
//
// Then computes avgU = DynamicRMA(upMoves, maxLength) and avgD = DynamicRMA(downMoves, maxLength).
// For each bar: rs = avgU[i] / avgD[i], rsi[i] = 100 - 100 / (1 + rs).
//
// Division by zero handling: if avgD[i] == 0, RSI = 100 (all gains, no losses).
// First bar (i=0): RSI = 50 (neutral, no movement data yet).
//
// Returns a slice of the same length as closes, or empty slice if closes is empty.
func DynamicRSI(closes []float64, maxLength int) []float64 {
	if len(closes) == 0 {
		return []float64{}
	}

	result := make([]float64, len(closes))

	// First bar: neutral (no movement data available)
	result[0] = 50

	// Need at least 2 bars to calculate movement
	if len(closes) < 2 {
		return result
	}

	// Calculate upward and downward movements
	// Index 0 is initialized to 0 (no movement on first bar)
	upMoves := make([]float64, len(closes))
	downMoves := make([]float64, len(closes))
	// upMoves[0] and downMoves[0] remain 0 (default Go zero value)

	for i := 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			upMoves[i] = change
			downMoves[i] = 0
		} else {
			upMoves[i] = 0
			downMoves[i] = -change // abs of negative change
		}
	}

	// Compute dynamic RMA for up and down movements
	avgU := DynamicRMA(upMoves, maxLength)
	avgD := DynamicRMA(downMoves, maxLength)

	// Calculate RSI for each bar
	for i := 1; i < len(closes); i++ {
		// Handle division by zero cases
		switch {
		case avgU[i] == 0 && avgD[i] == 0:
			// No movement (constant prices) - undefined, return 0
			result[i] = 0
		case avgD[i] == 0:
			// All gains, no losses
			result[i] = 100
		case avgU[i] == 0:
			// All losses, no gains
			result[i] = 0
		default:
			rs := avgU[i] / avgD[i]
			result[i] = 100 - 100/(1+rs)
		}
	}

	return result
}
