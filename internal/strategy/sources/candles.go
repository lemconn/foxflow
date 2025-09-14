package data

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CandlesModule K线数据模块
type CandlesModule struct {
	name    string
	candles map[string]*CandlesData
	mu      sync.RWMutex
}

// NewCandlesModule 创建K线数据模块
func NewCandlesModule() *CandlesModule {
	module := &CandlesModule{
		name:    "candles",
		candles: make(map[string]*CandlesData),
	}

	module.initMockData()
	return module
}

// GetName 获取模块名称
func (m *CandlesModule) GetName() string {
	return m.name
}

// GetData 获取数据
func (m *CandlesModule) GetData(ctx context.Context, entity, field string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	candleData, exists := m.candles[entity]
	if !exists {
		return nil, fmt.Errorf("no candle data found for entity: %s", entity)
	}

	switch field {
	case "last_px":
		return candleData.LastPx, nil
	case "last_volume":
		return candleData.LastVolume, nil
	case "open":
		return candleData.Open, nil
	case "high":
		return candleData.High, nil
	case "low":
		return candleData.Low, nil
	case "close":
		return candleData.Close, nil
	case "volume":
		return candleData.Volume, nil
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}
}

// GetHistoricalData 获取历史数据
func (m *CandlesModule) GetHistoricalData(ctx context.Context, entity, field string, period int) ([]interface{}, error) {
	// 获取当前数据作为基准
	currentData, err := m.GetData(ctx, entity, field)
	if err != nil {
		return nil, err
	}

	baseValue, ok := currentData.(float64)
	if !ok {
		return nil, fmt.Errorf("field %s is not a numeric value", field)
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
func (m *CandlesModule) initMockData() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 初始化K线数据
	m.candles["SOL"] = &CandlesData{
		Symbol:     "SOL",
		Open:       195.5,
		High:       210.2,
		Low:        190.1,
		Close:      205.8,
		Volume:     1500000.0,
		Timestamp:  time.Now(),
		LastPx:     205.8,
		LastVolume: 1500000.0,
	}

	m.candles["BTC"] = &CandlesData{
		Symbol:     "BTC",
		Open:       45000.0,
		High:       46000.0,
		Low:        44000.0,
		Close:      45500.0,
		Volume:     500.0,
		Timestamp:  time.Now(),
		LastPx:     45500.0,
		LastVolume: 500.0,
	}

	m.candles["ETH"] = &CandlesData{
		Symbol:     "ETH",
		Open:       3200.0,
		High:       3300.0,
		Low:        3100.0,
		Close:      3250.0,
		Volume:     2000.0,
		Timestamp:  time.Now(),
		LastPx:     3250.0,
		LastVolume: 2000.0,
	}
}

// UpdateCandlesData 更新K线数据（用于测试）
func (m *CandlesModule) UpdateCandlesData(symbol string, data *CandlesData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.candles[symbol] = data
}

// GetCandlesData 获取K线数据（用于测试）
func (m *CandlesModule) GetCandlesData(symbol string) (*CandlesData, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, exists := m.candles[symbol]
	return data, exists
}
