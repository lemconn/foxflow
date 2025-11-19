package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lemconn/foxflow/internal/cli/command"
	cliRender "github.com/lemconn/foxflow/internal/cli/render"
	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/grpc"
	"github.com/lemconn/foxflow/internal/news"
	"github.com/lemconn/foxflow/internal/repository"
	"github.com/lemconn/foxflow/internal/utils"
)

// ShowCommand 查看命令
type ShowCommand struct{}

func (c *ShowCommand) GetName() string {
	return "show"
}

func (c *ShowCommand) GetDescription() string {
	return "查看数据列表"
}

func (c *ShowCommand) GetUsage() string {
	return "show <type> [options]\n  types: exchange, account, balance, order, position, strategy, symbol, order, news\n  news: show news [count] - 显示最新新闻，count 为可选参数，默认为 10"
}

func (c *ShowCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "exchange":
		return c.handleExchangeCommand(ctx)
	case "account":
		return c.handleAccountCommand(ctx)
	case "balance":
		return c.handleBalanceCommand(ctx)
	case "order":
		return c.handleOrderCommand(ctx)
	case "position":
		return c.handlePositionCommand(ctx)
	case "strategy":
		fmt.Println(cliRender.RenderStrategies())
	case "symbol":
		return c.handleSymbolCommand(ctx, args[1:])
	case "news":
		return c.handleNewsCommand(ctx, args[1:])
	default:
		return fmt.Errorf("unknown show type: %s", args[0])
	}

	return nil
}

func (c *ShowCommand) handleSymbolCommand(ctx command.Context, args []string) error {
	if ctx.GetExchangeName() == "" {
		return errors.New("请先选择交易所")
	}

	exchangeInstance := ctx.GetExchangeInstance()
	if exchangeInstance == nil {
		return errors.New("交易所实例未正确初始化，请重新选择交易所")
	}

	keyword := ""
	if len(args) > 0 && args[0] != "" {
		keyword = strings.ToUpper(args[0])
	}

	if grpcClient := ctx.GetGRPCClient(); grpcClient != nil {
		fmt.Println(utils.RenderInfo("正在通过 gRPC 获取交易对列表..."))
		symbols, err := grpcClient.GetSymbols(exchangeInstance.Name, keyword)
		if err != nil {
			fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 获取交易对失败，回退到本地模式: %v", err)))
		} else {
			if len(symbols) == 0 {
				fmt.Println(utils.RenderWarning("暂无交易对数据"))
				return nil
			}

			renderSymbols := make([]cliRender.RenderSymbolsInfo, 0, len(symbols))
			for _, symbol := range symbols {
				renderSymbols = append(renderSymbols, cliRender.RenderSymbolsInfo{
					Exchange:    symbol.Exchange,
					Type:        symbol.Type,
					Name:        symbol.Name,
					Price:       symbol.Price,
					Volume:      symbol.Volume,
					High:        symbol.High,
					Low:         symbol.Low,
					Base:        symbol.Base,
					Quote:       symbol.Quote,
					MaxLeverage: symbol.MaxLeverage,
					MinSize:     symbol.MinSize,
					Contract:    symbol.Contract,
				})
			}

			fmt.Println(cliRender.RenderSymbols(renderSymbols))
			return nil
		}
	}

	return c.handleSymbolCommandLocal(ctx, args)
}

