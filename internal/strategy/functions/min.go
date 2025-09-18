package functions

import (
	"context"
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/strategy/datasources"
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
	if len(parts) != 3 || parts[0] != "kline" {
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
	data, err := f.getKlineData(ctx, evaluator, symbol, field)
	if err != nil {
		return nil, fmt.Errorf("failed to get kline data: %w", err)
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

// getKlineData 获取K线数据
func (f *MinFunction) getKlineData(ctx context.Context, evaluator Evaluator, symbol, field string) ([]float64, error) {
	// 获取K线数据源
	ds, exists := evaluator.GetDataSource("kline")
	if !exists {
		return nil, fmt.Errorf("kline data source not found")
	}

	// 类型断言为 Module 接口
	module, ok := ds.(datasources.Module)
	if !ok {
		return nil, fmt.Errorf("data source is not a Module")
	}

	// 使用新的 GetData 方法获取历史数据
	params := []datasources.DataParam{
		datasources.NewParam("period", 100),
	}
	historicalData, err := module.GetData(ctx, symbol, field, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to get kline historical data for %s.%s: %w", symbol, field, err)
	}

	// 检查返回的数据类型
	historicalArray, ok := historicalData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected []interface{} for historical data, got %T", historicalData)
	}

	// 转换为float64数组
	result := make([]float64, len(historicalArray))
	for i, v := range historicalArray {
		if f, ok := v.(float64); ok {
			result[i] = f
		} else {
			return nil, fmt.Errorf("invalid data type in historical data: %T", v)
		}
	}

	return result, nil
}
