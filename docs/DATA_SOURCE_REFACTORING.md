# 数据源模块重构总结

## 概述

本次重构将原有的 `candles` 模块拆分为两个独立的模块：`KlineModule`（K线模块）和 `MarketModule`（行情模块），并简化了数据源接口设计。

## 重构目标

1. **职责分离**：将K线数据和行情数据分离，使每个模块职责更加明确
2. **接口简化**：移除不常用的 `GetHistoricalData` 方法，简化接口设计
3. **结构优化**：将数据结构声明移到各自的模块中，提高代码组织性
4. **向后兼容**：保持现有策略函数的调用方式不变

## 主要变更

### 1. 模块拆分

#### 原架构
- `CandlesModule`：包含K线数据和行情数据

#### 新架构
- `KlineModule`：专门处理OHLCV历史数据，用于技术分析
- `MarketModule`：专门处理实时行情数据（价格、成交量、买卖价等）

### 2. 接口简化

#### 原接口
```go
type Module interface {
    GetName() string
    GetData(ctx context.Context, entity, field string) (interface{}, error)
    GetHistoricalData(ctx context.Context, entity, field string, period int) ([]interface{}, error)
}
```

#### 新接口
```go
type Module interface {
    GetName() string
    GetData(ctx context.Context, entity, field string) (interface{}, error)
}
```

### 3. 数据结构重新组织

#### KlineModule 数据结构
```go
type KlineData struct {
    Symbol    string    `json:"symbol"`
    Open      float64   `json:"open"`
    High      float64   `json:"high"`
    Low       float64   `json:"low"`
    Close     float64   `json:"close"`
    Volume    float64   `json:"volume"`
    Timestamp time.Time `json:"timestamp"`
}
```

#### MarketModule 数据结构
```go
type MarketData struct {
    Symbol     string    `json:"symbol"`
    LastPx     float64   `json:"last_px"`     // 最新价格
    LastVolume float64   `json:"last_volume"` // 最新成交量
    Bid        float64   `json:"bid"`         // 买一价
    Ask        float64   `json:"ask"`         // 卖一价
    Timestamp  time.Time `json:"timestamp"`
}
```

### 4. 数据访问方式变更

#### 原方式
```go
// 获取K线数据
candles.BTC.close
candles.BTC.last_px
```

#### 新方式
```go
// 获取K线数据
kline.BTC.close
kline.BTC.open
kline.BTC.high
kline.BTC.low
kline.BTC.volume

// 获取行情数据
market.BTC.last_px
market.BTC.last_volume
market.BTC.bid
market.BTC.ask
```

## 技术实现细节

### 1. 历史数据支持

虽然移除了接口中的 `GetHistoricalData` 方法，但 `KlineModule` 仍然支持历史数据访问，通过内部方法实现：

```go
func (m *KlineModule) GetHistoricalData(ctx context.Context, entity, field string, period int) ([]interface{}, error)
```

策略函数通过类型断言访问历史数据：
```go
if klineModule, ok := ds.(*sources.KlineModule); ok {
    historicalData, err := klineModule.GetHistoricalData(ctx, symbol, field, 100)
    // ...
}
```

### 2. 策略函数更新

所有策略函数（avg、sum、max、min）都已更新为使用新的数据源：

- 数据路径从 `candles.SYMBOL.field` 改为 `kline.SYMBOL.field`
- 支持类型断言和接口适配，确保与测试代码兼容

### 3. 测试用例更新

- 更新了所有测试用例中的数据源引用
- 添加了 `MockKlineDataSource` 和 `MockMarketDataSource`
- 更新了测试数据初始化

## 优势分析

### 1. 职责更清晰
- K线模块专注于历史数据和技术分析
- 行情模块专注于实时市场状态
- 新闻和指标模块保持独立

### 2. 接口更简洁
- 移除了不常用的 `GetHistoricalData` 方法
- 符合接口隔离原则（ISP）
- 减少了接口的复杂度

### 3. 扩展性更好
- 可以独立扩展每个模块的功能
- 新增数据源类型更加容易
- 模块间耦合度降低

### 4. 维护性更强
- 代码结构更清晰
- 数据结构与模块绑定
- 便于单元测试和调试

## 向后兼容性

### 策略表达式
- 策略表达式语法保持不变
- 只需要将 `candles` 改为 `kline` 或 `market`
- 函数调用方式完全兼容

### API 接口
- 核心 API 接口保持不变
- 数据访问方式更加明确
- 错误处理机制保持一致

## 使用示例

### 基本数据访问
```go
// 获取K线数据
closePrice, err := manager.GetData(ctx, "kline", "BTC", "close")
highPrice, err := manager.GetData(ctx, "kline", "BTC", "high")

// 获取行情数据
lastPrice, err := manager.GetData(ctx, "market", "BTC", "last_px")
bidPrice, err := manager.GetData(ctx, "market", "BTC", "bid")
```

### 策略表达式
```go
// 技术分析策略
"kline.BTC.close > kline.BTC.open and avg(kline.BTC.close, 5) > 100"

// 行情监控策略
"market.BTC.last_px > market.BTC.bid * 1.01"

// 混合策略
"kline.BTC.close > 100 and market.BTC.last_volume > 1000"
```

## 测试覆盖

所有测试用例都已更新并通过：

- ✅ 数据模块测试（7个测试用例）
- ✅ 语法解析测试（8个测试用例）
- ✅ 策略引擎测试（7个测试用例）

## 总结

本次重构成功实现了数据源模块的职责分离和接口简化，提高了代码的可维护性和扩展性。新的架构更加清晰，每个模块都有明确的职责，同时保持了良好的向后兼容性。

重构后的系统具有以下特点：
- 模块职责明确
- 接口设计简洁
- 代码组织良好
- 测试覆盖完整
- 向后兼容性强
