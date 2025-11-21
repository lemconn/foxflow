package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"github.com/lemconn/foxflow/internal/utils"
	"gorm.io/gorm"
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

	if grpcClient := ctx.GetGRPCClient(); grpcClient != nil {
		fmt.Println(utils.RenderInfo(fmt.Sprintf("正在通过 gRPC 更新 [%s] 的 %s 杠杆...", ctx.GetAccountName(), margin)))
		updatedAccount, err := grpcClient.UpdateTradeConfig(account.Id, margin, int64(leverageNum))
		if err == nil {
			ctx.SetAccountInstance(updatedAccount)
			fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户设置杠杆倍数成功：%s：%d", ctx.GetAccountName(), margin, leverageNum)))
			return nil
		}
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 更新杠杆失败，回退到本地模式: %v", err)))
	}

	return c.handleConfigCommandLocal(ctx, margin, leverageNum)
}

func (c *SetCommand) handleProxyCommand(ctx command.Context, args []string) error {
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

	if grpcClient := ctx.GetGRPCClient(); grpcClient != nil {
		fmt.Println(utils.RenderInfo(fmt.Sprintf("正在通过 gRPC 更新 [%s] 的网络代理...", ctx.GetAccountName())))
		updatedAccount, err := grpcClient.UpdateProxyConfig(account.Id, proxyUrl)
		if err == nil {
			ctx.SetAccountInstance(updatedAccount)
			if proxyUrl == "" {
				fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户取消网络代理设置成功", ctx.GetAccountName())))
			} else {
				fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户网络代理设置成功，默认代理：%s", ctx.GetAccountName(), proxyUrl)))
			}
			return nil
		}
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 更新代理失败，回退到本地模式: %v", err)))
	}

	return c.handleProxyCommandLocal(ctx, proxyUrl)
}

func (c *SetCommand) handleConfigCommandLocal(ctx command.Context, margin string, leverageNum int) error {
	accountID := ctx.GetAccountInstance().Id

	accountMarginLeverage, err := database.Adapter().FoxTradeConfig.Where(
		database.Adapter().FoxTradeConfig.AccountID.Eq(accountID),
		database.Adapter().FoxTradeConfig.Margin.Eq(margin),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("find account by name error: %w", err)
	}

	if accountMarginLeverage == nil {
		tradeConfig := &model.FoxTradeConfig{
			AccountID: accountID,
			Margin:    margin,
			Leverage:  int64(leverageNum),
		}
		if err := database.Adapter().FoxTradeConfig.Create(tradeConfig); err != nil {
			return fmt.Errorf("create account error: %w", err)
		}
	} else {
		accountMarginLeverage.Leverage = int64(leverageNum)
		if err := database.Adapter().FoxTradeConfig.Save(accountMarginLeverage); err != nil {
			return fmt.Errorf("update account error: %w", err)
		}
	}

	if account := ctx.GetAccountInstance(); account != nil {
		if margin == "cross" {
			account.CrossLeverage = int64(leverageNum)
		} else {
			account.IsolatedLeverage = int64(leverageNum)
		}
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户设置杠杆倍数成功：%s：%d", ctx.GetAccountName(), margin, leverageNum)))
	return nil
}

func (c *SetCommand) handleProxyCommandLocal(ctx command.Context, proxyUrl string) error {
	accountID := ctx.GetAccountInstance().Id
	accountConfigInfo, err := database.Adapter().FoxConfig.Where(
		database.Adapter().FoxConfig.AccountID.Eq(accountID),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if accountConfigInfo == nil {
		if proxyUrl == "" {
			fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户取消网络代理设置成功", ctx.GetAccountName())))
			return nil
		}

		configInfo := &model.FoxConfig{
			AccountID: accountID,
			ProxyURL:  proxyUrl,
		}
		if err := database.Adapter().FoxConfig.Create(configInfo); err != nil {
			return err
		}
	} else {
		accountConfigInfo.ProxyURL = proxyUrl
		if err := database.Adapter().FoxConfig.Save(accountConfigInfo); err != nil {
			return err
		}
	}

	if account := ctx.GetAccountInstance(); account != nil {
		account.ProxyUrl = proxyUrl
	}

	if proxyUrl == "" {
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户取消网络代理设置成功", ctx.GetAccountName())))
	} else {
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户网络代理设置成功，默认代理：%s", ctx.GetAccountName(), proxyUrl)))
	}
	return nil
}
