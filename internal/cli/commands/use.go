package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/grpc"
	"github.com/lemconn/foxflow/internal/utils"
)

// UseCommand 激活命令
type UseCommand struct{}

func (c *UseCommand) GetName() string        { return "use" }
func (c *UseCommand) GetDescription() string { return "激活交易所或用户" }
func (c *UseCommand) GetUsage() string {
	return `use <type> <name> \n  types: exchange（激活交易所）, account（激活交易账户）\n `
}

func (c *UseCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("%s", c.GetUsage())
	}

	switch strings.ToLower(args[0]) {
	case "exchange":
		return c.HandleExchangeCommand(ctx, args[1])
	case "account":
		return c.HandleAccountCommand(ctx, args[1])
	default:
		return fmt.Errorf("unknown use type: %s", args[0])
	}
}

func (c *UseCommand) HandleExchangeCommand(ctx command.Context, exchangeName string) error {
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	exchangeItem, err := grpcClient.UseExchange(exchangeName)
	if err != nil {
		return fmt.Errorf("激活交易所失败: %v", err)
	}

	ctx.SetExchangeName(exchangeItem.Name)
	ctx.SetExchangeInstance(&grpc.ShowExchangeItem{
		Name:        exchangeItem.Name,
		APIUrl:      exchangeItem.APIUrl,
		ProxyUrl:    exchangeItem.ProxyUrl,
		StatusValue: exchangeItem.StatusValue,
	})
	ctx.SetAccountName("")
	ctx.SetAccountInstance(&grpc.ShowAccountItem{})
	fmt.Println(utils.RenderSuccess(fmt.Sprintf("已激活交易所: %s", utils.MessageGreen(exchangeItem.Name))))
	return nil
}

func (c *UseCommand) HandleAccountCommand(ctx command.Context, name string) error {
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	accountItem, exchangeItem, err := grpcClient.UseAccount(name)
	if err != nil {
		return fmt.Errorf("gRPC 激活账户失败，回退到本地模式: %v", err)
	}

	ctx.SetExchangeName(exchangeItem.Name)
	ctx.SetExchangeInstance(exchangeItem)
	ctx.SetAccountName(accountItem.Name)
	ctx.SetAccountInstance(accountItem)

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("已激活用户: %s", utils.MessageGreen(accountItem.Name))))
	return nil
}
