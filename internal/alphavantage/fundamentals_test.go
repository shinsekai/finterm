package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClient_GetCompanyOverview_AAPL(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "overview_aapl.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	overview, err := client.GetCompanyOverview(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Verify all required fields are populated
	if overview.Symbol != "AAPL" {
		t.Errorf("expected Symbol AAPL, got %s", overview.Symbol)
	}
	if overview.Name != "Apple Inc." {
		t.Errorf("expected Name Apple Inc., got %s", overview.Name)
	}
	if overview.Exchange != "NASDAQ" {
		t.Errorf("expected Exchange NASDAQ, got %s", overview.Exchange)
	}
	if overview.Currency != "USD" {
		t.Errorf("expected Currency USD, got %s", overview.Currency)
	}
	if overview.Country != "USA" {
		t.Errorf("expected Country USA, got %s", overview.Country)
	}
	if overview.Sector != "Technology" {
		t.Errorf("expected Sector Technology, got %s", overview.Sector)
	}
	if overview.Industry != "Consumer Electronics" {
		t.Errorf("expected Industry Consumer Electronics, got %s", overview.Industry)
	}
	if overview.FiscalYearEnd != "September" {
		t.Errorf("expected FiscalYearEnd September, got %s", overview.FiscalYearEnd)
	}

	// Verify numeric field values
	expectedMC := 2848000000000.0
	if got := overview.MarketCapitalizationValue(); got != expectedMC {
		t.Errorf("expected MarketCapitalizationValue %f, got %f", expectedMC, got)
	}
	expectedPE := 29.42
	if got := overview.PERatioValue(); got != expectedPE {
		t.Errorf("expected PERatioValue %f, got %f", expectedPE, got)
	}
	expectedPEG := 2.68
	if got := overview.PEGRatioValue(); got != expectedPEG {
		t.Errorf("expected PEGRatioValue %f, got %f", expectedPEG, got)
	}
	expectedBV := 4.19
	if got := overview.BookValueValue(); got != expectedBV {
		t.Errorf("expected BookValueValue %f, got %f", expectedBV, got)
	}
	expectedDPS := 0.96
	if got := overview.DividendPerShareValue(); got != expectedDPS {
		t.Errorf("expected DividendPerShareValue %f, got %f", expectedDPS, got)
	}
	expectedDY := 0.56
	if got := overview.DividendYieldValue(); got != expectedDY {
		t.Errorf("expected DividendYieldValue %f, got %f", expectedDY, got)
	}
	expectedEPS := 6.16
	if got := overview.EPSValue(); got != expectedEPS {
		t.Errorf("expected EPSValue %f, got %f", expectedEPS, got)
	}
	expectedBeta := 1.28
	if got := overview.BetaValue(); got != expectedBeta {
		t.Errorf("expected BetaValue %f, got %f", expectedBeta, got)
	}
	expectedHigh := 199.62
	if got := overview.FiftyTwoWeekHighValue(); got != expectedHigh {
		t.Errorf("expected FiftyTwoWeekHighValue %f, got %f", expectedHigh, got)
	}
	expectedLow := 124.17
	if got := overview.FiftyTwoWeekLowValue(); got != expectedLow {
		t.Errorf("expected FiftyTwoWeekLowValue %f, got %f", expectedLow, got)
	}
	expectedShares := 15555300000.0
	if got := overview.SharesOutstandingValue(); got != expectedShares {
		t.Errorf("expected SharesOutstandingValue %f, got %f", expectedShares, got)
	}
}

