package ast

import (
	"context"
)

// DataProvider 数据提供者接口
type DataProvider interface {
	// GetData 获取指定模块、实体和字段的数据
	GetData(ctx context.Context, module DataType, entity, field string) (interface{}, error)

	// GetHistoricalData 获取历史数据（用于avg等函数）
	GetHistoricalData(ctx context.Context, module DataType, entity, field string, period int) ([]interface{}, error)
}
