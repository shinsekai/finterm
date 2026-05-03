// Package indicators provides the technical indicator interface for trend analysis.
package indicators

import (
	"context"
	"strings"
	"sync"
	"time"
)

// DataPoint represents a single data point with a date and value.
type DataPoint struct {
	Date  time.Time
	Value float64
}

// OHLCV represents Open, High, Low, Close, Volume data for a single time period.
type OHLCV struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// Indicator defines the contract for technical indicators.
// Implementations can fetch data from remote APIs or compute locally from OHLCV data.
type Indicator interface {
	// Compute returns indicator data points for the given symbol.
	// The context can be used for cancellation and timeout.
	Compute(ctx context.Context, symbol string, opts Options) ([]DataPoint, error)
}

// Options contains configuration for indicator computation.
type Options struct {
	// Period is the lookback period for the indicator (e.g., 14 for RSI).
	Period int
	// Interval is the time interval (e.g., "daily", "1min", "5min").
	Interval string
	// SeriesType is the price series to use (e.g., "close", "open", "high", "low").
	SeriesType string
}

// AssetClass represents the type of asset being analyzed.
type AssetClass int

const (
	// Equity represents traditional stocks and ETFs.
	Equity AssetClass = iota
	// Crypto represents cryptocurrencies.
	Crypto
	// Commodity represents commodities (e.g., WTI, BRENT, WHEAT).
	Commodity
)

// String returns the string representation of the AssetClass.
func (a AssetClass) String() string {
	switch a {
	case Equity:
		return "Equity"
	case Crypto:
		return "Crypto"
	case Commodity:
		return "Commodity"
	default:
		return "Unknown"
	}
}

// DefaultCommoditySymbols is the list of supported commodity symbols.
// These match the CommodityFunction constants in internal/alphavantage/commodities.go.
var DefaultCommoditySymbols = []string{
	"WTI", "BRENT", "NATURAL_GAS", "COPPER", "ALUMINUM",
	"WHEAT", "CORN", "COTTON", "SUGAR", "COFFEE",
}

// AssetClassDetector provides configurable asset class detection.
// The crypto and commodity lists are thread-safe and can be updated at runtime.
type AssetClassDetector struct {
	mu               sync.RWMutex
	cryptoSymbols    map[string]bool
	commoditySymbols map[string]bool
}

// NewAssetClassDetector creates a new detector with the provided lists of crypto and commodity symbols.
// The symbols are stored in uppercase for case-insensitive matching.
// If commoditySymbols is nil, DefaultCommoditySymbols is used.
func NewAssetClassDetector(cryptoSymbols, commoditySymbols []string) *AssetClassDetector {
	d := &AssetClassDetector{
		cryptoSymbols:    make(map[string]bool),
		commoditySymbols: make(map[string]bool),
	}
	for _, symbol := range cryptoSymbols {
		d.cryptoSymbols[strings.ToUpper(symbol)] = true
	}
	if commoditySymbols == nil {
		commoditySymbols = DefaultCommoditySymbols
	}
	for _, symbol := range commoditySymbols {
		d.commoditySymbols[strings.ToUpper(symbol)] = true
	}
	return d
}

// DetectAssetClass returns the asset class for the given symbol.
// Returns Commodity if the symbol matches any known commodity symbol,
// Crypto if the symbol matches any known crypto symbol,
// otherwise Equity (the default fallback).
func (d *AssetClassDetector) DetectAssetClass(symbol string) AssetClass {
	d.mu.RLock()
	defer d.mu.RUnlock()
	upper := strings.ToUpper(symbol)
	if d.commoditySymbols[upper] {
		return Commodity
	}
	if d.cryptoSymbols[upper] {
		return Crypto
	}
	return Equity
}

// SetCryptoSymbols updates the list of known crypto symbols.
// This is thread-safe and can be called at runtime.
func (d *AssetClassDetector) SetCryptoSymbols(cryptoSymbols []string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cryptoSymbols = make(map[string]bool)
	for _, symbol := range cryptoSymbols {
		d.cryptoSymbols[strings.ToUpper(symbol)] = true
	}
}

// AddCryptoSymbol adds a single crypto symbol to the detector.
// Thread-safe operation.
func (d *AssetClassDetector) AddCryptoSymbol(symbol string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cryptoSymbols[strings.ToUpper(symbol)] = true
}

// RemoveCryptoSymbol removes a crypto symbol from the detector.
// Thread-safe operation.
func (d *AssetClassDetector) RemoveCryptoSymbol(symbol string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.cryptoSymbols, strings.ToUpper(symbol))
}

// SetCommoditySymbols updates the list of known commodity symbols.
// This is thread-safe and can be called at runtime.
func (d *AssetClassDetector) SetCommoditySymbols(commoditySymbols []string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.commoditySymbols = make(map[string]bool)
	for _, symbol := range commoditySymbols {
		d.commoditySymbols[strings.ToUpper(symbol)] = true
	}
}

// AddCommoditySymbol adds a single commodity symbol to the detector.
// Thread-safe operation.
func (d *AssetClassDetector) AddCommoditySymbol(symbol string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.commoditySymbols[strings.ToUpper(symbol)] = true
}

// RemoveCommoditySymbol removes a commodity symbol from the detector.
// Thread-safe operation.
func (d *AssetClassDetector) RemoveCommoditySymbol(symbol string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.commoditySymbols, strings.ToUpper(symbol))
}
