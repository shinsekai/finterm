package indicators

import (
	"strings"
	"testing"
	"time"
)

func TestDetectAssetClass_Equity(t *testing.T) {
	cryptoSymbols := []string{"BTC", "ETH", "SOL"}
	detector := NewAssetClassDetector(cryptoSymbols, nil)

	tests := []struct {
		name   string
		symbol string
		want   AssetClass
	}{
		{
			name:   "AAPL classified as Equity",
			symbol: "AAPL",
			want:   Equity,
		},
		{
			name:   "MSFT classified as Equity",
			symbol: "MSFT",
			want:   Equity,
		},
		{
			name:   "GOOGL classified as Equity",
			symbol: "GOOGL",
			want:   Equity,
		},
		{
			name:   "AMZN classified as Equity",
			symbol: "AMZN",
			want:   Equity,
		},
		{
			name:   "NVDA classified as Equity",
			symbol: "NVDA",
			want:   Equity,
		},
		{
			name:   "TSLA classified as Equity",
			symbol: "TSLA",
			want:   Equity,
		},
		{
			name:   "lowercase equity symbol",
			symbol: "aapl",
			want:   Equity,
		},
		{
			name:   "mixed case equity symbol",
			symbol: "AaPl",
			want:   Equity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.DetectAssetClass(tt.symbol); got != tt.want {
				t.Errorf("DetectAssetClass(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestDetectAssetClass_Crypto(t *testing.T) {
	cryptoSymbols := []string{"BTC", "ETH", "SOL", "ADA", "DOT"}
	detector := NewAssetClassDetector(cryptoSymbols, nil)

	tests := []struct {
		name   string
		symbol string
		want   AssetClass
	}{
		{
			name:   "BTC classified as Crypto",
			symbol: "BTC",
			want:   Crypto,
		},
		{
			name:   "ETH classified as Crypto",
			symbol: "ETH",
			want:   Crypto,
		},
		{
			name:   "SOL classified as Crypto",
			symbol: "SOL",
			want:   Crypto,
		},
		{
			name:   "ADA classified as Crypto",
			symbol: "ADA",
			want:   Crypto,
		},
		{
			name:   "DOT classified as Crypto",
			symbol: "DOT",
			want:   Crypto,
		},
		{
			name:   "lowercase crypto symbol",
			symbol: "btc",
			want:   Crypto,
		},
		{
			name:   "mixed case crypto symbol",
			symbol: "BtC",
			want:   Crypto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.DetectAssetClass(tt.symbol); got != tt.want {
				t.Errorf("DetectAssetClass(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestDetectAssetClass_Unknown_DefaultsEquity(t *testing.T) {
	cryptoSymbols := []string{"BTC", "ETH"}
	detector := NewAssetClassDetector(cryptoSymbols, nil)

	tests := []struct {
		name   string
		symbol string
		want   AssetClass
	}{
		{
			name:   "Unknown symbol defaults to Equity",
			symbol: "UNKNOWN",
			want:   Equity,
		},
		{
			name:   "Random symbol defaults to Equity",
			symbol: "XYZ123",
			want:   Equity,
		},
		{
			name:   "Empty string defaults to Equity",
			symbol: "",
			want:   Equity,
		},
		{
			name:   "Partial crypto match not crypto (SOL vs SOLA)",
			symbol: "SOLA",
			want:   Equity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.DetectAssetClass(tt.symbol); got != tt.want {
				t.Errorf("DetectAssetClass(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestAssetClassDetector_SetCryptoSymbols(t *testing.T) {
	detector := NewAssetClassDetector([]string{"BTC", "ETH"}, nil)

	// Initial state
	if got := detector.DetectAssetClass("BTC"); got != Crypto {
		t.Errorf("Initial: DetectAssetClass(BTC) = %v, want Crypto", got)
	}
	if got := detector.DetectAssetClass("SOL"); got != Equity {
		t.Errorf("Initial: DetectAssetClass(SOL) = %v, want Equity", got)
	}

	// Update crypto symbols
	detector.SetCryptoSymbols([]string{"SOL", "ADA", "DOT"})

	// Verify new state
	if got := detector.DetectAssetClass("BTC"); got != Equity {
		t.Errorf("After update: DetectAssetClass(BTC) = %v, want Equity", got)
	}
	if got := detector.DetectAssetClass("SOL"); got != Crypto {
		t.Errorf("After update: DetectAssetClass(SOL) = %v, want Crypto", got)
	}
	if got := detector.DetectAssetClass("ADA"); got != Crypto {
		t.Errorf("After update: DetectAssetClass(ADA) = %v, want Crypto", got)
	}
}

func TestAssetClassDetector_AddCryptoSymbol(t *testing.T) {
	detector := NewAssetClassDetector([]string{"BTC"}, nil)

	// Initial state
	if got := detector.DetectAssetClass("ETH"); got != Equity {
		t.Errorf("Initial: DetectAssetClass(ETH) = %v, want Equity", got)
	}

	// Add ETH
	detector.AddCryptoSymbol("ETH")

	// Verify ETH is now detected as Crypto
	if got := detector.DetectAssetClass("ETH"); got != Crypto {
		t.Errorf("After add: DetectAssetClass(ETH) = %v, want Crypto", got)
	}
}

func TestAssetClassDetector_RemoveCryptoSymbol(t *testing.T) {
	detector := NewAssetClassDetector([]string{"BTC", "ETH", "SOL"}, nil)

	// Initial state
	if got := detector.DetectAssetClass("ETH"); got != Crypto {
		t.Errorf("Initial: DetectAssetClass(ETH) = %v, want Crypto", got)
	}

	// Remove ETH
	detector.RemoveCryptoSymbol("ETH")

	// Verify ETH is now detected as Equity
	if got := detector.DetectAssetClass("ETH"); got != Equity {
		t.Errorf("After remove: DetectAssetClass(ETH) = %v, want Equity", got)
	}
}

func TestAssetClassString(t *testing.T) {
	tests := []struct {
		name string
		a    AssetClass
		want string
	}{
		{
			name: "Equity string representation",
			a:    Equity,
			want: "Equity",
		},
		{
			name: "Crypto string representation",
			a:    Crypto,
			want: "Crypto",
		},
		{
			name: "Commodity string representation",
			a:    Commodity,
			want: "Commodity",
		},
		{
			name: "Unknown asset class",
			a:    AssetClass(99),
			want: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.String(); got != tt.want {
				t.Errorf("AssetClass.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDataPoint(t *testing.T) {
	// Test that DataPoint struct is properly usable
	dp := DataPoint{
		Date:  time.Time{},
		Value: 42.5,
	}
	if dp.Value != 42.5 {
		t.Errorf("DataPoint.Value = %f, want 42.5", dp.Value)
	}
}

func TestOHLCV(t *testing.T) {
	// Test that OHLCV struct is properly usable
	ohlcv := OHLCV{
		Date:   time.Time{},
		Open:   40.0,
		High:   45.0,
		Low:    38.0,
		Close:  42.5,
		Volume: 1000000,
	}
	if ohlcv.Close != 42.5 {
		t.Errorf("OHLCV.Close = %f, want 42.5", ohlcv.Close)
	}
	if ohlcv.High != 45.0 {
		t.Errorf("OHLCV.High = %f, want 45.0", ohlcv.High)
	}
	if ohlcv.Low != 38.0 {
		t.Errorf("OHLCV.Low = %f, want 38.0", ohlcv.Low)
	}
	if ohlcv.Volume != 1000000 {
		t.Errorf("OHLCV.Volume = %f, want 1000000", ohlcv.Volume)
	}
}

func TestOptions(t *testing.T) {
	// Test that Options struct is properly usable
	opts := Options{
		Period:     14,
		Interval:   "daily",
		SeriesType: "close",
	}
	if opts.Period != 14 {
		t.Errorf("Options.Period = %d, want 14", opts.Period)
	}
	if opts.Interval != "daily" {
		t.Errorf("Options.Interval = %q, want \"daily\"", opts.Interval)
	}
	if opts.SeriesType != "close" {
		t.Errorf("Options.SeriesType = %q, want \"close\"", opts.SeriesType)
	}
}

func TestAssetClassDetector_CaseInsensitive(t *testing.T) {
	cryptoSymbols := []string{"btc", "eth", "sol"} // Input in lowercase
	detector := NewAssetClassDetector(cryptoSymbols, nil)

	tests := []string{
		"BTC",
		"btc",
		"Btc",
		"BTc",
		"bTc",
		"ETH",
		"eth",
		"Eth",
		"SOL",
		"sol",
		"Sol",
	}

	for _, symbol := range tests {
		t.Run(symbol, func(t *testing.T) {
			if got := detector.DetectAssetClass(symbol); got != Crypto {
				t.Errorf("DetectAssetClass(%q) = %v, want Crypto (case insensitive)", symbol, got)
			}
		})
	}
}

func TestAssetClassDetector_Concurrent(t *testing.T) {
	cryptoSymbols := []string{"BTC", "ETH", "SOL"}
	detector := NewAssetClassDetector(cryptoSymbols, nil)

	// Run concurrent reads and writes
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			detector.DetectAssetClass("BTC")
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 100; i++ {
			detector.DetectAssetClass("AAPL")
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 10; i++ {
			detector.AddCryptoSymbol(strings.Repeat("X", i%10+1))
		}
		done <- true
	}()

	<-done
	<-done
	<-done

	// Verify final state
	if got := detector.DetectAssetClass("BTC"); got != Crypto {
		t.Errorf("After concurrent ops: DetectAssetClass(BTC) = %v, want Crypto", got)
	}
}

func TestDetectAssetClass_CommodityWTI(t *testing.T) {
	detector := NewAssetClassDetector([]string{}, DefaultCommoditySymbols)

	tests := []struct {
		name   string
		symbol string
		want   AssetClass
	}{
		{
			name:   "WTI classified as Commodity",
			symbol: "WTI",
			want:   Commodity,
		},
		{
			name:   "lowercase wti classified as Commodity",
			symbol: "wti",
			want:   Commodity,
		},
		{
			name:   "mixed case Wti classified as Commodity",
			symbol: "Wti",
			want:   Commodity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.DetectAssetClass(tt.symbol); got != tt.want {
				t.Errorf("DetectAssetClass(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestDetectAssetClass_CommodityBrent(t *testing.T) {
	detector := NewAssetClassDetector([]string{}, DefaultCommoditySymbols)

	tests := []struct {
		name   string
		symbol string
		want   AssetClass
	}{
		{
			name:   "BRENT classified as Commodity",
			symbol: "BRENT",
			want:   Commodity,
		},
		{
			name:   "lowercase brent classified as Commodity",
			symbol: "brent",
			want:   Commodity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.DetectAssetClass(tt.symbol); got != tt.want {
				t.Errorf("DetectAssetClass(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestDetectAssetClass_CommodityAllDefaults(t *testing.T) {
	detector := NewAssetClassDetector([]string{}, DefaultCommoditySymbols)

	tests := []struct {
		name   string
		symbol string
		want   AssetClass
	}{
		{"WTI", "WTI", Commodity},
		{"BRENT", "BRENT", Commodity},
		{"NATURAL_GAS", "NATURAL_GAS", Commodity},
		{"COPPER", "COPPER", Commodity},
		{"ALUMINUM", "ALUMINUM", Commodity},
		{"WHEAT", "WHEAT", Commodity},
		{"CORN", "CORN", Commodity},
		{"COTTON", "COTTON", Commodity},
		{"SUGAR", "SUGAR", Commodity},
		{"COFFEE", "COFFEE", Commodity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.DetectAssetClass(tt.symbol); got != tt.want {
				t.Errorf("DetectAssetClass(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestDetectAssetClass_CryptoStillWorks(t *testing.T) {
	detector := NewAssetClassDetector([]string{"BTC", "ETH"}, DefaultCommoditySymbols)

	tests := []struct {
		name   string
		symbol string
		want   AssetClass
	}{
		{
			name:   "BTC classified as Crypto",
			symbol: "BTC",
			want:   Crypto,
		},
		{
			name:   "ETH classified as Crypto",
			symbol: "ETH",
			want:   Crypto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.DetectAssetClass(tt.symbol); got != tt.want {
				t.Errorf("DetectAssetClass(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestDetectAssetClass_EquityFallback(t *testing.T) {
	detector := NewAssetClassDetector([]string{}, DefaultCommoditySymbols)

	tests := []struct {
		name   string
		symbol string
		want   AssetClass
	}{
		{
			name:   "AAPL defaults to Equity",
			symbol: "AAPL",
			want:   Equity,
		},
		{
			name:   "UNKNOWN defaults to Equity",
			symbol: "UNKNOWN",
			want:   Equity,
		},
		{
			name:   "Random ticker defaults to Equity",
			symbol: "XYZ123",
			want:   Equity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.DetectAssetClass(tt.symbol); got != tt.want {
				t.Errorf("DetectAssetClass(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestDetectAssetClass_CommodityPriority(t *testing.T) {
	// Test that commodity is checked before crypto
	// This matters if there's a symbol overlap
	detector := NewAssetClassDetector([]string{"CORN"}, DefaultCommoditySymbols)

	// CORN should be detected as Commodity (from DefaultCommoditySymbols)
	// even though it's also in the crypto list
	if got := detector.DetectAssetClass("CORN"); got != Commodity {
		t.Errorf("DetectAssetClass(CORN) = %v, want Commodity (commodity checked before crypto)", got)
	}
}

func TestDefaultCommoditySymbols_MatchesAlphaVantageConstants(t *testing.T) {
	// Verify DefaultCommoditySymbols contains the expected commodity symbols
	expected := []string{
		"WTI", "BRENT", "NATURAL_GAS", "COPPER", "ALUMINUM",
		"WHEAT", "CORN", "COTTON", "SUGAR", "COFFEE",
	}

	if len(DefaultCommoditySymbols) != len(expected) {
		t.Errorf("DefaultCommoditySymbols length = %d, want %d", len(DefaultCommoditySymbols), len(expected))
	}

	for _, want := range expected {
		found := false
		for _, got := range DefaultCommoditySymbols {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultCommoditySymbols missing expected symbol %q", want)
		}
	}
}

func TestAssetClassDetector_AddCommoditySymbol(t *testing.T) {
	detector := NewAssetClassDetector([]string{}, []string{"WTI"})

	// Initial state
	if got := detector.DetectAssetClass("GOLD"); got != Equity {
		t.Errorf("Initial: DetectAssetClass(GOLD) = %v, want Equity", got)
	}

	// Add GOLD as commodity
	detector.AddCommoditySymbol("GOLD")

	// Verify GOLD is now detected as Commodity
	if got := detector.DetectAssetClass("GOLD"); got != Commodity {
		t.Errorf("After add: DetectAssetClass(GOLD) = %v, want Commodity", got)
	}
}

func TestAssetClassDetector_RemoveCommoditySymbol(t *testing.T) {
	detector := NewAssetClassDetector([]string{}, []string{"WTI", "BRENT", "WHEAT"})

	// Initial state
	if got := detector.DetectAssetClass("BRENT"); got != Commodity {
		t.Errorf("Initial: DetectAssetClass(BRENT) = %v, want Commodity", got)
	}

	// Remove BRENT
	detector.RemoveCommoditySymbol("BRENT")

	// Verify BRENT is now detected as Equity
	if got := detector.DetectAssetClass("BRENT"); got != Equity {
		t.Errorf("After remove: DetectAssetClass(BRENT) = %v, want Equity", got)
	}
}

func TestAssetClassDetector_SetCommoditySymbols(t *testing.T) {
	detector := NewAssetClassDetector([]string{}, DefaultCommoditySymbols)

	// Initial state
	if got := detector.DetectAssetClass("WTI"); got != Commodity {
		t.Errorf("Initial: DetectAssetClass(WTI) = %v, want Commodity", got)
	}

	// Update to custom commodity list
	detector.SetCommoditySymbols([]string{"GOLD", "SILVER"})

	// Verify WTI is no longer detected as Commodity
	if got := detector.DetectAssetClass("WTI"); got != Equity {
		t.Errorf("After update: DetectAssetClass(WTI) = %v, want Equity", got)
	}

	// Verify new commodities are detected
	if got := detector.DetectAssetClass("GOLD"); got != Commodity {
		t.Errorf("After update: DetectAssetClass(GOLD) = %v, want Commodity", got)
	}
	if got := detector.DetectAssetClass("SILVER"); got != Commodity {
		t.Errorf("After update: DetectAssetClass(SILVER) = %v, want Commodity", got)
	}
}
