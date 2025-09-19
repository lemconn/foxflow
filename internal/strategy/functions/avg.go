package functions

import (
	"context"
	"fmt"
)

// AvgFunction avg函数实现
type AvgFunction struct {
	*BaseFunction
}

// NewAvgFunction 创建avg函数
func NewAvgFunction() *AvgFunction {
	signature := Signature{
		Name:        "avg",
		Description: "计算指定数据源和字段的平均值",
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

	return &AvgFunction{
		BaseFunction: NewBaseFunction("avg", "计算指定数据源和字段的平均值", signature),
	}
}

// Execute 执行avg函数
func (f *AvgFunction) Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error) {
	if err := f.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 第一个参数应该是数据数组
	data, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("first argument to avg must be a data array")
	}

	// 第二个参数应该是周期数
	period, err := toFloat64(args[1])
	if err != nil {
		return nil, fmt.Errorf("second argument to avg must be a number: %w", err)
	}

	// 计算平均值
	n := int(period)
	if len(data) < n {
		n = len(data)
	}
	if n == 0 {
		return 0.0, nil
	}

	sum := 0.0
	for _, v := range data[len(data)-n:] {
		if val, ok := v.(float64); ok {
			sum += val
		} else {
			return nil, fmt.Errorf("invalid data type in array: %T", v)
		}
	}

	return sum / float64(n), nil
}
