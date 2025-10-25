package engine

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/engine/syntax"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/models"
)

// Engine 策略引擎
type Engine struct {
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	exchangeMgr   *exchange.Manager
	syntaxEngine  *syntax.Engine
	checkInterval time.Duration
	running       bool
	mu            sync.RWMutex
}

// NewEngine 创建策略引擎
func NewEngine() *Engine {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建新的语法引擎（不再需要数据管理器）
	syntaxEngine := syntax.NewEngine()

	return &Engine{
		ctx:           ctx,
		cancel:        cancel,
		exchangeMgr:   exchange.GetManager(),
		syntaxEngine:  syntaxEngine,
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
		userOrders[order.AccountID] = append(userOrders[order.AccountID], order)
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
	var user models.FoxAccount
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

	// 解析语法表达式
	node, err := e.syntaxEngine.Parse(order.Strategy)
	if err != nil {
		return fmt.Errorf("failed to parse strategy syntax: %w", err)
	}

	// 验证AST
	if err := e.syntaxEngine.GetEvaluator().Validate(node); err != nil {
		return fmt.Errorf("failed to validate AST: %w", err)
	}

	// 执行AST并获取布尔结果
	conditionResult, err := e.syntaxEngine.ExecuteToBool(e.ctx, node)
	if err != nil {
		return fmt.Errorf("failed to execute strategy AST: %w", err)
	}

	// 如果条件满足，提交订单
	if conditionResult {
		log.Printf("策略条件满足，提交订单: ID=%d, Strategy=%s", order.ID, order.Strategy)
		return e.submitOrder(exchangeInstance, order)
	}

	log.Printf("策略条件不满足，跳过订单: ID=%d, Strategy=%s", order.ID, order.Strategy)
	return nil
}

// submitOrder 提交订单到交易所
func (e *Engine) submitOrder(exchangeInstance exchange.Exchange, order *models.FoxSS) error {
	preCheckOrder := &exchange.OrderCostReq{
		Symbol:     order.Symbol,
		Amount:     order.Sz,
		AmountType: order.SzType,
		MarginType: order.MarginType,
	}

	preOrder, err := exchangeInstance.CalcOrderCost(e.ctx, preCheckOrder)
	if err != nil {
		return fmt.Errorf("failed to pre-check order cost: %w", err)
	}

	if !preOrder.CanBuyWithTaker {
		return fmt.Errorf("insufficient balance to place order")
	}

	if order.Type == "close" {
		closeExchangeOrder := &exchange.ClosePosition{
			Symbol:  order.Symbol,
			Margin:  order.MarginType,
			PosSide: order.PosSide,
		}
		err := exchangeInstance.ClosePosition(e.ctx, closeExchangeOrder)
		if err != nil {
			order.Msg = err.Error()
			order.Status = "failed"
			if err := database.GetDB().Save(order).Error; err != nil {
				return fmt.Errorf("failed to update order: %w", err)
			}
			log.Printf("平仓失败: ID=%d, OrderID=%s, Error=%s", order.ID, order.OrderID, err.Error())
			return fmt.Errorf("failed to close position: %w", err)
		}
		order.Status = "closed"
		if err := database.GetDB().Save(order).Error; err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}
		log.Printf("平仓成功: ID=%d, OrderID=%s", order.ID, order.OrderID)
	}

	if order.Type == "open" {
		exchangeOrder := &exchange.Order{
			Symbol:     order.Symbol,
			Side:       order.Side,
			PosSide:    order.PosSide,
			MarginType: order.MarginType,
			Price:      order.Px,
			Size:       preOrder.Contracts,
			Type:       order.OrderType,
		}
		result, err := exchangeInstance.CreateOrder(e.ctx, exchangeOrder)
		if err != nil {
			order.Msg = err.Error()
			order.Status = "failed"
			if err := database.GetDB().Save(order).Error; err != nil {
				return fmt.Errorf("failed to update order: %w", err)
			}
			log.Printf("开仓失败: ID=%d, OrderID=%s, Error=%s", order.ID, order.OrderID, err.Error())
			return fmt.Errorf("failed to open position: %w", err)
		}
		order.Status = "opened"
		if err := database.GetDB().Save(order).Error; err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}
		log.Printf("开仓成功: ID=%d, OrderID=%s", order.ID, result.ID)
	}
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
