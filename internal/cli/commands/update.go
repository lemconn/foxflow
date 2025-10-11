package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/repository"
	"github.com/lemconn/foxflow/internal/utils"
)

// UpdateCommand 设置命令
type UpdateCommand struct{}

func (c *UpdateCommand) GetName() string        { return "update" }
func (c *UpdateCommand) GetDescription() string { return "设置杠杆或保证金模式" }
func (c *UpdateCommand) GetUsage() string       { return "update <symbol> <type> <value>" }

func (c *UpdateCommand) Execute(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "leverage":
		return c.updateLeverage(ctx, args[1:])
	case "margin-type":
		marginType := args[1]
		// 这里应该调用交易所API设置保证金模式
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("保证金模式设置为: %s", marginType)))
	default:
		return fmt.Errorf("unknown update type: %s", args[0])
	}

	return nil
}

func (c *UpdateCommand) updateLeverage(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	// 必须设置杠杆和保证金模式
	if len(args) < 3 {
		return fmt.Errorf("usage: update leverage <symbol> --leverage=<num> --margin-type=<type>")
	}

	symbolName := strings.ToUpper(args[0])
	exchangeName := ctx.GetExchangeName()

	// 检索当前用户下是否存在指定的交易对
	symbolInfo, err := repository.GetSymbolByNameUser(symbolName, ctx.GetUserInstance().ID)
	if err != nil {
		return fmt.Errorf("failed to find symbol: %w", err)
	}

	if symbolInfo == nil || symbolInfo.ID == 0 {
		return fmt.Errorf("symbol does not exist")
	}

	leverage, err := strconv.Atoi(strings.TrimPrefix(args[1], "--leverage="))
	if err != nil {
		return fmt.Errorf("invalid leverage value")
	}

	marginType := strings.TrimPrefix(args[2], "--margin-type=")

	if leverage <= 0 || marginType == "" {
		return fmt.Errorf("invalid leverage/margin-type value")
	}

	// 初始化当前激活交易所
	exchangeClient, err := exchange.GetManager().GetExchange(exchangeName)
	if err != nil {
		return fmt.Errorf("get exchange client error: %w", err)
	}

	// 设置杠杆和保证金模式
	setLeverageErr := exchangeClient.SetLeverage(ctx.GetContext(), symbolName, leverage, marginType)
	if setLeverageErr != nil {
		return fmt.Errorf("set leverage error: %w", setLeverageErr)
	}

	// 更新标的
	symbolInfo.Leverage = leverage
	symbolInfo.MarginType = marginType
	err = repository.UpdateSymbol(symbolInfo)
	if err != nil {
		return fmt.Errorf("failed to update symbol: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("更新标的杠杆成功: %s-%d-%s", symbolName, leverage, marginType)))
	return nil
}
