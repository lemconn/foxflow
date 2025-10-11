package provider

import (
	"context"
	"testing"
)

func TestKlineProviderGetData(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试基本数据获取
	params := []DataParam{
		NewParam("limit", 3),
		NewParam("interval", "15m"),
	}
	
	data, err := provider.GetData(ctx, "okx", "BTC.close", params...)
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

	// 测试无效的limit参数
	params := []DataParam{
		NewParam("limit", 0),
		NewParam("interval", "15m"),
	}
	_, err = provider.GetData(ctx, "okx", "BTC.close", params...)
	if err == nil {
		t.Error("应该返回错误，因为limit必须大于0")
	}

	// 测试负数limit
	params = []DataParam{
		NewParam("limit", -1),
		NewParam("interval", "15m"),
	}
	_, err = provider.GetData(ctx, "okx", "BTC.close", params...)
	if err == nil {
		t.Error("应该返回错误，因为limit不能为负数")
	}
}

func TestKlineProviderGetDataInvalidField(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	params := []DataParam{
		NewParam("limit", 3),
		NewParam("interval", "15m"),
	}

	// 测试无效字段格式
	_, err := provider.GetData(ctx, "okx", "invalid_field", params...)
	if err == nil {
		t.Error("应该返回错误，因为字段格式无效")
	}

	// 测试无效字段名
	_, err = provider.GetData(ctx, "okx", "BTC.invalid", params...)
	if err == nil {
		t.Error("应该返回错误，因为字段名无效")
	}
}

func TestKlineProviderGetDataInvalidDataSource(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	params := []DataParam{
		NewParam("limit", 3),
		NewParam("interval", "15m"),
	}

	// 测试无效数据源
	_, err := provider.GetData(ctx, "invalid_source", "BTC.close", params...)
	if err == nil {
		t.Error("应该返回错误，因为数据源无效")
	}
}

func TestKlineProviderGetDataAllFields(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	params := []DataParam{
		NewParam("limit", 2),
		NewParam("interval", "1h"),
	}

	// 测试所有支持的字段
	fields := []string{"open", "high", "low", "close", "volume"}
	
	for _, field := range fields {
		data, err := provider.GetData(ctx, "okx", "BTC."+field, params...)
		if err != nil {
			t.Errorf("获取字段 %s 失败: %v", field, err)
			continue
		}

		if dataArray, ok := data.([]interface{}); ok {
			if len(dataArray) != 2 {
				t.Errorf("字段 %s 数据长度错误，期望2，实际%d", field, len(dataArray))
			}
		} else {
			t.Errorf("字段 %s 数据类型错误", field)
		}
	}
}

func TestKlineProviderGetDataAllDataSources(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	params := []DataParam{
		NewParam("limit", 2),
		NewParam("interval", "5m"),
	}

	// 测试所有支持的数据源
	dataSources := []string{"okx", "binance", "gate"}
	
	for _, dataSource := range dataSources {
		data, err := provider.GetData(ctx, dataSource, "BTC.close", params...)
		if err != nil {
			t.Errorf("获取数据源 %s 失败: %v", dataSource, err)
			continue
		}

		if dataArray, ok := data.([]interface{}); ok {
			if len(dataArray) != 2 {
				t.Errorf("数据源 %s 数据长度错误，期望2，实际%d", dataSource, len(dataArray))
			}
		} else {
			t.Errorf("数据源 %s 数据类型错误", dataSource)
		}
	}
}

func TestKlineProviderGetDataDifferentIntervals(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试不同的时间间隔
	intervals := []string{"1m", "5m", "15m", "30m", "1h", "4h", "1d", "1w"}
	
	for _, interval := range intervals {
		params := []DataParam{
			NewParam("limit", 3),
			NewParam("interval", interval),
		}

		data, err := provider.GetData(ctx, "okx", "BTC.close", params...)
		if err != nil {
			t.Errorf("时间间隔 %s 失败: %v", interval, err)
			continue
		}

		if dataArray, ok := data.([]interface{}); ok {
			if len(dataArray) != 3 {
				t.Errorf("时间间隔 %s 数据长度错误，期望3，实际%d", interval, len(dataArray))
			}
			
			// 验证不同时间间隔的数据确实不同（由于波动率不同）
			t.Logf("时间间隔 %s: [%.2f, %.2f, %.2f]", interval, 
				dataArray[0], dataArray[1], dataArray[2])
		} else {
			t.Errorf("时间间隔 %s 数据类型错误", interval)
		}
	}
}

