// Package chart provides a candlestick chart with TPI overlay visualization.
package chart

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shinsekai/finterm/internal/alphavantage"
	"github.com/shinsekai/finterm/internal/cache"
	"github.com/shinsekai/finterm/internal/config"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
	"github.com/shinsekai/finterm/internal/tui/components"
	"github.com/shinsekai/finterm/internal/tui/quote"
)

// State represents the loading state of the chart.
type State int

const (
	// StateLoading is when data is being fetched.
	StateLoading State = iota
	// StateLoaded is when data has been successfully loaded.
	StateLoaded
	// StateError is when there was an error loading data.
	StateError
)

// Timeframe represents the chart timeframe.
type Timeframe int

const (
	// TimeframeIntraday is 1-minute intervals (equities only).
	TimeframeIntraday Timeframe = iota
	// TimeframeDaily is daily bars (default).
	TimeframeDaily
	// TimeframeWeekly is weekly bars.
	TimeframeWeekly
	// TimeframeMonthly is monthly bars.
	TimeframeMonthly
)

// String returns the string representation of Timeframe.
func (t Timeframe) String() string {
	switch t {
	case TimeframeIntraday:
		return "intraday"
	case TimeframeDaily:
		return "daily"
	case TimeframeWeekly:
		return "weekly"
	case TimeframeMonthly:
		return "monthly"
	default:
		return "unknown"
	}
}

// chartData holds the OHLCV data and TPI series.
type chartData struct {
	Bars []indicators.OHLCV
	TPI  []float64
}

// Model represents the chart view model.
type Model struct {
	// engine performs trend analysis.
	engine quote.Engine
	// timeSeriesClient fetches time series data for equities.
	timeSeriesClient *alphavantage.Client
	// cache is used for caching time series data.
	cache cache.Cache
	// watchlist is the combined list of symbols to cycle through.
	watchlist []string
	// equitiesCount is the number of equity symbols (for cycling).
	equitiesCount int
	// detector determines asset class for symbols.
	detector *indicators.AssetClassDetector
	// config is the application configuration.
	cfg *config.Config
	// cryptoFetcher fetches OHLCV data for crypto symbols.
	cryptoFetcher CryptoDataFetcher

	// View state
	symbol    string
	timeframe Timeframe
	window    int
	offset    int
	state     State
	data      chartData
	err       error
	barClose  time.Time

	// Terminal dimensions
	width, height int

	// Context for async operations
	ctx    context.Context
	cancel context.CancelFunc
}

// CryptoDataFetcher fetches OHLCV data for cryptocurrency symbols.
type CryptoDataFetcher interface {
	FetchCryptoOHLCV(ctx context.Context, symbol string) ([]indicators.OHLCV, error)
}

// NewModel creates a new chart model.
func NewModel() *Model {
	ctx, cancel := context.WithCancel(context.Background())
	return &Model{
		timeframe: TimeframeDaily,
		window:    110,
		offset:    0,
		state:     StateLoading,
		width:     80,
		height:    24,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Configure sets up the model with dependencies.
func (m *Model) Configure(
	ctx context.Context,
	engine quote.Engine,
	timeSeriesClient *alphavantage.Client,
	cacheStore cache.Cache,
	watchlist *config.WatchlistConfig,
	detector *indicators.AssetClassDetector,
	cfg *config.Config,
	cryptoFetcher CryptoDataFetcher,
) *Model {
	m.engine = engine
	m.timeSeriesClient = timeSeriesClient
	m.cache = cacheStore
	m.detector = detector
	m.cfg = cfg
	m.ctx, m.cancel = context.WithCancel(ctx)

	// Build combined watchlist from equities and crypto
	m.equitiesCount = len(watchlist.Equities)
	m.watchlist = append([]string{}, watchlist.Equities...)
	m.watchlist = append(m.watchlist, watchlist.Crypto...)

	// Set initial symbol if watchlist is not empty
	if len(m.watchlist) > 0 {
		m.symbol = m.watchlist[0]
		m.state = StateLoading
	}

	// Store crypto fetcher in the model for crypto data fetching
	m.cryptoFetcher = cryptoFetcher

	return m
}

// Init initializes the chart model and returns an initial command.
func (m Model) Init() tea.Cmd {
	if m.symbol == "" {
		return nil
	}
	return m.fetchDataCmd()
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ChangeSymbolMsg:
		m = m.handleChangeSymbol(msg)
		return m, m.fetchDataCmd()

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case RefreshMsg:
		return m, m.fetchDataCmd()

	case DataMsg:
		m.state = StateLoaded
		m.data = msg.Data
		m.barClose = msg.BarClose
		return m, nil

	case ErrorMsg:
		m.state = StateError
		m.err = msg.Err
		return m, nil
	}

	return m, nil
}

// View renders the chart view.
func (m Model) View() string {
	return renderView(&m)
}

// handleKeyMsg handles keyboard input messages.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "j":
			return m, m.nextSymbolCmd()
		case "k":
			return m, m.prevSymbolCmd()
		case "1":
			return m.setTimeframeCmd(TimeframeIntraday)
		case "2":
			return m.setTimeframeCmd(TimeframeDaily)
		case "3":
			return m.setTimeframeCmd(TimeframeWeekly)
		case "4":
			return m.setTimeframeCmd(TimeframeMonthly)
		case "+":
			m.zoomIn()
			return m, m.fetchDataCmd()
		case "-":
			m.zoomOut()
			return m, m.fetchDataCmd()
		case "h":
			m.panLeft()
			return m, nil
		case "l":
			m.panRight()
			return m, nil
		case "r":
			return m, m.fetchDataCmd()
		}

	case tea.KeyLeft:
		m.panLeft()
		return m, nil

	case tea.KeyRight:
		m.panRight()
		return m, nil
	}

	return m, nil
}

