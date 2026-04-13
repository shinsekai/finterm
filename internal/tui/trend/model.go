// Package trend provides the trend following TUI model.
package trend

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shinsekai/finterm/internal/config"
	trenddomain "github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
	"github.com/shinsekai/finterm/internal/tui/components"
)

// Engine defines the interface for trend analysis engines.
// This allows mocking in tests while the actual implementation uses trenddomain.Engine.
type Engine interface {
	AnalyzeWithSymbolDetection(ctx context.Context, symbol string) (*trenddomain.Result, error)
}

// State represents the model state.
type State int

const (
	// StateLoading is the initial state when data is being fetched.
	StateLoading State = iota
	// StateLoaded is when data has been successfully loaded.
	StateLoaded
	// StateCached is when data is loaded from cache (offline mode).
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

// RowData represents the data for a single ticker row in the table.
type RowData struct {
	// Symbol is the ticker symbol.
	Symbol string
	// State is the current state of this ticker's data.
	State State
	// Result contains the trend analysis result (valid when State is StateLoaded).
	Result *trenddomain.Result
	// Error contains any error that occurred (valid when State is StateError).
	Error error
}

// Model represents the trend following view model.
type Model struct {
	// engine performs trend analysis for symbols.
	engine Engine
	// watchlist is the list of symbols to analyze.
	watchlist []string
	// detector determines asset class for symbols.
	detector *indicators.AssetClassDetector
	// rows contains the data for each ticker in the watchlist.
	rows []RowData
	// activeRow is the index of the currently selected row.
	activeRow int
	// overallState is the overall state of the model.
	overallState State
	// ctx is the context for async operations.
	ctx context.Context
	// cancel cancels the context.
	cancel context.CancelFunc
	// width and height are the terminal dimensions.
	width, height int
	// equitiesCount is the number of equity symbols in the watchlist.
	equitiesCount int
}

// NewModel creates a new trend model.
func NewModel() *Model {
	return &Model{
		watchlist:    []string{},
		rows:         []RowData{},
		activeRow:    0,
		overallState: StateLoading,
		width:        80,
		height:       24,
	}
}

// Configure sets up the model with dependencies and configuration.
// This is called by the main app to inject dependencies.
func (m *Model) Configure(
	ctx context.Context,
	engine Engine,
	watchlist *config.WatchlistConfig,
	detector *indicators.AssetClassDetector,
) *Model {
	m.engine = engine
	m.detector = detector
	m.ctx, m.cancel = context.WithCancel(ctx)

	// Track where equities end and crypto begins
	m.equitiesCount = len(watchlist.Equities)

	// Build combined watchlist from equities and crypto
	m.watchlist = append([]string{}, watchlist.Equities...)
	m.watchlist = append(m.watchlist, watchlist.Crypto...)

	// Initialize rows with loading state
	m.rows = make([]RowData, len(m.watchlist))
	for i, symbol := range m.watchlist {
		m.rows[i] = RowData{
			Symbol: symbol,
			State:  StateLoading,
		}
	}

	return m
}

// Init initializes the trend model and returns an initial command.
// Triggers concurrent fetch for all watchlist tickers.
func (m Model) Init() tea.Cmd {
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
		// Refresh all tickers
		return m, m.refreshAllCmd()

	case TrendDataMsg:
		// Handle data for a single ticker
		m = m.handleTrendData(msg)
		// Trigger next ticker fetch
		nextIndex := msg.Index + 1
		if nextIndex < len(m.watchlist) {
			return m, m.fetchTickerCmd(nextIndex)
		}
		m = m.updateOverallState()
		return m, nil

	case TrendErrorMsg:
		// Handle error for a single ticker
		m = m.handleTrendError(msg)
		// Continue to next ticker even on error
		nextIndex := msg.Index + 1
		if nextIndex < len(m.watchlist) {
			return m, m.fetchTickerCmd(nextIndex)
		}
		m = m.updateOverallState()
		return m, nil
	}

	return m, nil
}

// View renders the trend view.
func (m Model) View() string {
	return NewView(m).Render()
}

// handleKeyMsg handles keyboard input messages.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
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
		// Check if 'r' key is pressed
		if len(msg.Runes) == 1 && msg.Runes[0] == 'r' {
			// Refresh all tickers
			return m, m.refreshAllCmd()
		}
	}

	return m, nil
}

