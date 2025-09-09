package strategy

import (
	"context"
	"fmt"

	"github.com/lemconn/foxflow/internal/exchange"
)

// Strategy 策略接口
type Strategy interface {
	// 基础信息
	GetName() string
	GetDescription() string
	GetParameters() map[string]interface{}

	// 策略执行
	Evaluate(ctx context.Context, exchange exchange.Exchange, symbol string, params map[string]interface{}) (bool, error)

	// 参数验证
	ValidateParameters(params map[string]interface{}) error
}

// StrategyResult 策略执行结果
type StrategyResult struct {
	StrategyName string                 `json:"strategy_name"`
	Symbol       string                 `json:"symbol"`
	Result       bool                   `json:"result"`
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data"`
}

// Manager 策略管理器
type Manager struct {
	strategies map[string]Strategy
}

// NewManager 创建策略管理器
func NewManager() *Manager {
	manager := &Manager{
		strategies: make(map[string]Strategy),
	}

	// 注册默认策略
	manager.RegisterStrategy(NewVolumeStrategy())
	manager.RegisterStrategy(NewMACDStrategy())
	manager.RegisterStrategy(NewRSIStrategy())
	manager.RegisterStrategy(NewCandlesStrategy())
	manager.RegisterStrategy(NewNewsStrategy())

	return manager
}

// RegisterStrategy 注册策略
func (m *Manager) RegisterStrategy(strategy Strategy) {
	m.strategies[strategy.GetName()] = strategy
}

// GetStrategy 获取策略
func (m *Manager) GetStrategy(name string) (Strategy, bool) {
	strategy, exists := m.strategies[name]
	return strategy, exists
}

// GetAvailableStrategies 获取可用策略列表
func (m *Manager) GetAvailableStrategies() []string {
	var strategies []string
	for name := range m.strategies {
		strategies = append(strategies, name)
	}
	return strategies
}

// ExecuteStrategy 执行策略
func (m *Manager) ExecuteStrategy(ctx context.Context, name string, exchange exchange.Exchange, symbol string, params map[string]interface{}) (*StrategyResult, error) {
	strategy, exists := m.GetStrategy(name)
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", name)
	}

	// 验证参数
	if err := strategy.ValidateParameters(params); err != nil {
		return nil, fmt.Errorf("invalid parameters for strategy %s: %w", name, err)
	}

	// 执行策略
	result, err := strategy.Evaluate(ctx, exchange, symbol, params)
	if err != nil {
		return &StrategyResult{
			StrategyName: name,
			Symbol:       symbol,
			Result:       false,
			Message:      err.Error(),
		}, err
	}

	return &StrategyResult{
		StrategyName: name,
		Symbol:       symbol,
		Result:       result,
		Message:      "strategy executed successfully",
	}, nil
}
