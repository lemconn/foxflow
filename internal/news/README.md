# 新闻模块 (News Module)

这个模块提供了统一的新闻数据获取接口，支持从不同的新闻源获取新闻数据。

## 功能特性

- 统一的新闻数据接口
- 支持多个新闻源
- 自动分页获取大量新闻
- 健康检查机制
- 并发安全的新闻管理器

## 核心组件

### 1. 接口定义 (`interface.go`)

- `NewsItem`: 统一的新闻数据结构
- `NewsSource`: 新闻源接口规范
- `NewsManager`: 新闻管理器接口

### 2. BlockBeats 新闻源 (`blockbeats.go`)

实现了 BlockBeats 新闻源的完整功能：
- 自动分页获取新闻
- HTML 内容清理
- 统一的链接格式处理
- 完整的请求头设置

### 3. 新闻管理器 (`manager.go`)

提供新闻源的统一管理：
- 新闻源注册和管理
- 并发安全的新闻获取
- 健康检查功能

## 使用方法

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/lemconn/foxflow/internal/news"
)

func main() {
    // 创建新闻管理器
    manager := news.NewManager()
    
    // 注册 BlockBeats 新闻源
    blockBeats := news.NewBlockBeats()
    manager.RegisterSource(blockBeats)
    
    // 创建上下文
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // 获取新闻
    newsList, err := manager.GetNewsFromSource(ctx, "blockbeats", 10)
    if err != nil {
        fmt.Printf("获取新闻失败: %v\n", err)
        return
    }
    
    // 处理新闻
    for _, item := range newsList {
        fmt.Printf("标题: %s\n", item.Title)
        fmt.Printf("来源: %s\n", item.Source)
        fmt.Printf("时间: %s\n", item.PublishedAt.Format("2006-01-02 15:04:05"))
        fmt.Printf("链接: %s\n", item.URL)
        fmt.Println("---")
    }
}
```

### 从所有新闻源获取新闻

```go
// 获取所有新闻源的新闻
allNews, err := manager.GetNewsFromAllSources(ctx, 5)
if err != nil {
    fmt.Printf("获取新闻失败: %v\n", err)
    return
}

for sourceName, newsList := range allNews {
    fmt.Printf("新闻源: %s, 新闻数量: %d\n", sourceName, len(newsList))
}
```

### 健康检查

```go
// 检查可用的新闻源
availableSources := manager.GetAvailableSources(ctx)
fmt.Printf("可用的新闻源: %v\n", availableSources)
```

## 扩展新的新闻源

要实现新的新闻源，只需要实现 `NewsSource` 接口：

```go
type MyNewsSource struct {
    // 你的字段
}

func (m *MyNewsSource) GetName() string {
    return "mynews"
}

func (m *MyNewsSource) GetDisplayName() string {
    return "My News"
}

func (m *MyNewsSource) GetNews(ctx context.Context, count int) ([]NewsItem, error) {
    // 实现获取新闻的逻辑
    return newsList, nil
}

func (m *MyNewsSource) IsHealthy(ctx context.Context) bool {
    // 实现健康检查逻辑
    return true
}
```

然后注册到管理器中：

```go
mySource := &MyNewsSource{}
manager.RegisterSource(mySource)
```

## 测试

运行测试：

```bash
go test ./internal/news/... -v
```

## 注意事项

1. 所有网络请求都有超时控制
2. BlockBeats 新闻源会自动处理分页
3. HTML 内容会被自动清理
4. 新闻链接会根据 API 返回的 URL 字段自动生成
5. 支持并发获取多个新闻源的数据
