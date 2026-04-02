# EPICS.md — Epics & Tasks

> **For Claude Code**: Each task below is self-contained. Before starting any task,
> read `CLAUDE.md` for project guidelines. Every task includes acceptance criteria
> and required unit tests. Run `go fmt`, `go vet`, and `go test ./...` before marking done.

---

## Epic 1 — Project Scaffolding & Configuration

> Bootstrap the project: Go module, directory structure, config loading, and build tooling.

### Task 1.1 — Initialize Go module and directory structure

**Description**: Create the Go module and the full directory tree as defined in `CLAUDE.md` §2.

**Deliverables**:
- `go.mod` with module path `github.com/owner/finterm` and Go 1.26+.
- All directories created with placeholder `.go` files where needed (to keep Go happy).
- `cmd/finterm/main.go` with a minimal `func main()` that prints "finterm" and exits.
- `Makefile` with targets: `build`, `run`, `test`, `lint`, `fmt`, `vet`, `clean`.
- `.gitignore` covering: `config.yaml`, binaries, coverage files, `.DS_Store`, vendor/.
- `config.example.yaml` with all fields documented (no real API key).

**Acceptance Criteria**:
- [ ] `go build ./cmd/finterm/` compiles without errors.
- [ ] `make build` produces a binary in `./bin/finterm`.
- [ ] `make test` runs (even if no real tests yet) without errors.
- [ ] `.gitignore` blocks `config.yaml` and binary artifacts.

**Tests**: None required (scaffolding only). Verify with `go build` and `make`.

---

### Task 1.2 — Configuration loading and validation

**Description**: Implement the `internal/config/` package to load and validate `config.yaml`.

**Deliverables**:
- `internal/config/config.go`:
  - `Config` struct matching the schema in `CLAUDE.md` §8.1.
  - `Load(path string) (*Config, error)` — reads YAML, applies env var overrides, validates.
  - `Validate() error` — checks required fields, ticker format, duration parsing.
  - Env var override: `FINTERM_AV_API_KEY` overrides `api.key`.
- `internal/config/config_test.go`.

**Acceptance Criteria**:
- [ ] Valid config loads without error.
- [ ] Missing API key returns descriptive error.
- [ ] Env var overrides YAML value.
- [ ] Invalid ticker format (special chars, > 10 chars) returns validation error.
- [ ] Invalid duration strings return parse error.
- [ ] Default values applied for optional fields.

**Tests** (table-driven):
```
TestLoad_ValidConfig
TestLoad_MissingAPIKey
TestLoad_EnvVarOverride
TestLoad_FileNotFound
TestValidate_InvalidTicker
TestValidate_InvalidDuration
TestValidate_Defaults
```

**Guidelines reference**: `CLAUDE.md` §8.

---

### Task 1.3 — Pre-commit hooks and CI setup

**Description**: Create `.pre-commit-config.yaml` with all hooks defined in `CLAUDE.md` §9.

**Deliverables**:
- `.pre-commit-config.yaml` with hooks: `go-fmt`, `go-vet`, `go-test`, `go-lint`, `go-vulncheck`, `no-secrets`, `claude-guidelines-check`.
- `.golangci.yml` with linters: `govet`, `errcheck`, `staticcheck`, `unused`, `gosimple`, `ineffassign`, `misspell`, `gocyclo` (max 15), `gocritic`, `revive`.

**Acceptance Criteria**:
- [ ] `golangci-lint run` passes on the current codebase.
- [ ] Pre-commit config is valid YAML and references correct entry points.

**Tests**: None (tooling config). Verify with `golangci-lint run` and `pre-commit run --all-files`.

---

## Epic 2 — Alpha Vantage API Client

> Build the HTTP client layer with rate limiting, retry, caching, and typed responses.

### Task 2.1 — Core HTTP client with rate limiting and retry

**Description**: Implement `internal/alphavantage/client.go` — the foundation for all API calls.

**Deliverables**:
- `internal/alphavantage/client.go`:
  - `Client` struct holding: base URL, API key, `*http.Client`, rate limiter, config.
  - `New(cfg ClientConfig) *Client` constructor.
  - `get(ctx context.Context, params map[string]string) ([]byte, error)` — internal method that:
    - Builds the query URL with `function=` and `apikey=` params.
    - Acquires a rate limit token (blocks if budget exhausted).
    - Executes GET with context timeout.
    - Retries on 5xx or timeout with exponential backoff (3 attempts, 1s base, 2x factor, ±200ms jitter).
    - Returns response body bytes or typed `APIError`.
  - `APIError` struct: `StatusCode int`, `Message string`, `Endpoint string`.
  - Rate limiter: token bucket allowing 70 requests/minute (configurable).
- `internal/alphavantage/client_test.go`.

