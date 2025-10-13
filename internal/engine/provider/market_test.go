package provider

import (
	"context"
	"testing"
	"time"
)

func TestNewMarketProvider(t *testing.T) {
	provider := NewMarketProvider()

	if provider == nil {
		t.Fatal("NewMarketProvider 返回 nil")
	}

	if provider.GetName() != "market" {
		t.Errorf("期望提供者名称为 'market'，实际得到 '%s'", provider.GetName())
	}

	if provider.market == nil {
		t.Error("market 映射不应为 nil")
	}

	// 验证初始化了 mock 数据
	expectedDataSources := []string{"okx", "binance", "gate"}
	for _, source := range expectedDataSources {
		if _, exists := provider.market[source]; !exists {
			t.Errorf("期望数据源 '%s' 存在，但未找到", source)
		}
	}
}

func TestMarketProviderGetDataValidFields(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 测试所有有效字段
	testCases := []struct {
		dataSource string
		field      string
		expected   interface{}
	}{
		{"okx", "BTC.last_px", 45500.0},
		{"okx", "BTC.last_volume", 500.0},
		{"okx", "BTC.bid", 45480.0},
		{"okx", "BTC.ask", 45520.0},
		{"binance", "BTC.last_px", 45600.0},
		{"binance", "BTC.last_volume", 600.0},
		{"binance", "BTC.bid", 45580.0},
		{"binance", "BTC.ask", 45620.0},
		{"gate", "BTC.last_px", 45700.0},
		{"gate", "BTC.last_volume", 700.0},
		{"gate", "BTC.bid", 45680.0},
		{"gate", "BTC.ask", 45720.0},
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

			// 对于 timestamp 字段，只检查类型
			if tc.field == "BTC.timestamp" {
				if _, ok := data.(time.Time); !ok {
					t.Errorf("timestamp 字段类型错误，期望 time.Time，实际 %T", data)
				}
				return
			}

			// 对于数值字段，检查具体值
			if value, ok := data.(float64); ok {
				if value != tc.expected {
					t.Errorf("期望值 %v，实际值 %v", tc.expected, value)
				}
			} else {
				t.Errorf("数据类型错误，期望 float64，实际 %T", data)
			}
		})
	}
}

func TestMarketProviderGetDataTimestamp(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 测试 timestamp 字段
	data, err := provider.GetData(ctx, "okx", "BTC.timestamp")
	if err != nil {
		t.Errorf("获取时间戳失败: %v", err)
		return
	}

	if data == nil {
		t.Error("时间戳不应为空")
		return
	}

	if timestamp, ok := data.(time.Time); ok {
		// 检查时间戳是否合理（不应该是零值）
		if timestamp.IsZero() {
			t.Error("时间戳不应为零值")
		}

		// 检查时间戳是否在合理范围内（最近1小时内）
		now := time.Now()
		if timestamp.After(now) || timestamp.Before(now.Add(-time.Hour)) {
			t.Errorf("时间戳不在合理范围内: %v", timestamp)
		}
	} else {
		t.Errorf("时间戳类型错误，期望 time.Time，实际 %T", data)
	}
}

func TestMarketProviderGetDataInvalidDataSource(t *testing.T) {
	provider := NewMarketProvider()
	ctx := context.Background()

	// 测试不存在的数据源
	_, err := provider.GetData(ctx, "nonexistent", "BTC.last_px")
	if err == nil {
		t.Error("应该返回错误，因为数据源不存在")
	}

	expectedError := "no market data found for data source: nonexistent"
	if err.Error() != expectedError {
		t.Errorf("错误信息不匹配，期望 '%s'，实际 '%s'", expectedError, err.Error())
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

	// 测试符号不匹配
	_, err := provider.GetData(ctx, "okx", "ETH.last_px")
	if err == nil {
		t.Error("应该返回错误，因为符号不匹配")
	}

	expectedError := "symbol mismatch: expected ETH, got BTC"
	if err.Error() != expectedError {
		t.Errorf("错误信息不匹配，期望 '%s'，实际 '%s'", expectedError, err.Error())
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

			data, err := provider.GetData(ctx, "okx", "BTC.last_px")
			if err != nil {
				t.Errorf("并发获取数据失败: %v", err)
				return
			}

			if data == nil {
				t.Error("并发获取的数据不应为空")
				return
			}

			if value, ok := data.(float64); ok {
				if value != 45500.0 {
					t.Errorf("并发获取的值错误，期望 45500.0，实际 %v", value)
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

	// 测试所有数据源的所有字段
	dataSources := []string{"okx", "binance", "gate"}
	fields := []string{"last_px", "last_volume", "bid", "ask", "timestamp"}

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

				// 验证数据类型
				if field == "timestamp" {
					if _, ok := data.(time.Time); !ok {
						t.Errorf("%s.%s 类型错误，期望 time.Time，实际 %T", dataSource, field, data)
					}
				} else {
					if _, ok := data.(float64); !ok {
						t.Errorf("%s.%s 类型错误，期望 float64，实际 %T", dataSource, field, data)
					}
				}
			})
		}
	}
}

