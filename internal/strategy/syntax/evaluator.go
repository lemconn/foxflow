package syntax

import (
	"context"
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/strategy"
	"github.com/lemconn/foxflow/internal/strategy/datasources"
)

// Evaluator AST求值器
type Evaluator struct {
	registry *strategy.Registry
}

// NewEvaluator 创建AST求值器
func NewEvaluator(registry *strategy.Registry) *Evaluator {
	return &Evaluator{
		registry: registry,
	}
}

// Evaluate 执行AST节点
func (e *Evaluator) Evaluate(ctx context.Context, node *Node) (interface{}, error) {
	return node.Evaluate(ctx, e)
}

// EvaluateToBool 执行AST节点并返回布尔值
func (e *Evaluator) EvaluateToBool(ctx context.Context, node *Node) (bool, error) {
	result, err := e.Evaluate(ctx, node)
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

// GetFieldValue 获取字段值
func (e *Evaluator) GetFieldValue(ctx context.Context, module, entity, field string) (interface{}, error) {
	// 使用统一注册器获取数据
	return e.registry.GetData(ctx, module, entity, field)
}

// GetFieldValueWithParams 获取字段值（带参数）
func (e *Evaluator) GetFieldValueWithParams(ctx context.Context, module, entity, field string, params ...interface{}) (interface{}, error) {
	// 获取数据源模块
	ds, exists := e.registry.GetDataSource(module)
	if !exists {
		return nil, fmt.Errorf("data source not found: %s", module)
	}

	// 获取函数参数映射
	paramMapping := ds.GetFunctionParamMapping()

	// 将参数转换为 DataParam
	dataParams := make([]datasources.DataParam, 0)

	for _, param := range params {
		if funcNode, ok := param.(*Node); ok && funcNode.Type == NodeFuncCall {
			// 从函数调用节点中提取参数
			funcName := funcNode.FuncName

			// 查找该函数的参数映射
			if paramInfo, exists := paramMapping[funcName]; exists {
				// 处理多个参数
				for _, paramDef := range paramInfo.Params {
					// 检查参数索引是否有效
					if paramDef.ParamIndex < len(funcNode.Args) {
						paramNode := funcNode.Args[paramDef.ParamIndex]

						// 求值参数节点
						value, err := paramNode.Evaluate(ctx, e)
						if err != nil {
							return nil, fmt.Errorf("failed to evaluate parameter for function %s: %w", funcName, err)
						}

						// 根据参数类型转换值
						paramValue, err := datasources.ConvertParamValue(value, paramDef.ParamType)
						if err != nil {
							return nil, fmt.Errorf("failed to convert parameter %s for function %s: %w",
								paramDef.ParamName, funcName, err)
						}

						dataParams = append(dataParams, datasources.NewParam(paramDef.ParamName, paramValue))
					} else if paramDef.Required {
						// 如果参数是必需的但没有提供，使用默认值
						if paramDef.Default != nil {
							dataParams = append(dataParams, datasources.NewParam(paramDef.ParamName, paramDef.Default))
						} else {
							return nil, fmt.Errorf("function %s requires parameter %s at index %d but only %d arguments provided",
								funcName, paramDef.ParamName, paramDef.ParamIndex, len(funcNode.Args))
						}
					}
				}
			}
		}
	}

	// 使用统一注册器获取数据
	return e.registry.GetData(ctx, module, entity, field, dataParams...)
}

// CallFunction 调用函数
func (e *Evaluator) CallFunction(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	fn, exists := e.registry.GetFunction(name)
	if !exists {
		return nil, fmt.Errorf("unknown function: %s", name)
	}

	return fn.Execute(ctx, args, e)
}

// GetDataSource 获取数据源
func (e *Evaluator) GetDataSource(name string) (interface{}, bool) {
	return e.registry.GetDataSource(name)
}

// EvaluateBinary 评估二元表达式
func (e *Evaluator) EvaluateBinary(op string, left, right interface{}) (interface{}, error) {
	switch op {
	case "and":
		return e.evaluateLogicalAnd(left, right)
	case "or":
		return e.evaluateLogicalOr(left, right)
	case ">":
		return e.evaluateComparison(left, right, func(l, r float64) bool { return l > r })
	case "<":
		return e.evaluateComparison(left, right, func(l, r float64) bool { return l < r })
	case ">=":
		return e.evaluateComparison(left, right, func(l, r float64) bool { return l >= r })
	case "<=":
		return e.evaluateComparison(left, right, func(l, r float64) bool { return l <= r })
	case "==":
		return e.evaluateEquality(left, right)
	case "!=":
		return e.evaluateInequality(left, right)
	case "in":
		return e.evaluateMembership(left, right, true)
	case "not_in":
		return e.evaluateMembership(left, right, false)
	case "has":
		return e.evaluateContains(left, right)
	default:
		return nil, fmt.Errorf("unsupported operator: %s", op)
	}
}

// evaluateLogicalAnd 评估逻辑AND
func (e *Evaluator) evaluateLogicalAnd(left, right interface{}) (bool, error) {
	leftBool, err := toBool(left)
	if err != nil {
		return false, fmt.Errorf("left operand is not boolean: %w", err)
	}

	rightBool, err := toBool(right)
	if err != nil {
		return false, fmt.Errorf("right operand is not boolean: %w", err)
	}

	return leftBool && rightBool, nil
}

// evaluateLogicalOr 评估逻辑OR
func (e *Evaluator) evaluateLogicalOr(left, right interface{}) (bool, error) {
	leftBool, err := toBool(left)
	if err != nil {
		return false, fmt.Errorf("left operand is not boolean: %w", err)
	}

	rightBool, err := toBool(right)
	if err != nil {
		return false, fmt.Errorf("right operand is not boolean: %w", err)
	}

	return leftBool || rightBool, nil
}

// evaluateComparison 评估比较操作
func (e *Evaluator) evaluateComparison(left, right interface{}, compare func(float64, float64) bool) (bool, error) {
	// 尝试数字比较
	if leftNum, rightNum, ok := toNumbers(left, right); ok {
		return compare(leftNum, rightNum), nil
	}

	// 尝试时间比较
	if leftTime, rightTime, ok := toTimes(left, right); ok {
		switch {
		case compare(1, 0): // >
			return leftTime.After(rightTime), nil
		case compare(0, 1): // <
			return leftTime.Before(rightTime), nil
		case compare(1, 1): // >=
			return leftTime.After(rightTime) || leftTime.Equal(rightTime), nil
		case compare(0, 0): // <=
			return leftTime.Before(rightTime) || leftTime.Equal(rightTime), nil
		}
	}

	// 尝试字符串比较
	leftStr := toString(left)
	rightStr := toString(right)
	switch {
	case compare(1, 0): // >
		return leftStr > rightStr, nil
	case compare(0, 1): // <
		return leftStr < rightStr, nil
	case compare(1, 1): // >=
		return leftStr >= rightStr, nil
	case compare(0, 0): // <=
		return leftStr <= rightStr, nil
	}

	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

// evaluateEquality 评估相等性
func (e *Evaluator) evaluateEquality(left, right interface{}) (bool, error) {
	return equals(left, right), nil
}

// evaluateInequality 评估不等性
func (e *Evaluator) evaluateInequality(left, right interface{}) (bool, error) {
	return !equals(left, right), nil
}

// evaluateMembership 评估成员关系
func (e *Evaluator) evaluateMembership(left, right interface{}, isIn bool) (bool, error) {
	// 将right转换为数组
	rightArray, err := toStringArray(right)
	if err != nil {
		return false, fmt.Errorf("right operand is not an array: %w", err)
	}

	// 检查left是否在right数组中
	leftStr := toString(left)
	contains := false
	for _, item := range rightArray {
		if leftStr == item {
			contains = true
			break
		}
	}

	if isIn {
		return contains, nil
	}
	return !contains, nil
}

// evaluateContains 评估包含关系
func (e *Evaluator) evaluateContains(left, right interface{}) (bool, error) {
	leftStr := toString(left)
	rightStr := toString(right)

	// 检查left是否包含right
	return contains(leftStr, rightStr), nil
}

// Validate 验证AST节点
func (e *Evaluator) Validate(node *Node) error {
	return e.validateNode(node)
}

// validateNode 验证单个节点
func (e *Evaluator) validateNode(node *Node) error {
	switch node.Type {
	case NodeBinary:
		return e.validateBinaryExpression(node)
	case NodeFuncCall:
		return e.validateFunctionCall(node)
	case NodeIdent:
		return e.validateIdentifier(node)
	case NodeLiteral:
		return e.validateLiteral(node)
	case NodeFieldAccess:
		return e.validateFieldAccess(node)
	default:
		return fmt.Errorf("unknown node type: %d", node.Type)
	}
}

// validateBinaryExpression 验证二元表达式
func (e *Evaluator) validateBinaryExpression(node *Node) error {
	// 验证操作符
	if !e.isValidOperator(node.Op) {
		return fmt.Errorf("invalid operator: %s", node.Op)
	}

	// 验证左操作数
	if err := e.validateNode(node.Left); err != nil {
		return fmt.Errorf("invalid left operand: %w", err)
	}

	// 验证右操作数
	if err := e.validateNode(node.Right); err != nil {
		return fmt.Errorf("invalid right operand: %w", err)
	}

	return nil
}

// validateFunctionCall 验证函数调用
func (e *Evaluator) validateFunctionCall(node *Node) error {
	// 验证函数是否存在
	_, exists := e.registry.GetFunction(node.FuncName)
	if !exists {
		return fmt.Errorf("unknown function: %s", node.FuncName)
	}

	// 验证每个参数
	for i, arg := range node.Args {
		if err := e.validateNode(arg); err != nil {
			return fmt.Errorf("invalid argument %d: %w", i, err)
		}
	}

	return nil
}

// validateIdentifier 验证标识符
func (e *Evaluator) validateIdentifier(node *Node) error {
	// 检查标识符格式
	parts := strings.Split(node.Ident, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid identifier format: %s", node.Ident)
	}

	// 验证模块类型
	module := parts[0]
	if !e.isValidModule(module) {
		return fmt.Errorf("invalid module: %s", module)
	}

	return nil
}

// validateLiteral 验证字面量
func (e *Evaluator) validateLiteral(node *Node) error {
	if node.Value == nil {
		return fmt.Errorf("literal value cannot be nil")
	}
	return nil
}

// validateFieldAccess 验证字段访问
func (e *Evaluator) validateFieldAccess(node *Node) error {
	// 验证模块类型
	if !e.isValidModule(node.Module) {
		return fmt.Errorf("invalid module: %s", node.Module)
	}

	// 验证实体名不为空
	if node.Entity == "" {
		return fmt.Errorf("entity name cannot be empty")
	}

	// 验证字段名不为空
	if node.Field == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	return nil
}

// isValidOperator 检查操作符是否有效
func (e *Evaluator) isValidOperator(op string) bool {
	validOps := []string{
		"and", "or",
		">", "<", ">=", "<=", "==", "!=",
		"in", "not_in", "has",
	}

	for _, validOp := range validOps {
		if op == validOp {
			return true
		}
	}
	return false
}

// isValidModule 检查模块是否有效
func (e *Evaluator) isValidModule(module string) bool {
	// 使用统一注册器检查数据源是否存在
	_, exists := e.registry.GetDataSource(module)
	return exists
}
