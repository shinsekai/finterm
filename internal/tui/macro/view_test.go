// Package macro provides tests for the macroeconomic dashboard TUI view.
package macro

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TestMacroView_NewView tests creating a new view.
func TestMacroView_NewView(t *testing.T) {
	model := NewModel()
	view := NewView(model)

	if view == nil {
		t.Fatal("Expected non-nil view")
	}

	if view.model.width != 80 {
		t.Errorf("Expected model width 80, got %d", view.model.width)
	}
}

// TestMacroView_SetTheme tests setting a theme.
func TestMacroView_SetTheme(t *testing.T) {
	model := NewModel()
	view := NewView(model)

	theme := &defaultTheme{}
	view.SetTheme(theme)

	if view.theme != theme {
		t.Error("Expected theme to be set")
	}
}

// TestMacroView_Render_WideTerminal tests rendering in wide terminal mode.
func TestMacroView_Render_WideTerminal(t *testing.T) {
	model := NewModel()
	model.width = 120 // Wide terminal

	// Set up mock data
	now := time.Now()
	model.gdp = GDPPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &GDPData{
			RealGDP:     "$22.67T",
			GDPChange:   "+2.1%",
			PerCapita:   "$67,891",
			Period:      "Q3 2025",
			LastUpdated: now,
		},
	}
	model.inflation = InflationPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &InflationData{
			CPI:         "312.23",
			CPIYoY:      "+3.1%",
			Inflation:   "3.0%",
			Period:      "2026-02",
			LastUpdated: now,
		},
	}
	model.employment = EmploymentPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &EmploymentData{
			Unemployment: "3.7%",
			Nonfarm:      "+216K",
			Trend:        "Stable",
			Period:       "2026-02",
			LastUpdated:  now,
		},
	}
	model.rates = RatesPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &RateData{
			FedFundsRate: "5.25%",
			Previous:     "5.25%",
			LastChange:   "Jul 2023",
			Period:       "2026-02",
			LastUpdated:  now,
		},
	}
	model.yields = YieldsPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &YieldData{
			Yield2Y:     "4.62%",
			Yield5Y:     "4.28%",
			Yield10Y:    "4.15%",
			Yield30Y:    "4.32%",
			Period:      "2026-02",
			LastUpdated: now,
		},
	}

	view := NewView(model)
	output := view.Render()

	// Check for expected content
	checkContains(t, output, "MACRO DASHBOARD")
	checkContains(t, output, "GDP")
	checkContains(t, output, "$22.67T")
	checkContains(t, output, "Inflation")
	checkContains(t, output, "312.23")
	checkContains(t, output, "Employment")
	checkContains(t, output, "3.7%")
	checkContains(t, output, "Interest Rates")
	checkContains(t, output, "5.25%")
	checkContains(t, output, "Treasury Yields")
	checkContains(t, output, "4.15%")
}

// TestMacroView_Render_NarrowTerminal tests rendering in narrow terminal mode.
func TestMacroView_Render_NarrowTerminal(t *testing.T) {
	model := NewModel()
	model.width = 80 // Narrow terminal

	// Set up mock data
	now := time.Now()
	model.gdp = GDPPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &GDPData{
			RealGDP:     "$22.67T",
			Period:      "Q3 2025",
			LastUpdated: now,
		},
	}

	view := NewView(model)
	output := view.Render()

	// Check for expected content
	checkContains(t, output, "MACRO DASHBOARD")
	checkContains(t, output, "GDP")
	checkContains(t, output, "$22.67T")
}

