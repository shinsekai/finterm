// Package quote provides tests for quote TUI model.
package quote

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/shinsekai/finterm/internal/alphavantage"
	trenddomain "github.com/shinsekai/finterm/internal/domain/trend"
	"github.com/shinsekai/finterm/internal/domain/trend/indicators"
)

// mockEngine is a mock implementation of Engine for testing.
type mockEngine struct {
	result *trenddomain.Result
	err    error
}

func (m *mockEngine) AnalyzeWithSymbolDetection(_ context.Context, _ string) (*trenddomain.Result, error) {
	return m.result, m.err
}

// mockClient is a mock implementation of QuoteClient for testing.
type mockClient struct {
	quote       *alphavantage.GlobalQuote
	cryptoDaily *alphavantage.CryptoDaily
	err         error
}

func (m *mockClient) GetGlobalQuote(_ context.Context, _ string) (*alphavantage.GlobalQuote, error) {
	return m.quote, m.err
}

// GetCryptoDaily implements CryptoQuoteClient for crypto testing.
func (m *mockClient) GetCryptoDaily(_ context.Context, symbol, market string) (*alphavantage.CryptoDaily, error) {
	if m.cryptoDaily != nil {
		return m.cryptoDaily, m.err
	}
	// If no cryptoDaily is set, return error
	return nil, fmt.Errorf("no crypto data configured for %s/%s", symbol, market)
}

// mockCommodityClient is a mock implementation of CommodityQuoteClient for testing.
type mockCommodityClient struct {
	series *alphavantage.CommoditySeries
	err    error
}

func (m *mockCommodityClient) GetGlobalQuote(_ context.Context, _ string) (*alphavantage.GlobalQuote, error) {
	return nil, nil
}

func (m *mockCommodityClient) GetCommodity(_ context.Context, _ alphavantage.CommodityFunction, _ string) (*alphavantage.CommoditySeries, error) {
	return m.series, m.err
}

// TestNewModel verifies model is initialized correctly.
func TestNewModel(t *testing.T) {
	model := NewModel()

	if model.state != StateIdle {
		t.Errorf("Expected state %v, got %v", StateIdle, model.state)
	}

	if model.textInput.Placeholder != "Enter ticker symbol (e.g., AAPL)" {
		t.Errorf("Expected placeholder %q, got %q", "Enter ticker symbol (e.g., AAPL)", model.textInput.Placeholder)
	}

	if model.textInput.CharLimit != 10 {
		t.Errorf("Expected char limit %d, got %d", 10, model.textInput.CharLimit)
	}

	if model.historyIndex != -1 {
		t.Errorf("Expected history index -1, got %d", model.historyIndex)
	}

	if len(model.lookupHistory) != 0 {
		t.Errorf("Expected empty history, got %v", model.lookupHistory)
	}
}

// TestQuoteModel_Init verifies Init returns nil command.
func TestQuoteModel_Init(t *testing.T) {
	model := NewModel()
	cmd := model.Init()

	if cmd != nil {
		t.Errorf("Expected nil command, got %v", cmd)
	}
}

// TestQuoteModel_Configure verifies Configure sets up dependencies correctly.
func TestQuoteModel_Configure(t *testing.T) {
	model := NewModel()
	ctx := context.Background()
	client := &mockClient{}
	engine := &mockEngine{}

	configured := model.Configure(ctx, client, engine, indicators.NewAssetClassDetector(nil, nil))

	if configured.client != client {
		t.Error("Client not configured")
	}

	if configured.engine != engine {
		t.Error("Engine not configured")
	}

	if configured.ctx == nil {
		t.Error("Context not configured")
	}

	if configured.cancel == nil {
		t.Error("Cancel function not configured")
	}
}

// TestQuoteModel_Update_KeyMsg_Enter_SubmitTicker verifies pressing Enter with a valid ticker submits fetch command.
func TestQuoteModel_Update_KeyMsg_Enter_SubmitTicker(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	// Set a ticker in input
	ti := model.textInput
	ti.SetValue("AAPL")
	model.textInput = ti

	// Press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Expected a command to be returned")
	}

	// Verify state changed to loading
	if newModel.(Model).state != StateLoading {
		t.Errorf("Expected state %v, got %v", StateLoading, newModel.(Model).state)
	}
}

// TestQuoteModel_Update_KeyMsg_Enter_EmptyTicker verifies pressing Enter with empty ticker does nothing.
func TestQuoteModel_Update_KeyMsg_Enter_EmptyTicker(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	// Press Enter with empty input
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)

	if cmd != nil {
		t.Errorf("Expected no command, got %v", cmd)
	}

	if newModel.(Model).state != StateIdle {
		t.Errorf("Expected state %v, got %v", StateIdle, newModel.(Model).state)
	}
}

// TestQuoteModel_Update_KeyMsg_Enter_InvalidTicker verifies pressing Enter with invalid ticker shows error.
func TestQuoteModel_Update_KeyMsg_Enter_InvalidTicker(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	// Set an invalid ticker
	ti := model.textInput
	ti.SetValue("INVALID@TICKER!")
	model.textInput = ti

	// Press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg)

	if cmd != nil {
		t.Errorf("Expected no command, got %v", cmd)
	}

	if newModel.(Model).state != StateError {
		t.Errorf("Expected state %v, got %v", StateError, newModel.(Model).state)
	}

	if newModel.(Model).err == nil {
		t.Error("Expected error to be set")
	}
}

// TestQuoteModel_Update_KeyMsg_Up_Navigation verifies Up arrow navigates backward through history.
func TestQuoteModel_Update_KeyMsg_Up_Navigation(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	// Add some history entries
	model.lookupHistory = []string{"AAPL", "MSFT", "GOOGL"}

	// Press Up - starts from index 0 showing most recent
	msg := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ := model.Update(msg)

	// First Up navigates to most recent entry
	if newModel.(Model).textInput.Value() != "GOOGL" {
		t.Errorf("Expected ticker GOOGL, got %s", newModel.(Model).textInput.Value())
	}

	if newModel.(Model).historyIndex != 0 {
		t.Errorf("Expected history index 0, got %d", newModel.(Model).historyIndex)
	}
}

