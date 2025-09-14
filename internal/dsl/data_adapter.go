package dsl

import (
	"context"
	"fmt"

	"github.com/lemconn/foxflow/internal/data"
)

// DataAdapter 数据适配器，将现有的数据模块适配到新的DSL框架
type DataAdapter struct {
	dataManager *data.Manager
}

// NewDataAdapter 创建数据适配器
func NewDataAdapter(dataManager *data.Manager) *DataAdapter {
	return &DataAdapter{
		dataManager: dataManager,
	}
}

// GetCandles 获取K线数据
func (da *DataAdapter) GetCandles(ctx context.Context, symbol, field string) ([]float64, error) {
	// 获取candles模块
	candlesModule, err := da.dataManager.GetModule("candles")
	if err != nil {
		return nil, fmt.Errorf("failed to get candles module: %w", err)
	}

	// 获取历史数据
	historicalData, err := candlesModule.GetHistoricalData(ctx, symbol, field, 100) // 获取最近100条数据
	if err != nil {
		return nil, fmt.Errorf("failed to get candles historical data for %s.%s: %w", symbol, field, err)
	}

	// 转换为float64数组
	result := make([]float64, len(historicalData))
	for i, v := range historicalData {
		if f, ok := v.(float64); ok {
			result[i] = f
		} else {
			return nil, fmt.Errorf("invalid data type in historical data: %T", v)
		}
	}

	return result, nil
}

// GetCandleField 获取K线单个字段值
func (da *DataAdapter) GetCandleField(ctx context.Context, symbol, field string) (interface{}, error) {
	// 直接使用数据管理器获取数据
	return da.dataManager.GetData(ctx, "candles", symbol, field)
}

// GetNewsField 获取新闻字段值
func (da *DataAdapter) GetNewsField(ctx context.Context, source, field string) (interface{}, error) {
	// 直接使用数据管理器获取数据
	return da.dataManager.GetData(ctx, "news", source, field)
}

// GetIndicatorField 获取指标字段值
func (da *DataAdapter) GetIndicatorField(ctx context.Context, symbol, field string) (interface{}, error) {
	// 直接使用数据管理器获取数据
	return da.dataManager.GetData(ctx, "indicators", symbol, field)
}
