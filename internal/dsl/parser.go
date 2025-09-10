package dsl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/ast"
)

// Parser DSL解析器
type Parser struct {
	// 正则表达式
	fieldAccessRegex *regexp.Regexp
	operatorRegex    *regexp.Regexp
	parenthesesRegex *regexp.Regexp
	stringArrayRegex *regexp.Regexp
	functionRegex    *regexp.Regexp

	// 占位符映射（用于括号处理）
	placeholders map[string]ast.Node
}

// NewParser 创建DSL解析器
func NewParser() *Parser {
	return &Parser{
		// 匹配字段访问：candles.SOL.last_px, news.theblockbeats.last_title（不匹配函数调用）
		fieldAccessRegex: regexp.MustCompile(`^(\w+)\.(\w+)\.(\w+)(?:\.(\w+))?$`),
		// 匹配操作符：and, or
		operatorRegex: regexp.MustCompile(`\b(and|or)\b`),
		// 匹配括号
		parenthesesRegex: regexp.MustCompile(`\(([^()]+)\)`),
		// 匹配字符串数组：["突破新高", "重大利好"]
		stringArrayRegex: regexp.MustCompile(`\[([^\]]+)\]`),
		// 匹配函数调用：avg(candles.BTC.close, 5)
		functionRegex: regexp.MustCompile(`(\w+)\(([^)]+)\)`),
		// 初始化占位符映射
		placeholders: make(map[string]ast.Node),
	}
}

// Parse 解析DSL表达式为AST
func (p *Parser) Parse(expression string) (ast.Node, error) {
	// 清理表达式
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// 重置占位符映射
	p.placeholders = make(map[string]ast.Node)

	// 解析表达式
	return p.parseExpression(expression)
}

// parseExpression 解析表达式（支持括号嵌套）
func (p *Parser) parseExpression(expression string) (ast.Node, error) {
	// 首先处理括号
	parentheses := p.parenthesesRegex.FindAllStringSubmatch(expression, -1)
	if len(parentheses) > 0 {
		// 有括号，需要递归处理
		return p.parseWithParentheses(expression)
	}

	// 没有括号，直接解析
	return p.parseSimpleExpression(expression)
}

// parseWithParentheses 解析带括号的表达式
func (p *Parser) parseWithParentheses(expression string) (ast.Node, error) {
	// 找到最外层的括号
	start := strings.Index(expression, "(")
	if start == -1 {
		return p.parseSimpleExpression(expression)
	}

	// 找到匹配的右括号
	level := 0
	end := -1
	for i := start; i < len(expression); i++ {
		if expression[i] == '(' {
			level++
		} else if expression[i] == ')' {
			level--
			if level == 0 {
				end = i
				break
			}
		}
	}

	if end == -1 {
		return nil, fmt.Errorf("unmatched parentheses")
	}

	// 检查是否是函数调用
	if start > 0 {
		beforeParen := strings.TrimSpace(expression[:start])
		// 如果括号前有标识符，可能是函数调用
		if p.isValidFunctionName(beforeParen) {
			// 检查函数调用后面是否还有操作符
			afterParen := strings.TrimSpace(expression[end+1:])
			if afterParen == "" {
				// 没有操作符，直接返回函数调用
				return p.parseFunctionCall(expression)
			} else {
				// 有操作符，需要进一步解析
				// 这里需要特殊处理，因为函数调用后面可能有操作符
				return p.parseExpressionWithFunctionCall(expression, start, end)
			}
		}
	}

	// 提取括号内的内容
	innerExpr := expression[start+1 : end]
	innerNode, err := p.parseExpression(innerExpr)
	if err != nil {
		return nil, err
	}

	// 处理括号前后的内容
	before := strings.TrimSpace(expression[:start])
	after := strings.TrimSpace(expression[end+1:])

	// 如果整个表达式都在括号内
	if before == "" && after == "" {
		return innerNode, nil
	}

	// 重新构建表达式，将括号内容替换为占位符
	placeholder := fmt.Sprintf("PLACEHOLDER_%d", len(p.placeholders))
	p.placeholders[placeholder] = innerNode

	// 构建新的表达式
	newExpr := before + " " + placeholder + " " + after
	newExpr = strings.TrimSpace(newExpr)

	// 递归解析新表达式
	return p.parseExpression(newExpr)
}

