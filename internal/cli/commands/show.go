package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/lemconn/foxflow/internal/cli/command"
	cliRender "github.com/lemconn/foxflow/internal/cli/render"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/models"
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
	return "show <type> [options]\n  types: exchanges, users, assets, orders, positions, strategies, symbols, ss, news\n  news: show news [count] - 显示最新新闻，count 为可选参数，默认为 10"
}

func (c *ShowCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "exchanges":
		exchanges, err := repository.ListExchanges()
		if err != nil {
			return fmt.Errorf("failed to get exchanges: %w", err)
		}

		if exchanges == nil || len(exchanges) == 0 {
			fmt.Println(utils.RenderWarning("No exchanges found"))
			return nil
		}

		fmt.Println(cliRender.RenderExchangesWithStatus(exchanges))

	case "users":
		users, err := repository.ListUsers()
		if err != nil {
			return fmt.Errorf("failed to get users: %w", err)
		}
		fmt.Println(cliRender.RenderUsers(users))

	case "assets":
		if !ctx.IsReady() {
			return errors.New("请先选择交易所和用户")
		}

		exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeInstance().Name)
		if err != nil {
			return fmt.Errorf("failed to get exchange client: %w", err)
		}
		assets, err := exchangeClient.GetBalance(ctx.GetContext())
		if err != nil {
			return fmt.Errorf("failed to get assets: %w", err)
		}

		fmt.Println(cliRender.RenderAssets(assets))

	case "orders":
		if !ctx.IsReady() {
			return errors.New("请先选择交易所和用户")
		}

		exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeInstance().Name)
		if err != nil {
			return fmt.Errorf("failed to get exchange client: %w", err)
		}
		orders, err := exchangeClient.GetOrders(ctx.GetContext(), "", "pending")
		if err != nil {
			return fmt.Errorf("failed to get orders: %w", err)
		}
		fmt.Println(cliRender.RenderOrders(orders))

	case "positions":
		if !ctx.IsReady() {
			return errors.New("请先选择交易所和用户")
		}

		exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeInstance().Name)
		if err != nil {
			return fmt.Errorf("failed to get exchange client: %w", err)
		}
		positions, err := exchangeClient.GetPositions(ctx.GetContext())
		if err != nil {
			return fmt.Errorf("failed to get positions: %w", err)
		}
		fmt.Println(cliRender.RenderPositions(positions))

	case "strategies":
		fmt.Println(cliRender.RenderStrategies())

	case "symbols":
		if !ctx.IsReady() {
			return errors.New("请先选择交易所和用户")
		}

		symbolList, err := repository.GetSymbolByUser(ctx.GetUserInstance().ID)
		if err != nil {
			return fmt.Errorf("failed to get symbols: %w", err)
		}

		// 获取当前交易所指定用户的标的张面值换算数据
		symbols := make([]string, 0, len(symbolList))
		for _, symbol := range symbolList {
			symbols = append(symbols, symbol.Name)
		}

		exchangeName := ctx.GetExchangeInstance().Name
		contractList, err := repository.GetSymbolContractByExchangeSymbols(exchangeName, symbols)
		if err != nil {
			return fmt.Errorf("get contract err: %w", err)
		}

		contractListMap := make(map[string]*models.FoxContractMultiplier)
		for _, contract := range contractList {
			contractListMap[contract.Symbol] = contract
		}

		symbolInfoList := make([]cliRender.RenderSymbolsInfo, 0)
		for _, symbolInfo := range symbolList {
			renderSymbolsInfo := cliRender.RenderSymbolsInfo{
				Name:       symbolInfo.Name,
				Exchange:   symbolInfo.Exchange,
				Leverage:   symbolInfo.Leverage,
				MarginType: symbolInfo.MarginType,
			}
			if contractInfo, ok := contractListMap[symbolInfo.Name]; ok {
				renderSymbolsInfo.Multiplier = contractInfo.Multiplier
			}

			symbolInfoList = append(symbolInfoList, renderSymbolsInfo)
		}

		fmt.Println(cliRender.RenderSymbols(symbolInfoList))

	case "ss":
		if !ctx.IsReady() {
			return errors.New("请先选择交易所和用户")
		}

		ss, err := repository.ListSSOrders(ctx.GetUserInstance().ID, []string{"waiting", "pending"})
		if err != nil {
			return fmt.Errorf("failed to get strategy orders: %w", err)
		}
		fmt.Println(cliRender.RenderStrategyOrders(ss))

	case "news":
		return c.handleNewsCommand(ctx, args[1:])

	default:
		return fmt.Errorf("unknown show type: %s", args[0])
	}

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

	// 创建新闻管理器
	manager := news.NewManager()

	// 注册 BlockBeats 新闻源
	blockBeats := news.NewBlockBeats()
	manager.RegisterSource(blockBeats)

	// 创建带超时的上下文
	newsCtx, cancel := context.WithTimeout(ctx.GetContext(), 30*time.Second)
	defer cancel()

	// 获取新闻
	fmt.Println(utils.RenderInfo(fmt.Sprintf("正在获取最新 %d 条新闻...", count)))
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
