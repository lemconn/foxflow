package provider

import (
	"context"
	"testing"
)

func TestKlineProviderGetData(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试基本数据获取
	data, err := provider.GetData(ctx, "okx", "BTC.close", "15m", 3)
	if err != nil {
		t.Errorf("获取K线数据失败: %v", err)
	}

	if data == nil {
		t.Error("K线数据不应为空")
	}

	// 验证返回的是数组
	if dataArray, ok := data.([]interface{}); ok {
		if len(dataArray) != 3 {
			t.Errorf("期望3个数据点，实际得到%d个", len(dataArray))
		}
		
		// 验证数据类型
		for i, value := range dataArray {
			if _, ok := value.(float64); !ok {
				t.Errorf("数据点[%d]类型错误，期望float64，实际%T", i, value)
			}
		}
	} else {
		t.Errorf("返回数据类型错误，期望[]interface{}，实际%T", data)
	}
}

func TestKlineProviderGetDataInvalidParams(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试缺少参数
	_, err := provider.GetData(ctx, "okx", "BTC.close")
	if err == nil {
		t.Error("应该返回错误，因为缺少必需参数")
	}

	// 测试无效参数类型
	_, err = provider.GetData(ctx, "okx", "BTC.close", "15m", "invalid")
	if err == nil {
		t.Error("应该返回错误，因为limit参数类型无效")
	}

	// 测试零limit
	_, err = provider.GetData(ctx, "okx", "BTC.close", "15m", 0)
	if err == nil {
		t.Error("应该返回错误，因为limit必须大于0")
	}

	// 测试负数limit
	_, err = provider.GetData(ctx, "okx", "BTC.close", "15m", -1)
	if err == nil {
		t.Error("应该返回错误，因为limit不能为负数")
	}
}

func TestKlineProviderGetDataInvalidField(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试无效字段格式
	_, err := provider.GetData(ctx, "okx", "invalid_field", "15m", 3)
	if err == nil {
		t.Error("应该返回错误，因为字段格式无效")
	}

	// 测试无效字段名
	_, err = provider.GetData(ctx, "okx", "BTC.invalid", "15m", 3)
	if err == nil {
		t.Error("应该返回错误，因为字段名无效")
	}
}

func TestKlineProviderGetDataInvalidDataSource(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试无效数据源
	_, err := provider.GetData(ctx, "invalid_source", "BTC.close", "15m", 3)
	if err == nil {
		t.Error("应该返回错误，因为数据源不存在")
	}
}

func TestKlineProviderGetDataDifferentIntervals(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试不同时间间隔
	intervals := []string{"1m", "5m", "15m", "1h"}
	
	for _, interval := range intervals {
		data, err := provider.GetData(ctx, "okx", "BTC.close", interval, 2)
		if err != nil {
			t.Errorf("获取%s间隔数据失败: %v", interval, err)
			continue
		}
		
		if data == nil {
			t.Errorf("%s间隔数据不应为空", interval)
			continue
		}
		
		if dataArray, ok := data.([]interface{}); ok {
			if len(dataArray) != 2 {
				t.Errorf("%s间隔期望2个数据点，实际得到%d个", interval, len(dataArray))
			}
		} else {
			t.Errorf("%s间隔返回数据类型错误，期望[]interface{}，实际%T", interval, data)
		}
	}
}

func TestKlineProviderGetDataDifferentSymbols(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试不同币种
	symbols := []string{"BTC", "ETH", "SOL"}
	
	for _, symbol := range symbols {
		data, err := provider.GetData(ctx, "okx", symbol+".close", "15m", 2)
		if err != nil {
			t.Errorf("获取%s数据失败: %v", symbol, err)
			continue
		}
		
		if data == nil {
			t.Errorf("%s数据不应为空", symbol)
			continue
		}
		
		if dataArray, ok := data.([]interface{}); ok {
			if len(dataArray) != 2 {
				t.Errorf("%s期望2个数据点，实际得到%d个", symbol, len(dataArray))
			}
		} else {
			t.Errorf("%s返回数据类型错误，期望[]interface{}，实际%T", symbol, data)
		}
	}
}

func TestKlineProviderGetDataDifferentFields(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试不同字段
	fields := []string{"open", "high", "low", "close", "volume"}
	
	for _, field := range fields {
		data, err := provider.GetData(ctx, "okx", "BTC."+field, "15m", 2)
		if err != nil {
			t.Errorf("获取%s字段数据失败: %v", field, err)
			continue
		}
		
		if data == nil {
			t.Errorf("%s字段数据不应为空", field)
			continue
		}
		
		if dataArray, ok := data.([]interface{}); ok {
			if len(dataArray) != 2 {
				t.Errorf("%s字段期望2个数据点，实际得到%d个", field, len(dataArray))
			}
		} else {
			t.Errorf("%s字段返回数据类型错误，期望[]interface{}，实际%T", field, data)
		}
	}
}

func TestKlineProviderGetDataWithDefaultInterval(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试使用默认时间间隔（不传递interval参数）
	data, err := provider.GetData(ctx, "okx", "BTC.close", "1m", 2)
	if err != nil {
		t.Errorf("使用默认间隔获取数据失败: %v", err)
	}

	if data == nil {
		t.Error("默认间隔数据不应为空")
	}

	if dataArray, ok := data.([]interface{}); ok {
		if len(dataArray) != 2 {
			t.Errorf("默认间隔期望2个数据点，实际得到%d个", len(dataArray))
		}
	} else {
		t.Errorf("默认间隔返回数据类型错误，期望[]interface{}，实际%T", data)
	}
}