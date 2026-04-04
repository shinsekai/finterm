// Package quote provides the single ticker quote TUI model.
package quote

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/owner/finterm/internal/alphavantage"
	trenddomain "github.com/owner/finterm/internal/domain/trend"
	"github.com/owner/finterm/internal/domain/trend/indicators"
	"github.com/owner/finterm/internal/tui/components"
)

// State represents the model state.
type State int

const (
	// StateIdle is when waiting for user input.
	StateIdle State = iota
	// StateLoading is when fetching quote data.
	StateLoading
	// StateLoaded is when quote data has been loaded.
	StateLoaded
	// StateError is when an error occurred.
	StateError
)

// String returns the string representation of the State.
func (s State) String() string {
	switch s {
	case StateIdle:
		return "Idle"
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

// Engine defines the interface for trend analysis engines.
// This allows mocking in tests while the actual implementation uses trenddomain.Engine.
type Engine interface {
	AnalyzeWithSymbolDetection(ctx context.Context, symbol string) (*trenddomain.Result, error)
}

// QuoteClient defines the interface for fetching quote data.
// This allows mocking in tests while the actual implementation uses alphavantage.Client.
//
//nolint:revive // type name stuttering is acceptable for package-scoped interfaces
type QuoteClient interface {
	GetGlobalQuote(ctx context.Context, symbol string) (*alphavantage.GlobalQuote, error)
}

// TimeSeriesClient defines the interface for fetching time series data for equities.
type TimeSeriesClient interface {
	GetDailyTimeSeries(ctx context.Context, symbol, outputsize string) (*alphavantage.TimeSeriesDaily, error)
}

// CryptoQuoteClient defines the interface for fetching crypto quote data.
type CryptoQuoteClient interface {
	GetCryptoDaily(ctx context.Context, symbol, market string) (*alphavantage.CryptoDaily, error)
}

// QuoteData contains all the data to display for a quote.
//
//nolint:revive // type name stuttering is acceptable for package-scoped types
type QuoteData struct {
	Quote      *alphavantage.GlobalQuote
	Indicators *trenddomain.Result
}

// Model represents the quote view model.
type Model struct {
	// engine performs trend analysis for symbols.
	engine Engine
	// client fetches quote data from Alpha Vantage.
	client QuoteClient
	// detector determines asset class for a symbol.
	detector *indicators.AssetClassDetector
	// ctx is the context for async operations.
	ctx context.Context
	// cancel cancels the context.
	cancel context.CancelFunc

	// state is the current model state.
	state State
	// textInput handles ticker symbol input.
	textInput textinput.Model
	// lookupHistory stores the last 10 ticker symbols looked up.
	lookupHistory []string
	// historyIndex is the current position in lookup history (-1 if not navigating).
	historyIndex int
	// maxHistorySize is the maximum number of entries in lookup history.
	maxHistorySize int

	// quoteData contains the loaded quote and indicators.
	quoteData *QuoteData
	// err contains any error that occurred.
	err error

	// width and height are the terminal dimensions.
	width, height int
}

// NewModel creates a new quote model.
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter ticker symbol (e.g., AAPL)"
	ti.CharLimit = 10
	ti.Focus()

	return Model{
		state:          StateIdle,
		textInput:      ti,
		lookupHistory:  []string{},
		historyIndex:   -1,
		maxHistorySize: 10,
		width:          80,
		height:         24,
	}
}

// Configure sets up the model with dependencies.
// This is called by the main app to inject dependencies.
func (m *Model) Configure(
	ctx context.Context,
	client QuoteClient,
	engine Engine,
	detector *indicators.AssetClassDetector,
) *Model {
	m.client = client
	m.engine = engine
	m.detector = detector
	m.ctx, m.cancel = context.WithCancel(ctx)
	return m
}

// Init initializes the quote model and returns an initial command.
func (m Model) Init() tea.Cmd {
	return nil
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
		// Refresh current ticker if we have one
		if m.quoteData != nil && m.quoteData.Quote != nil {
			symbol := m.quoteData.Quote.Symbol
			return m, m.fetchQuoteCmd(symbol)
		}
		return m, nil

	case QuoteResultMsg:
		m.state = StateLoaded
		m.quoteData = msg.Data
		m.err = nil

		// Add to lookup history if this is a new lookup (not from history navigation)
		if m.historyIndex == -1 && msg.Data.Quote != nil {
			m.addToHistory(msg.Data.Quote.Symbol)
		}
		m.historyIndex = -1 // Reset history index

		// Clear and re-focus input for next query
		m.textInput.SetValue("")
		m.textInput.Focus()

		return m, nil

	case QuoteErrorMsg:
		m.state = StateError
		m.err = msg.Err
		return m, nil
	}

	// Delegate text input updates - always allow typing regardless of state
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleKeyMsg handles keyboard input messages.
//
//nolint:gocyclo // complexity is acceptable for handling all key messages
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Submit ticker and fetch quote (from idle, loaded, or error state)
		if m.state == StateIdle || m.state == StateLoaded || m.state == StateError {
			ticker := strings.TrimSpace(m.textInput.Value())
			if ticker == "" {
				return m, nil
			}

			// Validate ticker format
			if err := validateTicker(ticker); err != nil {
				m.state = StateError
				m.err = err
				return m, nil
			}

			m.state = StateLoading
			return m, m.fetchQuoteCmd(ticker)
		}

	case tea.KeyUp:
		// Navigate backward through history
		if m.state == StateIdle && len(m.lookupHistory) > 0 {
			m.historyIndex++
			if m.historyIndex >= len(m.lookupHistory) {
				m.historyIndex = len(m.lookupHistory) - 1
			}
			m.textInput.SetValue(m.lookupHistory[len(m.lookupHistory)-1-m.historyIndex])
			return m, nil
		}

	case tea.KeyDown:
		// Navigate forward through history
		if m.state == StateIdle && m.historyIndex >= 0 {
			m.historyIndex--
			if m.historyIndex < 0 {
				m.historyIndex = -1
				m.textInput.SetValue("") // Clear input when going past history
			} else {
				m.textInput.SetValue(m.lookupHistory[len(m.lookupHistory)-1-m.historyIndex])
			}
			return m, nil
		}

	case tea.KeyEsc:
		// Clear input and reset to idle from any state
		m.textInput.SetValue("")
		m.textInput.Focus()
		m.historyIndex = -1
		m.state = StateIdle
		m.quoteData = nil
		m.err = nil
		return m, nil

	case tea.KeyRunes:
		// Reset history index when typing
		if m.state == StateIdle && len(msg.Runes) > 0 {
			m.historyIndex = -1
		}
	}

	// Delegate to text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// fetchQuoteCmd returns a command to fetch quote data for a symbol.
