# CLAUDE.md — Finterm Project Guidelines

> **This file is the single source of truth for project conventions.**
> It MUST be referenced in pre-commit checks and by any AI-assisted development tool.
> Every contributor (human or AI) must read and follow these guidelines before writing code.

---

## Project Identity

**Finterm** is a terminal-based financial analysis tool built with Go and Bubbletea.
It provides trend-following analysis, real-time quotes, macroeconomic dashboards,
and a sentiment-scored news feed — all powered by the Alpha Vantage API.

---

## 1. Language & Runtime

- **Go 1.26+** (use the latest stable release).
- Target platforms: Linux, macOS, Windows (terminal-based, no GUI dependencies).
- Module path: `github.com/shinsekai/finterm`.

---

## 2. Project Structure

```
finterm/
├── CLAUDE.md                          # This file — project guidelines
├── PRD.md                             # Product Requirements Document
├── EPICS.md                           # Epics and tasks
├── cmd/
│   └── finterm/
│       └── main.go                    # Entry point — config load, DI, app bootstrap
├── internal/
│   ├── tui/                           # Presentation layer (Bubbletea)
│   │   ├── app.go                     # Root model, tab router, keyboard nav
│   │   ├── theme.go                   # Colors, styles, Lipgloss definitions
│   │   ├── components/                # Shared UI components (spinner, table, etc.)
│   │   ├── trend/                     # Trend following view
│   │   │   ├── model.go
│   │   │   └── view.go
│   │   ├── quote/                     # Single ticker quote view
│   │   │   ├── model.go
│   │   │   └── view.go
│   │   ├── macro/                     # Macroeconomic dashboard view
│   │   │   ├── model.go
│   │   │   └── view.go
│   │   └── news/                      # News feed view
│   │       ├── model.go
│   │       └── view.go
│   ├── domain/                        # Business logic — NO Bubbletea imports
│   │   ├── trend/
│   │   │   ├── engine.go              # Orchestrator — picks indicator strategy
│   │   │   ├── scoring.go             # Composite signal scoring
│   │   │   └── indicators/
│   │   │       ├── indicator.go       # Indicator interface
│   │   │       ├── remote.go          # Alpha Vantage server-side (equities)
│   │   │       ├── local_rsi.go       # Pure Go RSI (crypto / fallback)
│   │   │       ├── local_ema.go       # Pure Go EMA (crypto / fallback)
│   │   │       ├── local_rsi_test.go
│   │   │       └── local_ema_test.go
│   │   └── valuation/
│   │       ├── rsi.go                 # RSI-based valuation scoring
│   │       └── rsi_test.go
│   ├── alphavantage/                  # API client layer
│   │   ├── client.go                  # HTTP client, rate limiter, retry
│   │   ├── client_test.go
│   │   ├── timeseries.go              # Time series + crypto endpoints
│   │   ├── timeseries_test.go
│   │   ├── technicals.go              # Technical indicator endpoints
│   │   ├── technicals_test.go
│   │   ├── macro.go                   # Economic data endpoints
│   │   ├── macro_test.go
│   │   ├── news.go                    # News sentiment endpoint
│   │   ├── news_test.go
│   │   └── models.go                  # Shared response types
│   ├── cache/                         # In-memory TTL cache
│   │   ├── cache.go
│   │   └── cache_test.go
│   └── config/                        # Configuration loading
│       ├── config.go
│       └── config_test.go
├── config.example.yaml                # Example configuration
├── go.mod
├── go.sum
├── Makefile
└── .pre-commit-config.yaml
```

### Rules

- Everything under `internal/` is unexported to external consumers.
- `internal/domain/` MUST NOT import `internal/tui/` or any Bubbletea package.
- `internal/tui/` depends on `internal/domain/` and `internal/alphavantage/`, never the reverse.
- `internal/alphavantage/` MUST NOT import `internal/domain/` or `internal/tui/`.
- `internal/cache/` is a generic utility — no business logic, no domain types.
- `cmd/finterm/main.go` wires everything together (dependency injection).

---

## 3. Coding Standards