**Acceptance Criteria**:
- [ ] Successful GET returns response bytes.
- [ ] 429 response triggers retry with backoff.
- [ ] 5xx response triggers retry with backoff.
- [ ] 4xx (non-429) response returns `APIError` immediately (no retry).
- [ ] Rate limiter blocks when budget exhausted and unblocks after refill.
- [ ] Context cancellation aborts request.
- [ ] API key is included in every request but never logged.
- [ ] All requests use HTTPS.

**Tests** (using `httptest.NewServer`):
```
TestClient_SuccessfulGet
TestClient_RetryOn5xx
TestClient_RetryOn429
TestClient_NoRetryOn4xx
TestClient_ContextCancellation
TestClient_RateLimiting
TestClient_MaxRetriesExhausted
TestClient_TimeoutRetry
TestClient_APIKeyInRequest
```

**Guidelines reference**: `CLAUDE.md` §4.3.

---

### Task 2.2 — Response models and JSON parsing

**Description**: Define typed Go structs for all Alpha Vantage API responses used by the app.

**Deliverables**:
- `internal/alphavantage/models.go`:
  - `GlobalQuote` — price, change, volume, etc.
  - `TimeSeriesDaily` — map of date → OHLCV.
  - `CryptoDaily` — map of date → OHLCV (with market-specific and USD prices).
  - `CryptoIntraday` — map of timestamp → OHLCV.
  - `RSIResponse` — map of date → RSI value.
  - `EMAResponse` — map of date → EMA value.
  - `NewsSentiment` — list of articles with scores.
  - `MacroDataPoint` — date + value (generic for GDP, CPI, etc.).
  - `MarketStatus` — list of markets with open/closed status.
  - Helper: `ParseFloat(s string) (float64, error)` — handles Alpha Vantage's string-encoded numbers.
  - Helper: `ParseDate(s string) (time.Time, error)` — handles `YYYY-MM-DD` format.
- `internal/alphavantage/models_test.go`.

**Acceptance Criteria**:
- [ ] All structs deserialize from real Alpha Vantage JSON fixtures.
- [ ] String → float64 parsing handles edge cases (empty string, "None", "-").
- [ ] Date parsing handles all AV date formats.
- [ ] No `map[string]interface{}` — everything is typed.

**Tests**:
```
TestParseGlobalQuote
TestParseCryptoDaily
TestParseRSIResponse
TestParseEMAResponse
TestParseNewsSentiment
TestParseMacroDataPoint
TestParseFloat_ValidNumber
TestParseFloat_EmptyString
TestParseFloat_None
TestParseDate_ValidDate
TestParseDate_InvalidFormat
```

**Test fixtures**: Store sample JSON responses in `internal/alphavantage/testdata/`.

---

### Task 2.3 — Time series endpoints

**Description**: Implement methods for fetching time series data (equities and crypto).

**Deliverables**:
- `internal/alphavantage/timeseries.go`:
  - `(c *Client) GetDailyTimeSeries(ctx, symbol string) (*TimeSeriesDaily, error)`
  - `(c *Client) GetCryptoDaily(ctx, symbol, market string) (*CryptoDaily, error)`
  - `(c *Client) GetCryptoIntraday(ctx, symbol, market, interval string) (*CryptoIntraday, error)`
  - `(c *Client) GetGlobalQuote(ctx, symbol string) (*GlobalQuote, error)`
  - `(c *Client) GetBulkQuotes(ctx, symbols []string) ([]GlobalQuote, error)` — uses `REALTIME_BULK_QUOTES`, batches symbols in groups of 100.
- `internal/alphavantage/timeseries_test.go`.

**Acceptance Criteria**:
- [ ] Each method calls the correct AV function with correct params.
- [ ] JSON response is parsed into typed structs.
- [ ] Empty or malformed response returns descriptive error.
- [ ] Bulk quotes handles > 100 symbols by batching.
- [ ] Context is propagated to underlying HTTP call.

**Tests**:
```
TestGetDailyTimeSeries_Success
TestGetDailyTimeSeries_InvalidSymbol
TestGetCryptoDaily_Success
TestGetCryptoDaily_WeekendDataIncluded
TestGetGlobalQuote_Success
TestGetBulkQuotes_Under100
TestGetBulkQuotes_Over100_Batches
TestGetBulkQuotes_PartialFailure
```

---

### Task 2.4 — Technical indicator endpoints

**Description**: Implement methods for fetching RSI and EMA from Alpha Vantage (equities path).

**Deliverables**:
- `internal/alphavantage/technicals.go`:
  - `(c *Client) GetRSI(ctx, symbol, interval string, period int) (*RSIResponse, error)`
  - `(c *Client) GetEMA(ctx, symbol, interval string, period int) (*EMAResponse, error)`
- `internal/alphavantage/technicals_test.go`.

