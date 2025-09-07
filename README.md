# FoxFlow 策略下单系统

FoxFlow 是一个基于 Go 语言开发的策略下单系统，支持多交易所、多策略的自动化交易。

## 功能特性

- 🏦 **多交易所支持**: 支持 OKX、Binance 等主流交易所
- 📊 **策略系统**: 内置成交量、MACD、RSI 等策略，支持自定义策略
- 🎯 **智能下单**: 支持策略条件触发下单和直接下单
- 💻 **交互式 CLI**: 提供完整的命令行界面，支持命令补全和历史记录
- 🔄 **后台引擎**: 独立的策略监听引擎，实时监控策略条件
- 🗄️ **数据持久化**: 使用 SQLite 数据库存储用户、订单、策略等数据
- 🔧 **灵活配置**: 支持环境变量配置，可配置代理、日志等

## 系统架构

```
foxflow/
├── bin/                   # 编译后的可执行文件
│   ├── foxflow-cli       # CLI 程序
│   └── foxflow-engine    # 策略引擎程序
├── cmd/                   # 可执行程序入口
│   ├── cli/              # CLI 程序源码
│   └── engine/           # 策略引擎程序源码
├── internal/              # 内部包
│   ├── cli/              # CLI 相关代码
│   ├── config/           # 配置管理
│   ├── database/         # 数据库连接
│   ├── exchange/         # 交易所接口和实现
│   ├── engine/           # 策略引擎
│   ├── models/           # 数据模型
│   └── strategy/         # 策略接口和实现
├── pkg/                   # 公共包
│   ├── parser/           # 策略表达式解析器
│   └── utils/            # 工具函数
├── scripts/              # 脚本和SQL文件
└── docs/                 # 文档
```

## 快速开始

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
cp config.env.example .env

# 根据需要修改配置
vim .env
```

### 4. 构建程序

```bash
# 使用构建脚本
./scripts/build.sh

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

#### 启动策略引擎

```bash
./bin/foxflow-engine
```

## 使用指南

### CLI 命令

#### 基础命令

- `show <type>` - 查看数据列表
- `use <type> <name>` - 激活交易所或用户
- `create <type> [options]` - 创建用户、标的或策略订单
- `update <type> <value>` - 设置杠杆或保证金模式
- `cancel <type> <id>` - 取消策略订单
- `delete <type> <name>` - 删除用户或标的
- `help [command]` - 显示帮助信息
- `exit` - 退出程序

#### 使用示例

```bash
# 1. 查看可用交易所
foxflow > show exchanges

# 2. 激活交易所
foxflow > use exchanges okx
foxflow [okx] > 

# 3. 创建用户
foxflow [okx] > create users --username=user1 --ak=your_api_key --sk=your_secret_key --trade_type=mock

# 4. 激活用户
foxflow [okx] > use users user1
foxflow [okx:user1] > 

# 5. 查看用户资产
foxflow [okx:user1] > show assets

# 6. 创建策略订单
foxflow [okx:user1] > create ss --symbol=BTC/USDT --side=buy --posSide=long --px=50000 --sz=0.01 --strategy=volume:volume>100

# 7. 查看策略订单
foxflow [okx:user1] > show ss
```

### 策略表达式

支持复杂的策略表达式，使用逻辑运算符组合多个策略：

```bash
# 单个策略
--strategy=volume:volume>100

# 多个策略组合
--strategy=(volume:volume>100 and macd:macd>50) or rsi:rsi<10
```

#### 支持的策略

- **volume**: 成交量策略
  - 参数: `threshold` - 成交量阈值
  - 示例: `volume:volume>1000`

- **macd**: MACD 策略
  - 参数: `threshold` - MACD 阈值
  - 示例: `macd:macd>50`

- **rsi**: RSI 策略
  - 参数: `threshold` - RSI 阈值 (0-100)
  - 示例: `rsi:rsi<30`

#### 逻辑运算符

- `and` - 逻辑与
- `or` - 逻辑或
- `()` - 分组

## 数据库结构

系统使用 SQLite 数据库存储以下数据：

- `fox_users` - 用户信息
- `fox_symbols` - 交易对配置
- `fox_ss` - 策略订单
- `fox_exchanges` - 交易所配置
- `fox_strategies` - 策略配置

## 开发指南

### 添加新交易所

1. 在 `internal/exchange/` 目录下实现 `Exchange` 接口
2. 在 `internal/exchange/manager.go` 中注册新交易所
3. 更新数据库中的交易所配置

### 添加新策略

1. 在 `internal/strategy/` 目录下实现 `Strategy` 接口
2. 在 `internal/strategy/interface.go` 中注册新策略
3. 更新数据库中的策略配置

### 扩展 CLI 命令

1. 在 `internal/cli/commands.go` 中实现 `Command` 接口
2. 在 `internal/cli/cli.go` 中注册新命令
3. 更新命令补全器

## 测试

```bash
# 测试 CLI 程序
./scripts/test_cli.sh

# 测试引擎程序
./scripts/test_engine.sh
```

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进项目。

## 注意事项

- 本系统仅用于学习和研究目的
- 使用前请确保了解相关风险
- 建议在模拟环境中充分测试后再用于实盘交易
- 请妥善保管 API 密钥，不要泄露给他人