func (c *ShowCommand) handleSymbolCommandLocal(ctx command.Context, args []string) error {
	exchangeInstance := ctx.GetExchangeInstance()
	exchangeName := exchangeInstance.Name
	symbolList, exists := config.ExchangeSymbolList[exchangeName]
	if !exists {
		return fmt.Errorf("exchange %s not found", exchangeName)
	}

	exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeInstance().Name)
	if err != nil {
		return fmt.Errorf("failed to get exchange client: %w", err)
	}

	tickerList, err := exchangeClient.GetTickers(ctx.GetContext())
	if err != nil {
		return fmt.Errorf("failed to get ticker list: %w", err)
	}
	tickerListSymbolMap := make(map[string]exchange.Ticker)
	for _, ticker := range tickerList {
		tickerListSymbolMap[ticker.Symbol] = ticker
	}

	symbolInfoList := make([]cliRender.RenderSymbolsInfo, 0)
	for _, symbolInfo := range symbolList {

		// 如果存在参数，并且参数不为空，并且symbolInfo.Name字段中不包含参数，则跳过
		if len(args) > 0 && args[0] != "" && !strings.Contains(symbolInfo.Name, strings.ToUpper(args[0])) {
			continue
		}

		renderSymbolsInfo := cliRender.RenderSymbolsInfo{
			Exchange:    exchangeName,
			Type:        symbolInfo.Type,
			Name:        symbolInfo.Name,
			Base:        symbolInfo.Base,
			Quote:       symbolInfo.Quote,
			MaxLeverage: symbolInfo.MaxLever,
			MinSize:     symbolInfo.MinSize,
			Contract:    symbolInfo.Contract,
		}

		if ticker, ok := tickerListSymbolMap[symbolInfo.Name]; ok {
			renderSymbolsInfo.Price = ticker.Price
			renderSymbolsInfo.Volume = ticker.Volume
			renderSymbolsInfo.High = ticker.High
			renderSymbolsInfo.Low = ticker.Low
		}

		symbolInfoList = append(symbolInfoList, renderSymbolsInfo)
	}

	fmt.Println(cliRender.RenderSymbols(symbolInfoList))
	return nil
}

func (c *ShowCommand) handlePositionCommand(ctx command.Context) error {
	if !ctx.IsReady() {
		return errors.New("请先选择交易所和用户")
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return c.handlePositionCommandLocal(ctx)
	}

	fmt.Println(utils.RenderInfo("正在通过 gRPC 获取仓位列表..."))
	positions, err := grpcClient.GetPositions(ctx.GetAccountInstance().Id)
	if err != nil {
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 获取仓位失败，回退到本地模式: %v", err)))
		return c.handlePositionCommandLocal(ctx)
	}

	if len(positions) == 0 {
		fmt.Println(utils.RenderWarning("暂无仓位数据"))
		return nil
	}

	fmt.Println(cliRender.RenderPositions(positions))
	return nil
}

func (c *ShowCommand) handlePositionCommandLocal(ctx command.Context) error {
	exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeInstance().Name)
	if err != nil {
		return fmt.Errorf("failed to get exchange client: %w", err)
	}

	positions, err := exchangeClient.GetPositions(ctx.GetContext())
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	if len(positions) == 0 {
		fmt.Println(utils.RenderWarning("暂无仓位数据"))
		return nil
	}

	renderPositions := make([]*grpc.ShowPositionItem, 0, len(positions))
	for _, pos := range positions {
		renderPositions = append(renderPositions, &grpc.ShowPositionItem{
			Symbol:     pos.Symbol,
			PosSide:    pos.PosSide,
			MarginType: pos.MarginType,
			Size:       pos.Size,
			AvgPrice:   pos.AvgPrice,
			UnrealPnl:  pos.UnrealPnl,
		})
	}

	fmt.Println(cliRender.RenderPositions(renderPositions))
	return nil
}

func (c *ShowCommand) handleOrderCommand(ctx command.Context) error {
	if !ctx.IsReady() {
		return errors.New("请先选择交易所和用户")
	}

	// 检查是否有 gRPC 客户端
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		// 如果没有 gRPC 客户端，使用本地模式
		return c.handleOrderCommandLocal(ctx)
	}

	// 使用 gRPC 获取订单列表
	fmt.Println(utils.RenderInfo("正在通过 gRPC 获取订单列表..."))
	orders, err := grpcClient.GetOrders(ctx.GetAccountInstance().Id, []string{})
	if err != nil {
		// 如果 gRPC 失败，回退到本地模式
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 获取订单失败，回退到本地模式: %v", err)))
		return c.handleOrderCommandLocal(ctx)
	}

	if len(orders) == 0 {
		fmt.Println(utils.RenderWarning("暂无订单数据"))
		return nil
	}

	// 渲染订单列表
	fmt.Println(cliRender.RenderOrders(orders))

	return nil
}

