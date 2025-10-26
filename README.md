# FoxFlow 智能策略交易系统

FoxFlow 是一个基于 Go 语言开发的专业级策略交易系统，支持多交易所接入、智能策略引擎和自动化交易执行。

## ✨ 核心特性

- **多交易所支持**: 支持 OKX、Binance 等主流交易所
- **智能策略引擎**: 基于 DSL 的策略表达式系统，支持复杂条件组合
- **实时数据分析**: 内置市场数据、K线、新闻等数据提供者
- **精准下单**: 支持策略条件触发和直接下单，支持止盈止损
- **交互式 CLI**: 完整的命令行界面，支持命令补全和历史记录
- **后台引擎**: 独立的策略监听引擎，实时监控策略条件
- **数据持久化**: 使用 SQLite 数据库存储账户、订单、策略等数据
- **安全认证**: 支持模拟盘和实盘交易，API 密钥安全管理

## 📁 系统架构

```
foxflow/
├── bin/                   # 编译后的可执行文件
│   ├── foxflow-cli       # CLI 交互程序
│   └── foxflow-engine    # 策略引擎程序
├── cmd/                   # 可执行程序入口
│   ├── cli/              # CLI 程序源码
│   └── engine/           # 策略引擎程序源码
├── internal/              # 内部包
│   ├── cli/              # CLI 相关代码
│   │   ├── commands/     # 命令实现
│   │   ├── command/      # 命令接口
│   │   └── render/       # 输出渲染
│   ├── config/           # 配置管理
│   ├── database/         # 数据库连接
│   ├── exchange/         # 交易所接口和实现
│   │   ├── interface.go  # 交易所接口定义
│   │   ├── okx.go        # OKX 交易所实现
│   │   └── manager.go    # 交易所管理器
│   ├── engine/           # 策略引擎
│   │   ├── builtin/      # 内置函数（avg, max, min, sum 等）
│   │   ├── provider/     # 数据提供者（市场、K线、新闻）
│   │   ├── registry/     # 函数和数据源注册器
│   │   └── syntax/       # DSL 语法解析器
│   ├── models/           # 数据模型
│   ├── news/             # 新闻源集成
│   ├── repository/       # 数据访问层
│   └── utils/            # 工具函数
├── scripts/              # 脚本和SQL文件
│   ├── build.sh          # 构建脚本
│   ├── foxflow.sql       # 数据库结构
│   └── test_*.sh         # 测试脚本
└── docs/                 # 文档
```

## 🚀 快速开始

### 1. 环境要求

- Go 1.21+
- SQLite3

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境

```bash
# 复制配置文件
cp .env.example .env

# 根据需要修改配置
vim .env
```

### 4. 构建程序

```bash
# 使用构建脚本
make build

# 或手动构建
mkdir -p bin
go build -o bin/foxflow-cli ./cmd/cli
go build -o bin/foxflow-engine ./cmd/engine
```

### 5. 运行程序

#### 启动 CLI 程序

```bash
./bin/foxflow-cli
```

#### 启动策略引擎（后台监听）

```bash
./bin/foxflow-engine
```

## 📖 使用指南

### CLI 命令大全

#### 基础命令

| 命令                       | 说明 |
|--------------------------|------|
| `show <type>`            | 查看数据列表（交易所、账户、资产、持仓等） |
| `use <type> <name>`      | 激活交易所或账户 |
| `create <type> [options]` | 创建账户或策略订单 |
| `update <type> [options]` | 更新账户信息或标的配置 |
| `cancel <type> <options>` | 取消策略订单 |
| `delete <type> <name>`   | 删除账户 |
| `close <symbol> [options]`      | 平仓指定标的 |
| `open <symbol> [options]`        | 手动执行策略订单 |
| `help`          | 显示帮助信息 |
| `exit`                   | 退出程序 |
| `quit`                   | 退出程序 |

### 详细使用示例

#### 1. 查看和激活交易所

```bash
# 查看所有可用交易所
foxflow > show exchange

# 激活 OKX 交易所
foxflow > use exchange okx
✓ 已激活交易所: okx
foxflow [okx] > 
```

#### 2. 创建和管理账户

```bash
# 创建模拟账户
foxflow [okx] > create account mock name=demo apiKey=your_api_key secretKey=your_secret_key passphrase=your_pass

# 创建实盘账户
foxflow [okx] > create account live name=demo apiKey=your_api_key secretKey=your_secret_key passphrase=your_pass

# 查看所有账户
foxflow [okx] > show account

# 激活账户
foxflow [okx] > use account demo
✓ 已激活用户: demo
foxflow [okx:demo] > 

# 更新账户信息
foxflow [okx:demo] > update account demo mock name=demo2 apiKey=new_key secretKey=new_secret passphrase=new_pass
```

