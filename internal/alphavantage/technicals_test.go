package alphavantage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetRSI_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("function") != functionRSI {
			t.Errorf("expected function %s, got %s", functionRSI, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("symbol") != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", r.URL.Query().Get("symbol"))
		}
		if r.URL.Query().Get("interval") != "daily" {
			t.Errorf("expected interval daily, got %s", r.URL.Query().Get("interval"))
		}
		if r.URL.Query().Get("time_period") != "14" {
			t.Errorf("expected time_period 14, got %s", r.URL.Query().Get("time_period"))
		}
		if r.URL.Query().Get("series_type") != "close" {
			t.Errorf("expected series_type close, got %s", r.URL.Query().Get("series_type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1: Symbol": "AAPL",
				"2: Indicator": "Relative Strength Index (RSI)",
				"3: Last Refreshed": "2024-04-02",
				"4: Interval": "daily",
				"5: Time Period": 14,
				"6: Series Type": "close",
				"7: Time Zone": "US/Eastern"
			},
			"Technical Analysis: RSI": {
				"2024-04-02": {
					"RSI": "58.2345"
				},
				"2024-04-01": {
					"RSI": "55.6789"
				},
				"2024-03-29": {
					"RSI": "52.3456"
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

	result, err := client.GetRSI(context.Background(), "AAPL", "daily", 14)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.MetaData.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", result.MetaData.Symbol)
	}

	if result.MetaData.Indicator != "Relative Strength Index (RSI)" {
		t.Errorf("expected indicator name, got %s", result.MetaData.Indicator)
	}

	if result.MetaData.SeriesType != "close" {
		t.Errorf("expected series_type close, got %s", result.MetaData.SeriesType)
	}

	if len(result.TechnicalAnalysis) != 3 {
		t.Errorf("expected 3 RSI entries, got %d", len(result.TechnicalAnalysis))
	}

	// Check specific entry
	entry, exists := result.TechnicalAnalysis["2024-04-02"]
	if !exists {
		t.Fatal("expected entry for 2024-04-02")
	}
	if entry.RSI != "58.2345" {
		t.Errorf("expected RSI 58.2345, got %s", entry.RSI)
	}
}

func TestGetRSI_InvalidSymbol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Error Message": "Invalid API call. Please retry or visit the documentation (https://www.alphavantage.co/documentation/) for TIME_SERIES_DAILY."
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetRSI(context.Background(), "INVALID_TICKER", "daily", 14)
	if err == nil {
		t.Fatal("expected error for invalid symbol, got nil")
	}

	if !strings.Contains(err.Error(), "Invalid API call") {
		t.Errorf("expected invalid API call error, got: %v", err)
	}
}

func TestGetRSI_ErrorInBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Note": "Thank you for using Alpha Vantage! Our standard API rate limit is 25 requests per day. Please subscribe to any of the premium plans at https://www.alphavantage.co/premium/ to instantly remove all daily rate limits."
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetRSI(context.Background(), "AAPL", "daily", 14)
	if err == nil {
		t.Fatal("expected error for rate limit, got nil")
	}

	if !strings.Contains(err.Error(), "API note") {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}

func TestGetRSI_EmptyData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1: Symbol": "INVALID",
				"2: Indicator": "Relative Strength Index (RSI)",
				"3: Last Refreshed": "2024-04-02",
				"4: Interval": "daily",
				"5: Time Period": 14,
				"6: Series Type": "close",
				"7: Time Zone": "US/Eastern"
			},
			"Technical Analysis: RSI": {}
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetRSI(context.Background(), "INVALID", "daily", 14)
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}

	if !strings.Contains(err.Error(), "empty RSI data") {
		t.Errorf("expected empty RSI data error, got: %v", err)
	}
}

func TestGetRSI_EmptySymbol(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetRSI(context.Background(), "", "daily", 14)
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	if !strings.Contains(err.Error(), "symbol cannot be empty") {
		t.Errorf("expected empty symbol error, got: %v", err)
	}
}

func TestGetRSI_EmptyInterval(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetRSI(context.Background(), "AAPL", "", 14)
	if err == nil {
		t.Fatal("expected error for empty interval, got nil")
	}

	if !strings.Contains(err.Error(), "interval cannot be empty") {
		t.Errorf("expected empty interval error, got: %v", err)
	}
}

func TestGetRSI_InvalidPeriod(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	tests := []struct {
		name   string
		period int
	}{
		{"zero period", 0},
		{"negative period", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetRSI(context.Background(), "AAPL", "daily", tt.period)
			if err == nil {
				t.Fatal("expected error for invalid period, got nil")
			}

			if !strings.Contains(err.Error(), "period must be positive") {
				t.Errorf("expected period validation error, got: %v", err)
			}
		})
	}
}