// handleOrderCommandLocal 本地模式处理订单命令
func (c *ShowCommand) handleOrderCommandLocal(ctx command.Context) error {
	orders, err := repository.ListSSOrders(ctx.GetAccountInstance().Id, []string{})
	if err != nil {
		return fmt.Errorf("failed to get strategy orders: %w", err)
	}

	if len(orders) == 0 {
		fmt.Println(utils.RenderWarning("暂无订单数据"))
		return nil
	}

	fmt.Println(cliRender.RenderStrategyOrders(orders))
	return nil
}

func (c *ShowCommand) handleBalanceCommand(ctx command.Context) error {
	if !ctx.IsReady() {
		return errors.New("请先选择交易所和用户")
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return c.handleBalanceCommandLocal(ctx)
	}

	fmt.Println(utils.RenderInfo("正在通过 gRPC 获取资产列表..."))
	assets, err := grpcClient.GetBalance(ctx.GetAccountInstance().Id)
	if err != nil {
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 获取资产失败，回退到本地模式: %v", err)))
		return c.handleBalanceCommandLocal(ctx)
	}

	if len(assets) == 0 {
		fmt.Println(utils.RenderWarning("暂无资产数据"))
		return nil
	}

	fmt.Println(cliRender.RenderAssets(assets))
	return nil
}

func (c *ShowCommand) handleBalanceCommandLocal(ctx command.Context) error {
	exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeInstance().Name)
	if err != nil {
		return fmt.Errorf("failed to get exchange client: %w", err)
	}

	assets, err := exchangeClient.GetBalance(ctx.GetContext())
	if err != nil {
		return fmt.Errorf("failed to get assets: %w", err)
	}

	if len(assets) == 0 {
		fmt.Println(utils.RenderWarning("暂无资产数据"))
		return nil
	}

	renderAssets := make([]*grpc.ShowAssetItem, 0, len(assets))
	for _, asset := range assets {
		renderAssets = append(renderAssets, &grpc.ShowAssetItem{
			Currency:  asset.Currency,
			Balance:   asset.Balance,
			Available: asset.Available,
			Frozen:    asset.Frozen,
		})
	}

	fmt.Println(cliRender.RenderAssets(renderAssets))
	return nil
}

func (c *ShowCommand) handleAccountCommand(ctx command.Context) error {
	// 检查是否有 gRPC 客户端
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		// 如果没有 gRPC 客户端，使用本地模式
		return c.handleAccountCommandLocal(ctx)
	}

	// 使用 gRPC 获取账户列表
	fmt.Println(utils.RenderInfo("正在通过 gRPC 获取账户列表..."))
	accounts, err := grpcClient.GetAccounts()
	if err != nil {
		// 如果 gRPC 失败，回退到本地模式
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 获取账户列表失败，回退到本地模式: %v", err)))
		return c.handleAccountCommandLocal(ctx)
	}

	if len(accounts) == 0 {
		fmt.Println(utils.RenderWarning("暂无账户数据"))
		return nil
	}

	// 渲染账户列表
	fmt.Println(cliRender.RenderAccounts(accounts))

	return nil
}

// handleAccountCommandLocal 本地模式处理账户命令
func (c *ShowCommand) handleAccountCommandLocal(ctx command.Context) error {
	accounts, err := repository.ExchangeAccountList(ctx.GetExchangeName())
	if err != nil {
		return fmt.Errorf("failed to get accounts: %w", err)
	}

	if len(accounts) == 0 {
		fmt.Println(utils.RenderWarning("暂无账户数据"))
		return nil
	}

	renderAccounts := make([]*grpc.ShowAccountItem, 0)
	for _, account := range accounts {
		renderAccount := &grpc.ShowAccountItem{
			Name:           account.Name,
			Exchange:       account.Exchange,
			TradeTypeValue: account.TradeType,
			StatusValue:    account.IsActive,
		}

		if len(account.TradeConfigs) > 0 {
			var crossLeverage, isolatedLeverage int64
			for _, tradeConfig := range account.TradeConfigs {
				if tradeConfig.Margin == "cross" {
					crossLeverage = tradeConfig.Leverage
				} else if tradeConfig.Margin == "isolated" {
					isolatedLeverage = tradeConfig.Leverage
				}
			}
			renderAccount.CrossLeverage = crossLeverage
			renderAccount.IsolatedLeverage = isolatedLeverage
		}

		// 处理代理地址
		if account.Config.ProxyURL != "" {
			renderAccount.ProxyUrl = account.Config.ProxyURL
		}

		renderAccounts = append(renderAccounts, renderAccount)
	}

	fmt.Println(cliRender.RenderAccounts(renderAccounts))
	return nil
}

