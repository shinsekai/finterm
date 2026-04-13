// Package blitz provides dynamic-length indicators for BLITZ analysis.
package blitz

import "math"

// PearsonCorrelation computes Pearson's correlation coefficient r between two series.
// For each bar i where i >= period - 1:
//
//	Take the last period values from x and y.
//	Compute Pearson's r:
//	  meanX = mean(xWindow)
//	  meanY = mean(yWindow)
//	  sumXY = Σ(xi - meanX)(yi - meanY)
//	  sumX2 = Σ(xi - meanX)²
//	  sumY2 = Σ(yi - meanY)²
//	  r = sumXY / sqrt(sumX2 * sumY2)
//
// If denominator is 0 (constant data), r = 0.
// For bars before period-1: r = 0 (insufficient data).
//
// Returns a slice of the same length as x, or empty slice if x is empty.
func PearsonCorrelation(x, y []float64, period int) []float64 {
	if len(x) == 0 {
		return []float64{}
	}

	result := make([]float64, len(x))

	// Different length slices: use minimum length
	minLen := len(x)
	if len(y) < minLen {
		minLen = len(y)
	}

	// For bars before period-1, correlation is undefined (use 0)
	for i := 0; i < period-1 && i < minLen; i++ {
		result[i] = 0
	}

	// Calculate correlation for each bar from period-1 onwards
	for i := period - 1; i < minLen; i++ {
		// Extract the window
		start := i - period + 1
		xWindow := x[start : i+1]
		yWindow := y[start : i+1]

		// Calculate means
		var meanX, meanY float64
		for j := 0; j < period; j++ {
			meanX += xWindow[j]
			meanY += yWindow[j]
		}
		meanX /= float64(period)
		meanY /= float64(period)

		// Calculate sums
		var sumXY, sumX2, sumY2 float64
		for j := 0; j < period; j++ {
			dx := xWindow[j] - meanX
			dy := yWindow[j] - meanY
			sumXY += dx * dy
			sumX2 += dx * dx
			sumY2 += dy * dy
		}

		// Calculate r
		denom := math.Sqrt(sumX2 * sumY2)
		if denom == 0 {
			// Constant data, no correlation
			result[i] = 0
		} else {
			result[i] = sumXY / denom
		}
	}

	return result
}

// TSI computes the Trend Strength Indicator as the correlation between close prices and bar index.
// Builds the bar index series: barIndex[i] = float64(i) for all i.
// Then calls PearsonCorrelation(closes, barIndex, period).
//
// This is the convenience wrapper that matches ta.correlation(close, bar_index, period).
// Returns a slice of the same length as closes, or empty slice if closes is empty.
func TSI(closes []float64, period int) []float64 {
	if len(closes) == 0 {
		return []float64{}
	}

	// Build bar index series
	barIndex := make([]float64, len(closes))
	for i := range barIndex {
		barIndex[i] = float64(i)
	}

	return PearsonCorrelation(closes, barIndex, period)
}