func TestClient_GetCompanyOverview_MissingNumericFieldsGraceful(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "overview_msft_missing_fields.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	overview, err := client.GetCompanyOverview(context.Background(), "MSFT")
	if err != nil {
		t.Fatalf("expected success with missing fields, got error: %v", err)
	}

	// Verify fields with "None" return 0 via value methods
	if got := overview.MarketCapitalizationValue(); got != 0 {
		t.Errorf("expected MarketCapitalizationValue 0 for None, got %f", got)
	}
	if got := overview.PEGRatioValue(); got != 0 {
		t.Errorf("expected PEGRatioValue 0 for None, got %f", got)
	}
	if got := overview.DividendPerShareValue(); got != 0 {
		t.Errorf("expected DividendPerShareValue 0 for None, got %f", got)
	}
	if got := overview.BetaValue(); got != 0 {
		t.Errorf("expected BetaValue 0 for None, got %f", got)
	}
	if got := overview.SharesOutstandingValue(); got != 0 {
		t.Errorf("expected SharesOutstandingValue 0 for None, got %f", got)
	}

	// Verify present numeric fields still work
	expectedPE := 36.82
	if got := overview.PERatioValue(); got != expectedPE {
		t.Errorf("expected PERatioValue %f, got %f", expectedPE, got)
	}
	expectedDY := 0.75
	if got := overview.DividendYieldValue(); got != expectedDY {
		t.Errorf("expected DividendYieldValue %f, got %f", expectedDY, got)
	}
}

func TestClient_GetCompanyOverview_EmptySymbolRejected(t *testing.T) {
	client := New(Config{
		Key:        "test-key",
		BaseURL:    "https://example.com",
		HTTPClient: &http.Client{},
	})

	_, err := client.GetCompanyOverview(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected error to mention empty symbol, got: %v", err)
	}
}

func TestClient_GetCompanyOverview_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"Symbol": "AAPL"}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetCompanyOverview(ctx, "AAPL")
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestClient_GetEarnings_AAPL(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "earnings_aapl.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	earnings, err := client.GetEarnings(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if earnings.Symbol != "AAPL" {
		t.Errorf("expected Symbol AAPL, got %s", earnings.Symbol)
	}

	// Verify annual earnings
	if len(earnings.Annual) != 4 {
		t.Errorf("expected 4 annual earnings entries, got %d", len(earnings.Annual))
	}

	firstAnnual := earnings.Annual[0]
	if firstAnnual.FiscalDateEnding != "2023-09-30" {
		t.Errorf("expected fiscal date 2023-09-30, got %s", firstAnnual.FiscalDateEnding)
	}
	expectedEPS := 6.16
	if got := firstAnnual.ReportedEPSValue(); got != expectedEPS {
		t.Errorf("expected ReportedEPSValue %f, got %f", expectedEPS, got)
	}

	// Verify quarterly earnings
	if len(earnings.Quarterly) != 4 {
		t.Errorf("expected 4 quarterly earnings entries, got %d", len(earnings.Quarterly))
	}

	firstQuarterly := earnings.Quarterly[0]
	if firstQuarterly.FiscalDateEnding != "2023-12-30" {
		t.Errorf("expected fiscal date 2023-12-30, got %s", firstQuarterly.FiscalDateEnding)
	}
	if firstQuarterly.ReportedDate != "2024-02-01" {
		t.Errorf("expected reported date 2024-02-01, got %s", firstQuarterly.ReportedDate)
	}
	expectedReportedEPS := 2.18
	if got := firstQuarterly.ReportedEPSValue(); got != expectedReportedEPS {
		t.Errorf("expected ReportedEPSValue %f, got %f", expectedReportedEPS, got)
	}
	expectedEstimatedEPS := 2.10
	if got := firstQuarterly.EstimatedEPSValue(); got != expectedEstimatedEPS {
		t.Errorf("expected EstimatedEPSValue %f, got %f", expectedEstimatedEPS, got)
	}
	expectedSurprise := 0.08
	if got := firstQuarterly.SurpriseValue(); got != expectedSurprise {
		t.Errorf("expected SurpriseValue %f, got %f", expectedSurprise, got)
	}
	expectedSurprisePct := 3.8095
	if got := firstQuarterly.SurprisePercentageValue(); got != expectedSurprisePct {
		t.Errorf("expected SurprisePercentageValue %f, got %f", expectedSurprisePct, got)
	}
}

func TestClient_GetEarnings_SymbolWithNoHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"symbol": "TEST", "annualEarnings": [], "quarterlyEarnings": []}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	earnings, err := client.GetEarnings(context.Background(), "TEST")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if earnings.Symbol != "TEST" {
		t.Errorf("expected Symbol TEST, got %s", earnings.Symbol)
	}
	if len(earnings.Annual) != 0 {
		t.Errorf("expected empty annual earnings, got %d entries", len(earnings.Annual))
	}
	if len(earnings.Quarterly) != 0 {
		t.Errorf("expected empty quarterly earnings, got %d entries", len(earnings.Quarterly))
	}
}

