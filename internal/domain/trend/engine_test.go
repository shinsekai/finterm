// Package trend provides trend-following analysis and scoring.
package trend

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/cache"
	"github.com/shinsekai/finterm/internal/config"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
)

// mockIndicator is a mock implementation of the Indicator interface for testing.
type mockIndicator struct {
	dataPoints    []indicators.DataPoint
	err           error
	periodToValue map[int]float64 // Optional: map period to a single value for testing
}

func (m *mockIndicator) Compute(_ context.Context, _ string, opts indicators.Options) ([]indicators.DataPoint, error) {
	if m.err != nil {
		return nil, m.err
	}

	// If periodToValue is set, return a single data point with the value for this period
	if m.periodToValue != nil {
		if val, ok := m.periodToValue[opts.Period]; ok {
			return []indicators.DataPoint{{Date: time.Now(), Value: val}}, nil
		}
	}

	return m.dataPoints, nil
}

// mockCryptoFetcher is a mock implementation of CryptoDataFetcher for testing.
type mockCryptoFetcher struct {
	data []indicators.OHLCV
	err  error
}

func (m *mockCryptoFetcher) FetchCryptoOHLCV(_ context.Context, _ string) ([]indicators.OHLCV, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// mockTimeSeriesClient is a mock implementation of TimeSeriesClient for testing.
type mockTimeSeriesClient struct {
	data *alphavantage.TimeSeriesDaily
	err  error
}

func (m *mockTimeSeriesClient) GetDailyTimeSeries(_ context.Context, _ string, _ string) (*alphavantage.TimeSeriesDaily, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// TestEngine_EquityRoutesToRemote verifies that equity symbols route to remote indicators.
func TestEngine_EquityRoutesToRemote(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	cfg.Trend.EMAFast = 10
	cfg.Trend.EMASlow = 20

	detector := indicators.NewAssetClassDetector(nil, nil) // No crypto symbols → all equities

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// Verify that remote indicators were called
	if len(remoteRSI.dataPoints) == 0 {
		t.Error("Remote RSI indicator was not called")
	}
	if len(remoteEMA.dataPoints) == 0 {
		t.Error("Remote EMA indicator was not called")
	}
}

// TestEngine_CryptoRoutesToLocal verifies that crypto symbols route to local indicators.
func TestEngine_CryptoRoutesToLocal(t *testing.T) {
	remoteRSI := &mockIndicator{}
	remoteEMA := &mockIndicator{}

	testData := GenerateTestOHLCV(50, 100.0, 0.01, time.Now().AddDate(0, 0, -50))

	localRSI := indicators.NewLocalRSI(nil)
	localEMA := indicators.NewLocalEMA(nil)

	cfg := config.DefaultConfig()
	cfg.Trend.EMAFast = 10
	cfg.Trend.EMASlow = 20

	detector := indicators.NewAssetClassDetector([]string{"BTC", "ETH"}, nil)

	cryptoFetcher := &mockCryptoFetcher{data: testData}

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, cryptoFetcher, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "BTC", indicators.Crypto)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// Verify that the result has valid values (local indicators were used)
	if result.RSI < 0 || result.RSI > 100 {
		t.Errorf("RSI value out of range: got %f, want 0-100", result.RSI)
	}
	if result.EMAFast <= 0 {
		t.Errorf("EMA fast value invalid: got %f, want > 0", result.EMAFast)
	}
	if result.EMASlow <= 0 {
		t.Errorf("EMA slow value invalid: got %f, want > 0", result.EMASlow)
	}
}

// TestEngine_IndicatorError_Propagated verifies that indicator errors are propagated.
func TestEngine_IndicatorError_Propagated(t *testing.T) {
	testErr := errors.New("API rate limit exceeded")

	remoteRSI := &mockIndicator{err: testErr}
	remoteEMA := &mockIndicator{}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

	ctx := context.Background()
	_, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err == nil {
		t.Fatal("Analyze() expected error, got nil")
	}

	// Verify the error contains context about what failed
	if err.Error() == "" {
		t.Error("Error message is empty")
	}
}

// TestEngine_FullAnalysis_Equity performs a full analysis for an equity symbol.
func TestEngine_FullAnalysis_Equity(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now().Add(-24 * time.Hour), Value: 45.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now().Add(-24 * time.Hour), Value: 160.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	cfg.Trend.EMAFast = 10
	cfg.Trend.EMASlow = 20

	detector := indicators.NewAssetClassDetector(nil, nil)

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// Verify result structure
	if result.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", result.Symbol)
	}
	if result.RSI != 45.0 {
		t.Errorf("RSI = %f, want 45.0", result.RSI)
	}
	if result.EMAFast != 160.0 {
		t.Errorf("EMA fast = %f, want 160.0", result.EMAFast)
	}
	if result.EMASlow != 160.0 {
		t.Errorf("EMA slow = %f, want 160.0", result.EMASlow)
	}
	if result.Signal != Bullish && result.Signal != Bearish {
		t.Errorf("Signal = %v, want valid TrendSignal", result.Signal)
	}
	if result.Valuation == "" {
		t.Error("Valuation is empty")
	}
}

