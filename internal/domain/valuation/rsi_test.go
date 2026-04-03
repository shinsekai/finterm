package valuation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValuate_Oversold(t *testing.T) {
	tests := []struct {
		name string
		rsi  float64
	}{
		{"RSI 0", 0},
		{"RSI 10", 10},
		{"RSI 20", 20},
		{"RSI 29.9", 29.9},
		{"Just below 30", 29.999},
	}

	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Valuate(tt.rsi, cfg)
			assert.Equal(t, "Oversold", result)
		})
	}
}

func TestValuate_Undervalued(t *testing.T) {
	tests := []struct {
		name string
		rsi  float64
	}{
		{"RSI 30", 30},
		{"RSI 35", 35},
		{"RSI 40", 40},
		{"RSI 44", 44},
		{"Just below 45", 44.999},
	}

	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Valuate(tt.rsi, cfg)
			assert.Equal(t, "Undervalued", result)
		})
	}
}

func TestValuate_FairValue(t *testing.T) {
	tests := []struct {
		name string
		rsi  float64
	}{
		{"RSI 45", 45},
		{"RSI 50", 50},
		{"RSI 52", 52},
		{"RSI 54", 54},
		{"Just below 55", 54.999},
	}

	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Valuate(tt.rsi, cfg)
			assert.Equal(t, "Fair value", result)
		})
	}
}

func TestValuate_Overvalued(t *testing.T) {
	tests := []struct {
		name string
		rsi  float64
	}{
		{"RSI 55", 55},
		{"RSI 60", 60},
		{"RSI 65", 65},
		{"RSI 69", 69},
		{"Just below 70", 69.999},
	}

	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Valuate(tt.rsi, cfg)
			assert.Equal(t, "Overvalued", result)
		})
	}
}

func TestValuate_Overbought(t *testing.T) {
	tests := []struct {
		name string
		rsi  float64
	}{
		{"RSI 70", 70},
		{"RSI 75", 75},
		{"RSI 80", 80},
		{"RSI 90", 90},
		{"RSI 100", 100},
	}

	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Valuate(tt.rsi, cfg)
			assert.Equal(t, "Overbought", result)
		})
	}
}

func TestValuate_BoundaryAt30(t *testing.T) {
	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	t.Run("Below boundary", func(t *testing.T) {
		result := Valuate(29.9, cfg)
		assert.Equal(t, "Oversold", result)
	})

	t.Run("At boundary", func(t *testing.T) {
		result := Valuate(30, cfg)
		assert.Equal(t, "Undervalued", result)
	})

	t.Run("Above boundary", func(t *testing.T) {
		result := Valuate(30.1, cfg)
		assert.Equal(t, "Undervalued", result)
	})
}

func TestValuate_BoundaryAt45(t *testing.T) {
	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	t.Run("Below boundary", func(t *testing.T) {
		result := Valuate(44.9, cfg)
		assert.Equal(t, "Undervalued", result)
	})

	t.Run("At boundary", func(t *testing.T) {
		result := Valuate(45, cfg)
		assert.Equal(t, "Fair value", result)
	})

	t.Run("Above boundary", func(t *testing.T) {
		result := Valuate(45.1, cfg)
		assert.Equal(t, "Fair value", result)
	})
}

func TestValuate_BoundaryAt55(t *testing.T) {
	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	t.Run("Below boundary", func(t *testing.T) {
		result := Valuate(54.9, cfg)
		assert.Equal(t, "Fair value", result)
	})

	t.Run("At boundary", func(t *testing.T) {
		result := Valuate(55, cfg)
		assert.Equal(t, "Overvalued", result)
	})

	t.Run("Above boundary", func(t *testing.T) {
		result := Valuate(55.1, cfg)
		assert.Equal(t, "Overvalued", result)
	})
}

func TestValuate_BoundaryAt70(t *testing.T) {
	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	t.Run("Below boundary", func(t *testing.T) {
		result := Valuate(69.9, cfg)
		assert.Equal(t, "Overvalued", result)
	})

	t.Run("At boundary", func(t *testing.T) {
		result := Valuate(70, cfg)
		assert.Equal(t, "Overbought", result)
	})

	t.Run("Above boundary", func(t *testing.T) {
		result := Valuate(70.1, cfg)
		assert.Equal(t, "Overbought", result)
	})
}

func TestValuate_ExtremeValues(t *testing.T) {
	cfg := Config{
		Oversold:    30,
		Undervalued: 45,
		FairLow:     45,
		FairHigh:    55,
		Overvalued:  70,
		Overbought:  70,
	}

	t.Run("RSI 0 maps to Oversold", func(t *testing.T) {
		result := Valuate(0, cfg)
		assert.Equal(t, "Oversold", result)
	})

	t.Run("RSI 100 maps to Overbought", func(t *testing.T) {
		result := Valuate(100, cfg)
		assert.Equal(t, "Overbought", result)
	})
}
