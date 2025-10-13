package provider

import (
	"context"
	"fmt"
	"sync"
)

// Provider 数据提供者接口
type Provider interface {
	// GetName 获取提供者名称
	GetName() string

	// GetData 获取数据
	// params 为可选参数，使用索引方式传递
	// 如果不传递参数，则使用默认行为
	GetData(ctx context.Context, entity, field string, params ...interface{}) (interface{}, error)
}

// Manager 数据提供者管理器
type Manager struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewManager 创建数据提供者管理器
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider 注册数据提供者
func (m *Manager) RegisterProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.GetName()] = provider
}

// GetProvider 获取数据提供者
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("data provider not found: %s", name)
	}

	return provider, nil
}

// GetData 获取数据
func (m *Manager) GetData(ctx context.Context, providerName, entity, field string, params ...interface{}) (interface{}, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetData(ctx, entity, field, params...)
}

// ListProviders 列出所有注册的提供者
func (m *Manager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.providers))
	for name := range m.providers {
		providers = append(providers, name)
	}

	return providers
}

// InitDefaultProviders 初始化默认数据提供者
func InitDefaultProviders() *Manager {
	manager := NewManager()

	// 注册所有默认提供者
	manager.RegisterProvider(NewKlineProvider())
	manager.RegisterProvider(NewMarketProvider())
	manager.RegisterProvider(NewNewsProvider())

	return manager
}
