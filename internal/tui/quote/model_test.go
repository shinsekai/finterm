// Package quote provides tests for quote TUI model.
package quote

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/owner/finterm/internal/alphavantage"
	trenddomain "github.com/owner/finterm/internal/domain/trend"
	"github.com/owner/finterm/internal/domain/trend/indicators"
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
	quote *alphavantage.GlobalQuote
	err   error
}

func (m *mockClient) GetGlobalQuote(_ context.Context, _ string) (*alphavantage.GlobalQuote, error) {
	return m.quote, m.err
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

	configured := model.Configure(ctx, client, engine, indicators.NewAssetClassDetector(nil))

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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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
		Symbol:    "AAPL",
		RSI:       52.3,
		EMAFast:   188.20,
		EMASlow:   186.45,
		Signal:    trenddomain.Bullish,
		Valuation: "Fair value",
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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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
				m.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))
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
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

	cmd := model.fetchQuoteCmd("AAPL")
	if cmd == nil {
		t.Error("Expected command to be created")
	}
}

// BenchmarkQuoteModel_Update benchmarks Update method.
func BenchmarkQuoteModel_Update(b *testing.B) {
	model := NewModel()
	model.Configure(context.Background(), &mockClient{}, &mockEngine{}, indicators.NewAssetClassDetector(nil))

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

// TestQuoteView_IndicatorsCard_TrendSignalIcon verifies trend signal icons.
func TestQuoteView_IndicatorsCard_TrendSignalIcon(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.state = StateLoaded
	model.quoteData = &QuoteData{
		Quote: &alphavantage.GlobalQuote{Symbol: "AAPL"},
		Indicators: &trenddomain.Result{
			Symbol:  "AAPL",
			RSI:     50,
			EMAFast: 150,
			EMASlow: 145,
		},
	}

	view := NewView(model).SetTheme(&defaultTheme{})

	// Test bullish
	model.quoteData.Indicators.Signal = trenddomain.Bullish
	card := view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")
	assert.Contains(t, card, "▲", "Bullish should contain up arrow")

	// Test bearish
	model.quoteData.Indicators.Signal = trenddomain.Bearish
	card = view.renderIndicatorsCard(model.quoteData.Indicators, "2026-04-01")
	assert.Contains(t, card, "▼", "Bearish should contain down arrow")
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
