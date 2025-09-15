package data

import (
	"context"
	"fmt"
	"sync"
)

// Module 数据模块接口
type Module interface {
	// GetName 获取模块名称
	GetName() string

	// GetData 获取数据
	GetData(ctx context.Context, entity, field string) (interface{}, error)
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
	manager.RegisterModule(NewKlineModule())
	manager.RegisterModule(NewMarketModule())
	manager.RegisterModule(NewNewsModule())
	manager.RegisterModule(NewIndicatorsModule())

	return manager
}
