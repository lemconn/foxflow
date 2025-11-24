package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/utils"
)

type SetCommand struct{}

func (c *SetCommand) GetName() string { return "update" }
func (c *SetCommand) GetDescription() string {
	return "设置默认保证金模式以及对应杠杆倍数/网络代理"
}
func (c *SetCommand) GetUsage() string {
	return "set <type> [options]\n  types: config（设置默认保证金模式以及对应杠杆倍数）, proxy（设置网络代理）\n "
}

func (c *SetCommand) Execute(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	switch strings.ToLower(args[0]) {
	case "config":
		return c.handleConfigCommand(ctx, args[1:])
	case "proxy":
		return c.handleProxyCommand(ctx, args[1:])
	default:
		return fmt.Errorf("unknown set type: %s", args[0])
	}
}

func (c *SetCommand) handleConfigCommand(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	if len(args) < 2 {
		return fmt.Errorf("当前参数不全，请补全参数，set config <margin> <leverage>  例：set config isolated/cross leverage=10")
	}

	account := ctx.GetAccountInstance()
	if account == nil {
		return fmt.Errorf("请先选择交易账户")
	}

	margin := strings.ToLower(args[0])
	if margin != "isolated" && margin != "cross" {
		return fmt.Errorf("margin 参数错误，只能为 isolated 或 cross")
	}

	leverage := ""
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "leverage=") {
			leverage = strings.TrimPrefix(arg, "leverage=")
		}
	}
	leverageNum, err := strconv.Atoi(leverage)
	if err != nil {
		return fmt.Errorf("leverage= 参数错误，只能为数字")
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	updatedAccount, err := grpcClient.UpdateTradeConfig(account.Id, margin, int64(leverageNum))
	if err != nil {
		return fmt.Errorf("更新杠杆失败: %v", err)
	}

	ctx.SetAccountInstance(updatedAccount)

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户设置杠杆倍数成功：%s：%d", ctx.GetAccountName(), margin, leverageNum)))
	return nil
}

func (c *SetCommand) handleProxyCommand(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return errors.New("请先选择交易所和用户")
	}

	account := ctx.GetAccountInstance()
	if account == nil {
		return fmt.Errorf("请先选择交易账户")
	}

	proxyUrl := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "url=") {
			proxyUrl = strings.TrimPrefix(arg, "url=")
		}
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	updatedAccount, err := grpcClient.UpdateProxyConfig(account.Id, proxyUrl)
	if err != nil {
		return fmt.Errorf("更新代理失败: %v", err)
	}

	ctx.SetAccountInstance(updatedAccount)
	if proxyUrl == "" {
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户取消网络代理设置成功", ctx.GetAccountName())))
	} else {
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户网络代理设置成功，默认代理：%s", ctx.GetAccountName(), proxyUrl)))
	}

	return nil
}
