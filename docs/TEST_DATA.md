# 测试数据说明

## 概述

项目已配置完整的测试数据，用于验证 Foxflow 交易策略引擎的各项功能。

## 测试数据统计

### 用户数据 (4个)
- **test_user_1**: Binance 用户，模拟交易，激活状态
- **test_user_2**: OKX 用户，真实交易，激活状态  
- **test_user_3**: Gate 用户，模拟交易，未激活状态
- **demo_trader**: Binance 用户，模拟交易，激活状态

### 标的数据 (5个)
- **BTCUSDT** (用户1): Binance 现货，10倍杠杆，逐仓
- **ETHUSDT** (用户1): Binance 现货，5倍杠杆，全仓
- **BTC-USDT-SWAP** (用户2): OKX 永续合约，20倍杠杆，逐仓
- **ETH-USDT-SWAP** (用户2): OKX 永续合约，15倍杠杆，全仓
- **ADAUSDT** (用户4): Binance 现货，3倍杠杆，逐仓

### 策略订单数据 (9个)

#### 按状态分布
- **waiting**: 4个订单（等待执行）
- **pending**: 2个订单（处理中）
- **filled**: 2个订单（已成交）
- **cancelled**: 1个订单（已取消）

#### 按策略分布
- **macd**: 4个订单（MACD策略）
- **volume**: 3个订单（成交量策略）
- **rsi**: 2个订单（RSI策略）

#### 按交易所分布
- **Binance**: 6个订单
- **OKX**: 3个订单

## 测试场景

### 1. 多用户场景
- 不同交易所的用户
- 不同交易类型（模拟/真实）
- 不同激活状态

### 2. 多标的场景
- 现货交易对（BTCUSDT, ETHUSDT, ADAUSDT）
- 永续合约（BTC-USDT-SWAP, ETH-USDT-SWAP）
- 不同杠杆倍数（3x-20x）
- 不同保证金类型（逐仓/全仓）

### 3. 多策略场景
- MACD 策略：趋势跟踪
- 成交量策略：量价分析
- RSI 策略：超买超卖

### 4. 多订单状态场景
- 等待执行：测试订单队列
- 处理中：测试订单执行
- 已成交：测试成交记录
- 已取消：测试订单取消

## 数据验证

### 使用 SQL 查询验证
```sql
-- 查看所有用户
SELECT * FROM fox_users;

-- 查看所有标的
SELECT * FROM fox_symbols;

-- 查看所有订单
SELECT * FROM fox_ss;

-- 按状态统计订单
SELECT status, COUNT(*) FROM fox_ss GROUP BY status;

-- 按策略统计订单
SELECT strategy, COUNT(*) FROM fox_ss GROUP BY strategy;
```

### 使用 Go 代码验证
```go
// 获取所有用户
var users []models.FoxUser
db.Find(&users)

// 获取所有订单
var orders []models.FoxSS
db.Find(&orders)

// 按用户统计订单
var userOrderStats []struct {
    UserID uint
    Count  int64
}
db.Model(&models.FoxSS{}).Select("user_id, count(*) as count").Group("user_id").Scan(&userOrderStats)
```

## 功能测试建议

1. **用户管理**: 测试用户激活/停用功能
2. **标的管理**: 测试标的添加/删除功能
3. **策略执行**: 测试不同策略的执行逻辑
4. **订单管理**: 测试订单创建/更新/取消功能
5. **数据查询**: 测试各种查询和统计功能
6. **交易所集成**: 测试不同交易所的API调用

## 注意事项

- 测试数据使用 `FirstOrCreate` 方法插入，避免重复
- 所有测试数据都是安全的模拟数据
- 真实交易的API密钥需要替换为实际值
- 测试数据会在数据库初始化时自动插入
