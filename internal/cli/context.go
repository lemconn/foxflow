package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/grpc"
)

// Context CLI上下文
type Context struct {
	ctx              context.Context
	currentExchange  string
	exchangeInstance *grpc.ShowExchangeItem
	currentAccount   string
	accountInstance  *grpc.ShowAccountItem
	exchange         exchange.Exchange
	grpcClient       *grpc.Client
	tokenExpiresAt   int64 // Token 过期时间
}

// NewContext 创建新的CLI上下文
func NewContext(ctx context.Context) *Context {
	cliCtx := &Context{
		ctx: ctx,
	}

	return cliCtx
}

// SetExchangeName 设置当前交易所
func (c *Context) SetExchangeName(exchangeName string) {
	c.currentExchange = strings.TrimSpace(exchangeName)
}

// GetExchangeName 获取当前交易所名称
func (c *Context) GetExchangeName() string {
	return c.currentExchange
}

// SetAccountName 设置当前用户
func (c *Context) SetAccountName(accountName string) {
	c.currentAccount = strings.TrimSpace(accountName)
}

// GetAccountName 获取当前用户
func (c *Context) GetAccountName() string {
	return c.currentAccount
}

// SetExchangeInstance 设置交易所实例
func (c *Context) SetExchangeInstance(ex *grpc.ShowExchangeItem) {
	c.exchangeInstance = ex
}

// GetExchangeInstance 获取交易所实例
func (c *Context) GetExchangeInstance() *grpc.ShowExchangeItem {
	return c.exchangeInstance
}

// SetAccountInstance 设置当前用户
func (c *Context) SetAccountInstance(account *grpc.ShowAccountItem) {
	c.accountInstance = account
}

// GetAccountInstance 获取当前用户
func (c *Context) GetAccountInstance() *grpc.ShowAccountItem {
	return c.accountInstance
}

// IsReady 检查是否已选择交易所和用户
func (c *Context) IsReady() bool {
	return c.currentExchange != "" && c.currentAccount != ""
}

// GetContext 获取底层context
func (c *Context) GetContext() context.Context {
	return c.ctx
}

// SetGRPCClient 设置 gRPC 客户端
func (c *Context) SetGRPCClient(client *grpc.Client) {
	c.grpcClient = client
}

// GetGRPCClient 获取 gRPC 客户端
func (c *Context) GetGRPCClient() *grpc.Client {
	return c.grpcClient
}

// SetTokenExpiry 设置 token 过期时间
func (c *Context) SetTokenExpiry(expiresAt int64) {
	c.tokenExpiresAt = expiresAt
}

// GetTokenExpiry 获取 token 过期时间
func (c *Context) GetTokenExpiry() int64 {
	return c.tokenExpiresAt
}

// IsTokenExpired 检查 token 是否过期
func (c *Context) IsTokenExpired() bool {
	if c.tokenExpiresAt == 0 {
		return true
	}
	return time.Now().Unix() >= c.tokenExpiresAt
}

// GetTokenStatus 获取 token 状态信息
func (c *Context) GetTokenStatus() string {
	if c.grpcClient == nil {
		return "本地模式"
	}

	if c.tokenExpiresAt == 0 {
		return "未认证"
	}

	expiresAt := time.Unix(c.tokenExpiresAt, 0)
	now := time.Now()

	if now.After(expiresAt) {
		return "已过期"
	}

	// 检查是否在 1 小时内过期
	oneHourFromNow := now.Add(1 * time.Hour)
	if expiresAt.Before(oneHourFromNow) {
		return fmt.Sprintf("即将过期 (%s)", expiresAt.Format("15:04:05"))
	}

	return fmt.Sprintf("有效至 %s", expiresAt.Format("15:04:05"))
}
