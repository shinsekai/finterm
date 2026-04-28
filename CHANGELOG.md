# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.8.0] — 2026-04-28

### Added
- Chart tab (tab `5`) with dual-pane visualization
  - Price pane: cell-based OHLC candlesticks using Unicode block characters
    - Bullish body: `█` (green), bearish body: `▓` (red), doji: `─` (neutral)
    - Wicks: `│` extending to high/low from body edge
  - TPI pane: braille line plotting composite TPI trajectory (FTEMA + BLITZ + DESTINY + FLOW + VORTEX) / 5
    - Zero reference line at LONG/CASH boundary
    - Fill below line: green (TPI > 0), red (TPI < 0), muted gray (±0.05 dead zone)
  - Keyboard navigation: `j`/`k` for ticker cycle, `1`/`2`/`3`/`4` for timeframes, `+`/`-` for zoom, `h`/`l` for pan, `r` for refresh
  - Status chip shows symbol, timeframe, window size, and bar-close date
  - Bar-close-only rule preserved — in-progress bars never plotted
- Sparkline component using Unicode block characters (`▁▂▃▄▅▆▇█`)
  - New Sparkline column in trend table between SYMBOL and TPI (fixed width 10 chars)
  - Color by net direction: bullish theme color when close-up, bearish when close-down, neutral when flat
  - Extended `trend.Result` with `PriceHistory []float64` (up to 30 most recent closes, bar-close-only)
- Braille canvas primitive with 2×4 subpixel resolution per character cell
  - `Canvas` struct with `Set`, `Line`, and `Render` methods for pixel-level control
  - Bresenham line algorithm for smooth curves
  - Foundation for chart TPI pane rendering
- Alpha Vantage MARKET_STATUS endpoint support
  - `GetMarketStatus()` client method returning typed `MarketStatus` and `MarketStatusEntry` structs
  - Covers Equity, Forex, Commodity, and Cryptocurrency market types with regional status data
- Market status strip in persistent app status bar (visible across all tabs)
  - Displays open/closed dots: `●` green (open), `○` gray (closed)
  - Layout: `NYSE ● · NASDAQ ● · LSE ○ · TSE ○ · FX ●`
  - Crypto venues excluded (24/7 markets convey no information)
  - Groups duplicate entries by primary exchange
  - Refreshes every 5 minutes via tea.Tick with 5-minute cache TTL
  - Loading state: `⋯` spinner; API failure: `markets: offline` in muted color
  - Colorblind theme uses glyph differentiation (`●` vs `○`) rather than hue

## [0.7.1] — 2026-04-23

### Fixed
- Added second-tier per-second burst spacing limiter (≤5 req/s) composed with existing per-minute rate limit
- Per-minute bucket now starts half-full to prevent cold-start burst of all tokens
- Classified burst-pattern API responses as transient errors (detects "premium-only" and "call frequency" messages)
- Retry transient API errors with exponential backoff (250ms base + 0-100ms jitter, capped at 5s)
- Added explicit backoff to network/timeout/read error retry paths

## [0.7.0] — 2026-04-21

### Added
- Global Ctrl+P command palette with fuzzy matching navigation
  - Execute any command or navigate to any view from anywhere in the app
  - Fuzzy search across commands, tickers, and actions
  - Keyboard-first design: type to filter, arrows to navigate, Enter to execute, Esc to cancel
  - Progress chip in trend view during parallel fetch operations
  - Real-time fetch progress with ticker-by-ticker status (✓ completed / ⟳ in progress / ○ pending)
- Parallel watchlist fetching for significantly reduced startup time
  - Concurrent ticker fetch using goroutine pools
  - Equities and crypto sections fetch in parallel batches
  - Rate limiting integration maintains API compliance while maximizing throughput
  - Overall fetch time reduced from O(n×rate_limit) to O(n/p + rate_limit) where p is parallelism
- Pure-Go SQLite-based cache layer replacing in-memory store
  - `internal/cache/sqlite.go` with full ACID compliance and WAL mode
  - Persistent cache across application restarts
  - Per-key TTL with automatic background cleanup
  - Thread-safe concurrent access with SQLite internal locking
  - Configurable cache path via `FINTERM_CACHE_PATH` environment variable or default to `~/.local/share/finterm/`

### Changed
- Cache architecture migrated from volatile memory-backed to persistent SQLite storage
- Data fetching strategy from sequential to parallel for watchlist operations
- Command accessibility streamlined through unified palette interface

