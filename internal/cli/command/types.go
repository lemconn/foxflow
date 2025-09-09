package command

import (
	"context"

	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/models"
)

// Context 定义命令执行所需的最小上下文接口
type Context interface {
	// 基础
	GetContext() context.Context

	// 交易所与用户状态
	IsReady() bool
	GetExchange() string
	SetExchange(exchangeName string)

	GetUser() *models.FoxUser
	SetUser(user *models.FoxUser)

	GetExchangeInstance() exchange.Exchange
	SetExchangeInstance(ex exchange.Exchange)
}

// Command 命令接口（供各业务命令实现）
type Command interface {
	GetName() string
	GetDescription() string
	GetUsage() string
	Execute(ctx Context, args []string) error
}
