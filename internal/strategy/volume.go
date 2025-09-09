package strategy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lemconn/foxflow/internal/exchange"
)

// VolumeStrategy 成交量策略
type VolumeStrategy struct {
	name        string
	description string
}

// NewVolumeStrategy 创建成交量策略
func NewVolumeStrategy() *VolumeStrategy {
	return &VolumeStrategy{
		name:        "volume",
		description: "成交量策略：当成交量大于指定阈值时触发",
	}
}

func (s *VolumeStrategy) GetName() string {
	return s.name
}

func (s *VolumeStrategy) GetDescription() string {
	return s.description
}

func (s *VolumeStrategy) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"threshold": "成交量阈值",
	}
}

func (s *VolumeStrategy) ValidateParameters(params map[string]interface{}) error {
	threshold, exists := params["threshold"]
	if !exists {
		return fmt.Errorf("missing required parameter: threshold")
	}

	// 尝试转换为数字
	switch v := threshold.(type) {
	case float64:
		if v <= 0 {
			return fmt.Errorf("threshold must be positive")
		}
	case string:
		if val, err := strconv.ParseFloat(v, 64); err != nil || val <= 0 {
			return fmt.Errorf("invalid threshold value: %s", v)
		}
	default:
		return fmt.Errorf("threshold must be a number")
	}

	return nil
}

func (s *VolumeStrategy) Evaluate(ctx context.Context, exchange exchange.Exchange, symbol string, params map[string]interface{}) (bool, error) {
	// 获取当前行情
	ticker, err := exchange.GetTicker(ctx, symbol)
	if err != nil {
		return false, fmt.Errorf("failed to get ticker: %w", err)
	}

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

	// 比较成交量
	result := ticker.Volume > thresholdValue

	return result, nil
}