func TestMarketProviderInitMockData(t *testing.T) {
	provider := NewMarketProvider()

	// 验证所有预期的数据源都已初始化
	expectedDataSources := map[string]struct {
		symbol     string
		lastPx     float64
		lastVolume float64
		bid        float64
		ask        float64
	}{
		"okx":     {"BTC", 45500.0, 500.0, 45480.0, 45520.0},
		"binance": {"BTC", 45600.0, 600.0, 45580.0, 45620.0},
		"gate":    {"BTC", 45700.0, 700.0, 45680.0, 45720.0},
	}

	for dataSource, expected := range expectedDataSources {
		marketData, exists := provider.market[dataSource]
		if !exists {
			t.Errorf("数据源 '%s' 不存在", dataSource)
			continue
		}

		if marketData.Symbol != expected.symbol {
			t.Errorf("数据源 '%s' 符号错误，期望 '%s'，实际 '%s'", dataSource, expected.symbol, marketData.Symbol)
		}

		if marketData.LastPx != expected.lastPx {
			t.Errorf("数据源 '%s' 最新价格错误，期望 %v，实际 %v", dataSource, expected.lastPx, marketData.LastPx)
		}

		if marketData.LastVolume != expected.lastVolume {
			t.Errorf("数据源 '%s' 最新成交量错误，期望 %v，实际 %v", dataSource, expected.lastVolume, marketData.LastVolume)
		}

		if marketData.Bid != expected.bid {
			t.Errorf("数据源 '%s' 买一价错误，期望 %v，实际 %v", dataSource, expected.bid, marketData.Bid)
		}

		if marketData.Ask != expected.ask {
			t.Errorf("数据源 '%s' 卖一价错误，期望 %v，实际 %v", dataSource, expected.ask, marketData.Ask)
		}

		// 检查时间戳不为零值
		if marketData.Timestamp.IsZero() {
			t.Errorf("数据源 '%s' 时间戳为零值", dataSource)
		}
	}
}

func TestMarketProviderGetDataWithContext(t *testing.T) {
	provider := NewMarketProvider()

	// 测试带 context 的数据获取
	ctx := context.Background()
	data, err := provider.GetData(ctx, "okx", "BTC.last_px")
	if err != nil {
		t.Errorf("使用 context 获取数据失败: %v", err)
		return
	}

	if data == nil {
		t.Error("使用 context 获取的数据不应为空")
		return
	}

	if value, ok := data.(float64); ok {
		if value != 45500.0 {
			t.Errorf("使用 context 获取的值错误，期望 45500.0，实际 %v", value)
		}
	} else {
		t.Errorf("使用 context 获取的数据类型错误，期望 float64，实际 %T", data)
	}
}

func TestMarketProviderGetDataWithNilContext(t *testing.T) {
	provider := NewMarketProvider()

	// 测试 nil context（虽然不推荐，但应该能正常工作）
	data, err := provider.GetData(nil, "okx", "BTC.last_px")
	if err != nil {
		t.Errorf("使用 nil context 获取数据失败: %v", err)
		return
	}

	if data == nil {
		t.Error("使用 nil context 获取的数据不应为空")
		return
	}

	if value, ok := data.(float64); ok {
		if value != 45500.0 {
			t.Errorf("使用 nil context 获取的值错误，期望 45500.0，实际 %v", value)
		}
	} else {
		t.Errorf("使用 nil context 获取的数据类型错误，期望 float64，实际 %T", data)
	}
}
