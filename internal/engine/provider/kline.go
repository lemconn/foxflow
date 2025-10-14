package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/exchange"
)

// 使用 exchange 包中的 KlineData 类型
type KlineData = exchange.KlineData

// KlineProvider K线数据提供者
type KlineProvider struct {
	*BaseProvider
	exchangeMgr *exchange.Manager
}

// NewKlineProvider 创建K线数据提供者
func NewKlineProvider() *KlineProvider {
	provider := &KlineProvider{
		BaseProvider: NewBaseProvider("kline"),
		exchangeMgr:  exchange.GetManager(),
	}

	return provider
}

// GetData 获取数据
// KlineProvider 通过 exchange 实时获取K线数据
// params 支持的参数：
// - params[0]: string - 时间间隔（可选，如 "15m", "1h", "1d"）
// - params[1]: int - 历史数据周期数（必需）
// - params[2]: time.Time - 开始时间（可选）
// - params[3]: time.Time - 结束时间（可选）
func (p *KlineProvider) GetData(ctx context.Context, dataSource, field string, params ...interface{}) (interface{}, error) {
	// 解析字段 - 支持多级字段如 "BTC.close"
	fieldParts := strings.Split(field, ".")
	if len(fieldParts) < 2 {
		return nil, fmt.Errorf("kline field must be in format 'SYMBOL.FIELD', got: %s", field)
	}

	symbol := fieldParts[0]
	fieldName := fieldParts[1]

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

	// 获取交易所实例
	exchangeInstance, err := p.exchangeMgr.GetExchange(dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange %s: %w", dataSource, err)
	}

	// 使用 exchange 的 ConvertIntervalFormat 转换时间间隔格式
	exchangeInterval := exchangeInstance.ConvertIntervalFormat(interval)

	// 使用 GetSwapSymbolByName 转换 symbol 参数
	exchangeSymbol := exchangeInstance.GetSwapSymbolByName(ctx, symbol)

	// 通过 exchange 实时获取K线数据
	klineData, err := exchangeInstance.GetKlineData(ctx, exchangeSymbol, exchangeInterval, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get kline data for %s %s %s: %w", dataSource, exchangeSymbol, exchangeInterval, err)
	}

	// 提取指定字段的历史数据
	result := make([]interface{}, 0, len(klineData))
	for _, kline := range klineData {
		var value interface{}
		switch fieldName {
		case "open":
			value = kline.Open
		case "high":
			value = kline.High
		case "low":
			value = kline.Low
		case "close":
			value = kline.Close
		case "volume":
			value = kline.Volume
		default:
			return nil, fmt.Errorf("unknown field: %s", fieldName)
		}
		result = append(result, value)
	}

	return result, nil
}

