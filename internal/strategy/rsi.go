package strategy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lemconn/foxflow/internal/exchange"
)

// RSIStrategy RSI策略
type RSIStrategy struct {
	name        string
	description string
}

// NewRSIStrategy 创建RSI策略
func NewRSIStrategy() *RSIStrategy {
	return &RSIStrategy{
		name:        "rsi",
		description: "RSI策略：当RSI值小于指定阈值时触发（超卖信号）",
	}
}

func (s *RSIStrategy) GetName() string {
	return s.name
}

func (s *RSIStrategy) GetDescription() string {
	return s.description
}

func (s *RSIStrategy) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"threshold": "RSI阈值",
	}
}

func (s *RSIStrategy) ValidateParameters(params map[string]interface{}) error {
	threshold, exists := params["threshold"]
	if !exists {
		return fmt.Errorf("missing required parameter: threshold")
	}

	// 尝试转换为数字
	switch v := threshold.(type) {
	case float64:
		if v < 0 || v > 100 {
			return fmt.Errorf("RSI threshold must be between 0 and 100")
		}
	case string:
		if val, err := strconv.ParseFloat(v, 64); err != nil || val < 0 || val > 100 {
			return fmt.Errorf("invalid RSI threshold value: %s (must be 0-100)", v)
		}
	default:
		return fmt.Errorf("threshold must be a number")
	}

	return nil
}

func (s *RSIStrategy) Evaluate(ctx context.Context, exchange exchange.Exchange, symbol string, params map[string]interface{}) (bool, error) {
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

	// 模拟RSI计算（实际应该从K线数据计算）
	// 这里使用价格位置作为简化的RSI值
	ticker, err := exchange.GetTicker(ctx, symbol)
	if err != nil {
		return false, fmt.Errorf("failed to get ticker: %w", err)
	}

	// 简化的RSI计算：基于价格在高低价区间的位置
	if ticker.High == ticker.Low {
		return false, fmt.Errorf("invalid price range: high equals low")
	}

	rsiValue := ((ticker.Price - ticker.Low) / (ticker.High - ticker.Low)) * 100

	// RSI小于阈值时触发（超卖信号）
	result := rsiValue < thresholdValue

	return result, nil
}
