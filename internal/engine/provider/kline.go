package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// KlineData K线数据
type KlineData struct {
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// ExchangeData 交易所数据结构
type ExchangeData map[string]map[string][]KlineData

// KlineProvider K线数据提供者
type KlineProvider struct {
	*BaseProvider
	klines map[string]ExchangeData
	mu     sync.RWMutex
}

// NewKlineProvider 创建K线数据提供者
func NewKlineProvider() *KlineProvider {
	provider := &KlineProvider{
		BaseProvider: NewBaseProvider("kline"),
		klines:       make(map[string]ExchangeData),
	}

	provider.initMockData()
	return provider
}

// GetData 获取数据
// KlineProvider 只支持历史数据数组，不支持单个数据值
// params 支持的参数：
// - params[0]: string - 时间间隔（可选，如 "15m", "1h", "1d"）
// - params[1]: int - 历史数据周期数（必需）
// - params[2]: time.Time - 开始时间（可选）
// - params[3]: time.Time - 结束时间（可选）
func (p *KlineProvider) GetData(ctx context.Context, dataSource, field string, params ...interface{}) (interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()


	// 获取可选参数 interval (params[0])
	interval := "1m" // 默认1分钟
	if len(params) > 0 {
		if intervalParam, ok := params[0].(string); ok {
			interval = intervalParam
		}
	}

	// 检查是否提供了 limit 参数
	if len(params) < 2 {
		return nil, fmt.Errorf("kline provider requires limit parameter")
	}

	// 获取必需参数 limit (params[1])
	var limit int
	switch v := params[1].(type) {
	case int:
		limit = v
	case float64:
		limit = int(v)
	case int64:
		limit = int(v)
	default:
		return nil, fmt.Errorf("kline provider requires limit parameter to be int, got %T", params[1])
	}

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0, got %d", limit)
	}

	// 返回历史数据数组
	return p.getHistoricalDataArray(dataSource, field, limit, interval)
}

// getHistoricalDataArray 获取历史数据数组
func (p *KlineProvider) getHistoricalDataArray(dataSource, field string, limit int, interval string) ([]interface{}, error) {
	// 解析字段 - 支持多级字段如 "BTC.close"
	fieldParts := strings.Split(field, ".")
	if len(fieldParts) < 2 {
		return nil, fmt.Errorf("kline field must be in format 'SYMBOL.FIELD', got: %s", field)
	}

	symbol := fieldParts[0]
	fieldName := fieldParts[1]

	// 获取交易所数据
	exchangeData, exists := p.klines[dataSource]
	if !exists {
		return nil, fmt.Errorf("no exchange data found for data source: %s", dataSource)
	}

	// 获取币种数据
	symbolData, exists := exchangeData[symbol]
	if !exists {
		return nil, fmt.Errorf("no symbol data found for symbol: %s", symbol)
	}

	// 获取时间间隔数据
	intervalData, exists := symbolData[interval]
	if !exists {
		return nil, fmt.Errorf("no interval data found for interval: %s", interval)
	}

	// 获取指定字段的历史数据
	result := make([]interface{}, 0, limit)
	count := 0
	
	// 从最新的数据开始取
	for i := len(intervalData) - 1; i >= 0 && count < limit; i-- {
		var value interface{}
		switch fieldName {
		case "open":
			value = intervalData[i].Open
		case "high":
			value = intervalData[i].High
		case "low":
			value = intervalData[i].Low
		case "close":
			value = intervalData[i].Close
		case "volume":
			value = intervalData[i].Volume
		default:
			return nil, fmt.Errorf("unknown field: %s", fieldName)
		}
		result = append(result, value)
		count++
	}

	// 如果数据不够，用最新的数据填充
	if len(result) < limit && len(intervalData) > 0 {
		latestData := intervalData[len(intervalData)-1]
		var currentValue interface{}
		switch fieldName {
		case "open":
			currentValue = latestData.Open
		case "high":
			currentValue = latestData.High
		case "low":
			currentValue = latestData.Low
		case "close":
			currentValue = latestData.Close
		case "volume":
			currentValue = latestData.Volume
		default:
			return nil, fmt.Errorf("unknown field: %s", fieldName)
		}

		for len(result) < limit {
			result = append([]interface{}{currentValue}, result...)
		}
	}

	return result, nil
}

