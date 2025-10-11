package provider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lemconn/foxflow/internal/news"
)

// NewsData 新闻数据
type NewsData struct {
	Title    string    `json:"title"`     // 新闻标题
	Content  string    `json:"content"`   // 新闻内容
	Datetime time.Time `json:"datetime"`  // 发布时间
}

// NewsProvider 新闻数据模块
type NewsProvider struct {
	*BaseProvider
	news     map[string]*NewsData
	mu       sync.RWMutex
	manager  *news.Manager
	ctx      context.Context
	cancel   context.CancelFunc
	stopChan chan struct{}
}

// NewNewsProvider 创建新闻数据模块
func NewNewsProvider() *NewsProvider {
	ctx, cancel := context.WithCancel(context.Background())
	
	module := &NewsProvider{
		BaseProvider: NewBaseProvider("news"),
		news:         make(map[string]*NewsData),
		manager:      news.NewManager(),
		ctx:          ctx,
		cancel:       cancel,
		stopChan:     make(chan struct{}),
	}

	// 注册新闻源
	module.registerNewsSources()
	
	// 启动协程定期更新新闻数据
	go module.startNewsUpdater()
	
	return module
}

// registerNewsSources 注册所有可用的新闻源
func (p *NewsProvider) registerNewsSources() {
	// 注册 BlockBeats 新闻源
	blockBeats := news.NewBlockBeats()
	p.manager.RegisterSource(blockBeats)
	
	// 可以在这里添加更多新闻源
	// 例如：coindesk, cointelegraph 等
}

// startNewsUpdater 启动新闻更新协程
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

// updateNewsData 更新新闻数据
func (p *NewsProvider) updateNewsData() {
	ctx, cancel := context.WithTimeout(p.ctx, 30*time.Second)
	defer cancel()
	
	// 从所有新闻源获取最新新闻
	allNews, err := p.manager.GetNewsFromAllSources(ctx, 1) // 每个源获取1条最新新闻
	if err != nil {
		fmt.Printf("获取新闻数据失败: %v\n", err)
		return
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// 更新每个新闻源的数据
	for sourceName, newsItems := range allNews {
		if len(newsItems) > 0 {
			// 获取最新的一条新闻
			latestNews := newsItems[0]
			
			// 直接转换为 NewsData 格式
			p.news[sourceName] = &NewsData{
				Title:    latestNews.Title,
				Content:  latestNews.Content,
				Datetime: latestNews.PublishedAt,
			}
		}
	}
}




// GetData 获取数据
// NewsProvider 只支持单个数据值，不支持历史数据
// params 参数（可选）：
// - 目前暂未使用，保留用于未来扩展
func (p *NewsProvider) GetData(ctx context.Context, dataSource, field string, params ...DataParam) (interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	newsData, exists := p.news[dataSource]
	if !exists {
		return nil, fmt.Errorf("no news data found for data source: %s", dataSource)
	}

	// News 模块支持简单字段名，不需要多级字段
	switch field {
	case "title":
		return newsData.Title, nil
	case "content":
		return newsData.Content, nil
	case "datetime":
		return newsData.Datetime, nil
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}
}

// Stop 停止新闻更新协程
func (p *NewsProvider) Stop() {
	p.cancel()
	close(p.stopChan)
}


// GetFunctionParamMapping 获取函数参数映射
func (p *NewsProvider) GetFunctionParamMapping() map[string]FunctionParamInfo {
	// News 模块目前不需要函数参数
	return map[string]FunctionParamInfo{}
}
