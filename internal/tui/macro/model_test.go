// Package macro provides tests for the macroeconomic dashboard TUI model.
package macro

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/charmbracelet/bubbletea"

	"github.com/owner/finterm/internal/alphavantage"
	"github.com/owner/finterm/internal/cache"
)

// mockMacroClient is a mock implementation of Client for testing.
type mockMacroClient struct {
	gdpData          []alphavantage.MacroDataPoint
	gdpPerCapita     []alphavantage.MacroDataPoint
	cpiData          []alphavantage.MacroDataPoint
	inflationData    []alphavantage.MacroDataPoint
	unemployment     []alphavantage.MacroDataPoint
	nonfarmPayroll   []alphavantage.MacroDataPoint
	fedFundsRate     []alphavantage.MacroDataPoint
	treasuryYield2Y  []alphavantage.MacroDataPoint
	treasuryYield5Y  []alphavantage.MacroDataPoint
	treasuryYield10Y []alphavantage.MacroDataPoint
	treasuryYield30Y []alphavantage.MacroDataPoint
	err              error
}

func (m *mockMacroClient) GetRealGDP(_ context.Context, _ string) ([]alphavantage.MacroDataPoint, error) {
	return m.gdpData, m.err
}

func (m *mockMacroClient) GetRealGDPPerCapita(_ context.Context) ([]alphavantage.MacroDataPoint, error) {
	return m.gdpPerCapita, m.err
}

func (m *mockMacroClient) GetCPI(_ context.Context, _ string) ([]alphavantage.MacroDataPoint, error) {
	return m.cpiData, m.err
}

func (m *mockMacroClient) GetInflation(_ context.Context) ([]alphavantage.MacroDataPoint, error) {
	return m.inflationData, m.err
}

func (m *mockMacroClient) GetUnemployment(_ context.Context) ([]alphavantage.MacroDataPoint, error) {
	return m.unemployment, m.err
}

func (m *mockMacroClient) GetNonfarmPayroll(_ context.Context) ([]alphavantage.MacroDataPoint, error) {
	return m.nonfarmPayroll, m.err
}

func (m *mockMacroClient) GetFedFundsRate(_ context.Context, _ string) ([]alphavantage.MacroDataPoint, error) {
	return m.fedFundsRate, m.err
}

func (m *mockMacroClient) GetTreasuryYield(_ context.Context, _, maturity string) ([]alphavantage.MacroDataPoint, error) {
	switch maturity {
	case "2year":
		return m.treasuryYield2Y, m.err
	case "5year":
		return m.treasuryYield5Y, m.err
	case "10year":
		return m.treasuryYield10Y, m.err
	case "30year":
		return m.treasuryYield30Y, m.err
	default:
		return nil, errors.New("unknown maturity")
	}
}

// TestMacroModel_Init_FetchesAll tests that Init returns commands to fetch all macro data.
func TestMacroModel_Init_FetchesAll(t *testing.T) {
	c := &cache.Store{}
	client := &mockMacroClient{
		gdpData:          []alphavantage.MacroDataPoint{{Date: "2025-Q3", Value: "22670000000000"}},
		gdpPerCapita:     []alphavantage.MacroDataPoint{{Date: "2025", Value: "67891"}},
		cpiData:          []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "312.23"}},
		inflationData:    []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "3.0"}},
		unemployment:     []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "3.7"}},
		nonfarmPayroll:   []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "216000"}},
		fedFundsRate:     []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "5.25"}},
		treasuryYield2Y:  []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "4.62"}},
		treasuryYield5Y:  []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "4.28"}},
		treasuryYield10Y: []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "4.15"}},
		treasuryYield30Y: []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "4.32"}},
	}

	model := NewModel()
	model.Configure(context.Background(), client, c)

	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init should return a command")
	}
}

