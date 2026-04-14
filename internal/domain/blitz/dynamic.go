// Package blitz provides dynamic-length moving average primitives for BLITZ indicators.
// These primitives adapt their lookback period based on how many bars are available,
// which is essential for indicator warm-up and proper initialization.
package blitz

import "math"

// DynamicLength returns the adaptive lookback period: min(maxLength, barIndex+1).
// On bar 0 the effective length is 1, growing until it reaches maxLength.
func DynamicLength(maxLength, barIndex int) int {
	if maxLength <= 0 {
		return 1
	}
	length := barIndex + 1
	if length > maxLength {
		return maxLength
	}
	return length
}

// DynamicSMA computes a Simple Moving Average with adaptive window length.
// For each bar i, it uses DynamicLength(maxLength, i) as the window size.
// Skip NaN/invalid values in the average calculation.
// Returns a slice of the same length as data, or empty slice if data is empty.
func DynamicSMA(data []float64, maxLength int) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	result := make([]float64, len(data))

	for i := range data {
		windowLen := DynamicLength(maxLength, i)
		startIdx := i - windowLen + 1
		if startIdx < 0 {
			startIdx = 0
		}

		var sum float64
		var count int
		for j := startIdx; j <= i; j++ {
			// Skip NaN values (match Pine's na handling)
			if data[j] == data[j] { // NaN check: NaN != NaN
				sum += data[j]
				count++
			}
		}

		if count > 0 {
			result[i] = sum / float64(count)
		} else {
			result[i] = 0
		}
	}

	return result
}

// DynamicRMA computes Wilder's RMA (Relative Moving Average) with adaptive length.
// For each bar i, the RMA is computed as: rma[i] = (rma[i-1] * (len-1) + data[i]) / len
// When len == 1, it falls back to DynamicSMA(data, 1) for that bar.
// The first bar (i=0) is seeded with data[0].
// Returns a slice of the same length as data, or empty slice if data is empty.
func DynamicRMA(data []float64, maxLength int) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	result := make([]float64, len(data))

	// First bar: seed with first value
	if data[0] == data[0] { // NaN check
		result[0] = data[0]
	} else {
		result[0] = 0
	}

	for i := 1; i < len(data); i++ {
		windowLen := DynamicLength(maxLength, i)

		// Fallback to SMA when length is 1
		if windowLen == 1 {
			sma := DynamicSMA(data, 1)
			result[i] = sma[i]
			continue
		}

		// Wilder's RMA formula: (prev * (len-1) + src) / len
		if data[i] == data[i] { // NaN check for current data point
			result[i] = (result[i-1]*float64(windowLen-1) + data[i]) / float64(windowLen)
		} else {
			result[i] = result[i-1] // Carry forward previous RMA if current is NaN
		}
	}

	return result
}

// DynamicEMA computes an Exponential Moving Average with adaptive length and alpha.
// For each bar i:
//
//	len = DynamicLength(maxLength, i)
//	alpha = 2.0 / float64(len + 1)
//	First valid bar (where i == len-1 for that length): seed ema = data[i]
//	Otherwise: ema[i] = (data[i] - ema[i-1]) * alpha + ema[i-1]
//
// Bars before the seed point carry forward the most recent ema (or 0 if no seed yet).
// Returns a slice of the same length as data, or empty slice if data is empty.
func DynamicEMA(data []float64, maxLength int) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	result := make([]float64, len(data))
	var seeded bool
	var lastEma float64

	for i := range data {
		windowLen := DynamicLength(maxLength, i)
		alpha := 2.0 / float64(windowLen+1)

		// First valid bar for this window length
		switch {
		case i == windowLen-1 && data[i] == data[i]:
			result[i] = data[i]
			seeded = true
			lastEma = result[i]
		case !seeded:
			// Before seed point: use 0 (no data yet)
			result[i] = 0
		default:
			// Standard EMA formula: ema = (src - prev) * alpha + prev
			if data[i] == data[i] { // NaN check
				result[i] = (data[i]-result[i-1])*alpha + result[i-1]
				lastEma = result[i]
			} else {
				result[i] = lastEma // Carry forward last known EMA if NaN
			}
		}
	}

	return result
}

// DynamicWMA computes a Weighted Moving Average with linearly decreasing weights.
// The most recent value has the highest weight (len), the oldest has weight 1.
//
// For each bar i:
//
//	len = DynamicLength(maxLength, i)
//	Iterate j from 0 to len-1, looking back: value = data[i-j]
//	Skip NaN values.
//	weight = len - j (corrected from original Pine code which had weight = len - 1)
//	weightedSum += value × weight
//	normalizationFactor += weight
//	result[i] = weightedSum / normalizationFactor (if > 0, else 0)
//
// Note: The weight calculation uses "len - j" instead of "len - 1" from the original
// Pine Script code, which was a typo that would give all elements the same weight.
//
// Returns a slice of the same length as data, or empty slice if data is empty.
func DynamicWMA(data []float64, maxLength int) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	result := make([]float64, len(data))

	for i := range data {
		windowLen := DynamicLength(maxLength, i)

		var weightedSum float64
		var normalizationFactor float64

		for j := 0; j < windowLen; j++ {
			idx := i - j
			if idx < 0 {
				break
			}

			// Skip NaN values
			if data[idx] != data[idx] {
				continue
			}

			weight := float64(windowLen - j)
			weightedSum += data[idx] * weight
			normalizationFactor += weight
		}

		if normalizationFactor > 0 {
			result[i] = weightedSum / normalizationFactor
		} else {
			result[i] = 0
		}
	}

	return result
}

