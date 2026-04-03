// Package trend provides the trend following TUI model.
package trend

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/owner/finterm/internal/config"
	trenddomain "github.com/owner/finterm/internal/domain/trend"
	"github.com/owner/finterm/internal/domain/trend/indicators"
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
		return m.handleTrendData(msg)

	case TrendErrorMsg:
		// Handle error for a single ticker
		return m.handleTrendError(msg)

	case FetchCompleteMsg:
		// All fetches completed
		m.updateOverallState()
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
func (m Model) handleTrendData(msg TrendDataMsg) (Model, tea.Cmd) {
	for i, row := range m.rows {
		if row.Symbol == msg.Symbol {
			m.rows[i].State = StateLoaded
			m.rows[i].Result = msg.Result
			m.rows[i].Error = nil
			return m, nil
		}
	}
	return m, nil
}

// handleTrendError updates the model with an error for a ticker.
func (m Model) handleTrendError(msg TrendErrorMsg) (Model, tea.Cmd) {
	for i, row := range m.rows {
		if row.Symbol == msg.Symbol {
			m.rows[i].State = StateError
			m.rows[i].Error = msg.Err
			return m, nil
		}
	}
	return m, nil
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

// fetchAllCmd returns a command that fetches data for all tickers concurrently.
func (m Model) fetchAllCmd() tea.Cmd {
	return func() tea.Msg {
		cmds := make([]tea.Cmd, len(m.watchlist))

		// Create a fetch command for each ticker
		for i, symbol := range m.watchlist {
			symbol := symbol // Capture loop variable
			cmds[i] = func() tea.Msg {
				return m.fetchTicker(symbol)
			}
		}

		// Execute all fetches concurrently
		// Return the first fetch complete message to trigger UI update
		// Individual results will arrive as separate messages
		return FetchCompleteMsg{}
	}
}

// fetchTicker fetches trend data for a single ticker.
func (m Model) fetchTicker(symbol string) tea.Msg {
	result, err := m.engine.AnalyzeWithSymbolDetection(m.ctx, symbol)

	if err != nil {
		return TrendErrorMsg{
			Symbol: symbol,
			Err:    err,
		}
	}

	return TrendDataMsg{
		Symbol: symbol,
		Result: result,
	}
}

// updateOverallState updates the overall state based on row states.
func (m *Model) updateOverallState() {
	var hasLoaded, hasError bool
	allLoaded := true

	for _, row := range m.rows {
		switch row.State {
		case StateLoading:
			allLoaded = false
		case StateLoaded:
			hasLoaded = true
		case StateError:
			hasError = true
		}
	}

	switch {
	case !allLoaded:
		m.overallState = StateLoading
	case hasError && !hasLoaded:
		m.overallState = StateError
	default:
		m.overallState = StateLoaded
	}
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

// RefreshMsg is a message to refresh all trend data.
type RefreshMsg struct{}

// TrendDataMsg is a message when trend data is loaded for a ticker.
type TrendDataMsg struct { //nolint:revive // type name stutters with package name
	Symbol string
	Result *trenddomain.Result
}

// TrendErrorMsg is a message when an error occurs fetching data for a ticker.
type TrendErrorMsg struct { //nolint:revive // type name stutters with package name
	Symbol string
	Err    error
}

// FetchCompleteMsg is a message when all fetch commands have been dispatched.
type FetchCompleteMsg struct{}

// FormatValue formats a float value for display with fixed decimal places.
func FormatValue(value float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, value)
}
