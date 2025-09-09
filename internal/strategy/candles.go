package strategy

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/lemconn/foxflow/internal/exchange"
)

// CandleData K线数据
type CandleData struct {
	Symbol     string    `json:"symbol"`
	Open       float64   `json:"open"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Close      float64   `json:"close"`
	Volume     float64   `json:"volume"`
	Timestamp  time.Time `json:"timestamp"`
	LastPx     float64   `json:"last_px"`     // 最新价格
	LastVolume float64   `json:"last_volume"` // 最新成交量
}

// CandlesStrategy K线策略
type CandlesStrategy struct {
	name        string
	description string
	mockData    map[string]*CandleData
	mu          sync.RWMutex
}

// NewCandlesStrategy 创建K线策略
func NewCandlesStrategy() *CandlesStrategy {
	strategy := &CandlesStrategy{
		name:        "candles",
		description: "K线数据策略：基于K线数据进行条件判断",
		mockData:    make(map[string]*CandleData),
	}

	// 初始化Mock数据
	strategy.initMockData()

	return strategy
}

// initMockData 初始化Mock数据
func (s *CandlesStrategy) initMockData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// SOL Mock数据
	s.mockData["SOL"] = &CandleData{
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

	// BTC Mock数据
	s.mockData["BTC"] = &CandleData{
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

	// ETH Mock数据
	s.mockData["ETH"] = &CandleData{
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

func (s *CandlesStrategy) GetName() string {
	return s.name
}

func (s *CandlesStrategy) GetDescription() string {
	return s.description
}

func (s *CandlesStrategy) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"symbol":   "交易对符号，如 SOL, BTC, ETH",
		"field":    "字段名：last_px, last_volume, open, high, low, close, volume",
		"operator": "操作符：>, <, >=, <=, ==, !=",
		"value":    "比较值",
	}
}

func (s *CandlesStrategy) ValidateParameters(params map[string]interface{}) error {
	// 验证必需参数
	requiredParams := []string{"symbol", "field", "operator", "value"}
	for _, param := range requiredParams {
		if _, exists := params[param]; !exists {
			return fmt.Errorf("missing required parameter: %s", param)
		}
	}

	// 验证符号
	symbol, ok := params["symbol"].(string)
	if !ok {
		return fmt.Errorf("symbol must be a string")
	}

	// 检查符号是否在Mock数据中存在
	s.mu.RLock()
	_, exists := s.mockData[symbol]
	s.mu.RUnlock()
	if !exists {
		return fmt.Errorf("invalid symbol: %s", symbol)
	}

	// 验证字段名
	field, ok := params["field"].(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	validFields := []string{"last_px", "last_volume", "open", "high", "low", "close", "volume"}
	validField := false
	for _, validF := range validFields {
		if field == validF {
			validField = true
			break
		}
	}
	if !validField {
		return fmt.Errorf("invalid field: %s, valid fields are: %v", field, validFields)
	}

	// 验证操作符
	operator, ok := params["operator"].(string)
	if !ok {
		return fmt.Errorf("operator must be a string")
	}

	validOperators := []string{">", "<", ">=", "<=", "==", "!="}
	validOp := false
	for _, validO := range validOperators {
		if operator == validO {
			validOp = true
			break
		}
	}
	if !validOp {
		return fmt.Errorf("invalid operator: %s, valid operators are: %v", operator, validOperators)
	}

	// 验证值
	value := params["value"]
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	// 尝试转换为数字
	switch v := value.(type) {
	case float64:
		// 数字值，检查是否为正数（对于价格和成交量）
		if (field == "last_px" || field == "open" || field == "high" || field == "low" || field == "close" || field == "last_volume" || field == "volume") && v < 0 {
			return fmt.Errorf("value must be positive for field %s", field)
		}
	case string:
		// 字符串值，尝试转换为数字
		if num, err := strconv.ParseFloat(v, 64); err != nil {
			return fmt.Errorf("invalid numeric value: %s", v)
		} else if (field == "last_px" || field == "open" || field == "high" || field == "low" || field == "close" || field == "last_volume" || field == "volume") && num < 0 {
			return fmt.Errorf("value must be positive for field %s", field)
		}
	default:
		return fmt.Errorf("value must be a number or numeric string")
	}

	return nil
}

func (s *CandlesStrategy) Evaluate(ctx context.Context, exchange exchange.Exchange, symbol string, params map[string]interface{}) (bool, error) {
	// 获取参数
	symbolParam, ok := params["symbol"].(string)
	if !ok {
		return false, fmt.Errorf("symbol parameter must be a string")
	}

	field, ok := params["field"].(string)
	if !ok {
		return false, fmt.Errorf("field parameter must be a string")
	}

	operator, ok := params["operator"].(string)
	if !ok {
		return false, fmt.Errorf("operator parameter must be a string")
	}

	value := params["value"]
	if value == nil {
		return false, fmt.Errorf("value parameter cannot be nil")
	}

	// 获取K线数据
	candleData, err := s.getCandleData(symbolParam)
	if err != nil {
		return false, fmt.Errorf("failed to get candle data for %s: %w", symbolParam, err)
	}

	// 获取字段值
	fieldValue, err := s.getFieldValue(candleData, field)
	if err != nil {
		return false, fmt.Errorf("failed to get field value: %w", err)
	}

	// 转换比较值
	compareValue, err := s.convertValue(value)
	if err != nil {
		return false, fmt.Errorf("failed to convert value: %w", err)
	}

	// 执行比较
	result, err := s.compareValues(fieldValue, operator, compareValue)
	if err != nil {
		return false, fmt.Errorf("failed to compare values: %w", err)
	}

	return result, nil
}

// getCandleData 获取K线数据
func (s *CandlesStrategy) getCandleData(symbol string) (*CandleData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.mockData[symbol]
	if !exists {
		return nil, fmt.Errorf("no candle data found for symbol: %s", symbol)
	}

	return data, nil
}

// getFieldValue 获取字段值
func (s *CandlesStrategy) getFieldValue(data *CandleData, field string) (float64, error) {
	switch field {
	case "last_px":
		return data.LastPx, nil
	case "last_volume":
		return data.LastVolume, nil
	case "open":
		return data.Open, nil
	case "high":
		return data.High, nil
	case "low":
		return data.Low, nil
	case "close":
		return data.Close, nil
	case "volume":
		return data.Volume, nil
	default:
		return 0, fmt.Errorf("unknown field: %s", field)
	}
}

// convertValue 转换值
func (s *CandlesStrategy) convertValue(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unsupported value type: %T", value)
	}
}

// compareValues 比较值
func (s *CandlesStrategy) compareValues(fieldValue float64, operator string, compareValue float64) (bool, error) {
	switch operator {
	case ">":
		return fieldValue > compareValue, nil
	case "<":
		return fieldValue < compareValue, nil
	case ">=":
		return fieldValue >= compareValue, nil
	case "<=":
		return fieldValue <= compareValue, nil
	case "==":
		return fieldValue == compareValue, nil
	case "!=":
		return fieldValue != compareValue, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// UpdateMockData 更新Mock数据（用于测试）
func (s *CandlesStrategy) UpdateMockData(symbol string, data *CandleData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mockData[symbol] = data
}

// GetMockData 获取Mock数据（用于测试）
func (s *CandlesStrategy) GetMockData(symbol string) (*CandleData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, exists := s.mockData[symbol]
	return data, exists
}
