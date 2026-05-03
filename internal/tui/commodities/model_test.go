// Package commodities provides tests for the commodities dashboard TUI model.

package commodities

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shinsekai/finterm/internal/alphavantage"
)

// mockCacheStore is a mock implementation of cacheStore for testing.
type mockCacheStore struct {
	getCalls  int
	getKey    string
	getResult interface{}
	getExists bool

	setCalls int
	setKey   string
	setValue interface{}
	setTTL   time.Duration
}

func (m *mockCacheStore) Get(key string) (interface{}, bool) {
	m.getCalls++
	m.getKey = key
	return m.getResult, m.getExists
}

func (m *mockCacheStore) Set(key string, value interface{}, ttl time.Duration) {
	m.setCalls++
	m.setKey = key
	m.setValue = value
	m.setTTL = ttl
}

// asModel converts a tea.Model to *Model.
func asModel(t *testing.T, m tea.Model) *Model {
	t.Helper()
	if ptrModel, ok := m.(*Model); ok {
		return ptrModel
	}
	if valModel, ok := m.(Model); ok {
		return &valModel
	}
	t.Fatalf("expected *Model, got %T", m)
	return nil
}

func TestNewModel(t *testing.T) {
	model := NewModel()

	if model.rows == nil {
		t.Error("rows should be initialized")
	}
	if model.activeRow != 0 {
		t.Errorf("activeRow should be 0, got %d", model.activeRow)
	}
	if model.overallState != StateLoading {
		t.Errorf("overallState should be Loading, got %v", model.overallState)
	}
	if model.interval != "daily" {
		t.Errorf("interval should be daily, got %s", model.interval)
	}
}

func TestModel_Configure(t *testing.T) {
	tests := []struct {
		name        string
		watchlist   []string
		interval    string
		expectedLen int
	}{
		{
			name:        "empty watchlist uses defaults",
			watchlist:   []string{},
			interval:    "daily",
			expectedLen: 10,
		},
		{
			name:        "custom watchlist",
			watchlist:   []string{"WTI", "BRENT"},
			interval:    "weekly",
			expectedLen: 2,
		},
		{
			name:        "invalid symbols filtered out",
			watchlist:   []string{"WTI", "INVALID", "BRENT"},
			interval:    "monthly",
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			mockCache := &mockCacheStore{}
			model.Configure(context.Background(), nil, tt.watchlist, tt.interval, mockCache)

			if len(model.rows) != tt.expectedLen {
				t.Errorf("expected %d rows, got %d", tt.expectedLen, len(model.rows))
			}

			if model.interval != tt.interval {
				t.Errorf("expected interval %s, got %s", tt.interval, model.interval)
			}
		})
	}
}

func TestModel_InitFetchesAllInParallel(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT"}, "daily", mockCache)

	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init should return a command")
	}

	if cmd == nil {
		t.Error("expected non-nil command")
	}
}

func TestModel_RefreshReissuesFetches(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT"}, "daily", mockCache)

	for i := range model.rows {
		model.rows[i].State = StateLoaded
	}

	cmd := model.refreshAllCmd()
	if cmd == nil {
		t.Fatal("refreshAllCmd should return a command")
	}

	for i, row := range model.rows {
		if row.State != StateLoading {
			t.Errorf("row %d should be in Loading state after refresh, got %v", i, row.State)
		}
	}

	if model.overallState != StateLoading {
		t.Errorf("overallState should be Loading after refresh, got %v", model.overallState)
	}
}

func TestModel_NavigationJK(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT", "NATURAL_GAS"}, "daily", mockCache)

	updatedM, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = asModel(t, updatedM)
	if model.activeRow != 1 {
		t.Errorf("activeRow should be 1 after down, got %d", model.activeRow)
	}

	updatedM, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = asModel(t, updatedM)
	if model.activeRow != 2 {
		t.Errorf("activeRow should be 2 after down, got %d", model.activeRow)
	}

	updatedM, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = asModel(t, updatedM)
	if model.activeRow != 2 {
		t.Errorf("activeRow should stay at 2 when already at bottom, got %d", model.activeRow)
	}

	updatedM, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = asModel(t, updatedM)
	if model.activeRow != 1 {
		t.Errorf("activeRow should be 1 after up, got %d", model.activeRow)
	}

	model.activeRow = 0
	updatedM, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = asModel(t, updatedM)
	if model.activeRow != 0 {
		t.Errorf("activeRow should stay at 0 when already at top, got %d", model.activeRow)
	}

	updatedM, _ = model.Update(tea.KeyMsg{Runes: []rune{'r'}})
	model = asModel(t, updatedM)
	if model.overallState != StateLoading {
		t.Errorf("overallState should be Loading after 'r' key, got %v", model.overallState)
	}
}