**Acceptance Criteria**:
- [ ] RSI request includes correct `function`, `symbol`, `interval`, `time_period`, `series_type=close`.
- [ ] EMA request includes correct params.
- [ ] Response parsed into typed structs with float64 values.
- [ ] Handles AV error responses (e.g., invalid symbol, rate limit message in JSON body).

**Tests**:
```
TestGetRSI_Success
TestGetRSI_InvalidSymbol
TestGetRSI_ErrorInBody
TestGetEMA_Success
TestGetEMA_InvalidPeriod
```

---

### Task 2.5 — Macro economic endpoints

**Description**: Implement methods for all macroeconomic data endpoints.

**Deliverables**:
- `internal/alphavantage/macro.go`:
  - `(c *Client) GetRealGDP(ctx, interval string) ([]MacroDataPoint, error)` — `interval`: annual/quarterly.
  - `(c *Client) GetRealGDPPerCapita(ctx) ([]MacroDataPoint, error)`
  - `(c *Client) GetCPI(ctx, interval string) ([]MacroDataPoint, error)`
  - `(c *Client) GetInflation(ctx) ([]MacroDataPoint, error)`
  - `(c *Client) GetFedFundsRate(ctx, interval string) ([]MacroDataPoint, error)`
  - `(c *Client) GetTreasuryYield(ctx, interval, maturity string) ([]MacroDataPoint, error)`
  - `(c *Client) GetUnemployment(ctx) ([]MacroDataPoint, error)`
  - `(c *Client) GetNonfarmPayroll(ctx) ([]MacroDataPoint, error)`
- `internal/alphavantage/macro_test.go`.

**Acceptance Criteria**:
- [ ] Each method calls the correct AV function.
- [ ] Response parsed as `[]MacroDataPoint` sorted by date descending.
- [ ] Treasury yield supports maturity values: 2year, 5year, 10year, 30year.
- [ ] Empty data set returns empty slice, not nil.

**Tests**:
```
TestGetRealGDP_Quarterly
TestGetRealGDP_Annual
TestGetCPI_Monthly
TestGetInflation_Success
TestGetFedFundsRate_Daily
TestGetTreasuryYield_10Year
TestGetUnemployment_Success
TestGetNonfarmPayroll_Success
TestMacro_EmptyResponse
```

---

### Task 2.6 — News sentiment endpoint

**Description**: Implement the news sentiment fetch with filtering.

**Deliverables**:
- `internal/alphavantage/news.go`:
  - `(c *Client) GetNewsSentiment(ctx context.Context, opts NewsOpts) (*NewsSentiment, error)`
  - `NewsOpts` struct: `Tickers []string`, `Topics []string`, `Sort string`, `Limit int`.
- `internal/alphavantage/news_test.go`.

**Acceptance Criteria**:
- [ ] Ticker filter applied as comma-separated `tickers` param.
- [ ] Topic filter applied as comma-separated `topics` param.
- [ ] Each article parsed with: title, URL, source, published time, sentiment score, relevance score, ticker sentiments.
- [ ] Default limit: 50 articles.

**Tests**:
```
TestGetNewsSentiment_NoFilter
TestGetNewsSentiment_TickerFilter
TestGetNewsSentiment_TopicFilter
TestGetNewsSentiment_CombinedFilters
TestGetNewsSentiment_EmptyResult
TestGetNewsSentiment_ArticleParsing
```

---

## Epic 3 — Cache Layer

> In-memory TTL cache to minimize API calls and provide offline resilience.

### Task 3.1 — TTL cache implementation

**Description**: Build a generic, thread-safe, in-memory cache with per-key TTL.

**Deliverables**:
- `internal/cache/cache.go`:
  - `Store` struct using `sync.RWMutex` for thread safety.
  - `New() *Store` constructor.
  - `Get(key string) (interface{}, bool)` — returns value + whether it exists and is not expired.
  - `Set(key string, value interface{}, ttl time.Duration)` — stores with expiry timestamp.
  - `Delete(key string)`.
  - `Flush()` — clears all entries.
  - `Len() int` — count of non-expired entries.
  - Lazy expiration: expired entries are cleaned up on `Get` (returns miss) and periodically via background goroutine (every 60s).
  - `CacheKey(parts ...string) string` — helper to build consistent keys like `rsi:AAPL:14:daily`.
- `internal/cache/cache_test.go`.

**Acceptance Criteria**:
- [ ] Set + Get returns stored value before TTL.
- [ ] Get after TTL returns miss.
- [ ] Concurrent Set/Get from multiple goroutines is safe (no race detector failures).
- [ ] Flush clears all entries.
- [ ] Len only counts non-expired.
- [ ] CacheKey produces deterministic, collision-free keys.

