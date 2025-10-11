package builtin

import (
	"context"
	"fmt"
)

// SumBuiltin sum函数实现
type SumBuiltin struct {
	*BaseBuiltin
}

// NewSumBuiltin 创建sum函数
func NewSumBuiltin() *SumBuiltin {
	signature := Signature{
		Name:        "sum",
		Description: "计算指定数据源和字段的总和",
		ReturnType:  "float64",
		Args: []ArgInfo{
			{
				Name:        "path",
				Type:        "string",
				Required:    true,
				Description: "数据路径，格式：kline.SYMBOL.field",
			},
			{
				Name:        "interval",
				Type:        "string",
				Required:    true,
				Description: "时间间隔，如：15m, 1h, 1d",
			},
			{
				Name:        "limit",
				Type:        "number",
				Required:    true,
				Description: "数据点数量",
			},
		},
	}

	return &SumBuiltin{
		BaseBuiltin: NewBaseBuiltin("sum", "计算指定数据源和字段的总和", signature),
	}
}

// Execute 执行sum函数
func (f *SumBuiltin) Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error) {
	if err := f.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 第一个参数应该是数据数组
	data, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("first argument to sum must be a data array")
	}

	// 第二个参数应该是时间间隔（字符串）
	_, ok = args[1].(string)
	if !ok {
		return nil, fmt.Errorf("second argument to sum must be a string (interval)")
	}

	// 第三个参数应该是数据点数量
	limit, err := toFloat64(args[2])
	if err != nil {
		return nil, fmt.Errorf("third argument to sum must be a number: %w", err)
	}

	// 计算总和
	n := int(limit)
	if len(data) < n {
		n = len(data)
	}

	sum := 0.0
	for _, v := range data[len(data)-n:] {
		if val, ok := v.(float64); ok {
			sum += val
		} else {
			return nil, fmt.Errorf("invalid data type in array: %T", v)
		}
	}

	return sum, nil
}