// TestQuoteModel_Update_KeyMsg_Down_Navigation verifies Down arrow navigates forward through history.
func TestQuoteModel_Update_KeyMsg_Down_Navigation(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	// Add some history and navigate back
	model.lookupHistory = []string{"AAPL", "MSFT"}
	model.historyIndex = 1 // At first entry
	model.textInput.SetValue("AAPL")

	// Press Down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msg)

	if newModel.(Model).textInput.Value() != "MSFT" {
		t.Errorf("Expected ticker MSFT, got %s", newModel.(Model).textInput.Value())
	}

	if newModel.(Model).historyIndex != 0 {
		t.Errorf("Expected history index 0, got %d", newModel.(Model).historyIndex)
	}
}

// TestQuoteModel_Update_KeyMsg_Esc_ClearInput verifies pressing Esc clears input and resets history.
func TestQuoteModel_Update_KeyMsg_Esc_ClearInput(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	// Set some text and navigate history
	ti := model.textInput
	ti.SetValue("AAPL")
	model.textInput = ti
	model.historyIndex = 1

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)

	if newModel.(Model).textInput.Value() != "" {
		t.Errorf("Expected empty input, got %s", newModel.(Model).textInput.Value())
	}

	if newModel.(Model).historyIndex != -1 {
		t.Errorf("Expected history index -1, got %d", newModel.(Model).historyIndex)
	}
}

// TestQuoteModel_Update_QuoteResultMsg verifies handling of successful quote result.
func TestQuoteModel_Update_QuoteResultMsg(t *testing.T) {
	model := NewModel()
	model.state = StateLoading

	// Create mock quote data
	quote := &alphavantage.GlobalQuote{
		Symbol:         "AAPL",
		Price:          "189.84",
		Change:         "2.34",
		ChangePercent:  "1.25%",
		Open:           "187.50",
		High:           "190.20",
		Low:            "186.80",
		Volume:         "52345678",
		PreviousClose:  "187.50",
		LastTradingDay: "2026-04-01",
	}

	indicators := &trenddomain.Result{
		Symbol:       "AAPL",
		RSI:          52.3,
		Signal:       trenddomain.Bullish,
		Valuation:    "Fair value",
		BlitzScore:   1,
		DestinyScore: 1,
		TPI:          0.67,
		TPISignal:    "LONG",
	}

	msg := QuoteResultMsg{
		Data: &QuoteData{
			Quote:      quote,
			Indicators: indicators,
		},
	}

	newModel, _ := model.Update(msg)

	if newModel.(Model).state != StateLoaded {
		t.Errorf("Expected state %v, got %v", StateLoaded, newModel.(Model).state)
	}

	if newModel.(Model).quoteData == nil {
		t.Error("Expected quote data to be set")
	}

	if newModel.(Model).quoteData.Quote.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", newModel.(Model).quoteData.Quote.Symbol)
	}

	// Should be added to history
	if len(newModel.(Model).lookupHistory) != 1 {
		t.Errorf("Expected history length 1, got %d", len(newModel.(Model).lookupHistory))
	}

	if newModel.(Model).historyIndex != -1 {
		t.Errorf("Expected history index -1, got %d", newModel.(Model).historyIndex)
	}
}

// TestQuoteModel_Update_QuoteErrorMsg verifies handling of error message.
func TestQuoteModel_Update_QuoteErrorMsg(t *testing.T) {
	model := NewModel()
	model.state = StateLoading

	expectedErr := errors.New("network error")
	msg := QuoteErrorMsg{Err: expectedErr}

	newModel, _ := model.Update(msg)

	if newModel.(Model).state != StateError {
		t.Errorf("Expected state %v, got %v", StateError, newModel.(Model).state)
	}

	if newModel.(Model).err == nil {
		t.Error("Expected error to be set")
	}

	if newModel.(Model).err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, newModel.(Model).err)
	}
}

// TestQuoteModel_Update_RefreshMsg verifies refreshing current ticker.
func TestQuoteModel_Update_RefreshMsg(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	// Set up existing quote data
	quote := &alphavantage.GlobalQuote{Symbol: "AAPL"}
	indicators := &trenddomain.Result{Symbol: "AAPL"}
	model.quoteData = &QuoteData{Quote: quote, Indicators: indicators}

	ti := model.textInput
	ti.SetValue("AAPL")
	model.textInput = ti

	// Send refresh message
	msg := RefreshMsg{}
	_, cmd := model.Update(msg)

	// Should return a command to fetch the quote
	if cmd == nil {
		t.Error("Expected a command to be returned")
	}
}

// TestQuoteModel_Update_WindowSizeMsg verifies window size is updated.
func TestQuoteModel_Update_WindowSizeMsg(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.height = 24

	msg := tea.WindowSizeMsg{Width: 120, Height: 30}
	newModel, _ := model.Update(msg)

	if newModel.(Model).width != 120 {
		t.Errorf("Expected width 120, got %d", newModel.(Model).width)
	}

	if newModel.(Model).height != 30 {
		t.Errorf("Expected height 30, got %d", newModel.(Model).height)
	}
}

// TestQuoteModel_Validation_TickerTooLong verifies ticker length validation.
func TestQuoteModel_Validation_TickerTooLong(t *testing.T) {
	err := validateTicker("WAYTOOLONGTICKER")
	if err == nil {
		t.Error("Expected error for ticker that is too long")
	}

	expectedMsg := "ticker exceeds maximum length of 10 characters"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain %q, got %q", expectedMsg, err.Error())
	}
}

// TestQuoteModel_Validation_EmptyTicker verifies empty ticker validation.
func TestQuoteModel_Validation_EmptyTicker(t *testing.T) {
	err := validateTicker("")
	if err == nil {
		t.Error("Expected error for empty ticker")
	}

	expectedMsg := "ticker cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

// TestQuoteModel_Validation_InvalidCharacters verifies ticker character validation.
func TestQuoteModel_Validation_InvalidCharacters(t *testing.T) {
	err := validateTicker("A@")
	if err == nil {
		t.Error("Expected error for ticker with invalid characters")
	}

	expectedMsg := "ticker contains invalid characters"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain %q, got %q", expectedMsg, err.Error())
	}
}

// TestQuoteModel_Validation_ValidTickers verifies valid tickers pass validation.
func TestQuoteModel_Validation_ValidTickers(t *testing.T) {
	validTickers := []string{"AAPL", "MSFT", "GOOGL", "BTC", "ETH", "BTC-USD", "BRK.B"}

	for _, ticker := range validTickers {
		err := validateTicker(ticker)
		if err != nil {
			t.Errorf("Expected no error for ticker %s, got %v", ticker, err)
		}
	}
}

