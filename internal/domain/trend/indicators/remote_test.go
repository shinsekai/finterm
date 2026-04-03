// Package indicators tests the remote indicator implementations.
package indicators

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/owner/finterm/internal/alphavantage"
)

// mockAVClient is a mock implementation of AVClient for testing.
type mockAVClient struct {
	rsiFunc func(_ context.Context, symbol string, interval string, period int) (*alphavantage.RSIResponse, error)
	emaFunc func(_ context.Context, symbol string, interval string, period int) (*alphavantage.EMAResponse, error)
}

func (m *mockAVClient) GetRSI(ctx context.Context, symbol string, interval string, period int) (*alphavantage.RSIResponse, error) {
	if m.rsiFunc != nil {
		return m.rsiFunc(ctx, symbol, interval, period)
	}
	return nil, nil
}

func (m *mockAVClient) GetEMA(ctx context.Context, symbol string, interval string, period int) (*alphavantage.EMAResponse, error) {
	if m.emaFunc != nil {
		return m.emaFunc(ctx, symbol, interval, period)
	}
	return nil, nil
}

// TestRemoteRSI_Success verifies that RemoteRSI calls the correct AV client method
// with the correct parameters and converts the response correctly.
func TestRemoteRSI_Success(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		interval string
		period   int
		resp     *alphavantage.RSIResponse
		wantLen  int
	}{
		{
			name:     "Standard daily RSI",
			symbol:   "AAPL",
			interval: "daily",
			period:   14,
			resp: &alphavantage.RSIResponse{
				MetaData: alphavantage.IndicatorMetadata{
					Symbol:     "AAPL",
					Indicator:  "Relative Strength Index (RSI)",
					Interval:   "daily",
					TimePeriod: "14",
					SeriesType: "close",
				},
				TechnicalAnalysis: map[string]alphavantage.RSIEntry{
					"2024-01-15": {RSI: "65.43"},
					"2024-01-14": {RSI: "62.12"},
					"2024-01-13": {RSI: "58.76"},
					"2024-01-12": {RSI: "55.34"},
					"2024-01-11": {RSI: "51.23"},
				},
			},
			wantLen: 5,
		},
		{
			name:     "Single data point",
			symbol:   "MSFT",
			interval: "weekly",
			period:   14,
			resp: &alphavantage.RSIResponse{
				TechnicalAnalysis: map[string]alphavantage.RSIEntry{
					"2024-01-15": {RSI: "50.0"},
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotSymbol, gotInterval string
			var gotPeriod int

			mockClient := &mockAVClient{
				rsiFunc: func(_ context.Context, symbol string, interval string, period int) (*alphavantage.RSIResponse, error) {
					gotSymbol = symbol
					gotInterval = interval
					gotPeriod = period
					return tt.resp, nil
				},
			}

			remoteRSI := NewRemoteRSI(mockClient)
			result, err := remoteRSI.Compute(context.Background(), tt.symbol, Options{
				Interval: tt.interval,
				Period:   tt.period,
			})

			if err != nil {
				t.Fatalf("Compute() failed: %v", err)
			}

			// Verify correct parameters were passed to client
			if gotSymbol != tt.symbol {
				t.Errorf("GetRSI() symbol = %q, want %q", gotSymbol, tt.symbol)
			}
			if gotInterval != tt.interval {
				t.Errorf("GetRSI() interval = %q, want %q", gotInterval, tt.interval)
			}
			if gotPeriod != tt.period {
				t.Errorf("GetRSI() period = %d, want %d", gotPeriod, tt.period)
			}

			// Verify result length
			if len(result) != tt.wantLen {
				t.Errorf("Compute() result length = %d, want %d", len(result), tt.wantLen)
			}

			// Verify sorting by date descending (newest first)
			for i := 1; i < len(result); i++ {
				if result[i].Date.After(result[i-1].Date) {
					t.Errorf("Result not sorted by date descending at index %d: %v comes after %v",
						i, result[i].Date, result[i-1].Date)
				}
			}
		})
	}
}

