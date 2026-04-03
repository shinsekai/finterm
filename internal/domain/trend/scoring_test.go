// Package trend provides trend-following analysis and scoring.
package trend

import (
	"testing"
)

func TestScore_Bullish_EMAFastAboveSlow(t *testing.T) {
	tests := []struct {
		name       string
		emaFast    float64
		emaSlow    float64
		wantSignal Signal
		wantString string
	}{
		{
			name:       "EMA fast slightly above slow",
			emaFast:    150.5,
			emaSlow:    150.0,
			wantSignal: Bullish,
			wantString: "Bullish",
		},
		{
			name:       "EMA fast significantly above slow",
			emaFast:    200.0,
			emaSlow:    150.0,
			wantSignal: Bullish,
			wantString: "Bullish",
		},
		{
			name:       "EMA fast negative values, fast above slow",
			emaFast:    -50.0,
			emaSlow:    -100.0,
			wantSignal: Bullish,
			wantString: "Bullish",
		},
		{
			name:       "EMA fast zero, slow negative",
			emaFast:    0.0,
			emaSlow:    -10.0,
			wantSignal: Bullish,
			wantString: "Bullish",
		},
		{
			name:       "EMA fast small positive, slow zero",
			emaFast:    0.01,
			emaSlow:    0.0,
			wantSignal: Bullish,
			wantString: "Bullish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.emaFast, tt.emaSlow)
			if got != tt.wantSignal {
				t.Errorf("Score(%f, %f) = %v, want %v", tt.emaFast, tt.emaSlow, got, tt.wantSignal)
			}
			if got.String() != tt.wantString {
				t.Errorf("Score(%f, %f).String() = %q, want %q", tt.emaFast, tt.emaSlow, got.String(), tt.wantString)
			}
		})
	}
}

func TestScore_Bearish_EMAFastBelowSlow(t *testing.T) {
	tests := []struct {
		name       string
		emaFast    float64
		emaSlow    float64
		wantSignal Signal
		wantString string
	}{
		{
			name:       "EMA fast slightly below slow",
			emaFast:    149.5,
			emaSlow:    150.0,
			wantSignal: Bearish,
			wantString: "Bearish",
		},
		{
			name:       "EMA fast significantly below slow",
			emaFast:    100.0,
			emaSlow:    150.0,
			wantSignal: Bearish,
			wantString: "Bearish",
		},
		{
			name:       "EMA fast negative values, fast below slow",
			emaFast:    -150.0,
			emaSlow:    -100.0,
			wantSignal: Bearish,
			wantString: "Bearish",
		},
		{
			name:       "EMA fast zero, slow positive",
			emaFast:    0.0,
			emaSlow:    10.0,
			wantSignal: Bearish,
			wantString: "Bearish",
		},
		{
			name:       "EMA fast negative, slow zero",
			emaFast:    -0.01,
			emaSlow:    0.0,
			wantSignal: Bearish,
			wantString: "Bearish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.emaFast, tt.emaSlow)
			if got != tt.wantSignal {
				t.Errorf("Score(%f, %f) = %v, want %v", tt.emaFast, tt.emaSlow, got, tt.wantSignal)
			}
			if got.String() != tt.wantString {
				t.Errorf("Score(%f, %f).String() = %q, want %q", tt.emaFast, tt.emaSlow, got.String(), tt.wantString)
			}
		})
	}
}

func TestScore_BoundaryEqual(t *testing.T) {
	tests := []struct {
		name       string
		emaFast    float64
		emaSlow    float64
		wantSignal Signal
		wantString string
	}{
		{
			name:       "EMA equal positive values",
			emaFast:    150.0,
			emaSlow:    150.0,
			wantSignal: Bearish, // Equal is treated as Bearish
			wantString: "Bearish",
		},
		{
			name:       "EMA equal zero",
			emaFast:    0.0,
			emaSlow:    0.0,
			wantSignal: Bearish,
			wantString: "Bearish",
		},
		{
			name:       "EMA equal negative values",
			emaFast:    -100.0,
			emaSlow:    -100.0,
			wantSignal: Bearish,
			wantString: "Bearish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.emaFast, tt.emaSlow)
			if got != tt.wantSignal {
				t.Errorf("Score(%f, %f) = %v, want %v", tt.emaFast, tt.emaSlow, got, tt.wantSignal)
			}
			if got.String() != tt.wantString {
				t.Errorf("Score(%f, %f).String() = %q, want %q", tt.emaFast, tt.emaSlow, got.String(), tt.wantString)
			}
		})
	}
}

func TestSignal_String(t *testing.T) {
	tests := []struct {
		name     string
		signal   Signal
		expected string
	}{
		{
			name:     "Bullish string",
			signal:   Bullish,
			expected: "Bullish",
		},
		{
			name:     "Bearish string",
			signal:   Bearish,
			expected: "Bearish",
		},
		{
			name:     "Unknown signal",
			signal:   Signal(99),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.signal.String(); got != tt.expected {
				t.Errorf("Signal(%d).String() = %q, want %q", tt.signal, got, tt.expected)
			}
		})
	}
}