// TestQuoteModel_LookupHistory_AddToHistory verifies adding to lookup history.
func TestQuoteModel_LookupHistory_AddToHistory(t *testing.T) {
	model := NewModel()
	model.lookupHistory = []string{"AAPL", "MSFT"}

	// Add a new symbol
	model.addToHistory("GOOGL")

	if len(model.lookupHistory) != 3 {
		t.Errorf("Expected history length 3, got %d", len(model.lookupHistory))
	}

	if model.lookupHistory[2] != "GOOGL" {
		t.Errorf("Expected GOOGL at end of history, got %s", model.lookupHistory[2])
	}
}

// TestQuoteModel_LookupHistory_Duplicate verifies duplicates are not added.
func TestQuoteModel_LookupHistory_Duplicate(t *testing.T) {
	model := NewModel()
	model.lookupHistory = []string{"AAPL", "MSFT"}

	// Add duplicate of last entry
	model.addToHistory("MSFT")

	if len(model.lookupHistory) != 2 {
		t.Errorf("Expected history length 2 (no duplicate), got %d", len(model.lookupHistory))
	}
}

// TestQuoteModel_LookupHistory_MaxSize verifies history is trimmed to max size.
func TestQuoteModel_LookupHistory_MaxSize(t *testing.T) {
	model := NewModel()
	model.maxHistorySize = 3

	// Fill up to max size
	model.lookupHistory = []string{"AAPL", "MSFT", "GOOGL"}

	// Add one more - should trim to max
	model.addToHistory("AMZN")

	if len(model.lookupHistory) != 3 {
		t.Errorf("Expected history length 3, got %d", len(model.lookupHistory))
	}

	// Oldest entry should be removed
	if model.lookupHistory[0] != "MSFT" {
		t.Errorf("Expected MSFT at start, got %s", model.lookupHistory[0])
	}

	if model.lookupHistory[2] != "AMZN" {
		t.Errorf("Expected AMZN at end, got %s", model.lookupHistory[2])
	}
}

// TestQuoteModel_LookupHistory_CaseNormalization verifies symbols are stored in uppercase.
func TestQuoteModel_LookupHistory_CaseNormalization(t *testing.T) {
	model := NewModel()
	model.addToHistory("aapl")

	if model.lookupHistory[0] != "AAPL" {
		t.Errorf("Expected symbol in uppercase, got %s", model.lookupHistory[0])
	}
}

// TestQuoteModel_View renders view without errors.
func TestQuoteModel_View(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	// Should not panic
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

// TestQuoteModel_View_Loading renders loading view.
func TestQuoteModel_View_Loading(t *testing.T) {
	model := NewModel()
	model.state = StateLoading
	ti := model.textInput
	ti.SetValue("AAPL")
	model.textInput = ti

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty loading view")
	}

	// Should contain loading text
	if len(view) < 10 {
		t.Error("Expected longer loading view output")
	}
}

// TestQuoteModel_View_Error renders error view.
func TestQuoteModel_View_Error(t *testing.T) {
	model := NewModel()
	model.state = StateError
	model.err = errors.New("test error")

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty error view")
	}

	// Should contain error text
	if len(view) < 10 {
		t.Error("Expected longer error view output")
	}
}

// TestQuoteModel_Getters verifies getter methods.
func TestQuoteModel_Getters(t *testing.T) {
	model := NewModel()
	model.state = StateLoaded
	model.width = 100
	model.height = 40

	quote := &alphavantage.GlobalQuote{Symbol: "AAPL"}
	indicators := &trenddomain.Result{Symbol: "AAPL"}
	model.quoteData = &QuoteData{Quote: quote, Indicators: indicators}

	expectedErr := errors.New("test error")
	model.err = expectedErr

	if model.GetState() != StateLoaded {
		t.Errorf("Expected state %v, got %v", StateLoaded, model.GetState())
	}

	if model.GetWidth() != 100 {
		t.Errorf("Expected width 100, got %d", model.GetWidth())
	}

	if model.GetHeight() != 40 {
		t.Errorf("Expected height 40, got %d", model.GetHeight())
	}

	if model.GetQuoteData() == nil {
		t.Error("Expected quote data to be returned")
	}

	if model.GetError() != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, model.GetError())
	}
}

// TestQuoteModel_ResultDisplay_CanRetryInErrorState verifies that Enter works in error state.
func TestQuoteModel_ResultDisplay_CanRetryInErrorState(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	// First submit
	ti := model.textInput
	ti.SetValue("AAPL")
	model.textInput = ti

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	model = newModel.(Model)

	// Simulate error - in real flow this would come from API
	model.state = StateError
	model.err = errors.New("network error")

	// Press Enter in error state - should retry and transition to Loading
	msg2 := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := model.Update(msg2)
	model = newModel.(Model)

	// Should transition to Loading state
	if model.state != StateLoading {
		t.Errorf("Expected state %v, got %v", StateLoading, model.state)
	}

	// A command should be returned to fetch the quote
	if cmd == nil {
		t.Error("Expected a command to be returned for retry")
	}
}

func TestQuoteModel_StateTransitions(t *testing.T) {
	tests := []struct {
		name     string
		prepare  func(Model) Model
		input    tea.Msg
		expected State
	}{
		{
			name: "Idle to Loading on Enter",
			prepare: func(m Model) Model {
				ti := m.textInput
				ti.SetValue("AAPL")
				m.textInput = ti
				m.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))
				return m
			},
			input:    tea.KeyMsg{Type: tea.KeyEnter},
			expected: StateLoading,
		},
		{
			name: "Loading to Loaded on QuoteResultMsg",
			prepare: func(m Model) Model {
				m.state = StateLoading
				return m
			},
			input: QuoteResultMsg{
				Data: &QuoteData{
					Quote:      &alphavantage.GlobalQuote{Symbol: "AAPL"},
					Indicators: &trenddomain.Result{Symbol: "AAPL"},
				},
			},
			expected: StateLoaded,
		},
		{
			name: "Loading to Error on QuoteErrorMsg",
			prepare: func(m Model) Model {
				m.state = StateLoading
				return m
			},
			input:    QuoteErrorMsg{Err: errors.New("error")},
			expected: StateError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tt.prepare(NewModel())
			newModel, _ := model.Update(tt.input)
			model = newModel.(Model)

			if model.state != tt.expected {
				t.Errorf("Expected state %v, got %v", tt.expected, model.state)
			}
		})
	}
}

