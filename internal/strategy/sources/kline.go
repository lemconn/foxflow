package data

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// KlineData K线数据
type KlineData struct {
	Symbol    string    `json:"symbol"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// KlineModule K线数据模块
type KlineModule struct {
	name   string
	klines map[string]*KlineData
	// 历史数据存储，用于技术分析
	historicalData map[string][]*KlineData
	mu             sync.RWMutex
}

// NewKlineModule 创建K线数据模块
func NewKlineModule() *KlineModule {
	module := &KlineModule{
		name:           "kline",
		klines:         make(map[string]*KlineData),
		historicalData: make(map[string][]*KlineData),
	}

	module.initMockData()
	return module
}

// GetName 获取模块名称
func (m *KlineModule) GetName() string {
	return m.name
}

// GetData 获取数据
func (m *KlineModule) GetData(ctx context.Context, entity, field string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	klineData, exists := m.klines[entity]
	if !exists {
		return nil, fmt.Errorf("no kline data found for entity: %s", entity)
	}

	switch field {
	case "open":
		return klineData.Open, nil
	case "high":
		return klineData.High, nil
	case "low":
		return klineData.Low, nil
	case "close":
		return klineData.Close, nil
	case "volume":
		return klineData.Volume, nil
	case "timestamp":
		return klineData.Timestamp, nil
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}
}

// GetHistoricalData 获取历史数据（内部方法，用于技术分析）
func (m *KlineModule) GetHistoricalData(ctx context.Context, entity, field string, period int) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	historicalData, exists := m.historicalData[entity]
	if !exists {
		// 如果没有历史数据，生成模拟数据
		return m.generateMockHistoricalData(entity, field, period)
	}

	// 获取指定字段的历史数据
	result := make([]interface{}, 0, period)
	count := 0
	for i := len(historicalData) - 1; i >= 0 && count < period; i-- {
		var value interface{}
		switch field {
		case "open":
			value = historicalData[i].Open
		case "high":
			value = historicalData[i].High
		case "low":
			value = historicalData[i].Low
		case "close":
			value = historicalData[i].Close
		case "volume":
			value = historicalData[i].Volume
		default:
			return nil, fmt.Errorf("unknown field: %s", field)
		}
		result = append(result, value)
		count++
	}

	// 如果数据不够，用当前数据填充
	if len(result) < period {
		currentData, err := m.GetData(ctx, entity, field)
		if err != nil {
			return nil, err
		}
		for len(result) < period {
			result = append([]interface{}{currentData}, result...)
		}
	}

	return result, nil
}

// generateMockHistoricalData 生成模拟历史数据
func (m *KlineModule) generateMockHistoricalData(entity, field string, period int) ([]interface{}, error) {
	// 获取当前数据作为基准
	currentData, exists := m.klines[entity]
	if !exists {
		return nil, fmt.Errorf("no kline data found for entity: %s", entity)
	}

	var baseValue float64
	switch field {
	case "open":
		baseValue = currentData.Open
	case "high":
		baseValue = currentData.High
	case "low":
		baseValue = currentData.Low
	case "close":
		baseValue = currentData.Close
	case "volume":
		baseValue = currentData.Volume
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}

	// 生成模拟的历史数据
	historicalData := make([]interface{}, period)
	for i := 0; i < period; i++ {
		// 模拟价格波动
		variation := float64(i-period/2) * 2.0
		historicalData[i] = baseValue + variation
	}

	return historicalData, nil
}

// initMockData 初始化Mock数据
func (m *KlineModule) initMockData() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 初始化K线数据
	m.klines["SOL"] = &KlineData{
		Symbol:    "SOL",
		Open:      195.5,
		High:      210.2,
		Low:       190.1,
		Close:     205.8,
		Volume:    1500000.0,
		Timestamp: time.Now(),
	}

	m.klines["BTC"] = &KlineData{
		Symbol:    "BTC",
		Open:      45000.0,
		High:      46000.0,
		Low:       44000.0,
		Close:     45500.0,
		Volume:    500.0,
		Timestamp: time.Now(),
	}

	m.klines["ETH"] = &KlineData{
		Symbol:    "ETH",
		Open:      3200.0,
		High:      3300.0,
		Low:       3100.0,
		Close:     3250.0,
		Volume:    2000.0,
		Timestamp: time.Now(),
	}
}

// UpdateKlineData 更新K线数据（用于测试）
func (m *KlineModule) UpdateKlineData(symbol string, data *KlineData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.klines[symbol] = data
}

// GetKlineData 获取K线数据（用于测试）
func (m *KlineModule) GetKlineData(symbol string) (*KlineData, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, exists := m.klines[symbol]
	return data, exists
}
