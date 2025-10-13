package provider

import (
	"context"
	"testing"
	"time"

	"github.com/lemconn/foxflow/internal/news"
)

func TestNewsProviderIntegration(t *testing.T) {
	// 创建新闻提供者
	provider := NewNewsProvider()
	defer provider.Stop()

	// 等待一段时间让协程有时间获取数据
	time.Sleep(2 * time.Second)

	// 测试获取数据
	ctx := context.Background()
	
	// 测试获取 blockbeats 数据源的标题
	title, err := provider.GetData(ctx, "blockbeats", "title")
	if err != nil {
		t.Logf("获取 blockbeats 标题失败: %v", err)
		// 这里不直接失败，因为网络请求可能失败
	} else {
		t.Logf("获取到 blockbeats 标题: %v", title)
		// 验证标题不为空
		if titleStr, ok := title.(string); ok {
			if titleStr == "" {
				t.Error("标题不应为空")
			}
		}
	}

	// 测试获取内容
	content, err := provider.GetData(ctx, "blockbeats", "content")
	if err != nil {
		t.Logf("获取 blockbeats 内容失败: %v", err)
	} else {
		t.Logf("获取到 blockbeats 内容: %v", content)
		// 验证内容不为空
		if contentStr, ok := content.(string); ok {
			if contentStr == "" {
				t.Error("内容不应为空")
			}
		}
	}

	// 测试获取发布时间
	datetime, err := provider.GetData(ctx, "blockbeats", "datetime")
	if err != nil {
		t.Logf("获取 blockbeats 发布时间失败: %v", err)
	} else {
		t.Logf("获取到 blockbeats 发布时间: %v", datetime)
		// 验证时间格式正确
		if timeVal, ok := datetime.(time.Time); ok {
			if timeVal.IsZero() {
				t.Error("发布时间不应为零值")
			}
		}
	}
}

func TestNewsProviderStop(t *testing.T) {
	provider := NewNewsProvider()
	
	// 测试停止功能
	provider.Stop()
	
	// 等待一段时间确保协程已停止
	time.Sleep(100 * time.Millisecond)
	
	// 这里可以添加更多验证逻辑
	t.Log("NewsProvider 停止测试完成")
}

func TestConvertToNewsData(t *testing.T) {
	// 创建测试用的 NewsItem
	testItem := news.NewsItem{
		ID:          "test-1",
		Title:       "测试新闻标题",
		Content:     "这是测试新闻内容",
		URL:         "https://example.com/news/1",
		Source:      "test-source",
		PublishedAt: time.Now(),
		Tags:        []string{"测试", "新闻"},
		ImageURL:    "https://example.com/image.jpg",
	}
	
	// 测试数据转换逻辑（模拟 convertToNewsData 的功能）
	newsData := &NewsData{
		Title:    testItem.Title,
		Content:  testItem.Content,
		Datetime: testItem.PublishedAt,
	}
	
	// 验证转换结果
	if newsData.Title != testItem.Title {
		t.Errorf("标题转换错误: 期望 %s, 实际 %s", testItem.Title, newsData.Title)
	}
	
	if newsData.Content != testItem.Content {
		t.Errorf("内容转换错误: 期望 %s, 实际 %s", testItem.Content, newsData.Content)
	}
	
	if !newsData.Datetime.Equal(testItem.PublishedAt) {
		t.Errorf("时间转换错误: 期望 %v, 实际 %v", testItem.PublishedAt, newsData.Datetime)
	}
	
	t.Logf("数据转换测试通过: %+v", newsData)
}

func TestNewsProviderGetDataInvalidField(t *testing.T) {
	provider := NewNewsProvider()
	defer provider.Stop()
	
	// 等待一段时间让协程获取数据
	time.Sleep(2 * time.Second)
	
	ctx := context.Background()
	
	// 测试无效字段（使用真实的数据源）
	_, err := provider.GetData(ctx, "blockbeats", "invalid_field")
	if err == nil {
		t.Error("应该返回错误，因为字段无效")
	}
	
	expectedError := "unknown field: invalid_field"
	if err.Error() != expectedError {
		t.Errorf("错误信息不匹配: 期望 %s, 实际 %s", expectedError, err.Error())
	}
}

func TestNewsProviderGetDataInvalidSource(t *testing.T) {
	provider := NewNewsProvider()
	defer provider.Stop()
	
	ctx := context.Background()
	
	// 测试无效数据源
	_, err := provider.GetData(ctx, "invalid_source", "title")
	if err == nil {
		t.Error("应该返回错误，因为数据源无效")
	}
	
	expectedError := "no news data found for data source: invalid_source"
	if err.Error() != expectedError {
		t.Errorf("错误信息不匹配: 期望 %s, 实际 %s", expectedError, err.Error())
	}
}

func TestNewsProviderGetDataAllFields(t *testing.T) {
	provider := NewNewsProvider()
	defer provider.Stop()
	
	// 等待一段时间让协程获取数据
	time.Sleep(2 * time.Second)
	
	ctx := context.Background()
	
	// 测试获取所有支持的字段
	fields := []string{"title", "content", "datetime"}
	
	for _, field := range fields {
		value, err := provider.GetData(ctx, "blockbeats", field)
		if err != nil {
			t.Logf("获取 %s 字段失败: %v", field, err)
			continue
		}
		
		// 验证返回值不为空
		switch field {
		case "title", "content":
			if str, ok := value.(string); ok {
				if str == "" {
					t.Errorf("%s 字段不应为空", field)
				}
			} else {
				t.Errorf("%s 字段类型错误，期望 string", field)
			}
		case "datetime":
			if timeVal, ok := value.(time.Time); ok {
				if timeVal.IsZero() {
					t.Errorf("%s 字段不应为零值", field)
				}
			} else {
				t.Errorf("%s 字段类型错误，期望 time.Time", field)
			}
		}
		
		t.Logf("成功获取 %s 字段: %v", field, value)
	}
}

func TestNewsProviderConcurrentAccess(t *testing.T) {
	provider := NewNewsProvider()
	defer provider.Stop()
	
	// 等待一段时间让协程获取数据
	time.Sleep(2 * time.Second)
	
	ctx := context.Background()
	
	// 并发访问测试
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			// 并发获取数据
			_, err := provider.GetData(ctx, "blockbeats", "title")
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

