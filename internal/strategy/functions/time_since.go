package functions

import (
	"context"
	"fmt"
	"time"
)

// TimeSinceFunction time_since函数实现
type TimeSinceFunction struct {
	*BaseFunction
}

// NewTimeSinceFunction 创建time_since函数
func NewTimeSinceFunction() *TimeSinceFunction {
	signature := Signature{
		Name:        "time_since",
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

	return &TimeSinceFunction{
		BaseFunction: NewBaseFunction("time_since", "计算从指定时间到现在的秒数", signature),
	}
}

// Execute 执行time_since函数
func (f *TimeSinceFunction) Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error) {
	if err := f.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 参数应该是时间
	timestamp, err := toTime(args[0])
	if err != nil {
		return nil, fmt.Errorf("argument to time_since must be a time: %w", err)
	}

	// 计算从指定时间到现在的秒数
	now := time.Now()
	duration := now.Sub(timestamp)
	return duration.Seconds(), nil
}