#### 3. 查看市场信息

```bash
# 查看账户资产
foxflow [okx:demo] > show balance

# 查看当前持仓
foxflow [okx:demo] > show position

# 查看所有交易对
foxflow [okx:demo] > show symbol

# 搜索特定交易对
foxflow [okx:demo] > show symbol BTC

# 查看最新新闻
foxflow [okx:demo] > show news 20
```

#### 4. 配置交易参数

```bash
# 设置标的杠杆和保证金模式
foxflow [okx:demo] > update symbol BTC-USDT-SWAP margin=cross leverage=10
✓ 更新标的杠杆成功: BTC-USDT-SWAP:cross:10
```

#### 5. 策略订单

```bash
# 简单条件订单：当OKX交易所中BTC价格大于 50000 时开多仓 10 个BTC（逐仓）
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with market.okx.BTC.price > 50000

# 简单条件订单：当OKX交易所中BTC价格大于 50000 时开空仓 100000 个USDT的BTC（全仓）
foxflow [okx:demo] > open BTC-USDT-SWAP short cross 100000U with market.okx.BTC.price < 50000

# 多条件组合策略订单：当OKX交易所中BTC价格大于 50000 并且24小时总成交量大于 100000 时开多仓 10 个BTC
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with market.okx.BTC.price > 50000 and market.okx.BTC.vollume > 100000

# 使用内置函数的策略：当OKX交易所中 15分钟 K线中最近5条线数据的平均收盘价大于 100000 时开多仓 10 个BTC
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with avg(kline.okx.BTC.close, "15m", 5) > 100000

# 基于新闻的策略：指定新闻源 theblockbeats 并且新闻标题中包含“新高”时开多仓 10 个BTC
foxflow [okx:demo] > open BTC-USDT-SWAP long isolated 10 with has(news.theblockbeats.title, "新高")

# 查看所有策略订单
foxflow [okx:demo] > show order

# 取消策略订单（仅可取消未完成订单），取消开多10个BTC的策略订单
foxflow [okx:demo] > cancel order BTC-USDT-SWAP:long:10

# 平仓：平掉逐仓做多BTC的仓位
foxflow [okx:demo] > close BTC-USDT-SWAP long isolated

# 平仓：平掉全仓做空BTC的仓位
foxflow [okx:demo] > close BTC-USDT-SWAP short cross

```

## 🎯 策略表达式系统

FoxFlow 使用强大的 DSL（领域特定语言）来定义策略条件，支持复杂的逻辑组合。

### 数据提供者

#### 1. market - 实时市场数据

```bash
# 可用字段
market.price      # 当前价格
market.volume     # 24小时成交量
market.high       # 24小时最高价
market.low        # 24小时最低价

# 示例
market.price > 50000
market.volume > 1000000
```

#### 2. kline - K线数据

```bash
# K线语法格式
kline.交易所.标的.字段

# K线需要配合时间周期和数量使用
# 时间周期: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 12h, 1d, 1w
# 回溯数量: 获取最近 N 根K线

# 例：回溯okx交易所BTC的5根15分钟K线收盘价的平均值
avg(kline.okx.BTC.close, "15m", 5)

# 可用字段
kline.open[1h, 1]    # 1小时前的开盘价
kline.high[1h, 1]    # 1小时前的最高价
kline.low[1h, 1]     # 1小时前的最低价
kline.close[1h, 1]   # 1小时前的收盘价
kline.volume[1h, 1]  # 1小时前的成交量
```

#### 3. news - 新闻数据

```bash
# 语法格式
news.新闻源.字段

# 通常与 has() 函数配合使用
has(news.theblockbeats.title, "Bitcoin")       # 最近10条新闻中是否包含"Bitcoin"
```

### 内置函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `avg(data)` | 计算平均值 | `avg(kline.close[1h, 10]) > 50000` |
| `max(data)` | 获取最大值 | `max(kline.high[1h, 5]) > 52000` |
| `min(data)` | 获取最小值 | `min(kline.low[1h, 5]) < 48000` |
| `sum(data)` | 计算总和 | `sum(kline.volume[5m, 3]) > 10000` |
| `has(data, keyword)` | 检查是否包含关键字 | `has(news.title[10], "Bitcoin")` |
| `ago(timestamp, duration)` | 时间回溯 | `ago(kline.timestamp[1h, 1], "2h")` |

### 逻辑运算符

| 运算符 | 说明 | 示例 |
|--------|------|------|
| `and` | 逻辑与 | `market.price > 50000 and market.volume > 1000` |
| `or` | 逻辑或 | `market.price > 60000 or market.volume > 2000` |
| `()` | 分组 | `(market.price > 50000 and market.volume > 1000) or kline.close[1h, 1] > 51000` |

