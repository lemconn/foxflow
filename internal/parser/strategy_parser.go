package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DataSource 数据源类型
type DataSource string

const (
	DataSourceCandles DataSource = "candles"
	DataSourceNews    DataSource = "news"
)

// FieldAccess 字段访问
type FieldAccess struct {
	Source   DataSource `json:"source"`    // 数据源：candles, news
	Symbol   string     `json:"symbol"`    // 交易对符号，如 SOL
	Field    string     `json:"field"`     // 字段名，如 last_px, last_volume
	SubField string     `json:"sub_field"` // 子字段，如 last_title, last_update_time
}

// StrategyExpression 策略表达式
type StrategyExpression struct {
	FieldAccess FieldAccess            `json:"field_access"`
	Operator    string                 `json:"operator"` // ">", "<", ">=", "<=", "==", "!=", "in", "not_in"
	Value       interface{}            `json:"value"`
	Values      []interface{}          `json:"values"` // 用于 in/not_in 操作
	Parameters  map[string]interface{} `json:"parameters"`
}

// StrategyCondition 策略条件
type StrategyCondition struct {
	Expressions   []StrategyExpression `json:"expressions"`
	Operators     []string             `json:"operators"`      // "and", "or"
	SubConditions []*StrategyCondition `json:"sub_conditions"` // 支持嵌套条件
}

// StrategyParser 策略解析器
type StrategyParser struct {
	// 匹配复杂表达式：candles.SOL.last_px >= 200
	fieldAccessRegex *regexp.Regexp
	// 匹配操作符：and, or
	operatorRegex *regexp.Regexp
	// 匹配括号
	parenthesesRegex *regexp.Regexp
	// 匹配字符串数组：["突破新高", "重大利好"]
	stringArrayRegex *regexp.Regexp
	// 占位符映射（用于括号处理）
	placeholders map[string]*StrategyCondition
}

// NewStrategyParser 创建策略解析器
func NewStrategyParser() *StrategyParser {
	return &StrategyParser{
		// 匹配字段访问：candles.SOL.last_px, news.theblockbeats.last_title
		fieldAccessRegex: regexp.MustCompile(`(\w+)\.(\w+)\.(\w+)(?:\.(\w+))?`),
		// 匹配操作符：and, or
		operatorRegex: regexp.MustCompile(`\b(and|or)\b`),
		// 匹配括号
		parenthesesRegex: regexp.MustCompile(`\(([^()]+)\)`),
		// 匹配字符串数组：["突破新高", "重大利好"]
		stringArrayRegex: regexp.MustCompile(`\[([^\]]+)\]`),
		// 初始化占位符映射
		placeholders: make(map[string]*StrategyCondition),
	}
}

// Parse 解析策略表达式
func (p *StrategyParser) Parse(expression string) (*StrategyCondition, error) {
	// 清理表达式
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// 处理括号嵌套
	return p.parseExpression(expression)
}

