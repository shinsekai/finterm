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

func TestTPI_AllBullish(t *testing.T) {
	// TPI(Bullish, 1, 1, 1, 1) returns 1.0 (all 5 signals agree long)
	got := TPI(Bullish, 1, 1, 1, 1)
	want := 1.0
	if got != want {
		t.Errorf("TPI(Bullish, 1, 1, 1, 1) = %f, want %f", got, want)
	}
}

func TestTPI_AllBearish(t *testing.T) {
	// TPI(Bearish, -1, -1, -1, -1) returns -1.0 (all 5 signals agree short)
	got := TPI(Bearish, -1, -1, -1, -1)
	want := -1.0
	if got != want {
		t.Errorf("TPI(Bearish, -1, -1, -1, -1) = %f, want %f", got, want)
	}
}

func TestTPI_Mixed_BullishMajority(t *testing.T) {
	// TPI(Bullish, 1, 0, 0, 0) returns 0.4 (2 of 5 long: EMA=1, blitz=1)
	got := TPI(Bullish, 1, 0, 0, 0)
	want := 0.4
	if got != want {
		t.Errorf("TPI(Bullish, 1, 0, 0, 0) = %f, want %f", got, want)
	}
}

func TestTPI_Mixed_BearishMajority(t *testing.T) {
	// TPI(Bearish, -1, 0, 0, 0) returns -0.4 (2 of 5 short: EMA=-1, blitz=-1)
	got := TPI(Bearish, -1, 0, 0, 0)
	want := -0.4
	if got != want {
		t.Errorf("TPI(Bearish, -1, 0, 0, 0) = %f, want %f", got, want)
	}
}

func TestTPI_Neutral(t *testing.T) {
	// TPI(Bullish, 1, 0, -1, 0) returns 0.2 (2 long, 1 short: EMA=1, blitz=1, destiny=-1 → (1+1+0-1+0)/5 = 1/5 = 0.2)
	got := TPI(Bullish, 1, 0, -1, 0)
	want := 0.2
	if got != want {
		t.Errorf("TPI(Bullish, 1, 0, -1, 0) = %f, want %f", got, want)
	}
}

func TestTPI_AllHold(t *testing.T) {
	// TPI(Bearish, 0, 0, 0, 0) returns -0.2 (only EMA is bearish at -1, all others HOLD at 0)
	got := TPI(Bearish, 0, 0, 0, 0)
	want := -1.0 / 5.0
	if got != want {
		t.Errorf("TPI(Bearish, 0, 0, 0, 0) = %f, want %f", got, want)
	}
}

func TestTPISignal_Positive(t *testing.T) {
	// TPISignal returns "LONG" for positive
	got := TPISignal(0.5)
	want := "LONG"
	if got != want {
		t.Errorf("TPISignal(0.5) = %q, want %q", got, want)
	}
}

func TestTPISignal_Zero(t *testing.T) {
	// TPISignal returns "CASH" for zero
	got := TPISignal(0.0)
	want := "CASH"
	if got != want {
		t.Errorf("TPISignal(0.0) = %q, want %q", got, want)
	}
}

func TestTPISignal_Negative(t *testing.T) {
	// TPISignal returns "CASH" for negative
	got := TPISignal(-0.5)
	want := "CASH"
	if got != want {
		t.Errorf("TPISignal(-0.5) = %q, want %q", got, want)
	}
}
