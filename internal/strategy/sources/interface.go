package data

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Module 数据模块接口
type Module interface {
	// GetName 获取模块名称
	GetName() string

	// GetData 获取数据
	GetData(ctx context.Context, entity, field string) (interface{}, error)

	// GetHistoricalData 获取历史数据
	GetHistoricalData(ctx context.Context, entity, field string, period int) ([]interface{}, error)
}

// CandlesData K线数据
type CandlesData struct {
	Symbol     string    `json:"symbol"`
	Open       float64   `json:"open"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Close      float64   `json:"close"`
	Volume     float64   `json:"volume"`
	Timestamp  time.Time `json:"timestamp"`
	LastPx     float64   `json:"last_px"`     // 最新价格
	LastVolume float64   `json:"last_volume"` // 最新成交量
}

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

// IndicatorsData 指标数据
type IndicatorsData struct {
	Symbol    string                 `json:"symbol"`
	Indicator string                 `json:"indicator"` // 指标名称：MACD, RSI, Volume等
	Value     float64                `json:"value"`     // 指标值
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"` // 额外元数据
}

// Manager 数据管理器
type Manager struct {
	modules map[string]Module
	mu      sync.RWMutex
}

// NewManager 创建数据管理器
func NewManager() *Manager {
	return &Manager{
		modules: make(map[string]Module),
	}
}

// RegisterModule 注册数据模块
func (m *Manager) RegisterModule(module Module) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modules[module.GetName()] = module
}

// GetModule 获取数据模块
func (m *Manager) GetModule(name string) (Module, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	module, exists := m.modules[name]
	if !exists {
		return nil, fmt.Errorf("data module not found: %s", name)
	}

	return module, nil
}

// GetData 获取数据
func (m *Manager) GetData(ctx context.Context, moduleName, entity, field string) (interface{}, error) {
	module, err := m.GetModule(moduleName)
	if err != nil {
		return nil, err
	}

	return module.GetData(ctx, entity, field)
}

// GetHistoricalData 获取历史数据
func (m *Manager) GetHistoricalData(ctx context.Context, moduleName, entity, field string, period int) ([]interface{}, error) {
	module, err := m.GetModule(moduleName)
	if err != nil {
		return nil, err
	}

	return module.GetHistoricalData(ctx, entity, field, period)
}

// ListModules 列出所有注册的模块
func (m *Manager) ListModules() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	modules := make([]string, 0, len(m.modules))
	for name := range m.modules {
		modules = append(modules, name)
	}

	return modules
}

// InitDefaultModules 初始化默认数据模块
func InitDefaultModules() *Manager {
	manager := NewManager()

	// 注册所有默认模块
	manager.RegisterModule(NewCandlesModule())
	manager.RegisterModule(NewNewsModule())
	manager.RegisterModule(NewIndicatorsModule())

	return manager
}