// TestMacroModel_Update_GDPData tests updating model with GDP data.
func TestMacroModel_Update_GDPData(t *testing.T) {
	c := &cache.Store{}
	client := &mockMacroClient{}
	model := NewModel()
	model.Configure(context.Background(), client, c)

	// Send GDP data message
	msg := GDPDataMsg{
		Data: &GDPData{
			RealGDP:     "$22.67T",
			GDPChange:   "+2.1%",
			PerCapita:   "$67,891",
			Period:      "Q3 2025",
			LastUpdated: time.Now(),
		},
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.gdp.State != PanelLoaded {
		t.Errorf("Expected PanelLoaded, got %v", m.gdp.State)
	}

	if m.gdp.Data == nil {
		t.Error("Expected GDP data to be set")
	}

	if m.gdp.Data.RealGDP != "$22.67T" {
		t.Errorf("Expected RealGDP $22.67T, got %s", m.gdp.Data.RealGDP)
	}
}

// TestMacroModel_Update_PartialLoad tests handling partial loading state.
func TestMacroModel_Update_PartialLoad(t *testing.T) {
	c := &cache.Store{}
	client := &mockMacroClient{}
	model := NewModel()
	model.Configure(context.Background(), client, c)

	// Update GDP panel
	msg1 := GDPDataMsg{
		Data: &GDPData{
			RealGDP:     "$22.67T",
			Period:      "Q3 2025",
			LastUpdated: time.Now(),
		},
	}
	newModel, _ := model.Update(msg1)
	model = newModel.(Model)

	// Inflation panel should still be loading
	if model.inflation.State != PanelLoading {
		t.Errorf("Expected PanelLoading for inflation, got %v", model.inflation.State)
	}

	// GDP panel should be loaded
	if model.gdp.State != PanelLoaded {
		t.Errorf("Expected PanelLoaded for GDP, got %v", model.gdp.State)
	}
}

// TestMacroModel_Update_WindowSize tests handling window size updates.
func TestMacroModel_Update_WindowSize(t *testing.T) {
	c := &cache.Store{}
	client := &mockMacroClient{}
	model := NewModel()
	model.Configure(context.Background(), client, c)

	msg := tea.WindowSizeMsg{Width: 120, Height: 30}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}

	if m.height != 30 {
		t.Errorf("Expected height 30, got %d", m.height)
	}
}

// TestMacroModel_Update_Refresh tests handling refresh message.
func TestMacroModel_Update_Refresh(t *testing.T) {
	c := &cache.Store{}
	client := &mockMacroClient{}
	model := NewModel()
	model.Configure(context.Background(), client, c)

	msg := RefreshMsg{}
	newModel, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Expected refresh command")
	}

	// Model should be returned unchanged
	m := newModel.(Model)
	if m.gdp.State != PanelLoading {
		t.Errorf("Expected PanelLoading after refresh, got %v", m.gdp.State)
	}
}

// TestMacroModel_Update_RefreshKey tests handling refresh key press.
func TestMacroModel_Update_RefreshKey(t *testing.T) {
	c := &cache.Store{}
	client := &mockMacroClient{}
	model := NewModel()
	model.Configure(context.Background(), client, c)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Expected refresh command")
	}

	m := newModel.(Model)
	if m.gdp.State != PanelLoading {
		t.Errorf("Expected PanelLoading after refresh, got %v", m.gdp.State)
	}
}

// TestMacroModel_ErrorHandling tests handling error messages.
func TestMacroModel_ErrorHandling(t *testing.T) {
	c := &cache.Store{}
	client := &mockMacroClient{}
	model := NewModel()
	model.Configure(context.Background(), client, c)

	testErr := errors.New("API rate limit exceeded")

	msg := GDPErrorMsg{Err: testErr}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.gdp.State != PanelError {
		t.Errorf("Expected PanelError, got %v", m.gdp.State)
	}

	if m.gdp.Error == nil {
		t.Error("Expected error to be set")
	}

	if m.gdp.Error.Error() != "API rate limit exceeded" {
		t.Errorf("Expected error 'API rate limit exceeded', got %v", m.gdp.Error)
	}
}

