package functions

import (
	"context"
	"fmt"
	"strings"
)

// MinFunction min函数实现
type MinFunction struct {
	*BaseFunction
}

// NewMinFunction 创建min函数
func NewMinFunction() *MinFunction {
	signature := Signature{
		Name:        "min",
		Description: "计算指定数据源和字段的最小值",
		ReturnType:  "float64",
		Args: []ArgInfo{
			{
				Name:        "path",
				Type:        "string",
				Required:    true,
				Description: "数据路径，格式：candles.SYMBOL.field",
			},
			{
				Name:        "period",
				Type:        "number",
				Required:    true,
				Description: "计算周期数",
			},
		},
	}

	return &MinFunction{
		BaseFunction: NewBaseFunction("min", "计算指定数据源和字段的最小值", signature),
	}
}

// Execute 执行min函数
func (f *MinFunction) Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error) {
	if err := f.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 第一个参数应该是数据引用路径
	path, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("first argument to min must be a string path")
	}

	// 解析路径
	parts := strings.Split(path, ".")
	if len(parts) != 3 || parts[0] != "candles" {
		return nil, fmt.Errorf("invalid path for min function: %s", path)
	}

	symbol := parts[1]
	field := parts[2]

	// 第二个参数应该是周期数
	period, err := toFloat64(args[1])
	if err != nil {
		return nil, fmt.Errorf("second argument to min must be a number: %w", err)
	}

	// 获取历史数据
	data, err := f.getCandlesData(ctx, evaluator, symbol, field)
	if err != nil {
		return nil, fmt.Errorf("failed to get candles data: %w", err)
	}

	// 计算最小值
	n := int(period)
	if len(data) < n {
		n = len(data)
	}
	if n == 0 {
		return 0.0, nil
	}

	min := data[len(data)-n]
	for _, v := range data[len(data)-n:] {
		if v < min {
			min = v
		}
	}

	return min, nil
}

// getCandlesData 获取K线数据
func (f *MinFunction) getCandlesData(ctx context.Context, evaluator Evaluator, symbol, field string) ([]float64, error) {
	// 获取历史数据
	historicalData, err := evaluator.GetHistoricalData(ctx, "candles", symbol, field, 100)
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
