package news

import (
	"context"
	"time"
)

// NewsItem 统一的新闻数据结构体
type NewsItem struct {
	ID          string    `json:"id"`           // 新闻唯一标识
	Title       string    `json:"title"`        // 新闻标题
	Content     string    `json:"content"`      // 新闻内容
	URL         string    `json:"url"`          // 新闻链接
	Source      string    `json:"source"`       // 新闻源名称
	PublishedAt time.Time `json:"published_at"` // 发布时间
	Tags        []string  `json:"tags"`         // 标签列表
	ImageURL    string    `json:"image_url"`    // 图片链接
}

// NewsSource 新闻源接口规范
type NewsSource interface {
	// GetName 获取新闻源名称
	GetName() string

	// GetDisplayName 获取新闻源展示名称
	GetDisplayName() string

	// GetNews 获取新闻列表
	// count: 获取新闻数量
	// 返回新闻列表和可能的错误
	GetNews(ctx context.Context, count int) ([]NewsItem, error)

	// IsHealthy 检查新闻源是否健康可用
	IsHealthy(ctx context.Context) bool
}

// NewsManager 新闻管理器接口
type NewsManager interface {
	// RegisterSource 注册新闻源
	RegisterSource(source NewsSource)

	// GetSource 根据名称获取新闻源
	GetSource(name string) (NewsSource, bool)

	// GetAllSources 获取所有已注册的新闻源
	GetAllSources() map[string]NewsSource

	// GetNewsFromSource 从指定新闻源获取新闻
	GetNewsFromSource(ctx context.Context, sourceName string, count int) ([]NewsItem, error)

	// GetNewsFromAllSources 从所有新闻源获取新闻
	GetNewsFromAllSources(ctx context.Context, count int) (map[string][]NewsItem, error)
}
