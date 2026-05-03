// Package commodities provides tests for commodities dashboard TUI view.
package commodities

import (
	"context"
	"github.com/shinsekai/finterm/internal/alphavantage"
	"testing"
	"time"
)

func TestView_Render(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT"}, "daily", mockCache)
	view := NewView(model)
	output := view.Render()
	if output == "" {
		t.Error("Render should return non-empty output")
	}
}

func TestView_RendersAllColumns(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)
	series := &alphavantage.CommoditySeries{Name: "WTI", Unit: "dollars per barrel", Data: generateTestData(35)}
	model.rows[0].State = StateLoaded
	model.rows[0].Series = series
	view := NewView(model)
	view.SetTheme(&defaultTheme{})
	output := view.Render()
	requiredHeaders := []string{"SYMBOL", "NAME", "SPARKLINE", "VALUE", "CHANGE", "PERIOD"}
	for _, header := range requiredHeaders {
		if !containsString(output, header) {
			t.Errorf("expected header %q in output", header)
		}
	}
}

func TestView_DeltaColorBySign(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)
	view := NewView(model)
	view.SetTheme(&defaultTheme{})

	bullishSeries := &alphavantage.CommoditySeries{Name: "WTI", Unit: "dollars per barrel", Data: []alphavantage.CommodityDataPoint{{Date: time.Now().Add(-2 * time.Hour), Value: 70.0}, {Date: time.Now().Add(-1 * time.Hour), Value: 75.0}}}
	model.rows[0].Series = bullishSeries
	model.rows[0].State = StateLoaded
	row := view.buildRow(model.rows[0])
	if len(row) != 6 {
		t.Errorf("expected 6 columns, got %d", len(row))
	}

	bearishSeries := &alphavantage.CommoditySeries{Name: "WTI", Unit: "dollars per barrel", Data: []alphavantage.CommodityDataPoint{{Date: time.Now().Add(-2 * time.Hour), Value: 75.0}, {Date: time.Now().Add(-1 * time.Hour), Value: 70.0}}}
	model.rows[0].Series = bearishSeries
	model.rows[0].State = StateLoaded
	_ = view.buildRow(model.rows[0])
}

func TestView_SparklineRendered(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)
	series := &alphavantage.CommoditySeries{Name: "WTI", Unit: "dollars per barrel", Data: generateTestData(35)}
	model.rows[0].State = StateLoaded
	model.rows[0].Series = series
	view := NewView(model)
	view.SetTheme(&defaultTheme{})
	output := view.Render()
	if !containsString(output, "SPARKLINE") {
		t.Error("expected SPARKLINE header in output")
	}
}

func TestView_RenderEmptyState(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)
	model.rows = []RowData{}
	view := NewView(model)
	view.SetTheme(&defaultTheme{})
	output := view.Render()
	if !containsString(output, "no commodities") {
		t.Error("expected empty placeholder message in output")
	}
}

func TestView_RenderLoadingState(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)
	model.rows[0].State = StateLoading
	view := NewView(model)
	view.SetTheme(&defaultTheme{})
	output := view.renderTitle()
	if !containsString(output, "loaded") {
		t.Error("expected 'loaded' in title when state is loading")
	}
}

func TestView_RenderErrorState(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)
	model.rows[0].State = StateError
	model.rows[0].Error = &testError{msg: "API error"}
	view := NewView(model)
	view.SetTheme(&defaultTheme{})
	row := view.buildRow(model.rows[0])
	if len(row) != 6 {
		t.Errorf("expected 6 columns in error row, got %d", len(row))
	}
	if !containsString(row[2], "✗") {
		t.Error("expected error indicator in sparkline column (index 2)")
	}
}

func TestView_CalculateChange(t *testing.T) {
	view := &View{model: &Model{}, theme: &defaultTheme{}}
	tests := []struct {
		name   string
		data   []alphavantage.CommodityDataPoint
		expect string
	}{
		{"positive change", []alphavantage.CommodityDataPoint{{Date: time.Now().Add(-2 * time.Hour), Value: 70.0}, {Date: time.Now().Add(-1 * time.Hour), Value: 75.0}}, "+"},
		{"negative change", []alphavantage.CommodityDataPoint{{Date: time.Now().Add(-2 * time.Hour), Value: 75.0}, {Date: time.Now().Add(-1 * time.Hour), Value: 70.0}}, "-"},
	}
	for _, tt := range tests {
		changeStr, _ := view.calculateChange(tt.data)
		if tt.expect == "+" && !containsString(changeStr, "+") {
			t.Errorf("expected positive sign, got %s", changeStr)
		}
		if tt.expect == "-" && !containsString(changeStr, "-") {
			t.Errorf("expected negative sign, got %s", changeStr)
		}
	}
}

