// Package commodities provides the commodities dashboard TUI model.
package commodities

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// State represents the model state.
type State int

const (
	// StateLoading is the initial state when data is being fetched.
	StateLoading State = iota
	// StateLoaded is when data has been successfully loaded.
	StateLoaded
	// StateCached is when data is loaded from cache.
	StateCached
	// StateError is when there was an error loading data.
	StateError
)

// String returns the string representation of the State.
func (s State) String() string {
	switch s {
	case StateLoading:
		return "Loading"
	case StateLoaded:
		return "Loaded"
	case StateCached:
		return "Cached"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// CommodityInfo holds metadata for a commodity.
type CommodityInfo struct {
	Symbol      string
	Name        string
	Function    alphavantage.CommodityFunction
	Description string
}

// RowData represents the data for a single commodity row in the table.
type RowData struct {
	CommodityInfo
	State      State
	Series     *alphavantage.CommoditySeries
	Error      error
	LastUpdate time.Time
}

// Model represents the commodities dashboard view model.
type Model struct {
	client        *alphavantage.Client
	watchlist     []CommodityInfo
	interval      string
	rows          []RowData
	activeRow     int
	overallState  State
	lastUpdate    time.Time
	ctx           context.Context
	cancel        context.CancelFunc
	width, height int
	cacheStore    cacheStore
}

// cacheStore defines the interface for cache operations.
// Using an interface here allows for cleaner testing.
type cacheStore interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}

// commodityWatchlist maps watchlist symbols to their metadata.
var commodityWatchlist = map[string]CommodityInfo{
	"WTI": {
		Symbol:      "WTI",
		Name:        "Crude Oil WTI",
		Function:    alphavantage.CommodityFunctionWTI,
		Description: "West Texas Intermediate",
	},
	"BRENT": {
		Symbol:      "BRENT",
		Name:        "Crude Oil Brent",
		Function:    alphavantage.CommodityFunctionBrent,
		Description: "Brent Crude Oil",
	},
	"NATURAL_GAS": {
		Symbol:      "NATURAL_GAS",
		Name:        "Natural Gas",
		Function:    alphavantage.CommodityFunctionNaturalGas,
		Description: "Natural Gas Futures",
	},
	"COPPER": {
		Symbol:      "COPPER",
		Name:        "Copper",
		Function:    alphavantage.CommodityFunctionCopper,
		Description: "Copper Futures",
	},
	"ALUMINUM": {
		Symbol:      "ALUMINUM",
		Name:        "Aluminum",
		Function:    alphavantage.CommodityFunctionAluminum,
		Description: "Aluminum Futures",
	},
	"WHEAT": {
		Symbol:      "WHEAT",
		Name:        "Wheat",
		Function:    alphavantage.CommodityFunctionWheat,
		Description: "Wheat Futures",
	},
	"CORN": {
		Symbol:      "CORN",
		Name:        "Corn",
		Function:    alphavantage.CommodityFunctionCorn,
		Description: "Corn Futures",
	},
	"COFFEE": {
		Symbol:      "COFFEE",
		Name:        "Coffee",
		Function:    alphavantage.CommodityFunctionCoffee,
		Description: "Coffee Futures",
	},
	"SUGAR": {
		Symbol:      "SUGAR",
		Name:        "Sugar",
		Function:    alphavantage.CommodityFunctionSugar,
		Description: "Sugar Futures",
	},
	"COTTON": {
		Symbol:      "COTTON",
		Name:        "Cotton",
		Function:    alphavantage.CommodityFunctionCotton,
		Description: "Cotton Futures",
	},
}

// NewModel creates a new commodities model.
func NewModel() *Model {
	return &Model{
		rows:         []RowData{},
		activeRow:    0,
		overallState: StateLoading,
		width:        80,
		height:       24,
		interval:     "daily",
	}
}

// Configure sets up the model with dependencies and configuration.
// This is called by the main app to inject dependencies.
func (m *Model) Configure(
	ctx context.Context,
	client *alphavantage.Client,
	watchlist []string,
	interval string,
	cacheStore cacheStore,
) *Model {
	m.client = client
	m.cacheStore = cacheStore
	m.ctx, m.cancel = context.WithCancel(ctx)

	// Validate and set interval
	if interval != "" {
		normalizedInterval := strings.ToLower(interval)
		switch normalizedInterval {
		case "daily", "weekly", "monthly", "quarterly", "annual":
			m.interval = normalizedInterval
		default:
			m.interval = "daily"
		}
	}

	// Build watchlist from config or use default
	m.watchlist = m.buildWatchlist(watchlist)

	// Initialize rows with loading state
	m.rows = make([]RowData, len(m.watchlist))
	for i, commodity := range m.watchlist {
		m.rows[i] = RowData{
			CommodityInfo: commodity,
			State:         StateLoading,
		}
	}

	return m
}

// buildWatchlist constructs the commodity watchlist from configuration symbols.
// Returns all available commodities if watchlist is empty.
func (m *Model) buildWatchlist(symbols []string) []CommodityInfo {
	if len(symbols) == 0 {
		return m.defaultWatchlist()
	}

	watchlist := make([]CommodityInfo, 0, len(symbols))
	for _, symbol := range symbols {
		commodity, ok := commodityWatchlist[strings.ToUpper(symbol)]
		if ok {
			watchlist = append(watchlist, commodity)
		}
	}

	if len(watchlist) == 0 {
		return m.defaultWatchlist()
	}

	return watchlist
}

// defaultWatchlist returns the default commodity watchlist.
func (m *Model) defaultWatchlist() []CommodityInfo {
	return []CommodityInfo{
		commodityWatchlist["WTI"],
		commodityWatchlist["BRENT"],
		commodityWatchlist["NATURAL_GAS"],
		commodityWatchlist["COPPER"],
		commodityWatchlist["ALUMINUM"],
		commodityWatchlist["WHEAT"],
		commodityWatchlist["CORN"],
		commodityWatchlist["COFFEE"],
		commodityWatchlist["SUGAR"],
		commodityWatchlist["COTTON"],
	}
}

// Init initializes the commodities model and returns an initial command.
// Triggers concurrent fetch for all watchlist commodities.
func (m Model) Init() tea.Cmd {
	if len(m.watchlist) == 0 {
		return nil
	}
	return m.fetchAllCmd()
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case RefreshMsg:
		return m, m.refreshAllCmd()

	case CommodityDataMsg:
		m = m.handleCommodityData(msg)
		m = m.updateOverallState()
		return m, nil

	case CommodityErrorMsg:
		m = m.handleCommodityError(msg)
		m = m.updateOverallState()
		return m, nil
	}

	return m, nil
}

