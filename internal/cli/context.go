package cli

import (
	"context"
	"fmt"
	"foxflow/internal/exchange"
	"foxflow/internal/models"
	"foxflow/pkg/utils"
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
	return &Context{
		ctx: ctx,
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
		return fmt.Sprintf("%s %s ",
			utils.MessageGreen("foxflow"),
			">",
		)
	}

	if c.currentUser == nil {
		return fmt.Sprintf("%s %s ",
			utils.MessageGreen("foxflow "+utils.MessageYellow("["+c.currentExchange+"]")),
			">",
		)
	}

	return fmt.Sprintf("%s %s ",
		utils.MessageGreen("foxflow "+utils.MessageYellow("["+c.currentExchange+":"+c.currentUser.Username+"]")),
		">",
	)
}

// IsReady 检查是否已选择交易所和用户
func (c *Context) IsReady() bool {
	return c.currentExchange != "" && c.currentUser != nil
}

// GetContext 获取底层context
func (c *Context) GetContext() context.Context {
	return c.ctx
}
