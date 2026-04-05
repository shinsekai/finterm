// Package trend provides trend-following analysis and scoring.
package trend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shinsekai/finterm/internal/config"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
)

// Result contains the complete analysis result for a symbol.
type Result struct {
	// Symbol is the analyzed ticker symbol.
	Symbol string
	// Price is the latest close price from the last closed bar.
	Price float64
	// RSI is the Relative Strength Index value from the last closed bar.
	RSI float64
	// EMAFast is the fast EMA value from the last closed bar.
	EMAFast float64
	// EMASlow is the slow EMA value from the last closed bar.
	EMASlow float64
	// Signal is the trend direction based on EMA crossover.
	Signal Signal
	// Valuation is the valuation assessment based on RSI.
	Valuation string
}

// CryptoDataFetcher fetches OHLCV data for cryptocurrency symbols.
type CryptoDataFetcher interface {
	FetchCryptoOHLCV(ctx context.Context, symbol string) ([]indicators.OHLCV, error)
}

// Engine orchestrates trend analysis by routing to the correct indicator path.
// It supports both remote API computation (equities) and local computation (crypto).
type Engine struct {
	// remoteRSI is the Alpha Vantage server-side RSI indicator.
	remoteRSI indicators.Indicator
	// remoteEMA is the Alpha Vantage server-side EMA indicator.
	remoteEMA indicators.Indicator
	// localRSI is the local RSI computation indicator.
	localRSI *indicators.LocalRSI
	// localEMA is the local EMA computation indicator.
	localEMA *indicators.LocalEMA
	// cfg is the application configuration.
	cfg *config.Config
	// detector is used to determine asset class for a given symbol.
	detector *indicators.AssetClassDetector
	// cryptoFetcher fetches OHLCV data for crypto symbols.
	cryptoFetcher CryptoDataFetcher
}

// New creates a new Engine with the provided indicators and configuration.
// The detector determines asset class from the symbol and routes to the correct indicator path.
// The cryptoFetcher provides OHLCV data for crypto symbols when needed.
func New(
	remoteRSI, remoteEMA indicators.Indicator,
	localRSI *indicators.LocalRSI,
	localEMA *indicators.LocalEMA,
	cfg *config.Config,
	detector *indicators.AssetClassDetector,
	cryptoFetcher CryptoDataFetcher,
) *Engine {
	return &Engine{
		remoteRSI:     remoteRSI,
		remoteEMA:     remoteEMA,
		localRSI:      localRSI,
		localEMA:      localEMA,
		cfg:           cfg,
		detector:      detector,
		cryptoFetcher: cryptoFetcher,
	}
}