func TestGetRSI_DifferentPeriods(t *testing.T) {
	tests := []struct {
		name   string
		period int
	}{
		{"period 5", 5},
		{"period 10", 10},
		{"period 14", 14},
		{"period 21", 21},
		{"period 50", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("time_period") != fmt.Sprintf("%d", tt.period) {
					t.Errorf("expected time_period %d, got %s", tt.period, r.URL.Query().Get("time_period"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{
					"Meta Data": {
						"1: Symbol": "AAPL",
						"2: Indicator": "Relative Strength Index (RSI)",
						"3: Last Refreshed": "2024-04-02",
						"4: Interval": "daily",
						"5: Time Period": %d,
						"6: Series Type": "close",
						"7: Time Zone": "US/Eastern"
					},
					"Technical Analysis: RSI": {
						"2024-04-02": {
							"RSI": "58.2345"
						}
					}
				}`, tt.period)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{Timeout: 5 * time.Second},
			})

			_, err := client.GetRSI(context.Background(), "AAPL", "daily", tt.period)
			if err != nil {
				t.Fatalf("expected success for period %d, got error: %v", tt.period, err)
			}
		})
	}
}

func TestGetEMA_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("function") != functionEMA {
			t.Errorf("expected function %s, got %s", functionEMA, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("symbol") != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", r.URL.Query().Get("symbol"))
		}
		if r.URL.Query().Get("interval") != "daily" {
			t.Errorf("expected interval daily, got %s", r.URL.Query().Get("interval"))
		}
		if r.URL.Query().Get("time_period") != "9" {
			t.Errorf("expected time_period 9, got %s", r.URL.Query().Get("time_period"))
		}
		if r.URL.Query().Get("series_type") != "close" {
			t.Errorf("expected series_type close, got %s", r.URL.Query().Get("series_type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1: Symbol": "AAPL",
				"2: Indicator": "Exponential Moving Average (EMA)",
				"3: Last Refreshed": "2024-04-02 20:00:00",
				"4: Interval": "daily",
				"5: Time Period": 9,
				"6: Series Type": "close",
				"7: Time Zone": "US/Eastern"
			},
			"Technical Analysis: EMA": {
				"2024-04-02": {
					"EMA": "168.7543"
				},
				"2024-04-01": {
					"EMA": "167.8234"
				},
				"2024-03-29": {
					"EMA": "166.5678"
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

	result, err := client.GetEMA(context.Background(), "AAPL", "daily", 9)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.MetaData.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", result.MetaData.Symbol)
	}

	if result.MetaData.Indicator != "Exponential Moving Average (EMA)" {
		t.Errorf("expected indicator name, got %s", result.MetaData.Indicator)
	}

	if result.MetaData.SeriesType != "close" {
		t.Errorf("expected series_type close, got %s", result.MetaData.SeriesType)
	}

	if len(result.TechnicalAnalysis) != 3 {
		t.Errorf("expected 3 EMA entries, got %d", len(result.TechnicalAnalysis))
	}

	// Check specific entry
	entry, exists := result.TechnicalAnalysis["2024-04-02"]
	if !exists {
		t.Fatal("expected entry for 2024-04-02")
	}
	if entry.EMA != "168.7543" {
		t.Errorf("expected EMA 168.7543, got %s", entry.EMA)
	}
}

func TestGetEMA_InvalidPeriod(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	tests := []struct {
		name   string
		period int
	}{
		{"zero period", 0},
		{"negative period", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetEMA(context.Background(), "AAPL", "daily", tt.period)
			if err == nil {
				t.Fatal("expected error for invalid period, got nil")
			}

			if !strings.Contains(err.Error(), "period must be positive") {
				t.Errorf("expected period validation error, got: %v", err)
			}
		})
	}
}

func TestGetEMA_DifferentPeriods(t *testing.T) {
	tests := []struct {
		name   string
		period int
	}{
		{"fast EMA period 9", 9},
		{"slow EMA period 21", 21},
		{"period 50", 50},
		{"period 200", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("time_period") != fmt.Sprintf("%d", tt.period) {
					t.Errorf("expected time_period %d, got %s", tt.period, r.URL.Query().Get("time_period"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{
					"Meta Data": {
						"1: Symbol": "AAPL",
						"2: Indicator": "Exponential Moving Average (EMA)",
						"3: Last Refreshed": "2024-04-02",
						"4: Interval": "daily",
						"5: Time Period": %d,
						"6: Series Type": "close",
						"7: Time Zone": "US/Eastern"
					},
					"Technical Analysis: EMA": {
						"2024-04-02": {
							"EMA": "168.7543"
						}
					}
				}`, tt.period)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{Timeout: 5 * time.Second},
			})

			_, err := client.GetEMA(context.Background(), "AAPL", "daily", tt.period)
			if err != nil {
				t.Fatalf("expected success for period %d, got error: %v", tt.period, err)
			}
		})
	}
}

func TestGetEMA_EmptySymbol(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetEMA(context.Background(), "", "daily", 9)
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	if !strings.Contains(err.Error(), "symbol cannot be empty") {
		t.Errorf("expected empty symbol error, got: %v", err)
	}
}

func TestGetEMA_EmptyInterval(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetEMA(context.Background(), "AAPL", "", 9)
	if err == nil {
		t.Fatal("expected error for empty interval, got nil")
	}

	if !strings.Contains(err.Error(), "interval cannot be empty") {
		t.Errorf("expected empty interval error, got: %v", err)
	}
}

func TestGetEMA_EmptyData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"Meta Data": {
				"1: Symbol": "INVALID",
				"2: Indicator": "Exponential Moving Average (EMA)",
				"3: Last Refreshed": "2024-04-02",
				"4: Interval": "daily",
				"5: Time Period": 9,
				"6: Series Type": "close",
				"7: Time Zone": "US/Eastern"
			},
			"Technical Analysis: EMA": {}
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetEMA(context.Background(), "INVALID", "daily", 9)
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}

	if !strings.Contains(err.Error(), "empty EMA data") {
		t.Errorf("expected empty EMA data error, got: %v", err)
	}
}

func TestGetRSI_ContextPropagation(t *testing.T) {
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

	_, err := client.GetRSI(ctx, "AAPL", "daily", 14)
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestGetEMA_ContextPropagation(t *testing.T) {
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

	_, err := client.GetEMA(ctx, "AAPL", "daily", 9)
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}
