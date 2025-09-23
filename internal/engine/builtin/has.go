package builtin

import (
	"context"
)

// HasBuiltin has函数实现
type HasBuiltin struct {
	*BaseBuiltin
}

// NewHasBuiltin 创建has函数
func NewHasBuiltin() *HasBuiltin {
	signature := Signature{
		Name:        "has",
		Description: "检查文本是否包含关键词",
		ReturnType:  "bool",
		Args: []ArgInfo{
			{
				Name:        "text",
				Type:        "string",
				Required:    true,
				Description: "要检查的文本",
			},
			{
				Name:        "keyword",
				Type:        "string",
				Required:    true,
				Description: "关键词",
			},
		},
	}

	return &HasBuiltin{
		BaseBuiltin: NewBaseBuiltin("has", "检查文本是否包含关键词", signature),
	}
}

// Execute 执行has函数
func (f *HasBuiltin) Execute(ctx context.Context, args []interface{}, evaluator Evaluator) (interface{}, error) {
	if err := f.ValidateArgs(args); err != nil {
		return nil, err
	}

	// 第一个参数应该是字符串
	text := toString(args[0])

	// 第二个参数应该是字符串
	keyword := toString(args[1])

	// 检查文本是否包含关键词
	return contains(text, keyword), nil
}