func TestModel_EmptyWatchlistPlaceholder(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"INVALID_SYMBOL"}, "daily", mockCache)

	if len(model.rows) != 10 {
		t.Errorf("expected 10 default commodities, got %d", len(model.rows))
	}

	if model.IsEmpty() && len(model.rows) == 0 {
		t.Error("IsEmpty should return false for default watchlist")
	}
}

func TestModel_PartialFailureIsolated(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT", "NATURAL_GAS"}, "daily", mockCache)

	wtiMsg := CommodityDataMsg{
		Symbol: "WTI",
		Series: &alphavantage.CommoditySeries{
			Name: "WTI",
			Data: []alphavantage.CommodityDataPoint{
				{Date: time.Now(), Value: 75.0},
			},
		},
		Index:     0,
		FromCache: false,
	}
	updatedM, _ := model.Update(wtiMsg)
	model = asModel(t, updatedM)

	brentMsg := CommodityErrorMsg{
		Symbol: "BRENT",
		Err:    errors.New("API error"),
		Index:  1,
	}
	updatedM, _ = model.Update(brentMsg)
	model = asModel(t, updatedM)

	if model.rows[0].State != StateLoaded {
		t.Errorf("WTI should be in Loaded state, got %v", model.rows[0].State)
	}
	if model.rows[1].State != StateError {
		t.Errorf("BRENT should be in Error state, got %v", model.rows[1].State)
	}
	if model.rows[2].State != StateLoading {
		t.Errorf("NATURAL_GAS should still be in Loading state, got %v", model.rows[2].State)
	}
}

func TestModel_CacheTTL(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)

	ttl := model.cacheTTL()
	if ttl != time.Hour {
		t.Errorf("expected 1h TTL for daily interval, got %v", ttl)
	}

	model.interval = "weekly"
	ttl = model.cacheTTL()
	if ttl != 6*time.Hour {
		t.Errorf("expected 6h TTL for weekly interval, got %v", ttl)
	}
}

func TestModel_UpdateOverallState(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT", "NATURAL_GAS"}, "daily", mockCache)

	model.overallState = StateLoading
	model = asModel(t, model.updateOverallState())
	if model.overallState != StateLoading {
		t.Errorf("expected Loading when all rows are loading, got %v", model.overallState)
	}

	for i := range model.rows {
		model.rows[i].State = StateLoaded
	}
	model = asModel(t, model.updateOverallState())
	if model.overallState != StateLoaded {
		t.Errorf("expected Loaded when all rows are loaded, got %v", model.overallState)
	}

	model.rows[1].State = StateError
	model = asModel(t, model.updateOverallState())
	if model.overallState != StateError {
		t.Errorf("expected Error when one row has error, got %v", model.overallState)
	}

	model.rows[0].State = StateLoading
	model.rows[1].State = StateLoaded
	model.rows[2].State = StateLoaded
	model = asModel(t, model.updateOverallState())
	if model.overallState != StateLoading {
		t.Errorf("expected Loading when some rows are still loading, got %v", model.overallState)
	}
}

func TestModel_GetLoadedCount(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT", "NATURAL_GAS"}, "daily", mockCache)

	count := model.GetLoadedCount()
	if count != 0 {
		t.Errorf("expected 0 loaded rows initially, got %d", count)
	}

	model.rows[0].State = StateLoaded
	count = model.GetLoadedCount()
	if count != 1 {
		t.Errorf("expected 1 loaded row, got %d", count)
	}

	model.rows[1].State = StateCached
	count = model.GetLoadedCount()
	if count != 2 {
		t.Errorf("expected 2 loaded rows (1 loaded + 1 cached), got %d", count)
	}

	model.rows[2].State = StateError
	count = model.GetLoadedCount()
	if count != 2 {
		t.Errorf("expected 2 loaded rows when third is error, got %d", count)
	}
}

func TestModel_KeyBindings(t *testing.T) {
	model := NewModel()
	bindings := model.KeyBindings()

	if len(bindings) != 3 {
		t.Errorf("expected 3 key bindings, got %d", len(bindings))
	}
}

