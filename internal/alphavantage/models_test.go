package alphavantage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// loadFixture loads a JSON fixture file from the testdata directory.
func loadFixture(t *testing.T, filename string) []byte {
	t.Helper()

	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", filename, err)
	}
	return data
}

func TestParseGlobalQuote(t *testing.T) {
	data := loadFixture(t, "global_quote_aapl.json")

	var resp GlobalQuoteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal GlobalQuoteResponse: %v", err)
	}

	if resp.GlobalQuote == nil {
		t.Fatal("GlobalQuote is nil")
	}

	gq := resp.GlobalQuote

	tests := []struct {
		name      string
		got       string
		want      string
		wantFloat float64
	}{
		{"symbol", gq.Symbol, "AAPL", 0},
		{"open", gq.Open, "169.5000", 169.5},
		{"high", gq.High, "171.2000", 171.2},
		{"low", gq.Low, "168.8000", 168.8},
		{"price", gq.Price, "170.7500", 170.75},
		{"volume", gq.Volume, "52134200", 52134200},
		{"last trading day", gq.LastTradingDay, "2024-04-02", 0},
		{"previous close", gq.PreviousClose, "169.2000", 169.2},
		{"change", gq.Change, "1.5500", 1.55},
		{"change percent", gq.ChangePercent, "0.9159%", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
			if tt.wantFloat != 0 {
				f, err := ParseFloat(tt.got)
				if err != nil {
					t.Errorf("ParseFloat failed: %v", err)
				}
				if f != tt.wantFloat {
					t.Errorf("ParseFloat: got %v, want %v", f, tt.wantFloat)
				}
			}
		})
	}
}

func TestParseCryptoDaily(t *testing.T) {
	data := loadFixture(t, "crypto_daily_btc.json")

	var resp CryptoDaily
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal CryptoDaily: %v", err)
	}

	// Check metadata
	if resp.MetaData.DigitalCode != "BTC" {
		t.Errorf("got DigitalCode %q, want BTC", resp.MetaData.DigitalCode)
	}
	if resp.MetaData.DigitalName != "Bitcoin" {
		t.Errorf("got DigitalName %q, want Bitcoin", resp.MetaData.DigitalName)
	}
	if resp.MetaData.MarketCode != "USD" {
		t.Errorf("got MarketCode %q, want USD", resp.MetaData.MarketCode)
	}

	// Check time series
	if len(resp.TimeSeries) < 2 {
		t.Errorf("got %d time series entries, want at least 2", len(resp.TimeSeries))
	}

	// Check first entry
	entry, ok := resp.TimeSeries["2024-04-02"]
	if !ok {
		t.Fatal("missing entry for 2024-04-02")
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"open", entry.OpenMarket, "67234.50000000"},
		{"high", entry.HighMarket, "68123.40000000"},
		{"low", entry.LowMarket, "66890.10000000"},
		{"close", entry.CloseMarket, "67987.20000000"},
		{"volume", entry.Volume, "28456.78900000"},
		{"market cap", entry.MarketCap, "1332456789000.00000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestParseRSIResponse(t *testing.T) {
	data := loadFixture(t, "rsi_btc_daily.json")

	var resp RSIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal RSIResponse: %v", err)
	}

	// Check metadata
	if resp.MetaData.Symbol != "BTC" {
		t.Errorf("got Symbol %q, want BTC", resp.MetaData.Symbol)
	}
	if resp.MetaData.Indicator != "Relative Strength Index (RSI)" {
		t.Errorf("got Indicator %q, want 'Relative Strength Index (RSI)'", resp.MetaData.Indicator)
	}
	// TimePeriod can be either float64 or string in JSON
	timePeriod := resp.MetaData.TimePeriod
	if f, ok := timePeriod.(float64); ok {
		if f != 14 {
			t.Errorf("got TimePeriod %v, want 14", f)
		}
	} else if s, ok := timePeriod.(string); ok {
		if s != "14" {
			t.Errorf("got TimePeriod %q, want '14'", s)
		}
	} else {
		t.Errorf("TimePeriod has unexpected type %T", timePeriod)
	}

	// Check technical analysis
	if len(resp.TechnicalAnalysis) < 4 {
		t.Errorf("got %d RSI entries, want at least 4", len(resp.TechnicalAnalysis))
	}

	// Check first entry
	entry, ok := resp.TechnicalAnalysis["2024-04-02"]
	if !ok {
		t.Fatal("missing RSI entry for 2024-04-02")
	}

	if entry.RSI != "58.2345" {
		t.Errorf("got RSI %q, want '58.2345'", entry.RSI)
	}

	// Parse and validate RSI values
	rsiValue, err := ParseFloat(entry.RSI)
	if err != nil {
		t.Errorf("ParseFloat failed: %v", err)
	}
	if rsiValue != 58.2345 {
		t.Errorf("got RSI value %v, want 58.2345", rsiValue)
	}
}