## [0.6.0] — 2026-04-16

### Added
- VORTEX trend following system — 7-MA TPI plus kernel-regression long-term trend filter
  - Three kernel-weighted deviation ratios (Epanechnikov, Logistic, Wave) blended with close
    into AV series, smoothed to 150-bar SMA Mid band
  - Wave-weighted regression line for parity with the Pine source
  - Asymmetric signal logic: LONG requires TPI > 0.5 AND close > Mid AND RSI rising AND RSI > 56
    SHORT fires on TPI < -0.5 OR close < Mid OR (RSI falling AND RSI < 56).
  - Latching score identical to DESTINY / FLOW
- VORTEX column in the trend tab showing ▲ LONG / ▼ SHORT
- VORTEX summary counts in the trend summary bar
- VORTEX row in the Quote view Signal Systems section
- VORTEX engine in `internal/domain/vortex/` with full test coverage
- Configurable VORTEX parameters in `config.example.yaml` (documentation only — wiring
    in a future task; `vortex.DefaultConfig()` is the current source of truth)

### Changed
- TPI formula updated from 4-signal to 5-signal average
  (FTEMA + BLITZ + DESTINY + FLOW + VORTEX) / 5
- TPI composite GoDoc updated to reflect the new 5-signal average

## [0.5.0] — 2026-04-15

### Added
- FLOW trend following system — double-smoothed Heikin-Ashi momentum scoring
  - Sebastine indicator: EMA-smoothed OHLC → Heikin-Ashi synthesis → second EMA smoothing
    → body ratio as percentage. First system to use full OHLC data, not just closes
  - Dynamic RSI with threshold confirmation (length 14, threshold 55)
  - Asymmetric signal logic matching DESTINY pattern (AND for longs, OR for shorts)
- FLOW column in trend and quote views showing ▲ LONG / ▼ SHORT
- FLOW summary counts in trend summary bar
- TPI numeric value displayed in trend table (e.g., `+0.50 LONG` or `-0.25 CASH`)
- TPI composite now averages 4 signals: FTEMA, BLITZ, DESTINY, FLOW
- Configurable FLOW parameters via `config.yaml` (rsi_length, rsi_threshold, fast_length, slow_length)

### Changed
- TPI formula updated from 3-signal to 4-signal average
- TPI column now shows: gauge + numeric value + colored LONG/CASH label
- LONG labels render in green, CASH and SHORT labels render in red across all columns
- Signal badge text alignment: ▲  LONG (7 chars) padded to match ▼ SHORT (7 chars)

### Fixed
- Signal badge colors not visible in trend table — theme styles were overridden by table
  row styling. Fixed by using hardcoded foreground colors that survive row composition
- Model receiver types converted to pointers to fix lint issues and prevent value copying

## [0.4.0] — 2026-04-15

### Added
- TPI (Trend Probability Indicator) composite signal — averages FTEMA, BLITZ, and DESTINY
  into a single score from -1 to +1. TPI > 0 signals LONG, TPI ≤ 0 signals CASH
- TPI gauge column in the trend table with per-character gradient color (red → orange → yellow → green)
- FTEMA column replacing the old SIGNAL column — same EMA crossover logic, new label
- Signal systems section in Quote view showing FTEMA, BLITZ, DESTINY badges with TPI composite gauge
- Shared `internal/domain/dynamo/` package for dynamic-length indicator primitives
  (extracted from blitz/ so both BLITZ and DESTINY import from a common foundation)

### Changed
- Trend table columns simplified: removed EMA FAST and EMA SLOW columns
- New column order: SYMBOL, TPI (with gauge), FTEMA, BLITZ, DESTINY, PRICE, RSI, VALUATION
- All signal badges unified: FTEMA, BLITZ, and DESTINY all show ▲ LONG (green) or ▼ SHORT (red)
  with empty cell when no signal — no more BULL/BEAR/HOLD labels
- Quote Technical Indicators card now shows signal systems + TPI composite instead of EMA values
- Upgraded Go from 1.24.2 to 1.26.2

### Fixed
- LONG badges invisible in trend table — BullishBadge background was stripped by table row
  styling, leaving dark text on dark background. Fixed by using foreground-only text colors
- TPI gauge now renders per-character gradient colors instead of flat single color
- Quote indicators card label alignment — fixed ANSI-aware padding for signal system labels
- Summary bar now shows "TPI:" prefix label before LONG/CASH counts