### 3.1 Formatting & Linting

- **gofmt**: All code must be formatted with `gofmt`. No exceptions.
- **golangci-lint**: Run with the project's `.golangci.yml` config. The following linters are enabled:
  `govet`, `errcheck`, `staticcheck`, `unused`, `gosimple`, `ineffassign`,
  `misspell`, `gocyclo` (max complexity: 15), `gocritic`, `revive`.
- **No `//nolint` directives** without an accompanying comment explaining why.

### 3.2 Naming Conventions

| Element          | Convention                       | Example                          |
|------------------|----------------------------------|----------------------------------|
| Packages         | Single lowercase word            | `trend`, `cache`, `news`         |
| Files            | `snake_case.go`                  | `local_rsi.go`, `client_test.go` |
| Exported types   | `PascalCase`                     | `TrendEngine`, `CacheStore`      |
| Unexported types | `camelCase`                      | `rsiCalculator`, `httpDoer`      |
| Interfaces       | Descriptive, not `-er` suffix    | `Indicator`, `DataFetcher`       |
| Constants        | `PascalCase` if exported         | `DefaultTTL`, `MaxRetries`       |
| Test functions   | `Test<Function>_<Scenario>`      | `TestRSI_InsufficientData`       |
| Test files       | Same package, `_test.go` suffix  | `cache_test.go`                  |

### 3.3 Error Handling

- **Always handle errors explicitly.** No blank `_` for error returns unless justified with a comment.
- **Wrap errors with context** using `fmt.Errorf("doing X: %w", err)`.
- **Use sentinel errors** for expected failure modes: `var ErrRateLimited = errors.New("rate limited")`.
- **Never panic** in library code. Panics are acceptable only in `main.go` for unrecoverable startup failures.
- **API errors** must be typed: define `APIError` struct with status code, message, and endpoint.

### 3.4 Concurrency

- Use **channels** for Bubbletea `tea.Cmd` / `tea.Msg` communication.
- Use `sync.Mutex` or `sync.RWMutex` only for shared state in the cache layer.
- **Never** use `sync.WaitGroup` inside Bubbletea models — use `tea.Cmd` for async work.
- Context propagation: every API call and long-running operation accepts `context.Context`.

### 3.5 Dependencies

Approved dependencies:

| Package                              | Purpose                        |
|--------------------------------------|--------------------------------|
| `github.com/charmbracelet/bubbletea` | TUI framework                  |
| `github.com/charmbracelet/lipgloss`  | Terminal styling                |
| `github.com/charmbracelet/bubbles`   | Pre-built Bubbletea components |
| `gopkg.in/yaml.v3`                   | Config file parsing            |
| `github.com/stretchr/testify`        | Test assertions (test only)    |

- **No other dependencies** without explicit justification documented in the PR.
- **Standard library first**: prefer `net/http`, `encoding/json`, `math`, `sync`, `time`, etc.

---

## 4. Architecture Rules

### 4.1 Layer Boundaries

```
┌──────────────────────────────────────────────────┐
│  Presentation (tui/)                             │
│  - Bubbletea models, views, components           │
│  - Lipgloss styling                              │
│  - Depends on: domain/, alphavantage/, cache/    │
├──────────────────────────────────────────────────┤
│  Domain (domain/)                                │
│  - Pure business logic                           │
│  - Indicator computation, scoring, valuation     │
│  - Depends on: NOTHING (interfaces only)         │
├──────────────────────────────────────────────────┤
│  Data (alphavantage/, cache/)                    │
│  - HTTP client, rate limiting, caching           │
│  - Depends on: NOTHING                           │
├──────────────────────────────────────────────────┤
│  Config (config/)                                │
│  - YAML parsing, validation, defaults            │
│  - Depends on: NOTHING                           │
└──────────────────────────────────────────────────┘
```

- Domain layer communicates with the data layer through **interfaces**, not concrete types.
- The `cmd/finterm/main.go` entry point performs all wiring (constructor injection).

### 4.2 Bubbletea Patterns

