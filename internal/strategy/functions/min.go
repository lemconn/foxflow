package functions

import (
	"context"
	"fmt"
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

	// 第一个参数应该是数据数组
	data, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("first argument to min must be a data array")
	}

	// 第二个参数应该是周期数
	period, err := toFloat64(args[1])
	if err != nil {
		return nil, fmt.Errorf("second argument to min must be a number: %w", err)
	}

	// 计算最小值
	n := int(period)
	if len(data) < n {
		n = len(data)
	}
	if n == 0 {
		return 0.0, nil
	}

	min := 0.0
	hasValue := false
	for _, v := range data[len(data)-n:] {
		if val, ok := v.(float64); ok {
			if !hasValue || val < min {
				min = val
				hasValue = true
			}
		} else {
			return nil, fmt.Errorf("invalid data type in array: %T", v)
		}
	}

	if !hasValue {
		return 0.0, nil
	}

	return min, nil
}