func TestParseEMAResponse(t *testing.T) {
	data := loadFixture(t, "ema_aapl_daily.json")

	var resp EMAResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal EMAResponse: %v", err)
	}

	// Check metadata
	if resp.MetaData.Symbol != "AAPL" {
		t.Errorf("got Symbol %q, want AAPL", resp.MetaData.Symbol)
	}
	if resp.MetaData.Indicator != "Exponential Moving Average (EMA)" {
		t.Errorf("got Indicator %q, want 'Exponential Moving Average (EMA)'", resp.MetaData.Indicator)
	}
	// TimePeriod can be either float64 or string in JSON
	timePeriod := resp.MetaData.TimePeriod
	if f, ok := timePeriod.(float64); ok {
		if f != 9 {
			t.Errorf("got TimePeriod %v, want 9", f)
		}
	} else if s, ok := timePeriod.(string); ok {
		if s != "9" {
			t.Errorf("got TimePeriod %q, want '9'", s)
		}
	} else {
		t.Errorf("TimePeriod has unexpected type %T", timePeriod)
	}

	// Check technical analysis
	if len(resp.TechnicalAnalysis) < 4 {
		t.Errorf("got %d EMA entries, want at least 4", len(resp.TechnicalAnalysis))
	}

	// Check first entry
	entry, ok := resp.TechnicalAnalysis["2024-04-02"]
	if !ok {
		t.Fatal("missing EMA entry for 2024-04-02")
	}

	if entry.EMA != "168.7543" {
		t.Errorf("got EMA %q, want '168.7543'", entry.EMA)
	}

	// Parse and validate EMA values
	emaValue, err := ParseFloat(entry.EMA)
	if err != nil {
		t.Errorf("ParseFloat failed: %v", err)
	}
	if emaValue != 168.7543 {
		t.Errorf("got EMA value %v, want 168.7543", emaValue)
	}
}

func TestParseNewsSentiment(t *testing.T) {
	data := loadFixture(t, "news_sentiment.json")

	var resp NewsSentiment
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal NewsSentiment: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Fatalf("got %d news items, want 2", len(resp.Items))
	}

	// Check first item
	item := resp.Items[0]
	if item.Title != "Apple Inc. announces quarterly earnings beating expectations" {
		t.Errorf("got title %q", item.Title)
	}
	if item.Source != "TechNews Daily" {
		t.Errorf("got source %q, want 'TechNews Daily'", item.Source)
	}
	if item.Topic != "Earnings" {
		t.Errorf("got topic %q, want 'Earnings'", item.Topic)
	}
	if item.SentimentLabel != "Somewhat-Bullish" {
		t.Errorf("got sentiment label %q, want 'Somewhat-Bullish'", item.SentimentLabel)
	}

	// Parse and validate sentiment score
	score, err := ParseFloat(item.OverallSentimentScore)
	if err != nil {
		t.Errorf("ParseFloat failed: %v", err)
	}
	if score != 0.345 {
		t.Errorf("got sentiment score %v, want 0.345", score)
	}

	// Check second item
	item2 := resp.Items[1]
	if item2.Title != "Market volatility continues amid economic uncertainty" {
		t.Errorf("got title %q", item2.Title)
	}
	if item2.SentimentLabel != "Somewhat-Bearish" {
		t.Errorf("got sentiment label %q, want 'Somewhat-Bearish'", item2.SentimentLabel)
	}
}

