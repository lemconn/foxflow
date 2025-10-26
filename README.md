# FoxFlow - Intelligent Strategy Trading System

A professional-grade strategy trading system based on Go, supporting multi-exchange integration, intelligent strategy engine, and automated trading execution.

## Language Versions / 语言版本

- [English](README.md) | [中文](README.zh-CN.md)

## Core Features

- **Multi-Exchange Support**: OKX and other mainstream exchanges
- **Intelligent Strategy Engine**: DSL-based strategy expression system
- **Real-time Data**: Market data, K-line, news and other data providers
- **Interactive CLI**: Complete command-line interface
- **Background Engine**: Independent strategy monitoring engine
- **Data Persistence**: SQLite database storage

## Quick Start

### Requirements
- Go 1.21+
- SQLite3

### Installation and Running

```bash
# Install dependencies
go mod tidy

# Build the program
make build

# Start CLI
./bin/foxflow-cli

# Start strategy engine (background monitoring)
./bin/foxflow-engine
```

## User Guide

### Basic Commands

| Command | Description |
|---------|-------------|
| `show <type>` | View data (exchanges, accounts, assets, etc.) |
| `use <type> <name>` | Activate exchange or account |
| `create <type> [options]` | Create account or strategy order |
| `open <symbol> [options]` | Execute strategy order |
| `close <symbol> [options]` | Close specified position |
| `cancel <type> <options>` | Cancel strategy order |

### Usage Examples

```bash
# View and activate exchanges
foxflow > show exchange
foxflow > use exchange okx

# Create account
foxflow [okx] > create account mock name=demo apiKey=xxx secretKey=xxx passphrase=xxx

# View assets and positions
foxflow [okx:demo] > show balance
foxflow [okx:demo] > show position

# Strategy order examples
# Price breakout strategy
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with market.okx.BTC.price > 50000

# K-line strategy
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with avg(kline.okx.BTC.close, "15m", 5) > 100000

# News strategy
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with has(news.blockbeats.title, "breakthrough")

# Close position
foxflow [okx:demo] > close BTC-USDT-SWAP long isolated
```

## Strategy Expression System

### Data Providers

**market** - Real-time market data
```bash
market.okx.BTC.price > 50000          # Current price
market.okx.BTC.volume > 1000000       # 24-hour volume
```

**kline** - K-line data
```bash
# Time periods: 1m, 5m, 15m, 1h, 4h, 1d
avg(kline.okx.BTC.close, "15m", 5) > 100000  # Average of 5 15-minute K-line closing prices
```

**news** - News data
```bash
has(news.blockbeats.title, "Bitcoin")     # News title contains keyword
```

### Built-in Functions

| Function | Description | Example |
|----------|-------------|---------|
| `avg(data)` | Average | `avg(kline.okx.BTC.close, "15m", 5) > 50000` |
| `max(data)` | Maximum | `max(kline.okx.BTC.close, "15m", 5) > 52000` |
| `min(data)` | Minimum | `min(kline.okx.BTC.close, "15m", 5) < 48000` |
| `has(data, keyword)` | Contains keyword | `has(news.blockbeats.title, "Bitcoin")` |

### Operators

| Type | Operators | Example |
|------|-----------|---------|
| Logical | `and`, `or`, `()` | `market.okx.BTC.price > 50000 and market.okx.BTC.volume > 1000` |
| Comparison | `>`, `>=`, `<`, `<=`, `==`, `!=` | `market.okx.BTC.price > 50000` |

### Strategy Examples

```bash
# Price breakout strategy
market.okx.BTC.price > 50000

# Volume confirmation strategy
market.okx.BTC.price > 50000 and market.okx.BTC.volume > 1000000

# K-line trend strategy
avg(kline.okx.BTC.close, "15m", 5) > avg(kline.okx.BTC.close, "5m", 5)

# News event strategy
has(news.blockbeats.title, "Bitcoin") and market.okx.BTC.price < 120000
```

## Configuration

### Environment Variables

Create `.env` file:

```bash
DB_PATH=.foxflow.db
LOG_LEVEL=info
ENGINE_CHECK_INTERVAL=5s
```

## Development Guide

### Extending Features

- **New Exchange**: Implement `Exchange` interface in `internal/exchange/`
- **Data Provider**: Implement `Provider` interface in `internal/engine/provider/`
- **Built-in Function**: Implement `Builtin` interface in `internal/engine/builtin/`
- **CLI Command**: Implement `Command` interface in `internal/cli/commands/`

### Testing

```bash
./scripts/test_cli.sh
./scripts/test_engine.sh
```

## Important Notes

**Risk Warning**: This system is for learning and research purposes only. Cryptocurrency trading involves high risks. Please operate with caution.

**Security Recommendations**: Keep API keys secure, use sub-accounts, and change keys regularly.

**Performance Tips**: Keep strategy expressions simple and set reasonable check intervals.

## License

Apache 2.0 License - See [LICENSE](LICENSE) for details.

---

**Disclaimer**: This software is for learning and research purposes only. Users are responsible for their own trading profits and losses.