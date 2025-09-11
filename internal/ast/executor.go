package ast

import (
	"context"
	"fmt"
)

// Executor AST执行器
type Executor struct {
	dataProvider DataProvider
}

// NewExecutor 创建AST执行器
func NewExecutor(dataProvider DataProvider) *Executor {
	return &Executor{
		dataProvider: dataProvider,
	}
}

// Execute 执行AST节点
func (e *Executor) Execute(ctx context.Context, node Node) (interface{}, error) {
	return node.Evaluate(ctx, e.dataProvider)
}

// ExecuteToBool 执行AST节点并返回布尔值
func (e *Executor) ExecuteToBool(ctx context.Context, node Node) (bool, error) {
	result, err := e.Execute(ctx, node)
	if err != nil {
		return false, err
	}

	// 尝试转换为布尔值
	switch v := result.(type) {
	case bool:
		return v, nil
	case string:
		return v == "true", nil
	case float64:
		return v != 0, nil
	case int:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert result %T to boolean", result)
	}
}

// Validate 验证AST节点
func (e *Executor) Validate(node Node) error {
	return e.validateNode(node)
}

// validateNode 验证单个节点
func (e *Executor) validateNode(node Node) error {
	switch n := node.(type) {
	case *BinaryExpression:
		return e.validateBinaryExpression(n)
	case *FunctionCall:
		return e.validateFunctionCall(n)
	case *DataRef:
		return e.validateDataRef(n)
	case *Value:
		return e.validateValue(n)
	default:
		return fmt.Errorf("unknown node type: %T", node)
	}
}

// validateBinaryExpression 验证二元表达式
func (e *Executor) validateBinaryExpression(expr *BinaryExpression) error {
	// 验证操作符
	if !isValidOperator(expr.Operator) {
		return fmt.Errorf("invalid operator: %s", expr.Operator)
	}

	// 验证左操作数
	if err := e.validateNode(expr.Left); err != nil {
		return fmt.Errorf("invalid left operand: %w", err)
	}

	// 验证右操作数
	if err := e.validateNode(expr.Right); err != nil {
		return fmt.Errorf("invalid right operand: %w", err)
	}

	// 验证操作符和操作数的兼容性
	return e.validateOperatorCompatibility(expr.Operator, expr.Left, expr.Right)
}

// validateFunctionCall 验证函数调用
func (e *Executor) validateFunctionCall(call *FunctionCall) error {
	// 验证函数名
	if !isValidFunction(call.Name) {
		return fmt.Errorf("unknown function: %s", call.Name)
	}

	// 验证参数数量
	expectedArgs := getFunctionArgCount(call.Name)
	if len(call.Args) != expectedArgs {
		return fmt.Errorf("function %s expects %d arguments, got %d", call.Name, expectedArgs, len(call.Args))
	}

	// 验证每个参数
	for i, arg := range call.Args {
		if err := e.validateNode(arg); err != nil {
			return fmt.Errorf("invalid argument %d: %w", i, err)
		}
	}

	return nil
}

// validateDataRef 验证数据引用
func (e *Executor) validateDataRef(ref *DataRef) error {
	// 验证模块类型
	if !isValidDataType(ref.Module) {
		return fmt.Errorf("invalid data type: %s", ref.Module)
	}

	// 验证实体名不为空
	if ref.Entity == "" {
		return fmt.Errorf("entity name cannot be empty")
	}

	// 验证字段名不为空
	if ref.Field == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	return nil
}

// validateValue 验证值节点
func (e *Executor) validateValue(value *Value) error {
	if value.Value == nil {
		return fmt.Errorf("value cannot be nil")
	}
	return nil
}

// validateOperatorCompatibility 验证操作符和操作数的兼容性
func (e *Executor) validateOperatorCompatibility(op Operator, left, right Node) error {
	switch op {
	case OpAnd, OpOr:
		// 逻辑操作符要求操作数能够转换为布尔值
		// 这里只做基本检查，实际类型检查在运行时进行
		return nil

	case OpGT, OpLT, OpGTE, OpLTE, OpEQ, OpNEQ:
		// 比较操作符要求操作数能够比较
		// 这里只做基本检查，实际类型检查在运行时进行
		return nil

	case OpIn, OpNotIn:
		// 成员操作符要求右操作数是数组
		if right.Type() != NodeTypeValue {
			return fmt.Errorf("right operand of %s must be a value (array)", op)
		}
		return nil

	case OpContains:
		// 包含操作符要求操作数是字符串
		// 这里只做基本检查，实际类型检查在运行时进行
		return nil

	default:
		return fmt.Errorf("unsupported operator: %s", op)
	}
}

// isValidOperator 检查操作符是否有效
func isValidOperator(op Operator) bool {
	validOps := []Operator{
		OpAnd, OpOr,
		OpGT, OpLT, OpGTE, OpLTE, OpEQ, OpNEQ,
		OpIn, OpNotIn, OpContains,
	}

	for _, validOp := range validOps {
		if op == validOp {
			return true
		}
	}
	return false
}

// isValidFunction 检查函数名是否有效
func isValidFunction(name string) bool {
	validFunctions := []string{
		"avg", "time_since", "contains",
	}

	for _, validFunc := range validFunctions {
		if name == validFunc {
			return true
		}
	}
	return false
}

// isValidDataType 检查数据类型是否有效
func isValidDataType(dataType DataType) bool {
	validTypes := []DataType{
		DataTypeCandles, DataTypeNews, DataTypeIndicators,
	}

	for _, validType := range validTypes {
		if dataType == validType {
			return true
		}
	}
	return false
}

// getFunctionArgCount 获取函数期望的参数数量
func getFunctionArgCount(name string) int {
	switch name {
	case "avg":
		return 2
	case "time_since":
		return 1
	case "contains":
		return 2
	default:
		return 0
	}
}
