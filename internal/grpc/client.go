package grpc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/lemconn/foxflow/internal/news"
	pb "github.com/lemconn/foxflow/proto/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client gRPC客户端
type Client struct {
	conn        *grpc.ClientConn
	client      pb.FoxFlowServiceClient
	accessToken string
	expiresAt   int64
	mu          sync.RWMutex
}

// NewClient 创建新的gRPC客户端
func NewClient(host string, port int) (*Client, error) {
	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := pb.NewFoxFlowServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Authenticate 认证
func (c *Client) Authenticate(username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.client.Authenticate(ctx, &pb.AuthRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		// 检查是否是连接错误
		if isConnectionError(err) {
			return fmt.Errorf("connection failed: %w", err)
		}
		return fmt.Errorf("authentication failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("authentication failed: %s", resp.Message)
	}

	// 保存 token 信息
	c.mu.Lock()
	c.accessToken = resp.AccessToken
	c.expiresAt = resp.ExpiresAt
	c.mu.Unlock()

	log.Printf("gRPC认证成功: %s, token 过期时间: %s", username, time.Unix(resp.ExpiresAt, 0).Format("2006-01-02 15:04:05"))
	return nil
}

// isConnectionError 检查是否是连接错误
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "unreachable") ||
		strings.Contains(errStr, "refused")
}

// RefreshToken 刷新 token
func (c *Client) RefreshToken() error {
	c.mu.RLock()
	currentToken := c.accessToken
	c.mu.RUnlock()

	if currentToken == "" {
		return fmt.Errorf("没有可用的 token 进行刷新")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.RefreshToken(ctx, &pb.RefreshTokenRequest{
		AccessToken: currentToken,
	})
	if err != nil {
		return fmt.Errorf("刷新 token 失败: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("刷新 token 失败: %s", resp.Message)
	}

	// 更新 token 信息
	c.mu.Lock()
	c.accessToken = resp.AccessToken
	c.expiresAt = resp.ExpiresAt
	c.mu.Unlock()

	log.Printf("Token 刷新成功, 新过期时间: %s", time.Unix(resp.ExpiresAt, 0).Format("2006-01-02 15:04:05"))
	return nil
}

// isTokenExpired 检查 token 是否即将过期（1小时内）
func (c *Client) isTokenExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.accessToken == "" || c.expiresAt == 0 {
		return true
	}

	// 检查是否在 1 小时内过期
	expiresAt := time.Unix(c.expiresAt, 0)
	oneHourFromNow := time.Now().Add(1 * time.Hour)

	return expiresAt.Before(oneHourFromNow)
}

// ensureValidToken 确保 token 有效，如果即将过期则自动刷新
func (c *Client) ensureValidToken() error {
	if c.isTokenExpired() {
		if err := c.RefreshToken(); err != nil {
			return fmt.Errorf("token 刷新失败: %w", err)
		}
	}
	return nil
}

// getAccessToken 获取当前的 access token
func (c *Client) getAccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken
}

// GetTokenExpiry 获取 token 过期时间
func (c *Client) GetTokenExpiry() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.expiresAt
}

// SendCommand 发送命令
func (c *Client) SendCommand(command string, args []string, exchange, account string) error {
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return fmt.Errorf("token 验证失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.SendCommand(ctx, &pb.CommandRequest{
		Command:     command,
		Args:        args,
		Exchange:    exchange,
		Account:     account,
		AccessToken: c.getAccessToken(),
	})
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("command failed: %s", resp.Message)
	}

	return nil
}