// setTimeframeCmd attempts to set the timeframe.
// Returns a command to fetch data if successful, nil if rejected.
func (m Model) setTimeframeCmd(tf Timeframe) (Model, tea.Cmd) {
	// Reject intraday for crypto symbols
	if tf == TimeframeIntraday && m.isCrypto() {
		// Log a non-fatal chip - in a real implementation this would use a logging system
		fmt.Printf("chart: intraday timeframe not available for crypto symbol %s\n", m.symbol)
		return m, nil
	}

	m.timeframe = tf
	m.state = StateLoading
	return m, m.fetchDataCmd()
}

// isCrypto returns true if the current symbol is a cryptocurrency.
func (m *Model) isCrypto() bool {
	if m.detector == nil {
		return false
	}
	return m.detector.DetectAssetClass(m.symbol) == indicators.Crypto
}

// zoomIn increases the window size.
func (m *Model) zoomIn() {
	switch m.window {
	case 30:
		m.window = 60
	case 60:
		m.window = 110
	}
}

// zoomOut decreases the window size.
func (m *Model) zoomOut() {
	switch m.window {
	case 110:
		m.window = 60
	case 60:
		m.window = 30
	}
}

// panLeft moves the view left by 10% of the window.
func (m *Model) panLeft() {
	panAmount := m.window / 10
	if panAmount < 1 {
		panAmount = 1
	}
	m.offset -= panAmount
	if m.offset < 0 {
		m.offset = 0
	}
}

// panRight moves the view right by 10% of the window.
func (m *Model) panRight() {
	if len(m.data.Bars) == 0 {
		return
	}
	panAmount := m.window / 10
	if panAmount < 1 {
		panAmount = 1
	}
	maxOffset := len(m.data.Bars) - m.window
	if maxOffset < 0 {
		maxOffset = 0
	}
	m.offset += panAmount
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

// nextSymbolCmd returns a command to load the next symbol in the watchlist.
func (m Model) nextSymbolCmd() tea.Cmd {
	return func() tea.Msg {
		if len(m.watchlist) == 0 {
			return nil
		}
		// Find current index
		currentIdx := -1
		for i, s := range m.watchlist {
			if s == m.symbol {
				currentIdx = i
				break
			}
		}
		// Move to next (wrap around)
		nextIdx := currentIdx + 1
		if nextIdx >= len(m.watchlist) {
			nextIdx = 0
		}
		return ChangeSymbolMsg{Symbol: m.watchlist[nextIdx]}
	}
}

// prevSymbolCmd returns a command to load the previous symbol in the watchlist.
func (m Model) prevSymbolCmd() tea.Cmd {
	return func() tea.Msg {
		if len(m.watchlist) == 0 {
			return nil
		}
		// Find current index
		currentIdx := -1
		for i, s := range m.watchlist {
			if s == m.symbol {
				currentIdx = i
				break
			}
		}
		// Move to previous (wrap around)
		prevIdx := currentIdx - 1
		if prevIdx < 0 {
			prevIdx = len(m.watchlist) - 1
		}
		return ChangeSymbolMsg{Symbol: m.watchlist[prevIdx]}
	}
}

// fetchDataCmd returns a command to fetch chart data for the current symbol.
func (m Model) fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		// Fetch OHLCV data based on asset class
		var bars []indicators.OHLCV
		assetClass := indicators.Equity // Default to equity
		if m.detector != nil {
			assetClass = m.detector.DetectAssetClass(m.symbol)
		}

		if assetClass == indicators.Crypto {
			if m.cryptoFetcher != nil {
				var fetchErr error
				bars, fetchErr = m.cryptoFetcher.FetchCryptoOHLCV(m.ctx, m.symbol)
				if fetchErr != nil {
					return ErrorMsg{Symbol: m.symbol, Err: fetchErr}
				}
			}
		} else {
			// For equities, fetch time series data
			if m.timeSeriesClient != nil {
				cacheKey := cache.Key("timeseries", "daily", m.symbol)
				var tsData *alphavantage.TimeSeriesDaily

				// Try cache first
				if cached, ok := m.cache.Get(cacheKey); ok {
					if ts, ok := cached.(*alphavantage.TimeSeriesDaily); ok {
						tsData = ts
					}
				}

				// Fetch from API if not cached
				if tsData == nil {
					var fetchErr error
					tsData, fetchErr = m.timeSeriesClient.GetDailyTimeSeries(m.ctx, m.symbol, "full")
					if fetchErr != nil {
						return ErrorMsg{Symbol: m.symbol, Err: fetchErr}
					}
					m.cache.Set(cacheKey, tsData, m.cfg.Cache.DailyTTL)
				}

				// Convert TimeSeriesDaily to OHLCV
				bars = convertTimeSeriesToOHLCV(tsData)
			}
		}

		// Compute TPI history
		tpiHistory, err := computeTPIHistory(bars, m.cfg)
		if err != nil {
			return ErrorMsg{Symbol: m.symbol, Err: fmt.Errorf("computing TPI history: %w", err)}
		}

		// Get bar-close date from the last bar
		var barClose time.Time
		if len(bars) > 0 {
			barClose = bars[len(bars)-1].Date
		}

		return DataMsg{
			Symbol:   m.symbol,
			Data:     chartData{Bars: bars, TPI: tpiHistory},
			BarClose: barClose,
		}
	}
}