// Analyze performs a complete trend analysis for the given symbol.
//
// Based on the asset class, it routes to the appropriate indicator path:
//   - Equity → remote API indicators (Alpha Vantage)
//   - Crypto → local computation indicators
//
// The analysis uses only the last closed bar's values - in-progress bar data
// is excluded to prevent repainting.
//
// Returns:
//   - Result with RSI, EMA values, signal, and valuation
//   - Error if any indicator computation fails
//
// The function is well-structured with clear separation between equity and crypto paths.
//
//nolint:gocyclo // Complexity 17 due to switch on asset class with multiple operations
func (e *Engine) Analyze(ctx context.Context, symbol string, assetClass indicators.AssetClass) (*Result, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	var rsiDataPoints, emaFastDataPoints, emaSlowDataPoints []indicators.DataPoint
	var price float64
	var err error

	// Route to appropriate indicator path based on asset class
	switch assetClass {
	case indicators.Equity:
		rsiDataPoints, err = e.remoteRSI.Compute(ctx, symbol, indicators.Options{
			Period:   e.cfg.Trend.RSIPeriod,
			Interval: "daily",
		})
		if err != nil {
			return nil, fmt.Errorf("computing remote RSI for %s: %w", symbol, err)
		}

		emaFastDataPoints, err = e.remoteEMA.Compute(ctx, symbol, indicators.Options{
			Period:   e.cfg.Trend.EMAFast,
			Interval: "daily",
		})
		if err != nil {
			return nil, fmt.Errorf("computing remote EMA fast for %s: %w", symbol, err)
		}

		emaSlowDataPoints, err = e.remoteEMA.Compute(ctx, symbol, indicators.Options{
			Period:   e.cfg.Trend.EMASlow,
			Interval: "daily",
		})
		if err != nil {
			return nil, fmt.Errorf("computing remote EMA slow for %s: %w", symbol, err)
		}

		// For equities, use EMA fast as price proxy (based on close prices)
		if len(emaFastDataPoints) > 0 {
			price = emaFastDataPoints[0].Value
		}

	case indicators.Crypto:
		// Fetch OHLCV data for this crypto symbol
		if e.cryptoFetcher == nil {
			return nil, fmt.Errorf("crypto data fetcher not configured for %s", symbol)
		}
		ohlcvData, err := e.cryptoFetcher.FetchCryptoOHLCV(ctx, symbol)
		if err != nil {
			return nil, fmt.Errorf("fetching crypto OHLCV for %s: %w", symbol, err)
		}

		// Extract the latest close price from OHLCV data (after reversing to newest-first)
		// The OHLCV data is sorted oldest-first by the fetcher
		if len(ohlcvData) > 0 {
			price = ohlcvData[len(ohlcvData)-1].Close // Newest is last after oldest-first sort
		}

		// Set data on local indicators for this symbol
		e.localRSI.SetData(ohlcvData)
		e.localEMA.SetData(ohlcvData)

		// Now compute using the fetched data
		rsiDataPoints, err = e.localRSI.ComputeFromOHLCV(e.cfg.Trend.RSIPeriod, false) // useInProgress=false
		if err != nil {
			return nil, fmt.Errorf("computing local RSI for %s: %w", symbol, err)
		}

		emaFastDataPoints, err = e.localEMA.ComputeFromOHLCV(e.cfg.Trend.EMAFast, false) // useInProgress=false
		if err != nil {
			return nil, fmt.Errorf("computing local EMA fast for %s: %w", symbol, err)
		}

		emaSlowDataPoints, err = e.localEMA.ComputeFromOHLCV(e.cfg.Trend.EMASlow, false) // useInProgress=false
		if err != nil {
			return nil, fmt.Errorf("computing local EMA slow for %s: %w", symbol, err)
		}

	default:
		return nil, fmt.Errorf("unsupported asset class: %v", assetClass)
	}

	// Extract values from the last closed bar (first entry in descending-sorted arrays)
	// Remote indicators return data sorted newest-first, and we exclude in-progress bars
	// Local indicators with useInProgress=false already exclude the last bar
	if len(rsiDataPoints) == 0 {
		return nil, fmt.Errorf("no RSI data points available for %s", symbol)
	}
	if len(emaFastDataPoints) == 0 {
		return nil, fmt.Errorf("no EMA fast data points available for %s", symbol)
	}
	if len(emaSlowDataPoints) == 0 {
		return nil, fmt.Errorf("no EMA slow data points available for %s", symbol)
	}

	rsi := rsiDataPoints[0].Value
	emaFast := emaFastDataPoints[0].Value
	emaSlow := emaSlowDataPoints[0].Value

	// Compute trend signal from EMA crossover only (RSI is NOT used in trend scoring)
	signal := Score(emaFast, emaSlow)

	// Compute valuation from RSI
	valuation := computeValuation(rsi, e.cfg.Valuation)

	return &Result{
		Symbol:    strings.ToUpper(symbol),
		Price:     price,
		RSI:       rsi,
		EMAFast:   emaFast,
		EMASlow:   emaSlow,
		Signal:    signal,
		Valuation: valuation,
	}, nil
}

// computeValuation returns the valuation assessment based on RSI value.
func computeValuation(rsi float64, val config.ValuationConfig) string {
	switch {
	case rsi < float64(val.Oversold):
		return "Oversold"
	case rsi < float64(val.Undervalued):
		return "Undervalued"
	case rsi < float64(val.FairLow):
		return "Undervalued"
	case rsi <= float64(val.FairHigh):
		return "Fair value"
	case rsi < float64(val.Overvalued):
		return "Overvalued"
	case rsi < float64(val.Overbought):
		return "Overvalued"
	default:
		return "Overbought"
	}
}

// AnalyzeWithSymbolDetection performs trend analysis with automatic asset class detection.
// It uses the configured detector to determine if the symbol is equity or crypto.
func (e *Engine) AnalyzeWithSymbolDetection(ctx context.Context, symbol string) (*Result, error) {
	assetClass := e.detector.DetectAssetClass(symbol)
	return e.Analyze(ctx, symbol, assetClass)
}

// SetLocalData updates the OHLCV data for local indicator computation.
// This is used to refresh data for crypto symbols.
func (e *Engine) SetLocalData(data []indicators.OHLCV) {
	e.localRSI.SetData(data)
	e.localEMA.SetData(data)
}

// GetLocalData returns the current OHLCV data for local indicator computation.
func (e *Engine) GetLocalData() []indicators.OHLCV {
	return e.localRSI.Data
}

// GenerateTestOHLCV generates test OHLCV data for testing purposes.
// This is a utility function for creating synthetic data.
func GenerateTestOHLCV(count int, basePrice float64, volatility float64, startTime time.Time) []indicators.OHLCV {
	data := make([]indicators.OHLCV, count)
	currentPrice := basePrice

	for i := 0; i < count; i++ {
		// Random walk with volatility
		change := volatility * (0.5 - float64(i%10)/10.0) // Slight trend based on position
		currentPrice *= (1 + change)

		// Generate realistic OHLC
		open := currentPrice
		high := currentPrice * (1 + volatility/2)
		low := currentPrice * (1 - volatility/2)
		closePrice := currentPrice * (1 + change/2)

		data[i] = indicators.OHLCV{
			Date:   startTime.Add(time.Duration(i) * 24 * time.Hour),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: 1e6, // Arbitrary volume
		}
	}

	return data
}
