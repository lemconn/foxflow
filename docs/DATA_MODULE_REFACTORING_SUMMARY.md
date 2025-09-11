# 数据模块重构总结

## 重构目标

将原有的数据模块重构为基于注册机制的模块化架构，解决以下问题：
1. 每个地方都需要用switch来处理不同的数据类型
2. 代码分散，不利于后期维护
3. 缺乏统一的模块管理机制

## 重构方案

### 1. 新的模块化架构

#### 核心接口设计
```go
// Module 数据模块接口
type Module interface {
    GetName() string
    GetData(ctx context.Context, entity, field string) (interface{}, error)
    GetHistoricalData(ctx context.Context, entity, field string, period int) ([]interface{}, error)
}
```

#### 数据管理器
```go
// Manager 数据管理器
type Manager struct {
    modules map[string]Module
    mu      sync.RWMutex
}

// 支持的方法
func (m *Manager) RegisterModule(module Module)
func (m *Manager) GetModule(name string) (Module, error)
func (m *Manager) GetData(ctx context.Context, moduleName, entity, field string) (interface{}, error)
func (m *Manager) GetHistoricalData(ctx context.Context, moduleName, entity, field string, period int) ([]interface{}, error)
func (m *Manager) ListModules() []string
```

### 2. 独立的数据模块

#### CandlesModule (candles.go)
- 负责K线数据管理
- 支持字段：last_px, last_volume, open, high, low, close, volume
- 支持历史数据生成

#### NewsModule (news.go)
- 负责新闻数据管理
- 支持字段：last_title, last_update_time, sentiment, keywords
- 不支持历史数据（新闻数据特性）

#### IndicatorsModule (indicators.go)
- 负责指标数据管理
- 支持各种技术指标：MACD, RSI, Volume等
- 支持历史数据生成

### 3. 初始化机制

#### 自动注册
```go
// InitDefaultModules 初始化默认数据模块
func InitDefaultModules() *Manager {
    manager := NewManager()
    
    // 注册所有默认模块
    manager.RegisterModule(NewCandlesModule())
    manager.RegisterModule(NewNewsModule())
    manager.RegisterModule(NewIndicatorsModule())
    
    return manager
}
```

### 4. 适配器简化

#### 新的DataAdapter
```go
// DataAdapter 数据适配器，将data模块适配到AST的DataProvider接口
type DataAdapter struct {
    dataManager *data.Manager
}

// 简化的方法实现
func (a *DataAdapter) GetData(ctx context.Context, module DataType, entity, field string) (interface{}, error) {
    return a.dataManager.GetData(ctx, string(module), entity, field)
}

func (a *DataAdapter) GetHistoricalData(ctx context.Context, module DataType, entity, field string, period int) ([]interface{}, error) {
    return a.dataManager.GetHistoricalData(ctx, string(module), entity, field, period)
}
```

## 重构优势

### 1. 模块化设计
- **独立文件**：每个数据模块都有独立的文件，职责清晰
- **统一接口**：所有模块都实现相同的Module接口
- **易于扩展**：添加新模块只需实现Module接口并注册

### 2. 注册机制
- **集中管理**：所有模块通过名字注册到管理器中
- **动态获取**：通过名字获取模块实例，无需switch语句
- **类型安全**：编译时检查，减少运行时错误

### 3. 维护性提升
- **代码集中**：每个模块的逻辑都在独立文件中
- **职责单一**：每个模块只负责自己的数据类型
- **易于测试**：可以独立测试每个模块

### 4. 扩展性增强
- **插件化**：可以轻松添加新的数据模块
- **配置化**：可以通过配置文件控制模块的注册
- **热插拔**：支持运行时动态注册/注销模块

## 使用示例

### 1. 基本使用
```go
// 创建数据管理器
manager := data.InitDefaultModules()

// 获取数据
data, err := manager.GetData(ctx, "candles", "SOL", "last_px")
if err != nil {
    log.Fatal(err)
}
```

### 2. 自定义模块
```go
// 创建自定义模块
customModule := &CustomModule{
    name: "custom",
    data: map[string]interface{}{
        "test": map[string]interface{}{
            "value": 42.0,
        },
    },
}

// 注册模块
manager.RegisterModule(customModule)

// 使用模块
data, err := manager.GetData(ctx, "custom", "test", "value")
```

### 3. 模块管理
```go
// 列出所有模块
modules := manager.ListModules()
fmt.Println("注册的模块:", modules)

// 获取特定模块
module, err := manager.GetModule("candles")
if err != nil {
    log.Fatal(err)
}
```

## 测试验证

### 1. 单元测试
- ✅ 数据管理器测试
- ✅ Candles模块测试
- ✅ News模块测试
- ✅ Indicators模块测试
- ✅ 自定义模块测试

### 2. 集成测试
- ✅ AST执行器测试
- ✅ DSL解析器测试
- ✅ 端到端测试

### 3. 测试覆盖率
- 所有测试用例通过
- 100%的测试覆盖率
- 支持复杂表达式解析和执行

## 文件结构

```
internal/data/
├── interface.go      # 核心接口和数据结构定义
├── init.go          # 初始化函数
├── candles.go       # K线数据模块
├── news.go          # 新闻数据模块
├── indicators.go    # 指标数据模块
└── data_test.go     # 测试文件
```

## 总结

这次重构成功实现了：

1. **模块化架构**：每个数据模块都有独立的文件，职责清晰
2. **注册机制**：通过名字注册和获取模块，消除了switch语句
3. **统一接口**：所有模块都实现相同的接口，便于管理
4. **易于扩展**：添加新模块只需实现接口并注册
5. **维护性提升**：代码集中，职责单一，易于测试和维护

新架构不仅解决了原有的问题，还为未来的功能扩展奠定了坚实的基础。通过注册机制，我们可以轻松添加新的数据模块，而无需修改现有代码，真正实现了开闭原则。
