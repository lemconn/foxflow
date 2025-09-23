package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NewsData 新闻数据
type NewsData struct {
	Source         string    `json:"source"`           // 新闻源，如 theblockbeats
	Title          string    `json:"title"`            // 新闻标题
	Content        string    `json:"content"`          // 新闻内容
	UpdateTime     time.Time `json:"update_time"`      // 更新时间
	LastTitle      string    `json:"last_title"`       // 最新标题
	LastUpdateTime time.Time `json:"last_update_time"` // 最新更新时间
	Keywords       []string  `json:"keywords"`         // 关键词
	Sentiment      string    `json:"sentiment"`        // 情感分析：positive, negative, neutral
}

// NewsProvider 新闻数据模块
type NewsProvider struct {
	*BaseProvider
	news map[string]*NewsData
	mu   sync.RWMutex
}

// NewNewsProvider 创建新闻数据模块
func NewNewsProvider() *NewsProvider {
	module := &NewsProvider{
		BaseProvider: NewBaseProvider("news"),
		news:       make(map[string]*NewsData),
	}

	module.initMockData()
	return module
}

// GetData 获取数据
// NewsProvider 只支持单个数据值，不支持历史数据
// params 参数（可选）：
// - 目前暂未使用，保留用于未来扩展
func (p *NewsProvider) GetData(ctx context.Context, entity, field string, params ...DataParam) (interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	newsData, exists := p.news[entity]
	if !exists {
		return nil, fmt.Errorf("no news data found for entity: %s", entity)
	}

	switch field {
	case "last_title":
		return newsData.LastTitle, nil
	case "last_update_time":
		return newsData.LastUpdateTime, nil
	case "sentiment":
		return newsData.Sentiment, nil
	case "keywords":
		return newsData.Keywords, nil
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}
}

// initMockData 初始化Mock数据
func (p *NewsProvider) initMockData() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 初始化新闻数据
	p.news["theblockbeats"] = &NewsData{
		Source:         "theblockbeats",
		Title:          "SOL突破新高，市值创新纪录",
		Content:        "Solana代币SOL价格突破200美元大关，创下历史新高...",
		UpdateTime:     time.Now().Add(-5 * time.Minute),
		LastTitle:      "SOL突破新高，市值创新纪录",
		LastUpdateTime: time.Now().Add(-5 * time.Minute),
		Keywords:       []string{"SOL", "突破", "新高", "市值"},
		Sentiment:      "positive",
	}

	p.news["coindesk"] = &NewsData{
		Source:         "coindesk",
		Title:          "比特币ETF获批，市场反应积极",
		Content:        "美国证券交易委员会批准了首个比特币ETF...",
		UpdateTime:     time.Now().Add(-10 * time.Minute),
		LastTitle:      "比特币ETF获批，市场反应积极",
		LastUpdateTime: time.Now().Add(-10 * time.Minute),
		Keywords:       []string{"比特币", "ETF", "获批", "市场"},
		Sentiment:      "positive",
	}

	p.news["cointelegraph"] = &NewsData{
		Source:         "cointelegraph",
		Title:          "以太坊2.0升级进展顺利",
		Content:        "以太坊网络升级进展顺利，交易费用显著降低...",
		UpdateTime:     time.Now().Add(-15 * time.Minute),
		LastTitle:      "以太坊2.0升级进展顺利",
		LastUpdateTime: time.Now().Add(-15 * time.Minute),
		Keywords:       []string{"以太坊", "升级", "交易费用", "网络"},
		Sentiment:      "positive",
	}
}

// UpdateNewsData 更新新闻数据（用于测试）
func (p *NewsProvider) UpdateNewsData(source string, data *NewsData) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.news[source] = data
}

// GetNewsData 获取新闻数据（用于测试）
func (p *NewsProvider) GetNewsData(source string) (*NewsData, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data, exists := p.news[source]
	return data, exists
}

// GetFunctionParamMapping 获取函数参数映射
func (p *NewsProvider) GetFunctionParamMapping() map[string]FunctionParamInfo {
	// News 模块目前不需要函数参数
	return map[string]FunctionParamInfo{}
}