// TestQuoteModel_InputMaxLength verifies input respects character limit.
func TestQuoteModel_InputMaxLength(t *testing.T) {
	model := NewModel()

	// Try to input more than char limit
	longInput := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ti := model.textInput
	ti.SetValue(longInput)

	// The textinput component should enforce the limit
	if len(ti.Value()) > model.textInput.CharLimit {
		t.Errorf("Input exceeds char limit of %d", model.textInput.CharLimit)
	}
}

// TestQuoteModel_FetchQuoteCmd_CreatesCommand verifies fetch command is created.
func TestQuoteModel_FetchQuoteCmd_CreatesCommand(t *testing.T) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	cmd := model.fetchQuoteCmd("AAPL")
	if cmd == nil {
		t.Error("Expected command to be created")
	}
}

// BenchmarkQuoteModel_Update benchmarks Update method.
func BenchmarkQuoteModel_Update(b *testing.B) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil, nil))

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A', 'A', 'P', 'L'}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newModel, _ := model.Update(msg)
		model = newModel.(Model)
	}
}

// BenchmarkQuoteModel_ValidateTicker benchmarks ticker validation.
func BenchmarkQuoteModel_ValidateTicker(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateTicker("AAPL")
	}
}

// TestQuoteView_PriceCard_RendersBorderedBox verifies rounded border characters.
func TestQuoteView_PriceCard_RendersBorderedBox(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:         "AAPL",
			Price:          "189.84",
			Change:         "2.34",
			ChangePercent:  "1.25%",
			Open:           "187.50",
			High:           "190.20",
			Low:            "186.80",
			Volume:         "52345678",
			PreviousClose:  "187.50",
			LastTradingDay: "2026-04-01",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	rendered := view.renderPriceCard(model.quoteData.Quote)

	// Check for rounded border characters
	assert.Contains(t, rendered, "╭", "Should contain top-left rounded corner")
	assert.Contains(t, rendered, "╮", "Should contain top-right rounded corner")
	assert.Contains(t, rendered, "╰", "Should contain bottom-left rounded corner")
	assert.Contains(t, rendered, "╯", "Should contain bottom-right rounded corner")
}

// TestQuoteView_IndicatorsCard_RSIBar verifies RSI bar renders with █ and ░ characters.
func TestQuoteView_IndicatorsCard_RSIBar(t *testing.T) {
	model := NewModel()
	model.width = 80

	view := NewView(model).SetTheme(&defaultTheme{})

	// Test with RSI = 50 (middle of range)
	rsiBar := view.renderRSIBar(50)

	assert.Contains(t, rsiBar, "█", "Should contain filled block character")
	assert.Contains(t, rsiBar, "░", "Should contain empty block character")
	assert.Contains(t, rsiBar, "0", "Should contain zone marker 0")
	assert.Contains(t, rsiBar, "30", "Should contain zone marker 30")
	assert.Contains(t, rsiBar, "70", "Should contain zone marker 70")
	assert.Contains(t, rsiBar, "100", "Should contain zone marker 100")
}

// TestQuoteView_IndicatorsCard_FTEMABadge_Bull verifies bullish FTEMA badge.
func TestQuoteView_IndicatorsCard_FTEMABadge_Bull(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bullish,
			BlitzScore:   0,
			DestinyScore: 0,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.Contains(t, card, "▲  LONG", "Bullish FTEMA should show ▲ LONG badge")
}

// TestQuoteView_IndicatorsCard_FTEMABadge_Bear verifies bearish FTEMA badge.
func TestQuoteView_IndicatorsCard_FTEMABadge_Bear(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bearish,
			BlitzScore:   0,
			DestinyScore: 0,
			TPI:          -0.33,
			TPISignal:    "CASH",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.Contains(t, card, "▼ SHORT", "Bearish FTEMA should show ▼ SHORT badge")
}

// TestQuoteView_IndicatorsCard_BlitzBadge verifies BLITZ badge rendering.
func TestQuoteView_IndicatorsCard_BlitzBadge(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 0,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.Contains(t, card, "▲  LONG", "BLITZ LONG should show ▲ LONG badge")

	// Test BLITZ SHORT
	model.quoteData.Indicators.BlitzScore = -1
	model.quoteData.Indicators.TPI = -0.67
	model.quoteData.Indicators.TPISignal = "CASH"
	card = view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")
	assert.Contains(t, card, "▼ SHORT", "BLITZ SHORT should show ▼ SHORT badge")

	// Test BLITZ HOLD - should show nothing
	model.quoteData.Indicators.BlitzScore = 0
	card = view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")
	// The line should only have the label, not the badge
	assert.NotContains(t, card, "─ HOLD", "BLITZ HOLD should not show badge")
}

// TestQuoteView_IndicatorsCard_DestinyBadge verifies DESTINY badge rendering.
func TestQuoteView_IndicatorsCard_DestinyBadge(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bullish,
			BlitzScore:   0,
			DestinyScore: 1,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.Contains(t, card, "▲  LONG", "DESTINY LONG should show ▲ LONG badge")

	// Test DESTINY SHORT
	model.quoteData.Indicators.DestinyScore = -1
	model.quoteData.Indicators.TPI = -0.67
	model.quoteData.Indicators.TPISignal = "CASH"
	card = view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")
	assert.Contains(t, card, "▼ SHORT", "DESTINY SHORT should show ▼ SHORT badge")

	// Test DESTINY HOLD - should show nothing
	model.quoteData.Indicators.DestinyScore = 0
	card = view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")
	// The line should only have the label, not the badge
	assert.NotContains(t, card, "○ HOLD", "DESTINY HOLD should not show badge")
}

// TestQuoteView_IndicatorsCard_TPIValue verifies TPI numeric value display.
func TestQuoteView_IndicatorsCard_TPIValue(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.Contains(t, card, "+0.67", "Positive TPI should show with + sign")

	// Test negative TPI
	model.quoteData.Indicators.TPI = -0.33
	model.quoteData.Indicators.TPISignal = "CASH"
	card = view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")
	assert.Contains(t, card, "-0.33", "Negative TPI should show with - sign")
}

