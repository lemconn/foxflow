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
	Symbol    string    `json:"symbol"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// KlineProvider K线数据提供者
type KlineProvider struct {
	*BaseProvider
	klines map[string]*KlineData
	// 历史数据存储，用于技术分析
	historicalData map[string][]*KlineData
	mu             sync.RWMutex
}

// NewKlineProvider 创建K线数据提供者
func NewKlineProvider() *KlineProvider {
	provider := &KlineProvider{
		BaseProvider:   NewBaseProvider("kline"),
		klines:         make(map[string]*KlineData),
		historicalData: make(map[string][]*KlineData),
	}

	provider.initMockData()
	return provider
}

// GetData 获取数据
// KlineProvider 只支持历史数据数组，不支持单个数据值
// params 支持的参数：
// - limit: int - 历史数据周期数（必需）
// - interval: string - 时间间隔（可选，如 "15m", "1h", "1d"）
// - start_time: time.Time - 开始时间（可选）
// - end_time: time.Time - 结束时间（可选）
func (p *KlineProvider) GetData(ctx context.Context, dataSource, field string, params ...DataParam) (interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 检查是否提供了参数
	if len(params) == 0 {
		return nil, fmt.Errorf("kline provider requires DataParam with 'limit' field")
	}

	// 获取必需参数 limit
	limit, err := p.GetIntParam(params, "limit")
	if err != nil {
		return nil, fmt.Errorf("kline provider requires 'limit' parameter: %w", err)
	}

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0, got %d", limit)
	}

	// 获取可选参数 interval
	interval := "1m" // 默认1分钟
	if intervalParam, err := p.GetParam(params, "interval", "string"); err == nil {
		interval = intervalParam.(string)
	}

	// 返回历史数据数组
	return p.getHistoricalDataArray(dataSource, field, limit, interval)
}

// getHistoricalDataArray 获取历史数据数组
func (p *KlineProvider) getHistoricalDataArray(dataSource, field string, limit int, interval string) ([]interface{}, error) {
	historicalData, exists := p.historicalData[dataSource]
	if !exists {
		// 如果没有历史数据，生成模拟数据
		return p.generateMockHistoricalData(dataSource, field, limit, interval)
	}

	// 解析字段 - 支持多级字段如 "BTC.close"
	fieldParts := strings.Split(field, ".")
	if len(fieldParts) < 2 {
		return nil, fmt.Errorf("kline field must be in format 'SYMBOL.FIELD', got: %s", field)
	}

	symbol := fieldParts[0]
	fieldName := fieldParts[1]

	// 验证符号是否匹配
	if historicalData[0].Symbol != symbol {
		return nil, fmt.Errorf("symbol mismatch: expected %s, got %s", symbol, historicalData[0].Symbol)
	}

	// 获取指定字段的历史数据
	result := make([]interface{}, 0, limit)
	count := 0
	for i := len(historicalData) - 1; i >= 0 && count < limit; i-- {
		var value interface{}
		switch fieldName {
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
			return nil, fmt.Errorf("unknown field: %s", fieldName)
		}
		result = append(result, value)
		count++
	}

	// 如果数据不够，用当前数据填充
	if len(result) < limit {
		klineData, exists := p.klines[dataSource]
		if !exists {
			return nil, fmt.Errorf("no kline data found for data source: %s", dataSource)
		}

		var currentValue interface{}
		switch fieldName {
		case "open":
			currentValue = klineData.Open
		case "high":
			currentValue = klineData.High
		case "low":
			currentValue = klineData.Low
		case "close":
			currentValue = klineData.Close
		case "volume":
			currentValue = klineData.Volume
		default:
			return nil, fmt.Errorf("unknown field: %s", fieldName)
		}

		for len(result) < limit {
			result = append([]interface{}{currentValue}, result...)
		}
	}

	return result, nil
}

// generateMockHistoricalData 生成模拟历史数据
func (p *KlineProvider) generateMockHistoricalData(dataSource, field string, limit int, interval string) ([]interface{}, error) {
	// 获取当前数据作为基准
	currentData, exists := p.klines[dataSource]
	if !exists {
		return nil, fmt.Errorf("no kline data found for data source: %s", dataSource)
	}

	// 解析字段
	fieldParts := strings.Split(field, ".")
	if len(fieldParts) < 2 {
		return nil, fmt.Errorf("kline field must be in format 'SYMBOL.FIELD', got: %s", field)
	}

	symbol := fieldParts[0]
	fieldName := fieldParts[1]

	// 验证符号是否匹配
	if currentData.Symbol != symbol {
		return nil, fmt.Errorf("symbol mismatch: expected %s, got %s", symbol, currentData.Symbol)
	}

	var baseValue float64
	switch fieldName {
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
		return nil, fmt.Errorf("unknown field: %s", fieldName)
	}

	// 根据时间间隔生成模拟的历史数据
	historicalData := make([]interface{}, limit)
	
	for i := 0; i < limit; i++ {
		// 根据时间间隔调整价格波动幅度
		// 时间间隔越长，波动幅度越大
		volatility := p.getVolatilityByInterval(interval)
		variation := float64(i-limit/2) * volatility
		
		// 添加一些随机性，使数据更真实
		randomFactor := float64((i*7)%10-5) * 0.1 // 简单的伪随机
		historicalData[i] = baseValue + variation + randomFactor
	}

	return historicalData, nil
}

// parseInterval 解析时间间隔字符串，返回分钟数
func (p *KlineProvider) parseInterval(interval string) int {
	switch interval {
	case "1m":
		return 1
	case "5m":
		return 5
	case "15m":
		return 15
	case "30m":
		return 30
	case "1h":
		return 60
	case "4h":
		return 240
	case "1d":
		return 1440
	case "1w":
		return 10080
	default:
		return 1 // 默认1分钟
	}
}

// getVolatilityByInterval 根据时间间隔获取波动率
func (p *KlineProvider) getVolatilityByInterval(interval string) float64 {
	switch interval {
	case "1m":
		return 0.5 // 1分钟波动较小
	case "5m":
		return 1.0
	case "15m":
		return 2.0
	case "30m":
		return 3.0
	case "1h":
		return 5.0
	case "4h":
		return 10.0
	case "1d":
		return 20.0
	case "1w":
		return 50.0
	default:
		return 2.0 // 默认波动率
	}
}

// initMockData 初始化Mock数据
func (p *KlineProvider) initMockData() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 初始化K线数据 - 支持多个数据源
	// OKX 数据源
	p.klines["okx"] = &KlineData{
		Symbol:    "BTC",
		Open:      45000.0,
		High:      46000.0,
		Low:       44000.0,
		Close:     45500.0,
		Volume:    500.0,
		Timestamp: time.Now(),
	}

	// Binance 数据源
	p.klines["binance"] = &KlineData{
		Symbol:    "BTC",
		Open:      45100.0,
		High:      46100.0,
		Low:       44100.0,
		Close:     45600.0,
		Volume:    600.0,
		Timestamp: time.Now(),
	}

	// Gate 数据源
	p.klines["gate"] = &KlineData{
		Symbol:    "BTC",
		Open:      45200.0,
		High:      46200.0,
		Low:       44200.0,
		Close:     45700.0,
		Volume:    700.0,
		Timestamp: time.Now(),
	}
}


// GetFunctionParamMapping 获取函数参数映射
func (p *KlineProvider) GetFunctionParamMapping() map[string]FunctionParamInfo {
	return map[string]FunctionParamInfo{
		"avg": {
			FunctionName: "avg",
			Params: []FunctionParam{
				{
					ParamIndex: 1, // 第二个参数（从0开始）
					ParamName:  "interval",
					ParamType:  ParamTypeString,
					Required:   true,
				},
				{
					ParamIndex: 2, // 第三个参数（从0开始）
					ParamName:  "limit",
					ParamType:  ParamTypeInt,
					Required:   true,
				},
			},
		},
		"max": {
			FunctionName: "max",
			Params: []FunctionParam{
				{
					ParamIndex: 1, // 第二个参数（从0开始）
					ParamName:  "interval",
					ParamType:  ParamTypeString,
					Required:   true,
				},
				{
					ParamIndex: 2, // 第三个参数（从0开始）
					ParamName:  "limit",
					ParamType:  ParamTypeInt,
					Required:   true,
				},
			},
		},
		"min": {
			FunctionName: "min",
			Params: []FunctionParam{
				{
					ParamIndex: 1, // 第二个参数（从0开始）
					ParamName:  "interval",
					ParamType:  ParamTypeString,
					Required:   true,
				},
				{
					ParamIndex: 2, // 第三个参数（从0开始）
					ParamName:  "limit",
					ParamType:  ParamTypeInt,
					Required:   true,
				},
			},
		},
		"sum": {
			FunctionName: "sum",
			Params: []FunctionParam{
				{
					ParamIndex: 1, // 第二个参数（从0开始）
					ParamName:  "interval",
					ParamType:  ParamTypeString,
					Required:   true,
				},
				{
					ParamIndex: 2, // 第三个参数（从0开始）
					ParamName:  "limit",
					ParamType:  ParamTypeInt,
					Required:   true,
				},
			},
		},
	}
}
