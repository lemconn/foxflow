package dsl

import (
	"context"
	"fmt"

	"github.com/lemconn/foxflow/internal/data"
)

// Engine DSL引擎
type Engine struct {
	parser      *Parser
	evaluator   *Evaluator
	registry    *Registry
	dataAdapter *DataAdapter
}

// NewEngine 创建DSL引擎
func NewEngine(dataManager *data.Manager) *Engine {
	// 创建组件
	parser := NewParser()
	registry := DefaultRegistry()
	dataAdapter := NewDataAdapter(dataManager)
	evaluator := NewEvaluator(registry, dataAdapter)

	return &Engine{
		parser:      parser,
		evaluator:   evaluator,
		registry:    registry,
		dataAdapter: dataAdapter,
	}
}

// Parse 解析DSL表达式
func (e *Engine) Parse(expression string) (*Node, error) {
	return e.parser.Parse(expression)
}

// Validate 验证DSL表达式
func (e *Engine) Validate(expression string) error {
	return e.parser.Validate(expression)
}

// Execute 执行AST节点
func (e *Engine) Execute(ctx context.Context, node *Node) (interface{}, error) {
	return e.evaluator.Evaluate(ctx, node)
}

// ExecuteToBool 执行AST节点并返回布尔值
func (e *Engine) ExecuteToBool(ctx context.Context, node *Node) (bool, error) {
	return e.evaluator.EvaluateToBool(ctx, node)
}

// ExecuteExpression 解析并执行DSL表达式
func (e *Engine) ExecuteExpression(ctx context.Context, expression string) (interface{}, error) {
	// 解析表达式
	node, err := e.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	// 验证AST
	if err := e.evaluator.Validate(node); err != nil {
		return nil, fmt.Errorf("failed to validate AST: %w", err)
	}

	// 执行AST
	return e.Execute(ctx, node)
}

// ExecuteExpressionToBool 解析并执行DSL表达式，返回布尔值
func (e *Engine) ExecuteExpressionToBool(ctx context.Context, expression string) (bool, error) {
	// 解析表达式
	node, err := e.Parse(expression)
	if err != nil {
		return false, fmt.Errorf("failed to parse expression: %w", err)
	}

	// 验证AST
	if err := e.evaluator.Validate(node); err != nil {
		return false, fmt.Errorf("failed to validate AST: %w", err)
	}

	// 执行AST
	return e.ExecuteToBool(ctx, node)
}

// RegisterFunction 注册自定义函数
func (e *Engine) RegisterFunction(name string, fn Function) {
	e.registry.Register(name, fn)
}

// GetRegistry 获取函数注册表
func (e *Engine) GetRegistry() *Registry {
	return e.registry
}

// GetDataAdapter 获取数据适配器
func (e *Engine) GetDataAdapter() *DataAdapter {
	return e.dataAdapter
}

// GetEvaluator 获取求值器
func (e *Engine) GetEvaluator() *Evaluator {
	return e.evaluator
}
