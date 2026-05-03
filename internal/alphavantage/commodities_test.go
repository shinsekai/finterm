package alphavantage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
				{"date": "2024-04-01", "value": "82.50"},
				{"date": "2024-03-31", "value": "83.12"},
				{"date": "2024-03-30", "value": "82.75"}
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
				{"date": "2024-04-01", "value": "87.25"},
				{"date": "2024-03-31", "value": "86.80"}
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
				{"date": "2024-04-01", "value": "9125.50"},
				{"date": "2024-03-01", "value": "8890.25"},
				{"date": "2024-02-01", "value": "8565.00"}
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
				{"date": "2024-04-01", "value": "1.85"},
				{"date": "2024-03-31", "value": "1.78"},
				{"date": "2024-03-30", "value": "1.82"}
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
				{"date": "2024-04-01", "value": "125.50"},
				{"date": "2024-01-01", "value": "122.30"}
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
				{"date": "2024-04-01", "value": "100.00"},
				{"date": "2024-03-31", "value": "."},
				{"date": "2024-03-30", "value": "98.50"},
				{"date": "2024-03-29", "value": "."}
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
		_, _ = fmt.Fprint(w, `{"name": "Test", "data": [{"date": "2024-01-01", "value": "100"}`)
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
			"data": [{"date": "invalid-date", "value": "100.00"}]
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
				"data": [{"date": "2024-01-01", "value": "100.50"}, {"date": "2024-01-02", "value": "99.75"}]
			}`,
			wantValues: []float64{100.50, 99.75},
		},
		{
			name: "integer values",
			jsonData: `{
				"name": "Test",
				"interval": "daily",
				"unit": "test",
				"data": [{"date": "2024-01-01", "value": "100"}, {"date": "2024-01-02", "value": "99"}]
			}`,
			wantValues: []float64{100, 99},
		},
		{
			name: "decimal values",
			jsonData: `{
				"name": "Test",
				"interval": "daily",
				"unit": "test",
				"data": [{"date": "2024-01-01", "value": "0.85"}, {"date": "2024-01-02", "value": "1.23"}]
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
				"data": [{"date": "2024-01-01", "value": "100"}, {"date": "2024-01-02", "value": "99"}]
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
				"data": [{"date": "2023-12-01", "value": "100"}, {"date": "2024-01-01", "value": "99"}]
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
		{"Wheat monthly", CommodityFunctionWheat, "monthly"},
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
					"data": [{"date": "2024-01-01", "value": "100.00"}]
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
			{"date": "2024-03-01", "value": "90"},
			{"date": "2024-02-01", "value": "95"},
			{"date": "2024-01-01", "value": "100"}
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

// TestCommoditySeries_UnmarshalRealShapeWTI tests that we can unmarshal a real WTI daily response.
func TestCommoditySeries_UnmarshalRealShapeWTI(t *testing.T) {
	data, err := os.ReadFile("testdata/commodity_wti_daily.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var series CommoditySeries
	if err := json.Unmarshal(data, &series); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	if series.Name != "Crude Oil Prices WTI" {
		t.Errorf("expected name 'Crude Oil Prices WTI', got %s", series.Name)
	}
	if series.Interval != "daily" {
		t.Errorf("expected interval 'daily', got %s", series.Interval)
	}
	if len(series.Data) == 0 {
		t.Error("expected at least one data point")
	}

	// Verify the first data point has correct date
	firstDate := time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)
	if !series.Data[0].Date.Equal(firstDate) {
		t.Errorf("expected first date %s, got %s", firstDate, series.Data[0].Date)
	}
}

// TestCommoditySeries_UnmarshalRealShapeBrent tests that we can unmarshal a real Brent daily response.
func TestCommoditySeries_UnmarshalRealShapeBrent(t *testing.T) {
	data, err := os.ReadFile("testdata/commodity_brent_daily.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var series CommoditySeries
	if err := json.Unmarshal(data, &series); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	if len(series.Data) == 0 {
		t.Error("expected at least one data point")
	}
}

// TestCommoditySeries_UnmarshalRealShapeMonthly tests that we can unmarshal a real monthly response.
func TestCommoditySeries_UnmarshalRealShapeMonthly(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
	}{
		{"Copper monthly", "testdata/commodity_copper_monthly.json"},
		{"Wheat monthly", "testdata/commodity_wheat_monthly.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.fixture)
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}

			var series CommoditySeries
			if err := json.Unmarshal(data, &series); err != nil {
				t.Fatalf("failed to unmarshal fixture: %v", err)
			}

			if series.Interval != "monthly" {
				t.Errorf("expected interval 'monthly', got %s", series.Interval)
			}
			if len(series.Data) == 0 {
				t.Error("expected at least one data point")
			}
		})
	}
}