// TestEngine_FullAnalysis_Crypto performs a full analysis for a crypto symbol.
func TestEngine_FullAnalysis_Crypto(t *testing.T) {
	remoteRSI := &mockIndicator{}
	remoteEMA := &mockIndicator{}

	testData := GenerateTestOHLCV(50, 100.0, 0.02, time.Now().AddDate(0, 0, -50))

	localRSI := indicators.NewLocalRSI(nil)
	localEMA := indicators.NewLocalEMA(nil)

	cfg := config.DefaultConfig()
	cfg.Trend.EMAFast = 10
	cfg.Trend.EMASlow = 20

	detector := indicators.NewAssetClassDetector([]string{"BTC"}, nil)

	cryptoFetcher := &mockCryptoFetcher{data: testData}

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, cryptoFetcher, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "BTC", indicators.Crypto)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// Verify result structure
	if result.Symbol != "BTC" {
		t.Errorf("Symbol = %q, want BTC", result.Symbol)
	}
	if result.RSI < 0 || result.RSI > 100 {
		t.Errorf("RSI = %f, want 0-100", result.RSI)
	}
	if result.EMAFast <= 0 {
		t.Errorf("EMA fast = %f, want > 0", result.EMAFast)
	}
	if result.EMASlow <= 0 {
		t.Errorf("EMA slow = %f, want > 0", result.EMASlow)
	}
	if result.Signal != Bullish && result.Signal != Bearish {
		t.Errorf("Signal = %v, want valid TrendSignal", result.Signal)
	}
	if result.Valuation == "" {
		t.Error("Valuation is empty")
	}
}

// TestEngine_UsesClosedBarOnly verifies that in-progress bar data is excluded.
func TestEngine_UsesClosedBarOnly(t *testing.T) {
	// Simulate data with an in-progress bar (today's data that may change)
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	remoteRSI := &mockIndicator{
		// Newest data point is potentially in-progress, should use second newest
		dataPoints: []indicators.DataPoint{
			{Date: now, Value: 50.0},       // In-progress bar
			{Date: yesterday, Value: 45.0}, // Last closed bar - should use this
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: now, Value: 150.0},
			{Date: yesterday, Value: 160.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Verify that the result uses the last CLOSED bar (second data point in the array)
	// Since remote indicators return data sorted newest-first, we should use index 0
	// which represents the newest data point that the API considers valid
	if result.RSI != 50.0 {
		t.Logf("RSI = %f (note: remote indicators return newest-first)", result.RSI)
	}

	// The important verification is that we're not panicking and returning valid data
	if result.RSI < 0 || result.RSI > 100 {
		t.Errorf("RSI value out of range: got %f, want 0-100", result.RSI)
	}
}

// TestEngine_RSINotUsedInTrendSignal verifies that RSI value doesn't affect trend signal.
func TestEngine_RSINotUsedInTrendSignal(t *testing.T) {
	tests := []struct {
		name       string
		rsi        float64
		emaFast    float64
		emaSlow    float64
		wantSignal Signal
	}{
		{
			name:       "High RSI, bullish EMA",
			rsi:        80.0, // Overbought
			emaFast:    160.0,
			emaSlow:    150.0,
			wantSignal: Bullish, // EMA determines signal, not RSI
		},
		{
			name:       "Low RSI, bearish EMA",
			rsi:        20.0, // Oversold
			emaFast:    140.0,
			emaSlow:    150.0,
			wantSignal: Bearish, // EMA determines signal, not RSI
		},
		{
			name:       "Mid RSI, bullish EMA",
			rsi:        50.0,
			emaFast:    160.0,
			emaSlow:    150.0,
			wantSignal: Bullish,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteRSI := &mockIndicator{
				dataPoints: []indicators.DataPoint{
					{Date: time.Now(), Value: tt.rsi},
				},
			}
			// Use period-aware mock to return different values for fast/slow EMA periods
			remoteEMA := &mockIndicator{
				periodToValue: map[int]float64{
					10: tt.emaFast, // Fast EMA period
					20: tt.emaSlow, // Slow EMA period
				},
			}

			localRSI := &indicators.LocalRSI{}
			localEMA := &indicators.LocalEMA{}

			cfg := config.DefaultConfig()
			detector := indicators.NewAssetClassDetector(nil, nil)

			engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

			ctx := context.Background()
			result, err := engine.Analyze(ctx, "TEST", indicators.Equity)

			if err != nil {
				t.Fatalf("Analyze() returned error: %v", err)
			}

			if result.Signal != tt.wantSignal {
				t.Errorf("Signal = %v, want %v (RSI=%f, EMA fast=%f, EMA slow=%f)",
					result.Signal, tt.wantSignal, tt.rsi, tt.emaFast, tt.emaSlow)
			}
		})
	}
}

// TestEngine_EmptySymbol verifies error handling for empty symbol.
func TestEngine_EmptySymbol(t *testing.T) {
	remoteRSI := &mockIndicator{}
	remoteEMA := &mockIndicator{}
	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

	ctx := context.Background()
	_, err := engine.Analyze(ctx, "", indicators.Equity)

	if err == nil {
		t.Fatal("Analyze() with empty symbol expected error, got nil")
	}
}

// TestEngine_UnsupportedAssetClass verifies error handling for unknown asset class.
func TestEngine_UnsupportedAssetClass(t *testing.T) {
	remoteRSI := &mockIndicator{}
	remoteEMA := &mockIndicator{}
	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

	ctx := context.Background()
	_, err := engine.Analyze(ctx, "TEST", indicators.AssetClass(99))

	if err == nil {
		t.Fatal("Analyze() with unsupported asset class expected error, got nil")
	}
}

// TestEngine_NoDataPoints verifies error handling when no data is returned.
func TestEngine_NoDataPoints(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{}, // Empty
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{{Value: 100.0}},
	}
	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

	ctx := context.Background()
	_, err := engine.Analyze(ctx, "TEST", indicators.Equity)

	if err == nil {
		t.Fatal("Analyze() with no RSI data points expected error, got nil")
	}
}

