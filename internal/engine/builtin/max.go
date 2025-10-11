package builtin

import (
	"context"
	"fmt"
)

// MaxBuiltin max函数实现
type MaxBuiltin struct {
	*BaseBuiltin
}

// NewMaxBuiltin 创建max函数
func NewMaxBuiltin() *MaxBuiltin {
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

	return &MaxBuiltin{
		BaseBuiltin: NewBaseBuiltin("max", "计算指定数据源和字段的最大值", signature),
	}
}

// Execute 执行max函数
func (f *MaxBuiltin) Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error) {
	if err := f.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 第一个参数应该是数据数组
	data, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("first argument to max must be a data array")
	}

	// 第二个参数应该是时间间隔（字符串）
	_, ok = args[1].(string)
	if !ok {
		return nil, fmt.Errorf("second argument to max must be a string (interval)")
	}

	// 第三个参数应该是数据点数量
	limit, err := toFloat64(args[2])
	if err != nil {
		return nil, fmt.Errorf("third argument to max must be a number: %w", err)
	}

	// 计算最大值
	n := int(limit)
	if len(data) < n {
		n = len(data)
	}
	if n == 0 {
		return 0.0, nil
	}

	max := 0.0
	hasValue := false
	for _, v := range data[len(data)-n:] {
		if val, ok := v.(float64); ok {
			if !hasValue || val > max {
				max = val
				hasValue = true
			}
		} else {
			return nil, fmt.Errorf("invalid data type in array: %T", v)
		}
	}

	if !hasValue {
		return 0.0, nil
	}

	return max, nil
}
