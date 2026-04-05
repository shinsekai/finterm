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

Finterm is a keyboard-driven financial analysis tool that runs entirely in your terminal. It provides EMA crossover trend signals, RSI-based valuation, macroeconomic indicators, and sentiment-scored news — all powered by the [Alpha Vantage](https://www.alphavantage.co/) API.

It's designed for traders and analysts who live in the terminal and want a fast, unified view of market data without leaving their workflow.

**Key principles:**

- **TradingView parity** — RSI and EMA calculations match TradingView's `ta.rsi()` and `ta.ema()` exactly. RSI uses Wilder's RMA smoothing (`alpha = 1/length`), EMA seeds with the first source value.
- **Bar close only** — All trend and valuation signals use completed (closed) bars exclusively. No intra-bar repainting. Once a bar closes, its signal is final.
- **Dual-path indicators** — Equities use Alpha Vantage's server-side technicals. Crypto uses locally computed indicators (since AV's technical endpoints don't cover weekend data). Local implementations also serve as a fallback.
- **Single binary** — `go build` produces one static binary. No runtime dependencies, no Docker, no Node.

## Features

**Trend Following** — Watchlist table showing EMA(10)/EMA(20) crossover signals alongside RSI-based valuation for each ticker. Supports both equities and crypto.

```
 Ticker │ Price    │ EMA(10) │ EMA(20) │ Signal    │ RSI(14) │ Valuation
 ───────┼──────────┼─────────┼─────────┼───────────┼─────────┼────────────
 AAPL   │ $189.84  │ 188.20  │ 186.45  │ ▲ BULLISH │  52.3   │ Fair value
 BTC    │ $67,432  │ 66,800  │ 65,200  │ ▲ BULLISH │  48.7   │ Fair value
 ETH    │ $3,521   │  3,480  │  3,620  │ ▼ BEARISH │  29.5   │ Oversold
```

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
| **Domain** | `internal/domain/` | Indicator computation (RSI, EMA), trend scoring, valuation |
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
| **Trend** | `EMA(10) > EMA(20)` → Bullish, `EMA(10) < EMA(20)` → Bearish |
| **Valuation** | RSI < 30 → Oversold, 30–45 → Undervalued, 45–55 → Fair, 55–70 → Overvalued, > 70 → Overbought |

### Data Flow

```
User input → tea.Cmd → fetch (AV client + cache) → domain engine → tea.Msg → view update
```

For equities, RSI and EMA are fetched from Alpha Vantage's server-side endpoints. For crypto, raw OHLCV data is fetched and indicators are computed locally in Go, because AV's technical endpoints follow equity market hours and exclude weekends.

## Requirements

- **Go 1.26+**
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
│   │   └── valuation/               # RSI-based valuation
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

## License

[GPL-3.0](LICENSE)

## Acknowledgements

- [Charm](https://charm.sh) — Bubbletea, Lipgloss, and Bubbles
- [Alpha Vantage](https://www.alphavantage.co) — Financial data API
- [TradingView](https://www.tradingview.com) — Reference implementation for indicator calculations
