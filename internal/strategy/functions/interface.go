package functions

import (
	"context"
)

// Function 函数接口定义
type Function interface {
	// GetName 获取函数名称
	GetName() string

	// GetDescription 获取函数描述
	GetDescription() string

	// GetSignature 获取函数签名信息
	GetSignature() Signature

	// Execute 执行函数
	Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error)
}

// Signature 函数签名信息
type Signature struct {
	// Name 函数名称
	Name string

	// Args 参数信息
	Args []ArgInfo

	// ReturnType 返回类型
	ReturnType string

	// Description 函数描述
	Description string
}

// ArgInfo 参数信息
type ArgInfo struct {
	// Name 参数名称
	Name string

	// Type 参数类型
	Type string

	// Required 是否必需
	Required bool

	// Description 参数描述
	Description string
}

// Evaluator 求值器接口，用于函数内部调用其他函数或获取数据
type Evaluator interface {
	// GetFieldValue 获取字段值
	GetFieldValue(ctx context.Context, module, entity, field string) (interface{}, error)

	// CallFunction 调用函数
	CallFunction(ctx context.Context, name string, args []interface{}) (interface{}, error)

	// GetHistoricalData 获取历史数据
	GetHistoricalData(ctx context.Context, source, entity, field string, period int) ([]interface{}, error)
}

// Registry 函数注册表接口
type Registry interface {
	// Register 注册函数
	Register(fn Function)

	// GetFunction 获取函数
	GetFunction(name string) (Function, bool)

	// ListFunctions 列出所有函数
	ListFunctions() []string
}
