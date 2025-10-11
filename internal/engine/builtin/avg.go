package builtin

import (
	"context"
	"fmt"
)

// AvgBuiltin avg函数实现
type AvgBuiltin struct {
	*BaseBuiltin
}

// NewAvgBuiltin 创建avg函数
func NewAvgBuiltin() *AvgBuiltin {
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

	return &AvgBuiltin{
		BaseBuiltin: NewBaseBuiltin("avg", "计算指定数据源和字段的平均值", signature),
	}
}

// Execute 执行avg函数
func (f *AvgBuiltin) Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error) {
	if err := f.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 第一个参数应该是数据数组
	data, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("first argument to avg must be a data array")
	}

	// 第二个参数应该是时间间隔（字符串）
	_, ok = args[1].(string)
	if !ok {
		return nil, fmt.Errorf("second argument to avg must be a string (interval)")
	}

	// 第三个参数应该是数据点数量
	limit, err := toFloat64(args[2])
	if err != nil {
		return nil, fmt.Errorf("third argument to avg must be a number: %w", err)
	}

	// 计算平均值
	n := int(limit)
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
