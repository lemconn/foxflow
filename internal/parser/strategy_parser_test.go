package parser

import (
	"testing"
)

func TestStrategyParser_Parse(t *testing.T) {
	parser := NewStrategyParser()

	tests := []struct {
		name        string
		expression  string
		expectError bool
		description string
	}{
		{
			name:        "simple_candles_expression",
			expression:  "candles.SOL.last_px >= 200",
			expectError: false,
			description: "简单的K线价格比较",
		},
		{
			name:        "simple_news_expression",
			expression:  "news.theblockbeats.last_title in [\"突破新高\"]",
			expectError: false,
			description: "简单的新闻标题包含检查",
		},
		{
			name:        "complex_and_expression",
			expression:  "candles.SOL.last_px >= 200 and candles.SOL.last_volume > 1000000",
			expectError: false,
			description: "复杂的AND条件",
		},
		{
			name:        "complex_or_expression",
			expression:  "candles.SOL.last_px >= 200 or news.theblockbeats.last_title in [\"突破新高\"]",
			expectError: false,
			description: "复杂的OR条件",
		},
		{
			name:        "complex_mixed_expression",
			expression:  "(candles.SOL.last_px >= 200 and candles.SOL.last_volume > 1000000) or (news.theblockbeats.last_title in [\"突破新高\"] and news.theblockbeats.last_update_time < 10)",
			expectError: false,
			description: "复杂的混合条件（带括号）",
		},
		{
			name:        "news_time_expression",
			expression:  "news.theblockbeats.last_update_time < 10",
			expectError: false,
			description: "新闻时间比较",
		},
		{
			name:        "invalid_expression",
			expression:  "invalid.source.field > 100",
			expectError: true,
			description: "无效的数据源",
		},
		{
			name:        "empty_expression",
			expression:  "",
			expectError: true,
			description: "空表达式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition, err := parser.Parse(tt.expression)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望错误但没有收到错误: %s", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("解析表达式时出错: %v, 表达式: %s", err, tt.expression)
				return
			}

			if condition == nil {
				t.Errorf("解析结果为空: %s", tt.expression)
				return
			}

			// 验证解析结果
			strategyNames := parser.GetStrategyNames(condition)
			parameters := parser.GetParameters(condition)

			t.Logf("表达式: %s", tt.expression)
			t.Logf("策略名称: %v", strategyNames)
			t.Logf("参数: %+v", parameters)
			t.Logf("表达式数量: %d", len(condition.Expressions))
			t.Logf("操作符数量: %d", len(condition.Operators))
			t.Logf("子条件数量: %d", len(condition.SubConditions))
		})
	}
}

func TestStrategyParser_ComplexExpression(t *testing.T) {
	parser := NewStrategyParser()

	// 测试复杂的表达式：用户提供的示例
	expression := "(candles.SOL.last_px >= 200 and candles.SOL.last_volume) or (news.theblockbeats.last_title in [\"突破新高\"] and news.theblockbeats.last_update_time < 10)"

	condition, err := parser.Parse(expression)
	if err != nil {
		t.Fatalf("解析复杂表达式失败: %v", err)
	}

	// 验证解析结果
	strategyNames := parser.GetStrategyNames(condition)
	parameters := parser.GetParameters(condition)

	t.Logf("复杂表达式: %s", expression)
	t.Logf("策略名称: %v", strategyNames)
	t.Logf("参数: %+v", parameters)

	// 验证策略名称
	expectedStrategies := []string{
		"candles.SOL.last_px",
		"candles.SOL.last_volume",
		"news.theblockbeats.last_title",
		"news.theblockbeats.last_update_time",
	}

	if len(strategyNames) != len(expectedStrategies) {
		t.Errorf("策略名称数量不匹配，期望: %d, 实际: %d", len(expectedStrategies), len(strategyNames))
	}

	// 验证参数结构
	for _, strategyName := range strategyNames {
		params, exists := parameters[strategyName]
		if !exists {
			t.Errorf("缺少策略参数: %s", strategyName)
			continue
		}

		// 验证必需参数
		requiredParams := []string{"source", "symbol", "field", "operator"}
		for _, param := range requiredParams {
			if _, exists := params[param]; !exists {
				t.Errorf("策略 %s 缺少必需参数: %s", strategyName, param)
			}
		}
	}
}

func TestStrategyParser_Evaluate(t *testing.T) {
	parser := NewStrategyParser()

	// 测试简单表达式
	expression := "candles.SOL.last_px >= 200"
	condition, err := parser.Parse(expression)
	if err != nil {
		t.Fatalf("解析表达式失败: %v", err)
	}

	// 模拟策略结果
	results := map[string]bool{
		"candles.SOL.last_px": true, // SOL价格 >= 200
	}

	result, err := parser.Evaluate(condition, results)
	if err != nil {
		t.Fatalf("评估表达式失败: %v", err)
	}

	if !result {
		t.Errorf("期望结果为true，但得到false")
	}

	// 测试AND表达式
	andExpression := "candles.SOL.last_px >= 200 and candles.SOL.last_volume > 1000000"
	andCondition, err := parser.Parse(andExpression)
	if err != nil {
		t.Fatalf("解析AND表达式失败: %v", err)
	}

	andResults := map[string]bool{
		"candles.SOL.last_px":     true,  // SOL价格 >= 200
		"candles.SOL.last_volume": false, // 成交量 <= 1000000
	}

	andResult, err := parser.Evaluate(andCondition, andResults)
	if err != nil {
		t.Fatalf("评估AND表达式失败: %v", err)
	}

	if andResult {
		t.Errorf("期望AND结果为false，但得到true")
	}

	// 测试OR表达式
	orExpression := "candles.SOL.last_px >= 200 or candles.SOL.last_volume > 1000000"
	orCondition, err := parser.Parse(orExpression)
	if err != nil {
		t.Fatalf("解析OR表达式失败: %v", err)
	}

	orResults := map[string]bool{
		"candles.SOL.last_px":     true,  // SOL价格 >= 200
		"candles.SOL.last_volume": false, // 成交量 <= 1000000
	}

	orResult, err := parser.Evaluate(orCondition, orResults)
	if err != nil {
		t.Fatalf("评估OR表达式失败: %v", err)
	}

	if !orResult {
		t.Errorf("期望OR结果为true，但得到false")
	}
}

func TestStrategyParser_FieldAccess(t *testing.T) {
	parser := NewStrategyParser()

	tests := []struct {
		expression string
		expected   FieldAccess
	}{
		{
			expression: "candles.SOL.last_px >= 200",
			expected: FieldAccess{
				Source:   DataSourceCandles,
				Symbol:   "SOL",
				Field:    "last_px",
				SubField: "",
			},
		},
		{
			expression: "news.theblockbeats.last_title in [\"突破新高\"]",
			expected: FieldAccess{
				Source:   DataSourceNews,
				Symbol:   "theblockbeats",
				Field:    "last_title",
				SubField: "",
			},
		},
		{
			expression: "news.theblockbeats.last_update_time < 10",
			expected: FieldAccess{
				Source:   DataSourceNews,
				Symbol:   "theblockbeats",
				Field:    "last_update_time",
				SubField: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			condition, err := parser.Parse(tt.expression)
			if err != nil {
				t.Fatalf("解析表达式失败: %v", err)
			}

			if len(condition.Expressions) == 0 {
				t.Fatalf("没有找到表达式")
			}

			actual := condition.Expressions[0].FieldAccess
			if actual != tt.expected {
				t.Errorf("字段访问不匹配，期望: %+v, 实际: %+v", tt.expected, actual)
			}
		})
	}
}
