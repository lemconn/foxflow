package provider

import (
	"context"
	"strings"
	"testing"
)

func TestNewMarketProvider(t *testing.T) {
	provider := NewMarketProvider()

	if provider == nil {
		t.Fatal("NewMarketProvider 返回 nil")
	}

	if provider.GetName() != "market" {
		t.Errorf("期望提供者名称为 'market'，实际得到 '%s'", provider.GetName())
	}

	if provider.exchangeMgr == nil {
		t.Error("exchangeMgr 不应为 nil")
	}
}

func TestMarketProviderGetDataValidFields(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 只测试可用的交易所（OKX）
	testCases := []struct {
		dataSource string
		field      string
	}{
		{"okx", "BTC.price"},
		{"okx", "BTC.volume"},
		{"okx", "BTC.high"},
		{"okx", "BTC.low"},
	}

	for _, tc := range testCases {
		t.Run(tc.dataSource+"_"+tc.field, func(t *testing.T) {
			data, err := provider.GetData(ctx, tc.dataSource, tc.field)
			if err != nil {
				t.Errorf("获取数据失败: %v", err)
				return
			}

			if data == nil {
				t.Error("数据不应为空")
				return
			}

			// 验证数据类型为 float64
			if value, ok := data.(float64); ok {
				// 验证数值合理性（价格应该大于0）
				if value <= 0 {
					t.Errorf("数值不合理，期望大于0，实际 %v", value)
				}
			} else {
				t.Errorf("数据类型错误，期望 float64，实际 %T", data)
			}
		})
	}
}

func TestMarketProviderGetDataInvalidDataSource(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 测试不存在的数据源
	_, err := provider.GetData(ctx, "nonexistent", "BTC.price")
	if err == nil {
		t.Error("应该返回错误，因为数据源不存在")
	}

	expectedError := "failed to get exchange nonexistent:"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("错误信息不匹配，期望包含 '%s'，实际 '%s'", expectedError, err.Error())
	}
}

func TestMarketProviderGetDataInvalidFieldFormat(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 测试无效字段格式（缺少点号）
	_, err := provider.GetData(ctx, "okx", "invalid_field")
	if err == nil {
		t.Error("应该返回错误，因为字段格式无效")
	}

	expectedError := "market field must be in format 'SYMBOL.FIELD', got: invalid_field"
	if err.Error() != expectedError {
		t.Errorf("错误信息不匹配，期望 '%s'，实际 '%s'", expectedError, err.Error())
	}
}

func TestMarketProviderGetDataSymbolMismatch(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 测试不存在的交易对
	_, err := provider.GetData(ctx, "okx", "INVALIDCOIN.price")
	if err == nil {
		t.Error("应该返回错误，因为交易对不存在")
		return
	}

	// 验证错误信息包含相关的错误提示
	expectedErrors := []string{"failed to get ticker data", "no ticker data found", "symbol mismatch"}
	found := false
	for _, expectedError := range expectedErrors {
		if strings.Contains(err.Error(), expectedError) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("错误信息不匹配，期望包含以下之一 %v，实际 '%s'", expectedErrors, err.Error())
	}
}

func TestMarketProviderGetDataUnknownField(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 测试未知字段
	_, err := provider.GetData(ctx, "okx", "BTC.unknown_field")
	if err == nil {
		t.Error("应该返回错误，因为字段未知")
	}

	expectedError := "unknown field: unknown_field"
	if err.Error() != expectedError {
		t.Errorf("错误信息不匹配，期望 '%s'，实际 '%s'", expectedError, err.Error())
	}
}

func TestMarketProviderGetDataConcurrency(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 并发测试
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			data, err := provider.GetData(ctx, "okx", "BTC.price")
			if err != nil {
				t.Errorf("并发获取数据失败: %v", err)
				return
			}

			if data == nil {
				t.Error("并发获取的数据不应为空")
				return
			}

			if value, ok := data.(float64); ok {
				// 验证数值合理性（价格应该大于0）
				if value <= 0 {
					t.Errorf("并发获取的值不合理，期望大于0，实际 %v", value)
				}
			} else {
				t.Errorf("并发获取的数据类型错误，期望 float64，实际 %T", data)
			}
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMarketProviderGetDataAllDataSources(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 只测试可用的数据源
	dataSources := []string{"okx"}
	fields := []string{"price", "volume", "high", "low"}

	for _, dataSource := range dataSources {
		for _, field := range fields {
			t.Run(dataSource+"_"+field, func(t *testing.T) {
				data, err := provider.GetData(ctx, dataSource, "BTC."+field)
				if err != nil {
					t.Errorf("获取 %s.%s 数据失败: %v", dataSource, field, err)
					return
				}

				if data == nil {
					t.Errorf("%s.%s 数据不应为空", dataSource, field)
					return
				}

				// 验证数据类型 - 所有字段都是 float64 类型
				if value, ok := data.(float64); ok {
					// 验证数值合理性（价格应该大于0）
					if value <= 0 {
						t.Errorf("%s.%s 数值不合理，期望大于0，实际 %v", dataSource, field, value)
					}
				} else {
					t.Errorf("%s.%s 类型错误，期望 float64，实际 %T", dataSource, field, data)
				}
			})
		}
	}
}

func TestMarketProviderGetDataWithContext(t *testing.T) {
	provider := NewMarketProvider()

	// 测试带 context 的数据获取
	ctx := context.Background()
	data, err := provider.GetData(ctx, "okx", "BTC.price")
	if err != nil {
		// 如果是限流错误，跳过测试
		if strings.Contains(err.Error(), "Too Many Requests") {
			t.Skipf("跳过测试：API 限流 - %v", err)
			return
		}
		t.Errorf("使用 context 获取数据失败: %v", err)
		return
	}

	if data == nil {
		t.Error("使用 context 获取的数据不应为空")
		return
	}

	if value, ok := data.(float64); ok {
		// 验证数值合理性（价格应该大于0）
		if value <= 0 {
			t.Errorf("使用 context 获取的值不合理，期望大于0，实际 %v", value)
		}
	} else {
		t.Errorf("使用 context 获取的数据类型错误，期望 float64，实际 %T", data)
	}
}

func TestMarketProviderGetDataWithContextTODO(t *testing.T) {
	provider := NewMarketProvider()

	// 测试使用 context.TODO() 获取数据
	data, err := provider.GetData(context.TODO(), "okx", "BTC.price")
	if err != nil {
		// 如果是限流错误，跳过测试
		if strings.Contains(err.Error(), "Too Many Requests") {
			t.Skipf("跳过测试：API 限流 - %v", err)
			return
		}
		t.Errorf("使用 context.TODO() 获取数据失败: %v", err)
		return
	}

	if data == nil {
		t.Error("使用 context.TODO() 获取的数据不应为空")
		return
	}

	if value, ok := data.(float64); ok {
		// 验证数值合理性（价格应该大于0）
		if value <= 0 {
			t.Errorf("使用 context.TODO() 获取的值不合理，期望大于0，实际 %v", value)
		}
	} else {
		t.Errorf("使用 context.TODO() 获取的数据类型错误，期望 float64，实际 %T", data)
	}
}