// TestRemoteRSI_ClientError verifies that RemoteRSI propagates client errors
// with appropriate context.
func TestRemoteRSI_ClientError(t *testing.T) {
	tests := []struct {
		name        string
		symbol      string
		interval    string
		period      int
		clientErr   error
		wantErrText string
	}{
		{
			name:        "Client returns network error",
			symbol:      "AAPL",
			interval:    "daily",
			period:      14,
			clientErr:   errors.New("network timeout"),
			wantErrText: "fetching RSI for AAPL",
		},
		{
			name:        "Client returns API error",
			symbol:      "INVALID",
			interval:    "daily",
			period:      14,
			clientErr:   &alphavantage.APIError{StatusCode: 400, Message: "Invalid API key", Endpoint: "/query"},
			wantErrText: "fetching RSI for INVALID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAVClient{
				rsiFunc: func(_ context.Context, _ string, _ string, _ int) (*alphavantage.RSIResponse, error) {
					return nil, tt.clientErr
				},
			}

			remoteRSI := NewRemoteRSI(mockClient)
			_, err := remoteRSI.Compute(context.Background(), tt.symbol, Options{
				Interval: tt.interval,
				Period:   tt.period,
			})

			if err == nil {
				t.Fatal("Compute() expected error, got nil")
			}

			// Verify error includes context
			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.wantErrText) {
				t.Errorf("Error message should contain %q, got %q", tt.wantErrText, errMsg)
			}

			// Verify underlying error is preserved
			if !errors.Is(err, tt.clientErr) {
				// This might not work if we wrap the error differently
				// but we should check that the error was propagated
				if !strings.Contains(errMsg, tt.clientErr.Error()) {
					t.Errorf("Error should wrap client error, got %q", errMsg)
				}
			}
		})
	}
}

// TestRemoteEMA_Success verifies that RemoteEMA calls the correct AV client method
// with the correct parameters and converts the response correctly.
func TestRemoteEMA_Success(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		interval string
		period   int
		resp     *alphavantage.EMAResponse
		wantLen  int
	}{
		{
			name:     "Fast EMA (10 period)",
			symbol:   "AAPL",
			interval: "daily",
			period:   10,
			resp: &alphavantage.EMAResponse{
				MetaData: alphavantage.IndicatorMetadata{
					Symbol:     "AAPL",
					Indicator:  "Exponential Moving Average (EMA)",
					Interval:   "daily",
					TimePeriod: "10",
					SeriesType: "close",
				},
				TechnicalAnalysis: map[string]alphavantage.EMAEntry{
					"2024-01-15": {EMA: "185.42"},
					"2024-01-14": {EMA: "184.23"},
					"2024-01-13": {EMA: "183.11"},
					"2024-01-12": {EMA: "182.05"},
				},
			},
			wantLen: 4,
		},
		{
			name:     "Slow EMA (20 period)",
			symbol:   "GOOGL",
			interval: "daily",
			period:   20,
			resp: &alphavantage.EMAResponse{
				TechnicalAnalysis: map[string]alphavantage.EMAEntry{
					"2024-01-15": {EMA: "140.50"},
					"2024-01-14": {EMA: "140.25"},
					"2024-01-13": {EMA: "140.00"},
				},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotSymbol, gotInterval string
			var gotPeriod int

			mockClient := &mockAVClient{
				emaFunc: func(_ context.Context, symbol string, interval string, period int) (*alphavantage.EMAResponse, error) {
					gotSymbol = symbol
					gotInterval = interval
					gotPeriod = period
					return tt.resp, nil
				},
			}

			remoteEMA := NewRemoteEMA(mockClient)
			result, err := remoteEMA.Compute(context.Background(), tt.symbol, Options{
				Interval: tt.interval,
				Period:   tt.period,
			})

			if err != nil {
				t.Fatalf("Compute() failed: %v", err)
			}

			// Verify correct parameters were passed to client
			if gotSymbol != tt.symbol {
				t.Errorf("GetEMA() symbol = %q, want %q", gotSymbol, tt.symbol)
			}
			if gotInterval != tt.interval {
				t.Errorf("GetEMA() interval = %q, want %q", gotInterval, tt.interval)
			}
			if gotPeriod != tt.period {
				t.Errorf("GetEMA() period = %d, want %d", gotPeriod, tt.period)
			}

			// Verify result length
			if len(result) != tt.wantLen {
				t.Errorf("Compute() result length = %d, want %d", len(result), tt.wantLen)
			}

			// Verify sorting by date descending (newest first)
			for i := 1; i < len(result); i++ {
				if result[i].Date.After(result[i-1].Date) {
					t.Errorf("Result not sorted by date descending at index %d: %v comes after %v",
						i, result[i].Date, result[i-1].Date)
				}
			}
		})
	}
}