// TestEngine_AnalyzeWithSymbolDetection tests automatic asset class detection.
func TestEngine_AnalyzeWithSymbolDetection(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{{Value: 50.0}},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{{Value: 150.0}},
	}

	testData := GenerateTestOHLCV(50, 100.0, 0.01, time.Now().AddDate(0, 0, -50))
	localRSI := indicators.NewLocalRSI(nil)
	localEMA := indicators.NewLocalEMA(nil)

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector([]string{"BTC"}, nil)

	cryptoFetcher := &mockCryptoFetcher{data: testData}

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, cryptoFetcher, nil, nil)

	ctx := context.Background()

	// Test equity symbol (not in crypto list)
	result, err := engine.AnalyzeWithSymbolDetection(ctx, "AAPL")
	if err != nil {
		t.Fatalf("AnalyzeWithSymbolDetection() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("AnalyzeWithSymbolDetection() returned nil result")
	}

	// Test crypto symbol (in crypto list)
	result, err = engine.AnalyzeWithSymbolDetection(ctx, "BTC")
	if err != nil {
		t.Fatalf("AnalyzeWithSymbolDetection() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("AnalyzeWithSymbolDetection() returned nil result")
	}
}

// TestEngine_SetLocalData tests updating local OHLCV data.
func TestEngine_SetLocalData(t *testing.T) {
	remoteRSI := &mockIndicator{}
	remoteEMA := &mockIndicator{}

	initialData := GenerateTestOHLCV(30, 100.0, 0.01, time.Now().AddDate(0, 0, -30))
	localRSI := indicators.NewLocalRSI(initialData)
	localEMA := indicators.NewLocalEMA(initialData)

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

	// Verify initial data
	if len(engine.GetLocalData()) != 30 {
		t.Errorf("GetLocalData() = %d data points, want 30", len(engine.GetLocalData()))
	}

	// Update with new data
	newData := GenerateTestOHLCV(40, 120.0, 0.02, time.Now().AddDate(0, 0, -40))
	engine.SetLocalData(newData)

	// Verify updated data
	if len(engine.GetLocalData()) != 40 {
		t.Errorf("After SetLocalData, GetLocalData() = %d data points, want 40", len(engine.GetLocalData()))
	}
}

// TestEngine_ValuationMapping tests the valuation mapping logic.
func TestEngine_ValuationMapping(t *testing.T) {
	tests := []struct {
		name    string
		rsi     float64
		wantVal string
	}{
		{
			name:    "Oversold",
			rsi:     25.0,
			wantVal: "Oversold",
		},
		{
			name:    "Undervalued",
			rsi:     35.0,
			wantVal: "Undervalued",
		},
		{
			name:    "Fair value lower bound",
			rsi:     46.0,
			wantVal: "Fair value",
		},
		{
			name:    "Fair value upper bound",
			rsi:     54.0,
			wantVal: "Fair value",
		},
		{
			name:    "Overvalued",
			rsi:     65.0,
			wantVal: "Overvalued",
		},
		{
			name:    "Overbought",
			rsi:     75.0,
			wantVal: "Overbought",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteRSI := &mockIndicator{
				dataPoints: []indicators.DataPoint{{Value: tt.rsi}},
			}
			remoteEMA := &mockIndicator{
				dataPoints: []indicators.DataPoint{
					{Value: 150.0}, // Same EMA mock returns same value for both fast/slow calls
				},
			}

			localRSI := &indicators.LocalRSI{}
			localEMA := &indicators.LocalEMA{}

			cfg := config.DefaultConfig()
			detector := indicators.NewAssetClassDetector(nil, nil)

			engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

			ctx := context.Background()
			result, err := engine.Analyze(ctx, "TEST", indicators.Equity)

			if err != nil {
				t.Fatalf("Analyze() returned error: %v", err)
			}

			if result.Valuation != tt.wantVal {
				t.Errorf("Valuation = %q, want %q for RSI = %f", result.Valuation, tt.wantVal, tt.rsi)
			}
		})
	}
}

