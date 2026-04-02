# PRD.md — Finterm Product Requirements Document

**Version:** 1.0
**Status:** Draft
**Last Updated:** 2026-03-26

---

## 1. Overview

### 1.1 Product Summary

Finterm is a terminal-based financial analysis application built with Go and Bubbletea.
It provides retail traders and technical analysts with a fast, keyboard-driven interface
for trend-following analysis, real-time price lookups, macroeconomic monitoring, and
sentiment-scored news — all within a single TUI.

### 1.2 Problem Statement

Retail traders and analysts who work primarily in the terminal lack a unified,
lightweight tool for financial analysis. Existing solutions are either:

- **Browser-based** — heavy, require context-switching, distraction-prone.
- **Paid desktop apps** — expensive, feature-bloated, not composable with terminal workflows.
- **CLI one-offs** — fragmented scripts that each solve one problem without cohesion.

Finterm fills this gap: one binary, zero browser dependencies, full keyboard navigation,
real data from Alpha Vantage.

### 1.3 Target Users

| Persona                  | Description                                                    |
|--------------------------|----------------------------------------------------------------|
| **Terminal-native trader**| Uses tmux/vim daily. Wants financial data without leaving the terminal. |
| **Technical analyst**    | Relies on RSI, EMA, and trend signals for entry/exit timing.   |
| **Macro observer**       | Monitors GDP, CPI, interest rates, and inflation for context.  |
| **Crypto + equity trader**| Trades both asset classes and needs unified analysis.          |

### 1.4 Non-Goals (v1)

- Portfolio tracking or P&L calculation.
- Order execution or broker integration.
- Backtesting engine.
- Multi-user or collaborative features.
- Mobile or web interface.

---

## 2. Data Source

### 2.1 Alpha Vantage API (Paid Subscription)

All market data is sourced from Alpha Vantage. The paid tier provides:

- **75 API calls per minute** (enforced client-side at 70/min for safety).
- **Equities**: OHLCV time series (intraday, daily, weekly, monthly), server-side technical
  indicators (RSI, EMA, MACD, and 40+ others), company fundamentals, earnings.
- **Crypto**: OHLCV time series (daily, weekly, monthly, intraday). **Note**: server-side
  technical indicator endpoints do not reliably cover crypto (they map to exchange-listed
  instruments and exclude weekends). Crypto indicators are computed locally.
- **Macro/Economic**: Real GDP, CPI, inflation, federal funds rate, treasury yields,
  unemployment, nonfarm payroll, retail sales, durable goods.
- **Commodities**: WTI, Brent, natural gas, gold, silver, copper, aluminum, wheat, corn, etc.
- **Forex**: Real-time exchange rates, intraday/daily/weekly/monthly FX time series.
- **News**: Sentiment-scored articles with topic/ticker filtering.
- **Market status**: Open/closed status for global trading venues.

### 2.2 Endpoint Mapping

| Feature          | Endpoints Used                                                          |
|------------------|-------------------------------------------------------------------------|
| Trend (equities) | `RSI`, `EMA`, `TIME_SERIES_DAILY`                                       |
| Trend (crypto)   | `DIGITAL_CURRENCY_DAILY`, `CRYPTO_INTRADAY` (RSI/EMA computed locally)  |
| Quote            | `GLOBAL_QUOTE`, `REALTIME_BULK_QUOTES` (up to 100 tickers)             |
| Macro dashboard  | `REAL_GDP`, `CPI`, `FEDERAL_FUNDS_RATE`, `TREASURY_YIELD`, `UNEMPLOYMENT`, `INFLATION`, `NONFARM_PAYROLL` |
| News feed        | `NEWS_SENTIMENT` (with ticker and topic filters)                        |
| Market status    | `MARKET_STATUS`                                                         |

### 2.3 Dual-Path Indicator Strategy

The key architectural insight: Alpha Vantage's technical indicator endpoints (RSI, EMA, etc.)
follow equity market trading days. Crypto trades 24/7 including weekends.

**Equities**: Fetch RSI and EMA directly from Alpha Vantage endpoints (server-side computation).
These return accurate, gap-free data aligned to market trading days.

**Crypto**: Fetch raw OHLCV data from `DIGITAL_CURRENCY_DAILY` or `CRYPTO_INTRADAY`, then
compute RSI and EMA locally in Go. This ensures weekend data is included and the trend
signals reflect actual market behavior.

