package syntax

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/lemconn/foxflow/internal/engine/registry"
)

// MockDataProvider 模拟数据提供者
type MockDataProvider struct {
	klineData  map[string]map[string][]float64
	marketData map[string]map[string]float64
	newsData   map[string]map[string]interface{}
}

// MockKlineDataSource 模拟K线数据源
type MockKlineDataSource struct {
	provider *MockDataProvider
}

func (m *MockKlineDataSource) GetName() string {
	return "kline"
}

func (m *MockKlineDataSource) GetData(ctx context.Context, dataSource, field string, params ...interface{}) (interface{}, error) {
	// 如果请求历史数据
	if len(params) > 0 {
		var limit int
		var interval string = "1m" // 默认值
		
		if len(params) > 0 {
			if l, ok := params[0].(int); ok && l > 0 {
				limit = l
			}
		}
		if len(params) > 1 {
			if i, ok := params[1].(string); ok {
				interval = i
			}
		}
		
		if limit > 0 {
			return m.provider.GetHistoricalData(ctx, "kline", dataSource, field, limit, interval)
		}
	}
	
	return nil, fmt.Errorf("no data found")
}

// MockMarketDataSource 模拟行情数据源
type MockMarketDataSource struct {
	provider *MockDataProvider
}

func (m *MockMarketDataSource) GetName() string {
	return "market"
}

func (m *MockMarketDataSource) GetData(ctx context.Context, dataSource, field string, params ...interface{}) (interface{}, error) {
	return m.provider.GetMarketData(ctx, "market", dataSource, field)
}

// MockNewsDataSource 模拟新闻数据源
type MockNewsDataSource struct {
	provider *MockDataProvider
}

func (m *MockNewsDataSource) GetName() string {
	return "news"
}

func (m *MockNewsDataSource) GetData(ctx context.Context, dataSource, field string, params ...interface{}) (interface{}, error) {
	return m.provider.GetNewsData(ctx, "news", dataSource, field)
}

// MockDataProvider 方法实现
func (m *MockDataProvider) GetHistoricalData(ctx context.Context, module, dataSource, field string, limit int, interval string) (interface{}, error) {
	// 解析字段
	parts := strings.Split(field, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid field format")
	}
	
	symbol := parts[0]
	fieldName := parts[1]
	
	// 获取数据
	if symbolData, exists := m.klineData[symbol]; exists {
		if intervalData, exists := symbolData[interval]; exists {
			// 返回指定数量的数据点
			if len(intervalData) >= limit {
				return intervalData[len(intervalData)-limit:], nil
			}
			return intervalData, nil
		}
	}
	
	return nil, fmt.Errorf("no data found for %s %s", symbol, fieldName)
}

func (m *MockDataProvider) GetMarketData(ctx context.Context, module, dataSource, field string) (interface{}, error) {
	// 解析字段
	parts := strings.Split(field, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid field format")
	}
	
	symbol := parts[0]
	fieldName := parts[1]
	
	// 获取数据
	if symbolData, exists := m.marketData[symbol]; exists {
		if value, exists := symbolData[fieldName]; exists {
			return value, nil
		}
	}
	
	return nil, fmt.Errorf("no data found for %s %s", symbol, fieldName)
}

func (m *MockDataProvider) GetNewsData(ctx context.Context, module, dataSource, field string) (interface{}, error) {
	// 获取数据
	if sourceData, exists := m.newsData[dataSource]; exists {
		if value, exists := sourceData[field]; exists {
			return value, nil
		}
	}
	
	return nil, fmt.Errorf("no data found for %s %s", dataSource, field)
}