// handleTrendData updates the model with trend analysis result for a ticker.
func (m Model) handleTrendData(msg TrendDataMsg) Model {
	for i, row := range m.rows {
		if row.Symbol == msg.Symbol {
			m.rows[i].State = StateLoaded
			m.rows[i].Result = msg.Result
			m.rows[i].Error = nil
			return m
		}
	}
	return m
}

// handleTrendError updates the model with an error for a ticker.
func (m Model) handleTrendError(msg TrendErrorMsg) Model {
	for i, row := range m.rows {
		if row.Symbol == msg.Symbol {
			m.rows[i].State = StateError
			m.rows[i].Error = msg.Err
			return m
		}
	}
	return m
}

// refreshAllCmd returns a command to refresh all tickers.
func (m Model) refreshAllCmd() tea.Cmd {
	// Reset all rows to loading state
	for i := range m.rows {
		m.rows[i].State = StateLoading
		m.rows[i].Result = nil
		m.rows[i].Error = nil
	}
	m.overallState = StateLoading
	return m.fetchAllCmd()
}

// fetchAllCmd returns a command that starts fetching tickers sequentially.
// Each result/error triggers the next fetch via fetchTickerCmd.
func (m Model) fetchAllCmd() tea.Cmd {
	if len(m.watchlist) == 0 {
		return nil
	}
	// Start with the first ticker
	return m.fetchTickerCmd(0)
}

// fetchTickerCmd fetches a single ticker and returns the result.
func (m Model) fetchTickerCmd(index int) tea.Cmd {
	if index >= len(m.watchlist) {
		return nil
	}
	symbol := m.watchlist[index]
	return func() tea.Msg {
		result, err := m.engine.AnalyzeWithSymbolDetection(m.ctx, symbol)
		if err != nil {
			return TrendErrorMsg{
				Symbol: symbol,
				Err:    err,
				Index:  index,
			}
		}
		return TrendDataMsg{
			Symbol: symbol,
			Result: result,
			Index:  index,
		}
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

// GetCryptoStartIndex returns the index of the first crypto row.
// If there are no crypto symbols, returns len(m.rows).
func (m Model) GetCryptoStartIndex() int {
	return m.equitiesCount
}

// GetLoadedCount returns the number of rows that have finished loading
// (either loaded or cached state).
func (m Model) GetLoadedCount() int {
	count := 0
	for _, row := range m.rows {
		if row.State == StateLoaded || row.State == StateCached {
			count++
		}
	}
	return count
}

// GetSignalCounts returns the count of bullish, bearish, and neutral signals
// across all loaded rows. Loading and error rows are ignored.
func (m Model) GetSignalCounts() (bullish, bearish, neutral int) {
	var b, br, n int
	for _, row := range m.rows {
		if row.State != StateLoaded && row.State != StateCached {
			continue
		}
		if row.Result == nil {
			continue
		}
		switch row.Result.Signal {
		case trenddomain.Bullish:
			b++
		case trenddomain.Bearish:
			br++
		default:
			n++
		}
	}
	return b, br, n
}

// GetBlitzCounts returns the count of LONG, SHORT, and HOLD BLITZ signals
// across all loaded rows. Loading and error rows are ignored.
func (m Model) GetBlitzCounts() (long, short, hold int) {
	var l, s, h int
	for _, row := range m.rows {
		if row.State != StateLoaded && row.State != StateCached {
			continue
		}
		if row.Result == nil {
			continue
		}
		switch row.Result.BlitzScore {
		case 1:
			l++
		case -1:
			s++
		default:
			h++
		}
	}
	return l, s, h
}

// KeyBindings returns the keyboard bindings for the trend view.
func (m Model) KeyBindings() []components.KeyBinding {
	return []components.KeyBinding{
		{Key: "↑", Description: "Move up"},
		{Key: "↓", Description: "Move down"},
		{Key: "r", Description: "Refresh all tickers"},
	}
}

// RefreshMsg is a message to refresh all trend data.
type RefreshMsg struct{}

// TrendDataMsg is a message when trend data is loaded for a ticker.
type TrendDataMsg struct { //nolint:revive // type name stutters with package name
	Symbol string
	Result *trenddomain.Result
	Index  int
}

// TrendErrorMsg is a message when an error occurs fetching data for a ticker.
type TrendErrorMsg struct { //nolint:revive // type name stutters with package name
	Symbol string
	Err    error
	Index  int
}

// FormatValue formats a float value for display with fixed decimal places.
func FormatValue(value float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, value)
}
