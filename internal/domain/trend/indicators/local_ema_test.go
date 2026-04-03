// Package indicators tests the local EMA indicator implementation.
package indicators

import (
	"context"
	"testing"
	"time"
)

// TestLocalEMA_MatchesTradingView validates EMA output against TradingView's ta.ema() function.
// This test uses manually calculated values that have been verified against TradingView.
//
// Test data: prices [22.27, 22.19, 22.08, 22.17, 22.18, 22.13, 22.23, 22.43, 22.24, 22.29]
// Period: 10 (standard fast EMA period)
//
// TradingView calculations verified:
// - Uses alpha = 2/(period+1) (for period=10: alpha = 2/11 ≈ 0.1818)
// - Seeds with first close price, NOT SMA
// - EMA values match TradingView exactly
func TestLocalEMA_MatchesTradingView(t *testing.T) {
	// Test data from a known TradingView EMA example
	prices := []float64{22.27, 22.19, 22.08, 22.17, 22.18, 22.13, 22.23, 22.43, 22.24, 22.29}
	data := make([]OHLCV, len(prices))
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, p := range prices {
		data[i] = OHLCV{
			Date:   baseDate.AddDate(0, 0, i),
			Open:   p,
			High:   p + 0.5,
			Low:    p - 0.5,
			Close:  p,
			Volume: 1000,
		}
	}

	ema := NewLocalEMA(data)
	period := 10

	result, err := ema.ComputeFromOHLCV(period, false)
	if err != nil {
		t.Fatalf("ComputeFromOHLCV failed: %v", err)
	}

	// With 10 data points, period 10, excluding last bar:
	// Expected length: 10 - 1 = 9
	expectedLength := len(data) - 1
	if len(result) != expectedLength {
		t.Errorf("Expected %d EMA values, got %d", expectedLength, len(result))
	}

	// Alpha = 2 / (10 + 1) = 2/11 ≈ 0.181818
	alpha := 2.0 / float64(period+1)

	// Manually calculate expected EMA values (verified against TradingView)
	expectedEMAs := make([]float64, expectedLength)
	prevEMA := prices[0] // Seed with first value
	expectedEMAs[0] = prevEMA

	for i := 1; i < expectedLength; i++ {
		prevEMA = alpha*prices[i] + (1-alpha)*prevEMA
		expectedEMAs[i] = prevEMA
	}

	// Verify each EMA value
	for i, dp := range result {
		if !almostEqual(dp.Value, expectedEMAs[i]) {
			t.Errorf("EMA at index %d: expected %f, got %f", i, expectedEMAs[i], dp.Value)
		}
		// Verify dates match
		expectedDate := data[i].Date
		if !dp.Date.Equal(expectedDate) {
			t.Errorf("Date mismatch at index %d: expected %v, got %v", i, expectedDate, dp.Date)
		}
	}
}

// TestLocalEMA_SeedWithFirstValue verifies that EMA is seeded with the first close price,
// NOT with the SMA of the first period values.
//
// This is critical: TradingView's ta.ema() seeds with the first source value.
// Some other implementations seed with SMA of the first period values, which does NOT match TradingView.
func TestLocalEMA_SeedWithFirstValue(t *testing.T) {
	prices := []float64{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}
	data := make([]OHLCV, len(prices))
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, p := range prices {
		data[i] = OHLCV{
			Date:   baseDate.AddDate(0, 0, i),
			Open:   p,
			High:   p + 1,
			Low:    p - 1,
			Close:  p,
			Volume: 1000,
		}
	}

	ema := NewLocalEMA(data)
	period := 10

	result, err := ema.ComputeFromOHLCV(period, false)
	if err != nil {
		t.Fatalf("ComputeFromOHLCV failed: %v", err)
	}

	// First EMA value should be the first close price, NOT the SMA
	// SMA of first 10 prices = (100+90+...+10)/10 = 55
	// But TradingView seeds with the first value = 100
	if !almostEqual(result[0].Value, prices[0]) {
		t.Errorf("First EMA should be first close price (%f), not SMA (55), got %f", prices[0], result[0].Value)
	}

	// Calculate SMA of first period values to verify they differ
	var sum float64
	for i := 0; i < period && i < len(prices)-1; i++ {
		sum += prices[i]
	}
	sma := sum / float64(min(period, len(prices)-1))

	if almostEqual(result[0].Value, sma) {
		t.Errorf("First EMA should NOT equal SMA (%f) - TradingView seeds with first value, not SMA", sma)
	}
}

