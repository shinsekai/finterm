package alphavantage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGetDailyTimeSeries_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("function") != functionDailyTimeSeries {
			t.Errorf("expected function %s, got %s", functionDailyTimeSeries, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("symbol") != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", r.URL.Query().Get("symbol"))
		}
		if r.URL.Query().Get("outputsize") != "compact" {
			t.Errorf("expected outputsize compact, got %s", r.URL.Query().Get("outputsize"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1. Information": "Daily Prices (open, high, low, close) and Volumes",
				"2. Symbol": "AAPL",
				"3. Last Refreshed": "2024-04-02",
				"4. Output Size": "Compact",
				"5. Time Zone": "US/Eastern"
			},
			"Time Series (Daily)": {
				"2024-04-02": {
					"1. open": "169.5000",
					"2. high": "171.2000",
					"3. low": "168.8000",
					"4. close": "170.7500",
					"5. volume": "52134200"
				},
				"2024-04-01": {
					"1. open": "168.0000",
					"2. high": "169.5000",
					"3. low": "167.5000",
					"4. close": "169.2000",
					"5. volume": "48765300"
				}
			}
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	result, err := client.GetDailyTimeSeries(context.Background(), "AAPL", "compact")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.MetaData.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", result.MetaData.Symbol)
	}

	if len(result.TimeSeries) != 2 {
		t.Errorf("expected 2 time series entries, got %d", len(result.TimeSeries))
	}

	// Check specific entry
	entry, exists := result.TimeSeries["2024-04-02"]
	if !exists {
		t.Fatal("expected entry for 2024-04-02")
	}
	if entry.Close != "170.7500" {
		t.Errorf("expected close 170.7500, got %s", entry.Close)
	}
}

func TestGetDailyTimeSeries_InvalidSymbol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1. Information": "Daily Prices (open, high, low, close) and Volumes",
				"2. Symbol": "INVALID",
				"3. Last Refreshed": "2024-04-02",
				"4. Output Size": "Compact",
				"5. Time Zone": "US/Eastern"
			},
			"Time Series (Daily)": {}
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetDailyTimeSeries(context.Background(), "INVALID", "compact")
	if err == nil {
		t.Fatal("expected error for invalid symbol, got nil")
	}

	if !strings.Contains(err.Error(), "empty time series data") {
		t.Errorf("expected empty time series error, got: %v", err)
	}
}

func TestGetDailyTimeSeries_EmptySymbol(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetDailyTimeSeries(context.Background(), "", "compact")
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	if !strings.Contains(err.Error(), "symbol cannot be empty") {
		t.Errorf("expected symbol empty error, got: %v", err)
	}
}

