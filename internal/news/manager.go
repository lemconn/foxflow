package news

import (
	"context"
	"fmt"
	"sync"
)

// Manager 新闻管理器实现
type Manager struct {
	sources map[string]NewsSource
	mutex   sync.RWMutex
}

// NewManager 创建新的新闻管理器
func NewManager() *Manager {
	return &Manager{
		sources: make(map[string]NewsSource),
	}
}

// RegisterSource 注册新闻源
func (m *Manager) RegisterSource(source NewsSource) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sources[source.GetName()] = source
}

// GetSource 根据名称获取新闻源
func (m *Manager) GetSource(name string) (NewsSource, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	source, exists := m.sources[name]
	return source, exists
}

// GetAllSources 获取所有已注册的新闻源
func (m *Manager) GetAllSources() map[string]NewsSource {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 返回副本以避免并发修改
	sources := make(map[string]NewsSource)
	for name, source := range m.sources {
		sources[name] = source
	}
	return sources
}

// GetNewsFromSource 从指定新闻源获取新闻
func (m *Manager) GetNewsFromSource(ctx context.Context, sourceName string, count int) ([]NewsItem, error) {
	source, exists := m.GetSource(sourceName)
	if !exists {
		return nil, fmt.Errorf("新闻源 '%s' 不存在", sourceName)
	}

	return source.GetNews(ctx, count)
}

// GetNewsFromAllSources 从所有新闻源获取新闻
func (m *Manager) GetNewsFromAllSources(ctx context.Context, count int) (map[string][]NewsItem, error) {
	sources := m.GetAllSources()
	results := make(map[string][]NewsItem)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for name, source := range sources {
		wg.Add(1)
		go func(sourceName string, newsSource NewsSource) {
			defer wg.Done()

			news, err := newsSource.GetNews(ctx, count)
			if err != nil {
				// 记录错误但不中断其他源的获取
				fmt.Printf("从新闻源 '%s' 获取新闻失败: %v\n", sourceName, err)
				return
			}

			mutex.Lock()
			results[sourceName] = news
			mutex.Unlock()
		}(name, source)
	}

	wg.Wait()
	return results, nil
}

// GetAvailableSources 获取所有可用的新闻源（健康检查）
func (m *Manager) GetAvailableSources(ctx context.Context) []string {
	sources := m.GetAllSources()
	var available []string

	for name, source := range sources {
		if source.IsHealthy(ctx) {
			available = append(available, name)
		}
	}

	return available
}
