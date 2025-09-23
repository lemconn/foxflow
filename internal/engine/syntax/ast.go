package syntax

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NodeType AST节点类型
type NodeType int

const (
	NodeBinary NodeType = iota
	NodeUnary
	NodeLiteral
	NodeIdent
	NodeFuncCall
	NodeFieldAccess
)

// Node AST节点
type Node struct {
	Type NodeType

	// 二元表达式
	Op    string
	Left  *Node
	Right *Node

	// 一元表达式
	Value interface{}

	// 标识符
	Ident string

	// 函数调用
	FuncName string
	Args     []*Node

	// 字段访问
	Module string
	Entity string
	Field  string

	// 上下文信息（用于传递函数参数）
	Parent *Node
}

// String 返回节点的字符串表示
func (n *Node) String() string {
	switch n.Type {
	case NodeBinary:
		return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Op, n.Right.String())
	case NodeUnary:
		return fmt.Sprintf("(%s %v)", n.Op, n.Value)
	case NodeLiteral:
		switch v := n.Value.(type) {
		case string:
			return fmt.Sprintf("\"%s\"", v)
		case []string:
			return fmt.Sprintf("[%s]", strings.Join(v, ", "))
		default:
			return fmt.Sprintf("%v", v)
		}
	case NodeIdent:
		return n.Ident
	case NodeFuncCall:
		args := make([]string, len(n.Args))
		for i, arg := range n.Args {
			args[i] = arg.String()
		}
		return fmt.Sprintf("%s(%s)", n.FuncName, strings.Join(args, ", "))
	case NodeFieldAccess:
		return fmt.Sprintf("%s.%s.%s", n.Module, n.Entity, n.Field)
	default:
		return "UNKNOWN"
	}
}

// Evaluate 执行节点并返回结果
func (n *Node) Evaluate(ctx context.Context, evaluator *Evaluator) (interface{}, error) {
	switch n.Type {
	case NodeLiteral:
		return n.Value, nil

	case NodeIdent:
		// 标识符需要解析为字段访问
		parts := strings.Split(n.Ident, ".")
		if len(parts) == 3 {
			// 创建字段访问节点
			fieldNode := &Node{
				Type:   NodeFieldAccess,
				Module: parts[0],
				Entity: parts[1],
				Field:  parts[2],
			}
			return fieldNode.Evaluate(ctx, evaluator)
		}
		return nil, fmt.Errorf("invalid identifier: %s", n.Ident)

	case NodeFieldAccess:
		// 检查是否在函数调用中，如果是，传递函数参数
		if n.isInFunctionCall() {
			funcNode := n.getFunctionCallNode()
			if funcNode != nil {
				// 提取函数参数作为数据源参数
				params := n.extractDataSourceParams(funcNode)
				return evaluator.GetFieldValueWithParams(ctx, n.Module, n.Entity, n.Field, params...)
			}
		}
		return evaluator.GetFieldValue(ctx, n.Module, n.Entity, n.Field)

	case NodeBinary:
		return n.evaluateBinary(ctx, evaluator)

	case NodeFuncCall:
		return n.evaluateFunction(ctx, evaluator)

	default:
		return nil, fmt.Errorf("unsupported node type: %d", n.Type)
	}
}

// evaluateBinary 评估二元表达式
func (n *Node) evaluateBinary(ctx context.Context, evaluator *Evaluator) (interface{}, error) {
	left, err := n.Left.Evaluate(ctx, evaluator)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate left operand: %w", err)
	}

	right, err := n.Right.Evaluate(ctx, evaluator)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate right operand: %w", err)
	}

	return evaluator.EvaluateBinary(n.Op, left, right)
}

// evaluateFunction 评估函数调用
func (n *Node) evaluateFunction(ctx context.Context, evaluator *Evaluator) (interface{}, error) {
	// 评估所有参数
	args := make([]interface{}, len(n.Args))
	for i, arg := range n.Args {
		value, err := arg.Evaluate(ctx, evaluator)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate argument %d: %w", i, err)
		}
		args[i] = value
	}

	// 调用函数
	return evaluator.CallFunction(ctx, n.FuncName, args)
}

// 辅助函数

// toFloat64 转换为float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// toBool 转换为bool
func toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return v == "true", nil
	case float64:
		return v != 0, nil
	case int:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// toString 转换为string
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, ",")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toTime 转换为time.Time
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
		if timestamp, err := strconv.ParseFloat(v, 64); err == nil {
			return time.Unix(int64(timestamp), 0), nil
		}
		return time.Time{}, fmt.Errorf("cannot parse time from string: %s", v)
	case float64:
		return time.Unix(int64(v), 0), nil
	case int64:
		return time.Unix(v, 0), nil
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time", value)
	}
}

// toStringArray 转换为字符串数组
func toStringArray(value interface{}) ([]string, error) {
	switch v := value.(type) {
	case []string:
		return v, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = toString(item)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []string", value)
	}
}

// equals 比较两个值是否相等
func equals(left, right interface{}) bool {
	// 尝试数字比较
	if leftNum, rightNum, ok := toNumbers(left, right); ok {
		return leftNum == rightNum
	}

	// 尝试时间比较
	if leftTime, rightTime, ok := toTimes(left, right); ok {
		return leftTime.Equal(rightTime)
	}

	// 字符串比较
	return toString(left) == toString(right)
}

// toNumbers 尝试将两个值转换为数字
func toNumbers(left, right interface{}) (float64, float64, bool) {
	leftNum, leftErr := toFloat64(left)
	rightNum, rightErr := toFloat64(right)
	return leftNum, rightNum, leftErr == nil && rightErr == nil
}

// toTimes 尝试将两个值转换为时间
func toTimes(left, right interface{}) (time.Time, time.Time, bool) {
	leftTime, leftErr := toTime(left)
	rightTime, rightErr := toTime(right)
	return leftTime, rightTime, leftErr == nil && rightErr == nil
}

// contains 检查字符串是否包含子字符串
func contains(text, substr string) bool {
	return strings.Contains(text, substr)
}

// isInFunctionCall 检查当前节点是否在函数调用中
func (n *Node) isInFunctionCall() bool {
	current := n.Parent
	for current != nil {
		if current.Type == NodeFuncCall {
			return true
		}
		current = current.Parent
	}
	return false
}

// getFunctionCallNode 获取包含当前节点的函数调用节点
func (n *Node) getFunctionCallNode() *Node {
	current := n.Parent
	for current != nil {
		if current.Type == NodeFuncCall {
			return current
		}
		current = current.Parent
	}
	return nil
}

// extractDataSourceParams 从函数调用中提取数据源参数
func (n *Node) extractDataSourceParams(funcNode *Node) []interface{} {
	// 这里不再硬编码函数名称，而是通过数据源模块的映射来获取参数
	// 具体的参数提取逻辑将在 Evaluator 中实现
	return []interface{}{funcNode}
}
