// Package quote provides the single ticker quote TUI view.
package quote

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shinsekai/finterm/internal/alphavantage"
)

// FundamentalsData contains company overview and earnings data.
type FundamentalsData struct {
	Overview *alphavantage.CompanyOverview
	Earnings *alphavantage.Earnings
}

// NextEarningsDate returns the next upcoming earnings date from quarterly earnings.
// Returns empty string if no future earnings are found.
func (f *FundamentalsData) NextEarningsDate() string {
	if f.Earnings == nil || len(f.Earnings.Quarterly) == 0 {
		return ""
	}

	today := time.Now()
	for _, q := range f.Earnings.Quarterly {
		if q.ReportedDate == "" {
			continue
		}
		reportedDate, err := time.Parse("2006-01-02", q.ReportedDate)
		if err != nil {
			continue
		}
		if reportedDate.After(today) {
			return q.ReportedDate
		}
	}
	return ""
}

// formatNumber formats a numeric string with thousands separators.
// Returns "—" for empty or invalid values.
func formatNumber(s string) string {
	if s == "" || s == "None" || s == "-" || s == "null" {
		return "—"
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil || val == 0 {
		return "—"
	}

	// Format with thousands separators
	str := fmt.Sprintf("%.0f", val)
	n := len(str)
	if n > 3 {
		var result []byte
		for i, c := range str {
			if i > 0 && (n-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		str = string(result)
	}
	return str
}

// formatPrice formats a price string with 2 decimal places.
// Returns "—" for empty or invalid values.
func formatPriceFundamentals(s string) string {
	if s == "" || s == "None" || s == "-" || s == "null" {
		return "—"
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil || val == 0 {
		return "—"
	}

	return fmt.Sprintf("$%.2f", val)
}

// formatPercent formats a percentage string.
// Returns "—" for empty or invalid values.
func formatPercentFundamentals(s string) string {
	if s == "" || s == "None" || s == "-" || s == "null" {
		return "—"
	}

	clean := strings.TrimSuffix(s, "%")
	pct, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return "—"
	}

	return fmt.Sprintf("%.2f%%", pct)
}

// formatMarketCap formats market capitalization with appropriate suffix.
// Returns "—" for empty or invalid values.
func formatMarketCap(s string) string {
	if s == "" || s == "None" || s == "-" || s == "null" {
		return "—"
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil || val == 0 {
		return "—"
	}

	switch {
	case val >= 1_000_000_000_000:
		return fmt.Sprintf("$%.2fT", val/1_000_000_000_000)
	case val >= 1_000_000_000:
		return fmt.Sprintf("$%.2fB", val/1_000_000_000)
	case val >= 1_000_000:
		return fmt.Sprintf("$%.2fM", val/1_000_000)
	default:
		return fmt.Sprintf("$%.0f", val)
	}
}

// formatField formats a label-value pair for fundamentals display.
func formatField(label, value string) string {
	if value == "" || value == "None" || value == "-" || value == "null" {
		value = "—"
	}
	return fmt.Sprintf("%-22s%s", label+":", value)
}

// renderFundamentalsPanel renders the fundamentals panel card.
func (v *View) renderFundamentalsPanel(fundamentals *FundamentalsData) string {
	if fundamentals == nil {
		return ""
	}

	overview := fundamentals.Overview
	if overview == nil {
		return ""
	}

	var content strings.Builder

	// Header
	content.WriteString(v.theme.CardHeader().Render("Fundamentals"))
	content.WriteString("\n")

	// Divider
	boxWidth := v.calculateBoxWidth()
	content.WriteString(v.theme.Divider().Render(strings.Repeat("─", boxWidth)))
	content.WriteString("\n\n")

	// Company name
	name := overview.Name
	if name == "" || name == "None" || name == "-" {
		name = "—"
	}
	fmt.Fprintf(&content, "%s\n\n", v.theme.Muted().Render(formatField("Company", name)))

	// Sector
	sector := overview.Sector
	if sector == "" || sector == "None" || sector == "-" {
		sector = "—"
	}
	fmt.Fprintf(&content, "%s\n\n", v.theme.Muted().Render(formatField("Sector", sector)))

	// Market cap
	marketCap := formatMarketCap(overview.MarketCapitalization)
	fmt.Fprintf(&content, "%s\n", v.theme.Muted().Render(formatField("Market Cap", marketCap)))

	// P/E Ratio
	pe := formatPriceFundamentals(overview.PERatio)
	if pe != "—" {
		pe = strings.TrimPrefix(pe, "$")
	}
	fmt.Fprintf(&content, "%s\n", v.theme.Muted().Render(formatField("P/E Ratio", pe)))

	// EPS
	eps := formatPriceFundamentals(overview.EPS)
	if eps != "—" {
		eps = strings.TrimPrefix(eps, "$")
	}
	fmt.Fprintf(&content, "%s\n", v.theme.Muted().Render(formatField("EPS", eps)))

	// Dividend Yield
	divYield := formatPercentFundamentals(overview.DividendYield)
	fmt.Fprintf(&content, "%s\n\n", v.theme.Muted().Render(formatField("Dividend Yield", divYield)))

	// Next earnings date
	nextEarnings := fundamentals.NextEarningsDate()
	if nextEarnings == "" {
		nextEarnings = "—"
	}
	fmt.Fprintf(&content, "%s\n\n", v.theme.Muted().Render(formatField("Next Earnings", nextEarnings)))

	// 52-week range
	high52 := formatPriceFundamentals(overview.FiftyTwoWeekHigh)
	low52 := formatPriceFundamentals(overview.FiftyTwoWeekLow)
	var range52 string
	if high52 == "—" || low52 == "—" {
		range52 = "—"
	} else {
		high := strings.TrimPrefix(high52, "$")
		low := strings.TrimPrefix(low52, "$")
		range52 = fmt.Sprintf("%s - %s", low, high)
	}
	fmt.Fprintf(&content, "%s\n", v.theme.Muted().Render(formatField("52-Week Range", range52)))

	return v.theme.Card().Render(content.String())
}
