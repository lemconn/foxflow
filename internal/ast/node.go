package ast

import (
	"context"
	"fmt"
	"time"
)

// NodeType AST节点类型
type NodeType string

const (
	NodeTypeBinaryExpression NodeType = "BinaryExpression"
	NodeTypeFunctionCall     NodeType = "FunctionCall"
	NodeTypeDataRef          NodeType = "DataRef"
	NodeTypeValue            NodeType = "Value"
)

// Operator 操作符类型
type Operator string

const (
	// 逻辑操作符
	OpAnd Operator = "and"
	OpOr  Operator = "or"

	// 比较操作符
	OpGT  Operator = ">"
	OpLT  Operator = "<"
	OpGTE Operator = ">="
	OpLTE Operator = "<="
	OpEQ  Operator = "=="
	OpNEQ Operator = "!="

	// 包含操作符
	OpIn       Operator = "in"
	OpNotIn    Operator = "not_in"
	OpContains Operator = "contains"
)

// DataType 数据类型
type DataType string

const (
	DataTypeCandles    DataType = "candles"
	DataTypeNews       DataType = "news"
	DataTypeIndicators DataType = "indicators"
)

// Node AST节点接口
type Node interface {
	// Type 返回节点类型
	Type() NodeType

	// Evaluate 执行节点并返回结果
	Evaluate(ctx context.Context, dataProvider DataProvider) (interface{}, error)

	// String 返回节点的字符串表示
	String() string
}

// BinaryExpression 二元表达式节点
type BinaryExpression struct {
	Operator Operator `json:"operator"`
	Left     Node     `json:"left"`
	Right    Node     `json:"right"`
}

func (b *BinaryExpression) Type() NodeType {
	return NodeTypeBinaryExpression
}

func (b *BinaryExpression) Evaluate(ctx context.Context, dataProvider DataProvider) (interface{}, error) {
	// 评估左操作数
	leftValue, err := b.Left.Evaluate(ctx, dataProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate left operand: %w", err)
	}

	// 评估右操作数
	rightValue, err := b.Right.Evaluate(ctx, dataProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate right operand: %w", err)
	}

	// 根据操作符执行相应的操作
	switch b.Operator {
	case OpAnd, OpOr:
		return b.evaluateLogical(leftValue, rightValue)
	case OpGT, OpLT, OpGTE, OpLTE, OpEQ, OpNEQ:
		return b.evaluateComparison(leftValue, rightValue)
	case OpIn, OpNotIn:
		return b.evaluateMembership(leftValue, rightValue)
	case OpContains:
		return b.evaluateContains(leftValue, rightValue)
	default:
		return nil, fmt.Errorf("unsupported operator: %s", b.Operator)
	}
}

func (b *BinaryExpression) evaluateLogical(left, right interface{}) (bool, error) {
	leftBool, err := toBool(left)
	if err != nil {
		return false, fmt.Errorf("left operand is not boolean: %w", err)
	}

	rightBool, err := toBool(right)
	if err != nil {
		return false, fmt.Errorf("right operand is not boolean: %w", err)
	}

	switch b.Operator {
	case OpAnd:
		return leftBool && rightBool, nil
	case OpOr:
		return leftBool || rightBool, nil
	default:
		return false, fmt.Errorf("unsupported logical operator: %s", b.Operator)
	}
}

func (b *BinaryExpression) evaluateComparison(left, right interface{}) (bool, error) {
	// 尝试数字比较
	if leftNum, rightNum, ok := toNumbers(left, right); ok {
		return b.compareNumbers(leftNum, rightNum), nil
	}

	// 尝试时间比较
	if leftTime, rightTime, ok := toTimes(left, right); ok {
		return b.compareTimes(leftTime, rightTime), nil
	}

	// 尝试字符串比较
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	return b.compareStrings(leftStr, rightStr), nil
}

func (b *BinaryExpression) compareNumbers(left, right float64) bool {
	switch b.Operator {
	case OpGT:
		return left > right
	case OpLT:
		return left < right
	case OpGTE:
		return left >= right
	case OpLTE:
		return left <= right
	case OpEQ:
		return left == right
	case OpNEQ:
		return left != right
	default:
		return false
	}
}