// TestMacroView_Render_AllPanels tests that all panels render with data.
func TestMacroView_Render_AllPanels(t *testing.T) {
	model := NewModel()
	model.width = 120

	// Set up all panels with data
	now := time.Now()
	model.gdp = GDPPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &GDPData{
			RealGDP:     "$22.67T",
			GDPChange:   "+2.1%",
			PerCapita:   "$67,891",
			Period:      "Q3 2025",
			LastUpdated: now,
		},
	}
	model.inflation = InflationPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &InflationData{
			CPI:         "312.23",
			CPIYoY:      "+3.1%",
			Inflation:   "3.0%",
			Period:      "2026-02",
			LastUpdated: now,
		},
	}
	model.employment = EmploymentPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &EmploymentData{
			Unemployment: "3.7%",
			Nonfarm:      "+216K",
			Trend:        "Stable",
			Period:       "2026-02",
			LastUpdated:  now,
		},
	}
	model.rates = RatesPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &RateData{
			FedFundsRate: "5.25%",
			Previous:     "5.25%",
			LastChange:   "Jul 2023",
			Period:       "2026-02",
			LastUpdated:  now,
		},
	}
	model.yields = YieldsPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &YieldData{
			Yield2Y:     "4.62%",
			Yield5Y:     "4.28%",
			Yield10Y:    "4.15%",
			Yield30Y:    "4.32%",
			Period:      "2026-02",
			LastUpdated: now,
		},
	}

	view := NewView(model)
	output := view.Render()

	// Verify all panels are present
	panelNames := []string{"GDP", "Inflation", "Employment", "Interest Rates", "Treasury Yields"}
	for _, name := range panelNames {
		if !contains(output, name) {
			t.Errorf("Expected panel '%s' to be present in output", name)
		}
	}

	// Verify data values
	dataValues := []string{"$22.67T", "312.23", "3.7%", "5.25%", "4.15%"}
	for _, value := range dataValues {
		if !contains(output, value) {
			t.Errorf("Expected value '%s' to be present in output", value)
		}
	}
}

// TestMacroView_Render_LoadingState tests rendering loading state with spinners.
func TestMacroView_Render_LoadingState(t *testing.T) {
	model := NewModel()
	model.width = 120

	// Set panels to loading
	model.gdp = GDPPanel{Panel: Panel{State: PanelLoading}}
	model.inflation = InflationPanel{Panel: Panel{State: PanelLoading}}
	model.employment = EmploymentPanel{Panel: Panel{State: PanelLoading}}
	model.rates = RatesPanel{Panel: Panel{State: PanelLoading}}
	model.yields = YieldsPanel{Panel: Panel{State: PanelLoading}}

	view := NewView(model)
	output := view.Render()

	// Check for loading indicators
	checkContains(t, output, "MACRO DASHBOARD")
	checkContains(t, output, "GDP")
	checkContains(t, output, "Inflation")
	checkContains(t, output, "Employment")
	checkContains(t, output, "Interest Rates")
	checkContains(t, output, "Treasury Yields")
}

// TestMacroView_Render_ErrorState tests rendering error state.
func TestMacroView_Render_ErrorState(t *testing.T) {
	model := NewModel()
	model.width = 120

	// Set one panel to error
	model.gdp = GDPPanel{
		Panel: Panel{
			State: PanelError,
			Error: &mockError{msg: "API rate limit exceeded"},
		},
	}
	model.inflation = InflationPanel{
		Panel: Panel{State: PanelLoaded},
		Data:  &InflationData{CPI: "312.23"},
	}
	model.employment = EmploymentPanel{
		Panel: Panel{State: PanelLoaded},
		Data:  &EmploymentData{Unemployment: "3.7%"},
	}
	model.rates = RatesPanel{
		Panel: Panel{State: PanelLoaded},
		Data:  &RateData{FedFundsRate: "5.25%"},
	}
	model.yields = YieldsPanel{
		Panel: Panel{State: PanelLoaded},
		Data:  &YieldData{Yield10Y: "4.15%"},
	}

	view := NewView(model)
	output := view.Render()

	// Check for error indicator
	checkContains(t, output, "Error loading data")
}

