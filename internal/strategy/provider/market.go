package provider

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

// MarketProvider 行情数据模块
type MarketProvider struct {
	*BaseProvider
	market map[string]*MarketData
	mu     sync.RWMutex
}

// NewMarketProvider 创建行情数据模块
func NewMarketProvider() *MarketProvider {
	module := &MarketProvider{
		BaseProvider: NewBaseProvider("market"),
		market:     make(map[string]*MarketData),
	}

	module.initMockData()
	return module
}

// GetData 获取数据
// MarketProvider 只支持单个数据值，不支持历史数据
// params 参数（可选）：
// - 目前暂未使用，保留用于未来扩展
func (p *MarketProvider) GetData(ctx context.Context, entity, field string, params ...DataParam) (interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	marketData, exists := p.market[entity]
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
func (p *MarketProvider) initMockData() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 初始化行情数据
	p.market["SOL"] = &MarketData{
		Symbol:     "SOL",
		LastPx:     205.8,
		LastVolume: 1500000.0,
		Bid:        205.5,
		Ask:        206.0,
		Timestamp:  time.Now(),
	}

	p.market["BTC"] = &MarketData{
		Symbol:     "BTC",
		LastPx:     45500.0,
		LastVolume: 500.0,
		Bid:        45480.0,
		Ask:        45520.0,
		Timestamp:  time.Now(),
	}

	p.market["ETH"] = &MarketData{
		Symbol:     "ETH",
		LastPx:     3250.0,
		LastVolume: 2000.0,
		Bid:        3245.0,
		Ask:        3255.0,
		Timestamp:  time.Now(),
	}
}

// UpdateMarketData 更新行情数据（用于测试）
func (p *MarketProvider) UpdateMarketData(symbol string, data *MarketData) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.market[symbol] = data
}

// GetMarketData 获取行情数据（用于测试）
func (p *MarketProvider) GetMarketData(symbol string) (*MarketData, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data, exists := p.market[symbol]
	return data, exists
}

// GetFunctionParamMapping 获取函数参数映射
func (p *MarketProvider) GetFunctionParamMapping() map[string]FunctionParamInfo {
	// Market 模块目前不需要函数参数
	return map[string]FunctionParamInfo{}
}
