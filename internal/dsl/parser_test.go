package dsl

import (
	"testing"

	"github.com/lemconn/foxflow/internal/ast"
)

func TestDSLParser(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		expression  string
		expectError bool
		description string
	}{
		{
			name:        "simple_candles_expression",
			expression:  "candles.SOL.last_px > 200",
			expectError: false,
			description: "简单的K线价格比较",
		},
		{
			name:        "simple_news_expression",
			expression:  "news.theblockbeats.last_title contains [\"新高\"]",
			expectError: false,
			description: "简单的新闻标题包含检查",
		},
		{
			name:        "complex_and_expression",
			expression:  "candles.SOL.last_px > 200 and candles.SOL.last_volume > 1000000",
			expectError: false,
			description: "复杂的AND条件",
		},
		{
			name:        "complex_or_expression",
			expression:  "candles.SOL.last_px > 200 or news.theblockbeats.last_title contains [\"新高\"]",
			expectError: false,
			description: "复杂的OR条件",
		},
		{
			name:        "complex_mixed_expression",
			expression:  "(candles.SOL.last_px > 200 and candles.SOL.last_volume > 1000000) or (news.theblockbeats.last_title contains [\"新高\"] and news.theblockbeats.last_update_time < 10)",
			expectError: false,
			description: "复杂的混合条件（带括号）",
		},
		{
			name:        "function_call_avg",
			expression:  "avg(candles.BTC.close, 5) > candles.BTC.last_px",
			expectError: false,
			description: "函数调用avg",
		},
		{
			name:        "function_call_time_since",
			expression:  "time_since(news.coindesk.last_update_time) < 600",
			expectError: false,
			description: "函数调用time_since",
		},
		{
			name:        "function_call_contains",
			expression:  "contains(news.theblockbeats.last_title, [\"新高\", \"SOL\"])",
			expectError: false,
			description: "函数调用contains",
		},
		{
			name:        "complex_function_expression",
			expression:  "(avg(candles.BTC.close, 5) > candles.BTC.last_px) and (time_since(news.coindesk.last_update_time) < 600)",
			expectError: false,
			description: "复杂的函数调用表达式",
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
			node, err := parser.Parse(tt.expression)

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

			if node == nil {
				t.Errorf("解析结果为空: %s", tt.expression)
				return
			}

			// 验证解析结果
			t.Logf("表达式: %s", tt.expression)
			t.Logf("AST节点: %s", node.String())
			t.Logf("节点类型: %s", node.Type())
		})
	}
}

func TestComplexExpression(t *testing.T) {
	parser := NewParser()

	// 测试用户提供的复杂表达式
	expression := "(avg(candles.BTC.close, 5) > candles.BTC.last_px) and (time_since(news.coindesk.last_update_time) < 600)"

	node, err := parser.Parse(expression)
	if err != nil {
		t.Fatalf("解析复杂表达式失败: %v", err)
	}

	// 验证解析结果
	t.Logf("复杂表达式: %s", expression)
	t.Logf("AST节点: %s", node.String())
	t.Logf("节点类型: %s", node.Type())

	// 验证节点结构
	binaryExpr, ok := node.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("根节点应该是BinaryExpression，但得到 %T", node)
	}

	if binaryExpr.Operator != ast.OpAnd {
		t.Errorf("根操作符应该是and，但得到 %s", binaryExpr.Operator)
	}

	// 验证左操作数（函数调用）
	leftBinaryExpr, ok := binaryExpr.Left.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("左操作数应该是BinaryExpression，但得到 %T", binaryExpr.Left)
	}

	if leftBinaryExpr.Operator != ast.OpGT {
		t.Errorf("左操作符应该是>，但得到 %s", leftBinaryExpr.Operator)
	}

	// 验证函数调用
	funcCall, ok := leftBinaryExpr.Left.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("左操作数的左操作数应该是FunctionCall，但得到 %T", leftBinaryExpr.Left)
	}

	if funcCall.Name != "avg" {
		t.Errorf("函数名应该是avg，但得到 %s", funcCall.Name)
	}

	if len(funcCall.Args) != 2 {
		t.Errorf("avg函数应该有2个参数，但得到 %d", len(funcCall.Args))
	}

	// 验证右操作数（函数调用）
	rightBinaryExpr, ok := binaryExpr.Right.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("右操作数应该是BinaryExpression，但得到 %T", binaryExpr.Right)
	}

	if rightBinaryExpr.Operator != ast.OpLT {
		t.Errorf("右操作符应该是<，但得到 %s", rightBinaryExpr.Operator)
	}

	// 验证time_since函数调用
	timeSinceCall, ok := rightBinaryExpr.Left.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("右操作数的左操作数应该是FunctionCall，但得到 %T", rightBinaryExpr.Left)
	}

	if timeSinceCall.Name != "time_since" {
		t.Errorf("函数名应该是time_since，但得到 %s", timeSinceCall.Name)
	}

	if len(timeSinceCall.Args) != 1 {
		t.Errorf("time_since函数应该有1个参数，但得到 %d", len(timeSinceCall.Args))
	}
}

func TestUserProvidedTestCases(t *testing.T) {
	parser := NewParser()

	// 用户提供的测试用例
	testCases := []struct {
		name       string
		expression string
	}{
		{
			name:       "simple_price_comparison",
			expression: "candles.SOL.last_px > 200",
		},
		{
			name:       "or_expression",
			expression: "candles.SOL.last_px > 200 or candles.SOL.last_volume > 100000",
		},
		{
			name:       "contains_function",
			expression: "contains(news.theblockbeats.last_title, [\"新高\", \"SOL\"]) and time_since(news.theblockbeats.last_update_time) < 300",
		},
		{
			name:       "complex_nested_expression",
			expression: "(candles.SOL.last_px > 200 and candles.SOL.last_volume > 100000) or (contains(news.theblockbeats.last_title, [\"新高\", \"SOL\"]) and time_since(news.theblockbeats.last_update_time) < 300)",
		},
		{
			name:       "avg_function_expression",
			expression: "(avg(candles.BTC.close, 5) > candles.BTC.last_px) and (time_since(news.coindesk.last_update_time) < 600)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := parser.Parse(tc.expression)
			if err != nil {
				t.Errorf("解析表达式失败: %v, 表达式: %s", err, tc.expression)
				return
			}

			if node == nil {
				t.Errorf("解析结果为空: %s", tc.expression)
				return
			}

			t.Logf("测试用例: %s", tc.name)
			t.Logf("表达式: %s", tc.expression)
			t.Logf("AST节点: %s", node.String())
			t.Logf("节点类型: %s", node.Type())
		})
	}
}

func TestValidate(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		expression  string
		expectError bool
	}{
		{
			name:        "valid_expression",
			expression:  "candles.SOL.last_px > 200",
			expectError: false,
		},
		{
			name:        "invalid_expression",
			expression:  "invalid.source.field > 100",
			expectError: true,
		},
		{
			name:        "empty_expression",
			expression:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.Validate(tt.expression)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望错误但没有收到错误: %s", tt.expression)
				}
			} else {
				if err != nil {
					t.Errorf("期望无错误但收到错误: %v, 表达式: %s", err, tt.expression)
				}
			}
		})
	}
}