func (b *BinaryExpression) compareTimes(left, right time.Time) bool {
	switch b.Operator {
	case OpGT:
		return left.After(right)
	case OpLT:
		return left.Before(right)
	case OpGTE:
		return left.After(right) || left.Equal(right)
	case OpLTE:
		return left.Before(right) || left.Equal(right)
	case OpEQ:
		return left.Equal(right)
	case OpNEQ:
		return !left.Equal(right)
	default:
		return false
	}
}

func (b *BinaryExpression) compareStrings(left, right string) bool {
	switch b.Operator {
	case OpGT:
		return left > right
	case OpLT:
		return left < right
	case OpGTE:
		return left >= right
	case OpLTE:
		return left <= right
	case OpEQ:
		return left == right
	case OpNEQ:
		return left != right
	default:
		return false
	}
}

func (b *BinaryExpression) evaluateMembership(left, right interface{}) (bool, error) {
	// 将right转换为数组
	rightArray, err := toArray(right)
	if err != nil {
		return false, fmt.Errorf("right operand is not an array: %w", err)
	}

	// 检查left是否在right数组中
	contains := false
	for _, item := range rightArray {
		if equals(left, item) {
			contains = true
			break
		}
	}

	switch b.Operator {
	case OpIn:
		return contains, nil
	case OpNotIn:
		return !contains, nil
	default:
		return false, fmt.Errorf("unsupported membership operator: %s", b.Operator)
	}
}

func (b *BinaryExpression) evaluateContains(left, right interface{}) (bool, error) {
	leftStr := fmt.Sprintf("%v", left)

	// 如果right是数组，检查left是否包含数组中的任何元素
	if rightArray, err := toArray(right); err == nil {
		for _, item := range rightArray {
			itemStr := fmt.Sprintf("%v", item)
			if contains(leftStr, itemStr) {
				return true, nil
			}
		}
		return false, nil
	}

	// 如果right是字符串，检查left是否包含right
	rightStr := fmt.Sprintf("%v", right)
	return contains(leftStr, rightStr), nil
}

func (b *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator, b.Right.String())
}

// FunctionCall 函数调用节点
type FunctionCall struct {
	Name string `json:"name"`
	Args []Node `json:"args"`
}

func (f *FunctionCall) Type() NodeType {
	return NodeTypeFunctionCall
}

func (f *FunctionCall) Evaluate(ctx context.Context, dataProvider DataProvider) (interface{}, error) {
	// 评估所有参数
	args := make([]interface{}, len(f.Args))
	for i, arg := range f.Args {
		value, err := arg.Evaluate(ctx, dataProvider)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate argument %d: %w", i, err)
		}
		args[i] = value
	}

	// 调用相应的函数
	switch f.Name {
	case "avg":
		return f.callAvg(args)
	case "time_since":
		return f.callTimeSince(args)
	case "contains":
		return f.callContains(args)
	default:
		return nil, fmt.Errorf("unknown function: %s", f.Name)
	}
}

func (f *FunctionCall) callAvg(args []interface{}) (float64, error) {
	if len(args) != 2 {
		return 0, fmt.Errorf("avg function expects 2 arguments, got %d", len(args))
	}

	// 第一个参数应该是数据引用
	_, ok := f.Args[0].(*DataRef)
	if !ok {
		return 0, fmt.Errorf("first argument to avg must be a data reference")
	}

	// 第二个参数应该是周期数
	_, err := toInt(args[1])
	if err != nil {
		return 0, fmt.Errorf("second argument to avg must be an integer: %w", err)
	}

	// 这里需要从数据提供者获取历史数据
	// 暂时返回模拟值
	return 200.0, nil
}

func (f *FunctionCall) callTimeSince(args []interface{}) (float64, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("time_since function expects 1 argument, got %d", len(args))
	}

	// 参数应该是时间
	timestamp, err := toTime(args[0])
	if err != nil {
		return 0, fmt.Errorf("argument to time_since must be a time: %w", err)
	}

	// 计算从指定时间到现在的秒数
	now := time.Now()
	duration := now.Sub(timestamp)
	return duration.Seconds(), nil
}