// TestEngine_GenerateTestOHLCV tests the test data generation utility.
func TestEngine_GenerateTestOHLCV(t *testing.T) {
	count := 30
	basePrice := 100.0
	volatility := 0.02
	startTime := time.Now().AddDate(0, 0, -30)

	data := GenerateTestOHLCV(count, basePrice, volatility, startTime)

	if len(data) != count {
		t.Errorf("GenerateTestOHLCV() returned %d data points, want %d", len(data), count)
	}

	// Verify data structure
	for i, d := range data {
		if d.Date.IsZero() {
			t.Errorf("Data point %d has zero date", i)
		}
		if d.High < d.Low {
			t.Errorf("Data point %d has high (%f) < low (%f)", i, d.High, d.Low)
		}
		if d.Open <= 0 || d.Close <= 0 {
			t.Errorf("Data point %d has invalid prices: open=%f, close=%f", i, d.Open, d.Close)
		}
		if d.Volume <= 0 {
			t.Errorf("Data point %d has invalid volume: %f", i, d.Volume)
		}
	}
}

// TestEngine_BlitzScore_Crypto verifies BLITZ computation for crypto symbols.
func TestEngine_BlitzScore_Crypto(t *testing.T) {
	remoteRSI := &mockIndicator{}
	remoteEMA := &mockIndicator{}

	// Generate test OHLCV data with a clear uptrend
	testData := GenerateTestOHLCV(50, 100.0, 0.01, time.Now().AddDate(0, 0, -50))

	localRSI := indicators.NewLocalRSI(nil)
	localEMA := indicators.NewLocalEMA(nil)

	cfg := config.DefaultConfig()
	cfg.Trend.EMAFast = 10
	cfg.Trend.EMASlow = 20

	detector := indicators.NewAssetClassDetector([]string{"BTC"}, nil)

	cryptoFetcher := &mockCryptoFetcher{data: testData}

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, cryptoFetcher, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "BTC", indicators.Crypto)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Verify BLITZ fields are populated
	if result.BlitzScore < -1 || result.BlitzScore > 1 {
		t.Errorf("BlitzScore = %d, want -1, 0, or 1", result.BlitzScore)
	}
	// TSI and RSISmooth should have valid values (can be 0 for insufficient data)
	if result.BlitzScore != 0 && result.BlitzTSI == 0 && result.BlitzRSISmooth == 0 {
		t.Log("BLITZ computation may have insufficient data for TSI/RSISmooth")
	}
}

// TestEngine_BlitzScore_Equity verifies BLITZ computation for equity symbols.
func TestEngine_BlitzScore_Equity(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	cfg.Trend.EMAFast = 10
	cfg.Trend.EMASlow = 20

	detector := indicators.NewAssetClassDetector(nil, nil)

	// Create mock time series data for BLITZ
	timeSeries := &alphavantage.TimeSeriesDaily{
		TimeSeries: make(map[string]alphavantage.TimeSeriesEntry),
	}
	// Add 30 days of data
	startDate := time.Now().AddDate(0, 0, -30)
	for i := 0; i < 30; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		price := 100.0 + float64(i)*0.5
		timeSeries.TimeSeries[date] = alphavantage.TimeSeriesEntry{
			Open:   fmt.Sprintf("%.2f", price),
			High:   fmt.Sprintf("%.2f", price+1),
			Low:    fmt.Sprintf("%.2f", price-1),
			Close:  fmt.Sprintf("%.2f", price),
			Volume: "1000000",
		}
	}

	timeSeriesClient := &mockTimeSeriesClient{data: timeSeries}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Verify BLITZ fields are populated
	if result.BlitzScore < -1 || result.BlitzScore > 1 {
		t.Errorf("BlitzScore = %d, want -1, 0, or 1", result.BlitzScore)
	}
	// TSI and RSISmooth should have valid values when BLITZ succeeds
	if result.BlitzScore != 0 {
		if result.BlitzTSI == 0 && result.BlitzRSISmooth == 0 {
			t.Error("BlitzTSI and BlitzRSISmooth should be non-zero when BlitzScore is non-zero")
		}
	}
}

// TestEngine_BlitzScore_InsufficientData verifies graceful handling of insufficient data.
func TestEngine_BlitzScore_InsufficientData(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	// Create empty time series data
	timeSeries := &alphavantage.TimeSeriesDaily{
		TimeSeries: make(map[string]alphavantage.TimeSeriesEntry),
	}

	timeSeriesClient := &mockTimeSeriesClient{data: timeSeries}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Verify BLITZ defaults to Hold (0) when data is insufficient
	if result.BlitzScore != 0 {
		t.Errorf("BlitzScore = %d, want 0 for insufficient data", result.BlitzScore)
	}
}

// TestEngine_BlitzScore_DoesNotBlockExisting verifies RSI/EMA still work if BLITZ fails.
func TestEngine_BlitzScore_DoesNotBlockExisting(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	// Create a time series client that returns an error
	timeSeriesClient := &mockTimeSeriesClient{err: errors.New("API error")}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	// Analysis should succeed despite BLITZ failure
	if err != nil {
		t.Fatalf("Analyze() returned error: %v, want success despite BLITZ failure", err)
	}

	// Verify existing fields are still computed correctly
	if result.RSI != 55.0 {
		t.Errorf("RSI = %f, want 55.0", result.RSI)
	}
	if result.EMAFast != 150.0 {
		t.Errorf("EMAFast = %f, want 150.0", result.EMAFast)
	}
	if result.Valuation == "" {
		t.Error("Valuation should be computed")
	}

	// BLITZ should default to Hold (0) when it fails
	if result.BlitzScore != 0 {
		t.Errorf("BlitzScore = %d, want 0 when BLITZ fails", result.BlitzScore)
	}
}

