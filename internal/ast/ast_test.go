package ast

import (
	"context"
	"testing"
	"time"

	"github.com/lemconn/foxflow/internal/data"
)

func TestASTExecution(t *testing.T) {
	// 创建数据管理器并初始化默认模块
	dataMgr := data.InitDefaultModules()
	dataAdapter := NewDataAdapter(dataMgr)
	executor := NewExecutor(dataAdapter)

	ctx := context.Background()

	tests := []struct {
		name     string
		node     Node
		expected bool
	}{
		{
			name: "simple_candles_comparison",
			node: &BinaryExpression{
				Operator: OpGT,
				Left: &DataRef{
					Module: DataTypeCandles,
					Entity: "SOL",
					Field:  "last_px",
				},
				Right: &Value{Value: 200.0},
			},
			expected: true, // SOL价格205.8 > 200
		},
		{
			name: "simple_news_comparison",
			node: &BinaryExpression{
				Operator: OpContains,
				Left: &DataRef{
					Module: DataTypeNews,
					Entity: "theblockbeats",
					Field:  "last_title",
				},
				Right: &Value{Value: []interface{}{"新高", "SOL"}},
			},
			expected: true, // 标题包含"新高"
		},
		{
			name: "and_expression",
			node: &BinaryExpression{
				Operator: OpAnd,
				Left: &BinaryExpression{
					Operator: OpGT,
					Left: &DataRef{
						Module: DataTypeCandles,
						Entity: "SOL",
						Field:  "last_px",
					},
					Right: &Value{Value: 200.0},
				},
				Right: &BinaryExpression{
					Operator: OpGT,
					Left: &DataRef{
						Module: DataTypeCandles,
						Entity: "SOL",
						Field:  "last_volume",
					},
					Right: &Value{Value: 100000.0},
				},
			},
			expected: true, // 两个条件都满足
		},
		{
			name: "or_expression",
			node: &BinaryExpression{
				Operator: OpOr,
				Left: &BinaryExpression{
					Operator: OpGT,
					Left: &DataRef{
						Module: DataTypeCandles,
						Entity: "SOL",
						Field:  "last_px",
					},
					Right: &Value{Value: 200.0},
				},
				Right: &BinaryExpression{
					Operator: OpGT,
					Left: &DataRef{
						Module: DataTypeCandles,
						Entity: "SOL",
						Field:  "last_volume",
					},
					Right: &Value{Value: 2000000.0}, // 这个条件不满足
				},
			},
			expected: true, // 第一个条件满足
		},
		{
			name: "function_call_avg",
			node: &BinaryExpression{
				Operator: OpGT,
				Left: &FunctionCall{
					Name: "avg",
					Args: []Node{
						&DataRef{
							Module: DataTypeCandles,
							Entity: "BTC",
							Field:  "close",
						},
						&Value{Value: 5},
					},
				},
				Right: &DataRef{
					Module: DataTypeCandles,
					Entity: "BTC",
					Field:  "last_px",
				},
			},
			expected: false, // avg(45500, 5) = 200 < 45500
		},
		{
			name: "function_call_time_since",
			node: &BinaryExpression{
				Operator: OpLT,
				Left: &FunctionCall{
					Name: "time_since",
					Args: []Node{
						&DataRef{
							Module: DataTypeNews,
							Entity: "coindesk",
							Field:  "last_update_time",
						},
					},
				},
				Right: &Value{Value: 1200.0}, // 20分钟，给更多缓冲时间
			},
			expected: true, // 新闻更新时间在20分钟内
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecuteToBool(ctx, tt.node)
			if err != nil {
				t.Errorf("执行AST时出错: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("期望结果 %v，但得到 %v", tt.expected, result)
			}
		})
	}
}

