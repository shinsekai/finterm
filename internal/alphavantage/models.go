// Package alphavantage provides typed structs for Alpha Vantage API responses.
// All structs are designed to deserialize from the actual API JSON responses.
package alphavantage

import (
	"encoding/json"
	"strconv"
	"time"
)

// GlobalQuote represents a global quote response for a single ticker.
// Example: {"Global Quote": {...}}
type GlobalQuote struct {
	Symbol         string `json:"01. symbol"`
	Open           string `json:"02. open"`
	High           string `json:"03. high"`
	Low            string `json:"04. low"`
	Price          string `json:"05. price"`
	Volume         string `json:"06. volume"`
	LastTradingDay string `json:"07. latest trading day"`
	PreviousClose  string `json:"08. previous close"`
	Change         string `json:"09. change"`
	ChangePercent  string `json:"10. change percent"`
}

// GlobalQuoteResponse wraps a GlobalQuote with metadata.
type GlobalQuoteResponse struct {
	GlobalQuote *GlobalQuote `json:"Global Quote"`
}

// TimeSeriesDaily represents the daily time series data for an equity.
type TimeSeriesDaily struct {
	MetaData   TimeSeriesMetadata         `json:"Meta Data"`
	TimeSeries map[string]TimeSeriesEntry `json:"Time Series (Daily)"`
}

// TimeSeriesMetadata contains common metadata for time series responses.
type TimeSeriesMetadata struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

// TimeSeriesEntry contains OHLCV data for a single time period.
type TimeSeriesEntry struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

// CryptoDaily represents the daily time series data for a cryptocurrency.
type CryptoDaily struct {
	MetaData   CryptoMetadata         `json:"Meta Data"`
	TimeSeries map[string]CryptoEntry `json:"Time Series (Digital Currency Daily)"`
}

// CryptoMetadata contains metadata for cryptocurrency responses.
type CryptoMetadata struct {
	Information   string `json:"1. Information"`
	DigitalCode   string `json:"2. Digital Currency Code"`
	DigitalName   string `json:"3. Digital Currency Name"`
	MarketCode    string `json:"4. Market Code"`
	MarketName    string `json:"5. Market Name"`
	LastRefreshed string `json:"6. Last Refreshed"`
	TimeZone      string `json:"7. Time Zone"`
}

// CryptoEntry contains OHLCV data for a cryptocurrency entry.
type CryptoEntry struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

// CryptoIntraday represents intraday time series data for a cryptocurrency.
type CryptoIntraday struct {
	MetaData   CryptoMetadata         `json:"Meta Data"`
	TimeSeries map[string]CryptoEntry `json:"Time Series Crypto (5min)"`
}

// RSIResponse represents the Relative Strength Index technical indicator response.
type RSIResponse struct {
	MetaData          IndicatorMetadata   `json:"Meta Data"`
	TechnicalAnalysis map[string]RSIEntry `json:"Technical Analysis: RSI"`
}

// IndicatorMetadata contains metadata for technical indicator responses.
// Note: TimePeriod can be either a number or string in API responses.
type IndicatorMetadata struct {
	Symbol        string `json:"1: Symbol"`
	Indicator     string `json:"2: Indicator"`
	LastRefreshed string `json:"3: Last Refreshed"`
	Interval      string `json:"4: Interval"`
	TimePeriod    any    `json:"5: Time Period"`
	SeriesType    string `json:"6: Series Type"`
	TimeZone      string `json:"7: Time Zone"`
}

// RSIEntry contains RSI value for a single time period.
type RSIEntry struct {
	RSI string `json:"RSI"`
}

// EMAResponse represents the Exponential Moving Average technical indicator response.
type EMAResponse struct {
	MetaData          IndicatorMetadata   `json:"Meta Data"`
	TechnicalAnalysis map[string]EMAEntry `json:"Technical Analysis: EMA"`
}

// EMAEntry contains EMA value for a single time period.
type EMAEntry struct {
	EMA string `json:"EMA"`
}

// NewsSentiment represents the news and sentiment data response.
type NewsSentiment struct {
	Items []NewsItem `json:"feed"`
}

// TickerSentiment contains sentiment data for a specific ticker within a news article.
type TickerSentiment struct {
	Ticker               string      `json:"ticker"`
	RelevanceScore       json.Number `json:"relevance_score"`
	TickerSentimentScore json.Number `json:"ticker_sentiment_score"`
	TickerSentimentLabel string      `json:"ticker_sentiment_label"`
}

// TopicItem represents a topic associated with a news article.
type TopicItem struct {
	Topic          string `json:"topic"`
	RelevanceScore string `json:"relevance_score"`
}

// NewsItem contains a single news article with sentiment analysis.
type NewsItem struct {
	Title                 string            `json:"title"`
	URL                   string            `json:"url"`
	TimePublished         string            `json:"time_published"`
	Authors               []string          `json:"authors"`
	Summary               string            `json:"summary"`
	BannerImage           string            `json:"banner_image"`
	Source                string            `json:"source"`
	CategoryWithin        string            `json:"category_within"`
	Topics                []TopicItem       `json:"topics"`
	OverallSentimentScore json.Number       `json:"overall_sentiment_score"`
	SentimentLabel        string            `json:"overall_sentiment_label"`
	Tickers               []TickerSentiment `json:"ticker_sentiment"`
}

// MacroDataPoint represents a single macroeconomic data point.
type MacroDataPoint struct {
	Date  string `json:"date"`
	Value string `json:"value"`
}

// MacroResponse represents a macroeconomic data response.
type MacroResponse struct {
	Name     string           `json:"name"`
	Interval string           `json:"interval"`
	Unit     string           `json:"unit"`
	Data     []MacroDataPoint `json:"data"`
}

// MarketStatus represents the market status for a single market.
type MarketStatus struct {
	MarketID      string `json:"market_id"`
	MarketName    string `json:"market_name"`
	Region        string `json:"region"`
	LocalOpen     string `json:"local_open"`
	LocalClose    string `json:"local_close"`
	CurrentStatus string `json:"current_status"`
}

// MarketStatusResponse represents the market status response.
type MarketStatusResponse struct {
	Markets []MarketStatus `json:"markets"`
}

// ParseFloat converts an Alpha Vantage string-encoded number to float64.
// Handles empty strings, "None", and "-" values by returning 0, error.
func ParseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	if s == "None" || s == "-" || s == "null" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

// ParseDate parses an Alpha Vantage date string in YYYY-MM-DD format.
// Also handles YYYY-MM-DD HH:MM:SS format for intraday data.
func ParseDate(s string) (time.Time, error) {
	// Try full timestamp format first (for intraday)
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t, nil
	}
	// Try date-only format
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, &time.ParseError{Layout: "2006-01-02", Value: s}
}
