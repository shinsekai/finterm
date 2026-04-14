<p align="center">
  <strong>finterm</strong><br/>
  <em>A terminal-based financial analysis tool built with Go and Bubbletea.</em>
</p>

<p align="center">
  <a href="https://github.com/shinsekai/finterm/actions/workflows/ci.yml"><img src="https://github.com/shinsekai/finterm/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/shinsekai/finterm"><img src="https://goreportcard.com/badge/github.com/shinsekai/finterm" alt="Go Report Card"></a>
  <a href="https://github.com/shinsekai/finterm/blob/main/LICENSE"><img src="https://img.shields.io/github/license/shinsekai/finterm" alt="License"></a>
  <a href="https://github.com/shinsekai/finterm/releases/latest"><img src="https://img.shields.io/github/v/release/shinsekai/finterm?include_prereleases" alt="Release"></a>
  <img src="https://img.shields.io/badge/go-%3E%3D1.24-blue?logo=go" alt="Go Version">
</p>

<p align="center">
  Trend-following signals · Real-time quotes · Macro dashboard · Sentiment-scored news
</p>

---

## What is Finterm?

Finterm is a keyboard-driven financial analysis tool that runs entirely in your terminal. It provides three independent trend signal systems (EMA crossover, BLITZ, and DESTINY), RSI-based valuation, macroeconomic indicators, and sentiment-scored news — all powered by the [Alpha Vantage](https://www.alphavantage.co/) API.

It's designed for traders and analysts who live in the terminal and want a fast, unified view of market data without leaving their workflow.

**Key principles:**

- **TradingView parity** — RSI and EMA calculations match TradingView's `ta.rsi()` and `ta.ema()` exactly. RSI uses Wilder's RMA smoothing (`alpha = 1/length`), EMA seeds with the first source value.
- **Bar close only** — All trend and valuation signals use completed (closed) bars exclusively. No intra-bar repainting. Once a bar closes, its signal is final.
- **Dual-path indicators** — Equities use Alpha Vantage's server-side technicals. Crypto uses locally computed indicators (since AV's technical endpoints don't cover weekend data). Local implementations also serve as a fallback.
- **Single binary** — `go build` produces one static binary. No runtime dependencies, no Docker, no Node.

## Features

**Trend Following** — Watchlist table with three independent signal systems per ticker:

- **EMA Crossover** — Classic EMA(10)/EMA(20) trend direction. Bullish when fast crosses above slow.
- **BLITZ System** — Correlation-based scoring combining TSI (Pearson correlation), adaptive RSI, and threshold confirmation. Fast, reactive signals.
- **DESTINY System** — Consensus-based scoring using the Trend Probability Indicator (TPI): average direction of 7 moving averages (SMA, EMA, DEMA, TEMA, WMA, HMA, LSMA) confirmed by adaptive RSI. More conservative, high-conviction signals.

Three perspectives at a glance. When all agree, conviction is highest.

SYMBOL   SIGNAL     BLITZ    DESTINY    PRICE       RSI   EMA FAST   EMA SLOW  VALUATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
AAPL     ▲ BULL    ▲ LONG   ▲ LONG    $189.84    52.3    $188.20    $186.45  ○ Fair val
MSFT     ▼ BEAR    ▼ SHORT  ▼ SHORT   $378.91    38.1    $375.10    $380.22  ◇ Underval
BTC      ▲ BULL    ▲ LONG   ○ HOLD   $67,432    48.7  $66,800    $65,200    ○ Fair val
ETH      ▼ BEAR    ○ HOLD   ▼ SHORT   $3,521    29.5   $3,480     $3,620    ◆ Oversold

**Quote Lookup** — Type any ticker to get real-time price, volume, change, and inline indicator analysis.

**Macro Dashboard** — Paneled view of GDP, CPI, inflation, federal funds rate, treasury yields, unemployment, and nonfarm payroll.

**News Feed** — Scrollable, sentiment-scored articles with ticker/topic filters, color-coded by sentiment (bullish/bearish/neutral).

**Navigation** — Fully keyboard-driven: `1-4` for tabs, `j/k` or arrows for rows, `r` to refresh, `?` for help, `q` to quit.

## Architecture

Finterm follows a strict layered architecture with clear dependency boundaries. The domain layer is pure business logic with zero HTTP or TUI imports.

<p align="center">
  <img src="docs/architecture.svg" alt="Finterm Architecture Diagram" width="800"/>
</p>

| Layer | Package | Responsibility |
|---|---|---|
| **Presentation** | `internal/tui/` | Bubbletea models, views, Lipgloss styling, tab routing |
| **Domain** | `internal/domain/` | Indicator computation (RSI, EMA), trend scoring, valuation, BLITZ + DESTINY engines |
| **Data** | `internal/alphavantage/`, `internal/cache/` | HTTP client with rate limiting and retry, TTL cache |
| **Config** | `internal/config/` | YAML + env var loading, validation |

The domain layer communicates with the data layer through interfaces — never concrete types. `cmd/finterm/main.go` wires everything together via constructor injection.

### Indicator Logic

| Indicator | Method | Details |
|---|---|---|
| **RSI** | Wilder's RMA smoothing | `alpha = 1/period`, seeded with SMA. Default period: 14 |
| **EMA** | Standard exponential | `alpha = 2/(period+1)`, seeded with first source value. Periods: 10 (fast), 20 (slow) |

| Signal | Rule |
|---|---|
| **EMA Trend** | `EMA(10) > EMA(20)` → Bullish, `EMA(10) < EMA(20)` → Bearish |
| **Valuation** | RSI < 30 → Oversold, 30–45 → Undervalued, 45–55 → Fair, 55–70 → Overvalued, > 70 → Overbought |

### BLITZ System

The BLITZ trend following system uses three independent confirmations to generate high-conviction signals:

| Component | Computation | Purpose |
|---|---|---|
| **TSI** | Pearson correlation of close vs bar index over 14 bars | Trend direction filter (+1 = up, -1 = down) |
| **Dynamic RSI** | Wilder's RSI with adaptive lookback (`min(12, bars_available)`) | Momentum strength |
| **RSI Smooth** | EMA of Dynamic RSI, same adaptive length | Noise reduction |

**Signal rules**: LONG when `TSI > 0` AND `RSI Smooth is rising` AND `RSI Smooth > 48`. SHORT when `TSI < 0` AND `RSI Smooth is falling` AND `RSI Smooth < 48`. Score latches — holds until the opposite signal fires.

**Dynamic length adaptation**: Unlike standard indicators that produce NaN for the first N bars, BLITZ adapts its lookback period to available data. At bar 5, a "12-period RSI" uses a 5-period lookback. This means signals start earlier with no warmup gap.

### DESTINY System

The DESTINY trend following system uses a consensus of 7 moving averages to build the Trend Probability Indicator (TPI), confirmed by Dynamic RSI:

| Component | Computation | Purpose |
|---|---|---|
| **SMA** | Simple MA, period 45 | Baseline trend |
| **EMA** | Exponential MA, period 45 | Responsive trend |
| **DEMA** | Double EMA, period 90 | Lag-reduced trend |
| **TEMA** | Triple EMA, period 135 | Minimal-lag trend |
| **WMA** | Weighted MA, period 45 | Recency-biased trend |
| **HMA** | Hull MA, period 45 | Fast-response trend |
| **LSMA** | Least Squares MA, period 45, offset 6 | Regression-projected trend |

Each MA is scored: +1 (rising), -1 (falling), 0 (flat). The TPI is the average of all 7 scores, ranging from -1 to +1.

**Signal rules**: LONG when `TPI > 0.5` AND `RSI Smooth is rising` AND `RSI Smooth > 56`. SHORT when `TPI < -0.5` OR (`RSI Smooth is falling` AND `RSI Smooth < 56`). The asymmetry is deliberate — entries require consensus, exits are more aggressive.

**Dynamic length adaptation**: All 7 MAs and the RSI use the same adaptive-length pattern as BLITZ, allowing signals on limited data.

### Data Flow

```
User input → tea.Cmd → fetch (AV client + cache) → domain engine → tea.Msg → view update
```

For equities, RSI and EMA are fetched from Alpha Vantage's server-side endpoints. For crypto, raw OHLCV data is fetched and indicators are computed locally in Go, because AV's technical endpoints follow equity market hours and exclude weekends.

## Requirements

- **Go 1.24+**
- **Alpha Vantage API key** — [get one here](https://www.alphavantage.co/support/#api-key) (paid tier recommended for 75 req/min)
- A terminal with 256-color support (iTerm2, Alacritty, kitty, Windows Terminal, etc.)

## Installation

### From source

```bash
git clone https://github.com/shinsekai/finterm.git
cd finterm
make build
./bin/finterm
```

### With `go install`

```bash
go install github.com/shinsekai/finterm/cmd/finterm@latest
```

## Configuration

Finterm loads configuration from `config.yaml` in the working directory. Environment variables take precedence.

```bash
# Minimum: set your API key
export FINTERM_AV_API_KEY="your_key_here"

# Or copy and edit the example config
cp config.example.yaml config.yaml
```

### Example `config.yaml`

```yaml
api:
  key: ""                          # Or use FINTERM_AV_API_KEY env var
  base_url: "https://www.alphavantage.co/query"
  rate_limit: 70                   # Requests per minute (AV premium: 75)
  timeout: 10s
  max_retries: 3

watchlist:
  equities: [AAPL, MSFT, GOOGL, AMZN, NVDA]
  crypto: [BTC, ETH, SOL]

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

blitz:
  rsi_length: 12                   # Dynamic RSI lookback period
  tsi_period: 14                   # Pearson correlation lookback period
  threshold: 48                     # RSI smooth threshold for signal generation

destiny:
  ma_length: 45                     # Base period for 7 MAs (DEMA uses 2×, TEMA uses 3×)
  rsi_length: 18                    # Dynamic RSI lookback period
  rsi_threshold: 56                 # RSI smooth threshold for signal confirmation
  lsma_offset: 6                    # LSMA projection offset

cache:
  intraday_ttl: 60s
  daily_ttl: 1h
  macro_ttl: 6h
  news_ttl: 5m
  crypto_ttl: 5m

theme:
  style: "default"                 # default | minimal | colorblind
```

## Keyboard Shortcuts

| Key | Action |
|---|---|
| `1` / `2` / `3` / `4` | Switch to Trend / Quote / Macro / News tab |
| `Tab` | Cycle to next tab |
| `j` / `k` or `↑` / `↓` | Navigate rows |
| `Enter` | Submit ticker (Quote view) / Open article (News view) |
| `r` | Refresh current view |
| `f` | Toggle filter (News view) |
| `s` | Toggle sort (News view) |
| `?` | Show help overlay |
| `q` / `Ctrl+C` | Quit |

## Development

### Prerequisites

```bash
# Install dependencies
go mod download

# Install linting tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Commands

```bash
make build          # Build binary to ./bin/finterm
make run            # Build and run
make test           # Run all tests
make test-race      # Run tests with race detector
make lint           # Run golangci-lint
make vet            # Run go vet
make fmt            # Format all code
make clean          # Remove build artifacts
```

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out

# Domain layer only (indicator logic)
go test ./internal/domain/...

# With race detector
go test -race ./...
```

### Project Structure

```
finterm/
├── cmd/finterm/main.go              # Entry point — config, DI, bootstrap
├── internal/
│   ├── tui/                         # Bubbletea views and components
│   │   ├── app.go                   # Root model + tab router
│   │   ├── trend/                   # Trend following view
│   │   ├── quote/                   # Quote lookup view
│   │   ├── macro/                   # Macro dashboard view
│   │   ├── news/                    # News feed view
│   │   └── components/              # Shared UI (spinner, table, help)
│   ├── domain/                      # Pure business logic
│   │   ├── trend/                   # Trend engine + scoring
│   │   │   └── indicators/          # RSI, EMA (local + remote)
│   │   ├── valuation/               # RSI-based valuation
│   │   ├── blitz/                   # BLITZ trend following + shared dynamic MA primitives
│   │   └── destiny/                 # DESTINY trend following system
│   ├── alphavantage/                # API client + typed models
│   ├── cache/                       # In-memory TTL cache
│   └── config/                      # YAML + env var loading
├── config.example.yaml
├── CLAUDE.md                        # Project guidelines (for AI-assisted dev)
├── PRD.md                           # Product requirements
├── EPICS.md                         # Epics and tasks
└── Makefile
```

### Contributing

1. Read `CLAUDE.md` — it's the project constitution covering coding standards, architecture rules, and testing requirements.
2. Pick a task from `EPICS.md` — each task is self-contained with acceptance criteria and test definitions.
3. Follow the conventions: `gofmt`, `go vet`, table-driven tests, error wrapping with context.
4. Run `make test lint vet` before submitting.
5. Commit messages: `<scope>: <description>` (e.g., `trend: add local RSI computation`).

### Design Decisions

**Why EMA crossover only for trend?** Simplicity and clarity. EMA(10) > EMA(20) gives a clean binary signal — the trend is up or it's down. RSI is reserved for valuation because it answers a different question: is the asset cheap or expensive relative to recent momentum.

**Why compute crypto indicators locally?** Alpha Vantage's technical endpoints (RSI, EMA) map crypto symbols to exchange-listed instruments that follow equity market hours — no weekend data, which creates gaps in 24/7 crypto markets. Fetching raw OHLCV from the crypto endpoints and computing locally ensures continuous, accurate signals.

**Why bar close only?** Intra-bar computation causes "repainting" — a signal that appears mid-bar might vanish by the time the bar closes. Using completed bars only means every signal is final. This matches how TradingView strategies execute by default.

**Why TradingView as the reference?** It's what most retail traders use. If Finterm's RSI says 52.3 and TradingView says 52.3 for the same data, trust is established immediately. The key subtlety: TradingView's RSI uses RMA smoothing (alpha = 1/length), not standard EMA smoothing (alpha = 2/(length+1)). Getting this wrong produces visibly different values.

**Why three trend signals (EMA + BLITZ + DESTINY)?** Each operates at a different speed and philosophy. EMA crossover is a simple, lagging binary — the trend is up or down. BLITZ uses Pearson correlation for trend direction and is fast and reactive (3 conditions, all AND). DESTINY polls 7 different moving averages for consensus and is more conservative — it waits for broad agreement before calling a trend. The asymmetric exit logic (OR for shorts) makes DESTINY quick to protect but slow to commit. Three signals, three speeds: when all agree, conviction is highest; when they diverge progressively, risk is rising.

**Why 7 moving averages in DESTINY?** Each MA type captures a different aspect of trend: SMA is stable but laggy, EMA reacts faster, DEMA and TEMA progressively reduce lag, WMA biases toward recent prices, HMA is the fastest responder, and LSMA projects the regression line forward. Averaging their direction votes creates a "wisdom of the crowd" signal — if 5+ out of 7 agree the trend is up, it's probably up. The TPI (Trend Probability Indicator) quantifies this consensus as a single number from -1 to +1.

**Why does BLITZ use dynamic-length indicators?** Standard indicators produce NaN for the first N bars, which is fine on TradingView with thousands of bars but problematic when data is limited. The dynamic length pattern adapts: at bar 5, a "12-period RSI" uses a 5-period lookback. This is ported directly from the Pine Script `getDynamicLength()` pattern to preserve exact signal parity.

## License

[GPL-3.0](LICENSE)

## Acknowledgements

- [Charm](https://charm.sh) — Bubbletea, Lipgloss, and Bubbles
- [Alpha Vantage](https://www.alphavantage.co) — Financial data API
- [TradingView](https://www.tradingview.com) — Reference implementation for indicator calculations
