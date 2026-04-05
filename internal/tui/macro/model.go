// Package macro provides the macroeconomic dashboard TUI model.
package macro

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/cache"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// PanelState represents the loading state of a data panel.
type PanelState int

const (
	// PanelLoading indicates the panel is fetching data.
	PanelLoading PanelState = iota
	// PanelLoaded indicates the panel has successfully loaded data.
	PanelLoaded
	// PanelError indicates the panel encountered an error.
	PanelError
)

// String returns the string representation of PanelState.
func (s PanelState) String() string {
	switch s {
	case PanelLoading:
		return "Loading"
	case PanelLoaded:
		return "Loaded"
	case PanelError:
		return "Error"
	default:
		return "Unknown"
	}
}

// GDPData contains GDP-related macroeconomic data.
type GDPData struct {
	RealGDP     string
	GDPChange   string
	PerCapita   string
	Period      string
	LastUpdated time.Time
}

// InflationData contains inflation-related macroeconomic data.
type InflationData struct {
	CPI         string
	CPIYoY      string
	Inflation   string
	Period      string
	LastUpdated time.Time
}

// EmploymentData contains employment-related macroeconomic data.
type EmploymentData struct {
	Unemployment string
	Nonfarm      string
	Trend        string
	Period       string
	LastUpdated  time.Time
}

// RateData contains interest rate data.
type RateData struct {
	FedFundsRate string
	Previous     string
	LastChange   string
	Period       string
	LastUpdated  time.Time
}

// YieldData contains treasury yield data for multiple maturities.
type YieldData struct {
	Yield2Y     string
	Yield5Y     string
	Yield10Y    string
	Yield30Y    string
	Period      string
	LastUpdated time.Time
}

// Panel represents a single macroeconomic data panel.
type Panel struct {
	State PanelState
	Error error
}

// GDPPanel holds GDP-specific data.
type GDPPanel struct {
	Panel
	Data *GDPData
}

// InflationPanel holds inflation-specific data.
type InflationPanel struct {
	Panel
	Data *InflationData
}

// EmploymentPanel holds employment-specific data.
type EmploymentPanel struct {
	Panel
	Data *EmploymentData
}

// RatesPanel holds interest rate data.
type RatesPanel struct {
	Panel
	Data *RateData
}

// YieldsPanel holds treasury yield data.
type YieldsPanel struct {
	Panel
	Data *YieldData
}

// Client defines the interface for fetching macroeconomic data.
// This allows mocking in tests while the actual implementation uses alphavantage.Client.
type Client interface {
	GetRealGDP(ctx context.Context, interval string) ([]alphavantage.MacroDataPoint, error)
	GetRealGDPPerCapita(ctx context.Context) ([]alphavantage.MacroDataPoint, error)
	GetCPI(ctx context.Context, interval string) ([]alphavantage.MacroDataPoint, error)
	GetInflation(ctx context.Context) ([]alphavantage.MacroDataPoint, error)
	GetUnemployment(ctx context.Context) ([]alphavantage.MacroDataPoint, error)
	GetNonfarmPayroll(ctx context.Context) ([]alphavantage.MacroDataPoint, error)
	GetFedFundsRate(ctx context.Context, interval string) ([]alphavantage.MacroDataPoint, error)
	GetTreasuryYield(ctx context.Context, interval, maturity string) ([]alphavantage.MacroDataPoint, error)
}

// CacheStore defines the interface for caching macro data.
type CacheStore interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}

// Model represents the macro view model.
type Model struct {
	// client fetches macroeconomic data from Alpha Vantage.
	client Client
	// cache stores macro data with TTL.
	cache CacheStore
	// ctx is the context for async operations.
	ctx context.Context
	// cancel cancels the context.
	cancel context.CancelFunc

	// Panel data
	gdp        GDPPanel
	inflation  InflationPanel
	employment EmploymentPanel
	rates      RatesPanel
	yields     YieldsPanel

	// Timestamps
	lastUpdate time.Time
	ttl        time.Duration

	// Terminal dimensions
	width, height int
}

// NewModel creates a new macro model.
func NewModel() Model {
	return Model{
		gdp:        GDPPanel{Panel: Panel{State: PanelLoading}},
		inflation:  InflationPanel{Panel: Panel{State: PanelLoading}},
		employment: EmploymentPanel{Panel: Panel{State: PanelLoading}},
		rates:      RatesPanel{Panel: Panel{State: PanelLoading}},
		yields:     YieldsPanel{Panel: Panel{State: PanelLoading}},
		ttl:        6 * time.Hour,
		width:      80,
		height:     24,
	}
}

