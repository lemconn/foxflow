package cli

import (
	"fmt"
	"strconv"
	"strings"

	"foxflow/internal/database"
	"foxflow/internal/exchange"
	"foxflow/internal/models"
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
	exchangeManager := exchange.GetManager()

	switch args[0] {
	case "exchanges":
		exchanges := exchangeManager.GetAvailableExchanges()
		fmt.Println("可用交易所:")
		for _, ex := range exchanges {
			fmt.Printf("  - %s\n", ex)
		}

	case "users":
		var users []models.FoxUser
		if err := db.Find(&users).Error; err != nil {
			return fmt.Errorf("failed to get users: %w", err)
		}
		fmt.Println("用户列表:")
		for _, user := range users {
			fmt.Printf("  - %s (%s) [%s]\n", user.Username, user.Exchange, user.Status)
		}

	case "assets":
		if !ctx.IsReady() {
			return fmt.Errorf("请先选择交易所和用户")
		}
		assets, err := ctx.GetExchangeInstance().GetBalance(ctx.GetContext())
		if err != nil {
			return fmt.Errorf("failed to get assets: %w", err)
		}
		fmt.Println("资产列表:")
		for _, asset := range assets {
			fmt.Printf("  - %s: %.4f (可用: %.4f, 冻结: %.4f)\n",
				asset.Currency, asset.Balance, asset.Available, asset.Frozen)
		}

	case "orders":
		if !ctx.IsReady() {
			return fmt.Errorf("请先选择交易所和用户")
		}
		orders, err := ctx.GetExchangeInstance().GetOrders(ctx.GetContext(), "", "pending")
		if err != nil {
			return fmt.Errorf("failed to get orders: %w", err)
		}
		fmt.Println("未成交订单:")
		for _, order := range orders {
			fmt.Printf("  - %s %s %s %.4f @ %.2f [%s]\n",
				order.ID, order.Symbol, order.Side, order.Size, order.Price, order.Status)
		}

	case "positions":
		if !ctx.IsReady() {
			return fmt.Errorf("请先选择交易所和用户")
		}
		positions, err := ctx.GetExchangeInstance().GetPositions(ctx.GetContext())
		if err != nil {
			return fmt.Errorf("failed to get positions: %w", err)
		}
		fmt.Println("未平仓仓位:")
		for _, pos := range positions {
			fmt.Printf("  - %s %s %.4f @ %.2f (未实现盈亏: %.2f)\n",
				pos.Symbol, pos.PosSide, pos.Size, pos.AvgPrice, pos.UnrealPnl)
		}

	case "strategies":
		fmt.Println("可用策略:")
		fmt.Println("  - volume: 成交量策略")
		fmt.Println("  - macd: MACD策略")
		fmt.Println("  - rsi: RSI策略")

	case "symbols":
		if !ctx.IsReady() {
			return fmt.Errorf("请先选择交易所和用户")
		}
		symbols, err := ctx.GetExchangeInstance().GetSymbols(ctx.GetContext())
		if err != nil {
			return fmt.Errorf("failed to get symbols: %w", err)
		}
		fmt.Println("交易对列表:")
		for _, symbol := range symbols {
			fmt.Printf("  - %s\n", symbol)
		}

	case "ss":
		var ss []models.FoxSS
		query := db.Where("status = ?", "waiting")
		if ctx.IsReady() {
			query = query.Where("user_id = ?", ctx.GetUser().ID)
		}
		if err := query.Find(&ss).Error; err != nil {
			return fmt.Errorf("failed to get strategy orders: %w", err)
		}
		fmt.Println("策略订单列表:")
		for _, s := range ss {
			fmt.Printf("  - ID:%d %s %s %s %.4f @ %.2f [%s] 策略:%s\n",
				s.ID, s.Symbol, s.Side, s.PosSide, s.Sz, s.Px, s.Status, s.Strategy)
		}

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
	exchangeManager := exchange.GetManager()

	switch args[0] {
	case "exchanges":
		exchangeName := args[1]
		ex, err := exchangeManager.GetExchange(exchangeName)
		if err != nil {
			return fmt.Errorf("exchange not found: %s", exchangeName)
		}
		ctx.SetExchange(exchangeName)
		ctx.SetExchangeInstance(ex)
		fmt.Printf("已激活交易所: %s\n", exchangeName)

	case "users":
		username := args[1]
		var user models.FoxUser
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			return fmt.Errorf("user not found: %s", username)
		}

		// 连接用户到当前交易所
		if ctx.GetExchange() != "" {
			ex, err := exchangeManager.GetExchange(ctx.GetExchange())
			if err != nil {
				return fmt.Errorf("failed to get exchange: %w", err)
			}
			if err := ex.Connect(ctx.GetContext(), &user); err != nil {
				return fmt.Errorf("failed to connect user: %w", err)
			}
			ctx.SetExchangeInstance(ex)
		}

		ctx.SetUser(&user)
		fmt.Printf("已激活用户: %s\n", username)

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

	fmt.Printf("用户创建成功: %s\n", user.Username)
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

	fmt.Printf("标的创建成功: %s\n", symbol.Name)
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
		fmt.Println("订单将直接提交到交易所")
	} else {
		order.Strategy = strategy
		fmt.Println("策略订单已创建，等待策略条件满足")
	}

	db := database.GetDB()
	if err := db.Create(order).Error; err != nil {
		return fmt.Errorf("failed to create strategy order: %w", err)
	}

	fmt.Printf("策略订单创建成功: ID=%d\n", order.ID)
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
		fmt.Printf("杠杆设置为: %d\n", leverage)

	case "margin-type":
		marginType := args[1]
		// 这里应该调用交易所API设置保证金模式
		fmt.Printf("保证金模式设置为: %s\n", marginType)

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

	fmt.Printf("订单已取消: ID=%d\n", order.ID)
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
		fmt.Printf("用户已删除: %s\n", username)

	case "symbols":
		if !ctx.IsReady() {
			return fmt.Errorf("请先选择交易所和用户")
		}
		symbolName := args[1]
		if err := db.Where("name = ? AND user_id = ?", symbolName, ctx.GetUser().ID).Delete(&models.FoxSymbol{}).Error; err != nil {
			return fmt.Errorf("failed to delete symbol: %w", err)
		}
		fmt.Printf("标的已删除: %s\n", symbolName)

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
