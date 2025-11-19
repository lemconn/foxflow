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

	accountMarginLeverage, err := database.Adapter().FoxTradeConfig.Where(
		database.Adapter().FoxTradeConfig.AccountID.Eq(ctx.GetAccountInstance().Id),
		database.Adapter().FoxTradeConfig.Margin.Eq(margin),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("find account by name error: %w", err)
	}

	if accountMarginLeverage == nil {
		tradeConfig := &model.FoxTradeConfig{
			AccountID: ctx.GetAccountInstance().Id,
			Margin:    margin,
			Leverage:  int64(leverageNum),
		}
		err = database.Adapter().FoxTradeConfig.Create(tradeConfig)
		if err != nil {
			return fmt.Errorf("create account error: %w", err)
		}
	} else {
		accountMarginLeverage.Leverage = int64(leverageNum)
		err = database.Adapter().FoxTradeConfig.Save(accountMarginLeverage)
		if err != nil {
			return fmt.Errorf("update account error: %w", err)
		}
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户设置杠杆倍数成功：%s：%d", ctx.GetAccountName(), margin, leverageNum)))

	return nil
}

func (c *SetCommand) handleProxyCommand(ctx command.Context, args []string) error {
	accountConfigInfo, err := database.Adapter().FoxConfig.Where(
		database.Adapter().FoxConfig.AccountID.Eq(ctx.GetAccountInstance().Id),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	proxyUrl := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "url=") {
			proxyUrl = strings.TrimPrefix(arg, "url=")
		}
	}

	if accountConfigInfo == nil {
		if proxyUrl == "" {
			return nil
		}

		configInfo := &model.FoxConfig{
			AccountID: ctx.GetAccountInstance().Id,
			ProxyURL:  proxyUrl,
		}
		err = database.Adapter().FoxConfig.Create(configInfo)
		if err != nil {
			return err
		}
	} else {
		accountConfigInfo.ProxyURL = proxyUrl
		err = database.Adapter().FoxConfig.Save(accountConfigInfo)
		if err != nil {
			return err
		}
	}

	if proxyUrl == "" {
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户取消网络代理设置成功", ctx.GetAccountName())))
	} else {
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("[%s]用户网络代理设置成功，默认代理：%s", ctx.GetAccountName(), proxyUrl)))
	}
	return nil
}
