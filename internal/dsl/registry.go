package dsl

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Function 函数签名
type Function func(ctx context.Context, args []interface{}, evaluator *Evaluator) (interface{}, error)

// DataProvider 数据提供者接口
type DataProvider interface {
	GetCandles(ctx context.Context, symbol, field string) ([]float64, error)
	GetCandleField(ctx context.Context, symbol, field string) (interface{}, error)
	GetNewsField(ctx context.Context, source, field string) (interface{}, error)
	GetIndicatorField(ctx context.Context, symbol, field string) (interface{}, error)
}

// Registry 函数注册表
type Registry struct {
	functions map[string]Function
}

// NewRegistry 创建函数注册表
func NewRegistry() *Registry {
	return &Registry{
		functions: make(map[string]Function),
	}
}

// Register 注册函数
func (r *Registry) Register(name string, fn Function) {
	r.functions[name] = fn
}

// GetFunction 获取函数
func (r *Registry) GetFunction(name string) (Function, bool) {
	fn, exists := r.functions[name]
	return fn, exists
}

// DefaultRegistry 创建默认函数注册表
func DefaultRegistry() *Registry {
	registry := NewRegistry()

	// 注册avg函数
	registry.Register("avg", func(ctx context.Context, args []interface{}, evaluator *Evaluator) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("avg function expects 2 arguments, got %d", len(args))
		}

		// 第一个参数应该是数据引用路径
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("first argument to avg must be a string path")
		}

		// 解析路径：candles.SYMBOL.field
		parts := strings.Split(path, ".")
		if len(parts) != 3 || parts[0] != "candles" {
			return nil, fmt.Errorf("invalid path for avg function: %s", path)
		}

		symbol := parts[1]
		field := parts[2]

		// 第二个参数应该是周期数
		period, err := toFloat64(args[1])
		if err != nil {
			return nil, fmt.Errorf("second argument to avg must be a number: %w", err)
		}

		// 获取历史数据
		data, err := evaluator.dataProvider.GetCandles(ctx, symbol, field)
		if err != nil {
			return nil, fmt.Errorf("failed to get candles data: %w", err)
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
			sum += v
		}

		return sum / float64(n), nil
	})

	// 注册time_since函数
	registry.Register("time_since", func(ctx context.Context, args []interface{}, evaluator *Evaluator) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("time_since function expects 1 argument, got %d", len(args))
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
	})

	// 注册has函数
	registry.Register("has", func(ctx context.Context, args []interface{}, evaluator *Evaluator) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("has function expects 2 arguments, got %d", len(args))
		}

		// 第一个参数应该是字符串
		text := toString(args[0])

		// 第二个参数应该是字符串
		keyword := toString(args[1])

		// 检查文本是否包含关键词
		return contains(text, keyword), nil
	})

	// 注册max函数
	registry.Register("max", func(ctx context.Context, args []interface{}, evaluator *Evaluator) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("max function expects 2 arguments, got %d", len(args))
		}

		// 第一个参数应该是数据引用路径
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("first argument to max must be a string path")
		}

		// 解析路径
		parts := strings.Split(path, ".")
		if len(parts) != 3 || parts[0] != "candles" {
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
		data, err := evaluator.dataProvider.GetCandles(ctx, symbol, field)
		if err != nil {
			return nil, fmt.Errorf("failed to get candles data: %w", err)
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
	})

	// 注册min函数
	registry.Register("min", func(ctx context.Context, args []interface{}, evaluator *Evaluator) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("min function expects 2 arguments, got %d", len(args))
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
		data, err := evaluator.dataProvider.GetCandles(ctx, symbol, field)
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
	})

	// 注册sum函数
	registry.Register("sum", func(ctx context.Context, args []interface{}, evaluator *Evaluator) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("sum function expects 2 arguments, got %d", len(args))
		}

		// 第一个参数应该是数据引用路径
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("first argument to sum must be a string path")
		}

		// 解析路径
		parts := strings.Split(path, ".")
		if len(parts) != 3 || parts[0] != "candles" {
			return nil, fmt.Errorf("invalid path for sum function: %s", path)
		}

		symbol := parts[1]
		field := parts[2]

		// 第二个参数应该是周期数
		period, err := toFloat64(args[1])
		if err != nil {
			return nil, fmt.Errorf("second argument to sum must be a number: %w", err)
		}

		// 获取历史数据
		data, err := evaluator.dataProvider.GetCandles(ctx, symbol, field)
		if err != nil {
			return nil, fmt.Errorf("failed to get candles data: %w", err)
		}

		// 计算总和
		n := int(period)
		if len(data) < n {
			n = len(data)
		}

		sum := 0.0
		for _, v := range data[len(data)-n:] {
			sum += v
		}

		return sum, nil
	})

	return registry
}