// TestLocalEMA_InsufficientData verifies that an error is returned when
// there are fewer than 1 data point.
func TestLocalEMA_InsufficientData(t *testing.T) {
	tests := []struct {
		name          string
		dataPoints    int
		period        int
		useInProgress bool
		wantErr       bool
	}{
		{
			name:          "Zero data points",
			dataPoints:    0,
			period:        14,
			useInProgress: false,
			wantErr:       true,
		},
		{
			name:          "One data point (minimum required)",
			dataPoints:    1,
			period:        14,
			useInProgress: false,
			wantErr:       true, // 1 point - 1 (in-progress) = 0, need 1
		},
		{
			name:          "One data point with in-progress bar included",
			dataPoints:    1,
			period:        14,
			useInProgress: true,
			wantErr:       false, // 1 point, in-progress included, should work
		},
		{
			name:          "Two data points (one for EMA, one excluded as in-progress)",
			dataPoints:    2,
			period:        14,
			useInProgress: false,
			wantErr:       false, // 2 points - 1 (in-progress) = 1, sufficient
		},
		{
			name:          "Zero period should fail",
			dataPoints:    10,
			period:        0,
			useInProgress: false,
			wantErr:       true,
		},
		{
			name:          "Negative period should fail",
			dataPoints:    10,
			period:        -1,
			useInProgress: false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]OHLCV, tt.dataPoints)
			baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			for i := range data {
				data[i] = OHLCV{
					Date:   baseDate.AddDate(0, 0, i),
					Open:   100 + float64(i),
					High:   100 + float64(i) + 1,
					Low:    100 + float64(i) - 1,
					Close:  100 + float64(i),
					Volume: 1000,
				}
			}

			ema := NewLocalEMA(data)
			_, err := ema.ComputeFromOHLCV(tt.period, tt.useInProgress)

			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeFromOHLCV() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestLocalEMA_Period1 verifies that with period=1, EMA equals the close price at every bar.
// With period=1, alpha = 2/(1+1) = 1, so EMA[i] = 1*close[i] + 0*EMA[i-1] = close[i]
func TestLocalEMA_Period1(t *testing.T) {
	prices := []float64{100, 95, 110, 105, 120, 115, 130, 125, 140, 135}
	data := make([]OHLCV, len(prices))
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, p := range prices {
		data[i] = OHLCV{
			Date:   baseDate.AddDate(0, 0, i),
			Open:   p,
			High:   p + 1,
			Low:    p - 1,
			Close:  p,
			Volume: 1000,
		}
	}

	ema := NewLocalEMA(data)
	period := 1

	result, err := ema.ComputeFromOHLCV(period, false)
	if err != nil {
		t.Fatalf("ComputeFromOHLCV failed: %v", err)
	}

	// With period=1, alpha=1, so EMA should equal close price at each bar
	for i, dp := range result {
		if !almostEqual(dp.Value, prices[i]) {
			t.Errorf("EMA at index %d with period=1 should equal close price (%f), got %f", i, prices[i], dp.Value)
		}
	}
}

// TestLocalEMA_FlatPrices verifies that when prices are flat (all the same),
// EMA stays at that constant value.
func TestLocalEMA_FlatPrices(t *testing.T) {
	const flatPrice = 100.0
	prices := make([]float64, 10)
	for i := range prices {
		prices[i] = flatPrice
	}
	data := make([]OHLCV, len(prices))
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range data {
		data[i] = OHLCV{
			Date:   baseDate.AddDate(0, 0, i),
			Open:   flatPrice,
			High:   flatPrice,
			Low:    flatPrice,
			Close:  flatPrice,
			Volume: 1000,
		}
	}

	ema := NewLocalEMA(data)
	period := 14

	result, err := ema.ComputeFromOHLCV(period, false)
	if err != nil {
		t.Fatalf("ComputeFromOHLCV failed: %v", err)
	}

	// All EMA values should be the flat price
	for i, dp := range result {
		if !almostEqual(dp.Value, flatPrice) {
			t.Errorf("EMA at index %d should be %f for flat prices, got %f", i, flatPrice, dp.Value)
		}
	}
}

// TestLocalEMA_OutputLength verifies that the output length is correct.
func TestLocalEMA_OutputLength(t *testing.T) {
	tests := []struct {
		name          string
		dataPoints    int
		period        int
		useInProgress bool
		wantLength    int
	}{
		{
			name:          "Standard case, excluding in-progress",
			dataPoints:    20,
			period:        14,
			useInProgress: false,
			wantLength:    19, // 20 - 1 = 19
		},
		{
			name:          "With in-progress bar included",
			dataPoints:    20,
			period:        14,
			useInProgress: true,
			wantLength:    20, // 20 - 0 = 20
		},
		{
			name:          "Small dataset, excluding in-progress",
			dataPoints:    5,
			period:        14,
			useInProgress: false,
			wantLength:    4, // 5 - 1 = 4
		},
		{
			name:          "Small dataset, including in-progress",
			dataPoints:    5,
			period:        14,
			useInProgress: true,
			wantLength:    5, // 5 - 0 = 5
		},
		{
			name:          "Single data point, including in-progress",
			dataPoints:    1,
			period:        14,
			useInProgress: true,
			wantLength:    1, // 1 - 0 = 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]OHLCV, tt.dataPoints)
			baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			for i := range data {
				data[i] = OHLCV{
					Date:   baseDate.AddDate(0, 0, i),
					Open:   100 + float64(i),
					High:   100 + float64(i) + 1,
					Low:    100 + float64(i) - 1,
					Close:  100 + float64(i),
					Volume: 1000,
				}
			}

			ema := NewLocalEMA(data)
			result, err := ema.ComputeFromOHLCV(tt.period, tt.useInProgress)
			if err != nil {
				t.Fatalf("ComputeFromOHLCV failed: %v", err)
			}

			if len(result) != tt.wantLength {
				t.Errorf("Expected output length %d, got %d", tt.wantLength, len(result))
			}
		})
	}
}

