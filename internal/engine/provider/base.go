package provider


// BaseProvider 基础提供者，提供通用的参数处理方法
type BaseProvider struct {
	name string
}

// NewBaseProvider 创建基础提供者
func NewBaseProvider(name string) *BaseProvider {
	return &BaseProvider{name: name}
}

// GetName 获取提供者名称
func (b *BaseProvider) GetName() string {
	return b.name
}

