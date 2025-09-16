package strategy

import (
	"context"
	"fmt"
	"sync"

	sources "github.com/lemconn/foxflow/internal/strategy/datasources"
	"github.com/lemconn/foxflow/internal/strategy/functions"
)

// Registry 注册器，支持函数和数据源注册
type Registry struct {
	functions   map[string]functions.Function
	dataSources map[string]sources.Module
	mu          sync.RWMutex
}

// NewRegistry 创建注册器
func NewRegistry() *Registry {
	return &Registry{
		functions:   make(map[string]functions.Function),
		dataSources: make(map[string]sources.Module),
	}
}

// RegisterFunction 注册函数
func (r *Registry) RegisterFunction(fn functions.Function) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.functions[fn.GetName()] = fn
}

// RegisterDataSource 注册数据源
func (r *Registry) RegisterDataSource(ds sources.Module) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dataSources[ds.GetName()] = ds
}

// GetFunction 获取函数
func (r *Registry) GetFunction(name string) (functions.Function, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, exists := r.functions[name]
	return fn, exists
}

// GetDataSource 获取数据源
func (r *Registry) GetDataSource(name string) (sources.Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ds, exists := r.dataSources[name]
	return ds, exists
}

// ListFunctions 列出所有函数
func (r *Registry) ListFunctions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}
	return names
}

// ListDataSources 列出所有数据源
func (r *Registry) ListDataSources() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.dataSources))
	for name := range r.dataSources {
		names = append(names, name)
	}
	return names
}

// GetData 获取数据（统一接口）
func (r *Registry) GetData(ctx context.Context, source, entity, field string) (interface{}, error) {
	ds, exists := r.GetDataSource(source)
	if !exists {
		return nil, fmt.Errorf("data source not found: %s", source)
	}

	return ds.GetData(ctx, entity, field)
}

// DefaultRegistry 创建默认注册器，注册所有默认函数和数据源
func DefaultRegistry() *Registry {
	registry := NewRegistry()

	// 注册默认函数
	registry.RegisterFunction(functions.NewAvgFunction())
	registry.RegisterFunction(functions.NewAgoFunction())
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