- Each view is its own `Model` implementing `tea.Model` (`Init`, `Update`, `View`).
- The root `App` model holds a slice of child models and routes messages by active tab.
- **Async data fetching** is done via `tea.Cmd` returning a `tea.Msg`. Never block in `Update`.
- Use custom message types per view: `TrendDataMsg`, `QuoteResultMsg`, etc.
- **Loading states**: every view must handle loading, success, and error states explicitly.

### 4.3 API Client Patterns

- A single `*alphavantage.Client` instance is created at startup and shared.
- **Rate limiting**: built into the client using a token bucket (`time.Ticker` or `golang.org/x/time/rate`).
  Alpha Vantage premium allows 75 requests/minute — enforce 70/min with headroom.
- **Retry with exponential backoff**: 3 attempts max, base delay 1s, factor 2x, jitter ±200ms.
- **Timeouts**: 10s per request. Use `context.WithTimeout`.
- **Response parsing**: decode JSON into typed structs, never `map[string]interface{}`.

### 4.4 Cache Strategy

| Data Type             | TTL        | Rationale                               |
|-----------------------|------------|-----------------------------------------|
| Intraday quotes       | 60s        | Balance freshness with rate limits      |
| Daily time series     | 1h         | Updates once per day after market close |
| Technical indicators  | 1h         | Derived from daily data                 |
| Macro data (GDP, CPI) | 6h         | Updated monthly/quarterly               |
| News sentiment        | 5min       | Relatively fast-changing                |
| Crypto OHLCV          | 5min       | 24/7 market, more volatile              |

---

## 5. Indicator Logic

> **Reference implementation: TradingView Pine Script v6.**
> All local indicator computations MUST match TradingView's built-in `ta.rsi()` and `ta.ema()`
> output exactly. Use TradingView charts as the ground truth for validation.

### 5.0 Bar Close Only Rule

**All trend and valuation computations use completed (closed) bars only.**

- Indicators are computed on the **previous bar's close**, never on the current in-progress bar.
- This eliminates intra-bar repainting: a signal that appeared on a closed bar will never change.
- For daily timeframe: use yesterday's close. For weekly: use last week's close. And so on.
- The current bar's real-time price is displayed in the Quote view for informational purposes only
  but is **never** fed into RSI, EMA, trend scoring, or valuation logic.