// TestQuoteView_IndicatorsCard_TPIGauge verifies TPI gauge rendering.
func TestQuoteView_IndicatorsCard_TPIGauge(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.Contains(t, card, "▓", "TPI gauge should contain filled block character")
	assert.Contains(t, card, "░", "TPI gauge should contain empty block character")
}

// TestQuoteView_IndicatorsCard_TPISignal_Long verifies LONG signal for positive TPI.
func TestQuoteView_IndicatorsCard_TPISignal_Long(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.Contains(t, card, "LONG", "Positive TPI should show LONG signal")
	// Verify gauge and signal are on same line (gauge + signal should be together)
	lines := strings.Split(card, "\n")
	for _, line := range lines {
		if strings.Contains(line, "▓") && strings.Contains(line, "LONG") {
			// Found a line with both gauge and signal
			return
		}
	}
	// If we get here, the gauge and signal aren't on the same line
	t.Error("TPI gauge and signal should be on the same line")
}

// TestQuoteView_IndicatorsCard_TPISignal_Cash verifies CASH signal for negative/zero TPI.
func TestQuoteView_IndicatorsCard_TPISignal_Cash(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bearish,
			BlitzScore:   -1,
			DestinyScore: -1,
			TPI:          -0.33,
			TPISignal:    "CASH",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.Contains(t, card, "CASH", "Negative TPI should show CASH signal")

	// Test zero TPI (also shows CASH)
	model.quoteData.Indicators.TPI = 0
	model.quoteData.Indicators.TPISignal = "CASH"
	card = view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")
	assert.Contains(t, card, "CASH", "Zero TPI should show CASH signal")
}

// TestQuoteView_IndicatorsCard_NoEMA verifies EMA values are not displayed.
func TestQuoteView_IndicatorsCard_NoEMA(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.NotContains(t, card, "EMA(10)", "Should not display EMA(10)")
	assert.NotContains(t, card, "EMA(20)", "Should not display EMA(20)")
}

// TestQuoteView_IndicatorsCard_NoTrendLabel verifies old Trend label is not displayed.
func TestQuoteView_IndicatorsCard_NoTrendLabel(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.NotContains(t, card, "Trend:", "Should not display old Trend: label")
}

// TestQuoteView_IndicatorsCard_RSIPreserved verifies RSI bar still renders.
func TestQuoteView_IndicatorsCard_RSIPreserved(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          50,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")

	assert.Contains(t, card, "RSI(14):", "Should display RSI label")
	assert.Contains(t, card, "50.0", "Should display RSI value")
	assert.Contains(t, card, "Fair value", "Should display RSI valuation")
	assert.Contains(t, card, "█", "Should display RSI progress bar")
}

// TestQuoteView_ChangeDisplay_BullishIcon verifies ▲ icon for positive change.
func TestQuoteView_ChangeDisplay_BullishIcon(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:        "AAPL",
			Price:         "189.84",
			Change:        "2.34",
			ChangePercent: "1.25%",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderPriceCard(model.quoteData.Quote)

	assert.Contains(t, card, "▲", "Positive change should show up arrow")
}

// TestQuoteView_ChangeDisplay_BearishIcon verifies ▼ icon for negative change.
func TestQuoteView_ChangeDisplay_BearishIcon(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:        "AAPL",
			Price:         "189.84",
			Change:        "-2.34",
			ChangePercent: "-1.25%",
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})
	card := view.renderPriceCard(model.quoteData.Quote)

	assert.Contains(t, card, "▼", "Negative change should show down arrow")
}

// TestQuoteView_ErrorCard_RedBorder verifies error card with red border.
func TestQuoteView_ErrorCard_RedBorder(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateError
	model.err = errors.New("test error")

	view := NewView(model).SetTheme(&defaultTheme{})
	rendered := view.renderError()

	// Check for ✗ in error message
	assert.Contains(t, rendered, "✗", "Error should contain X symbol")
	assert.Contains(t, rendered, "Error loading quote", "Should contain error message")
}

// TestQuoteView_InputLabel verifies "Symbol" label above input.
func TestQuoteView_InputLabel(t *testing.T) {
	model := NewModel()
	model.width = 80

	view := NewView(model).SetTheme(&defaultTheme{})
	rendered := view.renderIdle()

	assert.Contains(t, rendered, "Symbol", "Should contain Symbol label")
	assert.Contains(t, rendered, "(e.g. AAPL, BTC, ETH)", "Should contain example hint")
}

// TestQuoteView_EmDash verifies em dash used instead of double dash.
func TestQuoteView_EmDash(t *testing.T) {
	// Test parseChange
	_, changePercent := parseChange("", "")
	assert.Equal(t, "—", changePercent, "Empty change should return em dash")

	// Test formatVolume
	volume := formatVolume("", "AAPL", 100)
	assert.Equal(t, "—", volume, "Empty volume should return em dash")

	// Test formatPrice
	price := formatPrice("")
	assert.Equal(t, "—", price, "Empty price should return em dash")

	// Test formatPercent
	percent := formatPercent("")
	assert.Equal(t, "—", percent, "Empty percent should return em dash")
}

// TestQuoteModel_GetLookupHistory_Empty verifies empty slice for fresh model.
func TestQuoteModel_GetLookupHistory_Empty(t *testing.T) {
	model := NewModel()

	history := model.GetLookupHistory()

	if history == nil {
		t.Error("Expected non-nil history slice")
	}

	if len(history) != 0 {
		t.Errorf("Expected empty history, got %v", history)
	}
}

// TestQuoteModel_GetLookupHistory_AfterLookups verifies history contains lookups.
func TestQuoteModel_GetLookupHistory_AfterLookups(t *testing.T) {
	model := NewModel()

	// Add 3 tickers to history
	model.addToHistory("AAPL")
	model.addToHistory("MSFT")
	model.addToHistory("GOOGL")

	history := model.GetLookupHistory()

	if len(history) != 3 {
		t.Errorf("Expected history length 3, got %d", len(history))
	}

	expected := []string{"AAPL", "MSFT", "GOOGL"}
	for i, expectedTicker := range expected {
		if history[i] != expectedTicker {
			t.Errorf("Expected %s at index %d, got %s", expectedTicker, i, history[i])
		}
	}
}