// TestRemoteEMA_ClientError verifies that RemoteEMA propagates client errors
// with appropriate context.
func TestRemoteEMA_ClientError(t *testing.T) {
	tests := []struct {
		name        string
		symbol      string
		interval    string
		period      int
		clientErr   error
		wantErrText string
	}{
		{
			name:        "Client returns network error",
			symbol:      "AAPL",
			interval:    "daily",
			period:      10,
			clientErr:   errors.New("network timeout"),
			wantErrText: "fetching EMA for AAPL",
		},
		{
			name:        "Client returns rate limit error",
			symbol:      "MSFT",
			interval:    "daily",
			period:      20,
			clientErr:   &alphavantage.APIError{StatusCode: 429, Message: "Rate limit exceeded", Endpoint: "/query"},
			wantErrText: "fetching EMA for MSFT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAVClient{
				emaFunc: func(_ context.Context, _ string, _ string, _ int) (*alphavantage.EMAResponse, error) {
					return nil, tt.clientErr
				},
			}

			remoteEMA := NewRemoteEMA(mockClient)
			_, err := remoteEMA.Compute(context.Background(), tt.symbol, Options{
				Interval: tt.interval,
				Period:   tt.period,
			})

			if err == nil {
				t.Fatal("Compute() expected error, got nil")
			}

			// Verify error includes context
			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.wantErrText) {
				t.Errorf("Error message should contain %q, got %q", tt.wantErrText, errMsg)
			}
		})
	}
}