// parseExpression 解析表达式（支持括号嵌套）
func (p *StrategyParser) parseExpression(expression string) (*StrategyCondition, error) {
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
func (p *StrategyParser) parseWithParentheses(expression string) (*StrategyCondition, error) {
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

	// 提取括号内的内容
	innerExpr := expression[start+1 : end]
	innerCondition, err := p.parseExpression(innerExpr)
	if err != nil {
		return nil, err
	}

	// 处理括号前后的内容
	before := strings.TrimSpace(expression[:start])
	after := strings.TrimSpace(expression[end+1:])

	// 如果整个表达式都在括号内
	if before == "" && after == "" {
		return innerCondition, nil
	}

	// 重新构建表达式，将括号内容替换为占位符
	// 例如：(A and B) or (C and D) -> PLACEHOLDER_0 or PLACEHOLDER_1
	placeholder := fmt.Sprintf("PLACEHOLDER_%d", len(p.placeholders))
	p.placeholders[placeholder] = innerCondition

	// 构建新的表达式
	newExpr := before + " " + placeholder + " " + after
	newExpr = strings.TrimSpace(newExpr)

	// 递归解析新表达式
	return p.parseExpression(newExpr)
}

// parseSimpleExpression 解析简单表达式（无括号）
func (p *StrategyParser) parseSimpleExpression(expression string) (*StrategyCondition, error) {
	// 检查是否有 or 操作符
	if strings.Contains(expression, " or ") {
		return p.parseOrExpression(expression)
	}

	// 检查是否有 and 操作符
	if strings.Contains(expression, " and ") {
		return p.parseAndExpression(expression)
	}

	// 单个表达式
	expr, err := p.parseSingleExpression(strings.TrimSpace(expression))
	if err != nil {
		return nil, fmt.Errorf("failed to parse single expression '%s': %w", expression, err)
	}

	return &StrategyCondition{
		Expressions: []StrategyExpression{*expr},
		Operators:   []string{},
	}, nil
}

// parseOrExpression 解析OR表达式
func (p *StrategyParser) parseOrExpression(expression string) (*StrategyCondition, error) {
	parts := strings.Split(expression, " or ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid or expression")
	}

	condition := &StrategyCondition{
		Expressions:   make([]StrategyExpression, 0),
		Operators:     make([]string, 0, len(parts)-1),
		SubConditions: make([]*StrategyCondition, 0),
	}

	// 解析每个部分
	for i, part := range parts {
		part = strings.TrimSpace(part)

		// 检查是否是占位符
		if strings.HasPrefix(part, "PLACEHOLDER_") {
			if placeholder, exists := p.placeholders[part]; exists {
				condition.SubConditions = append(condition.SubConditions, placeholder)
			} else {
				return nil, fmt.Errorf("placeholder not found: %s", part)
			}
		} else {
			// 每个部分可能是单个表达式或AND表达式
			subCondition, err := p.parseAndExpression(part)
			if err != nil {
				return nil, fmt.Errorf("failed to parse or part '%s': %w", part, err)
			}

			// 如果是单个表达式，直接添加
			if len(subCondition.Expressions) == 1 && len(subCondition.SubConditions) == 0 {
				condition.Expressions = append(condition.Expressions, subCondition.Expressions[0])
			} else {
				// 如果是多个表达式或子条件，需要创建子条件
				condition.SubConditions = append(condition.SubConditions, subCondition)
			}
		}

		// 添加OR操作符（除了最后一个）
		if i < len(parts)-1 {
			condition.Operators = append(condition.Operators, "or")
		}
	}

	return condition, nil
}

// parseAndExpression 解析AND表达式
func (p *StrategyParser) parseAndExpression(expression string) (*StrategyCondition, error) {
	parts := strings.Split(expression, " and ")
	if len(parts) < 2 {
		// 单个表达式
		part := strings.TrimSpace(expression)

		// 检查是否是占位符
		if strings.HasPrefix(part, "PLACEHOLDER_") {
			if placeholder, exists := p.placeholders[part]; exists {
				return placeholder, nil
			} else {
				return nil, fmt.Errorf("placeholder not found: %s", part)
			}
		}

		expr, err := p.parseSingleExpression(part)
		if err != nil {
			return nil, fmt.Errorf("failed to parse single expression '%s': %w", expression, err)
		}
		return &StrategyCondition{
			Expressions: []StrategyExpression{*expr},
			Operators:   []string{},
		}, nil
	}

	condition := &StrategyCondition{
		Expressions:   make([]StrategyExpression, 0),
		Operators:     make([]string, 0, len(parts)-1),
		SubConditions: make([]*StrategyCondition, 0),
	}

	// 解析每个部分
	for i, part := range parts {
		part = strings.TrimSpace(part)

		// 检查是否是占位符
		if strings.HasPrefix(part, "PLACEHOLDER_") {
			if placeholder, exists := p.placeholders[part]; exists {
				condition.SubConditions = append(condition.SubConditions, placeholder)
			} else {
				return nil, fmt.Errorf("placeholder not found: %s", part)
			}
		} else {
			expr, err := p.parseSingleExpression(part)
			if err != nil {
				return nil, fmt.Errorf("failed to parse and part '%s': %w", part, err)
			}
			condition.Expressions = append(condition.Expressions, *expr)
		}

		// 添加AND操作符（除了最后一个）
		if i < len(parts)-1 {
			condition.Operators = append(condition.Operators, "and")
		}
	}

	return condition, nil
}

// splitByOperators 按操作符分割表达式
func (p *StrategyParser) splitByOperators(expression string) []string {
	// 先按 or 分割
	orParts := strings.Split(expression, " or ")
	if len(orParts) > 1 {
		// 有 or 操作符，返回所有部分
		var result []string
		for _, part := range orParts {
			// 按 and 分割每个 or 部分
			andParts := strings.Split(part, " and ")
			for _, andPart := range andParts {
				result = append(result, strings.TrimSpace(andPart))
			}
		}
		return result
	}

	// 没有 or，按 and 分割
	return strings.Split(expression, " and ")
}

