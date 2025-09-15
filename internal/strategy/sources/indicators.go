package data

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

// IndicatorsModule 指标数据模块
type IndicatorsModule struct {
	name       string
	indicators map[string]*IndicatorsData
	mu         sync.RWMutex
}

// NewIndicatorsModule 创建指标数据模块
func NewIndicatorsModule() *IndicatorsModule {
	module := &IndicatorsModule{
		name:       "indicators",
		indicators: make(map[string]*IndicatorsData),
	}

	module.initMockData()
	return module
}

// GetName 获取模块名称
func (m *IndicatorsModule) GetName() string {
	return m.name
}

// GetData 获取数据
func (m *IndicatorsModule) GetData(ctx context.Context, entity, field string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 构建指标键：symbol-indicator
	key := fmt.Sprintf("%s-%s", entity, field)
	indicatorsData, exists := m.indicators[key]
	if !exists {
		return nil, fmt.Errorf("no indicator data found for %s %s", entity, field)
	}

	return indicatorsData.Value, nil
}

// initMockData 初始化Mock数据
func (m *IndicatorsModule) initMockData() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 初始化指标数据
	m.indicators["SOL-MACD"] = &IndicatorsData{
		Symbol:    "SOL",
		Indicator: "MACD",
		Value:     0.5,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"signal":    0.3,
			"histogram": 0.2,
		},
	}

	m.indicators["BTC-RSI"] = &IndicatorsData{
		Symbol:    "BTC",
		Indicator: "RSI",
		Value:     65.5,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"overbought": false,
			"oversold":   false,
		},
	}

	m.indicators["ETH-Volume"] = &IndicatorsData{
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
func (m *IndicatorsModule) UpdateIndicatorsData(symbol, indicator string, data *IndicatorsData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s-%s", symbol, indicator)
	m.indicators[key] = data
}

// GetIndicatorsData 获取指标数据（用于测试）
func (m *IndicatorsModule) GetIndicatorsData(symbol, indicator string) (*IndicatorsData, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s-%s", symbol, indicator)
	data, exists := m.indicators[key]
	return data, exists
}