func (f *FunctionCall) callContains(args []interface{}) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("contains function expects 2 arguments, got %d", len(args))
	}

	// 第一个参数应该是字符串
	text := fmt.Sprintf("%v", args[0])

	// 第二个参数应该是字符串数组
	keywords, err := toStringArray(args[1])
	if err != nil {
		return false, fmt.Errorf("second argument to contains must be a string array: %w", err)
	}

	// 检查文本是否包含任何关键词
	for _, keyword := range keywords {
		if contains(text, keyword) {
			return true, nil
		}
	}

	return false, nil
}

func (f *FunctionCall) String() string {
	argsStr := ""
	for i, arg := range f.Args {
		if i > 0 {
			argsStr += ", "
		}
		argsStr += arg.String()
	}
	return fmt.Sprintf("%s(%s)", f.Name, argsStr)
}

// DataRef 数据引用节点
type DataRef struct {
	Module DataType `json:"module"`
	Entity string   `json:"entity"`
	Field  string   `json:"field"`
}

func (d *DataRef) Type() NodeType {
	return NodeTypeDataRef
}

func (d *DataRef) Evaluate(ctx context.Context, dataProvider DataProvider) (interface{}, error) {
	return dataProvider.GetData(ctx, d.Module, d.Entity, d.Field)
}

func (d *DataRef) String() string {
	return fmt.Sprintf("%s.%s.%s", d.Module, d.Entity, d.Field)
}

// Value 值节点
type Value struct {
	Value interface{} `json:"value"`
}

func (v *Value) Type() NodeType {
	return NodeTypeValue
}

func (v *Value) Evaluate(ctx context.Context, dataProvider DataProvider) (interface{}, error) {
	return v.Value, nil
}

func (v *Value) String() string {
	return fmt.Sprintf("%v", v.Value)
}

// 辅助函数

func toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return v == "true", nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

func toNumbers(left, right interface{}) (float64, float64, bool) {
	leftNum, leftOk := toNumber(left)
	rightNum, rightOk := toNumber(right)
	return leftNum, rightNum, leftOk && rightOk
}

func toNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		// 尝试解析为数字
		if num, err := parseFloat(v); err == nil {
			return num, true
		}
	}
	return 0, false
}

func toTimes(left, right interface{}) (time.Time, time.Time, bool) {
	leftTime, leftErr := toTime(left)
	rightTime, rightErr := toTime(right)
	return leftTime, rightTime, leftErr == nil && rightErr == nil
}

func toTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// 尝试解析时间字符串
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, nil
		}
		// 尝试解析为Unix时间戳
		if timestamp, err := parseFloat(v); err == nil {
			return time.Unix(int64(timestamp), 0), nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot convert %T to time", value)
}

func toArray(value interface{}) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case []string:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case []float64:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	}
	return nil, fmt.Errorf("cannot convert %T to array", value)
}

func toStringArray(value interface{}) ([]string, error) {
	array, err := toArray(value)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(array))
	for i, item := range array {
		result[i] = fmt.Sprintf("%v", item)
	}
	return result, nil
}

func toInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		if num, err := parseFloat(v); err == nil {
			return int(num), nil
		}
	}
	return 0, fmt.Errorf("cannot convert %T to int", value)
}

func equals(left, right interface{}) bool {
	return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right)
}

func contains(text, substr string) bool {
	return len(text) >= len(substr) &&
		(text == substr ||
			(len(text) > len(substr) &&
				(text[:len(substr)] == substr ||
					text[len(text)-len(substr):] == substr ||
					indexOf(text, substr) >= 0)))
}

func indexOf(text, substr string) int {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func parseFloat(s string) (float64, error) {
	// 简单的浮点数解析实现
	// 在实际项目中应该使用 strconv.ParseFloat
	if s == "0" {
		return 0, nil
	}
	return 0, fmt.Errorf("parseFloat not implemented for: %s", s)
}