// GetNews 获取新闻
func (c *Client) GetNews(count int, source string) ([]news.NewsItem, error) {
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.GetNews(ctx, &pb.GetNewsRequest{
		Count:       int32(count),
		Source:      source,
		AccessToken: c.getAccessToken(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get news: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get news failed: %s", resp.Message)
	}

	// 转换为内部格式
	var newsList []news.NewsItem
	for _, item := range resp.News {
		newsList = append(newsList, news.NewsItem{
			ID:          item.Id,
			Title:       item.Title,
			Content:     item.Content,
			URL:         item.Url,
			Source:      item.Source,
			PublishedAt: time.Unix(item.PublishedAt, 0),
			Tags:        item.Tags,
			ImageURL:    item.ImageUrl,
		})
	}

	return newsList, nil
}

// GetExchanges 获取交易所列表
func (c *Client) GetExchanges() ([]*ShowExchangeItem, error) {
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := c.client.GetExchanges(ctx, &pb.GetExchangesRequest{
		AccessToken: c.getAccessToken(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get exchanges: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get exchanges failed: %s", resp.Message)
	}

	var exchanges []*ShowExchangeItem
	for _, item := range resp.Exchanges {
		exchanges = append(exchanges, &ShowExchangeItem{
			Name:        item.Name,
			APIUrl:      item.ApiUrl,
			ProxyUrl:    item.ProxyUrl,
			StatusValue: item.StatusValue,
		})
	}

	return exchanges, nil
}

// GetAccounts 获取账户列表
func (c *Client) GetAccounts() ([]*ShowAccountItem, error) {
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.GetAccounts(ctx, &pb.GetAccountsRequest{
		AccessToken: c.getAccessToken(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get accounts failed: %s", resp.Message)
	}

	// 转换为内部格式
	var accounts []*ShowAccountItem
	for _, item := range resp.Accounts {
		accounts = append(accounts, &ShowAccountItem{
			Id:               item.Id,
			Name:             item.Name,
			Exchange:         item.Exchange,
			TradeTypeValue:   item.TradeTypeValue,
			StatusValue:      item.StatusValue,
			IsolatedLeverage: item.IsolatedLeverage,
			CrossLeverage:    item.CrossLeverage,
			ProxyUrl:         item.ProxyUrl,
		})
	}

	return accounts, nil
}

// GetBalance 获取资产列表
func (c *Client) GetBalance(accountID int64) ([]*ShowAssetItem, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	if accountID <= 0 {
		return nil, fmt.Errorf("account_id 是必填参数，且必须大于 0")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.GetBalance(ctx, &pb.GetBalanceRequest{
		AccessToken: c.getAccessToken(),
		AccountId:   accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get balance failed: %s", resp.Message)
	}

	var assets []*ShowAssetItem
	for _, item := range resp.Assets {
		assets = append(assets, &ShowAssetItem{
			Currency:  item.Currency,
			Balance:   item.Balance,
			Available: item.Available,
			Frozen:    item.Frozen,
		})
	}

	return assets, nil
}

// GetPositions 获取仓位列表
func (c *Client) GetPositions(accountID int64) ([]*ShowPositionItem, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	if accountID <= 0 {
		return nil, fmt.Errorf("account_id 是必填参数，且必须大于 0")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.GetPositions(ctx, &pb.GetPositionsRequest{
		AccessToken: c.getAccessToken(),
		AccountId:   accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get positions failed: %s", resp.Message)
	}

	var positions []*ShowPositionItem
	for _, item := range resp.Positions {
		positions = append(positions, &ShowPositionItem{
			Symbol:     item.Symbol,
			PosSide:    item.PosSide,
			MarginType: item.MarginType,
			Size:       item.Size,
			AvgPrice:   item.AvgPrice,
			UnrealPnl:  item.UnrealPnl,
		})
	}

	return positions, nil
}

// GetSymbols 获取交易对列表
func (c *Client) GetSymbols(exchangeName, keyword string) ([]*ShowSymbolItem, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	if exchangeName == "" {
		return nil, fmt.Errorf("exchange 是必填参数")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.GetSymbols(ctx, &pb.GetSymbolsRequest{
		AccessToken: c.getAccessToken(),
		Exchange:    exchangeName,
		Keyword:     keyword,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get symbols failed: %s", resp.Message)
	}

	var symbols []*ShowSymbolItem
	for _, item := range resp.Symbols {
		symbols = append(symbols, &ShowSymbolItem{
			Exchange:    item.Exchange,
			Type:        item.Type,
			Name:        item.Name,
			Price:       item.Price,
			Volume:      item.Volume,
			High:        item.High,
			Low:         item.Low,
			Base:        item.Base,
			Quote:       item.Quote,
			MaxLeverage: item.MaxLeverage,
			MinSize:     item.MinSize,
			Contract:    item.Contract,
		})
	}

	return symbols, nil
}

// UseExchange 激活交易所
func (c *Client) UseExchange(exchangeName string) (*ShowExchangeItem, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	if exchangeName == "" {
		return nil, fmt.Errorf("exchange 是必填参数")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.UseExchange(ctx, &pb.UseExchangeRequest{
		AccessToken: c.getAccessToken(),
		Exchange:    exchangeName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to use exchange: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("use exchange failed: %s", resp.Message)
	}

	if resp.Exchange == nil {
		return nil, fmt.Errorf("use exchange response 缺少 exchange 信息")
	}

	return &ShowExchangeItem{
		Name:        resp.Exchange.Name,
		APIUrl:      resp.Exchange.ApiUrl,
		ProxyUrl:    resp.Exchange.ProxyUrl,
		StatusValue: resp.Exchange.StatusValue,
	}, nil
}

// UseAccount 激活账户
func (c *Client) UseAccount(accountName string) (*ShowAccountItem, *ShowExchangeItem, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, nil, fmt.Errorf("token 验证失败: %w", err)
	}

	if accountName == "" {
		return nil, nil, fmt.Errorf("account 是必填参数")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.UseAccount(ctx, &pb.UseAccountRequest{
		AccessToken: c.getAccessToken(),
		Account:     accountName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to use account: %w", err)
	}

	if !resp.Success {
		return nil, nil, fmt.Errorf("use account failed: %s", resp.Message)
	}

	if resp.Account == nil || resp.Exchange == nil {
		return nil, nil, fmt.Errorf("use account response 缺少账户或交易所信息")
	}

	accountItem := &ShowAccountItem{
		Id:               resp.Account.Id,
		Name:             resp.Account.Name,
		Exchange:         resp.Account.Exchange,
		TradeTypeValue:   resp.Account.TradeTypeValue,
		StatusValue:      resp.Account.StatusValue,
		IsolatedLeverage: resp.Account.IsolatedLeverage,
		CrossLeverage:    resp.Account.CrossLeverage,
		ProxyUrl:         resp.Account.ProxyUrl,
	}

	exchangeItem := &ShowExchangeItem{
		Name:        resp.Exchange.Name,
		APIUrl:      resp.Exchange.ApiUrl,
		ProxyUrl:    resp.Exchange.ProxyUrl,
		StatusValue: resp.Exchange.StatusValue,
	}

	return accountItem, exchangeItem, nil
}

// UpdateTradeConfig 更新账户杠杆配置
func (c *Client) UpdateTradeConfig(accountID int64, margin string, leverage int64) (*ShowAccountItem, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	if accountID <= 0 {
		return nil, fmt.Errorf("account_id 是必填参数")
	}
	if margin != "isolated" && margin != "cross" {
		return nil, fmt.Errorf("margin 只能为 isolated 或 cross")
	}
	if leverage <= 0 {
		return nil, fmt.Errorf("leverage 必须大于 0")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.UpdateTradeConfig(ctx, &pb.UpdateTradeConfigRequest{
		AccessToken: c.getAccessToken(),
		AccountId:   accountID,
		Margin:      margin,
		Leverage:    leverage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update trade config: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("update trade config failed: %s", resp.Message)
	}

	if resp.Account == nil {
		return nil, fmt.Errorf("update trade config response 缺少账户信息")
	}

	return &ShowAccountItem{
		Id:               resp.Account.Id,
		Name:             resp.Account.Name,
		Exchange:         resp.Account.Exchange,
		TradeTypeValue:   resp.Account.TradeTypeValue,
		StatusValue:      resp.Account.StatusValue,
		IsolatedLeverage: resp.Account.IsolatedLeverage,
		CrossLeverage:    resp.Account.CrossLeverage,
		ProxyUrl:         resp.Account.ProxyUrl,
	}, nil
}

// UpdateProxyConfig 更新账户代理配置
func (c *Client) UpdateProxyConfig(accountID int64, proxyURL string) (*ShowAccountItem, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	if accountID <= 0 {
		return nil, fmt.Errorf("account_id 是必填参数")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.UpdateProxyConfig(ctx, &pb.UpdateProxyConfigRequest{
		AccessToken: c.getAccessToken(),
		AccountId:   accountID,
		ProxyUrl:    proxyURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update proxy config: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("update proxy config failed: %s", resp.Message)
	}

	if resp.Account == nil {
		return nil, fmt.Errorf("update proxy config response 缺少账户信息")
	}

	return &ShowAccountItem{
		Id:               resp.Account.Id,
		Name:             resp.Account.Name,
		Exchange:         resp.Account.Exchange,
		TradeTypeValue:   resp.Account.TradeTypeValue,
		StatusValue:      resp.Account.StatusValue,
		IsolatedLeverage: resp.Account.IsolatedLeverage,
		CrossLeverage:    resp.Account.CrossLeverage,
		ProxyUrl:         resp.Account.ProxyUrl,
	}, nil
}

// UpdateSymbol 更新标的杠杆配置
func (c *Client) UpdateSymbol(accountID int64, exchangeName, symbol, margin string, leverage int64) error {
	if err := c.ensureValidToken(); err != nil {
		return fmt.Errorf("token 验证失败: %w", err)
	}

	if accountID <= 0 {
		return fmt.Errorf("account_id 是必填参数")
	}
	if exchangeName == "" || symbol == "" {
		return fmt.Errorf("exchange 与 symbol 均为必填参数")
	}
	if margin != "isolated" && margin != "cross" {
		return fmt.Errorf("margin 只能为 isolated 或 cross")
	}
	if leverage <= 0 {
		return fmt.Errorf("leverage 必须大于 0")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.UpdateSymbol(ctx, &pb.UpdateSymbolRequest{
		AccessToken: c.getAccessToken(),
		AccountId:   accountID,
		Exchange:    exchangeName,
		Symbol:      symbol,
		Margin:      margin,
		Leverage:    leverage,
	})
	if err != nil {
		return fmt.Errorf("failed to update symbol: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("update symbol failed: %s", resp.Message)
	}
	return nil
}

// UpdateAccount 更新账户信息
func (c *Client) UpdateAccount(exchangeName, targetAccount, tradeType, name, apiKey, secretKey, passphrase string) (*ShowAccountItem, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	if exchangeName == "" || targetAccount == "" {
		return nil, fmt.Errorf("exchange 与 target_account 均为必填参数")
	}
	if tradeType == "" || name == "" || apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("缺少必要的账户参数")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.UpdateAccount(ctx, &pb.UpdateAccountRequest{
		AccessToken:   c.getAccessToken(),
		Exchange:      exchangeName,
		TargetAccount: targetAccount,
		TradeType:     tradeType,
		Name:          name,
		ApiKey:        apiKey,
		SecretKey:     secretKey,
		Passphrase:    passphrase,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("update account failed: %s", resp.Message)
	}
	if resp.Account == nil {
		return nil, fmt.Errorf("update account response 缺少账户信息")
	}

	return &ShowAccountItem{
		Id:               resp.Account.Id,
		Name:             resp.Account.Name,
		Exchange:         resp.Account.Exchange,
		TradeTypeValue:   resp.Account.TradeTypeValue,
		StatusValue:      resp.Account.StatusValue,
		IsolatedLeverage: resp.Account.IsolatedLeverage,
		CrossLeverage:    resp.Account.CrossLeverage,
		ProxyUrl:         resp.Account.ProxyUrl,
	}, nil
}

// OpenOrder 提交开仓订单
func (c *Client) OpenOrder(accountID int64, exchangeName, symbol, posSide, margin, amount, amountType, strategy string) (string, string, error) {
	if err := c.ensureValidToken(); err != nil {
		return "", "", fmt.Errorf("token 验证失败: %w", err)
	}

	if accountID <= 0 {
		return "", "", fmt.Errorf("account_id 是必填参数")
	}
	if exchangeName == "" || symbol == "" {
		return "", "", fmt.Errorf("exchange 和 symbol 均为必填参数")
	}
	if posSide != "long" && posSide != "short" {
		return "", "", fmt.Errorf("pos_side 只能为 long 或 short")
	}
	if margin != "isolated" && margin != "cross" {
		return "", "", fmt.Errorf("margin 只能为 isolated 或 cross")
	}
	if amount == "" {
		return "", "", fmt.Errorf("amount 是必填参数")
	}

	side := "buy"
	if posSide == "short" {
		side = "sell"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.OpenOrder(ctx, &pb.OpenOrderRequest{
		AccessToken: c.getAccessToken(),
		AccountId:   accountID,
		Exchange:    exchangeName,
		Symbol:      symbol,
		PosSide:     posSide,
		Margin:      margin,
		Amount:      amount,
		AmountType:  amountType,
		Side:        side,
		OrderType:   "market",
		Strategy:    strategy,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to open order: %w", err)
	}
	if !resp.Success {
		return "", "", fmt.Errorf("open order failed: %s", resp.Message)
	}

	orderID := ""
	if resp.Order != nil {
		orderID = resp.Order.OrderId
	}

	return resp.Message, orderID, nil
}

// CloseOrder 提交平仓订单
func (c *Client) CloseOrder(accountID int64, exchangeName, symbol, posSide, margin, strategy string) (string, string, error) {
	if err := c.ensureValidToken(); err != nil {
		return "", "", fmt.Errorf("token 验证失败: %w", err)
	}

	if accountID <= 0 {
		return "", "", fmt.Errorf("account_id 是必填参数")
	}
	if exchangeName == "" || symbol == "" {
		return "", "", fmt.Errorf("exchange 和 symbol 均为必填参数")
	}
	if posSide != "long" && posSide != "short" {
		return "", "", fmt.Errorf("pos_side 只能为 long 或 short")
	}
	if margin != "isolated" && margin != "cross" {
		return "", "", fmt.Errorf("margin 只能为 isolated 或 cross")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.CloseOrder(ctx, &pb.CloseOrderRequest{
		AccessToken: c.getAccessToken(),
		AccountId:   accountID,
		Exchange:    exchangeName,
		Symbol:      symbol,
		PosSide:     posSide,
		Margin:      margin,
		Strategy:    strategy,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to close order: %w", err)
	}
	if !resp.Success {
		return "", "", fmt.Errorf("close order failed: %s", resp.Message)
	}

	orderID := ""
	if resp.Order != nil {
		orderID = resp.Order.OrderId
	}

	return resp.Message, orderID, nil
}

// CreateAccount 创建账户
func (c *Client) CreateAccount(exchangeName, tradeType, name, apiKey, secretKey, passphrase string) (*ShowAccountItem, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	if tradeType == "" || name == "" || apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("缺少必要的账户参数")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.CreateAccount(ctx, &pb.CreateAccountRequest{
		AccessToken: c.getAccessToken(),
		Exchange:    exchangeName,
		TradeType:   tradeType,
		Name:        name,
		ApiKey:      apiKey,
		SecretKey:   secretKey,
		Passphrase:  passphrase,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("create account failed: %s", resp.Message)
	}
	if resp.Account == nil {
		return nil, fmt.Errorf("create account response 缺少账户信息")
	}

	return &ShowAccountItem{
		Id:               resp.Account.Id,
		Name:             resp.Account.Name,
		Exchange:         resp.Account.Exchange,
		TradeTypeValue:   resp.Account.TradeTypeValue,
		StatusValue:      resp.Account.StatusValue,
		IsolatedLeverage: resp.Account.IsolatedLeverage,
		CrossLeverage:    resp.Account.CrossLeverage,
		ProxyUrl:         resp.Account.ProxyUrl,
	}, nil
}

// DeleteAccount 删除账户
func (c *Client) DeleteAccount(name string) error {
	if err := c.ensureValidToken(); err != nil {
		return fmt.Errorf("token 验证失败: %w", err)
	}

	if name == "" {
		return fmt.Errorf("name 是必填参数")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.DeleteAccount(ctx, &pb.DeleteAccountRequest{
		AccessToken: c.getAccessToken(),
		Name:        name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("delete account failed: %s", resp.Message)
	}

	return nil
}

// CancelOrder 取消策略订单
func (c *Client) CancelOrder(accountID int64, exchangeName, symbol, side, posSide, amount, amountType string) (string, string, error) {
	if err := c.ensureValidToken(); err != nil {
		return "", "", fmt.Errorf("token 验证失败: %w", err)
	}

	if accountID <= 0 {
		return "", "", fmt.Errorf("account_id 是必填参数")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.CancelOrder(ctx, &pb.CancelOrderRequest{
		AccessToken: c.getAccessToken(),
		AccountId:   accountID,
		Exchange:    exchangeName,
		Symbol:      symbol,
		Side:        side,
		PosSide:     posSide,
		Amount:      amount,
		AmountType:  amountType,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to cancel order: %w", err)
	}
	if !resp.Success {
		return "", "", fmt.Errorf("cancel order failed: %s", resp.Message)
	}

	orderID := ""
	if resp.Order != nil {
		orderID = resp.Order.OrderId
	}

	return resp.Message, orderID, nil
}

// GetOrders 获取订单列表
func (c *Client) GetOrders(accountID int64, status []string) ([]*ShowOrderItem, error) {
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	// 验证必填参数
	if accountID <= 0 {
		return nil, fmt.Errorf("account_id 是必填参数，且必须大于 0")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.GetOrders(ctx, &pb.GetOrdersRequest{
		AccessToken: c.getAccessToken(),
		AccountId:   accountID,
		Status:      status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get orders failed: %s", resp.Message)
	}

	// 转换为内部格式
	var orders []*ShowOrderItem
	for _, item := range resp.Orders {
		orders = append(orders, &ShowOrderItem{
			ID:         item.Id,
			Exchange:   item.Exchange,
			AccountID:  item.AccountId,
			Symbol:     item.Symbol,
			Side:       item.Side,
			PosSide:    item.PosSide,
			MarginType: item.MarginType,
			Price:      item.Price,
			Size:       item.Size,
			SizeType:   item.SizeType,
			OrderType:  item.OrderType,
			Strategy:   item.Strategy,
			OrderID:    item.OrderId,
			Type:       item.Type,
			Status:     item.Status,
			Msg:        item.Msg,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}

	log.Printf("成功获取 %d 个订单", len(orders))
	return orders, nil
}
