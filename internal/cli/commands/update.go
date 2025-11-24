package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/grpc"
	"github.com/lemconn/foxflow/internal/utils"
)

// UpdateCommand 设置命令
type UpdateCommand struct{}

func (c *UpdateCommand) GetName() string        { return "update" }
func (c *UpdateCommand) GetDescription() string { return "设置杠杆或保证金模式" }
func (c *UpdateCommand) GetUsage() string {
	return "update <type> [options]\n  types: symbol（更新交易对杠杆倍数和保证金模式）, account（更新交易账户信息）\n "
}

func (c *UpdateCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "symbol":
		return c.handleSymbolCommand(ctx, args[1:])
	case "account":
		return c.handleAccountCommand(ctx, args[1:])
	default:
		return fmt.Errorf("unknown update type: %s", args[0])
	}
}

// handleSymbolCommand 更新标的的保证金模式/杠杆倍数
func (c *UpdateCommand) handleSymbolCommand(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	// 必须设置杠杆和保证金模式
	if len(args) < 3 {
		return fmt.Errorf("update symbol <symbol> margin=<type> leverage=<num>")
	}

	symbolName := strings.ToUpper(args[0])
	exchangeName := ctx.GetExchangeName()

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	exchangeSymbol, err := grpcClient.GetSymbols(exchangeName, "")
	if err != nil {
		return err
	}

	var symbolInfo *grpc.ShowSymbolItem
	for _, symbol := range exchangeSymbol {
		if symbol.Name == symbolName {
			symbolInfo = symbol
			break
		}
	}

	if symbolInfo == nil || symbolInfo.Name == "" {
		return fmt.Errorf("symbol does not exist")
	}

	var marginType, leverage string
	for _, arg := range args {
		if strings.HasPrefix(arg, "margin=") {
			marginType = strings.TrimPrefix(arg, "margin=")
		} else if strings.HasPrefix(arg, "leverage=") {
			leverage = strings.TrimPrefix(arg, "leverage=")
		}
	}

	leverageValue, err := strconv.ParseInt(leverage, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid leverage value")
	}

	if leverageValue <= 0 || marginType == "" {
		return fmt.Errorf("invalid leverage/margin value")
	}

	if leverageValue > symbolInfo.MaxLeverage {
		return fmt.Errorf("leverage value is too large, max leverage is %d", symbolInfo.MaxLeverage)
	}

	if ctx.GetAccountInstance() == nil {
		return fmt.Errorf("account instance is nil")
	}

	err = grpcClient.UpdateSymbol(ctx.GetAccountInstance().Id, exchangeName, symbolName, marginType, leverageValue)
	if err != nil {
		return fmt.Errorf("gRPC 更新标的失败: %v", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("更新标的杠杆成功: %s:%s:%s", symbolName, marginType, leverage)))
	return nil
}

// handleAccountCommand 更新账户信息
func (c *UpdateCommand) handleAccountCommand(ctx command.Context, args []string) error {
	if ctx.GetExchangeName() == "" {
		return fmt.Errorf("请优先选择交易所")
	}

	// 必须设置杠杆和保证金模式
	if len(args) < 6 {
		return fmt.Errorf("update account <name> <trade_type> name= apiKey=<apiKey> secretKey=<secretKey> passphrase=<passphrase>")
	}

	targetAccount := args[0]
	tradeType := args[1]
	var name, apiKey, secretKey, passphrase string
	for _, arg := range args[2:] {
		switch {
		case strings.HasPrefix(arg, "name="):
			name = strings.TrimPrefix(arg, "name=")
		case strings.HasPrefix(arg, "apiKey="):
			apiKey = strings.TrimPrefix(arg, "apiKey=")
		case strings.HasPrefix(arg, "secretKey="):
			secretKey = strings.TrimPrefix(arg, "secretKey=")
		case strings.HasPrefix(arg, "passphrase="):
			passphrase = strings.TrimPrefix(arg, "passphrase=")
		}
	}

	if tradeType == "" || name == "" || apiKey == "" || secretKey == "" {
		return fmt.Errorf("missing required parameters")
	}
	if ctx.GetExchangeName() == config.DefaultExchange && passphrase == "" {
		return fmt.Errorf("okx exchange passphrase is required")
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	accountItem, err := grpcClient.UpdateAccount(ctx.GetExchangeName(), targetAccount, tradeType, name, apiKey, secretKey, passphrase)
	if err != nil {
		return fmt.Errorf("更新账户失败: %v", err)
	}

	ctx.SetAccountName(accountItem.Name)
	ctx.SetAccountInstance(accountItem)

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("更新账户成功: %s", targetAccount)))
	return nil
}
