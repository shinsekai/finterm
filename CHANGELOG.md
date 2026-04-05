# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

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
