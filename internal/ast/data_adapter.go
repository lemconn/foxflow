package ast

import (
	"context"

	"github.com/lemconn/foxflow/internal/data"
)

// DataAdapter 数据适配器，将data模块适配到AST的DataProvider接口
type DataAdapter struct {
	dataManager *data.Manager
}

// NewDataAdapter 创建数据适配器
func NewDataAdapter(dataManager *data.Manager) *DataAdapter {
	return &DataAdapter{
		dataManager: dataManager,
	}
}

// GetData 获取数据
func (a *DataAdapter) GetData(ctx context.Context, module DataType, entity, field string) (interface{}, error) {
	return a.dataManager.GetData(ctx, string(module), entity, field)
}

// GetHistoricalData 获取历史数据
func (a *DataAdapter) GetHistoricalData(ctx context.Context, module DataType, entity, field string, period int) ([]interface{}, error) {
	return a.dataManager.GetHistoricalData(ctx, string(module), entity, field, period)
}