// Configure sets up the model with dependencies.
// This is called by the main app to inject dependencies.
func (m *Model) Configure(
	ctx context.Context,
	client Client,
	cacheStore CacheStore,
) *Model {
	m.client = client
	m.cache = cacheStore
	m.ctx, m.cancel = context.WithCancel(ctx)
	return m
}

// Init initializes the macro model and returns an initial command.
func (m Model) Init() tea.Cmd {
	return m.fetchAllMacroDataCmd()
}

// Update handles messages and updates the model state.
//
//nolint:gocyclo // complexity is acceptable for handling all message types
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			if msg.Runes[0] == 'r' || msg.Runes[0] == 'R' {
				// Refresh all macro data
				return m, m.fetchAllMacroDataCmd()
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case RefreshMsg:
		return m, m.fetchAllMacroDataCmd()

	case GDPDataMsg:
		m.gdp.State = PanelLoaded
		m.gdp.Data = msg.Data
		m.gdp.Error = nil
		return m, nil

	case GDPErrorMsg:
		m.gdp.State = PanelError
		m.gdp.Error = msg.Err
		return m, nil

	case InflationDataMsg:
		m.inflation.State = PanelLoaded
		m.inflation.Data = msg.Data
		m.inflation.Error = nil
		return m, nil

	case InflationErrorMsg:
		m.inflation.State = PanelError
		m.inflation.Error = msg.Err
		return m, nil

	case EmploymentDataMsg:
		m.employment.State = PanelLoaded
		m.employment.Data = msg.Data
		m.employment.Error = nil
		return m, nil

	case EmploymentErrorMsg:
		m.employment.State = PanelError
		m.employment.Error = msg.Err
		return m, nil

	case RatesDataMsg:
		m.rates.State = PanelLoaded
		m.rates.Data = msg.Data
		m.rates.Error = nil
		return m, nil

	case RatesErrorMsg:
		m.rates.State = PanelError
		m.rates.Error = msg.Err
		return m, nil

	case YieldsDataMsg:
		m.yields.State = PanelLoaded
		m.yields.Data = msg.Data
		m.yields.Error = nil
		return m, nil

	case YieldsErrorMsg:
		m.yields.State = PanelError
		m.yields.Error = msg.Err
		return m, nil
	}

	return m, nil
}

// fetchAllMacroDataCmd returns a command to fetch all macroeconomic data.
func (m Model) fetchAllMacroDataCmd() tea.Cmd {
	return tea.Batch(
		m.fetchGDPCmd(),
		m.fetchInflationCmd(),
		m.fetchEmploymentCmd(),
		m.fetchRatesCmd(),
		m.fetchYieldsCmd(),
	)
}

// fetchGDPCmd returns a command to fetch GDP data.
func (m Model) fetchGDPCmd() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		cacheKey := cache.Key("macro", "gdp")
		if cached, found := m.cache.Get(cacheKey); found {
			if data, ok := cached.(*GDPData); ok {
				return GDPDataMsg{Data: data}
			}
		}

		// Fetch real GDP (quarterly)
		gdpData, err := m.client.GetRealGDP(m.ctx, "quarterly")
		if err != nil {
			return GDPErrorMsg{Err: fmt.Errorf("fetching real GDP: %w", err)}
		}

		// Fetch GDP per capita
		perCapitaData, err := m.client.GetRealGDPPerCapita(m.ctx)
		if err != nil {
			return GDPErrorMsg{Err: fmt.Errorf("fetching GDP per capita: %w", err)}
		}

		data := m.parseGDPData(gdpData, perCapitaData)

		// Cache the result
		m.cache.Set(cacheKey, data, m.ttl)

		return GDPDataMsg{Data: data}
	}
}

// parseGDPData parses GDP data from Alpha Vantage responses.
func (m *Model) parseGDPData(gdp, perCapita []alphavantage.MacroDataPoint) *GDPData {
	data := &GDPData{
		RealGDP:     "--",
		GDPChange:   "--",
		PerCapita:   "--",
		Period:      "--",
		LastUpdated: time.Now(),
	}

	// Parse real GDP (most recent)
	if len(gdp) > 0 {
		data.RealGDP = formatGDPValue(gdp[0].Value)
		data.Period = gdp[0].Date

		// Calculate QoQ change if we have at least 2 data points
		if len(gdp) > 1 {
			current, err1 := alphavantage.ParseFloat(gdp[0].Value)
			previous, err2 := alphavantage.ParseFloat(gdp[1].Value)

			if err1 == nil && err2 == nil && previous != 0 {
				change := ((current - previous) / previous) * 100
				data.GDPChange = formatPercent(change)
			}
		}
	}

	// Parse GDP per capita (most recent)
	if len(perCapita) > 0 {
		data.PerCapita = "$" + perCapita[0].Value
	}

	return data
}