**Tests** (run with `-race`):
```
TestCache_SetAndGet
TestCache_TTLExpiry
TestCache_Delete
TestCache_Flush
TestCache_Len_ExcludesExpired
TestCache_ConcurrentAccess
TestCache_LazyCleanup
TestCacheKey_Deterministic
TestCacheKey_DifferentInputs
```

**Guidelines reference**: `CLAUDE.md` §4.4.

---

## Epic 4 — Domain Layer: Indicators & Scoring

> Pure business logic for trend analysis and valuation. No Bubbletea, no HTTP.

### Task 4.1 — Indicator interface and types

**Description**: Define the interface contract and shared types for indicators.

**Deliverables**:
- `internal/domain/trend/indicators/indicator.go`:
  - `DataPoint` struct: `Date time.Time`, `Value float64`.
  - `OHLCV` struct: `Date time.Time`, `Open, High, Low, Close, Volume float64`.
  - `Indicator` interface:
    ```go
    type Indicator interface {
        Compute(ctx context.Context, symbol string, opts Options) ([]DataPoint, error)
    }
    ```
  - `Options` struct: `Period int`, `Interval string`, `SeriesType string`.
  - `AssetClass` type: `Equity` or `Crypto`.
  - `DetectAssetClass(symbol string) AssetClass` — heuristic: known crypto symbols → Crypto, else Equity.
- `internal/domain/trend/indicators/indicator_test.go`.

**Acceptance Criteria**:
- [ ] Interface compiles and is implementable.
- [ ] `DetectAssetClass` correctly classifies: AAPL→Equity, BTC→Crypto, ETH→Crypto, MSFT→Equity.
- [ ] Crypto list is configurable (passed in, not hardcoded).

**Tests**:
```
TestDetectAssetClass_Equity
TestDetectAssetClass_Crypto
TestDetectAssetClass_Unknown_DefaultsEquity
```

---

### Task 4.2 — Local RSI implementation (TradingView `ta.rsi()` parity)

**Description**: Implement RSI calculation in pure Go matching TradingView's `ta.rsi()` exactly.
TradingView's RSI uses **RMA (Wilder's smoothing)** internally — NOT standard EMA.

**Deliverables**:
- `internal/domain/trend/indicators/local_rsi.go`:
  - `LocalRSI` struct implementing `Indicator` interface.
  - Accepts `[]OHLCV` data (from crypto daily endpoint).
  - Uses close prices of **completed bars only** (discard any in-progress bar).
  - Algorithm (must match TradingView `ta.rsi()` exactly):
    1. Calculate price changes: `change[i] = close[i] - close[i-1]`.
    2. Separate gains (`max(change, 0)`) and losses (`abs(min(change, 0))`).
    3. Smooth with **RMA (alpha = 1/period)**, NOT EMA (alpha = 2/(period+1)):
       - Seed: SMA of first `period` gains/losses.
       - Subsequent: `avg = (prev_avg * (period-1) + current) / period`.
    4. `RS = avg_gain / avg_loss`. `RSI = 100 - (100 / (1 + RS))`.
  - Returns `[]DataPoint` with RSI values for each date (first `period` dates have no value).
  - Returns error if len(data) < period + 1.
  - Handles edge case: avg_loss = 0 → RSI = 100, avg_gain = 0 → RSI = 0.
- `internal/domain/trend/indicators/local_rsi_test.go`.

**Acceptance Criteria**:
- [ ] RSI values match TradingView chart output for the same data (use pre-computed test vectors validated against TradingView).
- [ ] Uses RMA smoothing (alpha = 1/period), NOT EMA smoothing (alpha = 2/(period+1)).
- [ ] Insufficient data returns descriptive error.
- [ ] All gains → RSI = 100.
- [ ] All losses → RSI = 0.
- [ ] Mixed data produces values between 0 and 100.
- [ ] Output length = input length - period.
- [ ] Only completed bars are used (last bar discarded if flagged as in-progress).

**Tests** (table-driven with pre-computed expected values validated against TradingView):
```
TestLocalRSI_MatchesTradingView       # Core: compare against known TV output
TestLocalRSI_UsesRMA_NotEMA           # Verify alpha = 1/period, not 2/(period+1)
TestLocalRSI_InsufficientData
TestLocalRSI_AllGains
TestLocalRSI_AllLosses
TestLocalRSI_FlatPrices
TestLocalRSI_SinglePeriod
TestLocalRSI_OutputLength
TestLocalRSI_BarCloseOnly             # Verify in-progress bar is excluded
```

**Guidelines reference**: `CLAUDE.md` §5.0, §5.2.

---

### Task 4.3 — Local EMA implementation (TradingView `ta.ema()` parity)

**Description**: Implement EMA calculation in pure Go matching TradingView's `ta.ema()` exactly.

