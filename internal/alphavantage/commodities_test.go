package alphavantage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_GetCommodity_WTI_Daily(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != string(CommodityFunctionWTI) {
			t.Errorf("expected function %s, got %s", CommodityFunctionWTI, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("interval") != "daily" {
			t.Errorf("expected interval daily, got %s", r.URL.Query().Get("interval"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Crude Oil WTI Futures",
			"interval": "daily",
			"unit": "dollars per barrel",
			"data": [
				["2024-04-01", "82.50"],
				["2024-03-31", "83.12"],
				["2024-03-30", "82.75"]
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	series, err := client.GetCommodity(context.Background(), CommodityFunctionWTI, "daily")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if series.Name != "Crude Oil WTI Futures" {
		t.Errorf("expected name 'Crude Oil WTI Futures', got %s", series.Name)
	}
	if series.Unit != "dollars per barrel" {
		t.Errorf("expected unit 'dollars per barrel', got %s", series.Unit)
	}
	if len(series.Data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(series.Data))
	}

	// Verify data is sorted newest-first
	if !series.Data[0].Date.Equal(time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected first date 2024-04-01, got %s", series.Data[0].Date)
	}
	if series.Data[0].Value != 82.50 {
		t.Errorf("expected first value 82.50, got %f", series.Data[0].Value)
	}
}

func TestClient_GetCommodity_Brent_Daily(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != string(CommodityFunctionBrent) {
			t.Errorf("expected function %s, got %s", CommodityFunctionBrent, r.URL.Query().Get("function"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Brent Crude Oil Futures",
			"interval": "daily",
			"unit": "dollars per barrel",
			"data": [
				["2024-04-01", "87.25"],
				["2024-03-31", "86.80"]
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	series, err := client.GetCommodity(context.Background(), CommodityFunctionBrent, "daily")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(series.Data) != 2 {
		t.Errorf("expected 2 data points, got %d", len(series.Data))
	}

	if series.Data[0].Value != 87.25 {
		t.Errorf("expected first value 87.25, got %f", series.Data[0].Value)
	}
}

func TestClient_GetCommodity_Copper_Monthly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != string(CommodityFunctionCopper) {
			t.Errorf("expected function %s, got %s", CommodityFunctionCopper, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("interval") != "monthly" {
			t.Errorf("expected interval monthly, got %s", r.URL.Query().Get("interval"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Global Copper Price",
			"interval": "monthly",
			"unit": "dollars per ton",
			"data": [
				["2024-04-01", "9125.50"],
				["2024-03-01", "8890.25"],
				["2024-02-01", "8565.00"]
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	series, err := client.GetCommodity(context.Background(), CommodityFunctionCopper, "monthly")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if series.Unit != "dollars per ton" {
		t.Errorf("expected unit 'dollars per ton', got %s", series.Unit)
	}

	if len(series.Data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(series.Data))
	}

	if series.Data[0].Value != 9125.50 {
		t.Errorf("expected first value 9125.50, got %f", series.Data[0].Value)
	}
}

func TestClient_GetCommodity_NaturalGas_Daily(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != string(CommodityFunctionNaturalGas) {
			t.Errorf("expected function %s, got %s", CommodityFunctionNaturalGas, r.URL.Query().Get("function"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Natural Gas Futures",
			"interval": "daily",
			"unit": "dollars per million btu",
			"data": [
				["2024-04-01", "1.85"],
				["2024-03-31", "1.78"],
				["2024-03-30", "1.82"]
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	series, err := client.GetCommodity(context.Background(), CommodityFunctionNaturalGas, "daily")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(series.Data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(series.Data))
	}

	if series.Data[0].Value != 1.85 {
		t.Errorf("expected first value 1.85, got %f", series.Data[0].Value)
	}
}

func TestClient_GetCommodity_AllCommodities_Quarterly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != string(CommodityFunctionAllCommodities) {
			t.Errorf("expected function %s, got %s", CommodityFunctionAllCommodities, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("interval") != "quarterly" {
			t.Errorf("expected interval quarterly, got %s", r.URL.Query().Get("interval"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "All Commodities Index",
			"interval": "quarterly",
			"unit": "index",
			"data": [
				["2024-04-01", "125.50"],
				["2024-01-01", "122.30"]
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	series, err := client.GetCommodity(context.Background(), CommodityFunctionAllCommodities, "quarterly")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(series.Data) != 2 {
		t.Errorf("expected 2 data points, got %d", len(series.Data))
	}
}

func TestClient_GetCommodity_InvalidIntervalRejectedEarly(t *testing.T) {
	tests := []struct {
		name     string
		fn       CommodityFunction
		interval string
	}{
		{
			name:     "Copper with daily (invalid)",
			fn:       CommodityFunctionCopper,
			interval: "daily",
		},
		{
			name:     "All Commodities with daily (invalid)",
			fn:       CommodityFunctionAllCommodities,
			interval: "daily",
		},
		{
			name:     "WTI with hourly (invalid)",
			fn:       CommodityFunctionWTI,
			interval: "hourly",
		},
		{
			name:     "Unknown commodity function",
			fn:       "UNKNOWN",
			interval: "daily",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverCalled := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				serverCalled = true
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{"data": []}`)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			})

			_, err := client.GetCommodity(context.Background(), tt.fn, tt.interval)
			if err == nil {
				t.Fatal("expected error for invalid interval/function, got nil")
			}

			if !errors.Is(err, ErrUnsupportedInterval) && !strings.Contains(err.Error(), "unknown commodity function") {
				t.Errorf("expected ErrUnsupportedInterval or unknown function error, got: %v", err)
			}

			if serverCalled {
				t.Error("server should not have been called for invalid interval/function")
			}
		})
	}
}

func TestClient_GetCommodity_SentinelValuesSkipped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Response includes "." sentinel values that should be skipped
		_, _ = fmt.Fprint(w, `{
			"name": "Test Commodity",
			"interval": "daily",
			"unit": "test unit",
			"data": [
				["2024-04-01", "100.00"],
				["2024-03-31", "."],
				["2024-03-30", "98.50"],
				["2024-03-29", "."]
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	series, err := client.GetCommodity(context.Background(), CommodityFunctionWTI, "daily")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Should have only 2 data points (sentinel values skipped)
	if len(series.Data) != 2 {
		t.Errorf("expected 2 data points (sentinel values skipped), got %d", len(series.Data))
	}

	// Verify sentinel values are not present
	if series.Data[0].Value != 100.00 {
		t.Errorf("expected first value 100.00, got %f", series.Data[0].Value)
	}
	if series.Data[1].Value != 98.50 {
		t.Errorf("expected second value 98.50, got %f", series.Data[1].Value)
	}
}

func TestClient_GetCommodity_MalformedResponseWrappedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Malformed JSON - missing closing brace
		_, _ = fmt.Fprint(w, `{"name": "Test", "data": [[`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	_, err := client.GetCommodity(context.Background(), CommodityFunctionWTI, "daily")
	if err == nil {
		t.Fatal("expected error for malformed response, got nil")
	}

	// Error should be wrapped with context
	if !strings.Contains(err.Error(), "parsing commodity response") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestClient_GetCommodity_MissingDateValueWrappedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Data point with missing date format
		_, _ = fmt.Fprint(w, `{
			"name": "Test",
			"interval": "daily",
			"unit": "test",
			"data": [["invalid-date", "100.00"]]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	_, err := client.GetCommodity(context.Background(), CommodityFunctionWTI, "daily")
	if err == nil {
		t.Fatal("expected error for malformed date, got nil")
	}

	if !strings.Contains(err.Error(), "parsing date") {
		t.Errorf("expected wrapped date parsing error, got: %v", err)
	}
}

func TestClient_GetCommodity_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Simulate a slow request
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"data": []}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	_, err := client.GetCommodity(ctx, CommodityFunctionWTI, "daily")
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestClient_GetCommodity_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Test",
			"interval": "daily",
			"unit": "test",
			"data": []
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	series, err := client.GetCommodity(context.Background(), CommodityFunctionWTI, "daily")
	if err != nil {
		t.Fatalf("expected success for empty response, got error: %v", err)
	}

	// Empty data should return empty slice, not nil
	if series.Data == nil {
		t.Error("expected empty slice, not nil")
	}

	if len(series.Data) != 0 {
		t.Errorf("expected 0 data points, got %d", len(series.Data))
	}
}

func TestCommoditySeries_UnmarshalJSON_ParsesFloat(t *testing.T) {
	tests := []struct {
		name       string
		jsonData   string
		wantValues []float64
	}{
		{
			name: "standard floats",
			jsonData: `{
				"name": "Test",
				"interval": "daily",
				"unit": "test",
				"data": [["2024-01-01", "100.50"], ["2024-01-02", "99.75"]]
			}`,
			wantValues: []float64{100.50, 99.75},
		},
		{
			name: "integer values",
			jsonData: `{
				"name": "Test",
				"interval": "daily",
				"unit": "test",
				"data": [["2024-01-01", "100"], ["2024-01-02", "99"]]
			}`,
			wantValues: []float64{100, 99},
		},
		{
			name: "decimal values",
			jsonData: `{
				"name": "Test",
				"interval": "daily",
				"unit": "test",
				"data": [["2024-01-01", "0.85"], ["2024-01-02", "1.23"]]
			}`,
			wantValues: []float64{0.85, 1.23},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var series CommoditySeries
			if err := json.Unmarshal([]byte(tt.jsonData), &series); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if len(series.Data) != len(tt.wantValues) {
				t.Errorf("expected %d data points, got %d", len(tt.wantValues), len(series.Data))
			}

			for i, want := range tt.wantValues {
				if series.Data[i].Value != want {
					t.Errorf("data point %d: expected value %f, got %f", i, want, series.Data[i].Value)
				}
			}
		})
	}
}

func TestCommoditySeries_UnmarshalJSON_ParsesDate(t *testing.T) {
	tests := []struct {
		name      string
		jsonData  string
		wantDates []time.Time
	}{
		{
			name: "standard dates",
			jsonData: `{
				"name": "Test",
				"interval": "daily",
				"unit": "test",
				"data": [["2024-01-01", "100"], ["2024-01-02", "99"]]
			}`,
			wantDates: []time.Time{
				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "cross-year dates",
			jsonData: `{
				"name": "Test",
				"interval": "monthly",
				"unit": "test",
				"data": [["2023-12-01", "100"], ["2024-01-01", "99"]]
			}`,
			wantDates: []time.Time{
				time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var series CommoditySeries
			if err := json.Unmarshal([]byte(tt.jsonData), &series); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if len(series.Data) != len(tt.wantDates) {
				t.Errorf("expected %d data points, got %d", len(tt.wantDates), len(series.Data))
			}

			for i, want := range tt.wantDates {
				if !series.Data[i].Date.Equal(want) {
					t.Errorf("data point %d: expected date %s, got %s", i, want, series.Data[i].Date)
				}
			}
		})
	}
}

func TestClient_GetCommodity_AllFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       CommodityFunction
		interval string
	}{
		{"WTI daily", CommodityFunctionWTI, "daily"},
		{"Brent weekly", CommodityFunctionBrent, "weekly"},
		{"Natural Gas monthly", CommodityFunctionNaturalGas, "monthly"},
		{"Copper monthly", CommodityFunctionCopper, "monthly"},
		{"Aluminum quarterly", CommodityFunctionAluminum, "quarterly"},
		{"Wheat daily", CommodityFunctionWheat, "daily"},
		{"Corn weekly", CommodityFunctionCorn, "weekly"},
		{"Cotton monthly", CommodityFunctionCotton, "monthly"},
		{"Sugar quarterly", CommodityFunctionSugar, "quarterly"},
		{"Coffee annual", CommodityFunctionCoffee, "annual"},
		{"All Commodities monthly", CommodityFunctionAllCommodities, "monthly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("function") != string(tt.fn) {
					t.Errorf("expected function %s, got %s", tt.fn, r.URL.Query().Get("function"))
				}
				if r.URL.Query().Get("interval") != tt.interval {
					t.Errorf("expected interval %s, got %s", tt.interval, r.URL.Query().Get("interval"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{
					"name": "%s",
					"interval": "%s",
					"unit": "test unit",
					"data": [["2024-01-01", "100.00"]]
				}`, tt.fn, tt.interval)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			})

			series, err := client.GetCommodity(context.Background(), tt.fn, tt.interval)
			if err != nil {
				t.Fatalf("expected success for %s, got error: %v", tt.fn, err)
			}

			if len(series.Data) != 1 {
				t.Errorf("expected 1 data point, got %d", len(series.Data))
			}
		})
	}
}

func TestCommoditySeries_SortAscending(t *testing.T) {
	jsonData := `{
		"name": "Test",
		"interval": "daily",
		"unit": "test",
		"data": [
			["2024-03-01", "90"],
			["2024-02-01", "95"],
			["2024-01-01", "100"]
		]
	}`

	var series CommoditySeries
	if err := json.Unmarshal([]byte(jsonData), &series); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Initially in descending order (newest first)
	if !series.Data[0].Date.Equal(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected first date to be 2024-03-01, got %s", series.Data[0].Date)
	}

	// Sort to ascending
	series.SortAscending()

	if !series.Data[0].Date.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected first date to be 2024-01-01 after sorting, got %s", series.Data[0].Date)
	}
	if !series.Data[2].Date.Equal(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected last date to be 2024-03-01 after sorting, got %s", series.Data[2].Date)
	}
}

func TestCommoditySeries_SortAscending_NilSeries(_ *testing.T) {
	var series *CommoditySeries
	// Should not panic
	series.SortAscending()
}