// initMockData 初始化Mock数据
func (p *KlineProvider) initMockData() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	
	// 初始化OKX交易所数据
	p.klines["okx"] = ExchangeData{
		"BTC": {
			"1m": []KlineData{
				{Open: 45000.0, High: 45050.0, Low: 44980.0, Close: 45020.0, Volume: 12.5, Timestamp: now.Add(-4 * time.Minute)},
				{Open: 45020.0, High: 45080.0, Low: 45010.0, Close: 45060.0, Volume: 15.2, Timestamp: now.Add(-3 * time.Minute)},
				{Open: 45060.0, High: 45100.0, Low: 45040.0, Close: 45080.0, Volume: 18.7, Timestamp: now.Add(-2 * time.Minute)},
				{Open: 45080.0, High: 45120.0, Low: 45060.0, Close: 45100.0, Volume: 14.3, Timestamp: now.Add(-1 * time.Minute)},
				{Open: 45100.0, High: 45150.0, Low: 45080.0, Close: 45130.0, Volume: 16.8, Timestamp: now},
			},
			"5m": []KlineData{
				{Open: 44800.0, High: 45100.0, Low: 44750.0, Close: 45050.0, Volume: 120.5, Timestamp: now.Add(-20 * time.Minute)},
				{Open: 45050.0, High: 45200.0, Low: 45000.0, Close: 45150.0, Volume: 95.3, Timestamp: now.Add(-15 * time.Minute)},
				{Open: 45150.0, High: 45250.0, Low: 45100.0, Close: 45200.0, Volume: 110.7, Timestamp: now.Add(-10 * time.Minute)},
				{Open: 45200.0, High: 45280.0, Low: 45150.0, Close: 45250.0, Volume: 88.2, Timestamp: now.Add(-5 * time.Minute)},
				{Open: 45250.0, High: 45300.0, Low: 45200.0, Close: 45280.0, Volume: 102.1, Timestamp: now},
			},
			"15m": []KlineData{
				{Open: 44500.0, High: 45000.0, Low: 44400.0, Close: 44800.0, Volume: 350.0, Timestamp: now.Add(-60 * time.Minute)},
				{Open: 44800.0, High: 45200.0, Low: 44700.0, Close: 45000.0, Volume: 420.5, Timestamp: now.Add(-45 * time.Minute)},
				{Open: 45000.0, High: 45300.0, Low: 44900.0, Close: 45150.0, Volume: 380.2, Timestamp: now.Add(-30 * time.Minute)},
				{Open: 45150.0, High: 45400.0, Low: 45050.0, Close: 45300.0, Volume: 450.8, Timestamp: now.Add(-15 * time.Minute)},
				{Open: 45300.0, High: 45450.0, Low: 45200.0, Close: 45350.0, Volume: 320.6, Timestamp: now},
			},
			"1h": []KlineData{
				{Open: 44000.0, High: 45000.0, Low: 43800.0, Close: 44500.0, Volume: 1200.0, Timestamp: now.Add(-4 * time.Hour)},
				{Open: 44500.0, High: 45500.0, Low: 44300.0, Close: 45000.0, Volume: 1350.5, Timestamp: now.Add(-3 * time.Hour)},
				{Open: 45000.0, High: 45800.0, Low: 44800.0, Close: 45500.0, Volume: 1420.8, Timestamp: now.Add(-2 * time.Hour)},
				{Open: 45500.0, High: 46000.0, Low: 45300.0, Close: 45700.0, Volume: 1280.3, Timestamp: now.Add(-1 * time.Hour)},
				{Open: 45700.0, High: 46000.0, Low: 45500.0, Close: 45800.0, Volume: 1100.7, Timestamp: now},
			},
		},
		"ETH": {
			"1m": []KlineData{
				{Open: 3200.0, High: 3210.0, Low: 3195.0, Close: 3205.0, Volume: 45.2, Timestamp: now.Add(-4 * time.Minute)},
				{Open: 3205.0, High: 3220.0, Low: 3200.0, Close: 3215.0, Volume: 52.8, Timestamp: now.Add(-3 * time.Minute)},
				{Open: 3215.0, High: 3230.0, Low: 3210.0, Close: 3225.0, Volume: 48.6, Timestamp: now.Add(-2 * time.Minute)},
				{Open: 3225.0, High: 3240.0, Low: 3220.0, Close: 3235.0, Volume: 55.3, Timestamp: now.Add(-1 * time.Minute)},
				{Open: 3235.0, High: 3250.0, Low: 3230.0, Close: 3245.0, Volume: 61.7, Timestamp: now},
			},
			"5m": []KlineData{
				{Open: 3150.0, High: 3200.0, Low: 3140.0, Close: 3180.0, Volume: 280.5, Timestamp: now.Add(-20 * time.Minute)},
				{Open: 3180.0, High: 3220.0, Low: 3170.0, Close: 3200.0, Volume: 320.8, Timestamp: now.Add(-15 * time.Minute)},
				{Open: 3200.0, High: 3240.0, Low: 3190.0, Close: 3220.0, Volume: 350.2, Timestamp: now.Add(-10 * time.Minute)},
				{Open: 3220.0, High: 3260.0, Low: 3210.0, Close: 3240.0, Volume: 380.6, Timestamp: now.Add(-5 * time.Minute)},
				{Open: 3240.0, High: 3280.0, Low: 3230.0, Close: 3260.0, Volume: 420.3, Timestamp: now},
			},
			"15m": []KlineData{
				{Open: 3100.0, High: 3200.0, Low: 3080.0, Close: 3150.0, Volume: 850.0, Timestamp: now.Add(-60 * time.Minute)},
				{Open: 3150.0, High: 3250.0, Low: 3130.0, Close: 3200.0, Volume: 920.5, Timestamp: now.Add(-45 * time.Minute)},
				{Open: 3200.0, High: 3300.0, Low: 3180.0, Close: 3250.0, Volume: 980.2, Timestamp: now.Add(-30 * time.Minute)},
				{Open: 3250.0, High: 3350.0, Low: 3230.0, Close: 3300.0, Volume: 1050.8, Timestamp: now.Add(-15 * time.Minute)},
				{Open: 3300.0, High: 3380.0, Low: 3280.0, Close: 3350.0, Volume: 1120.6, Timestamp: now},
			},
			"1h": []KlineData{
				{Open: 3000.0, High: 3200.0, Low: 2980.0, Close: 3100.0, Volume: 3500.0, Timestamp: now.Add(-4 * time.Hour)},
				{Open: 3100.0, High: 3300.0, Low: 3080.0, Close: 3200.0, Volume: 3800.5, Timestamp: now.Add(-3 * time.Hour)},
				{Open: 3200.0, High: 3400.0, Low: 3180.0, Close: 3300.0, Volume: 4200.8, Timestamp: now.Add(-2 * time.Hour)},
				{Open: 3300.0, High: 3500.0, Low: 3280.0, Close: 3400.0, Volume: 4500.3, Timestamp: now.Add(-1 * time.Hour)},
				{Open: 3400.0, High: 3600.0, Low: 3380.0, Close: 3500.0, Volume: 4800.7, Timestamp: now},
			},
		},
		"SOL": {
			"1m": []KlineData{
				{Open: 180.0, High: 182.0, Low: 179.5, Close: 181.0, Volume: 125.8, Timestamp: now.Add(-4 * time.Minute)},
				{Open: 181.0, High: 183.5, Low: 180.5, Close: 182.5, Volume: 142.3, Timestamp: now.Add(-3 * time.Minute)},
				{Open: 182.5, High: 184.0, Low: 182.0, Close: 183.5, Volume: 138.6, Timestamp: now.Add(-2 * time.Minute)},
				{Open: 183.5, High: 185.0, Low: 183.0, Close: 184.5, Volume: 155.2, Timestamp: now.Add(-1 * time.Minute)},
				{Open: 184.5, High: 186.0, Low: 184.0, Close: 185.5, Volume: 168.7, Timestamp: now},
			},
			"5m": []KlineData{
				{Open: 175.0, High: 180.0, Low: 174.0, Close: 178.0, Volume: 680.5, Timestamp: now.Add(-20 * time.Minute)},
				{Open: 178.0, High: 182.0, Low: 177.0, Close: 180.0, Volume: 720.8, Timestamp: now.Add(-15 * time.Minute)},
				{Open: 180.0, High: 184.0, Low: 179.0, Close: 182.0, Volume: 750.2, Timestamp: now.Add(-10 * time.Minute)},
				{Open: 182.0, High: 186.0, Low: 181.0, Close: 184.0, Volume: 780.6, Timestamp: now.Add(-5 * time.Minute)},
				{Open: 184.0, High: 188.0, Low: 183.0, Close: 186.0, Volume: 820.3, Timestamp: now},
			},
			"15m": []KlineData{
				{Open: 170.0, High: 180.0, Low: 168.0, Close: 175.0, Volume: 1850.0, Timestamp: now.Add(-60 * time.Minute)},
				{Open: 175.0, High: 185.0, Low: 173.0, Close: 180.0, Volume: 1920.5, Timestamp: now.Add(-45 * time.Minute)},
				{Open: 180.0, High: 190.0, Low: 178.0, Close: 185.0, Volume: 1980.2, Timestamp: now.Add(-30 * time.Minute)},
				{Open: 185.0, High: 195.0, Low: 183.0, Close: 190.0, Volume: 2050.8, Timestamp: now.Add(-15 * time.Minute)},
				{Open: 190.0, High: 200.0, Low: 188.0, Close: 195.0, Volume: 2120.6, Timestamp: now},
			},
			"1h": []KlineData{
				{Open: 160.0, High: 180.0, Low: 158.0, Close: 170.0, Volume: 8500.0, Timestamp: now.Add(-4 * time.Hour)},
				{Open: 170.0, High: 190.0, Low: 168.0, Close: 180.0, Volume: 8800.5, Timestamp: now.Add(-3 * time.Hour)},
				{Open: 180.0, High: 200.0, Low: 178.0, Close: 190.0, Volume: 9200.8, Timestamp: now.Add(-2 * time.Hour)},
				{Open: 190.0, High: 210.0, Low: 188.0, Close: 200.0, Volume: 9500.3, Timestamp: now.Add(-1 * time.Hour)},
				{Open: 200.0, High: 220.0, Low: 198.0, Close: 210.0, Volume: 9800.7, Timestamp: now},
			},
		},
	}
}