The local implementations also serve as a **fallback for equities** if the API is
rate-limited or the technical endpoints are temporarily unavailable.

### 2.4 TradingView Parity

All indicator calculations (RSI, EMA) MUST match TradingView's Pine Script built-in functions
(`ta.rsi()`, `ta.ema()`) exactly. TradingView is the ground truth for validation.

Key implementation details:
- **RSI uses RMA (Wilder's smoothing)**, not EMA. Alpha = `1/length`, seeded with SMA.
- **EMA uses standard exponential smoothing**. Alpha = `2/(length+1)`, seeded with first source value (not SMA).
- See `CLAUDE.md` §5.2 and §5.3 for the exact step-by-step algorithms.

### 2.5 Bar Close Only Rule

**All trend and valuation computations use completed (closed) bars exclusively.**

- Indicators are never computed on the current in-progress bar.
- This eliminates intra-bar repainting: once a bar closes, its signal is final and immutable.
- For daily timeframe: yesterday's close is the most recent input to indicators.
- The current bar's real-time price is shown in the Quote view for informational purposes only,
  but is never fed into RSI, EMA, trend scoring, or valuation logic.
- When fetching data, discard the most recent data point if it represents an incomplete bar.

---

## 3. Features

### 3.1 Tab Navigation

The application presents four primary views, switchable via keyboard tabs:

| Key   | Tab          | Description                                      |
|-------|--------------|--------------------------------------------------|
| `1`   | Trend        | Trend-following analysis for watchlist tickers    |
| `2`   | Quote        | Single ticker price lookup                       |
| `3`   | Macro        | Macroeconomic indicators dashboard               |
| `4`   | News         | Sentiment-scored news feed                       |
| `q`   | —            | Quit application                                 |
| `?`   | —            | Help overlay                                     |
| `r`   | —            | Refresh current view                             |
| `Tab` | —            | Cycle to next tab                                |

### 3.2 Trend Following View

**Purpose**: Display bullish/bearish/neutral trend signals for a configurable watchlist.

**Display**:

```
 TREND FOLLOWING                                       [r] refresh  [?] help
 ─────────────────────────────────────────────────────────────────────────────
 Ticker │ Price    │ RSI(14) │ EMA(9)  │ EMA(21) │ Signal    │ Valuation
 ───────┼──────────┼─────────┼─────────┼─────────┼───────────┼────────────
 AAPL   │ $189.84  │  52.3   │ 188.20  │ 186.45  │ ▲ BULLISH │ Fair value
 MSFT   │ $378.91  │  38.1   │ 375.10  │ 380.22  │ ▼ BEARISH │ Undervalued
 NVDA   │ $875.30  │  71.8   │ 880.00  │ 860.15  │ ▲ BULLISH │ Overbought
 BTC    │ $67,432  │  48.7   │ 66,800  │ 65,200  │ ▲ BULLISH │ Fair value
 ETH    │ $3,521   │  29.5   │  3,480  │  3,620  │ ▼ BEARISH │ Oversold
 ───────┴──────────┴─────────┴─────────┴─────────┴───────────┴────────────
 Last updated: 2 min ago                          Equities: AV │ Crypto: Local
 Signals computed on last closed bar only
```

**Behavior**:

- On load: fetch data for all tickers in the watchlist. Use `REALTIME_BULK_QUOTES` for
  equity prices (single API call for up to 100 tickers), then fetch RSI/EMA per ticker.
- **Bar close only**: RSI, EMA, trend signal, and valuation are computed exclusively on
  completed (closed) bars. The current in-progress bar is never used for indicator computation.
  The "Price" column shows the latest available price for informational purposes.
- Trend signal is computed per the scoring rules in CLAUDE.md §5.4.
- Valuation label is computed per CLAUDE.md §5.5.
- Color coding: green for bullish, red for bearish, yellow for neutral.
- Auto-refresh interval: configurable, default 5 minutes.
- Arrow keys navigate rows; Enter on a row opens a detail sub-view (future enhancement).

**Data Flow**:

1. TUI sends `tea.Cmd` → triggers concurrent fetches for each ticker.
2. Domain engine receives raw data → calls RSI/EMA (remote or local) → scores signal.
3. Results return as `TrendDataMsg` → TUI updates table.

### 3.3 Quote View

**Purpose**: Look up the current price and key metrics for any single ticker.

**Display**:

```
 QUOTE LOOKUP                                          [r] refresh  [?] help
 ─────────────────────────────────────────────────────────────────────────────
 Enter ticker: AAPL█

 ┌─────────────────────────────────────────────┐
 │  AAPL — Apple Inc.                          │
 │                                             │
 │  Price:       $189.84                       │
 │  Change:      +$2.34 (+1.25%)               │
 │  Open:        $187.50                       │
 │  High:        $190.20                       │
 │  Low:         $186.80                       │
 │  Volume:      52,345,678                    │
 │  Prev Close:  $187.50                       │
 │                                             │
 │  RSI(14):     52.3  — Fair value            │
 │  EMA(9):      188.20                        │
 │  EMA(21):     186.45                        │
 │  Trend:       ▲ BULLISH                     │
 └─────────────────────────────────────────────┘
 Last updated: just now
```

**Behavior**:

- Text input field for ticker entry. Submit with Enter.
- Accepts both equity tickers (AAPL, MSFT) and crypto (BTC, ETH).
- Calls `GLOBAL_QUOTE` for equities or `CRYPTO_INTRADAY` for crypto.
- Also fetches RSI + EMA to show inline trend/valuation alongside the quote.
  **Note**: RSI, EMA, trend, and valuation are computed on the last closed bar only (see §2.5).
  The displayed price is the latest real-time price for informational purposes.
- Error state: invalid ticker → show "Ticker not found" with suggestion to use symbol search.
- History: last 10 lookups are accessible with up/down arrows in the input.

### 3.4 Macro Dashboard View

**Purpose**: Display key macroeconomic indicators in a dashboard layout.

**Display**:

```
 MACRO DASHBOARD                                       [r] refresh  [?] help
 ─────────────────────────────────────────────────────────────────────────────
 ┌── GDP ──────────────┐  ┌── Inflation ────────┐  ┌── Employment ────────┐
 │ Real GDP:  $22.67T   │  │ CPI:     312.23     │  │ Unemployment:  3.7%  │
 │ QoQ:       +2.1%     │  │ YoY:     +3.1%      │  │ Nonfarm:      +216K  │
 │ Per Capita: $67,891  │  │ Inflation: 3.0%     │  │ Trend:        Stable  │
 │ Period:    Q3 2025   │  │ Period:  Feb 2026   │  │ Period:    Feb 2026  │
 └──────────────────────┘  └─────────────────────┘  └──────────────────────┘
 ┌── Interest Rates ───────────────────────┐  ┌── Treasury Yields ─────────┐
 │ Fed Funds Rate:  5.25%                   │  │ 2Y:  4.62%                 │
 │ Previous:        5.25%                   │  │ 5Y:  4.28%                 │
 │ Last Change:     Jul 2023                │  │ 10Y: 4.15%                 │
 │                                          │  │ 30Y: 4.32%                 │
 └──────────────────────────────────────────┘  └─────────────────────────────┘
 Last updated: 3 hours ago                                    TTL: 6h
```

**Behavior**:

- Fetches all macro endpoints on initial load.
- Data is heavily cached (6h TTL) since most indicators update monthly/quarterly.
- Dashboard is read-only — no interactive inputs.
- Uses box-drawing characters for panel borders (Lipgloss styled).
- Responsive: panels reflow based on terminal width.

**Endpoints Used**:

| Panel           | Endpoint               | Refresh Rate   |
|-----------------|------------------------|----------------|
| GDP             | `REAL_GDP`, `REAL_GDP_PER_CAPITA` | Quarterly |
| Inflation       | `CPI`, `INFLATION`     | Monthly        |
| Employment      | `UNEMPLOYMENT`, `NONFARM_PAYROLL` | Monthly  |
| Interest rates  | `FEDERAL_FUNDS_RATE`   | Daily          |
| Treasury yields | `TREASURY_YIELD` (2Y, 5Y, 10Y, 30Y) | Daily  |

### 3.5 News Feed View

**Purpose**: Display recent market news with sentiment analysis.

**Display**:

```
 NEWS FEED                                             [r] refresh  [?] help
 ─────────────────────────────────────────────────────────────────────────────
 Filter: [all] equities  crypto  macro                   Sort: [newest] score
 ─────────────────────────────────────────────────────────────────────────────
 ▲ 0.72  AAPL │ Apple Reports Record Q1 Revenue Driven by Services
              │ Reuters — 2 hours ago
              │ Sentiment: Bullish │ Relevance: 0.95

 ▼ 0.28  BTC  │ Bitcoin Drops Below Key Support as Macro Fears Mount
              │ CoinDesk — 4 hours ago
              │ Sentiment: Bearish │ Relevance: 0.88

 ─ 0.51  MSFT │ Microsoft Expands Azure AI Offerings in Enterprise Push
              │ Bloomberg — 5 hours ago
              │ Sentiment: Neutral │ Relevance: 0.76
 ─────────────────────────────────────────────────────────────────────────────
 [j/k] navigate  [Enter] open in browser  [f] filter  [s] sort
```

**Behavior**:

- Fetches from `NEWS_SENTIMENT` endpoint with optional ticker/topic filters.
- Each article shows: sentiment score (0–1), ticker(s), headline, source, time, relevance.
- Color coding: green (> 0.6 bullish), red (< 0.4 bearish), yellow (neutral).
- Filter toggles: all, equities only, crypto only, macro topics.
- Sort options: newest first, sentiment score descending.
- Enter on an article copies the URL to clipboard or opens default browser.
- Scrollable list with vim-style `j`/`k` navigation.
- Auto-refresh: every 5 minutes.

---

## 4. Non-Functional Requirements

### 4.1 Performance

- **Startup time**: < 2 seconds to first rendered view (with cached data).
- **Refresh time**: < 5 seconds for full watchlist trend refresh (10 tickers).
- **Input latency**: < 50ms for keyboard navigation and tab switching.
- **Memory**: < 100MB RSS for typical usage (10-ticker watchlist, 50 news articles cached).

### 4.2 Reliability

- **Graceful degradation**: if Alpha Vantage is unreachable, show last cached data
  with a "stale data" indicator and timestamp.
- **Rate limit handling**: queue excess requests and retry, never crash on 429 responses.
- **Error recovery**: individual ticker failures do not block the rest of the watchlist.
  Failed tickers show "Error" in their row.

### 4.3 Usability

- **Zero configuration startup**: running `finterm` with just an API key set in the
  environment should work with sensible defaults.
- **Help overlay**: pressing `?` in any view shows context-sensitive key bindings.
- **Consistent navigation**: same keys work across all views (r=refresh, q=quit, Tab=next).
- **Color theme support**: default, minimal (fewer colors), and colorblind-friendly themes.

### 4.4 Compatibility

- **Terminals**: Must work correctly in: iTerm2, Terminal.app, Windows Terminal,
  Alacritty, kitty, tmux (with 256-color support).
- **OS**: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64).
- **Distribution**: Single static binary, no runtime dependencies.

