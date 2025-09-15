package syntax

import (
	"context"

	"github.com/lemconn/foxflow/internal/strategy"
	"github.com/lemconn/foxflow/internal/strategy/functions"
)

// Function 函数签名（保持向后兼容）
type Function func(ctx context.Context, args []interface{}, evaluator functions.Evaluator) (interface{}, error)

// DataProvider 数据提供者接口（保持向后兼容）
type DataProvider interface {
	GetKline(ctx context.Context, symbol, field string) ([]float64, error)
	GetKlineField(ctx context.Context, symbol, field string) (interface{}, error)
	GetMarketField(ctx context.Context, symbol, field string) (interface{}, error)
	GetNewsField(ctx context.Context, source, field string) (interface{}, error)
	GetIndicatorField(ctx context.Context, symbol, field string) (interface{}, error)
}

// Registry 函数注册表（保持向后兼容）
type Registry struct {
	unifiedRegistry *strategy.UnifiedRegistry
}

// NewRegistry 创建函数注册表
func NewRegistry() *Registry {
	return &Registry{
		unifiedRegistry: strategy.DefaultRegistry(),
	}
}

// Register 注册函数（保持向后兼容）
func (r *Registry) Register(name string, fn Function) {
	// 将旧的函数签名转换为新的函数接口
	wrapper := &FunctionWrapper{
		name: name,
		fn:   fn,
	}
	r.unifiedRegistry.RegisterFunction(wrapper)
}

// GetFunction 获取函数（保持向后兼容）
func (r *Registry) GetFunction(name string) (Function, bool) {
	fn, exists := r.unifiedRegistry.GetFunction(name)
	if !exists {
		return nil, false
	}

	// 如果是包装器，返回原始函数
	if wrapper, ok := fn.(*FunctionWrapper); ok {
		return wrapper.fn, true
	}

	// 如果是新式函数，创建适配器
	adapter := func(ctx context.Context, args []interface{}, evaluator functions.Evaluator) (interface{}, error) {
		return fn.Execute(ctx, args, evaluator)
	}
	return adapter, true
}

// FunctionWrapper 函数包装器，用于向后兼容
type FunctionWrapper struct {
	name string
	fn   Function
}

// GetName 获取函数名称
func (fw *FunctionWrapper) GetName() string {
	return fw.name
}

// GetDescription 获取函数描述
func (fw *FunctionWrapper) GetDescription() string {
	return "Legacy function wrapper"
}

// GetSignature 获取函数签名
func (fw *FunctionWrapper) GetSignature() functions.Signature {
	return functions.Signature{
		Name:        fw.name,
		Description: "Legacy function wrapper",
		ReturnType:  "interface{}",
		Args:        []functions.ArgInfo{},
	}
}

// Execute 执行函数
func (fw *FunctionWrapper) Execute(ctx context.Context, args []interface{}, evaluator functions.Evaluator) (interface{}, error) {
	return fw.fn(ctx, args, evaluator)
}

// DefaultRegistry 创建默认函数注册表
func DefaultRegistry() *Registry {
	return NewRegistry()
}
