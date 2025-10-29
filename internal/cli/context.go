package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/grpc"
	"github.com/lemconn/foxflow/internal/models"
)

// Context CLI上下文
type Context struct {
	ctx              context.Context
	currentExchange  string
	currentAccount   string
	accountInstance  *models.FoxAccount
	exchange         exchange.Exchange
	exchangeInstance *models.FoxExchange
	grpcClient       *grpc.Client
	tokenExpiresAt   int64 // Token 过期时间
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
		var activeAccount models.FoxAccount
		if err := db.Where("is_active = ? AND exchange = ?", true, activeExchange.Name).First(&activeAccount).Error; err == nil {
			c.currentAccount = activeAccount.Name

			// 连接用户到交易所
			if err := exchangeManager.ConnectAccount(c.ctx, activeExchange.Name, &activeAccount); err == nil {
				// 连接成功，更新交易所实例
				if ex, err := exchangeManager.GetExchange(activeExchange.Name); err == nil {
					c.exchange = ex
				}
			}
		}
	}
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
func (c *Context) SetExchangeInstance(ex *models.FoxExchange) {
	c.exchangeInstance = ex
}

// GetExchangeInstance 获取交易所实例
func (c *Context) GetExchangeInstance() *models.FoxExchange {
	return c.exchangeInstance
}

// SetAccountInstance 设置当前用户
func (c *Context) SetAccountInstance(account *models.FoxAccount) {
	c.accountInstance = account
}

// GetAccountInstance 获取当前用户
func (c *Context) GetAccountInstance() *models.FoxAccount {
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
