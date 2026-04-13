// Package blitz provides dynamic-length moving average primitives for BLITZ indicators.
// These primitives adapt their lookback period based on how many bars are available,
// which is essential for indicator warm-up and proper initialization.
package blitz

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
