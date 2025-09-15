package strategy

import (
	"context"
	"fmt"
	"sync"

	"github.com/lemconn/foxflow/internal/strategy/functions"
	sources "github.com/lemconn/foxflow/internal/strategy/sources"
)

// UnifiedRegistry 统一注册器，支持函数和数据源注册
type UnifiedRegistry struct {
	functions   map[string]functions.Function
	dataSources map[string]sources.Module
	mu          sync.RWMutex
}

// NewUnifiedRegistry 创建统一注册器
func NewUnifiedRegistry() *UnifiedRegistry {
	return &UnifiedRegistry{
		functions:   make(map[string]functions.Function),
		dataSources: make(map[string]sources.Module),
	}
}

// RegisterFunction 注册函数
func (r *UnifiedRegistry) RegisterFunction(fn functions.Function) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.functions[fn.GetName()] = fn
}

// RegisterDataSource 注册数据源
func (r *UnifiedRegistry) RegisterDataSource(ds sources.Module) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dataSources[ds.GetName()] = ds
}

// GetFunction 获取函数
func (r *UnifiedRegistry) GetFunction(name string) (functions.Function, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, exists := r.functions[name]
	return fn, exists
}

// GetDataSource 获取数据源
func (r *UnifiedRegistry) GetDataSource(name string) (sources.Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ds, exists := r.dataSources[name]
	return ds, exists
}

// ListFunctions 列出所有函数
func (r *UnifiedRegistry) ListFunctions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}
	return names
}

// ListDataSources 列出所有数据源
func (r *UnifiedRegistry) ListDataSources() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.dataSources))
	for name := range r.dataSources {
		names = append(names, name)
	}
	return names
}

// GetData 获取数据（统一接口）
func (r *UnifiedRegistry) GetData(ctx context.Context, source, entity, field string) (interface{}, error) {
	ds, exists := r.GetDataSource(source)
	if !exists {
		return nil, fmt.Errorf("data source not found: %s", source)
	}

	return ds.GetData(ctx, entity, field)
}

// DefaultRegistry 创建默认注册器，注册所有默认函数和数据源
func DefaultRegistry() *UnifiedRegistry {
	registry := NewUnifiedRegistry()

	// 注册默认函数
	registry.RegisterFunction(functions.NewAvgFunction())
	registry.RegisterFunction(functions.NewTimeSinceFunction())
	registry.RegisterFunction(functions.NewHasFunction())
	registry.RegisterFunction(functions.NewMaxFunction())
	registry.RegisterFunction(functions.NewMinFunction())
	registry.RegisterFunction(functions.NewSumFunction())

	// 注册默认数据源
	registry.RegisterDataSource(sources.NewKlineModule())
	registry.RegisterDataSource(sources.NewMarketModule())
	registry.RegisterDataSource(sources.NewNewsModule())
	registry.RegisterDataSource(sources.NewIndicatorsModule())

	return registry
}