---

## 5. Configuration

Configuration is loaded from `config.yaml` in the working directory, overridable by
environment variables. See CLAUDE.md §8 for the full schema.

Priority: Environment variables > config.yaml > defaults.

| Setting              | Env Var                      | Default          |
|----------------------|------------------------------|------------------|
| API key              | `FINTERM_AV_API_KEY`         | (required)       |
| Rate limit           | `FINTERM_RATE_LIMIT`         | 70 req/min       |
| Watchlist            | —                            | AAPL,MSFT,GOOGL,BTC,ETH |
| Refresh interval     | `FINTERM_REFRESH_INTERVAL`   | 5m               |
| Theme                | `FINTERM_THEME`              | default          |

---

## 6. Future Considerations (Post-v1)

These are explicitly out of scope for v1 but inform architectural decisions:

- **Charting**: ASCII/braille sparkline charts for price history in trend and quote views.
- **Alerts**: configurable threshold alerts (RSI crosses 70, EMA crossover event).
- **Commodity dashboard**: dedicated view for commodity prices (gold, oil, etc.).
- **FX view**: currency pair monitoring and cross-rate matrix.
- **Options chain**: display options data from `REALTIME_OPTIONS`.
- **Fundamentals view**: company overview, income statement, balance sheet.
- **Earnings calendar**: upcoming earnings dates.
- **Export**: CSV/JSON export of current view data.
- **Plugin system**: user-defined indicator compositions.

---

## 7. Success Metrics

| Metric                        | Target                         |
|-------------------------------|--------------------------------|
| Binary size                   | < 15MB                         |
| Startup to first render       | < 2s                           |
| Full watchlist refresh        | < 5s (10 tickers)              |
| Test coverage (domain layer)  | > 80%                          |
| Zero external runtime deps    | Single binary, no Docker/Node  |
| Clean `golangci-lint` run     | Zero warnings                  |
