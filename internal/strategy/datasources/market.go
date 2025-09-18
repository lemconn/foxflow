package datasources

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MarketData 行情数据
type MarketData struct {
	Symbol     string    `json:"symbol"`
	LastPx     float64   `json:"last_px"`     // 最新价格
	LastVolume float64   `json:"last_volume"` // 最新成交量
	Bid        float64   `json:"bid"`         // 买一价
	Ask        float64   `json:"ask"`         // 卖一价
	Timestamp  time.Time `json:"timestamp"`
}

// MarketModule 行情数据模块
type MarketModule struct {
	*BaseModule
	market map[string]*MarketData
	mu     sync.RWMutex
}

// NewMarketModule 创建行情数据模块
func NewMarketModule() *MarketModule {
	module := &MarketModule{
		BaseModule: NewBaseModule("market"),
		market:     make(map[string]*MarketData),
	}

	module.initMockData()
	return module
}

// GetData 获取数据
// MarketModule 只支持单个数据值，不支持历史数据
// params 参数（可选）：
// - 目前暂未使用，保留用于未来扩展
func (m *MarketModule) GetData(ctx context.Context, entity, field string, params ...DataParam) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	marketData, exists := m.market[entity]
	if !exists {
		return nil, fmt.Errorf("no market data found for entity: %s", entity)
	}

	switch field {
	case "last_px":
		return marketData.LastPx, nil
	case "last_volume":
		return marketData.LastVolume, nil
	case "bid":
		return marketData.Bid, nil
	case "ask":
		return marketData.Ask, nil
	case "timestamp":
		return marketData.Timestamp, nil
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}
}

// initMockData 初始化Mock数据
func (m *MarketModule) initMockData() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 初始化行情数据
	m.market["SOL"] = &MarketData{
		Symbol:     "SOL",
		LastPx:     205.8,
		LastVolume: 1500000.0,
		Bid:        205.5,
		Ask:        206.0,
		Timestamp:  time.Now(),
	}

	m.market["BTC"] = &MarketData{
		Symbol:     "BTC",
		LastPx:     45500.0,
		LastVolume: 500.0,
		Bid:        45480.0,
		Ask:        45520.0,
		Timestamp:  time.Now(),
	}

	m.market["ETH"] = &MarketData{
		Symbol:     "ETH",
		LastPx:     3250.0,
		LastVolume: 2000.0,
		Bid:        3245.0,
		Ask:        3255.0,
		Timestamp:  time.Now(),
	}
}

// UpdateMarketData 更新行情数据（用于测试）
func (m *MarketModule) UpdateMarketData(symbol string, data *MarketData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.market[symbol] = data
}

// GetMarketData 获取行情数据（用于测试）
func (m *MarketModule) GetMarketData(symbol string) (*MarketData, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, exists := m.market[symbol]
	return data, exists
}
