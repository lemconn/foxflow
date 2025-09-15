package functions

import (
	"context"
	"fmt"
	"strings"

	sources "github.com/lemconn/foxflow/internal/strategy/datasources"
)

// MaxFunction max函数实现
type MaxFunction struct {
	*BaseFunction
}

// NewMaxFunction 创建max函数
func NewMaxFunction() *MaxFunction {
	signature := Signature{
		Name:        "max",
		Description: "计算指定数据源和字段的最大值",
		ReturnType:  "float64",
		Args: []ArgInfo{
			{
				Name:        "path",
				Type:        "string",
				Required:    true,
				Description: "数据路径，格式：kline.SYMBOL.field",
			},
			{
				Name:        "period",
				Type:        "number",
				Required:    true,
				Description: "计算周期数",
			},
		},
	}

	return &MaxFunction{
		BaseFunction: NewBaseFunction("max", "计算指定数据源和字段的最大值", signature),
	}
}

// Execute 执行max函数
func (f *MaxFunction) Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error) {
	if err := f.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 第一个参数应该是数据引用路径
	path, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("first argument to max must be a string path")
	}

	// 解析路径
	parts := strings.Split(path, ".")
	if len(parts) != 3 || parts[0] != "kline" {
		return nil, fmt.Errorf("invalid path for max function: %s", path)
	}

	symbol := parts[1]
	field := parts[2]

	// 第二个参数应该是周期数
	period, err := toFloat64(args[1])
	if err != nil {
		return nil, fmt.Errorf("second argument to max must be a number: %w", err)
	}

	// 获取历史数据
	data, err := f.getKlineData(ctx, evaluator, symbol, field)
	if err != nil {
		return nil, fmt.Errorf("failed to get kline data: %w", err)
	}

	// 计算最大值
	n := int(period)
	if len(data) < n {
		n = len(data)
	}
	if n == 0 {
		return 0.0, nil
	}

	max := data[len(data)-n]
	for _, v := range data[len(data)-n:] {
		if v > max {
			max = v
		}
	}

	return max, nil
}

// getKlineData 获取K线数据
func (f *MaxFunction) getKlineData(ctx context.Context, evaluator Evaluator, symbol, field string) ([]float64, error) {
	// 获取K线数据源
	ds, exists := evaluator.GetDataSource("kline")
	if !exists {
		return nil, fmt.Errorf("kline data source not found")
	}

	// 尝试类型断言为KlineModule
	if klineModule, ok := ds.(*sources.KlineModule); ok {
		// 获取历史数据
		historicalData, err := klineModule.GetHistoricalData(ctx, symbol, field, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to get kline historical data for %s.%s: %w", symbol, field, err)
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

	// 如果不是KlineModule，尝试通过GetHistoricalData接口获取
	type HistoricalDataProvider interface {
		GetHistoricalData(ctx context.Context, entity, field string, period int) ([]interface{}, error)
	}

	if historicalProvider, ok := ds.(HistoricalDataProvider); ok {
		historicalData, err := historicalProvider.GetHistoricalData(ctx, symbol, field, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to get kline historical data for %s.%s: %w", symbol, field, err)
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

	return nil, fmt.Errorf("invalid kline module type")
}