// TestMacroModel_IsNarrowTerminal tests terminal width detection.
func TestMacroModel_IsNarrowTerminal(t *testing.T) {
	model := NewModel()

	// Wide terminal
	model.width = 120
	if model.IsNarrowTerminal() {
		t.Error("Expected false for wide terminal")
	}

	// Narrow terminal
	model.width = 80
	if !model.IsNarrowTerminal() {
		t.Error("Expected true for narrow terminal")
	}

	// Borderline case
	model.width = 99
	if !model.IsNarrowTerminal() {
		t.Error("Expected true for 99 width")
	}

	// Borderline case
	model.width = 100
	if model.IsNarrowTerminal() {
		t.Error("Expected false for 100 width")
	}
}

// TestMacroModel_IsDataStale tests stale data detection.
func TestMacroModel_IsDataStale(t *testing.T) {
	c := &cache.Store{}
	client := &mockMacroClient{}
	model := NewModel()
	model.Configure(context.Background(), client, c)

	// No data should NOT be considered stale (it's just uninitialized)
	if model.IsDataStale() {
		t.Error("Expected not stale when no data")
	}

	// Fresh data
	now := time.Now()
	model.gdp.Data = &GDPData{
		RealGDP:     "$22.67T",
		LastUpdated: now,
	}
	model.inflation.Data = &InflationData{
		CPI:         "312.23",
		LastUpdated: now,
	}
	model.employment.Data = &EmploymentData{
		Unemployment: "3.7%",
		LastUpdated:  now,
	}
	model.rates.Data = &RateData{
		FedFundsRate: "5.25%",
		LastUpdated:  now,
	}
	model.yields.Data = &YieldData{
		Yield10Y:    "4.15%",
		LastUpdated: now,
	}

	if model.IsDataStale() {
		t.Error("Expected fresh data not to be stale")
	}

	// Stale data (older than TTL)
	oldTime := now.Add(-7 * time.Hour) // Older than 6h TTL
	model.gdp.Data.LastUpdated = oldTime

	if !model.IsDataStale() {
		t.Error("Expected stale data when GDP is old")
	}
}

// TestMacroModel_parseGDPData tests GDP data parsing.
func TestMacroModel_parseGDPData(t *testing.T) {
	model := NewModel()

	gdpData := []alphavantage.MacroDataPoint{
		{Date: "2025-Q3", Value: "22670000000000"},
		{Date: "2025-Q2", Value: "22200000000000"},
	}
	perCapitaData := []alphavantage.MacroDataPoint{
		{Date: "2025", Value: "67891"},
	}

	result := model.parseGDPData(gdpData, perCapitaData)

	if result.RealGDP != "$22.67T" {
		t.Errorf("Expected RealGDP $22.67T, got %s", result.RealGDP)
	}

	if result.PerCapita != "$67891" {
		t.Errorf("Expected PerCapita $67891, got %s", result.PerCapita)
	}

	if result.Period != "2025-Q3" {
		t.Errorf("Expected Period 2025-Q3, got %s", result.Period)
	}

	// Check for positive QoQ change
	if result.GDPChange[0] != '+' {
		t.Errorf("Expected positive GDP change, got %s", result.GDPChange)
	}
}

// TestMacroModel_parseInflationData tests inflation data parsing.
func TestMacroModel_parseInflationData(t *testing.T) {
	model := NewModel()

	cpiData := []alphavantage.MacroDataPoint{
		{Date: "2026-02", Value: "312.23"},
	}
	inflationData := []alphavantage.MacroDataPoint{
		{Date: "2026-02", Value: "3.0"},
	}

	result := model.parseInflationData(cpiData, inflationData)

	if result.CPI != "312.23" {
		t.Errorf("Expected CPI 312.23, got %s", result.CPI)
	}

	if result.Inflation != "+3.0%" {
		t.Errorf("Expected Inflation +3.0%%, got %s", result.Inflation)
	}

	if result.Period != "2026-02" {
		t.Errorf("Expected Period 2026-02, got %s", result.Period)
	}
}