func (m Model) fetchQuoteCmd(symbol string) tea.Cmd {
	return func() tea.Msg {
		symbol = strings.ToUpper(symbol)

		var quote *alphavantage.GlobalQuote
		var err error

		// Detect asset class and use appropriate endpoint
		isCrypto := m.detector != nil && m.detector.DetectAssetClass(symbol) == indicators.Crypto
		if isCrypto {
			quote, err = m.fetchCryptoAsQuote(symbol)
		} else {
			// For stocks: use TIME_SERIES_DAILY, get latest closed bar
			quote, err = m.fetchStockAsQuote(symbol)
		}

		if err != nil {
			return QuoteErrorMsg{Err: fmt.Errorf("fetching quote for %s: %w", symbol, err)}
		}

		// Fetch indicators
		indicators, err := m.engine.AnalyzeWithSymbolDetection(m.ctx, symbol)
		if err != nil {
			return QuoteErrorMsg{Err: fmt.Errorf("fetching indicators for %s: %w", symbol, err)}
		}

		return QuoteResultMsg{
			Data: &QuoteData{Quote: quote, Indicators: indicators},
		}
	}
}

// fetchStockAsQuote fetches a stock quote from TIME_SERIES_DAILY endpoint.
func (m Model) fetchStockAsQuote(symbol string) (*alphavantage.GlobalQuote, error) {
	tsClient, ok := m.client.(TimeSeriesClient)
	if !ok {
		// Fallback to GlobalQuote if client doesn't support time series
		return m.client.GetGlobalQuote(m.ctx, symbol)
	}

	data, err := tsClient.GetDailyTimeSeries(m.ctx, symbol, "compact")
	if err != nil {
		return nil, err
	}

	// Find latest and previous dates
	latestDate, prevDate := findLatestTwoDates(data.TimeSeries)
	if latestDate == "" {
		return nil, fmt.Errorf("no data for %s", symbol)
	}

	entry := data.TimeSeries[latestDate]
	quote := &alphavantage.GlobalQuote{
		Symbol:         symbol,
		Open:           entry.Open,
		High:           entry.High,
		Low:            entry.Low,
		Price:          entry.Close,
		Volume:         entry.Volume,
		LastTradingDay: latestDate,
	}

	// Calculate change from previous day
	if prevDate != "" {
		prevEntry := data.TimeSeries[prevDate]
		closeVal, _ := alphavantage.ParseFloat(entry.Close)
		prevClose, _ := alphavantage.ParseFloat(prevEntry.Close)
		if prevClose > 0 {
			diff := closeVal - prevClose
			pct := (diff / prevClose) * 100
			quote.PreviousClose = prevEntry.Close
			quote.Change = fmt.Sprintf("%.4f", diff)
			quote.ChangePercent = fmt.Sprintf("%.4f%%", pct)
		}
	}

	return quote, nil
}