// parseSimpleExpression 解析简单表达式（无括号）
func (p *Parser) parseSimpleExpression(expression string) (ast.Node, error) {
	// 检查是否有 or 操作符
	if strings.Contains(expression, " or ") {
		return p.parseOrExpression(expression)
	}

	// 检查是否有 and 操作符
	if strings.Contains(expression, " and ") {
		return p.parseAndExpression(expression)
	}

	// 检查是否有比较操作符
	if p.containsOperator(expression) {
		return p.parseComparisonExpression(expression)
	}

	// 单个表达式
	return p.parseSingleExpression(strings.TrimSpace(expression))
}

// parseOrExpression 解析OR表达式
func (p *Parser) parseOrExpression(expression string) (ast.Node, error) {
	parts := strings.Split(expression, " or ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid or expression")
	}

	// 解析第一个部分
	left, err := p.parseAndExpression(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse left part of or expression: %w", err)
	}

	// 如果只有一个OR部分，直接返回
	if len(parts) == 2 {
		right, err := p.parseAndExpression(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("failed to parse right part of or expression: %w", err)
		}

		return &ast.BinaryExpression{
			Operator: ast.OpOr,
			Left:     left,
			Right:    right,
		}, nil
	}

	// 解析剩余部分
	right, err := p.parseOrExpression(strings.Join(parts[1:], " or "))
	if err != nil {
		return nil, fmt.Errorf("failed to parse right part of or expression: %w", err)
	}

	return &ast.BinaryExpression{
		Operator: ast.OpOr,
		Left:     left,
		Right:    right,
	}, nil
}

// parseAndExpression 解析AND表达式
func (p *Parser) parseAndExpression(expression string) (ast.Node, error) {
	parts := strings.Split(expression, " and ")
	if len(parts) < 2 {
		// 单个表达式
		return p.parseSingleExpression(strings.TrimSpace(expression))
	}

	// 解析第一个部分
	left, err := p.parseSingleExpression(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("failed to parse left part of and expression: %w", err)
	}

	// 解析剩余部分
	right, err := p.parseAndExpression(strings.Join(parts[1:], " and "))
	if err != nil {
		return nil, fmt.Errorf("failed to parse right part of and expression: %w", err)
	}

	return &ast.BinaryExpression{
		Operator: ast.OpAnd,
		Left:     left,
		Right:    right,
	}, nil
}

// parseExpressionWithFunctionCall 解析包含函数调用的表达式
func (p *Parser) parseExpressionWithFunctionCall(expression string, start, end int) (ast.Node, error) {
	// 解析函数调用部分
	funcCallExpr := expression[:end+1]
	funcCallNode, err := p.parseFunctionCall(funcCallExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse function call: %w", err)
	}

	// 获取函数调用后的部分
	afterFunc := strings.TrimSpace(expression[end+1:])
	if afterFunc == "" {
		return funcCallNode, nil
	}

	// 将函数调用替换为占位符，然后解析整个表达式
	placeholder := fmt.Sprintf("PLACEHOLDER_%d", len(p.placeholders))
	p.placeholders[placeholder] = funcCallNode

	// 构建新的表达式
	newExpr := placeholder + " " + afterFunc
	newExpr = strings.TrimSpace(newExpr)

	// 递归解析新表达式
	return p.parseExpression(newExpr)
}

