package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// IndicatorsData 指标数据
type IndicatorsData struct {
	Symbol    string                 `json:"symbol"`
	Indicator string                 `json:"indicator"` // 指标名称：MACD, RSI, Volume等
	Value     float64                `json:"value"`     // 指标值
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"` // 额外元数据
}

// IndicatorsProvider 指标数据模块
type IndicatorsProvider struct {
	*BaseProvider
	indicators map[string]*IndicatorsData
	mu         sync.RWMutex
}

// NewIndicatorsProvider 创建指标数据模块
func NewIndicatorsProvider() *IndicatorsProvider {
	module := &IndicatorsProvider{
		BaseProvider: NewBaseProvider("indicators"),
		indicators: make(map[string]*IndicatorsData),
	}

	module.initMockData()
	return module
}

// GetData 获取数据
// IndicatorsProvider 只支持单个数据值，不支持历史数据
// params 参数（可选）：
// - 目前暂未使用，保留用于未来扩展
func (p *IndicatorsProvider) GetData(ctx context.Context, entity, field string, params ...DataParam) (interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 构建指标键：symbol-indicator
	key := fmt.Sprintf("%s-%s", entity, field)
	indicatorsData, exists := p.indicators[key]
	if !exists {
		return nil, fmt.Errorf("no indicator data found for %s %s", entity, field)
	}

	return indicatorsData.Value, nil
}

// initMockData 初始化Mock数据
func (p *IndicatorsProvider) initMockData() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 初始化指标数据
	p.indicators["SOL-MACD"] = &IndicatorsData{
		Symbol:    "SOL",
		Indicator: "MACD",
		Value:     0.5,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"signal":    0.3,
			"histogram": 0.2,
		},
	}

	p.indicators["BTC-RSI"] = &IndicatorsData{
		Symbol:    "BTC",
		Indicator: "RSI",
		Value:     65.5,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"overbought": false,
			"oversold":   false,
		},
	}

	p.indicators["ETH-Volume"] = &IndicatorsData{
		Symbol:    "ETH",
		Indicator: "Volume",
		Value:     2500.0,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"avg_volume":   2000.0,
			"volume_ratio": 1.25,
		},
	}
}

// UpdateIndicatorsData 更新指标数据（用于测试）
func (p *IndicatorsProvider) UpdateIndicatorsData(symbol, indicator string, data *IndicatorsData) {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := fmt.Sprintf("%s-%s", symbol, indicator)
	p.indicators[key] = data
}

// GetIndicatorsData 获取指标数据（用于测试）
func (p *IndicatorsProvider) GetIndicatorsData(symbol, indicator string) (*IndicatorsData, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	key := fmt.Sprintf("%s-%s", symbol, indicator)
	data, exists := p.indicators[key]
	return data, exists
}

// GetFunctionParamMapping 获取函数参数映射
func (p *IndicatorsProvider) GetFunctionParamMapping() map[string]FunctionParamInfo {
	// Indicators 模块目前不需要函数参数
	return map[string]FunctionParamInfo{}
}
