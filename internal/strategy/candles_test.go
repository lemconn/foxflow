package strategy

import (
	"context"
	"testing"
	"time"
)

func TestCandlesStrategy_Evaluate(t *testing.T) {
	strategy := NewCandlesStrategy()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		expected    bool
		description string
	}{
		{
			name: "sol_price_greater_than_200",
			params: map[string]interface{}{
				"symbol":   "SOL",
				"field":    "last_px",
				"operator": ">=",
				"value":    200.0,
			},
			expectError: false,
			expected:    true, // SOL Mock数据中last_px是205.8
			description: "SOL价格大于等于200",
		},
		{
			name: "sol_price_less_than_210",
			params: map[string]interface{}{
				"symbol":   "SOL",
				"field":    "last_px",
				"operator": "<",
				"value":    210.0,
			},
			expectError: false,
			expected:    true, // SOL Mock数据中last_px是205.8
			description: "SOL价格小于210",
		},
		{
			name: "sol_volume_greater_than_1000000",
			params: map[string]interface{}{
				"symbol":   "SOL",
				"field":    "last_volume",
				"operator": ">",
				"value":    1000000.0,
			},
			expectError: false,
			expected:    true, // SOL Mock数据中last_volume是1500000.0
			description: "SOL成交量大于1000000",
		},
		{
			name: "btc_price_less_than_50000",
			params: map[string]interface{}{
				"symbol":   "BTC",
				"field":    "last_px",
				"operator": "<",
				"value":    50000.0,
			},
			expectError: false,
			expected:    true, // BTC Mock数据中last_px是45500.0
			description: "BTC价格小于50000",
		},
		{
			name: "invalid_symbol",
			params: map[string]interface{}{
				"symbol":   "INVALID",
				"field":    "last_px",
				"operator": ">=",
				"value":    100.0,
			},
			expectError: true,
			expected:    false,
			description: "无效的交易对符号",
		},
		{
			name: "invalid_field",
			params: map[string]interface{}{
				"symbol":   "SOL",
				"field":    "invalid_field",
				"operator": ">=",
				"value":    100.0,
			},
			expectError: true,
			expected:    false,
			description: "无效的字段名",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证参数
			err := strategy.ValidateParameters(tt.params)
			if tt.expectError {
				if err == nil {
					t.Errorf("期望参数验证错误但没有收到错误: %s", tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("参数验证失败: %v, 测试: %s", err, tt.description)
				return
			}

			// 执行策略
			result, err := strategy.Evaluate(context.Background(), nil, "SOL", tt.params)
			if err != nil {
				t.Errorf("策略执行失败: %v, 测试: %s", err, tt.description)
				return
			}

			if result != tt.expected {
				t.Errorf("策略结果不匹配，期望: %v, 实际: %v, 测试: %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestCandlesStrategy_UpdateMockData(t *testing.T) {
	strategy := NewCandlesStrategy()

	// 更新Mock数据
	newData := &CandleData{
		Symbol:     "TEST",
		Open:       100.0,
		High:       110.0,
		Low:        90.0,
		Close:      105.0,
		Volume:     1000.0,
		Timestamp:  time.Now(),
		LastPx:     105.0,
		LastVolume: 1000.0,
	}

	strategy.UpdateMockData("TEST", newData)

	// 验证更新
	retrievedData, exists := strategy.GetMockData("TEST")
	if !exists {
		t.Fatal("无法获取更新的Mock数据")
	}

	if retrievedData.Symbol != "TEST" {
		t.Errorf("符号不匹配，期望: TEST, 实际: %s", retrievedData.Symbol)
	}

	if retrievedData.LastPx != 105.0 {
		t.Errorf("价格不匹配，期望: 105.0, 实际: %f", retrievedData.LastPx)
	}
}

func TestCandlesStrategy_GetParameters(t *testing.T) {
	strategy := NewCandlesStrategy()

	params := strategy.GetParameters()
	if params == nil {
		t.Fatal("参数为空")
	}

	// 验证必需参数
	requiredParams := []string{"symbol", "field", "operator", "value"}
	for _, param := range requiredParams {
		if _, exists := params[param]; !exists {
			t.Errorf("缺少必需参数: %s", param)
		}
	}
}

func TestCandlesStrategy_GetName(t *testing.T) {
	strategy := NewCandlesStrategy()

	name := strategy.GetName()
	if name != "candles" {
		t.Errorf("策略名称不匹配，期望: candles, 实际: %s", name)
	}
}

func TestCandlesStrategy_GetDescription(t *testing.T) {
	strategy := NewCandlesStrategy()

	description := strategy.GetDescription()
	if description == "" {
		t.Error("策略描述为空")
	}
}