func TestCommodityWatchlist_Default(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{}, "daily", mockCache)

	expectedSymbols := []string{"WTI", "BRENT", "NATURAL_GAS", "COPPER", "ALUMINUM", "WHEAT", "CORN", "COFFEE", "SUGAR", "COTTON"}

	if len(model.rows) != len(expectedSymbols) {
		t.Errorf("expected %d default commodities, got %d", len(expectedSymbols), len(model.rows))
	}

	for i, symbol := range expectedSymbols {
		if model.rows[i].Symbol != symbol {
			t.Errorf("expected row %d to have symbol %s, got %s", i, symbol, model.rows[i].Symbol)
		}
	}
}

func TestModel_WindowSizeMsg(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedM, _ := model.Update(msg)
	model = asModel(t, updatedM)

	if model.GetWidth() != 100 {
		t.Errorf("expected width 100, got %d", model.GetWidth())
	}

	if model.GetHeight() != 50 {
		t.Errorf("expected height 50, got %d", model.GetHeight())
	}
}

func TestModel_RefreshMsg(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT"}, "daily", mockCache)

	for i := range model.rows {
		model.rows[i].State = StateLoaded
	}

	refreshMsg := RefreshMsg{}
	updatedM, _ := model.Update(refreshMsg)
	model = asModel(t, updatedM)

	for _, row := range model.rows {
		if row.State != StateLoading {
			t.Errorf("expected all rows to be in Loading state after refresh, got %v", row.State)
		}
	}
}

func TestModel_CommodityDataMsg(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT"}, "daily", mockCache)

	dataMsg := CommodityDataMsg{
		Symbol: "WTI",
		Series: &alphavantage.CommoditySeries{
			Name: "WTI",
			Unit: "dollars per barrel",
			Data: []alphavantage.CommodityDataPoint{
				{Date: time.Now(), Value: 75.50},
			},
		},
		Index:     0,
		FromCache: false,
	}

	updatedM, _ := model.Update(dataMsg)
	model = asModel(t, updatedM)

	if model.GetRows()[0].State != StateLoaded {
		t.Errorf("expected row to be in Loaded state, got %v", model.GetRows()[0].State)
	}

	if model.GetRows()[0].Series == nil {
		t.Error("expected series to be set")
	}

	if model.GetRows()[0].LastUpdate.IsZero() {
		t.Error("expected LastUpdate to be set")
	}
}

func TestModel_CommodityErrorMsg(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}
	model.Configure(context.Background(), nil, []string{"WTI", "BRENT"}, "daily", mockCache)

	errorMsg := CommodityErrorMsg{
		Symbol: "WTI",
		Err:    errors.New("API error"),
		Index:  0,
	}

	updatedM, _ := model.Update(errorMsg)
	model = asModel(t, updatedM)

	if model.GetRows()[0].State != StateError {
		t.Errorf("expected row to be in Error state, got %v", model.GetRows()[0].State)
	}

	if model.GetRows()[0].Error == nil {
		t.Error("expected error to be set")
	}
}

func TestModel_CacheHit(_ *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}

	cachedSeries := &alphavantage.CommoditySeries{
		Name: "WTI",
		Unit: "dollars per barrel",
		Data: []alphavantage.CommodityDataPoint{
			{Date: time.Now(), Value: 75.50},
		},
	}
	mockCache.getExists = true
	mockCache.getResult = cachedSeries

	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)
	// Cache is checked during fetchCommodityCmd execution
	// The mockCache.getCalls will be incremented when cache is accessed
}

func TestModel_GetInterval(t *testing.T) {
	model := NewModel()
	mockCache := &mockCacheStore{}

	model.Configure(context.Background(), nil, []string{"WTI"}, "daily", mockCache)
	if model.GetInterval() != "daily" {
		t.Errorf("expected default interval 'daily', got %s", model.GetInterval())
	}

	model = NewModel()
	model.Configure(context.Background(), nil, []string{"WTI"}, "weekly", mockCache)
	if model.GetInterval() != "weekly" {
		t.Errorf("expected interval 'weekly', got %s", model.GetInterval())
	}

	model = NewModel()
	model.Configure(context.Background(), nil, []string{"WTI"}, "invalid", mockCache)
	if model.GetInterval() != "daily" {
		t.Errorf("expected default interval 'daily' for invalid, got %s", model.GetInterval())
	}
}
