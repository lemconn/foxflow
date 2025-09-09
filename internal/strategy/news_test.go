package strategy

import (
	"context"
	"testing"
	"time"
)

func TestNewsStrategy_Evaluate(t *testing.T) {
	strategy := NewNewsStrategy()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		expected    bool
		description string
	}{
		{
			name: "title_contains_breakthrough",
			params: map[string]interface{}{
				"source":   "theblockbeats",
				"field":    "last_title",
				"operator": "contains",
				"value":    "突破",
			},
			expectError: false,
			expected:    true, // Mock数据中包含"突破"
			description: "标题包含'突破'关键词",
		},
		{
			name: "title_in_list",
			params: map[string]interface{}{
				"source":   "theblockbeats",
				"field":    "last_title",
				"operator": "in",
				"values":   []interface{}{"突破新高", "重大利好"},
			},
			expectError: false,
			expected:    true, // Mock数据中标题是"SOL突破新高，市值创新纪录"
			description: "标题在指定列表中",
		},
		{
			name: "title_not_in_list",
			params: map[string]interface{}{
				"source":   "theblockbeats",
				"field":    "last_title",
				"operator": "not_in",
				"values":   []interface{}{"下跌", "利空"},
			},
			expectError: false,
			expected:    true, // Mock数据中标题不在"下跌"、"利空"列表中
			description: "标题不在指定列表中",
		},
		{
			name: "update_time_less_than_10_minutes",
			params: map[string]interface{}{
				"source":   "theblockbeats",
				"field":    "last_update_time",
				"operator": "<",
				"value":    10.0, // 10分钟前
			},
			expectError: false,
			expected:    true, // Mock数据中更新时间是5分钟前
			description: "更新时间小于10分钟",
		},
		{
			name: "sentiment_positive",
			params: map[string]interface{}{
				"source":   "theblockbeats",
				"field":    "sentiment",
				"operator": "==",
				"value":    "positive",
			},
			expectError: false,
			expected:    true, // Mock数据中情感是"positive"
			description: "情感分析为正面",
		},
		{
			name: "invalid_source",
			params: map[string]interface{}{
				"source":   "invalid_source",
				"field":    "last_title",
				"operator": "contains",
				"value":    "test",
			},
			expectError: true,
			expected:    false,
			description: "无效的新闻源",
		},
		{
			name: "invalid_field",
			params: map[string]interface{}{
				"source":   "theblockbeats",
				"field":    "invalid_field",
				"operator": "contains",
				"value":    "test",
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

func TestNewsStrategy_UpdateMockData(t *testing.T) {
	strategy := NewNewsStrategy()

	// 更新Mock数据
	newData := &NewsData{
		Source:         "test_source",
		Title:          "测试新闻标题",
		Content:        "测试新闻内容",
		UpdateTime:     time.Now(),
		LastTitle:      "测试新闻标题",
		LastUpdateTime: time.Now(),
		Keywords:       []string{"测试", "新闻"},
		Sentiment:      "neutral",
	}

	strategy.UpdateMockData("test_source", newData)

	// 验证更新
	retrievedData, exists := strategy.GetMockData("test_source")
	if !exists {
		t.Fatal("无法获取更新的Mock数据")
	}

	if retrievedData.Source != "test_source" {
		t.Errorf("新闻源不匹配，期望: test_source, 实际: %s", retrievedData.Source)
	}

	if retrievedData.LastTitle != "测试新闻标题" {
		t.Errorf("标题不匹配，期望: 测试新闻标题, 实际: %s", retrievedData.LastTitle)
	}
}

func TestNewsStrategy_GetParameters(t *testing.T) {
	strategy := NewNewsStrategy()

	params := strategy.GetParameters()
	if params == nil {
		t.Fatal("参数为空")
	}

	// 验证必需参数
	requiredParams := []string{"source", "field", "operator", "value", "values"}
	for _, param := range requiredParams {
		if _, exists := params[param]; !exists {
			t.Errorf("缺少必需参数: %s", param)
		}
	}
}

func TestNewsStrategy_GetName(t *testing.T) {
	strategy := NewNewsStrategy()

	name := strategy.GetName()
	if name != "news" {
		t.Errorf("策略名称不匹配，期望: news, 实际: %s", name)
	}
}

func TestNewsStrategy_GetDescription(t *testing.T) {
	strategy := NewNewsStrategy()

	description := strategy.GetDescription()
	if description == "" {
		t.Error("策略描述为空")
	}
}

func TestNewsStrategy_TimeComparison(t *testing.T) {
	strategy := NewNewsStrategy()

	// 测试时间比较（使用秒数）
	params := map[string]interface{}{
		"source":   "theblockbeats",
		"field":    "last_update_time",
		"operator": ">",
		"value":    600.0, // 10分钟 = 600秒
	}

	result, err := strategy.Evaluate(context.Background(), nil, "SOL", params)
	if err != nil {
		t.Fatalf("时间比较测试失败: %v", err)
	}

	// 获取Mock数据进行调试
	newsData, exists := strategy.GetMockData("theblockbeats")
	if !exists {
		t.Fatal("无法获取Mock数据")
	}

	t.Logf("Mock数据更新时间: %v", newsData.LastUpdateTime)
	t.Logf("当前时间: %v", time.Now())
	t.Logf("10分钟前: %v", time.Now().Add(-600*time.Second))
	t.Logf("比较结果: %v", result)

	if !result {
		t.Error("时间比较结果应该为true（Mock数据是5分钟前，应该大于10分钟前）")
	}
}