// convertTimeSeriesToOHLCV converts Alpha Vantage time series to domain OHLCV format.
// Excludes today's in-progress bar (bar-close-only rule).
func convertTimeSeriesToOHLCV(ts *alphavantage.TimeSeriesDaily) []indicators.OHLCV {
	if ts == nil {
		return nil
	}

	today := time.Now().UTC().Format("2006-01-02")
	ohlcv := make([]indicators.OHLCV, 0)

	for dateStr, entry := range ts.TimeSeries {
		// Skip today's in-progress bar (bar-close-only rule)
		if dateStr >= today {
			continue
		}

		date, err := alphavantage.ParseDate(dateStr)
		if err != nil {
			continue
		}

		open, _ := alphavantage.ParseFloat(entry.Open)
		high, _ := alphavantage.ParseFloat(entry.High)
		low, _ := alphavantage.ParseFloat(entry.Low)
		closeVal, _ := alphavantage.ParseFloat(entry.Close)
		volume, _ := alphavantage.ParseFloat(entry.Volume)

		ohlcv = append(ohlcv, indicators.OHLCV{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closeVal,
			Volume: volume,
		})
	}

	// Sort oldest-first
	for i := 0; i < len(ohlcv)-1; i++ {
		for j := i + 1; j < len(ohlcv); j++ {
			if ohlcv[i].Date.After(ohlcv[j].Date) {
				ohlcv[i], ohlcv[j] = ohlcv[j], ohlcv[i]
			}
		}
	}

	return ohlcv
}

// handleChangeSymbol handles a ChangeSymbolMsg.
func (m Model) handleChangeSymbol(msg ChangeSymbolMsg) Model {
	m.symbol = msg.Symbol
	m.state = StateLoading
	m.data = chartData{}
	m.err = nil
	m.offset = 0
	return m
}

// KeyBindings returns the keyboard bindings for the chart view.
func (m Model) KeyBindings() []components.KeyBinding {
	return []components.KeyBinding{
		{Key: "j/k", Description: "Prev/next ticker"},
		{Key: "1", Description: "Timeframe: intraday (equities only)"},
		{Key: "2", Description: "Timeframe: daily (default)"},
		{Key: "3", Description: "Timeframe: weekly"},
		{Key: "4", Description: "Timeframe: monthly"},
		{Key: "+/-", Description: "Zoom in/out"},
		{Key: "h/l", Description: "Pan left/right"},
		{Key: "r", Description: "Refresh chart"},
	}
}

// RefreshMsg is a message to refresh chart data.
type RefreshMsg struct{}

// ChangeSymbolMsg is a message to change the current symbol.
type ChangeSymbolMsg struct {
	Symbol string
}

// DataMsg is a message when chart data is loaded.
type DataMsg struct {
	Symbol   string
	Data     chartData
	BarClose time.Time
}

// ErrorMsg is a message when an error occurs loading chart data.
type ErrorMsg struct {
	Symbol string
	Err    error
}