// parseSingleExpression 解析单个表达式
func (p *StrategyParser) parseSingleExpression(expression string) (*StrategyExpression, error) {
	// 检查是否是占位符
	if strings.HasPrefix(expression, "PLACEHOLDER_") {
		return nil, fmt.Errorf("placeholder found in single expression: %s", expression)
	}

	// 匹配字段访问模式：candles.SOL.last_px >= 200
	fieldAccessMatch := p.fieldAccessRegex.FindStringSubmatch(expression)
	if len(fieldAccessMatch) < 4 {
		return nil, fmt.Errorf("invalid field access pattern: %s", expression)
	}

	source := fieldAccessMatch[1]
	symbol := fieldAccessMatch[2]
	field := fieldAccessMatch[3]
	subField := ""
	if len(fieldAccessMatch) > 4 && fieldAccessMatch[4] != "" {
		subField = fieldAccessMatch[4]
	}

	// 验证数据源
	if source != string(DataSourceCandles) && source != string(DataSourceNews) {
		return nil, fmt.Errorf("unsupported data source: %s", source)
	}

	// 提取操作符和值
	operator, value, values, err := p.extractOperatorAndValue(expression)
	if err != nil {
		return nil, err
	}

	expr := &StrategyExpression{
		FieldAccess: FieldAccess{
			Source:   DataSource(source),
			Symbol:   symbol,
			Field:    field,
			SubField: subField,
		},
		Operator:   operator,
		Value:      value,
		Values:     values,
		Parameters: make(map[string]interface{}),
	}

	return expr, nil
}

// extractOperatorAndValue 提取操作符和值
func (p *StrategyParser) extractOperatorAndValue(expression string) (string, interface{}, []interface{}, error) {
	// 支持的比较操作符
	comparisonOps := []string{">=", "<=", "==", "!=", ">", "<"}

	// 支持的包含操作符
	containmentOps := []string{" in ", " not_in "}

	// 检查包含操作符
	for _, op := range containmentOps {
		if strings.Contains(expression, op) {
			parts := strings.Split(expression, op)
			if len(parts) != 2 {
				continue
			}

			// 解析值数组
			values, err := p.parseValueArray(strings.TrimSpace(parts[1]))
			if err != nil {
				continue
			}

			operator := strings.TrimSpace(op)
			return operator, nil, values, nil
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

			return op, value, nil, nil
		}
	}

	// 如果没有找到操作符，检查是否是简单的字段访问（如 candles.SOL.last_volume）
	// 这种情况下，我们假设是检查字段是否存在（非零值）
	if p.fieldAccessRegex.MatchString(expression) {
		return "!=", 0.0, nil, nil
	}

	return "", nil, nil, fmt.Errorf("no valid operator found in expression: %s", expression)
}

// parseValue 解析单个值
func (p *StrategyParser) parseValue(valueStr string) (interface{}, error) {
	// 尝试解析为数字
	if num, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return num, nil
	}

	// 尝试解析为整数
	if num, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return num, nil
	}

	// 尝试解析为布尔值
	if valueStr == "true" {
		return true, nil
	}
	if valueStr == "false" {
		return false, nil
	}

	// 作为字符串处理（去除引号）
	if len(valueStr) >= 2 && valueStr[0] == '"' && valueStr[len(valueStr)-1] == '"' {
		return valueStr[1 : len(valueStr)-1], nil
	}

	return valueStr, nil
}

