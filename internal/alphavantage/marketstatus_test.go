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

func TestClient_GetMarketStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("function") != functionMarketStatus {
			t.Errorf("expected function %s, got %s", functionMarketStatus, r.URL.Query().Get("function"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"endpoint": "Market Status",
			"markets": [
				{
					"market_type": "Equity",
					"region": "United States",
					"primary_exchanges": "NASDAQ, NYSE, AMEX",
					"local_open": "09:30",
					"local_close": "16:00",
					"current_status": "open",
					"notes": ""
				},
				{
					"market_type": "Cryptocurrency",
					"region": "Global",
					"primary_exchanges": "",
					"local_open": "00:00",
					"local_close": "23:59",
					"current_status": "open",
					"notes": "Crypto markets operate 24/7"
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

	result, err := client.GetMarketStatus(context.Background())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.Endpoint != "Market Status" {
		t.Errorf("expected endpoint 'Market Status', got %s", result.Endpoint)
	}

	if len(result.Markets) != 2 {
		t.Errorf("expected 2 market entries, got %d", len(result.Markets))
	}

	// Check first market entry
	usEquity := result.Markets[0]
	if usEquity.MarketType != "Equity" {
		t.Errorf("expected market type 'Equity', got %s", usEquity.MarketType)
	}
	if usEquity.Region != "United States" {
		t.Errorf("expected region 'United States', got %s", usEquity.Region)
	}
	if usEquity.CurrentStatus != "open" {
		t.Errorf("expected status 'open', got %s", usEquity.CurrentStatus)
	}

	// Check second market entry
	crypto := result.Markets[1]
	if crypto.MarketType != "Cryptocurrency" {
		t.Errorf("expected market type 'Cryptocurrency', got %s", crypto.MarketType)
	}
}

func TestClient_GetMarketStatus_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"endpoint": "Market Status", "markets": []}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetMarketStatus(ctx)
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestClient_GetMarketStatus_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Invalid JSON
		_, _ = fmt.Fprint(w, `{"endpoint": "Market Status", "markets": [}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetMarketStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed response, got nil")
	}

	if !strings.Contains(err.Error(), "parsing market status response") {
		t.Errorf("expected parsing error, got: %v", err)
	}
}

func TestClient_GetMarketStatus_CallsCorrectFunction(t *testing.T) {
	var functionCalled string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		functionCalled = r.URL.Query().Get("function")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"endpoint": "Market Status",
			"markets": [
				{
					"market_type": "Forex",
					"region": "Global",
					"primary_exchanges": "",
					"local_open": "00:00",
					"local_close": "23:59",
					"current_status": "open",
					"notes": "Forex markets operate 24/5"
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

	_, err := client.GetMarketStatus(context.Background())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if functionCalled != functionMarketStatus {
		t.Errorf("expected function %s to be called, got %s", functionMarketStatus, functionCalled)
	}
}

func TestClient_GetMarketStatus_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Valid JSON but missing endpoint field
		_, _ = fmt.Fprint(w, `{"markets": []}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetMarketStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}

	if !strings.Contains(err.Error(), "empty market status response") {
		t.Errorf("expected empty response error, got: %v", err)
	}
}

func TestClient_GetMarketStatus_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, `{"Note": "Thank you for using Alpha Vantage! Our standard API rate limit is 25 requests per day."}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	_, err := client.GetMarketStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for rate limit, got nil")
	}
}

func TestClient_GetMarketStatus_AllMarketTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"endpoint": "Market Status",
			"markets": [
				{
					"market_type": "Equity",
					"region": "United States",
					"primary_exchanges": "NASDAQ, NYSE, AMEX",
					"local_open": "09:30",
					"local_close": "16:00",
					"current_status": "open",
					"notes": ""
				},
				{
					"market_type": "Forex",
					"region": "Global",
					"primary_exchanges": "",
					"local_open": "00:00",
					"local_close": "23:59",
					"current_status": "open",
					"notes": "Forex markets operate 24/5"
				},
				{
					"market_type": "Commodity",
					"region": "Global",
					"primary_exchanges": "CME, ICE",
					"local_open": "00:00",
					"local_close": "23:59",
					"current_status": "open",
					"notes": ""
				},
				{
					"market_type": "Cryptocurrency",
					"region": "Global",
					"primary_exchanges": "",
					"local_open": "00:00",
					"local_close": "23:59",
					"current_status": "open",
					"notes": "Crypto markets operate 24/7"
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

	result, err := client.GetMarketStatus(context.Background())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(result.Markets) != 4 {
		t.Errorf("expected 4 market entries, got %d", len(result.Markets))
	}

	marketTypes := make(map[string]bool)
	for _, market := range result.Markets {
		marketTypes[market.MarketType] = true
	}

	expectedTypes := []string{"Equity", "Forex", "Commodity", "Cryptocurrency"}
	for _, expectedType := range expectedTypes {
		if !marketTypes[expectedType] {
			t.Errorf("expected market type %s not found in results", expectedType)
		}
	}
}
