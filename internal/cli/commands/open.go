package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/models"
	"github.com/lemconn/foxflow/internal/repository"
	"github.com/lemconn/foxflow/internal/utils"
)

// OpenCommand 退出命令
type OpenCommand struct{}

func (c *OpenCommand) GetName() string        { return "open" }
func (c *OpenCommand) GetDescription() string { return "开仓/下单" }
func (c *OpenCommand) GetUsage() string       { return "open <symbol> <direction> <amount> [with]" }

func (c *OpenCommand) Execute(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	symbolName := strings.ToUpper(args[0])
	posSide := strings.ToLower(args[1])
	amount := strings.ToUpper(args[2])

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
		return fmt.Errorf("amount 参数不能为空")
	}

	// 判断amount参数是否存在U的后缀（目前仅支持U）
	szType := ""
	if strings.HasSuffix(amount, "U") {
		amount = strings.TrimSuffix(amount, "U")
		szType = "U"
	}
	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("amount 参数错误，只能为数字")
	}
	if amountValue <= 0 {
		return fmt.Errorf("amount 参数错误，只能为大于0的数字")
	}

	stategry := ""
	if len(args) >= 5 {
		// @TODO 校验策略

		stategry = args[4]
	}

	// 解析参数
	order := &models.FoxSS{
		UserID:    ctx.GetAccountInstance().ID,
		Symbol:    symbolName,
		PosSide:   posSide,
		Sz:        amountValue,
		SzType:    szType,
		Side:      "buy",
		OrderType: "market",
		Strategy:  stategry,
		Type:      "open",
		Status:    "waiting",
	}

	// 提交到当前激活交易所
	exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeName())
	if err != nil {
		return fmt.Errorf("get exchange client error: %w", err)
	}

	exchangeSmbolInfo, err := exchangeClient.GetSymbols(ctx.GetContext(), symbolName)
	if err != nil {
		return fmt.Errorf("get exchange symbols error: %w", err)
	}

	fmt.Printf("==========[%+v]============[%+v]=========\n", exchangeSmbolInfo, order)

	//// 构造交易所订单
	//exOrder := &exchange.Order{
	//	Symbol:     order.Symbol,
	//	Side:       order.Side,
	//	PosSide:    order.PosSide,
	//	Price:      order.Px,
	//	Size:       order.Sz,
	//	Type:       order.OrderType,
	//	MarginType: "isolated",
	//}
	//
	//createdOrder, err := exchangeClient.CreateOrder(ctx.GetContext(), exOrder)
	//if err != nil {
	//	return fmt.Errorf("create exchange order error: %w", err)
	//}

	//fmt.Printf("-------------[%+v]---------------[%+v]---------------[%+v]---------\n", createdOrder, exOrder, order)

	if err := repository.CreateSSOrder(order); err != nil {
		return fmt.Errorf("failed to create strategy order: %w", err)
	}

	fmt.Println(utils.RenderInfo("策略订单已创建，等待策略条件满足"))

	return nil
}
