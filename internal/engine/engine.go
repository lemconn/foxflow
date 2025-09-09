package engine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/models"
	"github.com/lemconn/foxflow/internal/parser"
	"github.com/lemconn/foxflow/internal/strategy"
)

// Engine 策略引擎
type Engine struct {
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	strategyMgr   *strategy.Manager
	exchangeMgr   *exchange.Manager
	parser        *parser.StrategyParser
	checkInterval time.Duration
	running       bool
	mu            sync.RWMutex
}

// NewEngine 创建策略引擎
func NewEngine() *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		ctx:           ctx,
		cancel:        cancel,
		strategyMgr:   strategy.NewManager(),
		exchangeMgr:   exchange.GetManager(),
		parser:        parser.NewStrategyParser(),
		checkInterval: 5 * time.Second, // 每5秒检查一次
	}
}

// Start 启动引擎
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("engine is already running")
	}

	e.running = true
	log.Println("策略引擎启动")

	// 启动策略检查协程
	e.wg.Add(1)
	go e.run()

	return nil
}

// Stop 停止引擎
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return fmt.Errorf("engine is not running")
	}

	e.running = false
	log.Println("策略引擎停止中...")

	// 取消上下文
	e.cancel()

	// 等待所有协程结束
	e.wg.Wait()

	log.Println("策略引擎已停止")
	return nil
}

// run 运行策略检查循环
func (e *Engine) run() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			if err := e.checkStrategies(); err != nil {
				log.Printf("策略检查错误: %v", err)
			}
		}
	}
}

// checkStrategies 检查所有等待中的策略订单
func (e *Engine) checkStrategies() error {
	db := database.GetDB()

	// 获取所有等待中的策略订单
	var orders []models.FoxSS
	if err := db.Where("status = ?", "waiting").Find(&orders).Error; err != nil {
		return fmt.Errorf("failed to get waiting orders: %w", err)
	}

	// 按用户分组处理
	userOrders := make(map[uint][]models.FoxSS)
	for _, order := range orders {
		userOrders[order.UserID] = append(userOrders[order.UserID], order)
	}

	// 处理每个用户的订单
	for userID, userOrderList := range userOrders {
		if err := e.processUserOrders(userID, userOrderList); err != nil {
			log.Printf("处理用户 %d 订单时出错: %v", userID, err)
		}
	}

	return nil
}

// processUserOrders 处理单个用户的订单
func (e *Engine) processUserOrders(userID uint, orders []models.FoxSS) error {
	if len(orders) == 0 {
		return nil
	}

	// 获取用户信息
	db := database.GetDB()
	var user models.FoxUser
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 获取交易所实例
	exchangeInstance, err := e.exchangeMgr.GetExchange(user.Exchange)
	if err != nil {
		return fmt.Errorf("failed to get exchange: %w", err)
	}

	// 连接用户到交易所
	if err := exchangeInstance.Connect(e.ctx, &user); err != nil {
		return fmt.Errorf("failed to connect user to exchange: %w", err)
	}

	// 处理每个订单
	for _, order := range orders {
		if err := e.processOrder(exchangeInstance, &order); err != nil {
			log.Printf("处理订单 %d 时出错: %v", order.ID, err)
		}
	}

	return nil
}

// processOrder 处理单个订单
func (e *Engine) processOrder(exchangeInstance exchange.Exchange, order *models.FoxSS) error {
	// 如果没有策略，直接提交订单
	if order.Strategy == "" {
		return e.submitOrder(exchangeInstance, order)
	}

	// 解析策略表达式
	condition, err := e.parser.Parse(order.Strategy)
	if err != nil {
		return fmt.Errorf("failed to parse strategy: %w", err)
	}

	// 获取策略名称和参数
	strategyNames := e.parser.GetStrategyNames(condition)
	parameters := e.parser.GetParameters(condition)

	// 执行所有策略
	results := make(map[string]bool)
	for _, strategyKey := range strategyNames {
		// 解析策略键：candles.SOL.last_px -> candles
		strategyName, err := e.extractStrategyName(strategyKey)
		if err != nil {
			return fmt.Errorf("failed to extract strategy name from key %s: %w", strategyKey, err)
		}

		strategyInstance, exists := e.strategyMgr.GetStrategy(strategyName)
		if !exists {
			return fmt.Errorf("strategy not found: %s", strategyName)
		}

		result, err := strategyInstance.Evaluate(e.ctx, exchangeInstance, order.Symbol, parameters[strategyKey])
		if err != nil {
			return fmt.Errorf("failed to evaluate strategy %s: %w", strategyName, err)
		}

		results[strategyKey] = result
	}

	// 评估策略条件
	conditionResult, err := e.parser.Evaluate(condition, results)
	if err != nil {
		return fmt.Errorf("failed to evaluate condition: %w", err)
	}

	// 如果条件满足，提交订单
	if conditionResult {
		log.Printf("策略条件满足，提交订单: ID=%d", order.ID)
		return e.submitOrder(exchangeInstance, order)
	}

	return nil
}

// extractStrategyName 从策略键中提取策略名称
func (e *Engine) extractStrategyName(strategyKey string) (string, error) {
	// 策略键格式：candles.SOL.last_px -> candles
	// 或者：news.theblockbeats.last_title -> news
	parts := strings.Split(strategyKey, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid strategy key format: %s", strategyKey)
	}

	strategyName := parts[0]

	// 验证策略名称
	validStrategies := []string{"candles", "news", "volume", "macd", "rsi"}
	for _, valid := range validStrategies {
		if strategyName == valid {
			return strategyName, nil
		}
	}

	return "", fmt.Errorf("unknown strategy: %s", strategyName)
}

// submitOrder 提交订单到交易所
func (e *Engine) submitOrder(exchangeInstance exchange.Exchange, order *models.FoxSS) error {
	// 构建订单对象
	exchangeOrder := &exchange.Order{
		Symbol:  order.Symbol,
		Side:    order.Side,
		PosSide: order.PosSide,
		Price:   order.Px,
		Size:    order.Sz,
		Type:    order.OrderType,
	}

	// 提交订单
	result, err := exchangeInstance.CreateOrder(e.ctx, exchangeOrder)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// 更新数据库
	db := database.GetDB()
	order.OrderID = result.ID
	order.Status = "pending"

	if err := db.Save(order).Error; err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	log.Printf("订单已提交: ID=%d, OrderID=%s", order.ID, result.ID)
	return nil
}

// GetStatus 获取引擎状态
func (e *Engine) GetStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]interface{}{
		"running":        e.running,
		"check_interval": e.checkInterval.String(),
	}
}

// SetCheckInterval 设置检查间隔
func (e *Engine) SetCheckInterval(interval time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.checkInterval = interval
}