// TestMacroModel_parseEmploymentData tests employment data parsing.
func TestMacroModel_parseEmploymentData(t *testing.T) {
	model := NewModel()

	unemploymentData := []alphavantage.MacroDataPoint{
		{Date: "2026-02", Value: "3.7"},
	}
	nonfarmData := []alphavantage.MacroDataPoint{
		{Date: "2026-02", Value: "216000"},
	}

	result := model.parseEmploymentData(unemploymentData, nonfarmData)

	if result.Unemployment != "3.7%" {
		t.Errorf("Expected Unemployment 3.7%%, got %s", result.Unemployment)
	}

	if result.Nonfarm != "+216K" {
		t.Errorf("Expected Nonfarm +216K, got %s", result.Nonfarm)
	}

	if result.Period != "2026-02" {
		t.Errorf("Expected Period 2026-02, got %s", result.Period)
	}
}

// TestMacroModel_parseRatesData tests interest rate data parsing.
func TestMacroModel_parseRatesData(t *testing.T) {
	model := NewModel()

	ratesData := []alphavantage.MacroDataPoint{
		{Date: "2026-02", Value: "5.25"},
		{Date: "2026-01", Value: "5.25"},
		{Date: "2025-12", Value: "5.25"},
	}

	result := model.parseRatesData(ratesData)

	if result.FedFundsRate != "5.25%" {
		t.Errorf("Expected FedFundsRate 5.25%%, got %s", result.FedFundsRate)
	}

	if result.Previous != "5.25%" {
		t.Errorf("Expected Previous 5.25%%, got %s", result.Previous)
	}

	if result.Period != "2026-02" {
		t.Errorf("Expected Period 2026-02, got %s", result.Period)
	}
}

// TestMacroModel_parseYieldsData tests treasury yield data parsing.
func TestMacroModel_parseYieldsData(t *testing.T) {
	model := NewModel()

	yield2Y := []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "4.62"}}
	yield5Y := []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "4.28"}}
	yield10Y := []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "4.15"}}
	yield30Y := []alphavantage.MacroDataPoint{{Date: "2026-02", Value: "4.32"}}

	result := model.parseYieldsData(yield2Y, yield5Y, yield10Y, yield30Y)

	if result.Yield2Y != "4.62%" {
		t.Errorf("Expected Yield2Y 4.62%%, got %s", result.Yield2Y)
	}

	if result.Yield5Y != "4.28%" {
		t.Errorf("Expected Yield5Y 4.28%%, got %s", result.Yield5Y)
	}

	if result.Yield10Y != "4.15%" {
		t.Errorf("Expected Yield10Y 4.15%%, got %s", result.Yield10Y)
	}

	if result.Yield30Y != "4.32%" {
		t.Errorf("Expected Yield30Y 4.32%%, got %s", result.Yield30Y)
	}
}

// TestMacroModel_PanelState_String tests PanelState string representation.
func TestMacroModel_PanelState_String(t *testing.T) {
	tests := []struct {
		state    PanelState
		expected string
	}{
		{PanelLoading, "Loading"},
		{PanelLoaded, "Loaded"},
		{PanelError, "Error"},
		{PanelState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("PanelState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestMacroModel_GetPanel tests retrieving panels by name.
func TestMacroModel_GetPanel(t *testing.T) {
	c := &cache.Store{}
	client := &mockMacroClient{}
	model := NewModel()
	model.Configure(context.Background(), client, c)

	// Set some data
	model.gdp.Data = &GDPData{RealGDP: "$22.67T"}
	model.inflation.Data = &InflationData{CPI: "312.23"}

	tests := []struct {
		name string
	}{
		{"gdp"},
		{"inflation"},
		{"employment"},
		{"rates"},
		{"yields"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panel := model.GetPanel(tt.name)
			if panel == nil && tt.name != "unknown" {
				t.Errorf("Expected non-nil panel for %s", tt.name)
			}
			if panel != nil && tt.name == "unknown" {
				t.Error("Expected nil panel for unknown name")
			}
		})
	}
}
