// Package trend provides trend-following analysis and scoring.
package trend

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/cache"
	"github.com/shinsekai/finterm/internal/config"
	"github.com/shinsekai/finterm/internal/domain/blitz"
	"github.com/shinsekai/finterm/internal/domain/destiny"
	"github.com/shinsekai/finterm/internal/domain/flow"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
	"github.com/shinsekai/finterm/internal/domain/vortex"
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
	// Kept for internal computation, just not displayed.
	EMAFast float64
	// EMASlow is the slow EMA value from the last closed bar.
	// Kept for internal computation, just not displayed.
	EMASlow float64
	// Signal is the trend direction based on EMA crossover.
	// Renamed to FTEMA in view.
	Signal Signal
	// Valuation is the valuation assessment based on RSI.
	Valuation string
	// BlitzScore is the BLITZ trending system signal: +1 (Long), -1 (Short), or 0 (Hold).
	BlitzScore int
	// BlitzTSI is the latest TSI (Pearson correlation) value from BLITZ.
	BlitzTSI float64
	// BlitzRSISmooth is the latest smoothed RSI value from BLITZ.
	BlitzRSISmooth float64
	// DESTINY trend following system results.
	DestinyScore     int     // +1 (long), -1 (short), 0 (hold)
	DestinyTPI       float64 // Trend Probability Indicator value
	DestinyRSISmooth float64 // Smoothed RSI from DESTINY
	// FLOW trend following system results.
	FlowScore     int     // +1 (long), -1 (short), 0 (hold)
	FlowSebastine float64 // Latest Sebastine value
	FlowRSISmooth float64 // Smoothed RSI from FLOW
	// VORTEX trend following system results.
	VortexScore     int     // +1 (long), -1 (short), 0 (hold)
	VortexTPI       float64 // Trend Probability Indicator from VORTEX
	VortexRSISmooth float64 // Smoothed RSI from VORTEX
	VortexWave      float64 // Latest wave-weighted regression value
	VortexMid       float64 // Latest Mid band value
	// TPI composite signal.
	// TPI is the average of EMA signal, BLITZ, DESTINY, FLOW, and VORTEX (-1 to +1).
	TPI float64
	// TPISignal is the TPI signal label: "LONG" or "CASH".
	TPISignal string
}

// CryptoDataFetcher fetches OHLCV data for cryptocurrency symbols.
type CryptoDataFetcher interface {
	FetchCryptoOHLCV(ctx context.Context, symbol string) ([]indicators.OHLCV, error)
}

// TimeSeriesClient fetches daily time series data for equities.
// The concrete implementation is *alphavantage.Client from the alphavantage package.
type TimeSeriesClient interface {
	GetDailyTimeSeries(ctx context.Context, symbol, outputsize string) (*alphavantage.TimeSeriesDaily, error)
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
	// timeSeriesClient fetches daily time series for equities (for BLITZ computation).
	timeSeriesClient TimeSeriesClient
	// cache is used to cache time series data for BLITZ computation.
	cache cache.Cache
}

