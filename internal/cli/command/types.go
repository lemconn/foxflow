package command

import (
	"context"

	"github.com/lemconn/foxflow/internal/grpc"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
)

const (
	TopCommandHelp   = "help"
	TopCommandShow   = "show"
	TopCommandUse    = "use"
	TopCommandCreate = "create"
	TopCommandUpdate = "update"
	TopCommandOpen   = "open"
	TopCommandClose  = "close"
	TopCommandCancel = "cancel"
	TopCommandDelete = "delete"
	TopCommandExit   = "exit"
	TopCommandQuit   = "quit"
)

// Context 定义命令执行所需的最小上下文接口
type Context interface {
	// 基础
	GetContext() context.Context

	// 交易所与用户状态
	IsReady() bool

	GetExchangeName() string
	SetExchangeName(exchangeName string)

	GetAccountName() string
	SetAccountName(user string)

	GetExchangeInstance() *model.FoxExchange
	SetExchangeInstance(ex *model.FoxExchange)

	GetAccountInstance() *model.FoxAccount
	SetAccountInstance(user *model.FoxAccount)

	// gRPC 客户端
	GetGRPCClient() *grpc.Client
}

// Command 命令接口（供各业务命令实现）
type Command interface {
	GetName() string
	GetDescription() string
	GetUsage() string
	Execute(ctx Context, args []string) error
}
