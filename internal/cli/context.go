package cli

import (
	"context"
	"foxflow/internal/database"
	"foxflow/internal/exchange"
	"foxflow/internal/models"
)

// Context CLI上下文
type Context struct {
	ctx             context.Context
	currentExchange string
	currentUser     *models.FoxUser
	exchange        exchange.Exchange
}

// NewContext 创建新的CLI上下文
func NewContext(ctx context.Context) *Context {
	cliCtx := &Context{
		ctx: ctx,
	}

	// 从数据库恢复激活状态
	cliCtx.restoreActiveState()

	return cliCtx
}

// restoreActiveState 从数据库恢复激活状态
func (c *Context) restoreActiveState() {
	db := database.GetDB()
	exchangeManager := exchange.GetManager()

	// 查找激活的交易所
	var activeExchange models.FoxExchange
	if err := db.Where("is_active = ?", true).First(&activeExchange).Error; err == nil {
		c.currentExchange = activeExchange.Name

		// 获取交易所实例
		if ex, err := exchangeManager.GetExchange(activeExchange.Name); err == nil {
			c.exchange = ex
		}

		// 查找激活的用户
		var activeUser models.FoxUser
		if err := db.Where("is_active = ? AND exchange = ?", true, activeExchange.Name).First(&activeUser).Error; err == nil {
			c.currentUser = &activeUser

			// 连接用户到交易所
			if err := exchangeManager.ConnectUser(c.ctx, activeExchange.Name, &activeUser); err == nil {
				// 连接成功，更新交易所实例
				if ex, err := exchangeManager.GetExchange(activeExchange.Name); err == nil {
					c.exchange = ex
				}
			}
		}
	}
}

// SetExchange 设置当前交易所
func (c *Context) SetExchange(exchangeName string) {
	c.currentExchange = exchangeName
}

// GetExchange 获取当前交易所名称
func (c *Context) GetExchange() string {
	return c.currentExchange
}

// SetUser 设置当前用户
func (c *Context) SetUser(user *models.FoxUser) {
	c.currentUser = user
}

// GetUser 获取当前用户
func (c *Context) GetUser() *models.FoxUser {
	return c.currentUser
}

// SetExchangeInstance 设置交易所实例
func (c *Context) SetExchangeInstance(ex exchange.Exchange) {
	c.exchange = ex
}

// GetExchangeInstance 获取交易所实例
func (c *Context) GetExchangeInstance() exchange.Exchange {
	return c.exchange
}

// GetPrompt 获取当前提示符
func (c *Context) GetPrompt() string {
	if c.currentExchange == "" {
		return "foxflow > "
	}

	if c.currentUser == nil {
		return "foxflow [" + c.currentExchange + "] > "
	}

	return "foxflow [" + c.currentExchange + ":" + c.currentUser.Username + "] > "
}

// IsReady 检查是否已选择交易所和用户
func (c *Context) IsReady() bool {
	return c.currentExchange != "" && c.currentUser != nil
}

// GetContext 获取底层context
func (c *Context) GetContext() context.Context {
	return c.ctx
}