// parseComparisonExpression 解析比较表达式
func (p *Parser) parseComparisonExpression(expression string) (ast.Node, error) {
	// 支持的比较操作符（按优先级排序）
	operators := []string{">=", "<=", "==", "!=", ">", "<", " in ", " not_in ", " contains "}

	for _, op := range operators {
		if strings.Contains(expression, op) {
			parts := strings.Split(expression, op)
			if len(parts) == 2 {
				left := strings.TrimSpace(parts[0])
				right := strings.TrimSpace(parts[1])

				// 解析左操作数
				leftNode, err := p.parseSingleExpression(left)
				if err != nil {
					return nil, fmt.Errorf("failed to parse left operand: %w", err)
				}

				// 解析右操作数
				rightNode, err := p.parseSingleExpression(right)
				if err != nil {
					return nil, fmt.Errorf("failed to parse right operand: %w", err)
				}

				return &ast.BinaryExpression{
					Operator: ast.Operator(strings.TrimSpace(op)),
					Left:     leftNode,
					Right:    rightNode,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("no valid comparison operator found in expression: %s", expression)
}

// parseSingleExpression 解析单个表达式
func (p *Parser) parseSingleExpression(expression string) (ast.Node, error) {
	// 检查是否是占位符
	if strings.HasPrefix(expression, "PLACEHOLDER_") {
		if placeholder, exists := p.placeholders[expression]; exists {
			return placeholder, nil
		} else {
			return nil, fmt.Errorf("placeholder not found: %s", expression)
		}
	}

	// 检查是否是函数调用（优先于字段访问）
	if p.functionRegex.MatchString(expression) {
		return p.parseFunctionCall(expression)
	}

	// 检查是否是字段访问
	if p.fieldAccessRegex.MatchString(expression) {
		return p.parseFieldAccess(expression)
	}

	// 尝试解析为值
	return p.parseValue(expression)
}

// parseFunctionCall 解析函数调用
func (p *Parser) parseFunctionCall(expression string) (ast.Node, error) {
	matches := p.functionRegex.FindStringSubmatch(expression)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid function call format: %s", expression)
	}

	funcName := matches[1]
	argsStr := matches[2]

	// 解析参数
	args, err := p.parseFunctionArgs(argsStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse function arguments: %w", err)
	}

	return &ast.FunctionCall{
		Name: funcName,
		Args: args,
	}, nil
}

// parseFunctionArgs 解析函数参数
func (p *Parser) parseFunctionArgs(argsStr string) ([]ast.Node, error) {
	// 简单的参数分割（假设用逗号分隔）
	parts := strings.Split(argsStr, ",")
	args := make([]ast.Node, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		arg, err := p.parseSingleExpression(part)
		if err != nil {
			return nil, fmt.Errorf("failed to parse argument '%s': %w", part, err)
		}

		args = append(args, arg)
	}

	return args, nil
}

// parseFieldAccess 解析字段访问
func (p *Parser) parseFieldAccess(expression string) (ast.Node, error) {
	// 解析字段访问部分
	fieldAccessMatch := p.fieldAccessRegex.FindStringSubmatch(expression)
	if len(fieldAccessMatch) < 4 {
		return nil, fmt.Errorf("invalid field access pattern: %s", expression)
	}

	module := fieldAccessMatch[1]
	entity := fieldAccessMatch[2]
	field := fieldAccessMatch[3]

	// 验证模块类型
	dataType, err := p.parseDataType(module)
	if err != nil {
		return nil, err
	}

	// 创建数据引用
	dataRef := &ast.DataRef{
		Module: dataType,
		Entity: entity,
		Field:  field,
	}

	// 检查是否包含操作符，如果没有操作符，直接返回数据引用
	if !p.containsOperator(expression) {
		return dataRef, nil
	}

	// 提取操作符和值
	operator, value, values, err := p.extractOperatorAndValue(expression)
	if err != nil {
		return nil, err
	}

	// 创建右操作数
	var right ast.Node
	if values != nil {
		// 数组值
		right = &ast.Value{Value: values}
	} else {
		// 单个值
		right = &ast.Value{Value: value}
	}

	// 创建二元表达式
	return &ast.BinaryExpression{
		Operator: ast.Operator(operator),
		Left:     dataRef,
		Right:    right,
	}, nil
}

// parseValue 解析值
func (p *Parser) parseValue(valueStr string) (ast.Node, error) {
	// 尝试解析为数字
	if num, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return &ast.Value{Value: num}, nil
	}

	// 尝试解析为整数
	if num, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return &ast.Value{Value: float64(num)}, nil
	}

	// 尝试解析为布尔值
	if valueStr == "true" {
		return &ast.Value{Value: true}, nil
	}
	if valueStr == "false" {
		return &ast.Value{Value: false}, nil
	}

	// 尝试解析为字符串数组
	if p.stringArrayRegex.MatchString(valueStr) {
		values, err := p.parseStringArray(valueStr)
		if err != nil {
			return nil, err
		}
		return &ast.Value{Value: values}, nil
	}

	// 作为字符串处理（去除引号）
	if len(valueStr) >= 2 && valueStr[0] == '"' && valueStr[len(valueStr)-1] == '"' {
		return &ast.Value{Value: valueStr[1 : len(valueStr)-1]}, nil
	}

	// 作为字符串返回
	return &ast.Value{Value: valueStr}, nil
}

// parseStringArray 解析字符串数组
func (p *Parser) parseStringArray(valueStr string) ([]interface{}, error) {
	matches := p.stringArrayRegex.FindStringSubmatch(valueStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid array format: %s", valueStr)
	}

	content := matches[1]
	if content == "" {
		return []interface{}{}, nil
	}

	// 简单的字符串分割（假设用逗号分隔）
	parts := strings.Split(content, ",")
	values := make([]interface{}, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 去除引号
		if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
			part = part[1 : len(part)-1]
		}

		values = append(values, part)
	}

	return values, nil
}

