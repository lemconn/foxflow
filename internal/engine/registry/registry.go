package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/lemconn/foxflow/internal/engine/builtin"
	"github.com/lemconn/foxflow/internal/engine/provider"
)

// Registry 注册器，支持内置函数和数据提供者注册
type Registry struct {
	builtins  map[string]builtin.Builtin
	providers map[string]provider.Provider
	mu        sync.RWMutex
}

// NewRegistry 创建注册器
func NewRegistry() *Registry {
	return &Registry{
		builtins:  make(map[string]builtin.Builtin),
		providers: make(map[string]provider.Provider),
	}
}

// RegisterBuiltin 注册内置函数
func (r *Registry) RegisterBuiltin(fn builtin.Builtin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builtins[fn.GetName()] = fn
}

// RegisterProvider 注册数据提供者
func (r *Registry) RegisterProvider(ds provider.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[ds.GetName()] = ds
}

// GetBuiltin 获取函数
func (r *Registry) GetBuiltin(name string) (builtin.Builtin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, exists := r.builtins[name]
	return fn, exists
}

// GetProvider 获取数据源
func (r *Registry) GetProvider(name string) (provider.Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ds, exists := r.providers[name]
	return ds, exists
}

// ListBuiltins 列出所有函数
func (r *Registry) ListBuiltins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.builtins))
	for name := range r.builtins {
		names = append(names, name)
	}
	return names
}

// ListProviders 列出所有数据源
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// GetData 获取数据（统一接口）
func (r *Registry) GetData(ctx context.Context, source, entity, field string, params ...provider.DataParam) (interface{}, error) {
	ds, exists := r.GetProvider(source)
	if !exists {
		return nil, fmt.Errorf("data source not found: %s", source)
	}

	return ds.GetData(ctx, entity, field, params...)
}

// DefaultRegistry 创建默认注册器，注册所有默认函数和数据源
func DefaultRegistry() *Registry {
	registry := NewRegistry()

	// 注册默认函数
	registry.RegisterBuiltin(builtin.NewAvgBuiltin())
	registry.RegisterBuiltin(builtin.NewAgoBuiltin())
	registry.RegisterBuiltin(builtin.NewHasBuiltin())
	registry.RegisterBuiltin(builtin.NewMaxBuiltin())
	registry.RegisterBuiltin(builtin.NewMinBuiltin())
	registry.RegisterBuiltin(builtin.NewSumBuiltin())

	// 注册默认数据源
	registry.RegisterProvider(provider.NewKlineProvider())
	registry.RegisterProvider(provider.NewMarketProvider())
	registry.RegisterProvider(provider.NewNewsProvider())
	registry.RegisterProvider(provider.NewIndicatorsProvider())

	return registry
}