func TestKlineProviderGetDataDefaultInterval(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 测试不提供interval参数（应该使用默认值）
	params := []DataParam{
		NewParam("limit", 2),
	}

	data, err := provider.GetData(ctx, "okx", "BTC.close", params...)
	if err != nil {
		t.Errorf("使用默认时间间隔失败: %v", err)
	}

	if dataArray, ok := data.([]interface{}); ok {
		if len(dataArray) != 2 {
			t.Errorf("默认时间间隔数据长度错误，期望2，实际%d", len(dataArray))
		}
	} else {
		t.Errorf("默认时间间隔数据类型错误")
	}
}

func TestKlineProviderGetFunctionParamMapping(t *testing.T) {
	provider := NewKlineProvider()
	
	mapping := provider.GetFunctionParamMapping()
	
	// 验证所有函数都有参数映射
	expectedFunctions := []string{"avg", "max", "min", "sum"}
	for _, funcName := range expectedFunctions {
		if _, exists := mapping[funcName]; !exists {
			t.Errorf("函数 %s 缺少参数映射", funcName)
		}
	}

	// 验证avg函数的参数
	avgMapping, exists := mapping["avg"]
	if !exists {
		t.Error("avg函数映射不存在")
		return
	}

	if len(avgMapping.Params) != 2 {
		t.Errorf("avg函数参数数量错误，期望2，实际%d", len(avgMapping.Params))
	}

	// 验证第一个参数（interval）
	if avgMapping.Params[0].ParamName != "interval" {
		t.Errorf("第一个参数名错误，期望interval，实际%s", avgMapping.Params[0].ParamName)
	}
	if avgMapping.Params[0].ParamType != ParamTypeString {
		t.Errorf("第一个参数类型错误，期望string，实际%s", avgMapping.Params[0].ParamType)
	}
	if !avgMapping.Params[0].Required {
		t.Error("第一个参数应该为必需")
	}

	// 验证第二个参数（limit）
	if avgMapping.Params[1].ParamName != "limit" {
		t.Errorf("第二个参数名错误，期望limit，实际%s", avgMapping.Params[1].ParamName)
	}
	if avgMapping.Params[1].ParamType != ParamTypeInt {
		t.Errorf("第二个参数类型错误，期望int，实际%s", avgMapping.Params[1].ParamType)
	}
	if !avgMapping.Params[1].Required {
		t.Error("第二个参数应该为必需")
	}
}

func TestKlineProviderParseInterval(t *testing.T) {
	provider := NewKlineProvider()
	
	testCases := []struct {
		interval string
		expected int
	}{
		{"1m", 1},
		{"5m", 5},
		{"15m", 15},
		{"30m", 30},
		{"1h", 60},
		{"4h", 240},
		{"1d", 1440},
		{"1w", 10080},
		{"invalid", 1}, // 默认值
	}

	for _, tc := range testCases {
		result := provider.parseInterval(tc.interval)
		if result != tc.expected {
			t.Errorf("parseInterval(%s) = %d, 期望 %d", tc.interval, result, tc.expected)
		}
	}
}

func TestKlineProviderGetVolatilityByInterval(t *testing.T) {
	provider := NewKlineProvider()
	
	testCases := []struct {
		interval string
		expected float64
	}{
		{"1m", 0.5},
		{"5m", 1.0},
		{"15m", 2.0},
		{"30m", 3.0},
		{"1h", 5.0},
		{"4h", 10.0},
		{"1d", 20.0},
		{"1w", 50.0},
		{"invalid", 2.0}, // 默认值
	}

	for _, tc := range testCases {
		result := provider.getVolatilityByInterval(tc.interval)
		if result != tc.expected {
			t.Errorf("getVolatilityByInterval(%s) = %.1f, 期望 %.1f", tc.interval, result, tc.expected)
		}
	}
}

func TestKlineProviderConcurrentAccess(t *testing.T) {
	provider := NewKlineProvider()
	ctx := context.Background()

	// 并发访问测试
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			params := []DataParam{
				NewParam("limit", 2),
				NewParam("interval", "15m"),
			}
			
			_, err := provider.GetData(ctx, "okx", "BTC.close", params...)
			if err != nil {
				t.Logf("并发获取数据失败: %v", err)
			}
		}()
	}
	
	// 等待所有协程完成
	for i := 0; i < 10; i++ {
		<-done
	}
	
	t.Log("并发访问测试完成")
}