// View renders the commodities view.
func (m Model) View() string {
	return NewView(&m).Render()
}

// handleKeyMsg handles keyboard input messages.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.activeRow > 0 {
			m.activeRow--
		}
		return m, nil

	case tea.KeyDown:
		if m.activeRow < len(m.rows)-1 {
			m.activeRow++
		}
		return m, nil

	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == 'r' {
			return m, m.refreshAllCmd()
		}
	}

	return m, nil
}

// handleCommodityData updates the model with commodity data for a ticker.
func (m Model) handleCommodityData(msg CommodityDataMsg) Model {
	for i, row := range m.rows {
		if row.Symbol == msg.Symbol {
			m.rows[i].State = StateLoaded
			m.rows[i].Series = msg.Series
			m.rows[i].Error = nil
			m.rows[i].LastUpdate = time.Now()

			if msg.FromCache {
				m.rows[i].State = StateCached
			}

			return m
		}
	}
	return m
}

// handleCommodityError updates the model with an error for a commodity.
func (m Model) handleCommodityError(msg CommodityErrorMsg) Model {
	for i, row := range m.rows {
		if row.Symbol == msg.Symbol {
			m.rows[i].State = StateError
			m.rows[i].Error = msg.Err
			return m
		}
	}
	return m
}

// refreshAllCmd returns a command to refresh all commodities.
func (m Model) refreshAllCmd() tea.Cmd {
	for i := range m.rows {
		m.rows[i].State = StateLoading
		m.rows[i].Series = nil
		m.rows[i].Error = nil
	}
	m.overallState = StateLoading
	return m.fetchAllCmd()
}

// fetchAllCmd returns a command that starts fetching all commodities in parallel.
func (m Model) fetchAllCmd() tea.Cmd {
	if len(m.watchlist) == 0 {
		return nil
	}
	cmds := make([]tea.Cmd, len(m.watchlist))
	for i := range m.watchlist {
		cmds[i] = m.fetchCommodityCmd(i)
	}
	return tea.Batch(cmds...)
}

