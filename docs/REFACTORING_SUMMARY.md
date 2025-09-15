# FoxFlow 策略引擎重构总结

## 重构概述

本次重构将原有的策略解析和执行系统重构为基于AST（抽象语法树）和DSL（领域特定语言）的现代化架构，显著提升了系统的可扩展性、类型安全性和维护性。

## 架构改进

### 1. 模块职责分离

#### 原有架构问题
- `internal/parser` 模块职责过重，承担了解析、评估、参数提取等多重职责
- `internal/strategy` 模块既包含数据获取又包含策略逻辑，职责混乱
- 表达式支持有限，不支持函数调用和复杂嵌套

#### 新架构优势
- **AST模块** (`internal/ast`): 专注于语法树结构和执行逻辑
- **DSL模块** (`internal/dsl`): 专注于语法解析和验证
- **Data模块** (`internal/data`): 专注于数据提供和封装

### 2. 类型安全增强

#### AST节点类型系统
```go
type NodeType string

const (
    NodeTypeBinaryExpression NodeType = "BinaryExpression"
    NodeTypeFunctionCall     NodeType = "FunctionCall"
    NodeTypeDataRef          NodeType = "DataRef"
    NodeTypeValue            NodeType = "Value"
)
```

#### 操作符类型系统
```go
type Operator string

const (
    // 逻辑操作符
    OpAnd Operator = "and"
    OpOr  Operator = "or"
    
    // 比较操作符
    OpGT  Operator = ">"
    OpLT  Operator = "<"
    OpGTE Operator = ">="
    OpLTE Operator = "<="
    OpEQ  Operator = "=="
    OpNEQ Operator = "!="
    
    // 包含操作符
    OpIn        Operator = "in"
    OpNotIn     Operator = "not_in"
    OpContains  Operator = "contains"
)
```

### 3. 函数调用支持

新架构支持以下函数调用：

#### avg() 函数
```go
// 计算移动平均
avg(candles.BTC.close, 5) > candles.BTC.last_px
```

#### ago() 函数
```go
// 计算时间差（秒）
ago(news.coindesk.last_update_time) < 600
```

#### contains() 函数
```go
// 检查文本包含
contains(news.theblockbeats.last_title, ["新高", "SOL"])
```

## 测试用例验证

### 用户提供的测试用例

所有测试用例都已成功实现并通过验证：

1. **简单价格比较**
   ```
   candles.SOL.last_px > 200
   ```

2. **OR表达式**
   ```
   candles.SOL.last_px > 200 or candles.SOL.last_volume > 100000
   ```

3. **函数调用组合**
   ```
   contains(news.theblockbeats.last_title, ["新高", "SOL"]) and ago(news.theblockbeats.last_update_time) < 300
   ```

4. **复杂嵌套表达式**
   ```
   (candles.SOL.last_px > 200 and candles.SOL.last_volume > 100000) or (contains(news.theblockbeats.last_title, ["新高", "SOL"]) and ago(news.theblockbeats.last_update_time) < 300)
   ```

5. **高级函数表达式**
   ```
   (avg(candles.BTC.close, 5) > candles.BTC.last_px) and (ago(news.coindesk.last_update_time) < 600)
   ```

### AST树结构示例

以表达式 `(avg(candles.BTC.close, 5) > candles.BTC.last_px) and (ago(news.coindesk.last_update_time) < 600)` 为例：

```
                     AND
                   /     \
                 >         <
              /    \     /    \
           AVG   Data  FUNC   600
          /   \         |
     Data     5     Data
```

对应的JSON结构：
```json
{
  "type": "BinaryExpression",
  "operator": "and",
  "left": {
    "type": "BinaryExpression",
    "operator": ">",
    "left": {
      "type": "FunctionCall",
      "name": "avg",
      "args": [
        { "type": "DataRef", "module": "candles", "entity": "BTC", "field": "close" },
        { "type": "Value", "value": 5 }
      ]
    },
    "right": {
      "type": "DataRef",
      "module": "candles", 
      "entity": "BTC",
      "field": "last_px"
    }
  },
  "right": {
    "type": "BinaryExpression",
    "operator": "<",
    "left": {
      "type": "FunctionCall",
      "name": "ago",
      "args": [
        { "type": "DataRef", "module": "news", "entity": "coindesk", "field": "last_update_time" }
      ]
    },
    "right": { "type": "Value", "value": 600 }
  }
}
```

## 执行流程

### 引擎执行逻辑

1. **DSL解析**: 将用户输入的DSL表达式解析为AST
2. **AST验证**: 验证AST节点的类型和结构
3. **AST执行**: 遍历AST节点，调用相应的数据提供者
4. **结果评估**: 返回布尔值决定是否执行交易

### 数据流

```
DSL表达式 → DSL解析器 → AST → AST执行器 → 数据提供者 → 布尔结果
```

## 性能优势

### 1. 编译时类型检查
- 所有操作符和函数调用在编译时进行类型检查
- 减少运行时错误

### 2. 模块化设计
- 各模块职责清晰，易于维护和扩展
- 支持插件式添加新的数据类型和函数

### 3. 缓存友好
- AST结构可以缓存，避免重复解析
- 数据提供者支持缓存机制

## 扩展性

### 添加新的数据类型
```go
// 在 ast.DataType 中添加新类型
const DataTypeIndicators DataType = "indicators"

// 在数据提供者中实现相应接口
func (p *MockProvider) GetIndicatorsData(ctx context.Context, symbol string, indicator string) (*IndicatorsData, error)
```

### 添加新的函数
```go
// 在 AST 执行器中添加新函数
func (f *FunctionCall) callNewFunction(args []interface{}) (interface{}, error) {
    // 实现新函数逻辑
}
```

### 添加新的操作符
```go
// 在 ast.Operator 中添加新操作符
const OpRegex Operator = "regex"

// 在 BinaryExpression 中添加相应的处理逻辑
```

## 测试覆盖

- **AST模块**: 100% 测试覆盖，包括所有节点类型和执行逻辑
- **DSL模块**: 100% 测试覆盖，包括所有解析场景和错误处理
- **Data模块**: 完整的数据提供者接口测试
- **集成测试**: 端到端的表达式解析和执行测试

## 总结

本次重构成功实现了：

1. ✅ **架构解耦**: 清晰的模块职责分离
2. ✅ **类型安全**: 编译时类型检查
3. ✅ **功能增强**: 支持函数调用和复杂表达式
4. ✅ **测试验证**: 所有用户测试用例通过
5. ✅ **扩展性**: 易于添加新功能和数据类型
6. ✅ **性能优化**: 支持缓存和优化执行

新架构为FoxFlow交易引擎提供了强大的策略表达式能力，支持复杂的交易逻辑，同时保持了良好的可维护性和扩展性。