// parseValueArray 解析值数组
func (p *StrategyParser) parseValueArray(valueStr string) ([]interface{}, error) {
	// 匹配字符串数组：["突破新高", "重大利好"]
	matches := p.stringArrayRegex.FindStringSubmatch(valueStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid array format: %s", valueStr)
	}

	// 分割数组内容
	content := matches[1]
	if content == "" {
		return []interface{}{}, nil
	}

	// 简单的字符串分割（假设用逗号分隔）
	parts := strings.Split(content, ",")
	var values []interface{}

	for _, part := range parts {
		value, err := p.parseValue(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}

// isValidOperator 验证操作符是否有效
func (p *StrategyParser) isValidOperator(op string) bool {
	validOps := []string{">", "<", ">=", "<=", "==", "!=", "in", "not_in"}
	for _, validOp := range validOps {
		if op == validOp {
			return true
		}
	}
	return false
}

// Evaluate 评估策略条件
func (p *StrategyParser) Evaluate(condition *StrategyCondition, results map[string]bool) (bool, error) {
	if len(condition.Expressions) == 0 && len(condition.SubConditions) == 0 {
		return false, fmt.Errorf("no expressions to evaluate")
	}

	// 处理子条件
	if len(condition.SubConditions) > 0 {
		return p.evaluateSubConditions(condition, results)
	}

	// 如果只有一个表达式
	if len(condition.Expressions) == 1 {
		expr := condition.Expressions[0]
		key := p.generateExpressionKey(expr)
		result, exists := results[key]
		if !exists {
			return false, fmt.Errorf("strategy result not found: %s", key)
		}
		return result, nil
	}

	// 多个表达式需要按操作符组合
	if len(condition.Expressions) == 0 {
		return false, fmt.Errorf("no expressions to evaluate")
	}

	key := p.generateExpressionKey(condition.Expressions[0])
	result, exists := results[key]
	if !exists {
		return false, fmt.Errorf("strategy result not found: %s", key)
	}

	for i, operator := range condition.Operators {
		if i+1 >= len(condition.Expressions) {
			break
		}

		nextKey := p.generateExpressionKey(condition.Expressions[i+1])
		nextResult, exists := results[nextKey]
		if !exists {
			return false, fmt.Errorf("strategy result not found: %s", nextKey)
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

// evaluateSubConditions 评估子条件
func (p *StrategyParser) evaluateSubConditions(condition *StrategyCondition, results map[string]bool) (bool, error) {
	if len(condition.SubConditions) == 0 {
		return false, fmt.Errorf("no sub conditions to evaluate")
	}

	// 评估第一个子条件
	result, err := p.Evaluate(condition.SubConditions[0], results)
	if err != nil {
		return false, err
	}

	// 按操作符组合其他子条件
	for i, operator := range condition.Operators {
		if i+1 >= len(condition.SubConditions) {
			break
		}

		nextResult, err := p.Evaluate(condition.SubConditions[i+1], results)
		if err != nil {
			return false, err
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

// generateExpressionKey 生成表达式键
func (p *StrategyParser) generateExpressionKey(expr StrategyExpression) string {
	key := fmt.Sprintf("%s.%s.%s", expr.FieldAccess.Source, expr.FieldAccess.Symbol, expr.FieldAccess.Field)
	if expr.FieldAccess.SubField != "" {
		key += "." + expr.FieldAccess.SubField
	}
	return key
}

// GetStrategyNames 获取表达式中包含的策略名称
func (p *StrategyParser) GetStrategyNames(condition *StrategyCondition) []string {
	names := make([]string, 0)
	seen := make(map[string]bool)

	// 收集表达式中的策略名称
	for _, expr := range condition.Expressions {
		key := p.generateExpressionKey(expr)
		if !seen[key] {
			names = append(names, key)
			seen[key] = true
		}
	}

	// 递归收集子条件中的策略名称
	for _, subCondition := range condition.SubConditions {
		subNames := p.GetStrategyNames(subCondition)
		for _, name := range subNames {
			if !seen[name] {
				names = append(names, name)
				seen[name] = true
			}
		}
	}

	return names
}

// GetParameters 获取策略参数
func (p *StrategyParser) GetParameters(condition *StrategyCondition) map[string]map[string]interface{} {
	params := make(map[string]map[string]interface{})

	// 收集表达式中的参数
	for _, expr := range condition.Expressions {
		key := p.generateExpressionKey(expr)
		if params[key] == nil {
			params[key] = make(map[string]interface{})
		}

		// 添加字段访问信息
		params[key]["source"] = expr.FieldAccess.Source
		params[key]["symbol"] = expr.FieldAccess.Symbol
		params[key]["field"] = expr.FieldAccess.Field
		if expr.FieldAccess.SubField != "" {
			params[key]["sub_field"] = expr.FieldAccess.SubField
		}
		params[key]["operator"] = expr.Operator
		params[key]["value"] = expr.Value
		if len(expr.Values) > 0 {
			params[key]["values"] = expr.Values
		}

		// 合并其他参数
		for paramKey, value := range expr.Parameters {
			params[key][paramKey] = value
		}
	}

	// 递归收集子条件中的参数
	for _, subCondition := range condition.SubConditions {
		subParams := p.GetParameters(subCondition)
		for key, value := range subParams {
			params[key] = value
		}
	}

	return params
}
