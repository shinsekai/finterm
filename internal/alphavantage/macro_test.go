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

func TestGetRealGDP_Quarterly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("function") != functionRealGDP {
			t.Errorf("expected function %s, got %s", functionRealGDP, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("interval") != "quarterly" {
			t.Errorf("expected interval quarterly, got %s", r.URL.Query().Get("interval"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Real Gross Domestic Product",
			"interval": "quarterly",
			"unit": "billions of dollars",
			"data": [
				{
					"date": "2023-10-01",
					"value": "27610.2"
				},
				{
					"date": "2023-07-01",
					"value": "27434.8"
				},
				{
					"date": "2023-04-01",
					"value": "27110.3"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetRealGDP(context.Background(), "quarterly")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(data))
	}

	// Verify data is sorted by date descending
	if data[0].Date != "2023-10-01" {
		t.Errorf("expected first date 2023-10-01, got %s", data[0].Date)
	}
	if data[1].Date != "2023-07-01" {
		t.Errorf("expected second date 2023-07-01, got %s", data[1].Date)
	}
	if data[2].Date != "2023-04-01" {
		t.Errorf("expected third date 2023-04-01, got %s", data[2].Date)
	}

	// Verify value
	if data[0].Value != "27610.2" {
		t.Errorf("expected value 27610.2, got %s", data[0].Value)
	}
}

func TestGetRealGDP_Annual(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("interval") != "annual" {
			t.Errorf("expected interval annual, got %s", r.URL.Query().Get("interval"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Real Gross Domestic Product",
			"interval": "annual",
			"unit": "billions of dollars",
			"data": [
				{
					"date": "2023-01-01",
					"value": "27610.2"
				},
				{
					"date": "2022-01-01",
					"value": "25462.7"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetRealGDP(context.Background(), "annual")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(data) != 2 {
		t.Errorf("expected 2 data points, got %d", len(data))
	}
}

func TestGetRealGDP_DefaultInterval(t *testing.T) {
	// Test that empty interval defaults to quarterly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("interval") != "quarterly" {
			t.Errorf("expected interval quarterly as default, got %s", r.URL.Query().Get("interval"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Real Gross Domestic Product",
			"interval": "quarterly",
			"unit": "billions of dollars",
			"data": [
				{
					"date": "2023-10-01",
					"value": "27610.2"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	_, err := client.GetRealGDP(context.Background(), "")
	if err != nil {
		t.Fatalf("expected success with default interval, got error: %v", err)
	}
}

func TestGetRealGDP_InvalidInterval(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetRealGDP(context.Background(), "invalid")
	if err == nil {
		t.Fatal("expected error for invalid interval, got nil")
	}

	if !strings.Contains(err.Error(), "invalid interval") {
		t.Errorf("expected invalid interval error, got: %v", err)
	}
}

func TestGetRealGDPPerCapita_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != functionRealGDPPerCapita {
			t.Errorf("expected function %s, got %s", functionRealGDPPerCapita, r.URL.Query().Get("function"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Real GDP per Capita",
			"interval": "annual",
			"unit": "dollars",
			"data": [
				{
					"date": "2023-01-01",
					"value": "82385.4"
				},
				{
					"date": "2022-01-01",
					"value": "76028.2"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetRealGDPPerCapita(context.Background())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(data) != 2 {
		t.Errorf("expected 2 data points, got %d", len(data))
	}

	if data[0].Date != "2023-01-01" {
		t.Errorf("expected first date 2023-01-01, got %s", data[0].Date)
	}
}

func TestGetCPI_Monthly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != functionCPI {
			t.Errorf("expected function %s, got %s", functionCPI, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("interval") != "monthly" {
			t.Errorf("expected interval monthly, got %s", r.URL.Query().Get("interval"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Consumer Price Index",
			"interval": "monthly",
			"unit": "index",
			"data": [
				{
					"date": "2024-03-01",
					"value": "312.332"
				},
				{
					"date": "2024-02-01",
					"value": "310.326"
				},
				{
					"date": "2024-01-01",
					"value": "308.417"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetCPI(context.Background(), "monthly")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(data))
	}

	// Verify descending order
	if data[0].Date != "2024-03-01" {
		t.Errorf("expected first date 2024-03-01, got %s", data[0].Date)
	}
	if data[2].Date != "2024-01-01" {
		t.Errorf("expected third date 2024-01-01, got %s", data[2].Date)
	}
}

func TestGetCPI_InvalidInterval(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetCPI(context.Background(), "weekly")
	if err == nil {
		t.Fatal("expected error for invalid interval, got nil")
	}

	if !strings.Contains(err.Error(), "invalid interval") {
		t.Errorf("expected invalid interval error, got: %v", err)
	}
}

func TestGetInflation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != functionInflation {
			t.Errorf("expected function %s, got %s", functionInflation, r.URL.Query().Get("function"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Inflation",
			"interval": "annual",
			"unit": "percent",
			"data": [
				{
					"date": "2023-01-01",
					"value": "3.4"
				},
				{
					"date": "2022-01-01",
					"value": "8.0"
				},
				{
					"date": "2021-01-01",
					"value": "4.7"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetInflation(context.Background())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(data))
	}

	if data[0].Value != "3.4" {
		t.Errorf("expected first value 3.4, got %s", data[0].Value)
	}
}

func TestGetFedFundsRate_Daily(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != functionFedFundsRate {
			t.Errorf("expected function %s, got %s", functionFedFundsRate, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("interval") != "daily" {
			t.Errorf("expected interval daily, got %s", r.URL.Query().Get("interval"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Federal Funds Rate",
			"interval": "daily",
			"unit": "percent",
			"data": [
				{
					"date": "2024-04-01",
					"value": "5.33"
				},
				{
					"date": "2024-03-28",
					"value": "5.33"
				},
				{
					"date": "2024-03-27",
					"value": "5.33"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetFedFundsRate(context.Background(), "daily")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(data))
	}
}

func TestGetFedFundsRate_InvalidInterval(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetFedFundsRate(context.Background(), "hourly")
	if err == nil {
		t.Fatal("expected error for invalid interval, got nil")
	}

	if !strings.Contains(err.Error(), "invalid interval") {
		t.Errorf("expected invalid interval error, got: %v", err)
	}
}

func TestGetTreasuryYield_10Year(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != functionTreasuryYield {
			t.Errorf("expected function %s, got %s", functionTreasuryYield, r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("interval") != "daily" {
			t.Errorf("expected interval daily, got %s", r.URL.Query().Get("interval"))
		}
		if r.URL.Query().Get("maturity") != "10year" {
			t.Errorf("expected maturity 10year, got %s", r.URL.Query().Get("maturity"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Treasury Yield",
			"interval": "daily",
			"maturity": "10year",
			"unit": "percent",
			"data": [
				{
					"date": "2024-04-01",
					"value": "4.32"
				},
				{
					"date": "2024-03-28",
					"value": "4.20"
				},
				{
					"date": "2024-03-27",
					"value": "4.25"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetTreasuryYield(context.Background(), "daily", "10year")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(data))
	}

	if data[0].Date != "2024-04-01" {
		t.Errorf("expected first date 2024-04-01, got %s", data[0].Date)
	}
}

func TestGetTreasuryYield_InvalidMaturity(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetTreasuryYield(context.Background(), "daily", "20year")
	if err == nil {
		t.Fatal("expected error for invalid maturity, got nil")
	}

	if !strings.Contains(err.Error(), "invalid maturity") {
		t.Errorf("expected invalid maturity error, got: %v", err)
	}
}

func TestGetTreasuryYield_InvalidInterval(t *testing.T) {
	client := New(Config{
		Key: "test-key",
	})

	_, err := client.GetTreasuryYield(context.Background(), "hourly", "10year")
	if err == nil {
		t.Fatal("expected error for invalid interval, got nil")
	}

	if !strings.Contains(err.Error(), "invalid interval") {
		t.Errorf("expected invalid interval error, got: %v", err)
	}
}

func TestGetTreasuryYield_AllMaturities(t *testing.T) {
	tests := []struct {
		name     string
		maturity string
	}{
		{"2 year", "2year"},
		{"5 year", "5year"},
		{"10 year", "10year"},
		{"30 year", "30year"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("maturity") != tt.maturity {
					t.Errorf("expected maturity %s, got %s", tt.maturity, r.URL.Query().Get("maturity"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{
					"name": "Treasury Yield",
					"interval": "daily",
					"maturity": "`+tt.maturity+`",
					"unit": "percent",
					"data": [
						{
							"date": "2024-04-01",
							"value": "4.32"
						}
					]
				}`)
			}))
			defer server.Close()

			client := New(Config{
				Key:        "test-key",
				BaseURL:    server.URL,
				HTTPClient: &http.Client{},
			})

			_, err := client.GetTreasuryYield(context.Background(), "daily", tt.maturity)
			if err != nil {
				t.Fatalf("expected success for maturity %s, got error: %v", tt.maturity, err)
			}
		})
	}
}

func TestGetTreasuryYield_DefaultMaturity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default should be 10year
		if r.URL.Query().Get("maturity") != "10year" {
			t.Errorf("expected maturity 10year as default, got %s", r.URL.Query().Get("maturity"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Treasury Yield",
			"interval": "daily",
			"maturity": "10year",
			"unit": "percent",
			"data": [
				{
					"date": "2024-04-01",
					"value": "4.32"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	_, err := client.GetTreasuryYield(context.Background(), "daily", "")
	if err != nil {
		t.Fatalf("expected success with default maturity, got error: %v", err)
	}
}

func TestGetUnemployment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != functionUnemployment {
			t.Errorf("expected function %s, got %s", functionUnemployment, r.URL.Query().Get("function"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Unemployment Rate",
			"interval": "monthly",
			"unit": "percent",
			"data": [
				{
					"date": "2024-03-01",
					"value": "3.8"
				},
				{
					"date": "2024-02-01",
					"value": "3.9"
				},
				{
					"date": "2024-01-01",
					"value": "3.7"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetUnemployment(context.Background())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(data))
	}

	if data[0].Value != "3.8" {
		t.Errorf("expected first value 3.8, got %s", data[0].Value)
	}
}

func TestGetNonfarmPayroll_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("function") != functionNonfarmPayroll {
			t.Errorf("expected function %s, got %s", functionNonfarmPayroll, r.URL.Query().Get("function"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{
			"name": "Nonfarm Payroll",
			"interval": "monthly",
			"unit": "thousands",
			"data": [
				{
					"date": "2024-03-01",
					"value": "303"
				},
				{
					"date": "2024-02-01",
					"value": "270"
				},
				{
					"date": "2024-01-01",
					"value": "229"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetNonfarmPayroll(context.Background())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(data) != 3 {
		t.Errorf("expected 3 data points, got %d", len(data))
	}

	if data[0].Value != "303" {
		t.Errorf("expected first value 303, got %s", data[0].Value)
	}
}

func TestMacro_EmptyResponse(t *testing.T) {
	tests := []struct {
		name       string
		httpMethod func() ([]MacroDataPoint, error)
	}{
		{
			name: "Real GDP",
			httpMethod: func() ([]MacroDataPoint, error) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = fmt.Fprint(w, `{
						"name": "Real Gross Domestic Product",
						"interval": "quarterly",
						"unit": "billions of dollars",
						"data": []
					}`)
				}))
				defer server.Close()

				client := New(Config{
					Key:        "test-key",
					BaseURL:    server.URL,
					HTTPClient: &http.Client{},
				})

				return client.GetRealGDP(context.Background(), "quarterly")
			},
		},
		{
			name: "CPI",
			httpMethod: func() ([]MacroDataPoint, error) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = fmt.Fprint(w, `{
						"name": "Consumer Price Index",
						"interval": "monthly",
						"unit": "index",
						"data": []
					}`)
				}))
				defer server.Close()

				client := New(Config{
					Key:        "test-key",
					BaseURL:    server.URL,
					HTTPClient: &http.Client{},
				})

				return client.GetCPI(context.Background(), "monthly")
			},
		},
		{
			name: "Unemployment",
			httpMethod: func() ([]MacroDataPoint, error) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = fmt.Fprint(w, `{
						"name": "Unemployment Rate",
						"interval": "monthly",
						"unit": "percent",
						"data": []
					}`)
				}))
				defer server.Close()

				client := New(Config{
					Key:        "test-key",
					BaseURL:    server.URL,
					HTTPClient: &http.Client{},
				})

				return client.GetUnemployment(context.Background())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.httpMethod()
			if err != nil {
				t.Fatalf("expected success for empty response, got error: %v", err)
			}

			// Empty data should return empty slice, not nil
			if data == nil {
				t.Error("expected empty slice, not nil")
			}

			if len(data) != 0 {
				t.Errorf("expected 0 data points, got %d", len(data))
			}
		})
	}
}

func TestMacro_ContextPropagation(t *testing.T) {
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

	_, err := client.GetRealGDP(ctx, "quarterly")
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestMacro_SortingDescending(t *testing.T) {
	// Test that data is sorted by date descending
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Send data in random order
		_, _ = fmt.Fprint(w, `{
			"name": "Test",
			"interval": "monthly",
			"unit": "test",
			"data": [
				{
					"date": "2023-01-01",
					"value": "1"
				},
				{
					"date": "2023-12-01",
					"value": "12"
				},
				{
					"date": "2023-06-01",
					"value": "6"
				}
			]
		}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	data, err := client.GetInflation(context.Background())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Should be sorted descending
	if data[0].Date != "2023-12-01" {
		t.Errorf("expected first date 2023-12-01 (most recent), got %s", data[0].Date)
	}
	if data[1].Date != "2023-06-01" {
		t.Errorf("expected second date 2023-06-01, got %s", data[1].Date)
	}
	if data[2].Date != "2023-01-01" {
		t.Errorf("expected third date 2023-01-01 (oldest), got %s", data[2].Date)
	}
}