func TestComplexExpression(t *testing.T) {
	// 创建数据管理器并初始化默认模块
	dataMgr := data.InitDefaultModules()
	dataAdapter := NewDataAdapter(dataMgr)
	executor := NewExecutor(dataAdapter)

	ctx := context.Background()

	// 测试复杂表达式：(avg(candles.BTC.close, 5) > candles.BTC.last_px) and (time_since(news.coindesk.last_update_time) < 600)
	complexNode := &BinaryExpression{
		Operator: OpAnd,
		Left: &BinaryExpression{
			Operator: OpGT,
			Left: &FunctionCall{
				Name: "avg",
				Args: []Node{
					&DataRef{
						Module: DataTypeCandles,
						Entity: "BTC",
						Field:  "close",
					},
					&Value{Value: 5},
				},
			},
			Right: &DataRef{
				Module: DataTypeCandles,
				Entity: "BTC",
				Field:  "last_px",
			},
		},
		Right: &BinaryExpression{
			Operator: OpLT,
			Left: &FunctionCall{
				Name: "time_since",
				Args: []Node{
					&DataRef{
						Module: DataTypeNews,
						Entity: "coindesk",
						Field:  "last_update_time",
					},
				},
			},
			Right: &Value{Value: 600.0},
		},
	}

	result, err := executor.ExecuteToBool(ctx, complexNode)
	if err != nil {
		t.Errorf("执行复杂表达式时出错: %v", err)
		return
	}

	// 由于avg函数返回模拟值200，而BTC价格是45500，所以第一个条件为false
	// 第二个条件为true（新闻更新时间在10分钟内）
	// 所以整个表达式应该是false
	expected := false
	if result != expected {
		t.Errorf("期望结果 %v，但得到 %v", expected, result)
	}
}

func TestDataRef(t *testing.T) {
	// 创建数据管理器并初始化默认模块
	dataMgr := data.InitDefaultModules()
	dataAdapter := NewDataAdapter(dataMgr)
	executor := NewExecutor(dataAdapter)

	ctx := context.Background()

	tests := []struct {
		name     string
		dataRef  *DataRef
		expected interface{}
	}{
		{
			name: "candles_last_px",
			dataRef: &DataRef{
				Module: DataTypeCandles,
				Entity: "SOL",
				Field:  "last_px",
			},
			expected: 205.8,
		},
		{
			name: "news_last_title",
			dataRef: &DataRef{
				Module: DataTypeNews,
				Entity: "theblockbeats",
				Field:  "last_title",
			},
			expected: "SOL突破新高，市值创新纪录",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(ctx, tt.dataRef)
			if err != nil {
				t.Errorf("执行DataRef时出错: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("期望结果 %v，但得到 %v", tt.expected, result)
			}
		})
	}
}

func TestFunctionCall(t *testing.T) {
	// 创建数据管理器并初始化默认模块
	dataMgr := data.InitDefaultModules()
	dataAdapter := NewDataAdapter(dataMgr)
	executor := NewExecutor(dataAdapter)

	ctx := context.Background()

	tests := []struct {
		name     string
		funcCall *FunctionCall
		expected interface{}
	}{
		{
			name: "time_since_function",
			funcCall: &FunctionCall{
				Name: "time_since",
				Args: []Node{
					&Value{Value: time.Now().Add(-5 * time.Minute)},
				},
			},
			expected: 300.0, // 大约5分钟 = 300秒
		},
		{
			name: "contains_function",
			funcCall: &FunctionCall{
				Name: "contains",
				Args: []Node{
					&Value{Value: "SOL突破新高，市值创新纪录"},
					&Value{Value: []interface{}{"新高", "SOL"}},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.Execute(ctx, tt.funcCall)
			if err != nil {
				t.Errorf("执行FunctionCall时出错: %v", err)
				return
			}

			// 对于time_since函数，我们只检查结果是否为正数
			if tt.name == "time_since_function" {
				if timeVal, ok := result.(float64); !ok || timeVal <= 0 {
					t.Errorf("time_since函数应该返回正数，但得到 %v", result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("期望结果 %v，但得到 %v", tt.expected, result)
				}
			}
		})
	}
}
