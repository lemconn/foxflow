package data

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NewsModule 新闻数据模块
type NewsModule struct {
	name string
	news map[string]*NewsData
	mu   sync.RWMutex
}

// NewNewsModule 创建新闻数据模块
func NewNewsModule() *NewsModule {
	module := &NewsModule{
		name: "news",
		news: make(map[string]*NewsData),
	}

	module.initMockData()
	return module
}

// GetName 获取模块名称
func (m *NewsModule) GetName() string {
	return m.name
}

// GetData 获取数据
func (m *NewsModule) GetData(ctx context.Context, entity, field string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	newsData, exists := m.news[entity]
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

// GetHistoricalData 获取历史数据
func (m *NewsModule) GetHistoricalData(ctx context.Context, entity, field string, period int) ([]interface{}, error) {
	// 新闻数据通常不需要历史数据
	return nil, fmt.Errorf("news module does not support historical data")
}

// initMockData 初始化Mock数据
func (m *NewsModule) initMockData() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 初始化新闻数据
	m.news["theblockbeats"] = &NewsData{
		Source:         "theblockbeats",
		Title:          "SOL突破新高，市值创新纪录",
		Content:        "Solana代币SOL价格突破200美元大关，创下历史新高...",
		UpdateTime:     time.Now().Add(-5 * time.Minute),
		LastTitle:      "SOL突破新高，市值创新纪录",
		LastUpdateTime: time.Now().Add(-5 * time.Minute),
		Keywords:       []string{"SOL", "突破", "新高", "市值"},
		Sentiment:      "positive",
	}

	m.news["coindesk"] = &NewsData{
		Source:         "coindesk",
		Title:          "比特币ETF获批，市场反应积极",
		Content:        "美国证券交易委员会批准了首个比特币ETF...",
		UpdateTime:     time.Now().Add(-10 * time.Minute),
		LastTitle:      "比特币ETF获批，市场反应积极",
		LastUpdateTime: time.Now().Add(-10 * time.Minute),
		Keywords:       []string{"比特币", "ETF", "获批", "市场"},
		Sentiment:      "positive",
	}

	m.news["cointelegraph"] = &NewsData{
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
func (m *NewsModule) UpdateNewsData(source string, data *NewsData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.news[source] = data
}

// GetNewsData 获取新闻数据（用于测试）
func (m *NewsModule) GetNewsData(source string) (*NewsData, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, exists := m.news[source]
	return data, exists
}