## [0.3.0] — 2026-04-14

### Added
- DESTINY trend following system — consensus-based scoring using 7 moving averages
  - Trend Probability Indicator (TPI): average direction score of SMA, EMA, DEMA, TEMA, WMA, HMA, LSMA
  - Dynamic RSI with threshold confirmation (length 18, threshold 56)
  - Asymmetric signal logic: LONG requires TPI consensus + RSI confirmation (AND), SHORT can fire on TPI alone (OR)
  - Latching score: holds LONG (+1) or SHORT (-1) until the opposite signal fires
- DESTINY column in trend tab showing ▲ LONG / ▼ SHORT / ○ HOLD badges
- DESTINY summary counts in the trend summary bar
- New dynamic moving average primitives in `internal/domain/blitz/`:
  - `DynamicWMA` — Weighted MA with linearly decreasing weights (`weight = len - i`)
  - `DynamicDEMA` — Double EMA for lag reduction (`2 × EMA - EMA(EMA)`)
  - `DynamicTEMA` — Triple EMA for further lag reduction (`3 × (ema1 - ema2) + ema3`)
  - `DynamicHMA` — Hull MA for fast response (`WMA(2×WMA(half) - WMA(full), √len)`)
  - `DynamicLSMA` — Least Squares MA via linear regression projection
- DESTINY engine in `internal/domain/destiny/` with full test coverage
- Configuration section for DESTINY parameters (ma_length, rsi_length, rsi_threshold, lsma_offset)

## [0.2.0] — 2026-04-14

### Added
- BLITZ trend following system — a custom multi-signal scoring engine ported from Pine Script
  - Trend Strength Index (TSI): rolling Pearson correlation of price vs time for trend direction
  - Dynamic RSI: adaptive-length RSI that produces values from bar 1 (no NaN warmup period)
  - Dynamic EMA smoothing of RSI for noise reduction
  - Three-condition signal: TSI confirms direction, smoothed RSI confirms momentum, threshold confirms strength
  - Latching score: holds LONG (+1) or SHORT (-1) until the opposite signal fires
- BLITZ column in the trend tab showing ▲ LONG / ▼ SHORT / ○ HOLD badges alongside the existing EMA crossover signal
- BLITZ summary counts in the trend summary bar (e.g., "BLITZ: 5 LONG  2 SHORT  1 HOLD")
- Dynamic moving average primitives in `internal/domain/blitz/`:
  - `DynamicRMA` — Wilder's smoothing with per-bar adaptive length
  - `DynamicEMA` — exponential MA with per-bar adaptive length
  - `DynamicSMA` — simple MA with adaptive window and NaN handling
  - `PearsonCorrelation` — rolling window Pearson r computation
- Daily time series fetching for equities to support BLITZ computation
- GitHub Actions release workflow with goreleaser for cross-platform binaries
- CI workflow with build, test (race detector), vet, and golangci-lint

### Fixed
- golangci-lint v2.x compatibility for Go 1.24

## [0.1.0] — 2026-04-05

### Added
- Trend following tab with EMA(10)/EMA(20) crossover signals and RSI-based valuation
- Single ticker quote lookup with price cards, RSI gauge, day range bar, and EMA crossover delta
- Macroeconomic dashboard with GDP, CPI, employment, interest rates, and treasury yields
- Sentiment-scored news feed with filter (all/equities/crypto/macro) and sort (newest/score)
- Alpha Vantage API client with rate limiting (70 req/min), exponential backoff retry, and typed responses
- Dual-path indicator strategy: server-side for equities, local computation for crypto
- Local RSI and EMA implementations matching TradingView parity (Wilder's RMA for RSI)
- In-memory TTL cache with per-data-type expiration
- YAML + environment variable configuration with validation
- Three color themes: default (Dracula-inspired), minimal (monochrome), colorblind-friendly
- Tab navigation (1-4), keyboard shortcuts, context-sensitive help overlay (?)
- Graceful degradation: cached data shown when API is offline, per-ticker error isolation
- Status bar with connection state (● online/offline/rate limited), error count, and update time
- Signal summary bar in trend view with at-a-glance market read
- Active row cursor marker (▸) and alternating row backgrounds in trend table
- Equities/crypto section separator in trend watchlist
- Loading progress indicator during data fetch
- Lookup history with up/down arrow navigation in quote view
