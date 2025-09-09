package strategy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lemconn/foxflow/internal/exchange"
)

// MACDStrategy MACD策略
type MACDStrategy struct {
	name        string
	description string
}

// NewMACDStrategy 创建MACD策略
func NewMACDStrategy() *MACDStrategy {
	return &MACDStrategy{
		name:        "macd",
		description: "MACD策略：当MACD值大于指定阈值时触发",
	}
}

func (s *MACDStrategy) GetName() string {
	return s.name
}

func (s *MACDStrategy) GetDescription() string {
	return s.description
}

func (s *MACDStrategy) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"threshold": "MACD阈值",
	}
}

func (s *MACDStrategy) ValidateParameters(params map[string]interface{}) error {
	threshold, exists := params["threshold"]
	if !exists {
		return fmt.Errorf("missing required parameter: threshold")
	}

	// 尝试转换为数字
	switch v := threshold.(type) {
	case float64:
		// MACD可以是负数，所以不检查正负
	case string:
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return fmt.Errorf("invalid threshold value: %s", v)
		}
	default:
		return fmt.Errorf("threshold must be a number")
	}

	return nil
}

func (s *MACDStrategy) Evaluate(ctx context.Context, exchange exchange.Exchange, symbol string, params map[string]interface{}) (bool, error) {
	// 获取阈值参数
	threshold, exists := params["threshold"]
	if !exists {
		return false, fmt.Errorf("missing threshold parameter")
	}

	var thresholdValue float64
	switch v := threshold.(type) {
	case float64:
		thresholdValue = v
	case string:
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false, fmt.Errorf("invalid threshold value: %s", v)
		}
		thresholdValue = val
	default:
		return false, fmt.Errorf("threshold must be a number")
	}

	// 模拟MACD计算（实际应该从K线数据计算）
	// 这里使用价格作为简化的MACD值
	ticker, err := exchange.GetTicker(ctx, symbol)
	if err != nil {
		return false, fmt.Errorf("failed to get ticker: %w", err)
	}

	// 简化的MACD计算：使用价格变化率
	macdValue := (ticker.Price - ticker.Low) / (ticker.High - ticker.Low) * 100

	// 比较MACD值
	result := macdValue > thresholdValue

	return result, nil
}