// fetchInflationCmd returns a command to fetch inflation data.
func (m Model) fetchInflationCmd() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		cacheKey := cache.Key("macro", "inflation")
		if cached, found := m.cache.Get(cacheKey); found {
			if data, ok := cached.(*InflationData); ok {
				return InflationDataMsg{Data: data}
			}
		}

		// Fetch CPI (monthly)
		cpiData, err := m.client.GetCPI(m.ctx, "monthly")
		if err != nil {
			return InflationErrorMsg{Err: fmt.Errorf("fetching CPI: %w", err)}
		}

		// Fetch inflation rate
		inflationData, err := m.client.GetInflation(m.ctx)
		if err != nil {
			return InflationErrorMsg{Err: fmt.Errorf("fetching inflation rate: %w", err)}
		}

		data := m.parseInflationData(cpiData, inflationData)

		// Cache the result
		m.cache.Set(cacheKey, data, m.ttl)

		return InflationDataMsg{Data: data}
	}
}

// parseInflationData parses inflation data from Alpha Vantage responses.
func (m *Model) parseInflationData(cpi, inflation []alphavantage.MacroDataPoint) *InflationData {
	data := &InflationData{
		CPI:         "--",
		CPIYoY:      "--",
		Inflation:   "--",
		Period:      "--",
		LastUpdated: time.Now(),
	}

	// Parse CPI (most recent)
	if len(cpi) > 0 {
		data.CPI = cpi[0].Value
		data.Period = cpi[0].Date

		// Calculate YoY change if we have at least 12 data points (1 year)
		if len(cpi) >= 12 {
			current, err1 := alphavantage.ParseFloat(cpi[0].Value)
			yearAgo, err2 := alphavantage.ParseFloat(cpi[11].Value)

			if err1 == nil && err2 == nil && yearAgo != 0 {
				change := ((current - yearAgo) / yearAgo) * 100
				data.CPIYoY = formatPercent(change)
			}
		}
	}

	// Parse inflation rate (most recent)
	if len(inflation) > 0 {
		if rate, err := alphavantage.ParseFloat(inflation[0].Value); err == nil {
			data.Inflation = formatPercent(rate)
		}
	}

	return data
}

// fetchEmploymentCmd returns a command to fetch employment data.
func (m Model) fetchEmploymentCmd() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		cacheKey := cache.Key("macro", "employment")
		if cached, found := m.cache.Get(cacheKey); found {
			if data, ok := cached.(*EmploymentData); ok {
				return EmploymentDataMsg{Data: data}
			}
		}

		// Fetch unemployment rate
		unemploymentData, err := m.client.GetUnemployment(m.ctx)
		if err != nil {
			return EmploymentErrorMsg{Err: fmt.Errorf("fetching unemployment: %w", err)}
		}

		// Fetch nonfarm payroll
		nonfarmData, err := m.client.GetNonfarmPayroll(m.ctx)
		if err != nil {
			return EmploymentErrorMsg{Err: fmt.Errorf("fetching nonfarm payroll: %w", err)}
		}

		data := m.parseEmploymentData(unemploymentData, nonfarmData)

		// Cache the result
		m.cache.Set(cacheKey, data, m.ttl)

		return EmploymentDataMsg{Data: data}
	}
}

// parseEmploymentData parses employment data from Alpha Vantage responses.
func (m *Model) parseEmploymentData(unemployment, nonfarm []alphavantage.MacroDataPoint) *EmploymentData {
	data := &EmploymentData{
		Unemployment: "--",
		Nonfarm:      "--",
		Trend:        "Stable",
		Period:       "--",
		LastUpdated:  time.Now(),
	}

	// Parse unemployment (most recent)
	if len(unemployment) > 0 {
		if rate, err := alphavantage.ParseFloat(unemployment[0].Value); err == nil {
			data.Unemployment = fmt.Sprintf("%.1f%%", rate)
			data.Period = unemployment[0].Date

			// Determine trend based on recent changes
			if len(unemployment) >= 3 {
				current, _ := alphavantage.ParseFloat(unemployment[0].Value)
				prev, _ := alphavantage.ParseFloat(unemployment[1].Value)
				prevPrev, _ := alphavantage.ParseFloat(unemployment[2].Value)

				if current < prev && prev < prevPrev {
					data.Trend = "Improving"
				} else if current > prev && prev > prevPrev {
					data.Trend = "Worsening"
				}
			}
		}
	}

	// Parse nonfarm payroll (most recent)
	if len(nonfarm) > 0 {
		if jobs, err := alphavantage.ParseFloat(nonfarm[0].Value); err == nil {
			data.Nonfarm = fmt.Sprintf("%+dK", int(jobs/1000))
		}
	}

	return data
}

