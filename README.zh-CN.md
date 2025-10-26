# FoxFlow - 智能策略交易系统

基于 Go 语言的专业级策略交易系统，支持多交易所接入、智能策略引擎和自动化交易执行。

## 语言版本 / Language Versions

- [English](README.md) | [中文](README.zh-CN.md)

## 核心特性

- **多交易所支持**: OKX 等主流交易所
- **智能策略引擎**: 基于 DSL 的策略表达式系统
- **实时数据**: 市场数据、K线、新闻等数据提供者
- **交互式 CLI**: 完整的命令行界面
- **后台引擎**: 独立的策略监听引擎
- **数据持久化**: SQLite 数据库存储

## 快速开始

### 环境要求
- Go 1.21+
- SQLite3

### 安装和运行

```bash
# 安装依赖
go mod tidy

# 构建程序
make build

# 启动 CLI
./bin/foxflow-cli

# 启动策略引擎（后台监听）
./bin/foxflow-engine
```

## 使用指南

### 基础命令

| 命令 | 说明 |
|------|------|
| `show <type>` | 查看数据（交易所、账户、资产等） |
| `use <type> <name>` | 激活交易所或账户 |
| `create <type> [options]` | 创建账户或策略订单 |
| `open <symbol> [options]` | 执行策略订单 |
| `close <symbol> [options]` | 平仓指定标的 |
| `cancel <type> <options>` | 取消策略订单 |

### 使用示例

```bash
# 查看和激活交易所
foxflow > show exchange
foxflow > use exchange okx

# 创建账户
foxflow [okx] > create account mock name=demo apiKey=xxx secretKey=xxx passphrase=xxx

# 查看资产和持仓
foxflow [okx:demo] > show balance
foxflow [okx:demo] > show position

# 策略订单示例
# 价格突破策略
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with market.okx.BTC.price > 50000

# K线策略
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with avg(kline.okx.BTC.close, "15m", 5) > 100000

# 新闻策略
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with has(news.blockbeats.title, "新高")

# 平仓
foxflow [okx:demo] > close BTC-USDT-SWAP long isolated
```

## 策略表达式系统

### 数据提供者

**market** - 实时市场数据
```bash
market.okx.BTC.price > 50000          # 当前价格
market.okx.BTC.volume > 1000000       # 24小时成交量
```

**kline** - K线数据
```bash
# 时间周期: 1m, 5m, 15m, 1h, 4h, 1d
avg(kline.okx.BTC.close, "15m", 5) > 100000  # 5根15分钟K线平均收盘价
```

**news** - 新闻数据
```bash
has(news.blockbeats.title, "Bitcoin")     # 新闻标题包含关键字
```

### 内置函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `avg(data)` | 平均值 | `avg(kline.okx.BTC.close, "15m", 5) > 50000` |
| `max(data)` | 最大值 | `max(kline.okx.BTC.close, "15m", 5) > 52000` |
| `min(data)` | 最小值 | `min(kline.okx.BTC.close, "15m", 5) < 48000` |
| `has(data, keyword)` | 包含关键字 | `has(news.blockbeats.title, "Bitcoin")` |

### 运算符

| 类型 | 运算符 | 示例 |
|------|--------|------|
| 逻辑 | `and`, `or`, `()` | `market.okx.BTC.price > 50000 and market.okx.BTC.volume > 1000` |
| 比较 | `>`, `>=`, `<`, `<=`, `==`, `!=` | `market.okx.BTC.price > 50000` |

### 策略示例

```bash
# 价格突破策略
market.okx.BTC.price > 50000

# 成交量确认策略
market.okx.BTC.price > 50000 and market.okx.BTC.volume > 1000000

# K线趋势策略
avg(kline.okx.BTC.close, "15m", 5) > avg(kline.okx.BTC.close, "5m", 5)

# 新闻事件策略
has(news.blockbeats.title, "Bitcoin") and market.okx.BTC.price < 120000
```

## 配置说明

### 环境变量

创建 `.env` 文件：

```bash
DB_PATH=.foxflow.db
LOG_LEVEL=info
ENGINE_CHECK_INTERVAL=5s
```

## 开发指南

### 扩展功能

- **新交易所**: 在 `internal/exchange/` 实现 `Exchange` 接口
- **数据提供者**: 在 `internal/engine/provider/` 实现 `Provider` 接口  
- **内置函数**: 在 `internal/engine/builtin/` 实现 `Builtin` 接口
- **CLI命令**: 在 `internal/cli/commands/` 实现 `Command` 接口

### 测试

```bash
./scripts/test_cli.sh
./scripts/test_engine.sh
```

## 注意事项

**风险提示**: 本系统仅供学习研究，加密货币交易存在高风险，请谨慎操作

**安全建议**: 妥善保管API密钥，建议使用子账户，定期更换密钥

**性能建议**: 策略表达式不宜过于复杂，合理设置检查间隔

## 许可证

Apache 2.0 许可证 - 详见 [LICENSE](LICENSE)

---

**免责声明**: 本软件仅供学习研究使用，交易盈亏由使用者自行承担。
