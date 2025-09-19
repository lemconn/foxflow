package datasources

import (
	"context"
	"fmt"
	"testing"
)

func TestDataManager(t *testing.T) {
	// 创建数据管理器并初始化默认模块
	manager := InitDefaultModules()

	// 测试列出所有模块
	modules := manager.ListModules()
	expectedModules := []string{"kline", "market", "news", "indicators"}

	if len(modules) != len(expectedModules) {
		t.Errorf("期望 %d 个模块，但得到 %d 个", len(expectedModules), len(modules))
	}

	for _, expected := range expectedModules {
		found := false
		for _, module := range modules {
			if module == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("模块 %s 未找到", expected)
		}
	}
}

func TestKlineModule(t *testing.T) {
	manager := InitDefaultModules()
	ctx := context.Background()

	// 测试获取K线数据（需要 period 参数）
	params := []DataParam{
		NewParam("period", 1),
	}
	data, err := manager.GetData(ctx, "kline", "SOL", "close", params...)
	if err != nil {
		t.Errorf("获取K线数据失败: %v", err)
	}

	if data == nil {
		t.Errorf("K线数据为空")
	}

	// 测试获取K线模块并调用历史数据方法
	klineModule, err := manager.GetModule("kline")
	if err != nil {
		t.Errorf("获取K线模块失败: %v", err)
	}

	// 使用新的 GetData 方法获取历史数据
	historicalParams := []DataParam{
		NewParam("period", 5),
	}
	historicalData, err := klineModule.GetData(ctx, "SOL", "close", historicalParams...)
	if err != nil {
		t.Errorf("获取历史数据失败: %v", err)
	}

	// 检查返回的数据类型
	historicalArray, ok := historicalData.([]interface{})
	if !ok {
		t.Errorf("期望 []interface{} 类型，但得到 %T", historicalData)
	}

	if len(historicalArray) != 5 {
		t.Errorf("期望 5 个历史数据点，但得到 %d 个", len(historicalArray))
	}
}

func TestMarketModule(t *testing.T) {
	manager := InitDefaultModules()
	ctx := context.Background()

	// 测试获取行情数据
	data, err := manager.GetData(ctx, "market", "SOL", "last_px")
	if err != nil {
		t.Errorf("获取行情数据失败: %v", err)
	}

	if data == nil {
		t.Errorf("行情数据为空")
	}

	// 测试获取行情模块
	marketModule, err := manager.GetModule("market")
	if err != nil {
		t.Errorf("获取行情模块失败: %v", err)
	}

	// 类型断言为MarketModule
	if market, ok := marketModule.(*MarketModule); ok {
		// 测试获取行情数据
		marketData, exists := market.GetMarketData("SOL")
		if !exists {
			t.Errorf("未找到SOL的行情数据")
		}

		if marketData.Symbol != "SOL" {
			t.Errorf("期望Symbol为SOL，但得到 %s", marketData.Symbol)
		}
	} else {
		t.Errorf("行情模块类型断言失败")
	}
}

func TestNewsModule(t *testing.T) {
	manager := InitDefaultModules()
	ctx := context.Background()

	// 测试获取新闻数据
	data, err := manager.GetData(ctx, "news", "theblockbeats", "last_title")
	if err != nil {
		t.Errorf("获取新闻数据失败: %v", err)
	}

	if data == nil {
		t.Errorf("新闻数据为空")
	}

}

func TestIndicatorsModule(t *testing.T) {
	manager := InitDefaultModules()
	ctx := context.Background()

	// 测试获取指标数据
	data, err := manager.GetData(ctx, "indicators", "SOL", "MACD")
	if err != nil {
		t.Errorf("获取指标数据失败: %v", err)
	}

	if data == nil {
		t.Errorf("指标数据为空")
	}

}

func TestModuleNotFound(t *testing.T) {
	manager := InitDefaultModules()
	ctx := context.Background()

	// 测试获取不存在的模块
	_, err := manager.GetData(ctx, "nonexistent", "entity", "field")
	if err == nil {
		t.Errorf("应该返回模块不存在的错误")
	}
}

func TestCustomModule(t *testing.T) {
	// 创建自定义模块
	customModule := &CustomModule{
		name: "custom",
		data: map[string]interface{}{
			"test": map[string]interface{}{
				"value": 42.0,
			},
		},
	}

	// 创建管理器并注册自定义模块
	manager := NewManager()
	manager.RegisterModule(customModule)

	ctx := context.Background()

	// 测试获取自定义模块数据
	data, err := manager.GetData(ctx, "custom", "test", "value")
	if err != nil {
		t.Errorf("获取自定义模块数据失败: %v", err)
	}

	if data != 42.0 {
		t.Errorf("期望数据 42.0，但得到 %v", data)
	}
}

// CustomModule 自定义模块用于测试
type CustomModule struct {
	name string
	data map[string]interface{}
}

func (m *CustomModule) GetName() string {
	return m.name
}

func (m *CustomModule) GetData(ctx context.Context, entity, field string, params ...DataParam) (interface{}, error) {
	entityData, exists := m.data[entity]
	if !exists {
		return nil, fmt.Errorf("entity not found: %s", entity)
	}

	entityMap, ok := entityData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid entity data type")
	}

	value, exists := entityMap[field]
	if !exists {
		return nil, fmt.Errorf("field not found: %s", field)
	}

	// 如果请求历史数据，返回重复的值数组
	if len(params) > 0 {
		for _, param := range params {
			if param.Name == "period" {
				if period, ok := param.Value.(int); ok && period > 0 {
					result := make([]interface{}, period)
					for i := 0; i < period; i++ {
						result[i] = value
					}
					return result, nil
				}
			}
		}
	}

	return value, nil
}

func (m *CustomModule) GetFunctionParamMapping() map[string]FunctionParamInfo {
	// Custom 模块不需要函数参数
	return map[string]FunctionParamInfo{}
}