// TestCommoditySeries_SkipsSentinelValues tests that "." sentinel values are properly skipped.
func TestCommoditySeries_SkipsSentinelValues(t *testing.T) {
	jsonData := `{
		"name": "Test",
		"interval": "daily",
		"unit": "test",
		"data": [
			{"date": "2024-04-01", "value": "100.00"},
			{"date": "2024-03-31", "value": "."},
			{"date": "2024-03-30", "value": "98.50"}
		]
	}`

	var series CommoditySeries
	if err := json.Unmarshal([]byte(jsonData), &series); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(series.Data) != 2 {
		t.Errorf("expected 2 data points (sentinel skipped), got %d", len(series.Data))
	}
	if series.Data[0].Value != 100.00 {
		t.Errorf("expected first value 100.00, got %f", series.Data[0].Value)
	}
	if series.Data[1].Value != 98.50 {
		t.Errorf("expected second value 98.50, got %f", series.Data[1].Value)
	}
}

// TestCommoditySeries_RejectsArrayShapeWithClearError tests that the old array-shaped data
// is rejected with a clear error message indicating the API may have changed.
func TestCommoditySeries_RejectsArrayShapeWithClearError(t *testing.T) {
	// Old-style array format that should fail
	jsonData := `{
		"name": "Test",
		"interval": "daily",
		"unit": "test",
		"data": [["2024-04-01", "100.00"], ["2024-03-31", "99.50"]]
	}`

	var series CommoditySeries
	err := json.Unmarshal([]byte(jsonData), &series)
	if err == nil {
		t.Fatal("expected error for array-shaped data, got nil")
	}

	// Check that the error message hints at the API shape change
	if !strings.Contains(err.Error(), "unexpected commodity data point shape") {
		t.Errorf("expected error message about unexpected shape, got: %v", err)
	}
}

// TestCommoditySeries_PreservesNewestFirstOrdering tests that data is kept in
// the order returned by the API (newest first).
func TestCommoditySeries_PreservesNewestFirstOrdering(t *testing.T) {
	jsonData := `{
		"name": "Test",
		"interval": "daily",
		"unit": "test",
		"data": [
			{"date": "2024-04-05", "value": "105"},
			{"date": "2024-04-04", "value": "104"},
			{"date": "2024-04-03", "value": "103"},
			{"date": "2024-04-02", "value": "102"},
			{"date": "2024-04-01", "value": "101"}
		]
	}`

	var series CommoditySeries
	if err := json.Unmarshal([]byte(jsonData), &series); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify order is preserved (newest first)
	if !series.Data[0].Date.Equal(time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected first date 2024-04-05, got %s", series.Data[0].Date)
	}
	if !series.Data[4].Date.Equal(time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected last date 2024-04-01, got %s", series.Data[4].Date)
	}
}

// TestCommoditySeries_FixtureShapeRegression tests that fixture files contain
// the correct object-shaped data, not the old array-shaped data.
func TestCommoditySeries_FixtureShapeRegression(t *testing.T) {
	fixtures := []string{
		"testdata/commodity_wti_daily.json",
		"testdata/commodity_brent_daily.json",
		"testdata/commodity_natural_gas_daily.json",
		"testdata/commodity_copper_monthly.json",
		"testdata/commodity_wheat_monthly.json",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			data, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}

			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("failed to parse fixture as raw JSON: %v", err)
			}

			dataField, ok := raw["data"]
			if !ok {
				t.Fatal("fixture missing 'data' field")
			}

			dataArray, ok := dataField.([]interface{})
			if !ok {
				t.Fatal("'data' field is not an array")
			}

			if len(dataArray) == 0 {
				t.Fatal("fixture has empty data array")
			}

			// Check the first data point is an object, not an array
			first, ok := dataArray[0].(map[string]interface{})
			if !ok {
				t.Fatalf("fixture data point %d is not an object (it's %T), has the API shape changed?", 0, dataArray[0])
			}

			// Verify it has the expected keys
			if _, ok := first["date"]; !ok {
				t.Error("data point missing 'date' key")
			}
			if _, ok := first["value"]; !ok {
				t.Error("data point missing 'value' key")
			}
		})
	}
}