// TestQuoteView_PriceRendered verifies price is rendered correctly.
func TestQuoteView_PriceRendered(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol:         "AAPL",
			Price:          "155.50",
			Change:         "2.34",
			ChangePercent:  "1.25%",
			Open:           "153.00",
			High:           "156.00",
			Low:            "152.50",
			Volume:         "50000000",
			PreviousClose:  "153.16",
			LastTradingDay: "2026-04-01",
		},
		Indicators: &trenddomain.Result{
			Symbol:       "AAPL",
			RSI:          52.3,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			TPI:          0.67,
			TPISignal:    "LONG",
		},
	}

	view := model.View()

	// Price should be rendered - look for "$155.50" or "155.50"
	if !assert.Contains(t, view, "155.50", "View should contain price 155.50") {
		t.Logf("Full view output:\n%s", view)
	}
}

// TestQuoteView_DayRangeBar_Normal verifies day range bar with valid data.
func TestQuoteView_DayRangeBar_Normal(t *testing.T) {
	model := NewModel()
	model.width = 80
	view := NewView(model).SetTheme(&defaultTheme{})

	bar := view.renderDayRangeBar(155.00, "150", "160")

	assert.Contains(t, bar, "150", "Should contain low value")
	assert.Contains(t, bar, "160", "Should contain high value")
	assert.Contains(t, bar, "▓", "Should contain filled block character")
	assert.Contains(t, bar, "░", "Should contain empty block character")
}

// TestQuoteView_DayRangeBar_MissingLow verifies day range bar is skipped when Low is missing.
func TestQuoteView_DayRangeBar_MissingLow(t *testing.T) {
	model := NewModel()
	model.width = 80
	view := NewView(model).SetTheme(&defaultTheme{})

	bar := view.renderDayRangeBar(155.00, "", "160")

	assert.Empty(t, bar, "Should return empty string when Low is missing")
}

// TestQuoteView_DayRangeBar_MissingHigh verifies day range bar is skipped when High is missing.
func TestQuoteView_DayRangeBar_MissingHigh(t *testing.T) {
	model := NewModel()
	model.width = 80
	view := NewView(model).SetTheme(&defaultTheme{})

	bar := view.renderDayRangeBar(155.00, "150", "")

	assert.Empty(t, bar, "Should return empty string when High is missing")
}

// TestQuoteView_DayRangeBar_PriceAtLow verifies bar shows minimal fill at low.
func TestQuoteView_DayRangeBar_PriceAtLow(t *testing.T) {
	model := NewModel()
	model.width = 80
	view := NewView(model).SetTheme(&defaultTheme{})

	bar := view.renderDayRangeBar(150.00, "150", "160")

	assert.Contains(t, bar, "░", "Should contain empty block character when at low")
}

// TestQuoteView_DayRangeBar_PriceAtHigh verifies bar is fully filled at high.
func TestQuoteView_DayRangeBar_PriceAtHigh(t *testing.T) {
	model := NewModel()
	model.width = 80
	view := NewView(model).SetTheme(&defaultTheme{})

	bar := view.renderDayRangeBar(160.00, "150", "160")

	// When price is at high, bar should be fully filled (no empty blocks)
	assert.Contains(t, bar, "▓", "Should contain filled block character")
	assert.NotContains(t, bar, "░", "Should not contain empty block character when at high")
}

// TestQuoteView_IdleRecentLookups verifies recent lookups shown in idle state.
func TestQuoteView_IdleRecentLookups(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.lookupHistory = []string{"AAPL", "MSFT", "GOOGL"}

	view := NewView(model).SetTheme(&defaultTheme{})
	rendered := view.renderIdle()

	assert.Contains(t, rendered, "Recent:", "Should show Recent: label")
	assert.Contains(t, rendered, "AAPL", "Should show AAPL")
	assert.Contains(t, rendered, "MSFT", "Should show MSFT")
	assert.Contains(t, rendered, "GOOGL", "Should show GOOGL")
}

// TestQuoteView_IdleNoHistory verifies no recent section when history is empty.
func TestQuoteView_IdleNoHistory(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.lookupHistory = []string{}

	view := NewView(model).SetTheme(&defaultTheme{})
	rendered := view.renderIdle()

	assert.NotContains(t, rendered, "Recent:", "Should not show Recent: label when history is empty")
}

// TestQuoteView_IdleRecentMaxFive verifies maximum 5 recent entries shown.
func TestQuoteView_IdleRecentMaxFive(t *testing.T) {
	model := NewModel()
	model.width = 80
	// Add 8 entries to history
	model.lookupHistory = []string{"AAA", "BBB", "CCC", "DDD", "EEE", "FFF", "GGG", "HHH"}

	view := NewView(model).SetTheme(&defaultTheme{})
	rendered := view.renderIdle()

	assert.Contains(t, rendered, "Recent:", "Should show Recent: label")
	// Check that only 5 entries are shown (most recent 5: DDD, EEE, FFF, GGG, HHH)
	assert.Contains(t, rendered, "DDD", "Should show DDD (4th most recent)")
	assert.Contains(t, rendered, "EEE", "Should show EEE (5th most recent)")
	assert.Contains(t, rendered, "FFF", "Should show FFF (6th most recent)")
	assert.Contains(t, rendered, "GGG", "Should show GGG (7th most recent)")
	assert.Contains(t, rendered, "HHH", "Should show HHH (8th most recent)")
	assert.NotContains(t, rendered, "AAA", "Should not show AAA (oldest, beyond limit)")
	assert.NotContains(t, rendered, "BBB", "Should not show BBB (second oldest, beyond limit)")
	assert.NotContains(t, rendered, "CCC", "Should not show CCC (third oldest, beyond limit)")
}

// TestQuoteView_ErrorRetryHint verifies modern retry hint in error state.
func TestQuoteView_ErrorRetryHint(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateError
	model.err = errors.New("test error")

	view := NewView(model).SetTheme(&defaultTheme{})
	rendered := view.renderError()

	assert.Contains(t, rendered, "retry", "Should show 'retry' hint")
	assert.Contains(t, rendered, "clear", "Should show 'clear' hint")
	assert.Contains(t, rendered, "Enter", "Should show Enter key hint")
	assert.Contains(t, rendered, "Esc", "Should show Esc key hint")
}