// New creates a new Engine with the provided indicators and configuration.
// The detector determines asset class from the symbol and routes to the correct indicator path.
// The cryptoFetcher provides OHLCV data for crypto symbols when needed.
// The timeSeriesClient provides daily time series for equities (BLITZ computation).
// The cache is used to cache time series data.
func New(
	remoteRSI, remoteEMA indicators.Indicator,
	localRSI *indicators.LocalRSI,
	localEMA *indicators.LocalEMA,
	cfg *config.Config,
	detector *indicators.AssetClassDetector,
	cryptoFetcher CryptoDataFetcher,
	timeSeriesClient TimeSeriesClient,
	cacheStore cache.Cache,
) *Engine {
	return &Engine{
		remoteRSI:        remoteRSI,
		remoteEMA:        remoteEMA,
		localRSI:         localRSI,
		localEMA:         localEMA,
		cfg:              cfg,
		detector:         detector,
		cryptoFetcher:    cryptoFetcher,
		timeSeriesClient: timeSeriesClient,
		cache:            cacheStore,
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
//   - Result with RSI, EMA values, signal, valuation, and BLITZ score
//   - Error if any indicator computation fails (BLITZ failures are logged but don't fail the analysis)
//
// The function is well-structured with clear separation between equity and crypto paths.
//
//nolint:gocyclo // Complexity 20 due to switch on asset class with BLITZ integration
func (e *Engine) Analyze(ctx context.Context, symbol string, assetClass indicators.AssetClass) (*Result, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	var rsiDataPoints, emaFastDataPoints, emaSlowDataPoints []indicators.DataPoint
	var price float64
	var err error
	var blitzScore int
	var blitzTSI, blitzRSISmooth float64
	var destinyScore int
	var destinyTPI, destinyRSISmooth float64
	var flowScore int
	var flowSebastine, flowRSISmooth float64
	var vortexScore int
	var vortexTPI, vortexRSISmooth, vortexWave, vortexMid float64

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

		// Compute BLITZ score for equities using daily time series
		if e.timeSeriesClient != nil {
			cacheKey := cache.Key("timeseries", "daily", symbol)
			var tsData *alphavantage.TimeSeriesDaily

			// Try to get from cache first
			if cached, ok := e.cache.Get(cacheKey); ok {
				if ts, ok := cached.(*alphavantage.TimeSeriesDaily); ok {
					tsData = ts
				}
			}

			// If not in cache, fetch from API
			if tsData == nil {
				tsData, err = e.timeSeriesClient.GetDailyTimeSeries(ctx, symbol, "compact")
				if err != nil {
					// BLITZ computation failed - log and continue with default values
					fmt.Printf("warning: failed to fetch time series for BLITZ computation for %s: %v\n", symbol, err)
					blitzScore, blitzTSI, blitzRSISmooth = 0, 0, 0
				} else {
					// Cache the result
					e.cache.Set(cacheKey, tsData, e.cfg.Cache.DailyTTL)
				}
			}

			// Extract OHLC data and compute BLITZ, DESTINY, and FLOW
			if tsData != nil {
				opens, highs, lows, closePrices, err := extractOHLCFromTimeSeries(tsData)
				if err != nil {
					fmt.Printf("warning: failed to extract OHLC data for BLITZ/DESTINY/FLOW/VORTEX computation for %s: %v\n", symbol, err)
					blitzScore, blitzTSI, blitzRSISmooth = 0, 0, 0
					destinyScore, destinyTPI, destinyRSISmooth = 0, 0, 0
					flowScore, flowSebastine, flowRSISmooth = 0, 0, 0
					vortexScore, vortexTPI, vortexRSISmooth, vortexWave, vortexMid = 0, 0, 0, 0, 0
				} else {
					blitzScore, blitzTSI, blitzRSISmooth = computeBlitz(closePrices)
					// Compute DESTINY after BLITZ
					destinyScore, destinyTPI, destinyRSISmooth = computeDestiny(closePrices)
					// Compute FLOW using full OHLC data
					flowScore, flowSebastine, flowRSISmooth = computeFlow(opens, highs, lows, closePrices)
					// Compute VORTEX using close prices
					vortexScore, vortexTPI, vortexRSISmooth, vortexWave, vortexMid = computeVortex(closePrices)
				}
			}
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

		// Compute BLITZ, DESTINY, and FLOW score for crypto using existing OHLCV data
		closes := make([]float64, len(ohlcvData))
		opens := make([]float64, len(ohlcvData))
		highs := make([]float64, len(ohlcvData))
		lows := make([]float64, len(ohlcvData))
		for i, ohlcv := range ohlcvData {
			closes[i] = ohlcv.Close
			opens[i] = ohlcv.Open
			highs[i] = ohlcv.High
			lows[i] = ohlcv.Low
		}
		blitzScore, blitzTSI, blitzRSISmooth = computeBlitz(closes)
		// Compute DESTINY score for crypto using existing OHLCV data
		destinyScore, destinyTPI, destinyRSISmooth = computeDestiny(closes)
		// Compute FLOW score for crypto using full OHLCV data
		flowScore, flowSebastine, flowRSISmooth = computeFlow(opens, highs, lows, closes)
		// Compute VORTEX score for crypto using close prices
		vortexScore, vortexTPI, vortexRSISmooth, vortexWave, vortexMid = computeVortex(closes)

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

	// Compute TPI composite signal
	tpi := TPI(signal, blitzScore, destinyScore, flowScore, vortexScore)
	tpiSignal := TPISignal(tpi)

	return &Result{
		Symbol:           strings.ToUpper(symbol),
		Price:            price,
		RSI:              rsi,
		EMAFast:          emaFast,
		EMASlow:          emaSlow,
		Signal:           signal,
		Valuation:        valuation,
		BlitzScore:       blitzScore,
		BlitzTSI:         blitzTSI,
		BlitzRSISmooth:   blitzRSISmooth,
		DestinyScore:     destinyScore,
		DestinyTPI:       destinyTPI,
		DestinyRSISmooth: destinyRSISmooth,
		FlowScore:        flowScore,
		FlowSebastine:    flowSebastine,
		FlowRSISmooth:    flowRSISmooth,
		VortexScore:      vortexScore,
		VortexTPI:        vortexTPI,
		VortexRSISmooth:  vortexRSISmooth,
		VortexWave:       vortexWave,
		VortexMid:        vortexMid,
		TPI:              tpi,
		TPISignal:        tpiSignal,
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

// extractClosePricesFromTimeSeries extracts close prices from a TimeSeriesDaily response.
// The response contains dates as string keys with values sorted newest-first.
// This function returns closes sorted oldest-first for BLITZ computation.
// nolint:unused // Kept for reference, used in future BLITZ implementation
func extractClosePricesFromTimeSeries(ts *alphavantage.TimeSeriesDaily) ([]float64, error) {
	if ts == nil || len(ts.TimeSeries) == 0 {
		return nil, fmt.Errorf("empty time series data")
	}

	// The time series map has date string keys, sorted newest-first
	// We need to sort them oldest-first and extract close prices
	dates := make([]string, 0, len(ts.TimeSeries))
	for date := range ts.TimeSeries {
		dates = append(dates, date)
	}

	// Sort dates (they're in YYYY-MM-DD format, so string comparison works)
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[i] > dates[j] {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}

	// Extract close prices in sorted order
	closes := make([]float64, 0, len(dates))
	for _, date := range dates {
		entry := ts.TimeSeries[date]
		closePrice, err := strconv.ParseFloat(entry.Close, 64)
		if err != nil {
			continue
		}
		closes = append(closes, closePrice)
	}

	if len(closes) == 0 {
		return nil, fmt.Errorf("no valid close prices found in time series")
	}

	return closes, nil
}

// extractOHLCFromTimeSeries extracts OHLC data from a TimeSeriesDaily response.
// The response contains dates as string keys with values sorted newest-first.
// This function returns OHLC arrays sorted oldest-first for FLOW computation.
func extractOHLCFromTimeSeries(ts *alphavantage.TimeSeriesDaily) (opens, highs, lows, closes []float64, err error) {
	if ts == nil || len(ts.TimeSeries) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("empty time series data")
	}

	// The time series map has date string keys, sorted newest-first
	// We need to sort them oldest-first and extract OHLC data
	dates := make([]string, 0, len(ts.TimeSeries))
	for date := range ts.TimeSeries {
		dates = append(dates, date)
	}

	// Sort dates (they're in YYYY-MM-DD format, so string comparison works)
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[i] > dates[j] {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}

	// Extract OHLC data in sorted order
	opens = make([]float64, 0, len(dates))
	highs = make([]float64, 0, len(dates))
	lows = make([]float64, 0, len(dates))
	closes = make([]float64, 0, len(dates))

	for _, date := range dates {
		entry := ts.TimeSeries[date]
		open, err := strconv.ParseFloat(entry.Open, 64)
		if err == nil {
			opens = append(opens, open)
		}
		high, err := strconv.ParseFloat(entry.High, 64)
		if err == nil {
			highs = append(highs, high)
		}
		low, err := strconv.ParseFloat(entry.Low, 64)
		if err == nil {
			lows = append(lows, low)
		}
		closePrice, err := strconv.ParseFloat(entry.Close, 64)
		if err == nil {
			closes = append(closes, closePrice)
		}
	}

	if len(opens) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("no valid OHLC data found in time series")
	}

	return opens, highs, lows, closes, nil
}

// computeFlow computes FLOW signal from OHLC data.
// Returns score=0 if computation fails (e.g., insufficient data).
func computeFlow(opens, highs, lows, closes []float64) (score int, sebastine, rsiSmooth float64) {
	// Convert closes to flow.OHLCV format
	n := len(closes)
	if n == 0 {
		return 0, 0, 0
	}

	flowOHLCV := make([]flow.OHLCV, n)
	for i := 0; i < n; i++ {
		flowOHLCV[i] = flow.OHLCV{
			Date:   time.Time{}, // Not used by FLOW
			Open:   opens[i],
			High:   highs[i],
			Low:    lows[i],
			Close:  closes[i],
			Volume: 0,
		}
	}

	result, err := flow.Compute(flowOHLCV, flow.DefaultConfig())
	if err != nil {
		// FLOW computation failed - return default values
		return 0, 0, 0
	}

	return result.Score, result.Sebastine, result.RSISmooth
}

// computeBlitz computes BLITZ signal from close prices.
// Returns score=0 if computation fails (e.g., insufficient data).
func computeBlitz(closes []float64) (score int, tsi, rsiSmooth float64) {
	blitzResult, err := blitz.ComputeSingle(closes)
	if err != nil {
		// BLITZ computation failed - return default values
		return 0, 0, 0
	}
	return int(blitzResult.Current), blitzResult.TSI, blitzResult.RSISmooth
}

// computeDestiny computes DESTINY signal from close prices.
// Returns score=0 if computation fails (e.g., insufficient data).
func computeDestiny(closes []float64) (score int, tpi, rsiSmooth float64) {
	// Convert closes to destiny.OHLCV format
	ohlcvData := make([]destiny.OHLCV, len(closes))
	for i, close := range closes {
		ohlcvData[i] = destiny.OHLCV{
			Date:   float64(i),
			Open:   close,
			High:   close,
			Low:    close,
			Close:  close,
			Volume: 0,
		}
	}

	destinyResult, err := destiny.Compute(ohlcvData, destiny.DefaultConfig())
	if err != nil {
		// DESTINY computation failed - return default values
		return 0, 0, 0
	}
	return destinyResult.Score, destinyResult.TPI, destinyResult.RSISmooth
}

// computeVortex computes the VORTEX signal from close prices.
// Returns score = 0 if computation fails.
func computeVortex(closes []float64) (score int, tpi, rsiSmooth, wave, mid float64) {
	ohlcv := make([]vortex.OHLCV, len(closes))
	for i, c := range closes {
		ohlcv[i] = vortex.OHLCV{Close: c}
	}
	result, err := vortex.Compute(ohlcv, vortex.DefaultConfig())
	if err != nil {
		return 0, 0, 0, 0, 0
	}
	return result.Score, result.TPI, result.RSISmooth, result.Wave, result.Mid
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