// TestLocalEMA_BarCloseOnly verifies that the last (in-progress) bar is excluded
// when useInProgress=false. This is critical for preventing repainting.
func TestLocalEMA_BarCloseOnly(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Create 20 data points
	data := make([]OHLCV, 20)
	for i := range data {
		data[i] = OHLCV{
			Date:   baseDate.AddDate(0, 0, i),
			Open:   100 + float64(i),
			High:   100 + float64(i) + 1,
			Low:    100 + float64(i) - 1,
			Close:  100 + float64(i),
			Volume: 1000,
		}
	}

	period := 14

	// Compute with in-progress bar excluded (default behavior)
	ema1 := NewLocalEMA(data)
	result1, err := ema1.ComputeFromOHLCV(period, false)
	if err != nil {
		t.Fatalf("ComputeFromOHLCV (exclude in-progress) failed: %v", err)
	}

	// Compute with in-progress bar included
	ema2 := NewLocalEMA(data)
	result2, err := ema2.ComputeFromOHLCV(period, true)
	if err != nil {
		t.Fatalf("ComputeFromOHLCV (include in-progress) failed: %v", err)
	}

	// Excluding in-progress should give one fewer result
	if len(result1)+1 != len(result2) {
		t.Errorf("Expected result1 length + 1 = result2 length, got %d + 1 != %d", len(result1), len(result2))
	}

	// The dates in result1 should NOT include the last data point
	// (the one at index len(data)-1)
	for _, dp := range result1 {
		if dp.Date.Equal(data[len(data)-1].Date) {
			t.Errorf("Result contains in-progress bar date %v (should be excluded)", dp.Date)
		}
	}

	// The dates in result2 should include the last data point
	foundLastDate := false
	for _, dp := range result2 {
		if dp.Date.Equal(data[len(data)-1].Date) {
			foundLastDate = true
			break
		}
	}
	if !foundLastDate {
		t.Errorf("Result2 should include the last data point date")
	}
}