// TestFetchQuoteCmd_RoutesCommodityToCommodityEndpoint verifies that commodity symbols
// are routed to the commodity quote endpoint.
func TestFetchQuoteCmd_RoutesCommodityToCommodityEndpoint(t *testing.T) {
	model := NewModel()
	ctx := context.Background()

	// Create mock commodity client with test data
	mockCommodity := &mockCommodityClient{
		series: &alphavantage.CommoditySeries{
			Name:     "West Texas Intermediate Crude Oil",
			Interval: "daily",
			Unit:     "USD per Barrel",
			Data: []alphavantage.CommodityDataPoint{
				{Value: 75.50, Date: mustParseTime("2024-01-15")},
				{Value: 74.00, Date: mustParseTime("2024-01-14")},
			},
		},
	}

	// Configure with commodity detector and mock client
	detector := indicators.NewAssetClassDetector(nil, indicators.DefaultCommoditySymbols)
	model.Configure(ctx, mockCommodity, &mockEngine{}, detector)

	// Execute the command
	cmd := model.fetchQuoteCmd("WTI")
	msg := cmd()

	resultMsg, ok := msg.(QuoteResultMsg)
	if !ok {
		t.Fatalf("Expected QuoteResultMsg, got %T", msg)
	}

	// Verify quote was created with commodity data
	if resultMsg.Data.Quote == nil {
		t.Fatal("Expected quote to be populated")
	}

	if resultMsg.Data.Quote.Symbol != "WTI" {
		t.Errorf("Expected symbol WTI, got %s", resultMsg.Data.Quote.Symbol)
	}

	// Verify price from commodity data
	price, err := alphavantage.ParseFloat(resultMsg.Data.Quote.Price)
	if err != nil {
		t.Fatalf("Failed to parse price: %v", err)
	}
	if price != 75.50 {
		t.Errorf("Expected price 75.50, got %.2f", price)
	}

	// Verify change calculation
	change, err := alphavantage.ParseFloat(resultMsg.Data.Quote.Change)
	if err != nil {
		t.Fatalf("Failed to parse change: %v", err)
	}
	expectedChange := 75.50 - 74.00
	if change != expectedChange {
		t.Errorf("Expected change %.2f, got %.2f", expectedChange, change)
	}

	// Verify indicators are nil for commodities
	if resultMsg.Data.Indicators != nil {
		t.Error("Expected indicators to be nil for commodities")
	}
}

// TestFetchQuoteCmd_RoutesCryptoUnchanged verifies that crypto symbols
// are still routed to the crypto quote endpoint (regression test).
func TestFetchQuoteCmd_RoutesCryptoUnchanged(t *testing.T) {
	model := NewModel()
	ctx := context.Background()

	// Create mock crypto client with cryptoDaily data
	mockCrypto := &mockClient{
		cryptoDaily: &alphavantage.CryptoDaily{
			TimeSeries: map[string]alphavantage.CryptoEntry{
				"2024-01-15": {
					Open:   "49500.00",
					High:   "51000.00",
					Low:    "49000.00",
					Close:  "50000.00",
					Volume: "1000.0",
				},
				"2024-01-14": {
					Open:   "49200.00",
					High:   "49800.00",
					Low:    "49000.00",
					Close:  "49500.00",
					Volume: "950.0",
				},
			},
		},
	}

	// Create mock engine with a result
	mockEngine := &mockEngine{
		result: &trenddomain.Result{
			RSI:          50.0,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			FlowScore:    1,
			VortexScore:  1,
			TPI:          0.75,
			TPISignal:    "Bullish",
		},
	}

	// Configure with crypto detector
	detector := indicators.NewAssetClassDetector([]string{"BTC"}, nil)
	model.Configure(ctx, mockCrypto, mockEngine, detector)

	// Execute the command
	cmd := model.fetchQuoteCmd("BTC")
	msg := cmd()

	resultMsg, ok := msg.(QuoteResultMsg)
	if !ok {
		t.Fatalf("Expected QuoteResultMsg, got %T", msg)
	}

	if resultMsg.Data.Quote.Symbol != "BTC" {
		t.Errorf("Expected symbol BTC, got %s", resultMsg.Data.Quote.Symbol)
	}

	// Verify indicators are fetched for crypto (not nil)
	if resultMsg.Data.Indicators == nil {
		t.Error("Expected indicators to be populated for crypto")
	}
}

// TestFetchQuoteCmd_RoutesEquityUnchanged verifies that equity symbols
// are still routed to the equity quote endpoint (regression test).
func TestFetchQuoteCmd_RoutesEquityUnchanged(t *testing.T) {
	model := NewModel()
	ctx := context.Background()

	// Create mock equity client
	mockEquity := &mockClient{
		quote: &alphavantage.GlobalQuote{
			Symbol:         "AAPL",
			Price:          "185.00",
			Open:           "183.00",
			High:           "186.00",
			Low:            "182.00",
			Volume:         "50000000",
			LastTradingDay: "2024-01-15",
			PreviousClose:  "183.00",
			Change:         "2.00",
			ChangePercent:  "1.09%",
		},
	}

	// Create mock engine with a result
	mockEngine := &mockEngine{
		result: &trenddomain.Result{
			RSI:          55.0,
			Signal:       trenddomain.Bullish,
			BlitzScore:   1,
			DestinyScore: 1,
			FlowScore:    1,
			VortexScore:  1,
			TPI:          0.50,
			TPISignal:    "Bullish",
		},
	}

	// Configure with default detector (no crypto/commodity symbols)
	detector := indicators.NewAssetClassDetector(nil, nil)
	model.Configure(ctx, mockEquity, mockEngine, detector)

	// Execute the command
	cmd := model.fetchQuoteCmd("AAPL")
	msg := cmd()

	resultMsg, ok := msg.(QuoteResultMsg)
	if !ok {
		t.Fatalf("Expected QuoteResultMsg, got %T", msg)
	}

	if resultMsg.Data.Quote.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", resultMsg.Data.Quote.Symbol)
	}

	// Verify indicators are fetched for equity (not nil)
	if resultMsg.Data.Indicators == nil {
		t.Error("Expected indicators to be populated for equity")
	}
}