// extractOperatorAndValue 提取操作符和值
func (p *Parser) extractOperatorAndValue(expression string) (string, interface{}, []interface{}, error) {
	// 支持的比较操作符
	comparisonOps := []string{">=", "<=", "==", "!=", ">", "<"}

	// 支持的包含操作符
	containmentOps := []string{" in ", " not_in "}

	// 支持的包含操作符（用于字符串）
	containsOps := []string{" contains "}

	// 检查包含操作符
	for _, op := range containmentOps {
		if strings.Contains(expression, op) {
			parts := strings.Split(expression, op)
			if len(parts) != 2 {
				continue
			}

			// 解析值数组
			values, err := p.parseStringArray(strings.TrimSpace(parts[1]))
			if err != nil {
				continue
			}

			operator := strings.TrimSpace(op)
			return operator, nil, values, nil
		}
	}

	// 检查字符串包含操作符
	for _, op := range containsOps {
		if strings.Contains(expression, op) {
			parts := strings.Split(expression, op)
			if len(parts) != 2 {
				continue
			}

			// 解析值数组
			values, err := p.parseStringArray(strings.TrimSpace(parts[1]))
			if err != nil {
				continue
			}

			return "contains", nil, values, nil
		}
	}

	// 检查比较操作符
	for _, op := range comparisonOps {
		if strings.Contains(expression, op) {
			parts := strings.Split(expression, op)
			if len(parts) != 2 {
				continue
			}

			value, err := p.parseValue(strings.TrimSpace(parts[1]))
			if err != nil {
				continue
			}

			// 提取值节点的值
			if valueNode, ok := value.(*ast.Value); ok {
				return op, valueNode.Value, nil, nil
			}
		}
	}

	// 如果没有找到操作符，检查是否是简单的字段访问
	if p.fieldAccessRegex.MatchString(expression) {
		return "", nil, nil, nil
	}

	return "", nil, nil, fmt.Errorf("no valid operator found in expression: %s", expression)
}

// parseDataType 解析数据类型
func (p *Parser) parseDataType(module string) (ast.DataType, error) {
	switch module {
	case "candles":
		return ast.DataTypeCandles, nil
	case "news":
		return ast.DataTypeNews, nil
	case "indicators":
		return ast.DataTypeIndicators, nil
	default:
		return "", fmt.Errorf("unsupported data type: %s", module)
	}
}

// isValidFunctionName 检查是否是有效的函数名
func (p *Parser) isValidFunctionName(name string) bool {
	validFunctions := []string{"avg", "time_since", "contains"}

	for _, validFunc := range validFunctions {
		if name == validFunc {
			return true
		}
	}

	return false
}

// containsOperator 检查表达式是否包含操作符
func (p *Parser) containsOperator(expression string) bool {
	// 支持的比较操作符
	comparisonOps := []string{">=", "<=", "==", "!=", ">", "<"}

	// 支持的包含操作符
	containmentOps := []string{" in ", " not_in "}

	// 支持的包含操作符（用于字符串）
	containsOps := []string{" contains "}

	// 检查所有操作符
	allOps := append(comparisonOps, containmentOps...)
	allOps = append(allOps, containsOps...)

	for _, op := range allOps {
		if strings.Contains(expression, op) {
			return true
		}
	}

	return false
}

// Validate 验证DSL表达式
func (p *Parser) Validate(expression string) error {
	// 尝试解析表达式
	_, err := p.Parse(expression)
	return err
}