func (c *ShowCommand) handleExchangeCommand(ctx command.Context) error {
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return c.handleExchangeCommandLocal()
	}

	fmt.Println(utils.RenderInfo("正在通过 gRPC 获取交易所列表..."))
	exchanges, err := grpcClient.GetExchanges()
	if err != nil {
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 获取交易所失败，回退到本地模式: %v", err)))
		return c.handleExchangeCommandLocal()
	}

	if len(exchanges) == 0 {
		fmt.Println(utils.RenderWarning("暂无交易所数据"))
		return nil
	}

	fmt.Println(cliRender.RenderExchangesWithStatus(exchanges))
	return nil
}

func (c *ShowCommand) handleExchangeCommandLocal() error {
	exchanges, err := repository.ListExchanges()
	if err != nil {
		return fmt.Errorf("failed to get exchanges: %w", err)
	}

	if len(exchanges) == 0 {
		fmt.Println(utils.RenderWarning("暂无交易所数据"))
		return nil
	}

	renderItems := make([]*grpc.ShowExchangeItem, 0, len(exchanges))
	for _, exchange := range exchanges {
		renderItems = append(renderItems, &grpc.ShowExchangeItem{
			Name:        exchange.Name,
			APIUrl:      exchange.APIURL,
			ProxyUrl:    exchange.ProxyURL,
			StatusValue: exchange.IsActive,
		})
	}

	fmt.Println(cliRender.RenderExchangesWithStatus(renderItems))
	return nil
}

// handleNewsCommand 处理新闻命令
func (c *ShowCommand) handleNewsCommand(ctx command.Context, args []string) error {
	// 默认获取 10 条新闻
	count := 10

	// 如果提供了数量参数，解析它
	if len(args) > 0 {
		if parsedCount, err := strconv.Atoi(args[0]); err == nil && parsedCount > 0 {
			count = parsedCount
		} else {
			return fmt.Errorf("无效的新闻数量参数: %s，请输入正整数", args[0])
		}
	}

	// 检查是否有 gRPC 客户端
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		// 如果没有 gRPC 客户端，使用本地模式
		return c.handleNewsCommandLocal(ctx, count)
	}

	// 使用 gRPC 获取新闻
	fmt.Println(utils.RenderInfo(fmt.Sprintf("正在通过 gRPC 获取最新 %d 条新闻...", count)))
	newsList, err := grpcClient.GetNews(count, "blockbeats")
	if err != nil {
		// 如果 gRPC 失败，回退到本地模式
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 获取新闻失败，回退到本地模式: %v", err)))
		return c.handleNewsCommandLocal(ctx, count)
	}

	if len(newsList) == 0 {
		fmt.Println(utils.RenderWarning("暂无新闻数据"))
		return nil
	}

	// 渲染新闻
	fmt.Println(cliRender.RenderNews(newsList))

	return nil
}

// handleNewsCommandLocal 本地模式处理新闻命令
func (c *ShowCommand) handleNewsCommandLocal(ctx command.Context, count int) error {
	// 创建新闻管理器
	manager := news.NewManager()

	// 注册 BlockBeats 新闻源
	blockBeats := news.NewBlockBeats()
	manager.RegisterSource(blockBeats)

	// 创建带超时的上下文
	newsCtx, cancel := context.WithTimeout(ctx.GetContext(), 30*time.Second)
	defer cancel()

	// 获取新闻
	fmt.Println(utils.RenderInfo(fmt.Sprintf("正在本地获取最新 %d 条新闻...", count)))
	newsList, err := manager.GetNewsFromSource(newsCtx, "blockbeats", count)
	if err != nil {
		return fmt.Errorf("获取新闻失败: %w", err)
	}

	if len(newsList) == 0 {
		fmt.Println(utils.RenderWarning("暂无新闻数据"))
		return nil
	}

	// 渲染新闻
	fmt.Println(cliRender.RenderNews(newsList))

	return nil
}
