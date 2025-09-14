package functions

import (
	"fmt"
)

// BaseFunction 基础函数实现
type BaseFunction struct {
	name        string
	description string
	signature   Signature
}

// GetName 获取函数名称
func (f *BaseFunction) GetName() string {
	return f.name
}

// GetDescription 获取函数描述
func (f *BaseFunction) GetDescription() string {
	return f.description
}

// GetSignature 获取函数签名
func (f *BaseFunction) GetSignature() Signature {
	return f.signature
}

// NewBaseFunction 创建基础函数
func NewBaseFunction(name, description string, signature Signature) *BaseFunction {
	return &BaseFunction{
		name:        name,
		description: description,
		signature:   signature,
	}
}

// ValidateArgs 验证参数数量和类型
func (f *BaseFunction) ValidateArgs(args []interface{}) error {
	expectedCount := len(f.signature.Args)
	if len(args) != expectedCount {
		return fmt.Errorf("function %s expects %d arguments, got %d", f.name, expectedCount, len(args))
	}

	// 这里可以添加更详细的类型验证
	for i, arg := range f.signature.Args {
		if arg.Required && i < len(args) && args[i] == nil {
			return fmt.Errorf("required argument %s cannot be nil", arg.Name)
		}
	}

	return nil
}
