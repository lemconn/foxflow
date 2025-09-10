package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/config"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/models"
	"github.com/lemconn/foxflow/internal/repository"
	"github.com/lemconn/foxflow/internal/utils"
)

// CreateCommand 创建命令
type CreateCommand struct{}

func (c *CreateCommand) GetName() string        { return "create" }
func (c *CreateCommand) GetDescription() string { return "创建用户、标的或策略订单" }
func (c *CreateCommand) GetUsage() string       { return "create <type> [options]" }

func (c *CreateCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "users":
		return c.createUser(ctx, args[1:])
	case "symbols":
		return c.createSymbol(ctx, args[1:])
	case "ss":
		return c.createStrategyOrder(ctx, args[1:])
	default:
		return fmt.Errorf("unknown create type: %s", args[0])
	}
}

func (c *CreateCommand) createUser(ctx command.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: create users --username=<name> --ak=<key> --sk=<secret> --trade_type=<type>")
	}

	user := &models.FoxUser{}

	for _, arg := range args {
		if strings.HasPrefix(arg, "--username=") {
			user.Username = strings.TrimPrefix(arg, "--username=")
		} else if strings.HasPrefix(arg, "--ak=") {
			user.AccessKey = strings.TrimPrefix(arg, "--ak=")
		} else if strings.HasPrefix(arg, "--sk=") {
			user.SecretKey = strings.TrimPrefix(arg, "--sk=")
		} else if strings.HasPrefix(arg, "--trade_type=") {
			user.TradeType = strings.TrimPrefix(arg, "--trade_type=")
		}
	}

	if user.Username == "" || user.AccessKey == "" || user.SecretKey == "" || user.TradeType == "" {
		return fmt.Errorf("missing required parameters")
	}

	user.Exchange = ctx.GetExchangeName()
	if user.Exchange == "" {
		user.Exchange = config.DefaultExchange // 默认交易所
	}

	exchangeInfo, err := repository.GetExchange("")
	if err != nil {
		return fmt.Errorf("get exchange error: %w", err)
	}

	if exchangeInfo.Name == "" {

	}

	// 到指定交易交易所验证当前用户

	if err := repository.CreateUser(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("用户创建成功: %s", user.Username)))
	return nil
}

func (c *CreateCommand) createSymbol(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: create symbols <symbol> [--leverage=<num>] [--margin-type=<type>]")
	}

	symbol := &models.FoxSymbol{
		Name:       args[0],
		UserID:     ctx.GetUserInstance().ID,
		Exchange:   ctx.GetExchangeName(),
		Leverage:   1,
		MarginType: "isolated",
	}

	// 解析可选参数
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "--leverage=") {
			leverage, err := strconv.Atoi(strings.TrimPrefix(arg, "--leverage="))
			if err != nil {
				return fmt.Errorf("invalid leverage value")
			}
			symbol.Leverage = leverage
		} else if strings.HasPrefix(arg, "--margin-type=") {
			symbol.MarginType = strings.TrimPrefix(arg, "--margin-type=")
		}
	}

	if err := repository.CreateSymbol(symbol); err != nil {
		return fmt.Errorf("failed to create symbol: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("标的创建成功: %s", symbol.Name)))
	return nil
}

func (c *CreateCommand) createStrategyOrder(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	// 解析参数
	order := &models.FoxSS{
		UserID:    ctx.GetUserInstance().ID,
		OrderType: "limit",
		Type:      "open",
		Status:    "waiting",
	}

	var strategy string

	for _, arg := range args {
		if strings.HasPrefix(arg, "--symbol=") {
			order.Symbol = strings.TrimPrefix(arg, "--symbol=")
		} else if strings.HasPrefix(arg, "--side=") {
			order.Side = strings.TrimPrefix(arg, "--side=")
		} else if strings.HasPrefix(arg, "--posSide=") {
			order.PosSide = strings.TrimPrefix(arg, "--posSide=")
		} else if strings.HasPrefix(arg, "--px=") {
			px, err := strconv.ParseFloat(strings.TrimPrefix(arg, "--px="), 64)
			if err != nil {
				return fmt.Errorf("invalid price value")
			}
			order.Px = px
		} else if strings.HasPrefix(arg, "--sz=") {
			sz, err := strconv.ParseFloat(strings.TrimPrefix(arg, "--sz="), 64)
			if err != nil {
				return fmt.Errorf("invalid size value")
			}
			order.Sz = sz
		} else if arg == "--limit" {
			order.OrderType = "limit"
		} else if arg == "--market" {
			order.OrderType = "market"
		} else if strings.HasPrefix(arg, "--strategy=") {
			strategy = strings.TrimPrefix(arg, "--strategy=")
		}
	}

	if order.Symbol == "" || order.Side == "" || order.Sz == 0 {
		return fmt.Errorf("missing required parameters: symbol, side, size")
	}

	// 如果没有策略，直接提交订单
	if strategy == "" {
		// 这里应该直接提交到交易所
		order.Status = "pending"
		fmt.Println(utils.RenderInfo("订单将直接提交到交易所"))
	} else {
		order.Strategy = strategy
		fmt.Println(utils.RenderInfo("策略订单已创建，等待策略条件满足"))
	}

	if err := repository.CreateSSOrder(order); err != nil {
		return fmt.Errorf("failed to create strategy order: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("策略订单创建成功: ID=%d", order.ID)))
	return nil
}