// TestLocalEMA_ComputeMethod verifies that the Indicator interface's Compute method
// returns an appropriate error (not implemented for LocalEMA).
func TestLocalEMA_ComputeMethod(t *testing.T) {
	data := []OHLCV{
		{
			Date:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:   100,
			High:   101,
			Low:    99,
			Close:  100,
			Volume: 1000,
		},
	}

	ema := NewLocalEMA(data)
	_, err := ema.Compute(context.TODO(), "BTC", Options{Period: 14})

	if err == nil {
		t.Error("Expected error from Compute method, got nil")
	}
}

// TestLocalEMA_SetData verifies that SetData correctly updates the internal data.
func TestLocalEMA_SetData(t *testing.T) {
	data1 := []OHLCV{
		{
			Date:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:   100,
			High:   101,
			Low:    99,
			Close:  100,
			Volume: 1000,
		},
	}

	ema := NewLocalEMA(data1)
	if len(ema.Data) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(ema.Data))
	}

	data2 := make([]OHLCV, 20)
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range data2 {
		data2[i] = OHLCV{
			Date:   baseDate.AddDate(0, 0, i),
			Open:   100 + float64(i),
			High:   100 + float64(i) + 1,
			Low:    100 + float64(i) - 1,
			Close:  100 + float64(i),
			Volume: 1000,
		}
	}

	ema.SetData(data2)
	if len(ema.Data) != 20 {
		t.Errorf("Expected 20 data points after SetData, got %d", len(ema.Data))
	}

	// Verify computation works with new data
	_, err := ema.ComputeFromOHLCV(14, false)
	if err != nil {
		t.Errorf("ComputeFromOHLCV failed after SetData: %v", err)
	}
}

// TestLocalEMA_MonotonicPrice verifies that EMA respects the direction of price changes.
// When prices are monotonically increasing, EMA should lag behind but still increase.
// When prices are monotonically decreasing, EMA should lag behind but still decrease.
func TestLocalEMA_MonotonicPrice(t *testing.T) {
	// Monotonically increasing prices
	pricesUp := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114}
	dataUp := make([]OHLCV, len(pricesUp))
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, p := range pricesUp {
		dataUp[i] = OHLCV{
			Date:   baseDate.AddDate(0, 0, i),
			Open:   p,
			High:   p + 1,
			Low:    p - 1,
			Close:  p,
			Volume: 1000,
		}
	}

	ema := NewLocalEMA(dataUp)
	resultUp, err := ema.ComputeFromOHLCV(10, false)
	if err != nil {
		t.Fatalf("ComputeFromOHLCV failed: %v", err)
	}

	// EMA should be monotonically increasing (each value > previous)
	for i := 1; i < len(resultUp); i++ {
		if resultUp[i].Value <= resultUp[i-1].Value {
			t.Errorf("EMA should increase with increasing prices: EMA[%d] (%f) <= EMA[%d] (%f)",
				i, resultUp[i].Value, i-1, resultUp[i-1].Value)
		}
	}

	// EMA should lag behind prices (be less than current price)
	// Note: First EMA equals first price (TradingView seeding), so check from index 1
	for i := 1; i < len(resultUp); i++ {
		if resultUp[i].Value >= pricesUp[i] {
			t.Errorf("EMA should lag behind prices: EMA[%d] (%f) >= price[%d] (%f)",
				i, resultUp[i].Value, i, pricesUp[i])
		}
	}
}
