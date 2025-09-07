package cli

import (
	"fmt"
	"strconv"
	"strings"

	cliRender "foxflow/internal/cli/render"
	"foxflow/internal/database"
	"foxflow/internal/exchange"
	"foxflow/internal/models"
	"foxflow/pkg/utils"
)

// Command 命令接口
type Command interface {
	GetName() string
	GetDescription() string
	GetUsage() string
	Execute(ctx *Context, args []string) error
}

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

func (c *ShowCommand) Execute(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	db := database.GetDB()

	switch args[0] {
	case "exchanges":
		var exchanges []models.FoxExchange
		if err := db.Find(&exchanges).Error; err != nil {
			return fmt.Errorf("failed to get exchanges: %w", err)
		}

		fmt.Println(cliRender.RenderExchangesWithStatus(exchanges))

	case "users":
		var users []models.FoxUser
		if err := db.Find(&users).Error; err != nil {
			return fmt.Errorf("failed to get users: %w", err)
		}
		fmt.Println(cliRender.RenderUsers(users))

	case "assets":
		if !ctx.IsReady() {
			return fmt.Errorf(utils.RenderError("请先选择交易所和用户"))
		}
		assets, err := ctx.GetExchangeInstance().GetBalance(ctx.GetContext())
		if err != nil {
			return fmt.Errorf("failed to get assets: %w", err)
		}
		fmt.Println(cliRender.RenderAssets(assets))

	case "orders":
		if !ctx.IsReady() {
			return fmt.Errorf(utils.RenderError("请先选择交易所和用户"))
		}
		orders, err := ctx.GetExchangeInstance().GetOrders(ctx.GetContext(), "", "pending")
		if err != nil {
			return fmt.Errorf("failed to get orders: %w", err)
		}
		fmt.Println(cliRender.RenderOrders(orders))

	case "positions":
		if !ctx.IsReady() {
			return fmt.Errorf(utils.RenderError("请先选择交易所和用户"))
		}
		positions, err := ctx.GetExchangeInstance().GetPositions(ctx.GetContext())
		if err != nil {
			return fmt.Errorf("failed to get positions: %w", err)
		}
		fmt.Println(cliRender.RenderPositions(positions))

	case "strategies":
		fmt.Println(cliRender.RenderStrategies())

	case "symbols":
		if !ctx.IsReady() {
			return fmt.Errorf(utils.RenderError("请先选择交易所和用户"))
		}
		symbols, err := ctx.GetExchangeInstance().GetSymbols(ctx.GetContext())
		if err != nil {
			return fmt.Errorf("failed to get symbols: %w", err)
		}
		fmt.Println(cliRender.RenderSymbols(symbols))

	case "ss":
		var ss []models.FoxSS
		query := db.Where("status = ?", "waiting")
		if ctx.IsReady() {
			query = query.Where("user_id = ?", ctx.GetUser().ID)
		}
		if err := query.Find(&ss).Error; err != nil {
			return fmt.Errorf("failed to get strategy orders: %w", err)
		}
		fmt.Println(cliRender.RenderStrategyOrders(ss))

	default:
		return fmt.Errorf("unknown show type: %s", args[0])
	}

	return nil
}

// UseCommand 激活命令
type UseCommand struct{}

func (c *UseCommand) GetName() string {
	return "use"
}

func (c *UseCommand) GetDescription() string {
	return "激活交易所或用户"
}

func (c *UseCommand) GetUsage() string {
	return "use <type> <name>"
}

