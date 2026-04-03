// Package valuation provides asset valuation analysis functions.
package valuation

// Config contains thresholds for RSI-based valuation.
type Config struct {
	Oversold    float64
	Undervalued float64
	FairLow     float64
	FairHigh    float64
	Overvalued  float64
	Overbought  float64
}

// Valuate returns a valuation label based on the RSI value.
// All computations use the RSI value from the last closed bar only.
// Thresholds represent upper bounds: Oversold(<30), Undervalued(<45), Fair(<55), Overvalued(<70), Overbought(>=70).
func Valuate(rsi float64, cfg Config) string {
	switch {
	case rsi >= cfg.Overbought:
		return "Overbought"
	case rsi >= cfg.FairHigh:
		return "Overvalued"
	case rsi >= cfg.Undervalued:
		return "Fair value"
	case rsi >= cfg.Oversold:
		return "Undervalued"
	default:
		return "Oversold"
	}
}
