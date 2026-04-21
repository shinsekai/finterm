package cache

import (
	"encoding/gob"

	"github.com/shinsekai/finterm/internal/alphavantage"
)

func init() {
	gob.Register(&alphavantage.GlobalQuote{})
	gob.Register(&alphavantage.GlobalQuoteResponse{})
	gob.Register(&alphavantage.TimeSeriesDaily{})
	gob.Register(&alphavantage.TimeSeriesMetadata{})
	gob.Register(&alphavantage.TimeSeriesEntry{})
	gob.Register(&alphavantage.CryptoDaily{})
	gob.Register(&alphavantage.CryptoMetadata{})
	gob.Register(&alphavantage.CryptoEntry{})
	gob.Register(&alphavantage.CryptoIntraday{})
	gob.Register(&alphavantage.RSIResponse{})
	gob.Register(&alphavantage.IndicatorMetadata{})
	gob.Register(&alphavantage.RSIEntry{})
	gob.Register(&alphavantage.EMAResponse{})
	gob.Register(&alphavantage.EMAEntry{})
	gob.Register(&alphavantage.NewsSentiment{})
	gob.Register(&alphavantage.TickerSentiment{})
	gob.Register(&alphavantage.TopicItem{})
	gob.Register(&alphavantage.NewsItem{})
	gob.Register(&alphavantage.MacroDataPoint{})
	gob.Register(&alphavantage.MacroResponse{})
	gob.Register(&alphavantage.MarketStatus{})
	gob.Register(&alphavantage.MarketStatusResponse{})
}
