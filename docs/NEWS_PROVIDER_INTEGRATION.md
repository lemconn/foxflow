# NewsProvider 集成真实数据源

## 概述

本次更新将 `internal/engine/provider/news.go` 中的 `NewsProvider` 从使用 Mock 数据改为集成 `internal/news` 包中的真实数据源，实现了自动化的新闻数据获取和更新。

## 主要更改

### 1. 结构体更新

```go
type NewsProvider struct {
    *BaseProvider
    news     map[string]*NewsData
    mu       sync.RWMutex
    manager  *news.Manager          // 新增：新闻管理器
    ctx      context.Context        // 新增：上下文控制
    cancel   context.CancelFunc     // 新增：取消函数
    stopChan chan struct{}          // 新增：停止通道
}
```

### 2. 初始化流程变更

**之前：**
```go
func NewNewsProvider() *NewsProvider {
    module := &NewsProvider{
        BaseProvider: NewBaseProvider("news"),
        news:         make(map[string]*NewsData),
    }
    module.initMockData()  // 使用 Mock 数据
    return module
}
```

**现在：**
```go
func NewNewsProvider() *NewsProvider {
    ctx, cancel := context.WithCancel(context.Background())
    
    module := &NewsProvider{
        BaseProvider: NewBaseProvider("news"),
        news:         make(map[string]*NewsData),
        manager:      news.NewManager(),  // 集成新闻管理器
        ctx:          ctx,
        cancel:       cancel,
        stopChan:     make(chan struct{}),
    }

    module.registerNewsSources()  // 注册新闻源
    go module.startNewsUpdater()  // 启动更新协程
    
    return module
}
```

### 3. 新增功能

#### 3.1 新闻源注册
```go
func (p *NewsProvider) registerNewsSources() {
    // 注册 BlockBeats 新闻源
    blockBeats := news.NewBlockBeats()
    p.manager.RegisterSource(blockBeats)
    
    // 可以在这里添加更多新闻源
    // 例如：coindesk, cointelegraph 等
}
```

#### 3.2 自动更新协程
```go
func (p *NewsProvider) startNewsUpdater() {
    ticker := time.NewTicker(5 * time.Minute) // 每5分钟更新一次
    defer ticker.Stop()
    
    // 立即执行一次更新
    p.updateNewsData()
    
    for {
        select {
        case <-ticker.C:
            p.updateNewsData()
        case <-p.ctx.Done():
            return
        case <-p.stopChan:
            return
        }
    }
}
```

#### 3.3 数据转换
```go
func (p *NewsProvider) convertToNewsData(sourceName string, item news.NewsItem) *NewsData {
    // 简单的关键词提取（基于标题和内容）
    keywords := p.extractKeywords(item.Title + " " + item.Content)
    
    // 简单的情感分析（基于关键词）
    sentiment := p.analyzeSentiment(item.Title + " " + item.Content)
    
    return &NewsData{
        Source:         sourceName,
        Title:          item.Title,
        Content:        item.Content,
        UpdateTime:     item.PublishedAt,
        LastTitle:      item.Title,
        LastUpdateTime: item.PublishedAt,
        Keywords:       keywords,
        Sentiment:      sentiment,
    }
}
```

#### 3.4 智能分析功能

**关键词提取：**
- 支持中英文加密货币相关关键词
- 包括：比特币、BTC、以太坊、ETH、SOL、DeFi、NFT、Web3 等

**情感分析：**
- 正面词汇：上涨、突破、新高、利好、增长、成功、批准、通过、积极
- 负面词汇：下跌、暴跌、利空、失败、拒绝、风险、警告、负面
- 返回：positive、negative、neutral

#### 3.5 协程管理
```go
func (p *NewsProvider) Stop() {
    p.cancel()
    close(p.stopChan)
}
```

## 使用方式

### 基本使用
```go
// 创建新闻提供者
provider := NewNewsProvider()
defer provider.Stop()

// 等待数据更新
time.Sleep(2 * time.Second)

// 获取数据
ctx := context.Background()
title, err := provider.GetData(ctx, "blockbeats", "title")
sentiment, err := provider.GetData(ctx, "blockbeats", "sentiment")
keywords, err := provider.GetData(ctx, "blockbeats", "keywords")
updateTime, err := provider.GetData(ctx, "blockbeats", "last_update_time")
```

### 支持的字段
- `title` / `last_title`: 新闻标题
- `last_update_time`: 最新更新时间
- `sentiment`: 情感分析结果
- `keywords`: 关键词列表

## 测试验证

### 运行测试
```bash
go test -v ./internal/engine/provider/ -run TestNewsProvider
```

### 测试结果
- ✅ 成功获取真实 BlockBeats 新闻数据
- ✅ 情感分析功能正常
- ✅ 关键词提取功能正常
- ✅ 协程启动和停止功能正常
- ✅ 数据自动更新功能正常

## 优势

1. **真实数据**: 使用真实的 BlockBeats API 数据，不再是 Mock 数据
2. **自动更新**: 每5分钟自动更新新闻数据，无需手动刷新
3. **智能分析**: 内置关键词提取和情感分析功能
4. **可扩展**: 易于添加更多新闻源（CoinDesk、CoinTelegraph 等）
5. **并发安全**: 使用读写锁保证并发安全
6. **资源管理**: 提供优雅的启动和停止机制

## 注意事项

1. **网络依赖**: 需要网络连接才能获取新闻数据
2. **API 限制**: 受 BlockBeats API 限制影响
3. **错误处理**: 网络错误不会影响程序运行，会记录错误并继续
4. **内存使用**: 协程会持续运行，需要适当的内存管理

## 未来扩展

1. 添加更多新闻源（CoinDesk、CoinTelegraph 等）
2. 改进关键词提取算法（使用 NLP 库）
3. 增强情感分析功能（使用机器学习模型）
4. 添加新闻缓存机制
5. 支持新闻分类和过滤功能
