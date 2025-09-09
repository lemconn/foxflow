package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// StrategyExpression 策略表达式
type StrategyExpression struct {
	StrategyName string
	Parameters   map[string]interface{}
	Operator     string // ">", "<", ">=", "<=", "==", "!="
	Value        interface{}
}

// StrategyCondition 策略条件
type StrategyCondition struct {
	Expressions []StrategyExpression
	Operators   []string // "and", "or"
}

// StrategyParser 策略解析器
type StrategyParser struct {
	expressionRegex *regexp.Regexp
	operatorRegex   *regexp.Regexp
}

// NewStrategyParser 创建策略解析器
func NewStrategyParser() *StrategyParser {
	return &StrategyParser{
		// 匹配策略表达式：strategy1:volume>100
		expressionRegex: regexp.MustCompile(`(\w+):(\w+)([><=!]+)([\d.]+)`),
		// 匹配操作符：and, or
		operatorRegex: regexp.MustCompile(`\b(and|or)\b`),
	}
}

// Parse 解析策略表达式
func (p *StrategyParser) Parse(expression string) (*StrategyCondition, error) {
	// 清理表达式
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// 提取所有策略表达式
	expressions := p.expressionRegex.FindAllStringSubmatch(expression, -1)
	if len(expressions) == 0 {
		return nil, fmt.Errorf("no valid strategy expressions found")
	}

	// 提取操作符
	operators := p.operatorRegex.FindAllString(expression, -1)

	// 构建策略条件
	condition := &StrategyCondition{
		Expressions: make([]StrategyExpression, 0, len(expressions)),
		Operators:   operators,
	}

	// 解析每个表达式
	for _, match := range expressions {
		if len(match) != 5 {
			continue
		}

		strategyName := match[1]
		parameter := match[2]
		operator := match[3]
		valueStr := match[4]

		// 解析数值
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value in expression: %s", valueStr)
		}

		// 验证操作符
		if !p.isValidOperator(operator) {
			return nil, fmt.Errorf("invalid operator: %s", operator)
		}

		expr := StrategyExpression{
			StrategyName: strategyName,
			Parameters: map[string]interface{}{
				parameter: value,
			},
			Operator: operator,
			Value:    value,
		}

		condition.Expressions = append(condition.Expressions, expr)
	}

	// 验证操作符数量
	if len(operators) != len(expressions)-1 {
		return nil, fmt.Errorf("invalid number of operators")
	}

	return condition, nil
}

// isValidOperator 验证操作符是否有效
func (p *StrategyParser) isValidOperator(op string) bool {
	validOps := []string{">", "<", ">=", "<=", "==", "!="}
	for _, validOp := range validOps {
		if op == validOp {
			return true
		}
	}
	return false
}

// Evaluate 评估策略条件
func (p *StrategyParser) Evaluate(condition *StrategyCondition, results map[string]bool) (bool, error) {
	if len(condition.Expressions) == 0 {
		return false, fmt.Errorf("no expressions to evaluate")
	}

	// 如果只有一个表达式
	if len(condition.Expressions) == 1 {
		expr := condition.Expressions[0]
		result, exists := results[expr.StrategyName]
		if !exists {
			return false, fmt.Errorf("strategy result not found: %s", expr.StrategyName)
		}
		return result, nil
	}

	// 多个表达式需要按操作符组合
	result := results[condition.Expressions[0].StrategyName]

	for i, operator := range condition.Operators {
		if i+1 >= len(condition.Expressions) {
			break
		}

		nextResult, exists := results[condition.Expressions[i+1].StrategyName]
		if !exists {
			return false, fmt.Errorf("strategy result not found: %s", condition.Expressions[i+1].StrategyName)
		}

		switch operator {
		case "and":
			result = result && nextResult
		case "or":
			result = result || nextResult
		default:
			return false, fmt.Errorf("unsupported operator: %s", operator)
		}
	}

	return result, nil
}

// GetStrategyNames 获取表达式中包含的策略名称
func (p *StrategyParser) GetStrategyNames(condition *StrategyCondition) []string {
	names := make([]string, 0, len(condition.Expressions))
	seen := make(map[string]bool)

	for _, expr := range condition.Expressions {
		if !seen[expr.StrategyName] {
			names = append(names, expr.StrategyName)
			seen[expr.StrategyName] = true
		}
	}

	return names
}

// GetParameters 获取策略参数
func (p *StrategyParser) GetParameters(condition *StrategyCondition) map[string]map[string]interface{} {
	params := make(map[string]map[string]interface{})

	for _, expr := range condition.Expressions {
		if params[expr.StrategyName] == nil {
			params[expr.StrategyName] = make(map[string]interface{})
		}

		// 合并参数
		for key, value := range expr.Parameters {
			params[expr.StrategyName][key] = value
		}
	}

	return params
}