- When fetching data from Alpha Vantage, discard the most recent data point if it represents
  an incomplete/in-progress bar (e.g., today's intraday data before market close on daily timeframe).

### 5.1 Dual-Path Strategy

| Asset Class | RSI Source            | EMA Source            |
|-------------|----------------------|----------------------|
| **Equities**| Alpha Vantage `RSI`  | Alpha Vantage `EMA`  |
| **Crypto**  | Local computation    | Local computation    |

- The trend engine detects asset class from the ticker and routes accordingly.
- Local implementations serve as fallback for equities if rate-limited.
- Both paths produce identical mathematical results for the same input data.

### 5.2 RSI Implementation (Local) — TradingView `ta.rsi()`

TradingView's RSI uses **RMA (Relative Moving Average)** for smoothing, also known as
Wilder's Smoothing or SMMA. This is NOT the same as EMA — the alpha differs.

**Step-by-step algorithm (must match TradingView exactly):**

```
Input:  closes[]  — array of close prices, oldest first
        period    — lookback period (default: 14)

1. Calculate price changes:
   change[i] = closes[i] - closes[i-1]    (for i = 1 to len-1)

2. Separate gains and losses:
   gain[i] = max(change[i], 0)
   loss[i] = abs(min(change[i], 0))

3. Smooth with RMA (Wilder's smoothing):
   alpha = 1 / period

   Seed (first value, at index = period):
     avg_gain = SMA(gain[1..period])      — simple average of first `period` gains
     avg_loss = SMA(loss[1..period])       — simple average of first `period` losses

   Subsequent values (i > period):
     avg_gain = alpha * gain[i] + (1 - alpha) * prev_avg_gain
     avg_loss = alpha * loss[i] + (1 - alpha) * prev_avg_loss

   Equivalent to:
     avg_gain = (prev_avg_gain * (period - 1) + gain[i]) / period
     avg_loss = (prev_avg_loss * (period - 1) + loss[i]) / period

4. Compute RSI:
   RS  = avg_gain / avg_loss
   RSI = 100 - (100 / (1 + RS))

   Edge cases:
     avg_loss == 0  → RSI = 100
     avg_gain == 0  → RSI = 0
```

**Critical difference from EMA-based RSI:**
- RMA alpha = `1 / period` (for period=14: alpha = 0.0714)
- EMA alpha = `2 / (period + 1)` (for period=14: alpha = 0.1333)
- Using EMA alpha instead of RMA alpha will produce DIFFERENT values from TradingView.

- Default period: **14**.
- Return error if fewer than `period + 1` data points are provided.
- All computations use **close prices of completed bars only** (see §5.0).

### 5.3 EMA Implementation (Local) — TradingView `ta.ema()`

TradingView's EMA uses the standard exponential smoothing formula:

```
Input:  closes[]  — array of close prices, oldest first
        period    — lookback period

alpha = 2 / (period + 1)

Seed (first value):
  EMA[0] = closes[0]                    — first source value, NOT SMA

Subsequent values:
  EMA[i] = alpha * closes[i] + (1 - alpha) * EMA[i-1]
```

**Critical: TradingView seeds EMA with the first source value, not with SMA.**
This means the very first EMA value equals the first close price. Some implementations
seed with SMA of the first `period` values — that will NOT match TradingView output.

- Default periods: **10** (fast) and **20** (slow).
- Return error if fewer than `period` data points are provided.
- All computations use **close prices of completed bars only** (see §5.0).

### 5.4 Trend Scoring

A ticker is scored using the EMA crossover.
**All inputs are from the last closed bar** — never the current in-progress bar.

| Signal     | Condition                                      |
|------------|-------------------------------------------------|
| **Bullish**| EMA(10) > EMA(20)                               |
| **Bearish**| EMA(10) < EMA(20)                               |

These thresholds are configurable via `config.yaml`.

> **Note**: RSI is NOT used in trend scoring. RSI is used exclusively for valuation (§5.5).

### 5.5 Valuation (RSI-Based)

Uses the RSI value from the **last closed bar** only.

| RSI Range | Label         |
|-----------|---------------|
| < 30      | Oversold      |
| 30–45     | Undervalued   |
| 45–55     | Fair value    |
| 55–70     | Overvalued    |
| > 70      | Overbought    |

---

## 6. Security

- **API key** is loaded from environment variable `FINTERM_AV_API_KEY` or from `config.yaml`.
  Environment variable takes precedence.
- **Never log the API key.** Redact in any debug output.
- **No secrets in source control.** The `config.yaml` with a real key must be in `.gitignore`.
- **Input validation**: all user-entered tickers must be sanitized (alphanumeric + dots + dashes only,
  max 10 characters) before being passed to the API client.
- **TLS only**: the Alpha Vantage client must enforce HTTPS. No HTTP fallback.
- **Dependency auditing**: run `govulncheck ./...` in CI.

---

## 7. Testing

### 7.1 Requirements

- **Every function in `domain/` must have unit tests.** No exceptions.
- **Every public method in `alphavantage/` must have unit tests** using HTTP mocks.
- **Cache must have tests** covering TTL expiry, concurrent access, and eviction.
- **Table-driven tests** are the default pattern. Use `t.Run` for subtests.
- **No real API calls in tests.** Use `httptest.NewServer` or interface mocks.

### 7.2 Test Naming

```go
func TestRSI_ValidData(t *testing.T) { ... }
func TestRSI_InsufficientData(t *testing.T) { ... }
func TestRSI_AllGains(t *testing.T) { ... }
func TestEMA_SeedWithFirstValue(t *testing.T) { ... }
func TestCache_TTLExpiry(t *testing.T) { ... }
func TestClient_RateLimitRetry(t *testing.T) { ... }
```

### 7.3 Coverage

- Target: **80%** coverage on `domain/` and `alphavantage/`.
- Run: `go test -coverprofile=coverage.out ./internal/...`
- View: `go tool cover -func=coverage.out`

### 7.4 Test Data

- Store fixture JSON files under `internal/alphavantage/testdata/`.
- Name fixtures after the endpoint: `rsi_btc_daily.json`, `global_quote_aapl.json`.

---

## 8. Configuration

### 8.1 File Format

```yaml
# config.yaml
api:
  key: ""                              # Or use FINTERM_AV_API_KEY env var
  base_url: "https://www.alphavantage.co/query"
  rate_limit: 70                       # Requests per minute
  timeout: 10s
  max_retries: 3

watchlist:
  equities:
    - AAPL
    - MSFT
    - GOOGL
    - AMZN
    - NVDA
  crypto:
    - BTC
    - ETH
    - SOL

trend:
  rsi_period: 14
  ema_fast: 10
  ema_slow: 20

valuation:
  rsi_period: 14
  oversold: 30
  undervalued: 45
  fair_low: 45
  fair_high: 55
  overvalued: 70
  overbought: 70

cache:
  intraday_ttl: 60s
  daily_ttl: 1h
  macro_ttl: 6h
  news_ttl: 5m
  crypto_ttl: 5m

theme:
  style: "default"                     # default | minimal | colorblind
```

### 8.2 Validation

- Config is validated at startup. Missing API key → fatal error with clear message.
- Invalid ticker formats → warning logged, ticker skipped.
- TTL values → must parse as `time.Duration`.

---

## 9. Pre-Commit Checks

The following checks run on every commit via `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: claude-guidelines-check
        name: CLAUDE.md compliance
        entry: bash -c 'echo "Ensure code follows CLAUDE.md guidelines" && exit 0'
        language: system
        always_run: true
        pass_filenames: false

      - id: go-fmt
        name: gofmt
        entry: gofmt -l -d .
        language: system
        types: [go]

      - id: go-vet
        name: go vet
        entry: go vet ./...
        language: system
        types: [go]
        pass_filenames: false

      - id: go-test
        name: go test
        entry: go test ./...
        language: system
        types: [go]
        pass_filenames: false

      - id: go-lint
        name: golangci-lint
        entry: golangci-lint run
        language: system
        types: [go]
        pass_filenames: false

      - id: go-vulncheck
        name: govulncheck
        entry: govulncheck ./...
        language: system
        types: [go]
        pass_filenames: false

      - id: no-secrets
        name: No secrets in code
        entry: bash -c 'grep -rn "apikey\|api_key\|secret" --include="*.go" | grep -v "_test.go" | grep -v "os.Getenv" | grep -v "config\." | grep -v "// " && echo "FAIL: Possible secret in source" && exit 1 || exit 0'
        language: system
        pass_filenames: false
```

---

## 10. Git Conventions

- **Branch naming**: `epic/<name>`, `task/<epic>-<number>`, `fix/<description>`.
- **Commit messages**: imperative mood, max 72 chars subject line.
  Format: `<scope>: <description>` — e.g., `trend: add local RSI computation`.
  Scopes: `trend`, `quote`, `macro`, `news`, `client`, `cache`, `config`, `tui`, `ci`, `docs`.
- **No force pushes** on `main`.
- **Squash merge** feature branches into `main`.

---

## 11. AI Development Rules (Claude Code)

When working on this project with Claude Code:

1. **Read this file first.** Always. Before writing any code.
2. **Follow the structure exactly.** Do not create files outside the defined structure without updating this file.
3. **Write tests alongside code.** Every task deliverable includes passing tests.
4. **Respect layer boundaries.** If a domain function needs data, use an interface — never import the client directly.
5. **Use table-driven tests.** No single-assertion test functions.
6. **Run checks before declaring done:** `go fmt`, `go vet`, `go test ./...`.
7. **Reference EPICS.md** for task definitions. Each task has acceptance criteria and test requirements.
8. **Do not modify config.yaml with real API keys.** Use `config.example.yaml` as the template.
9. **When in doubt, keep it simple.** Prefer standard library over third-party packages.
10. **Document exported types and functions** with GoDoc comments.