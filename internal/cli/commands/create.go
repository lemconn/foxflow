package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/exchange"

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
	case "account":
		return c.createAccount(ctx, args[1:])
	case "ss":
		return c.createStrategyOrder(ctx, args[1:])
	default:
		return fmt.Errorf("unknown create type: %s", args[0])
	}
}

func (c *CreateCommand) createAccount(ctx command.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: create account <trade_type> name=<name> apiKey=<key> secretKey=<secret> [passphrase=<passphrase>]")
	}

	account := &models.FoxAccount{}

	account.TradeType = args[0]
	for _, arg := range args {
		if strings.HasPrefix(arg, "name=") {
			account.Name = strings.TrimPrefix(arg, "name=")
		} else if strings.HasPrefix(arg, "apiKey=") {
			account.AccessKey = strings.TrimPrefix(arg, "apiKey=")
		} else if strings.HasPrefix(arg, "secretKey=") {
			account.SecretKey = strings.TrimPrefix(arg, "secretKey=")
		} else if strings.HasPrefix(arg, "passphrase=") {
			account.Passphrase = strings.TrimPrefix(arg, "passphrase=")
		}
	}

	if account.TradeType == "" || account.Name == "" || account.AccessKey == "" || account.SecretKey == "" {
		return fmt.Errorf("missing required parameters")
	}

	// 根据用户名获取用户信息
	accountInfo, err := repository.FindAccountByName(account.Name)
	if err != nil {
		return fmt.Errorf("find username err: %w", err)
	}

	// 用户存在则不允许创建
	if accountInfo != nil && accountInfo.ID > 0 {
		return fmt.Errorf("account already exists, name: %s", accountInfo.Name)
	}

	account.Exchange = ctx.GetExchangeName()
	if account.Exchange == "" {
		account.Exchange = config.DefaultExchange // 默认交易所
	}

	exchangeInfo, err := repository.GetExchange(account.Exchange)
	if err != nil {
		return fmt.Errorf("get exchange error: %w", err)
	}

	if exchangeInfo.Name == "" {
		return fmt.Errorf("exchange is not found")
	}

	if exchangeInfo.Name == config.DefaultExchange && account.Passphrase == "" {
		return fmt.Errorf("okx exchange passphrase is required")
	}

	// 到指定交易交易所验证当前用户
	exchangeClient, err := exchange.GetManager().GetExchange(exchangeInfo.Name)
	if err != nil {
		return fmt.Errorf("get exchange client error: %w", err)
	}
	err = exchangeClient.Connect(ctx.GetContext(), account)
	if err != nil {
		return fmt.Errorf("connect exchange error: %w", err)
	}

	if err = repository.CreateAccount(account); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	useCommand := UseCommand{}
	err = useCommand.HandleAccountCommand(ctx, account.Name)
	if err != nil {
		return fmt.Errorf("create account & use error: %w", err)
	}

	return nil
}

func (c *CreateCommand) createStrategyOrder(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	// 解析参数
	order := &models.FoxSS{
		AccountID: ctx.GetAccountInstance().ID,
		OrderType: "limit",
		Type:      "open",
		Status:    "waiting",
	}

	var strategy string

	for _, arg := range args {
		if strings.HasPrefix(arg, "--symbol=") {
			order.Symbol = strings.ToUpper(strings.TrimPrefix(arg, "--symbol="))
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

	// 提交到当前激活交易所
	exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeName())
	if err != nil {
		return fmt.Errorf("get exchange client error: %w", err)
	}

	//// 获取当前用户要构面的交易对信息是否存在
	//localSymbol, err := repository.GetSymbolByNameUser(order.Symbol, ctx.GetUserInstance().ID)
	//if err != nil {
	//	return fmt.Errorf("get local symbol err: %w", err)
	//}
	//
	//if localSymbol == nil || localSymbol.ID == 0 {
	//	return fmt.Errorf("symbol not exists")
	//}

	// 如果没有策略，直接提交订单；否则仅写库
	if strategy == "" {

		// 构造交易所订单
		exOrder := &exchange.Order{
			Symbol:  order.Symbol,
			Side:    order.Side,
			PosSide: order.PosSide,
			Price:   order.Px,
			Size:    order.Sz,
			Type:    order.OrderType,
			//MarginType: localSymbol.MarginType,
		}

		createdOrder, err := exchangeClient.CreateOrder(ctx.GetContext(), exOrder)
		if err != nil {
			return fmt.Errorf("create exchange order error: %w", err)
		}

		// 将交易所返回的订单ID回写
		order.OrderID = createdOrder.ID
		order.Status = "pending"
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("交易所下单成功，orderId=%s", order.OrderID)))
	} else {
		order.Strategy = strategy
	}

	if err := repository.CreateSSOrder(order); err != nil {
		return fmt.Errorf("failed to create strategy order: %w", err)
	}

	fmt.Println(utils.RenderInfo("策略订单已创建，等待策略条件满足"))
	return nil
}