// fetchCommodityCmd fetches a single commodity and returns the result.
func (m Model) fetchCommodityCmd(index int) tea.Cmd {
	if index >= len(m.watchlist) {
		return nil
	}
	commodity := m.watchlist[index]
	return func() tea.Msg {

		// Check cache first
		cacheKey := m.cacheKey(commodity.Symbol, m.interval)
		if m.cacheStore != nil {
			if cached, exists := m.cacheStore.Get(cacheKey); exists {
				if series, ok := cached.(*alphavantage.CommoditySeries); ok {
					return CommodityDataMsg{
						Symbol:    commodity.Symbol,
						Series:    series,
						Index:     index,
						FromCache: true,
					}
				}
			}
		}

		// Fetch from API
		// Fetch from API
		if m.client == nil {
			return CommodityErrorMsg{
				Symbol: commodity.Symbol,
				Err:    fmt.Errorf("client is nil"),
				Index:  index,
			}
		}
		series, err := m.client.GetCommodity(m.ctx, commodity.Function, m.interval)
		if err != nil {
			return CommodityErrorMsg{
				Symbol: commodity.Symbol,
				Err:    err,
				Index:  index,
			}
		}

		// Cache the result
		ttl := m.cacheTTL()
		if m.cacheStore != nil {
			m.cacheStore.Set(cacheKey, series, ttl)
		}

		return CommodityDataMsg{
			Symbol:    commodity.Symbol,
			Series:    series,
			Index:     index,
			FromCache: false,
		}
	}
}

// cacheKey returns the cache key for a commodity and interval.
func (m *Model) cacheKey(symbol, interval string) string {
	return fmt.Sprintf("commodity:%s:%s", symbol, interval)
}

// cacheTTL returns the appropriate cache TTL based on interval.
func (m *Model) cacheTTL() time.Duration {
	switch m.interval {
	case "daily":
		return time.Hour
	default:
		return 6 * time.Hour
	}
}

// updateOverallState updates the overall state based on the status of all rows.
func (m Model) updateOverallState() Model {
	allLoaded := true
	hasLoading := false

	for _, row := range m.rows {
		switch row.State {
		case StateLoading:
			allLoaded = false
			hasLoading = true
		case StateError:
			allLoaded = false
		}
	}

	switch {
	case hasLoading:
		m.overallState = StateLoading
	case allLoaded:
		m.overallState = StateLoaded
	default:
		m.overallState = StateError
	}
	return m
}

// GetRows returns the current rows data.
func (m Model) GetRows() []RowData {
	return m.rows
}

// GetActiveRow returns the index of the active row.
func (m Model) GetActiveRow() int {
	return m.activeRow
}

// GetOverallState returns the overall state of the model.
func (m Model) GetOverallState() State {
	return m.overallState
}

// GetWidth returns the current width.
func (m Model) GetWidth() int {
	return m.width
}

// GetHeight returns the current height.
func (m Model) GetHeight() int {
	return m.height
}

// GetInterval returns the current interval.
func (m Model) GetInterval() string {
	return m.interval
}

// GetLoadedCount returns the number of rows that have finished loading.
func (m Model) GetLoadedCount() int {
	count := 0
	for _, row := range m.rows {
		if row.State == StateLoaded || row.State == StateCached {
			count++
		}
	}
	return count
}

// IsEmpty returns true if the watchlist is empty.
func (m Model) IsEmpty() bool {
	return len(m.rows) == 0
}

// KeyBindings returns the keyboard bindings for the commodities view.
func (m Model) KeyBindings() []components.KeyBinding {
	return []components.KeyBinding{
		{Key: "↑", Description: "Move up"},
		{Key: "↓", Description: "Move down"},
		{Key: "r", Description: "Refresh all commodities"},
	}
}

// RefreshMsg is a message to refresh all commodity data.
type RefreshMsg struct{}

// CommodityDataMsg is a message when commodity data is loaded.
type CommodityDataMsg struct {
	Symbol    string
	Series    *alphavantage.CommoditySeries
	Index     int
	FromCache bool
}

// CommodityErrorMsg is a message when an error occurs fetching commodity data.
type CommodityErrorMsg struct {
	Symbol string
	Err    error
	Index  int
}