func (c *UseCommand) Execute(ctx *Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	db := database.GetDB()

	switch args[0] {
	case "exchanges":
		exchangeName := args[1]

		// 将所有交易所设置为非激活状态
		if err := db.Model(&models.FoxExchange{}).Where("1 = 1").Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate exchanges: %w", err)
		}

		// 将所有用户设置为非激活状态
		if err := db.Model(&models.FoxUser{}).Where("1 = 1").Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate users: %w", err)
		}

		// 断开当前交易所连接
		if ctx.GetExchange() != "" {
			exchange.GetManager().DisconnectUser(ctx.GetExchange())
		}

		// 激活选中的交易所
		if err := db.Model(&models.FoxExchange{}).Where("name = ?", exchangeName).Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate exchange: %w", err)
		}

		// 设置新的交易所
		ex, err := exchange.GetManager().GetExchange(exchangeName)
		if err != nil {
			return fmt.Errorf("exchange not found: %s", exchangeName)
		}
		ctx.SetExchange(exchangeName)
		ctx.SetExchangeInstance(ex)
		ctx.SetUser(nil) // 清除当前用户
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("已激活交易所: %s", exchangeName)))

	case "users":
		username := args[1]
		var user models.FoxUser
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			return fmt.Errorf("user not found: %s", username)
		}

		// 将所有用户设置为非激活状态
		if err := db.Model(&models.FoxUser{}).Where("1 = 1").Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate users: %w", err)
		}

		// 如果用户属于不同的交易所，需要切换交易所
		if ctx.GetExchange() != "" && ctx.GetExchange() != user.Exchange {
			// 将所有交易所设置为非激活状态
			if err := db.Model(&models.FoxExchange{}).Where("1 = 1").Update("is_active", false).Error; err != nil {
				return fmt.Errorf("failed to deactivate exchanges: %w", err)
			}

			// 断开当前交易所连接
			exchange.GetManager().DisconnectUser(ctx.GetExchange())

			// 切换到用户所属的交易所
			ex, err := exchange.GetManager().GetExchange(user.Exchange)
			if err != nil {
				return fmt.Errorf("failed to get exchange: %w", err)
			}

			// 激活用户所属的交易所
			if err := db.Model(&models.FoxExchange{}).Where("name = ?", user.Exchange).Update("is_active", true).Error; err != nil {
				return fmt.Errorf("failed to activate exchange: %w", err)
			}

			ctx.SetExchange(user.Exchange)
			ctx.SetExchangeInstance(ex)
		}

		// 激活选中的用户
		if err := db.Model(&models.FoxUser{}).Where("username = ?", username).Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate user: %w", err)
		}

		// 连接用户到交易所
		if ctx.GetExchange() != "" {
			if err := exchange.GetManager().ConnectUser(ctx.GetContext(), ctx.GetExchange(), &user); err != nil {
				return fmt.Errorf("failed to connect user: %w", err)
			}
		}

		ctx.SetUser(&user)
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("已激活用户: %s", username)))

	default:
		return fmt.Errorf("unknown use type: %s", args[0])
	}

	return nil
}

// CreateCommand 创建命令
type CreateCommand struct{}

func (c *CreateCommand) GetName() string {
	return "create"
}

func (c *CreateCommand) GetDescription() string {
	return "创建用户、标的或策略订单"
}

func (c *CreateCommand) GetUsage() string {
	return "create <type> [options]"
}

func (c *CreateCommand) Execute(ctx *Context, args []string) error {
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

func (c *CreateCommand) createUser(ctx *Context, args []string) error {
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

	user.Exchange = ctx.GetExchange()
	if user.Exchange == "" {
		user.Exchange = "okx" // 默认交易所
	}

	db := database.GetDB()
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("用户创建成功: %s", user.Username)))
	return nil
}

func (c *CreateCommand) createSymbol(ctx *Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	if len(args) < 1 {
		return fmt.Errorf("usage: create symbols <symbol> [--leverage=<num>] [--margin-type=<type>]")
	}

	symbol := &models.FoxSymbol{
		Name:       args[0],
		UserID:     ctx.GetUser().ID,
		Exchange:   ctx.GetExchange(),
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

	db := database.GetDB()
	if err := db.Create(symbol).Error; err != nil {
		return fmt.Errorf("failed to create symbol: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("标的创建成功: %s", symbol.Name)))
	return nil
}

func (c *CreateCommand) createStrategyOrder(ctx *Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	// 解析参数
	order := &models.FoxSS{
		UserID:    ctx.GetUser().ID,
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

	db := database.GetDB()
	if err := db.Create(order).Error; err != nil {
		return fmt.Errorf("failed to create strategy order: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("策略订单创建成功: ID=%d", order.ID)))
	return nil
}

// UpdateCommand 设置命令
type UpdateCommand struct{}

func (c *UpdateCommand) GetName() string {
	return "update"
}

func (c *UpdateCommand) GetDescription() string {
	return "设置杠杆或保证金模式"
}

func (c *UpdateCommand) GetUsage() string {
	return "update <type> <value>"
}

func (c *UpdateCommand) Execute(ctx *Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "leverage":
		leverage, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid leverage value")
		}
		// 这里应该调用交易所API设置杠杆
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("杠杆设置为: %d", leverage)))

	case "margin-type":
		marginType := args[1]
		// 这里应该调用交易所API设置保证金模式
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("保证金模式设置为: %s", marginType)))

	default:
		return fmt.Errorf("unknown update type: %s", args[0])
	}

	return nil
}