func TestGetDailyTimeSeries_ContextPropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetDailyTimeSeries(ctx, "AAPL", "compact")
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestGetDailyTimeSeries_FullOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("outputsize") != "full" {
			t.Errorf("expected outputsize full, got %s", r.URL.Query().Get("outputsize"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1. Information": "Daily Prices",
				"2. Symbol": "AAPL",
				"3. Last Refreshed": "2024-04-02",
				"4. Output Size": "Full size",
				"5. Time Zone": "US/Eastern"
			},
			"Time Series (Daily)": {
				"2024-04-02": {
					"1. open": "169.5000",
					"2. high": "171.2000",
					"3. low": "168.8000",
					"4. close": "170.7500",
					"5. volume": "52134200"
				}
			}
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetDailyTimeSeries(context.Background(), "AAPL", "full")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestGetCryptoDaily_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("function") != functionCryptoDaily {
			t.Errorf("expected function %s, got %s", functionCryptoDaily, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("symbol") != "BTC" {
			t.Errorf("expected symbol BTC, got %s", r.URL.Query().Get("symbol"))
		}
		if r.URL.Query().Get("market") != "USD" {
			t.Errorf("expected market USD, got %s", r.URL.Query().Get("market"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1. Information": "Daily Prices and Volumes for Digital Currency",
				"2. Digital Currency Code": "BTC",
				"3. Digital Currency Name": "Bitcoin",
				"4. Market Code": "USD",
				"5. Market Name": "United States Dollar",
				"6. Last Refreshed": "2024-04-02 20:00:00",
				"7. Time Zone": "UTC"
			},
			"Time Series (Digital Currency Daily)": {
				"2024-04-02": {
					"1a. open (USD)": "67234.50000000",
					"2a. high (USD)": "68123.40000000",
					"3a. low (USD)": "66890.10000000",
					"4a. close (USD)": "67987.20000000",
					"5. volume": "28456.78900000",
					"6. market cap (USD)": "1332456789000.00000000"
				},
				"2024-04-01": {
					"1a. open (USD)": "66543.20000000",
					"2a. high (USD)": "67345.60000000",
					"3a. low (USD)": "66000.00000000",
					"4a. close (USD)": "67123.80000000",
					"5. volume": "26789.45600000",
					"6. market cap (USD)": "1315678901000.00000000"
				}
			}
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	result, err := client.GetCryptoDaily(context.Background(), "BTC", "USD")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.MetaData.DigitalCode != "BTC" {
		t.Errorf("expected digital code BTC, got %s", result.MetaData.DigitalCode)
	}

	if result.MetaData.MarketCode != "USD" {
		t.Errorf("expected market code USD, got %s", result.MetaData.MarketCode)
	}

	if len(result.TimeSeries) != 2 {
		t.Errorf("expected 2 time series entries, got %d", len(result.TimeSeries))
	}

	// Check specific entry
	entry, exists := result.TimeSeries["2024-04-02"]
	if !exists {
		t.Fatal("expected entry for 2024-04-02")
	}
	if entry.CloseMarket != "67987.20000000" {
		t.Errorf("expected close 67987.20000000, got %s", entry.CloseMarket)
	}
}

func TestGetCryptoDaily_WeekendDataIncluded(t *testing.T) {
	// Crypto markets operate 24/7, so weekend data should be included
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1. Information": "Daily Prices and Volumes for Digital Currency",
				"2. Digital Currency Code": "BTC",
				"3. Digital Currency Name": "Bitcoin",
				"4. Market Code": "USD",
				"5. Market Name": "United States Dollar",
				"6. Last Refreshed": "2024-04-02 20:00:00",
				"7. Time Zone": "UTC"
			},
			"Time Series (Digital Currency Daily)": {
				"2024-04-02": {
					"1a. open (USD)": "67234.50000000",
					"2a. high (USD)": "68123.40000000",
					"3a. low (USD)": "66890.10000000",
					"4a. close (USD)": "67987.20000000",
					"5. volume": "28456.78900000",
					"6. market cap (USD)": "1332456789000.00000000"
				},
				"2024-03-31": {
					"1a. open (USD)": "65500.00000000",
					"2a. high (USD)": "66000.00000000",
					"3a. low (USD)": "65000.00000000",
					"4a. close (USD)": "65800.00000000",
					"5. volume": "25000.00000000",
					"6. market cap (USD)": "1280000000000.00000000"
				},
				"2024-03-30": {
					"1a. open (USD)": "65000.00000000",
					"2a. high (USD)": "65700.00000000",
					"3a. low (USD)": "64500.00000000",
					"4a. close (USD)": "65500.00000000",
					"5. volume": "24000.00000000",
					"6. market cap (USD)": "1275000000000.00000000"
				}
			}
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	result, err := client.GetCryptoDaily(context.Background(), "BTC", "USD")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// 2024-03-30 is a Saturday and 2024-03-31 is a Sunday
	// Crypto should have weekend data
	_, satExists := result.TimeSeries["2024-03-30"]
	_, sunExists := result.TimeSeries["2024-03-31"]

	if !satExists {
		t.Error("expected Saturday (2024-03-30) data for crypto")
	}
	if !sunExists {
		t.Error("expected Sunday (2024-03-31) data for crypto")
	}
}

func TestGetCryptoDaily_EmptySymbol(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetCryptoDaily(context.Background(), "", "USD")
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	if !strings.Contains(err.Error(), "crypto symbol cannot be empty") {
		t.Errorf("expected empty symbol error, got: %v", err)
	}
}

func TestGetCryptoDaily_EmptyMarket(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetCryptoDaily(context.Background(), "BTC", "")
	if err == nil {
		t.Fatal("expected error for empty market, got nil")
	}

	if !strings.Contains(err.Error(), "market cannot be empty") {
		t.Errorf("expected empty market error, got: %v", err)
	}
}

func TestGetCryptoIntraday_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("function") != functionCryptoIntraday {
			t.Errorf("expected function %s, got %s", functionCryptoIntraday, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("symbol") != "BTC" {
			t.Errorf("expected symbol BTC, got %s", r.URL.Query().Get("symbol"))
		}
		if r.URL.Query().Get("market") != "USD" {
			t.Errorf("expected market USD, got %s", r.URL.Query().Get("market"))
		}
		if r.URL.Query().Get("interval") != "5min" {
			t.Errorf("expected interval 5min, got %s", r.URL.Query().Get("interval"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1. Information": "Crypto Intraday Prices",
				"2. Digital Currency Code": "BTC",
				"3. Digital Currency Name": "Bitcoin",
				"4. Market Code": "USD",
				"5. Market Name": "United States Dollar",
				"6. Last Refreshed": "2024-04-02 20:00:00",
				"7. Time Zone": "UTC"
			},
			"Time Series Crypto (5min)": {
				"2024-04-02 20:00:00": {
					"1a. open (USD)": "67234.50000000",
					"2a. high (USD)": "68123.40000000",
					"3a. low (USD)": "66890.10000000",
					"4a. close (USD)": "67987.20000000",
					"5. volume": "28456.78900000",
					"6. market cap (USD)": "1332456789000.00000000"
				},
				"2024-04-02 19:55:00": {
					"1a. open (USD)": "66543.20000000",
					"2a. high (USD)": "67345.60000000",
					"3a. low (USD)": "66000.00000000",
					"4a. close (USD)": "67234.50000000",
					"5. volume": "26789.45600000",
					"6. market cap (USD)": "1315678901000.00000000"
				}
			}
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	result, err := client.GetCryptoIntraday(context.Background(), "BTC", "USD", "5min")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.MetaData.DigitalCode != "BTC" {
		t.Errorf("expected digital code BTC, got %s", result.MetaData.DigitalCode)
	}

	if len(result.TimeSeries) != 2 {
		t.Errorf("expected 2 time series entries, got %d", len(result.TimeSeries))
	}
}

func TestGetCryptoIntraday_InvalidInterval(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetCryptoIntraday(context.Background(), "BTC", "USD", "10min")
	if err == nil {
		t.Fatal("expected error for invalid interval, got nil")
	}

	if !strings.Contains(err.Error(), "invalid interval") {
		t.Errorf("expected invalid interval error, got: %v", err)
	}
}

func TestGetCryptoIntraday_ValidIntervals(t *testing.T) {
	tests := []struct {
		name     string
		interval string
	}{
		{"1 minute", "1min"},
		{"5 minutes", "5min"},
		{"15 minutes", "15min"},
		{"30 minutes", "30min"},
		{"60 minutes", "60min"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("interval") != tt.interval {
					t.Errorf("expected interval %s, got %s", tt.interval, r.URL.Query().Get("interval"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{
					"Meta Data": {
						"1. Information": "Crypto Intraday Prices",
						"2. Digital Currency Code": "BTC",
						"3. Digital Currency Name": "Bitcoin",
						"4. Market Code": "USD",
						"5. Market Name": "United States Dollar",
						"6. Last Refreshed": "2024-04-02 20:00:00",
						"7. Time Zone": "UTC"
					},
					"Time Series Crypto (`+tt.interval+`)": {
						"2024-04-02 20:00:00": {
							"1a. open (USD)": "67234.50000000",
							"2a. high (USD)": "68123.40000000",
							"3a. low (USD)": "66890.10000000",
							"4a. close (USD)": "67987.20000000",
							"5. volume": "28456.78900000",
							"6. market cap (USD)": "1332456789000.00000000"
						}
					}
				}`)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{Timeout: 5 * time.Second},
			})

			_, err := client.GetCryptoIntraday(context.Background(), "BTC", "USD", tt.interval)
			if err != nil {
				t.Fatalf("expected success for interval %s, got error: %v", tt.interval, err)
			}
		})
	}
}

func TestGetCryptoIntraday_EmptySymbol(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetCryptoIntraday(context.Background(), "", "USD", "5min")
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	if !strings.Contains(err.Error(), "crypto symbol cannot be empty") {
		t.Errorf("expected empty symbol error, got: %v", err)
	}
}

func TestGetCryptoIntraday_EmptyMarket(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetCryptoIntraday(context.Background(), "BTC", "", "5min")
	if err == nil {
		t.Fatal("expected error for empty market, got nil")
	}

	if !strings.Contains(err.Error(), "market cannot be empty") {
		t.Errorf("expected empty market error, got: %v", err)
	}
}

func TestGetCryptoIntraday_EmptyInterval(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetCryptoIntraday(context.Background(), "BTC", "USD", "")
	if err == nil {
		t.Fatal("expected error for empty interval, got nil")
	}

	if !strings.Contains(err.Error(), "interval cannot be empty") {
		t.Errorf("expected empty interval error, got: %v", err)
	}
}

func TestGetGlobalQuote_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("function") != functionGlobalQuote {
			t.Errorf("expected function %s, got %s", functionGlobalQuote, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("symbol") != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", r.URL.Query().Get("symbol"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Global Quote": {
				"01. symbol": "AAPL",
				"02. open": "169.5000",
				"03. high": "171.2000",
				"04. low": "168.8000",
				"05. price": "170.7500",
				"06. volume": "52134200",
				"07. latest trading day": "2024-04-02",
				"08. previous close": "169.2000",
				"09. change": "1.5500",
				"10. change percent": "0.9159%"
			}
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	result, err := client.GetGlobalQuote(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", result.Symbol)
	}

	if result.Price != "170.7500" {
		t.Errorf("expected price 170.7500, got %s", result.Price)
	}

	if result.Volume != "52134200" {
		t.Errorf("expected volume 52134200, got %s", result.Volume)
	}
}

func TestGetGlobalQuote_EmptySymbol(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetGlobalQuote(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	if !strings.Contains(err.Error(), "symbol cannot be empty") {
		t.Errorf("expected empty symbol error, got: %v", err)
	}
}

func TestGetGlobalQuote_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"Global Quote": null}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetGlobalQuote(context.Background(), "INVALID")
	if err == nil {
		t.Fatal("expected error for empty quote, got nil")
	}

	if !strings.Contains(err.Error(), "empty global quote") {
		t.Errorf("expected empty quote error, got: %v", err)
	}
}

func TestGetGlobalQuote_ContextPropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetGlobalQuote(ctx, "AAPL")
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestGetBulkQuotes_Under100(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("function") != functionRealtimeBulkQuotes {
			t.Errorf("expected function %s, got %s", functionRealtimeBulkQuotes, r.URL.Query().Get("function"))
		}

		symbols := r.URL.Query().Get("symbols")
		expectedSymbols := "AAPL,MSFT,GOOGL"
		if symbols != expectedSymbols {
			t.Errorf("expected symbols %s, got %s", expectedSymbols, symbols)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"data": [
				{
					"01. symbol": "AAPL",
					"02. open": "169.5000",
					"03. high": "171.2000",
					"04. low": "168.8000",
					"05. price": "170.7500",
					"06. volume": "52134200",
					"07. latest trading day": "2024-04-02",
					"08. previous close": "169.2000",
					"09. change": "1.5500",
					"10. change percent": "0.9159%"
				},
				{
					"01. symbol": "MSFT",
					"02. open": "420.1000",
					"03. high": "425.3000",
					"04. low": "418.5000",
					"05. price": "423.8000",
					"06. volume": "22154300",
					"07. latest trading day": "2024-04-02",
					"08. previous close": "420.5000",
					"09. change": "3.3000",
					"10. change percent": "0.7848%"
				},
				{
					"01. symbol": "GOOGL",
					"02. open": "170.2000",
					"03. high": "173.1000",
					"04. low": "169.5000",
					"05. price": "172.8000",
					"06. volume": "18543200",
					"07. latest trading day": "2024-04-02",
					"08. previous close": "170.0000",
					"09. change": "2.8000",
					"10. change percent": "1.6471%"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	symbols := []string{"AAPL", "MSFT", "GOOGL"}
	quotes, err := client.GetBulkQuotes(context.Background(), symbols)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(quotes) != 3 {
		t.Errorf("expected 3 quotes, got %d", len(quotes))
	}

	// Verify symbols
	symbolsFound := make(map[string]bool)
	for _, q := range quotes {
		symbolsFound[q.Symbol] = true
	}

	for _, s := range symbols {
		if !symbolsFound[s] {
			t.Errorf("expected symbol %s not found in results", s)
		}
	}
}

func TestGetBulkQuotes_Over100_Batches(t *testing.T) {
	const totalSymbols = 250
	batchCallCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		batchCallCount++

		symbolsParam := r.URL.Query().Get("symbols")
		symbols := strings.Split(symbolsParam, ",")
		numSymbols := len(symbols)

		// Verify batch size is at most 100
		if numSymbols > 100 {
			t.Errorf("expected batch size <= 100, got %d", numSymbols)
		}

		// Build response with matching number of symbols
		var data []string
		for _, sym := range symbols {
			data = append(data, fmt.Sprintf(`{
				"01. symbol": "%s",
				"02. open": "100.0000",
				"03. high": "105.0000",
				"04. low": "99.0000",
				"05. price": "102.0000",
				"06. volume": "1000000",
				"07. latest trading day": "2024-04-02",
				"08. previous close": "100.0000",
				"09. change": "2.0000",
				"10. change percent": "2.0000%%"
			}`, sym))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"data": [%s]}`, strings.Join(data, ","))
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	// Create 250 symbols
	symbols := make([]string, totalSymbols)
	for i := 0; i < totalSymbols; i++ {
		symbols[i] = fmt.Sprintf("STK%03d", i)
	}

	quotes, err := client.GetBulkQuotes(context.Background(), symbols)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(quotes) != totalSymbols {
		t.Errorf("expected %d quotes, got %d", totalSymbols, len(quotes))
	}

	// Should have made 3 batches (100, 100, 50)
	expectedBatches := 3
	if batchCallCount != expectedBatches {
		t.Errorf("expected %d batch calls, got %d", expectedBatches, batchCallCount)
	}
}

func TestGetBulkQuotes_PartialFailure(t *testing.T) {
	batchCallCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		batchCallCount++

		symbolsParam := r.URL.Query().Get("symbols")
		symbols := strings.Split(symbolsParam, ",")

		// First batch succeeds, second batch fails
		if batchCallCount == 1 {
			var data []string
			for _, sym := range symbols {
				data = append(data, fmt.Sprintf(`{
					"01. symbol": "%s",
					"02. open": "100.0000",
					"03. high": "105.0000",
					"04. low": "99.0000",
					"05. price": "102.0000",
					"06. volume": "1000000",
					"07. latest trading day": "2024-04-02",
					"08. previous close": "100.0000",
					"09. change": "2.0000",
					"10. change percent": "2.0000%%"
				}`, sym))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"data": [%s]}`, strings.Join(data, ","))
		} else {
			// Second batch fails
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, `{"error": "internal server error"}`)
		}
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	// Create 150 symbols to trigger 2 batches
	symbols := make([]string, 150)
	for i := 0; i < 150; i++ {
		symbols[i] = fmt.Sprintf("STK%03d", i)
	}

	quotes, err := client.GetBulkQuotes(context.Background(), symbols)
	if err == nil {
		t.Fatal("expected error for partial failure, got nil")
	}

	// Should have received quotes from first batch only (100 symbols)
	if len(quotes) != 100 {
		t.Errorf("expected 100 quotes from first batch, got %d", len(quotes))
	}

	if !strings.Contains(err.Error(), "batch 101-150") {
		t.Errorf("expected batch error, got: %v", err)
	}
}

func TestGetBulkQuotes_EmptySymbols(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetBulkQuotes(context.Background(), []string{})
	if err == nil {
		t.Fatal("expected error for empty symbols list, got nil")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected empty symbols error, got: %v", err)
	}
}

func TestGetBulkQuotes_AllEmptySymbols(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetBulkQuotes(context.Background(), []string{"", "", ""})
	if err == nil {
		t.Fatal("expected error for all empty symbols, got nil")
	}

	if !strings.Contains(err.Error(), "only empty strings") {
		t.Errorf("expected empty strings error, got: %v", err)
	}
}

func TestGetBulkQuotes_FilterEmptySymbols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		symbolsParam := r.URL.Query().Get("symbols")
		// Should not include empty symbols
		if strings.Contains(symbolsParam, ",,") {
			t.Error("empty symbols should be filtered out")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"data": [
				{
					"01. symbol": "AAPL",
					"02. open": "169.5000",
					"03. high": "171.2000",
					"04. low": "168.8000",
					"05. price": "170.7500",
					"06. volume": "52134200",
					"07. latest trading day": "2024-04-02",
					"08. previous close": "169.2000",
					"09. change": "1.5500",
					"10. change percent": "0.9159%"
				},
				{
					"01. symbol": "GOOGL",
					"02. open": "170.2000",
					"03. high": "173.1000",
					"04. low": "169.5000",
					"05. price": "172.8000",
					"06. volume": "18543200",
					"07. latest trading day": "2024-04-02",
					"08. previous close": "170.0000",
					"09. change": "2.8000",
					"10. change percent": "1.6471%"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	symbols := []string{"AAPL", "", "GOOGL", ""}
	quotes, err := client.GetBulkQuotes(context.Background(), symbols)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(quotes) != 2 {
		t.Errorf("expected 2 quotes (empty symbols filtered), got %d", len(quotes))
	}
}

func TestGetBulkQuotes_Exact100(t *testing.T) {
	batchCallCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batchCallCount++

		symbolsParam := r.URL.Query().Get("symbols")
		symbols := strings.Split(symbolsParam, ",")

		if len(symbols) != 100 {
			t.Errorf("expected exactly 100 symbols in batch, got %d", len(symbols))
		}

		var data []string
		for _, sym := range symbols {
			data = append(data, fmt.Sprintf(`{
				"01. symbol": "%s",
				"02. open": "100.0000",
				"03. high": "105.0000",
				"04. low": "99.0000",
				"05. price": "102.0000",
				"06. volume": "1000000",
				"07. latest trading day": "2024-04-02",
				"08. previous close": "100.0000",
				"09. change": "2.0000",
				"10. change percent": "2.0000%%"
			}`, sym))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"data": [%s]}`, strings.Join(data, ","))
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	// Create exactly 100 symbols
	symbols := make([]string, 100)
	for i := 0; i < 100; i++ {
		symbols[i] = fmt.Sprintf("STK%03d", i)
	}

	quotes, err := client.GetBulkQuotes(context.Background(), symbols)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(quotes) != 100 {
		t.Errorf("expected 100 quotes, got %d", len(quotes))
	}

	// Should have made exactly 1 batch
	if batchCallCount != 1 {
		t.Errorf("expected 1 batch call for 100 symbols, got %d", batchCallCount)
	}
}
