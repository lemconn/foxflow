package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	cliRender "github.com/lemconn/foxflow/internal/cli/render"
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
		return fmt.Errorf("请先选择交易所")
	}

	exchangeInstance := ctx.GetExchangeInstance()
	if exchangeInstance == nil {
		return fmt.Errorf("交易所实例未正确初始化，请重新选择交易所")
	}

	keyword := ""
	if len(args) > 0 && args[0] != "" {
		keyword = strings.ToUpper(args[0])
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	symbols, err := grpcClient.GetSymbols(exchangeInstance.Name, keyword)
	if err != nil {
		return fmt.Errorf("获取交易对失败: %v", err)
	}

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

func (c *ShowCommand) handlePositionCommand(ctx command.Context) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	positions, err := grpcClient.GetPositions(ctx.GetAccountInstance().Id)
	if err != nil {
		return fmt.Errorf("获取仓位失败: %w", err)
	}

	if len(positions) == 0 {
		fmt.Println(utils.RenderWarning("暂无仓位数据"))
		return nil
	}

	fmt.Println(cliRender.RenderPositions(positions))
	return nil
}

func (c *ShowCommand) handleOrderCommand(ctx command.Context) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	// 检查是否有 gRPC 客户端
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	orders, err := grpcClient.GetOrders(ctx.GetAccountInstance().Id, []string{})
	if err != nil {
		return fmt.Errorf("获取订单失败: %w", err)
	}

	if len(orders) == 0 {
		fmt.Println(utils.RenderWarning("暂无订单数据"))
		return nil
	}

	// 渲染订单列表
	fmt.Println(cliRender.RenderOrders(orders))

	return nil
}

func (c *ShowCommand) handleBalanceCommand(ctx command.Context) error {
	if !ctx.IsReady() {
		return errors.New("请先选择交易所和用户")
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	assets, err := grpcClient.GetBalance(ctx.GetAccountInstance().Id)
	if err != nil {
		return fmt.Errorf("获取资产失败: %w", err)
	}

	if len(assets) == 0 {
		fmt.Println(utils.RenderWarning("暂无资产数据"))
		return nil
	}

	fmt.Println(cliRender.RenderAssets(assets))
	return nil
}

func (c *ShowCommand) handleAccountCommand(ctx command.Context) error {
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	accounts, err := grpcClient.GetAccounts()
	if err != nil {
		return fmt.Errorf("获取账户列表失败: %w", err)
	}

	if len(accounts) == 0 {
		fmt.Println(utils.RenderWarning("暂无账户数据"))
		return nil
	}

	fmt.Println(cliRender.RenderAccounts(accounts))
	return nil
}

func (c *ShowCommand) handleExchangeCommand(ctx command.Context) error {
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	exchanges, err := grpcClient.GetExchanges()
	if err != nil {
		return fmt.Errorf("获取交易所失败: %w", err)
	}

	if len(exchanges) == 0 {
		fmt.Println(utils.RenderWarning("暂无交易所数据"))
		return nil
	}

	fmt.Println(cliRender.RenderExchangesWithStatus(exchanges))
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

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	newsList, err := grpcClient.GetNews(count, "blockbeats")
	if err != nil {
		return fmt.Errorf("获取新闻失败: %v", err)
	}

	if len(newsList) == 0 {
		fmt.Println(utils.RenderWarning("暂无新闻数据"))
		return nil
	}

	// 渲染新闻
	fmt.Println(cliRender.RenderNews(newsList))

	return nil
}