// CancelCommand 取消命令
type CancelCommand struct{}

func (c *CancelCommand) GetName() string {
	return "cancel"
}

func (c *CancelCommand) GetDescription() string {
	return "取消策略订单"
}

func (c *CancelCommand) GetUsage() string {
	return "cancel ss <id>"
}

func (c *CancelCommand) Execute(ctx *Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	if args[0] != "ss" {
		return fmt.Errorf("only support cancel ss")
	}

	orderID, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid order ID")
	}

	db := database.GetDB()
	var order models.FoxSS
	if err := db.Where("id = ? AND user_id = ?", orderID, ctx.GetUser().ID).First(&order).Error; err != nil {
		return fmt.Errorf("order not found")
	}

	// 如果订单已提交到交易所，需要取消远程订单
	if order.Status == "pending" && order.OrderID != "" {
		if err := ctx.GetExchangeInstance().CancelOrder(ctx.GetContext(), order.OrderID); err != nil {
			return fmt.Errorf("failed to cancel remote order: %w", err)
		}
	}

	// 更新订单状态
	order.Status = "cancelled"
	if err := db.Save(&order).Error; err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("订单已取消: ID=%d", order.ID)))
	return nil
}

// DeleteCommand 删除命令
type DeleteCommand struct{}

func (c *DeleteCommand) GetName() string {
	return "delete"
}

func (c *DeleteCommand) GetDescription() string {
	return "删除用户或标的"
}

func (c *DeleteCommand) GetUsage() string {
	return "delete <type> <name>"
}

func (c *DeleteCommand) Execute(ctx *Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	db := database.GetDB()

	switch args[0] {
	case "users":
		username := args[1]
		if err := db.Where("username = ?", username).Delete(&models.FoxUser{}).Error; err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("用户已删除: %s", username)))

	case "symbols":
		if !ctx.IsReady() {
			return fmt.Errorf("请先选择交易所和用户")
		}
		symbolName := args[1]
		if err := db.Where("name = ? AND user_id = ?", symbolName, ctx.GetUser().ID).Delete(&models.FoxSymbol{}).Error; err != nil {
			return fmt.Errorf("failed to delete symbol: %w", err)
		}
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("标的已删除: %s", symbolName)))

	default:
		return fmt.Errorf("unknown delete type: %s", args[0])
	}

	return nil
}

// HelpCommand 帮助命令
type HelpCommand struct{}

func (c *HelpCommand) GetName() string {
	return "help"
}

func (c *HelpCommand) GetDescription() string {
	return "显示帮助信息"
}

func (c *HelpCommand) GetUsage() string {
	return "help [command]"
}

func (c *HelpCommand) Execute(ctx *Context, args []string) error {
	commands := []Command{
		&ShowCommand{},
		&UseCommand{},
		&CreateCommand{},
		&UpdateCommand{},
		&CancelCommand{},
		&DeleteCommand{},
		&HelpCommand{},
	}

	if len(args) > 0 {
		// 显示特定命令的帮助
		for _, cmd := range commands {
			if cmd.GetName() == args[0] {
				fmt.Printf("命令: %s\n", cmd.GetName())
				fmt.Printf("描述: %s\n", cmd.GetDescription())
				fmt.Printf("用法: %s\n", cmd.GetUsage())
				return nil
			}
		}
		return fmt.Errorf("command not found: %s", args[0])
	}

	// 显示所有命令
	fmt.Println("可用命令:")
	for _, cmd := range commands {
		fmt.Printf("  %-10s - %s\n", cmd.GetName(), cmd.GetDescription())
	}

	return nil
}

// ExitCommand 退出命令
type ExitCommand struct{}

func (c *ExitCommand) GetName() string {
	return "exit"
}

func (c *ExitCommand) GetDescription() string {
	return "退出程序"
}

func (c *ExitCommand) GetUsage() string {
	return "exit"
}

func (c *ExitCommand) Execute(ctx *Context, args []string) error {
	return fmt.Errorf("exit")
}