**Deliverables**:
- `internal/domain/trend/indicators/local_ema.go`:
  - `LocalEMA` struct implementing `Indicator` interface.
  - Accepts `[]OHLCV` data.
  - Uses close prices of **completed bars only** (discard any in-progress bar).
  - Algorithm (must match TradingView `ta.ema()` exactly):
    1. Multiplier: `alpha = 2 / (period + 1)`.
    2. **Seed: EMA[0] = first close price** (NOT SMA of first `period` values —
       TradingView seeds with the first source value).
    3. `EMA[i] = alpha * close[i] + (1 - alpha) * EMA[i-1]`.
  - Returns `[]DataPoint` with EMA values.
  - Returns error if len(data) < 1 (at least one data point needed).
- `internal/domain/trend/indicators/local_ema_test.go`.

**Acceptance Criteria**:
- [ ] First EMA value equals the first close price (NOT SMA — TradingView behavior).
- [ ] Subsequent values follow the EMA formula exactly.
- [ ] EMA values match TradingView chart output for the same data.
- [ ] Insufficient data returns descriptive error.
- [ ] Period 1 → EMA equals close price at every bar.
- [ ] Only completed bars are used (last bar discarded if flagged as in-progress).

**Tests**:
```
TestLocalEMA_MatchesTradingView       # Core: compare against known TV output
TestLocalEMA_SeedWithFirstValue       # Verify seed is first src, NOT SMA
TestLocalEMA_InsufficientData
TestLocalEMA_Period1
TestLocalEMA_FlatPrices
TestLocalEMA_OutputLength
TestLocalEMA_BarCloseOnly             # Verify in-progress bar is excluded
```

**Guidelines reference**: `CLAUDE.md` §5.0, §5.3.

---

### Task 4.4 — Remote indicator wrapper (equities path)

**Description**: Implement the `Indicator` interface by wrapping Alpha Vantage's server-side endpoints.

**Deliverables**:
- `internal/domain/trend/indicators/remote.go`:
  - `RemoteRSI` struct implementing `Indicator`. Wraps `alphavantage.Client.GetRSI`.
  - `RemoteEMA` struct implementing `Indicator`. Wraps `alphavantage.Client.GetEMA`.
  - Both accept the AV client via constructor injection (interface, not concrete type).
  - Both convert AV response format to `[]DataPoint`.
- `internal/domain/trend/indicators/remote_test.go`.

**Acceptance Criteria**:
- [ ] Calls correct AV client method with correct params.
- [ ] Converts response to `[]DataPoint` sorted by date descending.
- [ ] AV client errors are propagated with context.
- [ ] Uses interface for AV client (testable with mocks).

**Tests** (using mock AV client):
```
TestRemoteRSI_Success
TestRemoteRSI_ClientError
TestRemoteEMA_Success
TestRemoteEMA_ClientError
TestRemoteRSI_ResponseConversion
```

---

### Task 4.5 — Trend engine and scoring

**Description**: Orchestrator that routes to the correct indicator path and computes composite trend signals.

**Deliverables**:
- `internal/domain/trend/engine.go`:
  - `Engine` struct holding: RSI indicator (remote + local), EMA indicator (remote + local), scoring config.
  - `New(remoteFetcher, localDataFetcher, scoringCfg) *Engine` constructor.
  - `Analyze(ctx, symbol string, assetClass AssetClass) (*TrendResult, error)`:
    1. Based on `assetClass`, choose remote or local indicators.
    2. Fetch RSI (period 14), EMA(9), EMA(21).
    3. **Use only the last closed bar's values** — discard any in-progress bar data.
    4. Pass latest closed-bar values to scoring.
    5. Return `TrendResult`.
  - `TrendResult` struct: `Symbol`, `RSI float64`, `EMAFast float64`, `EMASlow float64`, `Signal TrendSignal`, `Valuation string`.
  - `TrendSignal` type: `Bullish`, `Bearish`, `Neutral`.
- `internal/domain/trend/scoring.go`:
  - `Score(rsi, emaFast, emaSlow float64, cfg ScoringConfig) TrendSignal` — pure function.
  - `ScoringConfig` struct with threshold fields from `CLAUDE.md` §5.4.
- `internal/domain/trend/engine_test.go` and `internal/domain/trend/scoring_test.go`.

**Acceptance Criteria**:
- [ ] Equity symbol routes to remote indicators.
- [ ] Crypto symbol routes to local indicators.
- [ ] Scoring matches the rules in `CLAUDE.md` §5.4 exactly.
- [ ] All three signal types are reachable.
- [ ] Engine propagates indicator errors without panicking.

**Tests**:
```
TestScore_Bullish
TestScore_Bearish_LowRSI
TestScore_Bearish_HighRSI
TestScore_Neutral
TestScore_BoundaryValues
TestEngine_EquityRoutesToRemote
TestEngine_CryptoRoutesToLocal
TestEngine_IndicatorError_Propagated
TestEngine_FullAnalysis_Equity
TestEngine_FullAnalysis_Crypto
TestEngine_UsesClosedBarOnly          # Verify in-progress bar data is excluded
```

