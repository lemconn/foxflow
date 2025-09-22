package builtin

import (
	"context"
	"fmt"
	"time"
)

// AgoBuiltin ago函数实现
type AgoBuiltin struct {
	*BaseBuiltin
}

// NewAgoBuiltin 创建ago函数
func NewAgoBuiltin() *AgoBuiltin {
	signature := Signature{
		Name:        "ago",
		Description: "计算从指定时间到现在的秒数",
		ReturnType:  "float64",
		Args: []ArgInfo{
			{
				Name:        "timestamp",
				Type:        "time",
				Required:    true,
				Description: "时间戳",
			},
		},
	}

	return &AgoBuiltin{
		BaseBuiltin: NewBaseBuiltin("ago", "计算从指定时间到现在的秒数", signature),
	}
}

// Execute 执行ago函数
func (f *AgoBuiltin) Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error) {
	if err := f.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 参数应该是时间
	timestamp, err := toTime(args[0])
	if err != nil {
		return nil, fmt.Errorf("argument to ago must be a time: %w", err)
	}

	// 计算从指定时间到现在的秒数
	now := time.Now()
	duration := now.Sub(timestamp)
	return duration.Seconds(), nil
}