func TestParseMacroDataPoint(t *testing.T) {
	data := loadFixture(t, "macro_gdp.json")

	var resp MacroResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal MacroResponse: %v", err)
	}

	// Check metadata
	if resp.Name != "Real Gross Domestic Product" {
		t.Errorf("got name %q, want 'Real Gross Domestic Product'", resp.Name)
	}
	if resp.Interval != "quarterly" {
		t.Errorf("got interval %q, want 'quarterly'", resp.Interval)
	}
	if resp.Unit != "billions of dollars" {
		t.Errorf("got unit %q, want 'billions of dollars'", resp.Unit)
	}

	// Check data
	if len(resp.Data) != 4 {
		t.Fatalf("got %d data points, want 4", len(resp.Data))
	}

	// Check first data point
	dp := resp.Data[0]
	if dp.Date != "2023-10-01" {
		t.Errorf("got date %q, want '2023-10-01'", dp.Date)
	}

	// Parse and validate value
	value, err := ParseFloat(dp.Value)
	if err != nil {
		t.Errorf("ParseFloat failed: %v", err)
	}
	if value != 27610.2 {
		t.Errorf("got value %v, want 27610.2", value)
	}
}

func TestParseFloat_ValidNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"integer", "123", 123},
		{"positive float", "123.456", 123.456},
		{"negative number", "-123.456", -123.456},
		{"zero", "0", 0},
		{"with trailing zeros", "169.5000", 169.5},
		{"large number", "1332456789000.00000000", 1.332456789e12},
		{"very small", "0.000123", 0.000123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFloat(tt.input)
			if err != nil {
				t.Fatalf("ParseFloat(%q) failed: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseFloat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFloat_EmptyString(t *testing.T) {
	got, err := ParseFloat("")
	if err != nil {
		t.Fatalf("ParseFloat(\"\") failed: %v", err)
	}
	if got != 0 {
		t.Errorf("ParseFloat(\"\") = %v, want 0", got)
	}
}

func TestParseFloat_None(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"None string", "None", 0},
		{"dash", "-", 0},
		{"null string", "null", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFloat(tt.input)
			if err != nil {
				t.Fatalf("ParseFloat(%q) failed: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseFloat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFloat_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid characters", "abc"},
		{"mixed invalid", "12a34"},
		{"only decimal point", "."},
		{"multiple decimal points", "12.34.56"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFloat(tt.input)
			if err == nil {
				t.Errorf("ParseFloat(%q) should have returned error", tt.input)
			}
		})
	}
}

func TestParseDate_ValidDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{
			name:  "standard date",
			input: "2024-04-02",
			want:  time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "with time",
			input: "2024-04-02 14:30:00",
			want:  time.Date(2024, 4, 2, 14, 30, 0, 0, time.UTC),
		},
		{
			name:  "end of day",
			input: "2024-04-02 23:59:59",
			want:  time.Date(2024, 4, 2, 23, 59, 59, 0, time.UTC),
		},
		{
			name:  "start of year",
			input: "2024-01-01",
			want:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate(%q) failed: %v", tt.input, err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseDate(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDate_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"wrong separator", "2024/04/02"},
		{"missing leading zeros", "2024-4-2"},
		{"invalid month", "2024-13-01"},
		{"invalid day", "2024-04-32"},
		{"garbage", "not-a-date"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDate(tt.input)
			if err == nil {
				t.Errorf("ParseDate(%q) should have returned error", tt.input)
			}
		})
	}
}

func TestParseDate_AlphaVantageFormat(t *testing.T) {
	// Alpha Vantage news uses a different format: YYYYMMDDTHHMMSS
	// This tests that we handle the common formats properly
	t.Run("news format with manual parsing", func(t *testing.T) {
		// Parse the Alpha Vantage news time_published format
		// Format: 20240402T143000 -> 2024-04-02 14:30:00
		input := "20240402T143000"
		// Manually transform to our supported format
		if len(input) >= 15 {
			transformed := input[0:4] + "-" + input[4:6] + "-" + input[6:8] + " " +
				input[9:11] + ":" + input[11:13] + ":" + input[13:15]
			got, err := ParseDate(transformed)
			if err != nil {
				t.Fatalf("ParseDate(%q) failed: %v", transformed, err)
			}
			want := time.Date(2024, 4, 2, 14, 30, 0, 0, time.UTC)
			if !got.Equal(want) {
				t.Errorf("got %v, want %v", got, want)
			}
		}
	})
}