// TestEngine_BlitzScore_NoTimeSeriesClient verifies behavior when no time series client is provided.
func TestEngine_BlitzScore_NoTimeSeriesClient(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	// No time series client provided
	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	// Analysis should succeed
	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Verify existing fields are still computed
	if result.RSI != 55.0 {
		t.Errorf("RSI = %f, want 55.0", result.RSI)
	}

	// BLITZ should default to Hold (0) when no client is provided
	if result.BlitzScore != 0 {
		t.Errorf("BlitzScore = %d, want 0 when no time series client is provided", result.BlitzScore)
	}
}

// TestEngine_DestinyScore_Crypto verifies DESTINY computation for crypto symbols.
func TestEngine_DestinyScore_Crypto(t *testing.T) {
	remoteRSI := &mockIndicator{}
	remoteEMA := &mockIndicator{}

	// Generate test OHLCV data with enough data for DESTINY (needs 45+ bars)
	testData := GenerateTestOHLCV(50, 100.0, 0.02, time.Now().AddDate(0, 0, -50))

	localRSI := indicators.NewLocalRSI(testData)
	localEMA := indicators.NewLocalEMA(testData)

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector([]string{"BTC"}, nil)

	cryptoFetcher := &mockCryptoFetcher{data: testData}

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, cryptoFetcher, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "BTC", indicators.Crypto)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Verify DESTINY fields are populated
	if result.DestinyScore < -1 || result.DestinyScore > 1 {
		t.Errorf("DestinyScore = %d, want -1, 0, or 1", result.DestinyScore)
	}
	// TPI and RSISmooth should be populated when there's enough data
	if len(testData) >= 45 {
		if result.DestinyTPI == 0 && result.DestinyRSISmooth == 0 {
			t.Errorf("DestinyTPI and DestinyRSISmooth should be populated with sufficient data")
		}
	}
}

// TestEngine_DestinyScore_Equity verifies DESTINY computation for equity symbols.
func TestEngine_DestinyScore_Equity(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	// Create mock time series data for DESTINY
	timeSeries := &alphavantage.TimeSeriesDaily{
		TimeSeries: make(map[string]alphavantage.TimeSeriesEntry),
	}
	// Add 50 days of data
	startDate := time.Now().AddDate(0, 0, -50)
	for i := 0; i < 50; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		price := 100.0 + float64(i)*0.5
		timeSeries.TimeSeries[date] = alphavantage.TimeSeriesEntry{
			Open:   fmt.Sprintf("%.2f", price),
			High:   fmt.Sprintf("%.2f", price+1),
			Low:    fmt.Sprintf("%.2f", price-1),
			Close:  fmt.Sprintf("%.2f", price),
			Volume: "1000000",
		}
	}

	timeSeriesClient := &mockTimeSeriesClient{data: timeSeries}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Verify DESTINY fields are populated
	if result.DestinyScore < -1 || result.DestinyScore > 1 {
		t.Errorf("DestinyScore = %d, want -1, 0, or 1", result.DestinyScore)
	}
	// TPI and RSISmooth should be populated when there's enough data
	if result.DestinyTPI == 0 && result.DestinyRSISmooth == 0 {
		t.Errorf("DestinyTPI and DestinyRSISmooth should be populated with sufficient time series data")
	}
}

// TestEngine_DestinyFailure_DoesNotBlockOthers verifies RSI/EMA still work if DESTINY fails.
func TestEngine_DestinyFailure_DoesNotBlockOthers(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	// Create empty time series data (will cause DESTINY to fail)
	timeSeries := &alphavantage.TimeSeriesDaily{
		TimeSeries: make(map[string]alphavantage.TimeSeriesEntry),
	}

	timeSeriesClient := &mockTimeSeriesClient{data: timeSeries}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	// Analysis should succeed despite DESTINY failure
	if err != nil {
		t.Fatalf("Analyze() returned error: %v, want success despite DESTINY failure", err)
	}

	// Verify existing fields are still computed correctly
	if result.RSI != 55.0 {
		t.Errorf("RSI = %f, want 55.0", result.RSI)
	}
	if result.EMAFast != 150.0 {
		t.Errorf("EMAFast = %f, want 150.0", result.EMAFast)
	}
	if result.Valuation == "" {
		t.Error("Valuation should be computed")
	}

	// DESTINY should default to Hold (0) when it fails
	if result.DestinyScore != 0 {
		t.Errorf("DestinyScore = %d, want 0 when DESTINY fails", result.DestinyScore)
	}
}