// TestFetchCommodityAsQuote_BuildsValidGlobalQuote verifies that commodity data
// is correctly converted to a GlobalQuote structure.
func TestFetchCommodityAsQuote_BuildsValidGlobalQuote(t *testing.T) {
	model := NewModel()
	ctx := context.Background()

	// Create mock commodity client with test data
	mockCommodity := &mockCommodityClient{
		series: &alphavantage.CommoditySeries{
			Name:     "Wheat",
			Interval: "daily",
			Unit:     "USD per Bushel",
			Data: []alphavantage.CommodityDataPoint{
				{Value: 6.50, Date: mustParseTime("2024-01-15")},
				{Value: 6.40, Date: mustParseTime("2024-01-14")},
			},
		},
	}

	detector := indicators.NewAssetClassDetector(nil, indicators.DefaultCommoditySymbols)
	model.Configure(ctx, mockCommodity, &mockEngine{}, detector)

	// Fetch commodity as quote
	quote, err := model.fetchCommodityAsQuote("WHEAT")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify quote structure
	if quote.Symbol != "WHEAT" {
		t.Errorf("Expected symbol WHEAT, got %s", quote.Symbol)
	}

	// Verify OHLC are all the same (commodity API doesn't provide separate values)
	if quote.Open != quote.High || quote.High != quote.Low || quote.Low != quote.Price {
		t.Error("Expected Open, High, Low, and Price to be the same for commodities")
	}

	// Verify volume is empty (commodity API doesn't provide it)
	if quote.Volume != "" {
		t.Errorf("Expected empty volume, got %s", quote.Volume)
	}

	// Verify last trading day
	if quote.LastTradingDay != "2024-01-15" {
		t.Errorf("Expected last trading day 2024-01-15, got %s", quote.LastTradingDay)
	}

	// Verify change calculation
	change, _ := alphavantage.ParseFloat(quote.Change)
	expectedChange := 6.50 - 6.40
	if change != expectedChange {
		t.Errorf("Expected change %.2f, got %.2f", expectedChange, change)
	}
}

// TestFetchCommodityAsQuote_HandlesSinglePointSeries verifies that a single
// data point is handled gracefully (no change calculation).
func TestFetchCommodityAsQuote_HandlesSinglePointSeries(t *testing.T) {
	model := NewModel()
	ctx := context.Background()

	// Create mock commodity client with single data point
	mockCommodity := &mockCommodityClient{
		series: &alphavantage.CommoditySeries{
			Name:     "Copper",
			Interval: "monthly",
			Unit:     "USD per Metric Ton",
			Data: []alphavantage.CommodityDataPoint{
				{Value: 8500.00, Date: mustParseTime("2024-01-01")},
			},
		},
	}

	detector := indicators.NewAssetClassDetector(nil, indicators.DefaultCommoditySymbols)
	model.Configure(ctx, mockCommodity, &mockEngine{}, detector)

	// Fetch commodity as quote
	quote, err := model.fetchCommodityAsQuote("COPPER")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify quote was created
	if quote == nil {
		t.Fatal("Expected quote to be created")
	}

	// Verify no change is calculated (only one data point)
	if quote.Change != "" {
		t.Errorf("Expected empty change, got %s", quote.Change)
	}

	if quote.ChangePercent != "" {
		t.Errorf("Expected empty change percent, got %s", quote.ChangePercent)
	}
}

// TestFetchCommodityAsQuote_PropagatesAPIError verifies that API errors
// are properly propagated.
func TestFetchCommodityAsQuote_PropagatesAPIError(t *testing.T) {
	model := NewModel()
	ctx := context.Background()

	// Create mock commodity client that returns an error
	mockCommodity := &mockCommodityClient{
		err: errors.New("API rate limit exceeded"),
	}

	detector := indicators.NewAssetClassDetector(nil, indicators.DefaultCommoditySymbols)
	model.Configure(ctx, mockCommodity, &mockEngine{}, detector)

	// Fetch commodity as quote should fail
	_, err := model.fetchCommodityAsQuote("WTI")
	if err == nil {
		t.Error("Expected error to be returned")
	}

	if !strings.Contains(err.Error(), "API rate limit exceeded") {
		t.Errorf("Expected error to contain 'API rate limit exceeded', got %v", err)
	}
}

// TestQuoteView_CommoditySuppressesIndicatorsPanel verifies that the indicators
// panel is suppressed for commodities.
func TestQuoteView_CommoditySuppressesIndicatorsPanel(t *testing.T) {
	model := NewModel()
	model.width = 120
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol: "WTI",
			Price:  "75.50",
		},
		Indicators: nil,
	}
	detector := indicators.NewAssetClassDetector(nil, indicators.DefaultCommoditySymbols)
	model.detector = detector

	view := NewView(model).SetTheme(&defaultTheme{})
	rendered := view.renderQuoteCards()

	// Verify muted message is shown for commodities
	assert.Contains(t, rendered, "signals unavailable for commodities", "Should show muted message for commodities")

	// Verify technical indicators header is shown but with muted message
	assert.Contains(t, rendered, "Technical Indicators", "Should show Technical Indicators header")
}

// TestQuoteView_CommodityShowsMutedMessage verifies the muted message
// is displayed for commodities.
func TestQuoteView_CommodityShowsMutedMessage(t *testing.T) {
	model := NewModel()
	model.width = 120
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{
			Symbol: "WHEAT",
			Price:  "6.50",
		},
		Indicators: nil,
	}
	detector := indicators.NewAssetClassDetector(nil, indicators.DefaultCommoditySymbols)
	model.detector = detector

	view := NewView(model).SetTheme(&defaultTheme{})
	rendered := view.renderCommodityIndicatorsMessage()

	// Verify the key components are present
	assert.Contains(t, rendered, "Technical Indicators", "Should show Technical Indicators header")
	assert.Contains(t, rendered, "signals unavailable for commodities", "Should show muted message")
}

// TestModel_IsCommodity verifies the IsCommodity method works correctly.
func TestModel_IsCommodity(t *testing.T) {
	tests := []struct {
		name      string
		symbol    string
		commodity bool
	}{
		{"WTI is commodity", "WTI", true},
		{"WHEAT is commodity", "WHEAT", true},
		{"BRENT is commodity", "BRENT", true},
		{"BTC is not commodity", "BTC", false},
		{"AAPL is not commodity", "AAPL", false},
		{"No quote data", "", false},
		{"Nil quote data", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			if tt.symbol != "" {
				model.quoteData = &QuoteData{
					Quote: &alphavantage.GlobalQuote{
						Symbol: tt.symbol,
					},
				}
				detector := indicators.NewAssetClassDetector(nil, indicators.DefaultCommoditySymbols)
				model.detector = detector
			}

			result := model.IsCommodity()
			if result != tt.commodity {
				t.Errorf("IsCommodity() = %v, want %v", result, tt.commodity)
			}
		})
	}
}

// mustParseTime is a helper to parse a time string in tests.
func mustParseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}