// TestMacroView_Render_StaleIndicator tests stale data indicator.
func TestMacroView_Render_StaleIndicator(t *testing.T) {
	model := NewModel()
	model.width = 120

	// Set up old data (older than TTL)
	oldTime := time.Now().Add(-7 * time.Hour) // Older than 6h TTL
	model.gdp = GDPPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &GDPData{
			RealGDP:     "$22.67T",
			Period:      "Q3 2025",
			LastUpdated: oldTime,
		},
	}
	model.inflation = InflationPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &InflationData{
			CPI:         "312.23",
			Period:      "2026-02",
			LastUpdated: oldTime,
		},
	}
	model.employment = EmploymentPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &EmploymentData{
			Unemployment: "3.7%",
			Period:       "2026-02",
			LastUpdated:  oldTime,
		},
	}
	model.rates = RatesPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &RateData{
			FedFundsRate: "5.25%",
			Period:       "2026-02",
			LastUpdated:  oldTime,
		},
	}
	model.yields = YieldsPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &YieldData{
			Yield10Y:    "4.15%",
			Period:      "2026-02",
			LastUpdated: oldTime,
		},
	}

	view := NewView(model)
	output := view.Render()

	// Check for stale indicator
	checkContains(t, output, "Data may be stale")
}

// TestMacroView_Render_NoData tests rendering when no data is available.
func TestMacroView_Render_NoData(t *testing.T) {
	model := NewModel()
	model.width = 120

	// Set panels to loaded with nil data
	model.gdp = GDPPanel{Panel: Panel{State: PanelLoaded}, Data: nil}
	model.inflation = InflationPanel{Panel: Panel{State: PanelLoaded}, Data: nil}
	model.employment = EmploymentPanel{Panel: Panel{State: PanelLoaded}, Data: nil}
	model.rates = RatesPanel{Panel: Panel{State: PanelLoaded}, Data: nil}
	model.yields = YieldsPanel{Panel: Panel{State: PanelLoaded}, Data: nil}

	view := NewView(model)
	output := view.Render()

	// Check for no data message
	checkContains(t, output, "No data available")
}

// TestMacroView_Render_WithTheme tests rendering with custom theme.
func TestMacroView_Render_WithTheme(t *testing.T) {
	model := NewModel()
	model.width = 120

	// Set up mock data
	now := time.Now()
	model.gdp = GDPPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &GDPData{
			RealGDP:     "$22.67T",
			Period:      "Q3 2025",
			LastUpdated: now,
		},
	}

	view := NewView(model)
	view.SetTheme(&customTheme{})
	output := view.Render()

	// Check for expected content
	checkContains(t, output, "MACRO DASHBOARD")
	checkContains(t, output, "GDP")
}

// TestMacroView_Render_Footer tests footer rendering.
func TestMacroView_Render_Footer(t *testing.T) {
	model := NewModel()
	model.width = 120

	// Set up mock data
	now := time.Now()
	model.gdp = GDPPanel{
		Panel: Panel{State: PanelLoaded},
		Data: &GDPData{
			RealGDP:     "$22.67T",
			Period:      "Q3 2025",
			LastUpdated: now,
		},
	}

	view := NewView(model)
	output := view.Render()

	// Check for footer elements
	checkContains(t, output, "Last updated")
	checkContains(t, output, "TTL")
	checkContains(t, output, "r: refresh")
	checkContains(t, output, "q: quit")
}

// mockError is a mock error for testing.
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

// customTheme is a custom theme for testing.
type customTheme struct{}

func (t *customTheme) Title() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func (t *customTheme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Faint(true)
}

func (t *customTheme) Warning() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")).Bold(true)
}

func (t *customTheme) Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Bold(true)
}

func (t *customTheme) Box() lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder())
}

func (t *customTheme) BoxBorder() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("gray"))
}

func (t *customTheme) BoxTitle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")).Bold(true)
}

func (t *customTheme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))
}

func (t *customTheme) SpinnerText() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("gray"))
}

// checkContains is a helper to check if a string contains a substring.
func checkContains(t *testing.T, s, substr string) {
	t.Helper()
	if !contains(s, substr) {
		t.Errorf("Expected output to contain %q", substr)
	}
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOfSubstring(s, substr) >= 0
}

// indexOfSubstring finds the index of a substring.
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