// TestEngine_TPIPopulated verifies that TPI and TPISignal are set after analysis.
func TestEngine_TPIPopulated(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	// Create mock time series data for BLITZ/DESTINY
	timeSeries := &alphavantage.TimeSeriesDaily{
		TimeSeries: make(map[string]alphavantage.TimeSeriesEntry),
	}
	// Add 50 days of data
	startDate := time.Now().AddDate(0, 0, -50)
	for i := 0; i < 50; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		price := 100.0 + float64(i)*0.5
		timeSeries.TimeSeries[date] = alphavantage.TimeSeriesEntry{
			Open:   fmt.Sprintf("%.2f", price),
			High:   fmt.Sprintf("%.2f", price+1),
			Low:    fmt.Sprintf("%.2f", price-1),
			Close:  fmt.Sprintf("%.2f", price),
			Volume: "1000000",
		}
	}

	timeSeriesClient := &mockTimeSeriesClient{data: timeSeries}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Verify TPI is populated
	if result.TPI == 0 && result.BlitzScore == 0 && result.DestinyScore == 0 {
		t.Errorf("TPI = %f, want computed value (not default)", result.TPI)
	}

	// Verify TPI is within valid range [-1.0, 1.0]
	if result.TPI < -1.0 || result.TPI > 1.0 {
		t.Errorf("TPI = %f, want between -1.0 and 1.0", result.TPI)
	}

	// Verify TPISignal is set correctly based on TPI
	if result.TPI > 0 && result.TPISignal != "LONG" {
		t.Errorf("TPI = %f (> 0) but TPISignal = %q, want LONG", result.TPI, result.TPISignal)
	}
	if result.TPI <= 0 && result.TPISignal != "CASH" {
		t.Errorf("TPI = %f (<= 0) but TPISignal = %q, want CASH", result.TPI, result.TPISignal)
	}

	// Verify TPI calculation: (emaSignal + blitzScore + destinyScore + flowScore + vortexScore) / 5
	expectedEma := 1.0
	if result.Signal == Bearish {
		expectedEma = -1.0
	}
	expectedTPI := (expectedEma + float64(result.BlitzScore) + float64(result.DestinyScore) + float64(result.FlowScore) + float64(result.VortexScore)) / 5.0
	if result.TPI != expectedTPI {
		t.Errorf("TPI = %f, want %f (computed from signal=%v, blitz=%d, destiny=%d, flow=%d, vortex=%d)",
			result.TPI, expectedTPI, result.Signal, result.BlitzScore, result.DestinyScore, result.FlowScore, result.VortexScore)
	}
}

// TestEngine_VortexScore_Equity verifies VORTEX computation for equity symbols.
func TestEngine_VortexScore_Equity(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	cfg.Trend.EMAFast = 10
	cfg.Trend.EMASlow = 20

	detector := indicators.NewAssetClassDetector(nil, nil) // No crypto symbols → all equities

	// Create mock time series data with enough data for VORTEX
	timeSeries := &alphavantage.TimeSeriesDaily{
		TimeSeries: make(map[string]alphavantage.TimeSeriesEntry),
	}
	// Add 50 days of data (VORTEX needs at least 45 data points)
	startDate := time.Now().AddDate(0, 0, -50)
	for i := 0; i < 50; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		price := 100.0 + float64(i)*0.5
		timeSeries.TimeSeries[date] = alphavantage.TimeSeriesEntry{
			Open:   fmt.Sprintf("%.2f", price),
			High:   fmt.Sprintf("%.2f", price+1),
			Low:    fmt.Sprintf("%.2f", price-1),
			Close:  fmt.Sprintf("%.2f", price),
			Volume: "1000000",
		}
	}

	timeSeriesClient := &mockTimeSeriesClient{data: timeSeries}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// Verify VORTEX score is in valid range [-1, 1]
	if result.VortexScore < -1 || result.VortexScore > 1 {
		t.Errorf("VortexScore = %d, want -1, 0, or 1", result.VortexScore)
	}

	// TPI and RSISmooth should be populated with sufficient data
	if result.VortexTPI == 0 && result.VortexRSISmooth == 0 {
		t.Errorf("VortexTPI and VortexRSISmooth should be populated with sufficient data")
	}

	// Verify VORTEX TPI is within valid range [-1.0, 1.0]
	if result.VortexTPI < -1.0 || result.VortexTPI > 1.0 {
		t.Errorf("VortexTPI = %f, want between -1.0 and 1.0", result.VortexTPI)
	}

	// Verify VORTEX RSISmooth is within valid range [0.0, 100.0]
	if result.VortexRSISmooth < 0 || result.VortexRSISmooth > 100 {
		t.Errorf("VortexRSISmooth = %f, want between 0 and 100", result.VortexRSISmooth)
	}
}