// fetchCryptoAsQuote fetches a crypto quote from DIGITAL_CURRENCY_DAILY endpoint.
func (m Model) fetchCryptoAsQuote(symbol string) (*alphavantage.GlobalQuote, error) {
	cryptoClient, ok := m.client.(CryptoQuoteClient)
	if !ok {
		return nil, fmt.Errorf("client does not support crypto quotes")
	}

	data, err := cryptoClient.GetCryptoDaily(m.ctx, symbol, "USD")
	if err != nil {
		return nil, err
	}

	// Find latest and previous dates
	latestDate, prevDate := findLatestTwoDates(data.TimeSeries)
	if latestDate == "" {
		return nil, fmt.Errorf("no data for %s", symbol)
	}

	entry := data.TimeSeries[latestDate]
	quote := &alphavantage.GlobalQuote{
		Symbol:         symbol,
		Open:           entry.Open,
		High:           entry.High,
		Low:            entry.Low,
		Price:          entry.Close,
		Volume:         entry.Volume,
		LastTradingDay: latestDate,
	}

	// Calculate change from previous day
	if prevDate != "" {
		prevEntry := data.TimeSeries[prevDate]
		closeVal, _ := alphavantage.ParseFloat(entry.Close)
		prevClose, _ := alphavantage.ParseFloat(prevEntry.Close)
		if prevClose > 0 {
			diff := closeVal - prevClose
			pct := (diff / prevClose) * 100
			quote.PreviousClose = prevEntry.Close
			quote.Change = fmt.Sprintf("%.4f", diff)
			quote.ChangePercent = fmt.Sprintf("%.4f%%", pct)
		}
	}

	return quote, nil
}

// findLatestTwoDates finds the two most recent dates from a time series map.
func findLatestTwoDates[V any](ts map[string]V) (latest, prev string) {
	for date := range ts {
		if date > latest {
			prev = latest
			latest = date
		} else if date > prev {
			prev = date
		}
	}
	return
}

// addToHistory adds a symbol to the lookup history, maintaining max size.
func (m *Model) addToHistory(symbol string) {
	symbol = strings.ToUpper(symbol)
	// Check if it's already the last entry
	if len(m.lookupHistory) > 0 && m.lookupHistory[len(m.lookupHistory)-1] == symbol {
		return
	}

	m.lookupHistory = append(m.lookupHistory, symbol)
	// Trim history if it exceeds max size
	if len(m.lookupHistory) > m.maxHistorySize {
		m.lookupHistory = m.lookupHistory[len(m.lookupHistory)-m.maxHistorySize:]
	}
}

// validateTicker validates that a ticker symbol is valid.
func validateTicker(ticker string) error {
	if ticker == "" {
		return fmt.Errorf("ticker cannot be empty")
	}

	if len(ticker) > 10 {
		return fmt.Errorf("ticker exceeds maximum length of 10 characters (got %d)", len(ticker))
	}

	// Allow: A-Z, a-z, 0-9, dots, dashes
	matched, err := regexp.MatchString(`^[A-Za-z0-9.-]+$`, ticker)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !matched {
		return fmt.Errorf("ticker contains invalid characters (only alphanumeric, dot, and dash allowed)")
	}

	return nil
}

// GetState returns the current state.
func (m Model) GetState() State {
	return m.state
}

// GetTextInput returns the text input model.
func (m Model) GetTextInput() textinput.Model {
	return m.textInput
}

// GetQuoteData returns the loaded quote data.
func (m Model) GetQuoteData() *QuoteData {
	return m.quoteData
}

// GetError returns the current error.
func (m Model) GetError() error {
	return m.err
}

// GetWidth returns the current width.
func (m Model) GetWidth() int {
	return m.width
}

// GetHeight returns the current height.
func (m Model) GetHeight() int {
	return m.height
}

// KeyBindings returns the keyboard bindings for the quote view.
func (m Model) KeyBindings() []components.KeyBinding {
	return []components.KeyBinding{
		{Key: "Enter", Description: "Fetch quote"},
		{Key: "↑", Description: "Previous in history"},
		{Key: "↓", Description: "Next in history"},
		{Key: "Esc", Description: "Clear input"},
		{Key: "r", Description: "Refresh ticker"},
	}
}

// RefreshMsg is a message to refresh the quote data.
type RefreshMsg struct{}

// QuoteResultMsg is a message when quote data is loaded.
//
//nolint:revive // type name stuttering is acceptable for package-scoped types
type QuoteResultMsg struct {
	Data *QuoteData
}

// QuoteErrorMsg is a message when an error occurs.
//
//nolint:revive // type name stuttering is acceptable for package-scoped types
type QuoteErrorMsg struct {
	Err error
}
