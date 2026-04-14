# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

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