// fetchRatesCmd returns a command to fetch interest rate data.
func (m Model) fetchRatesCmd() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		cacheKey := cache.Key("macro", "rates")
		if cached, found := m.cache.Get(cacheKey); found {
			if data, ok := cached.(*RateData); ok {
				return RatesDataMsg{Data: data}
			}
		}

		// Fetch federal funds rate (monthly)
		ratesData, err := m.client.GetFedFundsRate(m.ctx, "monthly")
		if err != nil {
			return RatesErrorMsg{Err: fmt.Errorf("fetching fed funds rate: %w", err)}
		}

		data := m.parseRatesData(ratesData)

		// Cache the result
		m.cache.Set(cacheKey, data, m.ttl)

		return RatesDataMsg{Data: data}
	}
}

// parseRatesData parses interest rate data from Alpha Vantage responses.
func (m *Model) parseRatesData(rates []alphavantage.MacroDataPoint) *RateData {
	data := &RateData{
		FedFundsRate: "--",
		Previous:     "--",
		LastChange:   "--",
		Period:       "--",
		LastUpdated:  time.Now(),
	}

	// Parse fed funds rate (most recent)
	if len(rates) > 0 {
		if rate, err := alphavantage.ParseFloat(rates[0].Value); err == nil {
			data.FedFundsRate = fmt.Sprintf("%.2f%%", rate)
			data.Period = rates[0].Date
		}

		// Get previous rate
		if len(rates) > 1 {
			if prevRate, err := alphavantage.ParseFloat(rates[1].Value); err == nil {
				data.Previous = fmt.Sprintf("%.2f%%", prevRate)

				// Find when the rate last changed significantly
				for i := 1; i < len(rates); i++ {
					current, _ := alphavantage.ParseFloat(rates[i-1].Value)
					previous, _ := alphavantage.ParseFloat(rates[i].Value)
					if abs(current-previous) > 0.01 { // More than 1 basis point
						data.LastChange = rates[i].Date
						break
					}
				}
			}
		}
	}

	return data
}

// fetchYieldsCmd returns a command to fetch treasury yield data.
func (m Model) fetchYieldsCmd() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		cacheKey := cache.Key("macro", "yields")
		if cached, found := m.cache.Get(cacheKey); found {
			if data, ok := cached.(*YieldData); ok {
				return YieldsDataMsg{Data: data}
			}
		}

		// Fetch yields for different maturities (daily)
		y2y, err := m.client.GetTreasuryYield(m.ctx, "daily", "2year")
		if err != nil {
			return YieldsErrorMsg{Err: fmt.Errorf("fetching 2Y treasury yield: %w", err)}
		}

		y5y, err := m.client.GetTreasuryYield(m.ctx, "daily", "5year")
		if err != nil {
			return YieldsErrorMsg{Err: fmt.Errorf("fetching 5Y treasury yield: %w", err)}
		}

		y10y, err := m.client.GetTreasuryYield(m.ctx, "daily", "10year")
		if err != nil {
			return YieldsErrorMsg{Err: fmt.Errorf("fetching 10Y treasury yield: %w", err)}
		}

		y30y, err := m.client.GetTreasuryYield(m.ctx, "daily", "30year")
		if err != nil {
			return YieldsErrorMsg{Err: fmt.Errorf("fetching 30Y treasury yield: %w", err)}
		}

		data := m.parseYieldsData(y2y, y5y, y10y, y30y)

		// Cache the result
		m.cache.Set(cacheKey, data, m.ttl)

		return YieldsDataMsg{Data: data}
	}
}

// parseYieldsData parses treasury yield data from Alpha Vantage responses.
func (m *Model) parseYieldsData(y2y, y5y, y10y, y30y []alphavantage.MacroDataPoint) *YieldData {
	data := &YieldData{
		Yield2Y:     "--",
		Yield5Y:     "--",
		Yield10Y:    "--",
		Yield30Y:    "--",
		Period:      "--",
		LastUpdated: time.Now(),
	}

	// Use 10Y period as the primary period (most referenced)
	if len(y10y) > 0 {
		data.Period = y10y[0].Date
	}

	// Parse 2Y yield
	if len(y2y) > 0 {
		if yield, err := alphavantage.ParseFloat(y2y[0].Value); err == nil {
			data.Yield2Y = fmt.Sprintf("%.2f%%", yield)
		}
	}

	// Parse 5Y yield
	if len(y5y) > 0 {
		if yield, err := alphavantage.ParseFloat(y5y[0].Value); err == nil {
			data.Yield5Y = fmt.Sprintf("%.2f%%", yield)
		}
	}

	// Parse 10Y yield
	if len(y10y) > 0 {
		if yield, err := alphavantage.ParseFloat(y10y[0].Value); err == nil {
			data.Yield10Y = fmt.Sprintf("%.2f%%", yield)
		}
	}

	// Parse 30Y yield
	if len(y30y) > 0 {
		if yield, err := alphavantage.ParseFloat(y30y[0].Value); err == nil {
			data.Yield30Y = fmt.Sprintf("%.2f%%", yield)
		}
	}

	return data
}

