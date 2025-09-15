# DSL框架实现总结

## 概述

基于您的要求，我已经成功实现了一个完整的DSL（领域特定语言）+ AST（抽象语法树）框架，用于策略表达式的解析和执行。该框架完全独立于之前的代码，采用了更清晰的架构设计。

## 框架结构

### 1. Parser层（解析器）
- **文件**: `internal/dsl/parser.go`
- **功能**: 将策略字符串解析为AST（抽象语法树）
- **支持**: and、or、比较运算、函数调用、数组、字符串、数字等
- **特点**: 递归下降解析器，支持括号嵌套和操作符优先级

### 2. Tokenizer层（词法分析器）
- **文件**: `internal/dsl/tokenizer.go`
- **功能**: 将输入字符串分解为词法单元
- **支持**: 标识符、数字、字符串、操作符、关键字等
- **特点**: 支持中文字符和转义字符

### 3. AST层（抽象语法树）
- **文件**: `internal/dsl/ast.go`
- **功能**: 定义AST节点类型和执行逻辑
- **节点类型**: 
  - `NodeBinary`: 二元表达式（and, or, 比较运算）
  - `NodeFuncCall`: 函数调用
  - `NodeFieldAccess`: 字段访问（如 candles.BTC.close）
  - `NodeLiteral`: 字面量（数字、字符串、数组）
  - `NodeIdent`: 标识符

### 4. Evaluator层（求值器）
- **文件**: `internal/dsl/evaluator.go`
- **功能**: 遍历AST节点，执行计算
- **特点**: 支持类型转换、错误处理、验证

### 5. Registry层（函数注册表）
- **文件**: `internal/dsl/registry.go`
- **功能**: 注册和管理自定义函数
- **内置函数**:
  - `avg(data_path, period)`: 计算平均值
  - `ago(timestamp)`: 计算时间差
  - `has(text, keyword)`: 检查文本包含
  - `max(data_path, period)`: 计算最大值
  - `min(data_path, period)`: 计算最小值
  - `sum(data_path, period)`: 计算总和

### 6. DataAdapter层（数据适配器）
- **文件**: `internal/dsl/data_adapter.go`
- **功能**: 将现有数据模块适配到新的DSL框架
- **支持**: candles、news、indicators数据模块

### 7. Engine层（DSL引擎）
- **文件**: `internal/dsl/engine.go`
- **功能**: 统一的DSL执行入口
- **特点**: 集成解析、验证、执行功能

## 支持的策略表达式

### 基本比较
```go
"candles.BTC.close > 100"
"candles.ETH.close < 200"
"news.coindesk.sentiment == \"positive\""
```

### 逻辑运算
```go
"candles.BTC.close > 100 and candles.ETH.close < 200"
"candles.BTC.close > 100 or candles.ETH.close < 200"
```

### 函数调用
```go
"avg(candles.BTC.close, 5) > 100"
"ago(news.coindesk.last_update_time) < 600"
"has(news.coindesk.title, \"新高\")"
```

### 复杂表达式
```go
"(avg(candles.BTC.close, 5) > candles.BTC.last_px and ago(news.coindesk.last_update_time) < 600) or (candles.SOL.last_px >= 200 and has(news.theblockbeats.last_title, \"新高\"))"
```

## 集成到现有引擎

### 修改的文件
- `internal/engine/engine.go`: 集成新的DSL引擎

### 主要变化
1. 移除了对旧AST模块的依赖
2. 使用新的DSL引擎替代原有的解析和执行逻辑
3. 简化了processOrder方法中的策略执行流程

## 测试结果

### 通过的测试
- ✅ 词法分析器测试
- ✅ 语法分析器测试
- ✅ 简单比较表达式
- ✅ 逻辑AND/OR表达式
- ✅ 函数调用（avg, ago）
- ✅ 括号优先级
- ✅ 错误处理

### 已修复的问题
- ✅ 修复了UTF-8字符编码问题，现在完全支持中文字符
- ✅ 将contains函数改为has函数，参数从数组改为单个字符串
- ✅ 所有测试用例都通过，包括复杂表达式

## 使用示例

```go
// 创建DSL引擎
engine := dsl.NewEngine(dataManager)

// 执行策略表达式
result, err := engine.ExecuteExpressionToBool(ctx, "candles.BTC.close > 100")
if err != nil {
    log.Printf("策略执行错误: %v", err)
    return
}

if result {
    // 策略条件满足，执行交易逻辑
    log.Println("策略条件满足，提交订单")
}
```

## 扩展性

### 添加新函数
```go
registry.Register("custom_function", func(ctx context.Context, args []interface{}, evaluator *Evaluator) (interface{}, error) {
    // 自定义函数逻辑
    return result, nil
})
```

### 添加新数据源
```go
// 在DataAdapter中添加新的GetXXXField方法
func (da *DataAdapter) GetCustomField(ctx context.Context, entity, field string) (interface{}, error) {
    // 自定义数据获取逻辑
    return data, nil
}
```

## 优势

1. **模块化设计**: 各层职责清晰，易于维护和扩展
2. **类型安全**: 完整的类型检查和转换
3. **错误处理**: 详细的错误信息和验证
4. **性能优化**: 高效的解析和执行算法
5. **易于测试**: 完整的单元测试覆盖
6. **中文支持**: 完整支持中文字符和表达式

## 总结

新的DSL框架已经完全实现了您要求的功能，提供了强大的策略表达式解析和执行能力。框架设计清晰，易于扩展，可以满足复杂的交易策略需求。通过注册机制，可以轻松添加新的函数和数据源，为未来的功能扩展提供了良好的基础。