### 比较运算符

| 运算符 | 说明 |
|--------|------|
| `>` | 大于 |
| `>=` | 大于等于 |
| `<` | 小于 |
| `<=` | 小于等于 |
| `==` | 等于 |
| `!=` | 不等于 |

### 策略示例

#### 简单价格突破策略

```bash
# 当前价格突破50000
market.okx.BTC.price > 50000
```

#### 成交量确认策略

```bash
# 价格突破且成交量放大
market.okx.BTC.price > 50000 and market.okx.BTC.volume > 1000000
```

#### K线趋势策略

```bash
# 当前价格高于1小时前收盘价，且最近10根1小时K线平均收盘价上升
market.price > kline.close[1h, 1] and avg(kline.close[1h, 10]) > avg(kline.close[1h, 20])
```

#### 支撑位买入策略

```bash
# 价格接近最近5根K线的最低点，且有反弹迹象
market.price < min(kline.low[1h, 5]) * 1.01 and market.price > kline.close[5m, 1]
```

#### 新闻事件策略

```bash
# 最近10条新闻提到"Bitcoin"且价格下跌
has(news.title, "Bitcoin") and market.price < kline.close[1h, 1]
```

#### 综合策略

```bash
# 多因子确认：价格突破 + 成交量放大 + 均线支持 + 新闻利好
(market.price > 50000 and market.volume > 1000000) and 
(avg(kline.close[1h, 10]) > avg(kline.close[1h, 20])) and 
has(news.title[10], "bullish")
```

## 🔧 配置说明

### 环境变量

创建 `.env` 文件，配置以下参数：

```bash
# 数据库配置
DB_PATH=.foxflow.db

# 代理配置（可选）
PROXY_URL=http://127.0.0.1:7890

# 日志配置
LOG_LEVEL=info
LOG_FILE=logs/foxflow.log

# 引擎配置
ENGINE_CHECK_INTERVAL=5s
```

### 交易所配置

系统支持多个交易所，每个交易所需要在数据库中配置：

- 交易所名称
- API 地址
- 代理地址（可选）

## 🗄️ 数据库结构

系统使用 SQLite 数据库存储以下数据：

| 表名 | 说明 |
|------|------|
| `fox_accounts` | 交易账户信息（API密钥、密码等） |
| `fox_exchanges` | 交易所配置（名称、API地址等） |
| `fox_orders` | 策略订单（策略表达式、执行状态等） |
| `fox_symbols` | 交易对配置（杠杆、保证金等） |


## 📊 开发指南

### 添加新交易所

1. 在 `internal/exchange/` 目录下实现 `Exchange` 接口
2. 实现所有必需方法（连接、下单、查询等）
3. 在 `internal/exchange/manager.go` 中注册新交易所
4. 更新数据库中的交易所配置

### 添加新的数据提供者

1. 在 `internal/engine/provider/` 目录下实现 `Provider` 接口
2. 实现 `GetName()` 和 `GetData()` 方法
3. 在 `internal/engine/registry/` 中注册数据提供者

### 添加新的内置函数

1. 在 `internal/engine/builtin/` 目录下实现 `Builtin` 接口
2. 实现 `GetName()` 和 `Execute()` 方法
3. 在 `internal/engine/registry/` 中注册函数

### 扩展 CLI 命令

1. 在 `internal/cli/commands/` 中创建新命令文件
2. 实现 `Command` 接口（GetName, GetDescription, GetUsage, Execute）
3. 在 `internal/cli/cli.go` 中注册新命令

## 🧪 测试

```bash
# 测试 CLI 程序
./scripts/test_cli.sh

# 测试引擎程序
./scripts/test_engine.sh
```

## ⚠️ 注意事项

1. **风险提示**
   - 本系统仅用于学习和研究目的
   - 加密货币交易存在高风险，请谨慎操作
   - 建议在模拟环境中充分测试后再用于实盘交易

2. **安全建议**
   - 妥善保管 API 密钥，不要泄露给他人
   - 建议使用子账户进行交易，限制 API 权限
   - 定期更换 API 密钥
   - 不要在公共网络环境下使用

3. **性能建议**
   - 策略表达式不要过于复杂，避免影响执行效率
   - 合理设置策略引擎检查间隔
   - 及时清理已完成的历史订单

4. **最佳实践**
   - 先在模拟盘测试策略
   - 设置合理的止盈止损
   - 控制仓位大小
   - 分散投资风险

## 📄 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进项目。

## 📮 联系方式

如有问题或建议，请通过 GitHub Issues 联系我们。

---

**免责声明**: 本软件仅供学习和研究使用，使用本软件进行交易的任何盈亏由使用者自行承担。