// GetWidth returns the current terminal width.
func (m Model) GetWidth() int {
	return m.width
}

// GetHeight returns the current terminal height.
func (m Model) GetHeight() int {
	return m.height
}

// IsNarrowTerminal returns true if the terminal is too narrow for horizontal layout.
func (m Model) IsNarrowTerminal() bool {
	return m.width < 100
}

// GetLastUpdate returns the last update timestamp.
func (m Model) GetLastUpdate() time.Time {
	return m.lastUpdate
}

// GetTTL returns the data TTL.
func (m Model) GetTTL() time.Duration {
	return m.ttl
}

// KeyBindings returns the keyboard bindings for the macro view.
func (m Model) KeyBindings() []components.KeyBinding {
	return []components.KeyBinding{
		{Key: "r", Description: "Refresh all panels"},
	}
}

// IsDataStale returns true if any panel data is older than TTL.
func (m Model) IsDataStale() bool {
	now := time.Now()

	checkStale := func(lastUpdated time.Time) bool {
		if lastUpdated.IsZero() {
			return true
		}
		return now.Sub(lastUpdated) > m.ttl
	}

	if m.gdp.Data != nil && checkStale(m.gdp.Data.LastUpdated) {
		return true
	}
	if m.inflation.Data != nil && checkStale(m.inflation.Data.LastUpdated) {
		return true
	}
	if m.employment.Data != nil && checkStale(m.employment.Data.LastUpdated) {
		return true
	}
	if m.rates.Data != nil && checkStale(m.rates.Data.LastUpdated) {
		return true
	}
	if m.yields.Data != nil && checkStale(m.yields.Data.LastUpdated) {
		return true
	}

	return false
}

// GetPanel returns a panel by name for testing purposes.
func (m Model) GetPanel(name string) interface{} {
	switch name {
	case "gdp":
		return m.gdp
	case "inflation":
		return m.inflation
	case "employment":
		return m.employment
	case "rates":
		return m.rates
	case "yields":
		return m.yields
	default:
		return nil
	}
}

// formatGDPValue formats a GDP value with appropriate suffix.
func formatGDPValue(value string) string {
	if val, err := alphavantage.ParseFloat(value); err == nil {
		if val >= 1e12 {
			return fmt.Sprintf("$%.2fT", val/1e12)
		} else if val >= 1e9 {
			return fmt.Sprintf("$%.2fB", val/1e9)
		}
		return fmt.Sprintf("$%.2f", val)
	}
	return "--"
}

// formatPercent formats a percentage value with sign.
func formatPercent(value float64) string {
	sign := ""
	if value > 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%.1f%%", sign, value)
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Message types

// RefreshMsg is a message to refresh all macro data.
type RefreshMsg struct{}

// GDPDataMsg is a message when GDP data is loaded.
type GDPDataMsg struct {
	Data *GDPData
}

// GDPErrorMsg is a message when GDP data fetch fails.
type GDPErrorMsg struct {
	Err error
}

// InflationDataMsg is a message when inflation data is loaded.
type InflationDataMsg struct {
	Data *InflationData
}

// InflationErrorMsg is a message when inflation data fetch fails.
type InflationErrorMsg struct {
	Err error
}

// EmploymentDataMsg is a message when employment data is loaded.
type EmploymentDataMsg struct {
	Data *EmploymentData
}

// EmploymentErrorMsg is a message when employment data fetch fails.
type EmploymentErrorMsg struct {
	Err error
}

// RatesDataMsg is a message when interest rate data is loaded.
type RatesDataMsg struct {
	Data *RateData
}

// RatesErrorMsg is a message when interest rate data fetch fails.
type RatesErrorMsg struct {
	Err error
}

// YieldsDataMsg is a message when treasury yield data is loaded.
type YieldsDataMsg struct {
	Data *YieldData
}

// YieldsErrorMsg is a message when treasury yield data fetch fails.
type YieldsErrorMsg struct {
	Err error
}