func TestClient_GetEarnings_EmptySymbolRejected(t *testing.T) {
	client := New(Config{
		Key:        "test-key",
		BaseURL:    "https://example.com",
		HTTPClient: &http.Client{},
	})

	_, err := client.GetEarnings(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty symbol, got nil")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected error to mention empty symbol, got: %v", err)
	}
}

func TestClient_GetEarnings_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"symbol": "AAPL"}`)
	}))
	defer server.Close()

	client := New(Config{
		Key:        "test-key",
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetEarnings(ctx, "AAPL")
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	if !strings.Contains(err.Error(), "canceled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestCompanyOverview_UnmarshalJSON_FullFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "overview_aapl.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	var overview CompanyOverview
	if err := json.Unmarshal(data, &overview); err != nil {
		t.Fatalf("unmarshaling: %v", err)
	}

	// Verify all required fields
	if overview.Symbol != "AAPL" {
		t.Errorf("expected Symbol AAPL, got %s", overview.Symbol)
	}
	if overview.Name != "Apple Inc." {
		t.Errorf("expected Name Apple Inc., got %s", overview.Name)
	}
	if overview.Exchange != "NASDAQ" {
		t.Errorf("expected Exchange NASDAQ, got %s", overview.Exchange)
	}
	if overview.Currency != "USD" {
		t.Errorf("expected Currency USD, got %s", overview.Currency)
	}
	if overview.Country != "USA" {
		t.Errorf("expected Country USA, got %s", overview.Country)
	}
	if overview.Sector != "Technology" {
		t.Errorf("expected Sector Technology, got %s", overview.Sector)
	}
	if overview.Industry != "Consumer Electronics" {
		t.Errorf("expected Industry Consumer Electronics, got %s", overview.Industry)
	}
	if overview.FiscalYearEnd != "September" {
		t.Errorf("expected FiscalYearEnd September, got %s", overview.FiscalYearEnd)
	}
}

func TestCompanyOverview_UnmarshalJSON_NoneSentinel(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "overview_msft_missing_fields.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	var overview CompanyOverview
	if err := json.Unmarshal(data, &overview); err != nil {
		t.Fatalf("unmarshaling: %v", err)
	}

	// Verify "None" values are parsed as strings
	if overview.MarketCapitalization != "None" {
		t.Errorf("expected MarketCapitalization None, got %s", overview.MarketCapitalization)
	}
	if overview.PEGRatio != "None" {
		t.Errorf("expected PEGRatio None, got %s", overview.PEGRatio)
	}

	// Verify value methods return 0 for "None"
	if got := overview.MarketCapitalizationValue(); got != 0 {
		t.Errorf("expected MarketCapitalizationValue 0, got %f", got)
	}
	if got := overview.PEGRatioValue(); got != 0 {
		t.Errorf("expected PEGRatioValue 0, got %f", got)
	}
}

func TestValidateSymbol(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		wantErr bool
	}{
		{
			name:    "valid uppercase",
			symbol:  "AAPL",
			wantErr: false,
		},
		{
			name:    "valid lowercase",
			symbol:  "aapl",
			wantErr: false,
		},
		{
			name:    "valid with dot",
			symbol:  "BRK.B",
			wantErr: false,
		},
		{
			name:    "valid with dash",
			symbol:  "GOOGL-TEST",
			wantErr: false,
		},
		{
			name:    "empty string",
			symbol:  "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			symbol:  "   ",
			wantErr: true,
		},
		{
			name:    "too long",
			symbol:  "TOOLONGSYMBOL",
			wantErr: true,
		},
		{
			name:    "invalid character",
			symbol:  "AAPL$",
			wantErr: true,
		},
		{
			name:    "special characters",
			symbol:  "AAPL@#",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSymbol(tt.symbol)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSymbol(%q) error = %v, wantErr %v", tt.symbol, err, tt.wantErr)
			}
		})
	}
}