// DynamicDEMA computes a Double Exponential Moving Average.
// Formula: dema[i] = 2 × ema1[i] - ema2[i]
// Where ema1 = DynamicEMA(data, maxLength) and ema2 = DynamicEMA(ema1, maxLength).
//
// DEMA is designed to have less lag than a standard EMA by using double smoothing.
//
// Returns a slice of the same length as data, or empty slice if data is empty.
func DynamicDEMA(data []float64, maxLength int) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	ema1 := DynamicEMA(data, maxLength)
	ema2 := DynamicEMA(ema1, maxLength)

	result := make([]float64, len(data))
	for i := range data {
		result[i] = 2*ema1[i] - ema2[i]
	}

	return result
}

// DynamicTEMA computes a Triple Exponential Moving Average.
// Formula: tema[i] = 3 × (ema1[i] - ema2[i]) + ema3[i]
// Where:
//
//	ema1 = DynamicEMA(data, maxLength)
//	ema2 = DynamicEMA(ema1, maxLength)
//	ema3 = DynamicEMA(ema2, maxLength)
//
// TEMA is designed to have even less lag than DEMA by using triple smoothing.
//
// Returns a slice of the same length as data, or empty slice if data is empty.
func DynamicTEMA(data []float64, maxLength int) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	ema1 := DynamicEMA(data, maxLength)
	ema2 := DynamicEMA(ema1, maxLength)
	ema3 := DynamicEMA(ema2, maxLength)

	result := make([]float64, len(data))
	for i := range data {
		result[i] = 3*(ema1[i]-ema2[i]) + ema3[i]
	}

	return result
}

// DynamicHMA computes a Hull Moving Average, designed to be very responsive with minimal lag.
//
// Steps:
//
//	halfLen = round(maxLength / 2)
//	sqrtLen = round(sqrt(maxLength))
//	wma1 = DynamicWMA(data, halfLen) — WMA with half period
//	wma2 = DynamicWMA(data, maxLength) — WMA with full period
//	diff[i] = 2 × wma1[i] - wma2[i]
//	hma = DynamicWMA(diff, sqrtLen) — WMA of the difference with sqrt period
//
// The Hull Moving Average responds faster to price changes than standard SMA/EMA
// while maintaining smoothness.
//
// Returns a slice of the same length as data, or empty slice if data is empty.
func DynamicHMA(data []float64, maxLength int) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	halfLen := int(math.Round(float64(maxLength) / 2))
	sqrtLen := int(math.Round(math.Sqrt(float64(maxLength))))

	// Ensure minimum lengths
	if halfLen < 1 {
		halfLen = 1
	}
	if sqrtLen < 1 {
		sqrtLen = 1
	}

	wma1 := DynamicWMA(data, halfLen)
	wma2 := DynamicWMA(data, maxLength)

	// Build intermediate series: diff = 2*wma1 - wma2
	diff := make([]float64, len(data))
	for i := range data {
		diff[i] = 2*wma1[i] - wma2[i]
	}

	// Final HMA is WMA of the difference series
	return DynamicWMA(diff, sqrtLen)
}

// DynamicLSMA computes a Least Squares Moving Average (linear regression projected value).
//
// For each bar i:
//
//	len = DynamicLength(maxLength, i)
//	Window: data[max(0, i-len+1)..i]
//	Perform linear regression on the window:
//	  x values: 0, 1, 2, ..., len-1 (bar position within window)
//	  y values: data[startIdx], data[startIdx+1], ..., data[i]
//	  Skip NaN values.
//	  Compute slope and intercept using least squares:
//	    Σx, Σy, Σxy, Σx² are summed over valid (non-NaN) points
//	    count = number of valid points
//	    slope = (count × Σxy - Σx × Σy) / (count × Σx² - (Σx)²)
//	    intercept = (Σy - slope × Σx) / count
//	  lsma[i] = slope × (offset - 1) + intercept
//
// If the denominator is zero (all x values the same), use the mean of y.
// If count is 0 (all NaN), result is 0.
//
// Returns a slice of the same length as data, or empty slice if data is empty.
func DynamicLSMA(data []float64, maxLength, offset int) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	result := make([]float64, len(data))

	for i := range data {
		windowLen := DynamicLength(maxLength, i)
		startIdx := i - windowLen + 1
		if startIdx < 0 {
			startIdx = 0
		}

		var sumX, sumY, sumXY, sumX2 float64
		var count int

		// Build regression: x is position within window, y is data value
		for j := startIdx; j <= i; j++ {
			// Skip NaN values
			if data[j] != data[j] {
				continue
			}

			x := float64(j - startIdx)
			y := data[j]

			sumX += x
			sumY += y
			sumXY += x * y
			sumX2 += x * x
			count++
		}

		if count == 0 {
			result[i] = 0
			continue
		}

		// Compute slope and intercept
		denominator := float64(count)*sumX2 - sumX*sumX

		if denominator != 0 {
			slope := (float64(count)*sumXY - sumX*sumY) / denominator
			intercept := (sumY - slope*sumX) / float64(count)
			// Project to offset position (offset-1 because x starts at 0)
			result[i] = slope*float64(offset-1) + intercept
		} else {
			// All x values are the same (should only happen with count==1)
			// Use the mean of y
			result[i] = sumY / float64(count)
		}
	}

	return result
}