**Guidelines reference**: `CLAUDE.md` §5.0, §5.4.

---

### Task 4.6 — RSI valuation scoring

**Description**: Implement the RSI-based valuation labels. Uses RSI from the last closed bar only.

**Deliverables**:
- `internal/domain/valuation/rsi.go`:
  - `Valuate(rsi float64, cfg ValuationConfig) string` — pure function.
  - `ValuationConfig` struct with thresholds from `CLAUDE.md` §5.5.
  - Returns one of: "Oversold", "Undervalued", "Fair value", "Overvalued", "Overbought".
- `internal/domain/valuation/rsi_test.go`.

**Acceptance Criteria**:
- [ ] Each RSI range maps to the correct label per `CLAUDE.md` §5.5.
- [ ] Boundary values (exactly 30, 45, 55, 70) are handled consistently.
- [ ] RSI = 0 → "Oversold", RSI = 100 → "Overbought".

**Tests** (table-driven, exhaustive boundaries):
```
TestValuate_Oversold
TestValuate_Undervalued
TestValuate_FairValue
TestValuate_Overvalued
TestValuate_Overbought
TestValuate_BoundaryAt30
TestValuate_BoundaryAt45
TestValuate_BoundaryAt55
TestValuate_BoundaryAt70
TestValuate_ExtremeValues
```

**Guidelines reference**: `CLAUDE.md` §5.5.

---

## Epic 5 — TUI Presentation Layer

> Bubbletea models, views, styling, and navigation.

### Task 5.1 — Theme and shared components

**Description**: Define the visual theme (Lipgloss styles) and shared UI components.

**Deliverables**:
- `internal/tui/theme.go`:
  - Color palette constants for default theme, minimal theme, colorblind theme.
  - Lipgloss styles for: tab bar, active tab, table headers, table rows, signal colors
    (bullish green, bearish red, neutral yellow), box borders, help text, error text, loading spinner.
  - `Theme` struct with all styles, switchable by config.
- `internal/tui/components/spinner.go`: Reusable loading spinner component.
- `internal/tui/components/table.go`: Reusable table component with column alignment.
- `internal/tui/components/help.go`: Help overlay component showing key bindings.

**Acceptance Criteria**:
- [ ] Three themes defined and switchable.
- [ ] All components render without panics for empty data.
- [ ] Table handles variable-width columns and truncates overflow.
- [ ] Spinner animates through frames.

**Tests**:
```
TestTheme_DefaultColors
TestTheme_ColorblindColors
TestTable_EmptyData
TestTable_ColumnAlignment
TestTable_Truncation
TestHelp_RendersBindings
```

---

### Task 5.2 — Root app model and tab router

**Description**: Build the root Bubbletea model that manages tab navigation between views.

**Deliverables**:
- `internal/tui/app.go`:
  - `App` model implementing `tea.Model`.
  - Holds child models: trend, quote, macro, news.
  - `Init()` → triggers data load for the default (first) tab.
  - `Update()`:
    - `1`/`2`/`3`/`4` keys → switch active tab.
    - `Tab` → cycle to next tab.
    - `q` / `Ctrl+C` → quit.
    - `?` → toggle help overlay.
    - `r` → send refresh command to active child.
    - All other keys → delegate to active child model.
  - `View()` → renders tab bar + active child view + status bar.
  - Status bar: shows last update time, data source indicator, error count.
- `cmd/finterm/main.go`: wire up config → client → cache → domain → TUI → `tea.NewProgram`.

**Acceptance Criteria**:
- [ ] Tab switching works with number keys and Tab.
- [ ] `q` quits cleanly.
- [ ] `?` toggles help overlay.
- [ ] `r` triggers refresh on active view.
- [ ] Child model receives delegated messages.
- [ ] Status bar shows meaningful info.

**Tests**:
```
TestApp_TabSwitching
TestApp_QuitKey
TestApp_HelpToggle
TestApp_RefreshDelegation
TestApp_DefaultTab
```

---

### Task 5.3 — Trend following view

**Description**: Build the trend view Bubbletea model showing watchlist analysis.

**Deliverables**:
- `internal/tui/trend/model.go`:
  - `Model` implementing `tea.Model`.
  - States: loading, loaded, error.
  - `Init()` → dispatches concurrent fetch commands for all watchlist tickers.
  - `Update()`:
    - Handles `TrendDataMsg` (single ticker result) → updates table row.
    - Handles `TrendErrorMsg` → marks ticker as errored.
    - Arrow keys navigate rows.
    - `r` → re-fetches all tickers.
  - Holds `[]TrendResult` as table data.