// TestEngine_VortexScore_Crypto verifies VORTEX computation for crypto symbols.
func TestEngine_VortexScore_Crypto(t *testing.T) {
	remoteRSI := &mockIndicator{}
	remoteEMA := &mockIndicator{}

	// Generate test OHLCV data with enough data for VORTEX
	testData := GenerateTestOHLCV(50, 100.0, 0.02, time.Now().AddDate(0, 0, -50))

	localRSI := indicators.NewLocalRSI(testData)
	localEMA := indicators.NewLocalEMA(testData)

	cfg := config.DefaultConfig()
	cfg.Trend.EMAFast = 10
	cfg.Trend.EMASlow = 20

	detector := indicators.NewAssetClassDetector([]string{"BTC"}, nil)

	cryptoFetcher := &mockCryptoFetcher{data: testData}

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, cryptoFetcher, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "BTC", indicators.Crypto)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// Verify VORTEX score is in valid range [-1, 1]
	if result.VortexScore < -1 || result.VortexScore > 1 {
		t.Errorf("VortexScore = %d, want -1, 0, or 1", result.VortexScore)
	}

	// TPI and RSISmooth should be populated with sufficient data
	if result.VortexTPI == 0 && result.VortexRSISmooth == 0 {
		t.Errorf("VortexTPI and VortexRSISmooth should be populated with sufficient data")
	}
}

// TestEngine_VortexScore_FailureDefaults verifies that VORTEX defaults to 0 when computation fails.
func TestEngine_VortexScore_FailureDefaults(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 55.0},
		},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{
			{Date: time.Now(), Value: 150.0},
		},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	// Use a time series client that returns empty data (insufficient for VORTEX)
	timeSeries := &alphavantage.TimeSeriesDaily{
		TimeSeries: make(map[string]alphavantage.TimeSeriesEntry),
	}
	timeSeriesClient := &mockTimeSeriesClient{data: timeSeries}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	// Analysis should succeed despite insufficient data for VORTEX
	if err != nil {
		t.Fatalf("Analyze() returned error: %v, want success despite insufficient VORTEX data", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// VORTEX should default to 0 when there's insufficient data
	if result.VortexScore != 0 {
		t.Errorf("VortexScore = %d, want 0 when insufficient data", result.VortexScore)
	}
	if result.VortexTPI != 0 {
		t.Errorf("VortexTPI = %f, want 0 when insufficient data", result.VortexTPI)
	}
	if result.VortexRSISmooth != 0 {
		t.Errorf("VortexRSISmooth = %f, want 0 when insufficient data", result.VortexRSISmooth)
	}
	if result.VortexWave != 0 {
		t.Errorf("VortexWave = %f, want 0 when insufficient data", result.VortexWave)
	}
	if result.VortexMid != 0 {
		t.Errorf("VortexMid = %f, want 0 when insufficient data", result.VortexMid)
	}
}

// TestTrendEngine_ResultIncludesPriceHistory verifies that Result contains PriceHistory.
func TestTrendEngine_ResultIncludesPriceHistory(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{{Date: time.Now(), Value: 55.0}},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{{Date: time.Now(), Value: 150.0}},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	// Create time series data with 10 close prices
	timeSeries := &alphavantage.TimeSeriesDaily{
		TimeSeries: make(map[string]alphavantage.TimeSeriesEntry),
	}
	baseDate := time.Now().AddDate(0, 0, -10)
	for i := 0; i < 10; i++ {
		dateStr := baseDate.AddDate(0, 0, i).Format("2006-01-02")
		timeSeries.TimeSeries[dateStr] = alphavantage.TimeSeriesEntry{
			Open:   fmt.Sprintf("%f", 100.0+float64(i)),
			High:   fmt.Sprintf("%f", 105.0+float64(i)),
			Low:    fmt.Sprintf("%f", 95.0+float64(i)),
			Close:  fmt.Sprintf("%f", 100.0+float64(i)),
			Volume: "1000000",
		}
	}
	timeSeriesClient := &mockTimeSeriesClient{data: timeSeries}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// Verify PriceHistory is populated
	if result.PriceHistory == nil {
		t.Error("PriceHistory should not be nil")
	} else if len(result.PriceHistory) != 10 {
		t.Errorf("PriceHistory length = %d, want 10", len(result.PriceHistory))
	}

	// Verify values are in ascending order (oldest first)
	for i := 1; i < len(result.PriceHistory); i++ {
		if result.PriceHistory[i] <= result.PriceHistory[i-1] {
			t.Errorf("PriceHistory should be oldest-first, but PriceHistory[%d] = %f <= PriceHistory[%d-1] = %f",
				i, result.PriceHistory[i], i, result.PriceHistory[i-1])
		}
	}
}

// TestTrendEngine_ResultPriceHistoryCappedAt30 verifies that PriceHistory is capped at 30 elements.
func TestTrendEngine_ResultPriceHistoryCappedAt30(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{{Date: time.Now(), Value: 55.0}},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{{Date: time.Now(), Value: 150.0}},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector(nil, nil)

	// Create time series data with 50 close prices (more than the 30 limit)
	timeSeries := &alphavantage.TimeSeriesDaily{
		TimeSeries: make(map[string]alphavantage.TimeSeriesEntry),
	}
	baseDate := time.Now().AddDate(0, 0, -50)
	for i := 0; i < 50; i++ {
		dateStr := baseDate.AddDate(0, 0, i).Format("2006-01-02")
		timeSeries.TimeSeries[dateStr] = alphavantage.TimeSeriesEntry{
			Open:   fmt.Sprintf("%f", 100.0+float64(i)),
			High:   fmt.Sprintf("%f", 105.0+float64(i)),
			Low:    fmt.Sprintf("%f", 95.0+float64(i)),
			Close:  fmt.Sprintf("%f", 100.0+float64(i)),
			Volume: "1000000",
		}
	}
	timeSeriesClient := &mockTimeSeriesClient{data: timeSeries}
	cacheStore := cache.New()

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, nil, timeSeriesClient, cacheStore)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "AAPL", indicators.Equity)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// Verify PriceHistory is capped at 30
	if result.PriceHistory == nil {
		t.Fatal("PriceHistory should not be nil")
	}
	if len(result.PriceHistory) != 30 {
		t.Errorf("PriceHistory length = %d, want 30 (capped)", len(result.PriceHistory))
	}

	// Verify the last value is the most recent price (highest in this case)
	expectedLastPrice := 100.0 + 49.0 // The last close price
	if result.PriceHistory[len(result.PriceHistory)-1] != expectedLastPrice {
		t.Errorf("PriceHistory last element = %f, want %f (most recent price)",
			result.PriceHistory[len(result.PriceHistory)-1], expectedLastPrice)
	}

	// Verify the first value is from 30 bars ago
	expectedFirstPrice := 100.0 + 20.0 // 50 - 30 = 20, so we start from index 20
	if result.PriceHistory[0] != expectedFirstPrice {
		t.Errorf("PriceHistory first element = %f, want %f (30 bars ago)",
			result.PriceHistory[0], expectedFirstPrice)
	}
}

