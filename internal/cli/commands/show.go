package commands

import (
	"errors"
	"fmt"

	"github.com/lemconn/foxflow/internal/cli/command"
	cliRender "github.com/lemconn/foxflow/internal/cli/render"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/models"
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
	return "show <type> [options]"
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

		exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeInstance().Name)
		if err != nil {
			return fmt.Errorf("failed to get exchange client: %w", err)
		}
		_, err = exchangeClient.GetSymbols(ctx.GetContext(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get symbols: %w", err)
		}
		fmt.Println(cliRender.RenderSymbols([]string{}))

	case "ss":
		var ss []models.FoxSS
		var uid *uint
		if ctx.IsReady() {
			u := ctx.GetUserInstance().ID
			uid = &u
		}
		ss, err := repository.ListWaitingSSOrders(uid)
		if err != nil {
			return fmt.Errorf("failed to get strategy orders: %w", err)
		}
		fmt.Println(cliRender.RenderStrategyOrders(ss))

	default:
		return fmt.Errorf("unknown show type: %s", args[0])
	}

	return nil
}
