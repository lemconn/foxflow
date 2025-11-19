package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/engine/syntax"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"github.com/lemconn/foxflow/internal/utils"
)

// CloseCommand 退出命令
type CloseCommand struct{}

func (c *CloseCommand) GetName() string        { return "close" }
func (c *CloseCommand) GetDescription() string { return "平仓" }
func (c *CloseCommand) GetUsage() string {
	return "close <symbol> <direction> <margin> [with] [strategy]"
}

func (c *CloseCommand) Execute(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	if len(args) < 3 {
		return fmt.Errorf("参数缺失，请补全参数，例：close <symbol> <direction> <margin> [with] [strategy]")
	}

	symbolName := strings.ToUpper(args[0])
	posSide := strings.ToLower(args[1])
	margin := strings.ToLower(args[2])

	if posSide != "long" && posSide != "short" {
		return fmt.Errorf("direction 参数错误，只能为 long 或 short")
	}
	if margin != "isolated" && margin != "cross" {
		return fmt.Errorf("margin 参数错误，只能为 isolated 或 cross")
	}

	// 校验策略
	var stategry string
	if len(args) >= 5 {
		engineClient := syntax.NewEngine()
		// 解析语法表达式
		node, err := engineClient.Parse(args[4])
		if err != nil {
			return fmt.Errorf("failed to parse strategy syntax: %w", err)
		}

		// 验证AST
		if err := engineClient.GetEvaluator().Validate(node); err != nil {
			return fmt.Errorf("failed to validate AST: %w", err)
		}
	}

	var side string
	if posSide == "long" {
		side = "sell"
	} else {
		side = "buy"
	}

	// 激活交易所
	exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeName())
	if err != nil {
		return fmt.Errorf("get exchange client error: %w", err)
	}

	// 解析参数
	order := &model.FoxOrder{
		OrderID:    exchangeClient.GetClientOrderId(ctx.GetContext()),
		Exchange:   ctx.GetExchangeName(),
		AccountID:  ctx.GetAccountInstance().Id,
		Symbol:     symbolName,
		PosSide:    posSide,
		MarginType: margin,
		Side:       side,
		OrderType:  "market",
		Strategy:   stategry,
		Type:       "close",
		Status:     "waiting",
	}
	if err := database.Adapter().FoxOrder.Create(order); err != nil {
		return fmt.Errorf("create order error: %w", err)
	}

	fmt.Println(utils.RenderInfo("策略订单已创建，等待策略条件满足"))

	return nil
}
