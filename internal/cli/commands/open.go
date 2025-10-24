package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/engine/syntax"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/models"
	"github.com/lemconn/foxflow/internal/repository"
	"github.com/lemconn/foxflow/internal/utils"
)

// OpenCommand 退出命令
type OpenCommand struct{}

func (c *OpenCommand) GetName() string        { return "open" }
func (c *OpenCommand) GetDescription() string { return "开仓/下单" }
func (c *OpenCommand) GetUsage() string {
	return "open <symbol> <direction> <margin> <amount> [with] [strategy]"
}

func (c *OpenCommand) Execute(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	if len(args) < 4 {
		return fmt.Errorf("当前参数不全，请补全参数，例：open BTC-USDT-SWAP isolated long 1000/1000U [with] [strategy]")
	}

	symbolName := strings.ToUpper(args[0])
	posSide := strings.ToLower(args[1])
	margin := strings.ToLower(args[2])
	amount := strings.ToUpper(args[3])

	exchangeSymbolList, exist := config.ExchangeSymbolList[ctx.GetExchangeName()]
	if !exist {
		return fmt.Errorf("交易所 %s 交易对信息不存在", ctx.GetExchangeName())
	}

	// 校验交易对数据
	symbolInfo := config.SymbolInfo{}
	for _, symbol := range exchangeSymbolList {
		if symbol.Name == symbolName {
			symbolInfo = symbol
			break
		}
	}
	if symbolInfo.Name == "" {
		return fmt.Errorf("交易所 %s 交易对 %s 信息不存在", ctx.GetExchangeName(), symbolName)
	}

	if posSide != "long" && posSide != "short" {
		// 校验是否存在U的后缀
		return fmt.Errorf("direction 参数错误，只能为 long 或 short")
	}

	if amount == "" {
		return fmt.Errorf("amount 参数不能为空，例：100/100U")
	}

	// 判断amount参数是否存在U的后缀（目前仅支持U）
	amountType := ""
	if strings.HasSuffix(amount, "U") {
		amount = strings.TrimSuffix(amount, "U")
		amountType = "USDT"
	}
	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("amount 参数错误，只能为数字或者带有U单位，例：100/100U")
	}
	if amountValue <= 0 {
		return fmt.Errorf("amount 参数错误，只能为大于0的数字或者带有U单位，例：100/100U")
	}

	var side string
	if posSide == "long" {
		side = "buy"
	} else {
		side = "sell"
	}

	// 激活交易所
	exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeName())
	if err != nil {
		return fmt.Errorf("get exchange client error: %w", err)
	}

	// 校验当前用户提交的数据（按照当前标的价格计算校验）
	costRes, costErr := exchangeClient.CalcOrderCost(ctx.GetContext(), &exchange.OrderCostReq{
		Side:       side,
		Symbol:     symbolName,
		Amount:     amountValue,
		AmountType: amountType,
		MarginType: margin,
	})
	if costErr != nil {
		return costErr
	}

	if costRes.CanBuyWithTaker == false {
		return fmt.Errorf("当前暂时暂时不可提交订单，标的价格：%f，期望购买数（张）：%f，可用资金：%f，手续费（%s交易所收取）:%f，需要总资金：%f", costRes.MarkPrice, costRes.Contracts, costRes.AvailableFunds, ctx.GetExchangeName(), costRes.Fee, costRes.TotalRequired)
	}

	// 校验策略
	var stategry string
	if len(args) >= 6 {
		engineClient := syntax.NewEngine()
		// 解析语法表达式
		node, err := engineClient.Parse(args[5])
		if err != nil {
			return fmt.Errorf("failed to parse strategy syntax: %w", err)
		}

		// 验证AST
		if err := engineClient.GetEvaluator().Validate(node); err != nil {
			return fmt.Errorf("failed to validate AST: %w", err)
		}
	}

	// 解析参数
	order := &models.FoxSS{
		Exchange:   ctx.GetExchangeName(),
		UserID:     ctx.GetAccountInstance().ID,
		Symbol:     symbolName,
		PosSide:    posSide,
		MarginType: margin,
		Sz:         amountValue,
		SzType:     amountType,
		Side:       side,
		OrderType:  "market",
		Strategy:   stategry,
		Type:       "open",
		Status:     "waiting",
	}

	if err := repository.CreateSSOrder(order); err != nil {
		return fmt.Errorf("create order error: %w", err)
	}

	fmt.Println(utils.RenderInfo("策略订单已创建，等待策略条件满足"))

	return nil
}