// TestGetCommodity_LiveAPI tests against the live Alpha Vantage API.
// This test is gated by the RUN_LIVE_TESTS environment variable.
func TestGetCommodity_LiveAPI(t *testing.T) {
	if os.Getenv("RUN_LIVE_TESTS") != "1" {
		t.Skip("skipping live API test (set RUN_LIVE_TESTS=1 to run)")
	}

	apiKey := os.Getenv("FINTERM_AV_API_KEY")
	if apiKey == "" {
		t.Skip("skipping live API test (FINTERM_AV_API_KEY not set)")
	}

	client := New(Config{
		Key:        apiKey,
		BaseURL:    "https://www.alphavantage.co/query",
		HTTPClient: &http.Client{},
	})

	series, err := client.GetCommodity(context.Background(), CommodityFunctionWTI, "daily")
	if err != nil {
		t.Fatalf("live API call failed: %v", err)
	}

	if len(series.Data) == 0 {
		t.Error("live API returned empty data")
	}

	// Verify we got a reasonable response
	if series.Name == "" {
		t.Error("live API returned empty name")
	}
	if series.Unit == "" {
		t.Error("live API returned empty unit")
	}

	t.Logf("Live API test passed: got %d data points from %s", len(series.Data), series.Name)
}

// TestCommodityFunctionFromSymbol_ExhaustiveMapping tests that all commodity symbols
// can be mapped to their CommodityFunction.
func TestCommodityFunctionFromSymbol_ExhaustiveMapping(t *testing.T) {
	tests := []struct {
		name   string
		symbol string
		want   CommodityFunction
	}{
		{"WTI", "WTI", CommodityFunctionWTI},
		{"BRENT", "BRENT", CommodityFunctionBrent},
		{"NATURAL_GAS", "NATURAL_GAS", CommodityFunctionNaturalGas},
		{"COPPER", "COPPER", CommodityFunctionCopper},
		{"ALUMINUM", "ALUMINUM", CommodityFunctionAluminum},
		{"WHEAT", "WHEAT", CommodityFunctionWheat},
		{"CORN", "CORN", CommodityFunctionCorn},
		{"COTTON", "COTTON", CommodityFunctionCotton},
		{"SUGAR", "SUGAR", CommodityFunctionSugar},
		{"COFFEE", "COFFEE", CommodityFunctionCoffee},
		{"ALL_COMMODITIES", "ALL_COMMODITIES", CommodityFunctionAllCommodities},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := CommodityFunctionFromSymbol(tt.symbol)
			if !ok {
				t.Errorf("CommodityFunctionFromSymbol(%q) returned false, want true", tt.symbol)
			}
			if got != tt.want {
				t.Errorf("CommodityFunctionFromSymbol(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

// TestCommodityFunctionFromSymbol_CaseInsensitive tests that the lookup is
// case-insensitive.
func TestCommodityFunctionFromSymbol_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name   string
		symbol string
		want   CommodityFunction
	}{
		{"WTI lowercase", "wti", CommodityFunctionWTI},
		{"WTI mixed case", "Wti", CommodityFunctionWTI},
		{"BRENT lowercase", "brent", CommodityFunctionBrent},
		{"BRENT mixed case", "BrEnT", CommodityFunctionBrent},
		{"WHEAT lowercase", "wheat", CommodityFunctionWheat},
		{"WHEAT mixed case", "WhEaT", CommodityFunctionWheat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := CommodityFunctionFromSymbol(tt.symbol)
			if !ok {
				t.Errorf("CommodityFunctionFromSymbol(%q) returned false, want true", tt.symbol)
			}
			if got != tt.want {
				t.Errorf("CommodityFunctionFromSymbol(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

// TestCommodityFunctionFromSymbol_UnknownReturnsFalse tests that unknown symbols
// return false as the second return value.
func TestCommodityFunctionFromSymbol_UnknownReturnsFalse(t *testing.T) {
	tests := []struct {
		name   string
		symbol string
	}{
		{"UNKNOWN", "UNKNOWN"},
		{"GOLD", "GOLD"},
		{"SILVER", "SILVER"},
		{"PLATINUM", "PLATINUM"},
		{"empty", ""},
		{"random", "RANDOM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := CommodityFunctionFromSymbol(tt.symbol)
			if ok {
				t.Errorf("CommodityFunctionFromSymbol(%q) returned true, want false", tt.symbol)
			}
		})
	}
}

// TestCommoditySupportedIntervals tests that the supported intervals
// are correctly returned for each commodity function.
func TestCommoditySupportedIntervals(t *testing.T) {
	tests := []struct {
		name          string
		fn            CommodityFunction
		wantIntervals []string
		wantOK        bool
	}{
		{
			"WTI",
			CommodityFunctionWTI,
			[]string{"daily", "weekly", "monthly", "quarterly"},
			true,
		},
		{
			"BRENT",
			CommodityFunctionBrent,
			[]string{"daily", "weekly", "monthly", "quarterly"},
			true,
		},
		{
			"COPPER",
			CommodityFunctionCopper,
			[]string{"monthly", "quarterly", "annual"},
			true,
		},
		{
			"ALUMINUM",
			CommodityFunctionAluminum,
			[]string{"monthly", "quarterly", "annual"},
			true,
		},
		{
			"WHEAT",
			CommodityFunctionWheat,
			[]string{"daily", "weekly", "monthly", "quarterly", "annual"},
			true,
		},
		{
			"UNKNOWN",
			"UNKNOWN",
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := CommoditySupportedIntervals(tt.fn)
			if ok != tt.wantOK {
				t.Errorf("CommoditySupportedIntervals(%q) ok = %v, want %v", tt.fn, ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if len(got) != len(tt.wantIntervals) {
				t.Errorf("CommoditySupportedIntervals(%q) returned %d intervals, want %d", tt.fn, len(got), len(tt.wantIntervals))
			}
			for i, want := range tt.wantIntervals {
				if got[i] != want {
					t.Errorf("CommoditySupportedIntervals(%q)[%d] = %q, want %q", tt.fn, i, got[i], want)
				}
			}
		})
	}
}

// TestBestSupportedInterval_ExactMatch tests that exact matches are returned unchanged.
func TestBestSupportedInterval_ExactMatch(t *testing.T) {
	tests := []struct {
		name      string
		fn        CommodityFunction
		requested string
		want      string
	}{
		{
			name:      "WTI daily - exact match",
			fn:        CommodityFunctionWTI,
			requested: "daily",
			want:      "daily",
		},
		{
			name:      "WTI weekly - exact match",
			fn:        CommodityFunctionWTI,
			requested: "weekly",
			want:      "weekly",
		},
		{
			name:      "WTI monthly - exact match",
			fn:        CommodityFunctionWTI,
			requested: "monthly",
			want:      "monthly",
		},
		{
			name:      "WTI quarterly - exact match",
			fn:        CommodityFunctionWTI,
			requested: "quarterly",
			want:      "quarterly",
		},
		{
			name:      "Copper monthly - exact match",
			fn:        CommodityFunctionCopper,
			requested: "monthly",
			want:      "monthly",
		},
		{
			name:      "Copper quarterly - exact match",
			fn:        CommodityFunctionCopper,
			requested: "quarterly",
			want:      "quarterly",
		},
		{
			name:      "Copper annual - exact match",
			fn:        CommodityFunctionCopper,
			requested: "annual",
			want:      "annual",
		},
		{
			name:      "AllCommodities annual - exact match",
			fn:        CommodityFunctionAllCommodities,
			requested: "annual",
			want:      "annual",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BestSupportedInterval(tt.fn, tt.requested)
			if got != tt.want {
				t.Errorf("BestSupportedInterval(%v, %q) = %q, want %q", tt.fn, tt.requested, got, tt.want)
			}
		})
	}
}

// TestBestSupportedInterval_FallbackToCoarser tests fallback to next coarser interval.
func TestBestSupportedInterval_FallbackToCoarser(t *testing.T) {
	tests := []struct {
		name      string
		fn        CommodityFunction
		requested string
		want      string
	}{
		{
			name:      "Copper daily to monthly",
			fn:        CommodityFunctionCopper,
			requested: "daily",
			want:      "monthly",
		},
		{
			name:      "Copper weekly to monthly",
			fn:        CommodityFunctionCopper,
			requested: "weekly",
			want:      "monthly",
		},
		{
			name:      "Aluminum daily to monthly",
			fn:        CommodityFunctionAluminum,
			requested: "daily",
			want:      "monthly",
		},
		{
			name:      "Aluminum weekly to monthly",
			fn:        CommodityFunctionAluminum,
			requested: "weekly",
			want:      "monthly",
		},
		{
			name:      "AllCommodities daily to monthly",
			fn:        CommodityFunctionAllCommodities,
			requested: "daily",
			want:      "monthly",
		},
		{
			name:      "AllCommodities weekly to monthly",
			fn:        CommodityFunctionAllCommodities,
			requested: "weekly",
			want:      "monthly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BestSupportedInterval(tt.fn, tt.requested)
			if got != tt.want {
				t.Errorf("BestSupportedInterval(%v, %q) = %q, want %q", tt.fn, tt.requested, got, tt.want)
			}
		})
	}
}

// TestBestSupportedInterval_AllSupportedIntervalsPerFunction tests all intervals per function.
func TestBestSupportedInterval_AllSupportedIntervalsPerFunction(t *testing.T) {
	tests := []struct {
		name          string
		fn            CommodityFunction
		requested     string
		expected      string
		explainReason string
	}{
		// WTI - supports daily, weekly, monthly, quarterly
		{
			name:          "WTI: exact match - daily",
			fn:            CommodityFunctionWTI,
			requested:     "daily",
			expected:      "daily",
			explainReason: "WTI supports daily, so exact match",
		},
		{
			name:          "WTI: exact match - weekly",
			fn:            CommodityFunctionWTI,
			requested:     "weekly",
			expected:      "weekly",
			explainReason: "WTI supports weekly, so exact match",
		},
		{
			name:          "WTI: exact match - monthly",
			fn:            CommodityFunctionWTI,
			requested:     "monthly",
			expected:      "monthly",
			explainReason: "WTI supports monthly, so exact match",
		},
		{
			name:          "WTI: exact match - quarterly",
			fn:            CommodityFunctionWTI,
			requested:     "quarterly",
			expected:      "quarterly",
			explainReason: "WTI supports quarterly, so exact match",
		},
		{
			name:          "WTI: requested annual, fallback to quarterly",
			fn:            CommodityFunctionWTI,
			requested:     "annual",
			expected:      "quarterly",
			explainReason: "WTI doesn't support annual, falls back to quarterly (next coarser)",
		},
		// Copper - supports monthly, quarterly, annual
		{
			name:          "Copper: daily falls back to monthly",
			fn:            CommodityFunctionCopper,
			requested:     "daily",
			expected:      "monthly",
			explainReason: "Copper doesn't support daily or weekly, falls back to monthly",
		},
		{
			name:          "Copper: weekly falls back to monthly",
			fn:            CommodityFunctionCopper,
			requested:     "weekly",
			expected:      "monthly",
			explainReason: "Copper doesn't support daily or weekly, falls back to monthly",
		},
		{
			name:          "Copper: exact match - monthly",
			fn:            CommodityFunctionCopper,
			requested:     "monthly",
			expected:      "monthly",
			explainReason: "Copper supports monthly, so exact match",
		},
		{
			name:          "Copper: exact match - quarterly",
			fn:            CommodityFunctionCopper,
			requested:     "quarterly",
			expected:      "quarterly",
			explainReason: "Copper supports quarterly, so exact match",
		},
		{
			name:          "Copper: exact match - annual",
			fn:            CommodityFunctionCopper,
			requested:     "annual",
			expected:      "annual",
			explainReason: "Copper supports annual, so exact match",
		},
		// Aluminum - supports monthly, quarterly, annual (same as Copper)
		{
			name:          "Aluminum: daily falls back to monthly",
			fn:            CommodityFunctionAluminum,
			requested:     "daily",
			expected:      "monthly",
			explainReason: "Aluminum doesn't support daily or weekly, falls back to monthly",
		},
		{
			name:          "Aluminum: weekly falls back to monthly",
			fn:            CommodityFunctionAluminum,
			requested:     "weekly",
			expected:      "monthly",
			explainReason: "Aluminum doesn't support daily or weekly, falls back to monthly",
		},
		{
			name:          "Aluminum: exact match - monthly",
			fn:            CommodityFunctionAluminum,
			requested:     "monthly",
			expected:      "monthly",
			explainReason: "Aluminum supports monthly, so exact match",
		},
		{
			name:          "Aluminum: exact match - quarterly",
			fn:            CommodityFunctionAluminum,
			requested:     "quarterly",
			expected:      "quarterly",
			explainReason: "Aluminum supports quarterly, so exact match",
		},
		{
			name:          "Aluminum: exact match - annual",
			fn:            CommodityFunctionAluminum,
			requested:     "annual",
			expected:      "annual",
			explainReason: "Aluminum supports annual, so exact match",
		},
		// Wheat - supports daily, weekly, monthly, quarterly, annual
		{
			name:          "Wheat: exact match - daily",
			fn:            CommodityFunctionWheat,
			requested:     "daily",
			expected:      "daily",
			explainReason: "Wheat supports daily, so exact match",
		},
		{
			name:          "Wheat: exact match - weekly",
			fn:            CommodityFunctionWheat,
			requested:     "weekly",
			expected:      "weekly",
			explainReason: "Wheat supports weekly, so exact match",
		},
		{
			name:          "Wheat: exact match - monthly",
			fn:            CommodityFunctionWheat,
			requested:     "monthly",
			expected:      "monthly",
			explainReason: "Wheat supports monthly, so exact match",
		},
		{
			name:          "Wheat: exact match - quarterly",
			fn:            CommodityFunctionWheat,
			requested:     "quarterly",
			expected:      "quarterly",
			explainReason: "Wheat supports quarterly, so exact match",
		},
		{
			name:          "Wheat: exact match - annual",
			fn:            CommodityFunctionWheat,
			requested:     "annual",
			expected:      "annual",
			explainReason: "Wheat supports annual, so exact match",
		},
		// AllCommodities - supports monthly, quarterly, annual
		{
			name:          "AllCommodities: daily falls back to monthly",
			fn:            CommodityFunctionAllCommodities,
			requested:     "daily",
			expected:      "monthly",
			explainReason: "AllCommodities doesn't support daily or weekly, falls back to monthly",
		},
		{
			name:          "AllCommodities: weekly falls back to monthly",
			fn:            CommodityFunctionAllCommodities,
			requested:     "weekly",
			expected:      "monthly",
			explainReason: "AllCommodities doesn't support daily or weekly, falls back to monthly",
		},
		{
			name:          "AllCommodities: exact match - monthly",
			fn:            CommodityFunctionAllCommodities,
			requested:     "monthly",
			expected:      "monthly",
			explainReason: "AllCommodities supports monthly, so exact match",
		},
		{
			name:          "AllCommodities: exact match - quarterly",
			fn:            CommodityFunctionAllCommodities,
			requested:     "quarterly",
			expected:      "quarterly",
			explainReason: "AllCommodities supports quarterly, so exact match",
		},
		{
			name:          "AllCommodities: exact match - annual",
			fn:            CommodityFunctionAllCommodities,
			requested:     "annual",
			expected:      "annual",
			explainReason: "AllCommodities supports annual, so exact match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BestSupportedInterval(tt.fn, tt.requested)
			if got != tt.expected {
				t.Errorf("BestSupportedInterval(%v, %q) = %q, want %q (%s)",
					tt.fn, tt.requested, got, tt.expected, tt.explainReason)
			}
		})
	}
}

// TestBestSupportedInterval_UnknownFunction tests that unknown functions return empty string.
func TestBestSupportedInterval_UnknownFunction(t *testing.T) {
	got := BestSupportedInterval(CommodityFunction("UNKNOWN"), "daily")
	if got != "" {
		t.Errorf("BestSupportedInterval(UNKNOWN, daily) = %q, want empty string", got)
	}
}

// TestBestSupportedInterval_InvalidInterval tests that invalid intervals return the coarsest supported.
func TestBestSupportedInterval_InvalidInterval(t *testing.T) {
	// For truly invalid intervals (not in precedence), return the coarsest supported interval
	got := BestSupportedInterval(CommodityFunctionWTI, "invalid")
	expected := "quarterly" // Coarsest supported for WTI
	if got != expected {
		t.Errorf("BestSupportedInterval(WTI, invalid) = %q, want %q (coarsest supported)", got, expected)
	}
}