// TestRemoteRSI_ResponseConversion verifies the conversion logic from Alpha Vantage
// response format to []DataPoint, including date parsing and value parsing.
func TestRemoteRSI_ResponseConversion(t *testing.T) {
	tests := []struct {
		name       string
		resp       *alphavantage.RSIResponse
		wantErr    bool
		wantValues []float64
		wantDates  []time.Time
	}{
		{
			name: "Valid response with multiple entries",
			resp: &alphavantage.RSIResponse{
				TechnicalAnalysis: map[string]alphavantage.RSIEntry{
					"2024-01-15": {RSI: "70.5"},
					"2024-01-14": {RSI: "65.25"},
					"2024-01-13": {RSI: "60.0"},
				},
			},
			wantErr:    false,
			wantValues: []float64{70.5, 65.25, 60.0},
			wantDates: []time.Time{
				time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Response with RSI boundary values",
			resp: &alphavantage.RSIResponse{
				TechnicalAnalysis: map[string]alphavantage.RSIEntry{
					"2024-01-15": {RSI: "0.0"},
					"2024-01-14": {RSI: "50.0"},
					"2024-01-13": {RSI: "100.0"},
				},
			},
			wantErr:    false,
			wantValues: []float64{0.0, 50.0, 100.0},
		},
		{
			name:    "Nil response",
			resp:    nil,
			wantErr: true,
		},
		{
			name: "Empty technical analysis",
			resp: &alphavantage.RSIResponse{
				TechnicalAnalysis: map[string]alphavantage.RSIEntry{},
			},
			wantErr: true,
		},
		{
			name: "Invalid date format",
			resp: &alphavantage.RSIResponse{
				TechnicalAnalysis: map[string]alphavantage.RSIEntry{
					"invalid-date": {RSI: "50.0"},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid RSI value format",
			resp: &alphavantage.RSIResponse{
				TechnicalAnalysis: map[string]alphavantage.RSIEntry{
					"2024-01-15": {RSI: "not-a-number"},
				},
			},
			wantErr: true,
		},
		{
			name: "Response with intraday timestamp",
			resp: &alphavantage.RSIResponse{
				TechnicalAnalysis: map[string]alphavantage.RSIEntry{
					"2024-01-15 10:30:00": {RSI: "55.5"},
					"2024-01-15 10:25:00": {RSI: "54.3"},
				},
			},
			wantErr:    false,
			wantValues: []float64{55.5, 54.3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAVClient{
				rsiFunc: func(_ context.Context, _ string, _ string, _ int) (*alphavantage.RSIResponse, error) {
					return tt.resp, nil
				},
			}

			remoteRSI := NewRemoteRSI(mockClient)
			result, err := remoteRSI.Compute(context.Background(), "AAPL", Options{
				Interval: "daily",
				Period:   14,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Compute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify values
			if len(result) != len(tt.wantValues) {
				t.Fatalf("Expected %d data points, got %d", len(tt.wantValues), len(result))
			}

			// Note: results are sorted descending by date, so order may differ from input
			// We need to check that all expected values are present
			for _, expectedValue := range tt.wantValues {
				found := false
				for _, dp := range result {
					if almostEqual(dp.Value, expectedValue) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected value %f not found in results", expectedValue)
				}
			}

			// Verify dates if specified
			if len(tt.wantDates) > 0 {
				for i, expectedDate := range tt.wantDates {
					// Find matching date in results (since sorted)
					found := false
					for _, dp := range result {
						if dp.Date.Equal(expectedDate) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected date %v (index %d) not found in results", expectedDate, i)
					}
				}
			}
		})
	}
}

// TestRemoteRSI_InvalidInput verifies error handling for invalid input parameters.
func TestRemoteRSI_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		interval string
		period   int
		wantErr  bool
	}{
		{
			name:     "Zero period",
			symbol:   "AAPL",
			interval: "daily",
			period:   0,
			wantErr:  true,
		},
		{
			name:     "Negative period",
			symbol:   "AAPL",
			interval: "daily",
			period:   -5,
			wantErr:  true,
		},
		{
			name:     "Empty interval",
			symbol:   "AAPL",
			interval: "",
			period:   14,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAVClient{
				rsiFunc: func(_ context.Context, _ string, _ string, _ int) (*alphavantage.RSIResponse, error) {
					return &alphavantage.RSIResponse{}, nil
				},
			}

			remoteRSI := NewRemoteRSI(mockClient)
			_, err := remoteRSI.Compute(context.Background(), tt.symbol, Options{
				Interval: tt.interval,
				Period:   tt.period,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Compute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRemoteEMA_InvalidInput verifies error handling for invalid input parameters.
func TestRemoteEMA_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		interval string
		period   int
		wantErr  bool
	}{
		{
			name:     "Zero period",
			symbol:   "AAPL",
			interval: "daily",
			period:   0,
			wantErr:  true,
		},
		{
			name:     "Negative period",
			symbol:   "AAPL",
			interval: "daily",
			period:   -10,
			wantErr:  true,
		},
		{
			name:     "Empty interval",
			symbol:   "AAPL",
			interval: "",
			period:   20,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAVClient{
				emaFunc: func(_ context.Context, _ string, _ string, _ int) (*alphavantage.EMAResponse, error) {
					return &alphavantage.EMAResponse{}, nil
				},
			}

			remoteEMA := NewRemoteEMA(mockClient)
			_, err := remoteEMA.Compute(context.Background(), tt.symbol, Options{
				Interval: tt.interval,
				Period:   tt.period,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Compute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRemoteEMA_ResponseConversion verifies the conversion logic from Alpha Vantage
// EMA response format to []DataPoint.
func TestRemoteEMA_ResponseConversion(t *testing.T) {
	tests := []struct {
		name       string
		resp       *alphavantage.EMAResponse
		wantErr    bool
		wantValues []float64
	}{
		{
			name: "Valid EMA response",
			resp: &alphavantage.EMAResponse{
				TechnicalAnalysis: map[string]alphavantage.EMAEntry{
					"2024-01-15": {EMA: "185.50"},
					"2024-01-14": {EMA: "184.25"},
				},
			},
			wantErr:    false,
			wantValues: []float64{185.50, 184.25},
		},
		{
			name: "EMA with decimal precision",
			resp: &alphavantage.EMAResponse{
				TechnicalAnalysis: map[string]alphavantage.EMAEntry{
					"2024-01-15": {EMA: "123.456789"},
				},
			},
			wantErr:    false,
			wantValues: []float64{123.456789},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAVClient{
				emaFunc: func(_ context.Context, _ string, _ string, _ int) (*alphavantage.EMAResponse, error) {
					return tt.resp, nil
				},
			}

			remoteEMA := NewRemoteEMA(mockClient)
			result, err := remoteEMA.Compute(context.Background(), "AAPL", Options{
				Interval: "daily",
				Period:   10,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Compute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify values
			if len(result) != len(tt.wantValues) {
				t.Fatalf("Expected %d data points, got %d", len(tt.wantValues), len(result))
			}

			for _, expectedValue := range tt.wantValues {
				found := false
				for _, dp := range result {
					if almostEqual(dp.Value, expectedValue) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected value %f not found in results", expectedValue)
				}
			}
		})
	}
}