func TestEvaluatorBasicOperations(t *testing.T) {
	// 创建模拟数据提供者
	mockProvider := &MockDataProvider{
		klineData: map[string]map[string][]float64{
			"BTC": {
				"1m":  {100.0, 101.0, 102.0, 103.0, 104.0},
				"15m": {100.0, 102.0, 104.0, 106.0, 108.0},
			},
		},
		marketData: map[string]map[string]float64{
			"BTC": {
				"last_px": 105.0,
				"bid":     104.5,
				"ask":     105.5,
			},
		},
		newsData: map[string]map[string]interface{}{
			"blockbeats": {
				"title":   "Test News Title",
				"content": "Test News Content",
			},
		},
	}
	
	// 创建注册器并注册模拟数据源
	registry := registry.NewRegistry()
	registry.RegisterProvider(&MockKlineDataSource{provider: mockProvider})
	registry.RegisterProvider(&MockMarketDataSource{provider: mockProvider})
	registry.RegisterProvider(&MockNewsDataSource{provider: mockProvider})
	
	// 创建求值器
	evaluator := NewEvaluator(registry)
	ctx := context.Background()
	
	// 测试基本字段访问
	value, err := evaluator.GetFieldValue(ctx, "market", "okx", "BTC.last_px")
	if err != nil {
		t.Errorf("获取行情数据失败: %v", err)
	}
	
	if value != 105.0 {
		t.Errorf("期望行情数据 105.0，实际得到 %v", value)
	}
	
	// 测试带参数的字段访问
	historicalData, err := evaluator.GetFieldValueWithParams(ctx, "kline", "okx", "BTC.close", 3, "15m")
	if err != nil {
		t.Errorf("获取历史数据失败: %v", err)
	}
	
	if dataArray, ok := historicalData.([]float64); ok {
		if len(dataArray) != 3 {
			t.Errorf("期望3个数据点，实际得到%d个", len(dataArray))
		}
	} else {
		t.Errorf("期望[]float64类型，实际得到%T", historicalData)
	}
	
	// 测试新闻数据访问
	newsTitle, err := evaluator.GetFieldValue(ctx, "news", "blockbeats", "title")
	if err != nil {
		t.Errorf("获取新闻标题失败: %v", err)
	}
	
	if newsTitle != "Test News Title" {
		t.Errorf("期望新闻标题 'Test News Title'，实际得到 %v", newsTitle)
	}
}

func TestEvaluatorComparison(t *testing.T) {
	// 创建模拟数据提供者
	mockProvider := &MockDataProvider{
		marketData: map[string]map[string]float64{
			"BTC": {
				"last_px": 105.0,
			},
		},
	}
	
	// 创建注册器并注册模拟数据源
	registry := registry.NewRegistry()
	registry.RegisterProvider(&MockMarketDataSource{provider: mockProvider})
	
	// 创建求值器
	evaluator := NewEvaluator(registry)
	
	// 测试比较操作
	result, err := evaluator.EvaluateBinary(">", 105.0, 100.0)
	if err != nil {
		t.Errorf("比较操作失败: %v", err)
	}
	
	if result != true {
		t.Error("期望 105.0 > 100.0 为 true")
	}
	
	// 测试相等性比较
	result, err = evaluator.EvaluateBinary("==", 105.0, 105.0)
	if err != nil {
		t.Errorf("相等性比较失败: %v", err)
	}
	
	if result != true {
		t.Error("期望 105.0 == 105.0 为 true")
	}
}

func TestEvaluatorLogicalOperations(t *testing.T) {
	// 创建求值器
	registry := registry.NewRegistry()
	evaluator := NewEvaluator(registry)
	
	// 测试逻辑AND操作
	result, err := evaluator.EvaluateBinary("and", true, true)
	if err != nil {
		t.Errorf("逻辑AND操作失败: %v", err)
	}
	
	if result != true {
		t.Error("期望 true AND true 为 true")
	}
	
	// 测试逻辑OR操作
	result, err = evaluator.EvaluateBinary("or", false, true)
	if err != nil {
		t.Errorf("逻辑OR操作失败: %v", err)
	}
	
	if result != true {
		t.Error("期望 false OR true 为 true")
	}
}