- `internal/tui/trend/view.go`:
  - Renders the table per PRD §3.2 wireframe.
  - Color-codes signals: green/red/yellow.
  - Shows "Loading..." for in-flight tickers.
  - Shows "Error" for failed tickers.

**Acceptance Criteria**:
- [ ] Loading state shows spinner for each ticker.
- [ ] Results populate incrementally (first ticker appears before last is done).
- [ ] Error in one ticker does not block others.
- [ ] Signals are color-coded.
- [ ] Arrow navigation highlights active row.

**Tests**:
```
TestTrendModel_Init_DispatchesCommands
TestTrendModel_Update_DataMsg
TestTrendModel_Update_ErrorMsg
TestTrendModel_Update_ArrowNavigation
TestTrendModel_View_LoadingState
TestTrendModel_View_LoadedState
TestTrendModel_View_MixedState
```

---

### Task 5.4 — Quote lookup view

**Description**: Build the quote view with text input and result display.

**Deliverables**:
- `internal/tui/quote/model.go`:
  - `Model` implementing `tea.Model`.
  - Text input field (using `bubbles/textinput`).
  - States: idle (input focused), loading, loaded, error.
  - `Update()`:
    - Enter → submit ticker → dispatch fetch command.
    - Up/Down in idle → cycle through last 10 lookups.
    - Handles `QuoteResultMsg`, `QuoteErrorMsg`.
- `internal/tui/quote/view.go`:
  - Renders per PRD §3.3 wireframe.
  - Shows inline RSI, EMA, trend, and valuation alongside the quote.

**Acceptance Criteria**:
- [ ] Text input accepts and submits ticker.
- [ ] Loading state shows spinner.
- [ ] Result displays all fields from `GlobalQuote` + indicators.
- [ ] Invalid ticker shows error message.
- [ ] Lookup history navigable with arrows.

**Tests**:
```
TestQuoteModel_SubmitTicker
TestQuoteModel_LoadingState
TestQuoteModel_ResultDisplay
TestQuoteModel_ErrorState
TestQuoteModel_LookupHistory
TestQuoteModel_InputValidation
```

---

### Task 5.5 — Macro dashboard view

**Description**: Build the macro dashboard with paneled layout.

**Deliverables**:
- `internal/tui/macro/model.go`:
  - `Model` implementing `tea.Model`.
  - Fetches all macro endpoints on init.
  - Holds structured data for each panel: GDP, inflation, employment, rates, yields.
- `internal/tui/macro/view.go`:
  - Renders per PRD §3.4 wireframe.
  - Box-drawing panels using Lipgloss borders.
  - Responsive: adapts panel layout to terminal width.
  - Shows last update timestamp and TTL.

**Acceptance Criteria**:
- [ ] All five panels render with data.
- [ ] Loading state shows per-panel spinners.
- [ ] Stale data shows warning indicator.
- [ ] Layout adapts to narrow terminals (stacks vertically).

**Tests**:
```
TestMacroModel_Init_FetchesAll
TestMacroModel_Update_GDPData
TestMacroModel_Update_PartialLoad
TestMacroModel_View_AllPanels
TestMacroModel_View_NarrowTerminal
TestMacroModel_View_StaleIndicator
```

---

### Task 5.6 — News feed view

**Description**: Build the scrollable news feed with filtering and sorting.

**Deliverables**:
- `internal/tui/news/model.go`:
  - `Model` implementing `tea.Model`.
  - Scrollable list of articles.
  - Filter state: all, equities, crypto, macro.
  - Sort state: newest, score.
  - `Update()`:
    - `j`/`k` or arrows → navigate articles.
    - `f` → cycle filter.
    - `s` → cycle sort.
    - `Enter` → open article URL (or copy to clipboard).
    - `r` → refresh.
- `internal/tui/news/view.go`:
  - Renders per PRD §3.5 wireframe.
  - Sentiment-colored indicators per article.
  - Shows filter and sort state in header.

**Acceptance Criteria**:
- [ ] Articles render with sentiment scores, tickers, headlines, sources, timestamps.
- [ ] Filter toggles work and re-render the list.
- [ ] Sort toggles work.
- [ ] Vim-style `j`/`k` navigation works.
- [ ] Enter on article opens URL or copies to clipboard.
- [ ] Empty result set shows "No articles found" message.

**Tests**:
```
TestNewsModel_Init_FetchesArticles
TestNewsModel_Update_Navigation
TestNewsModel_Update_FilterToggle
TestNewsModel_Update_SortToggle
TestNewsModel_View_ArticleRendering
TestNewsModel_View_SentimentColors
TestNewsModel_View_EmptyState
TestNewsModel_Update_OpenArticle
```

---

## Epic 6 — Integration & Polish

> Wire everything together, end-to-end testing, error handling, and UX refinements.

### Task 6.1 — Main entry point wiring

**Description**: Complete `cmd/finterm/main.go` with full dependency injection.