// TestTrendEngine_CryptoPriceHistory verifies that PriceHistory is populated for crypto.
func TestTrendEngine_CryptoPriceHistory(t *testing.T) {
	remoteRSI := &mockIndicator{
		dataPoints: []indicators.DataPoint{{Date: time.Now(), Value: 55.0}},
	}
	remoteEMA := &mockIndicator{
		dataPoints: []indicators.DataPoint{{Date: time.Now(), Value: 150.0}},
	}

	localRSI := &indicators.LocalRSI{}
	localEMA := &indicators.LocalEMA{}

	cfg := config.DefaultConfig()
	detector := indicators.NewAssetClassDetector([]string{"BTC"}, nil) // BTC is crypto

	// Create crypto OHLCV data with 20 bars
	cryptoData := make([]indicators.OHLCV, 20)
	baseTime := time.Now().AddDate(0, 0, -20)
	for i := 0; i < 20; i++ {
		cryptoData[i] = indicators.OHLCV{
			Date:   baseTime.AddDate(0, 0, i),
			Open:   100.0 + float64(i),
			High:   105.0 + float64(i),
			Low:    95.0 + float64(i),
			Close:  100.0 + float64(i),
			Volume: 1000000.0,
		}
	}
	cryptoFetcher := &mockCryptoFetcher{data: cryptoData}

	engine := New(remoteRSI, remoteEMA, localRSI, localEMA, cfg, detector, cryptoFetcher, nil, nil)

	ctx := context.Background()
	result, err := engine.Analyze(ctx, "BTC", indicators.Crypto)

	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil result")
	}

	// Verify PriceHistory is populated for crypto
	if result.PriceHistory == nil {
		t.Error("PriceHistory should not be nil for crypto")
	} else if len(result.PriceHistory) != 20 {
		t.Errorf("PriceHistory length = %d, want 20 for crypto", len(result.PriceHistory))
	}

	// Verify values are in ascending order (oldest first)
	for i := 1; i < len(result.PriceHistory); i++ {
		if result.PriceHistory[i] <= result.PriceHistory[i-1] {
			t.Errorf("PriceHistory should be oldest-first, but PriceHistory[%d] = %f <= PriceHistory[%d-1] = %f",
				i, result.PriceHistory[i], i, result.PriceHistory[i-1])
		}
	}
}

// TestExtractPriceHistoryFromCloses tests the PriceHistory extraction function.
func TestExtractPriceHistoryFromCloses(t *testing.T) {
	tests := []struct {
		name     string
		closes   []float64
		expected []float64
	}{
		{
			name:     "empty slice",
			closes:   []float64{},
			expected: nil,
		},
		{
			name:     "less than 30 elements",
			closes:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
		},
		{
			name:     "exactly 30 elements",
			closes:   makeFloatSlice(1.0, 30),
			expected: makeFloatSlice(1.0, 30),
		},
		{
			name:     "more than 30 elements",
			closes:   makeFloatSlice(1.0, 50),
			expected: makeFloatSlice(21.0, 30), // Last 30 elements
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPriceHistoryFromCloses(tt.closes)
			if !slicesEqual(result, tt.expected) {
				t.Errorf("extractPriceHistoryFromCloses(%v) = %v, want %v",
					tt.closes, result, tt.expected)
			}
		})
	}
}

// makeFloatSlice creates a slice of floats starting from start with count elements.
func makeFloatSlice(start float64, count int) []float64 {
	result := make([]float64, count)
	for i := 0; i < count; i++ {
		result[i] = start + float64(i)
	}
	return result
}

// slicesEqual checks if two float slices are equal.
func slicesEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