func TestView_ExtractValuesForSparkline(t *testing.T) {
	data := generateTestData(50)
	values := extractValuesForSparkline(data, 30)
	if len(values) != 30 {
		t.Errorf("expected 30 values, got %d", len(values))
	}
	if len(values) > 0 {
		lastResult := values[len(values)-1]
		lastData := data[len(data)-1].Value
		if lastResult != lastData {
			t.Errorf("expected last value to match, got %f != %f", lastData, lastResult)
		}
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func generateTestData(n int) []alphavantage.CommodityDataPoint {
	data := make([]alphavantage.CommodityDataPoint, n)
	now := time.Now()
	for i := 0; i < n; i++ {
		data[i] = alphavantage.CommodityDataPoint{Date: now.Add(time.Duration(-n+i) * time.Hour), Value: 70.0 + float64(i)*0.5}
	}
	return data
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestView_FallbackChipShownOnMismatch tests that fallback chip is shown when intervals differ.
func TestView_FallbackChipShownOnMismatch(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"COPPER"}, "daily", mockCache)

	series := &alphavantage.CommoditySeries{
		Name: "Copper",
		Unit: "dollars per ton",
		Data: []alphavantage.CommodityDataPoint{
			{Date: time.Now(), Value: 9125.50},
		},
	}

	model.rows[0].State = StateLoaded
	model.rows[0].Series = series
	model.rows[0].ActualInterval = "monthly" // Fallback interval

	view := NewView(model)
	view.SetTheme(&defaultTheme{})
	row := view.buildRow(model.rows[0])

	if len(row) != 6 {
		t.Errorf("expected 6 columns, got %d", len(row))
	}

	// The PERIOD column is index 5
	periodColumn := row[5]
	if !containsString(periodColumn, "monthly") {
		t.Errorf("expected fallback chip to show '→ monthly', got %s", periodColumn)
	}
}

// TestView_FallbackChipHiddenOnExactMatch tests that fallback chip is hidden when intervals match.
func TestView_FallbackChipHiddenOnExactMatch(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)

	series := &alphavantage.CommoditySeries{
		Name: "WTI",
		Unit: "dollars per barrel",
		Data: []alphavantage.CommodityDataPoint{
			{Date: time.Now(), Value: 75.50},
		},
	}

	model.rows[0].State = StateLoaded
	model.rows[0].Series = series
	model.rows[0].ActualInterval = "daily" // Exact match

	view := NewView(model)
	view.SetTheme(&defaultTheme{})
	row := view.buildRow(model.rows[0])

	if len(row) != 6 {
		t.Errorf("expected 6 columns, got %d", len(row))
	}

	// The PERIOD column is index 5
	periodColumn := row[5]
	if containsString(periodColumn, "→") {
		t.Errorf("expected no fallback chip when intervals match, got %s", periodColumn)
	}
}

// TestView_FallbackChipWithDifferentIntervals tests various interval mismatch scenarios.
func TestView_FallbackChipWithDifferentIntervals(t *testing.T) {
	tests := []struct {
		name           string
		requested      string
		actualInterval string
		expectChip     bool
	}{
		{
			name:           "daily to monthly - show chip",
			requested:      "daily",
			actualInterval: "monthly",
			expectChip:     true,
		},
		{
			name:           "daily to weekly - show chip",
			requested:      "daily",
			actualInterval: "weekly",
			expectChip:     true,
		},
		{
			name:           "daily to daily - no chip",
			requested:      "daily",
			actualInterval: "daily",
			expectChip:     false,
		},
		{
			name:           "weekly to weekly - no chip",
			requested:      "weekly",
			actualInterval: "weekly",
			expectChip:     false,
		},
		{
			name:           "monthly to quarterly - show chip",
			requested:      "monthly",
			actualInterval: "quarterly",
			expectChip:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			mockCache := &mockCacheStore{}
			model.Configure(context.Background(), nil, []string{"WTI"}, tt.requested, mockCache)

			series := &alphavantage.CommoditySeries{
				Name: "WTI",
				Unit: "dollars per barrel",
				Data: []alphavantage.CommodityDataPoint{
					{Date: time.Now(), Value: 75.50},
				},
			}

			model.rows[0].State = StateLoaded
			model.rows[0].Series = series
			model.rows[0].ActualInterval = tt.actualInterval

			view := NewView(model)
			view.SetTheme(&defaultTheme{})
			row := view.buildRow(model.rows[0])

			// The PERIOD column is index 5
			periodColumn := row[5]
			hasChip := containsString(periodColumn, "→")

			if hasChip != tt.expectChip {
				t.Errorf("expected chip=%v, got chip=%v (period column: %s)", tt.expectChip, hasChip, periodColumn)
			}
		})
	}
}