**Deliverables**:
- `cmd/finterm/main.go`:
  - Loads config via `config.Load`.
  - Creates `alphavantage.Client`.
  - Creates `cache.Store`.
  - Creates domain engines (trend, valuation) with injected dependencies.
  - Creates TUI app model with all child models wired.
  - Starts `tea.NewProgram` with alt screen and mouse support.
  - Graceful shutdown on SIGINT/SIGTERM.

**Acceptance Criteria**:
- [ ] `go run ./cmd/finterm/` starts the TUI.
- [ ] Missing config → clear error message.
- [ ] Missing API key → clear error message.
- [ ] Ctrl+C exits gracefully.
- [ ] `make build` produces a working binary.

**Tests**: Manual verification (integration). Ensure `go build ./cmd/finterm/` compiles.

---

### Task 6.2 — Graceful degradation and error states

**Description**: Ensure the app handles failures gracefully at every level.

**Deliverables**:
- API unreachable → show cached data with "offline" indicator.
- Rate limited → queue and retry, show "rate limited" in status bar.
- Individual ticker failure → show "Error" in that row, don't block others.
- Invalid config → startup error with actionable message.
- Terminal resize → reflow layout without crash.

**Acceptance Criteria**:
- [ ] App never panics on API errors.
- [ ] Stale cache data is shown with timestamp when API fails.
- [ ] Status bar reflects current health: "online", "rate limited", "offline".
- [ ] Terminal resize triggers clean re-render.

**Tests**:
```
TestApp_APIUnavailable_ShowsCachedData
TestApp_RateLimited_StatusBar
TestApp_TickerError_IsolatedFailure
TestApp_TerminalResize
```

---

### Task 6.3 — Help system and keybinding documentation

**Description**: Implement the `?` help overlay with context-sensitive bindings.

**Deliverables**:
- `internal/tui/components/help.go`:
  - `HelpOverlay` model that renders a centered panel.
  - Shows global bindings (always) + view-specific bindings (based on active tab).
  - Dismissible with `?` or `Esc`.
- Each view provides its own `KeyBindings() []KeyBinding` method.

**Acceptance Criteria**:
- [ ] Help overlay renders centered on screen.
- [ ] Shows global + view-specific bindings.
- [ ] `?` and `Esc` dismiss it.
- [ ] Overlay blocks input to underlying view while visible.

**Tests**:
```
TestHelpOverlay_Render
TestHelpOverlay_Dismiss
TestHelpOverlay_ContextSensitive
```

---

### Task 6.4 — Final review and documentation

**Description**: Final pass over code quality, documentation, and README.

**Deliverables**:
- `README.md` with: project description, features, installation, configuration, usage, screenshots (placeholder), development setup, architecture overview.
- GoDoc comments on all exported types and functions.
- `go vet ./...` clean.
- `golangci-lint run` clean.
- `go test -race ./...` passes.
- Test coverage report: `go test -coverprofile=coverage.out ./internal/...`.

**Acceptance Criteria**:
- [ ] README is complete and actionable.
- [ ] All exported symbols have GoDoc comments.
- [ ] Zero lint warnings.
- [ ] Zero race conditions.
- [ ] Coverage > 80% on `domain/` and `alphavantage/`.
- [ ] `make build` produces a clean binary.

**Tests**: All existing tests pass. Coverage target met.

---

## Task Dependency Graph

```
Epic 1 (Scaffolding)
  1.1 ──► 1.2 ──► 1.3
             │
             ▼
Epic 2 (API Client)
  2.1 ──► 2.2 ──► 2.3 ──► 2.4
             │       │       │
             │       ▼       ▼
             │     2.5     2.6
             │
             ▼
Epic 3 (Cache)
  3.1 (parallel with Epic 2 after 2.1)
             │
             ▼
Epic 4 (Domain)
  4.1 ──► 4.2 ──► 4.4 ──► 4.5 ──► 4.6
     │       │
     │       ▼
     └──► 4.3
             │
             ▼
Epic 5 (TUI)
  5.1 ──► 5.2 ──► 5.3
             │       │
             ├──► 5.4
             ├──► 5.5
             └──► 5.6
                    │
                    ▼
Epic 6 (Integration)
  6.1 ──► 6.2 ──► 6.3 ──► 6.4
```

**Critical path**: 1.1 → 1.2 → 2.1 → 2.2 → 2.3/2.4 → 4.1 → 4.2/4.3 → 4.5 → 5.2 → 6.1 → 6.4

**Parallelizable**:
- Task 3.1 (cache) can start after 1.2.
- Tasks 2.5 and 2.6 can run parallel to 2.3/2.4.
- Tasks 4.2 and 4.3 (local RSI/EMA) are independent and parallelizable.
- Tasks 5.3, 5.4, 5.5, 5.6 can run in parallel after 5.2.