package provider

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDataManager(t *testing.T) {
	// 创建数据管理器并初始化默认模块
	manager := InitDefaultProviders()

	// 测试列出所有模块
	modules := manager.ListProviders()
	expectedProviders := []string{"kline", "market", "news", "indicators"}

	if len(modules) != len(expectedProviders) {
		t.Errorf("期望 %d 个模块，但得到 %d 个", len(expectedProviders), len(modules))
	}

	for _, expected := range expectedProviders {
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

func TestKlineProvider(t *testing.T) {
	manager := InitDefaultProviders()
	ctx := context.Background()

	// 测试获取K线数据（需要 limit 参数）
	params := []DataParam{
		NewParam("limit", 1),
		NewParam("interval", "15m"),
	}
	data, err := manager.GetData(ctx, "kline", "okx", "BTC.close", params...)
	if err != nil {
		t.Errorf("获取K线数据失败: %v", err)
	}

	if data == nil {
		t.Errorf("K线数据为空")
	}

	// 测试获取K线模块并调用历史数据方法
	klineProvider, err := manager.GetProvider("kline")
	if err != nil {
		t.Errorf("获取K线模块失败: %v", err)
	}

	// 使用新的 GetData 方法获取历史数据
	historicalParams := []DataParam{
		NewParam("limit", 5),
		NewParam("interval", "1h"),
	}
	historicalData, err := klineProvider.GetData(ctx, "okx", "BTC.close", historicalParams...)
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

func TestMarketProvider(t *testing.T) {
	manager := InitDefaultProviders()
	ctx := context.Background()

	// 测试获取行情数据
	data, err := manager.GetData(ctx, "market", "okx", "BTC.last_px")
	if err != nil {
		t.Errorf("获取行情数据失败: %v", err)
	}

	if data == nil {
		t.Errorf("行情数据为空")
	}

	// 测试获取行情模块
	marketProvider, err := manager.GetProvider("market")
	if err != nil {
		t.Errorf("获取行情模块失败: %v", err)
	}

	// 测试行情数据获取
	lastPx, err := marketProvider.GetData(ctx, "okx", "BTC.last_px")
	if err != nil {
		t.Errorf("获取行情价格失败: %v", err)
	}

	if lastPx == nil {
		t.Errorf("行情价格不应为空")
	}
}

func TestNewsProvider(t *testing.T) {
	manager := InitDefaultProviders()
	ctx := context.Background()

	// 等待一段时间让新闻提供者获取数据
	time.Sleep(2 * time.Second)

	// 测试获取新闻数据（使用 blockbeats 数据源）
	data, err := manager.GetData(ctx, "news", "blockbeats", "title")
	if err != nil {
		t.Errorf("获取新闻数据失败: %v", err)
	}

	if data == nil {
		t.Errorf("新闻数据为空")
	}

	// 验证数据是字符串类型
	if title, ok := data.(string); ok {
		if title == "" {
			t.Errorf("新闻标题不应为空")
		}
		t.Logf("获取到新闻标题: %s", title)
	} else {
		t.Errorf("新闻标题类型错误，期望 string，实际 %T", data)
	}
}

func TestIndicatorsProvider(t *testing.T) {
	manager := InitDefaultProviders()
	ctx := context.Background()

	// 测试获取指标数据
	data, err := manager.GetData(ctx, "indicators", "okx", "rsi")
	if err != nil {
		t.Errorf("获取指标数据失败: %v", err)
	}

	if data == nil {
		t.Errorf("指标数据为空")
	}

}

func TestProviderNotFound(t *testing.T) {
	manager := InitDefaultProviders()
	ctx := context.Background()

	// 测试获取不存在的模块
	_, err := manager.GetData(ctx, "nonexistent", "dataSource", "field")
	if err == nil {
		t.Errorf("应该返回模块不存在的错误")
	}
}

func TestCustomProvider(t *testing.T) {
	// 创建自定义模块
	customProvider := &CustomProvider{
		name: "custom",
		data: map[string]interface{}{
			"test": map[string]interface{}{
				"value": 42.0,
			},
		},
	}

	// 创建管理器并注册自定义模块
	manager := NewManager()
	manager.RegisterProvider(customProvider)

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

// CustomProvider 自定义模块用于测试
type CustomProvider struct {
	name string
	data map[string]interface{}
}

func (p *CustomProvider) GetName() string {
	return p.name
}

func (p *CustomProvider) GetData(ctx context.Context, dataSource, field string, params ...DataParam) (interface{}, error) {
	entityData, exists := p.data[dataSource]
	if !exists {
		return nil, fmt.Errorf("data source not found: %s", dataSource)
	}

	entityMap, ok := entityData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data source data type")
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

func (p *CustomProvider) GetFunctionParamMapping() map[string]FunctionParamInfo {
	// Custom 模块不需要函数参数
	return map[string]FunctionParamInfo{}
}
