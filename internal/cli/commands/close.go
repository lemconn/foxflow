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

	strategy := ""
	if len(args) >= 5 {
		strategy = args[4]
		engineClient := syntax.NewEngine()
		node, err := engineClient.Parse(strategy)
		if err != nil {
			return fmt.Errorf("failed to parse strategy syntax: %w", err)
		}
		if err := engineClient.GetEvaluator().Validate(node); err != nil {
			return fmt.Errorf("failed to validate AST: %w", err)
		}
	}

	if grpcClient := ctx.GetGRPCClient(); grpcClient != nil {
		fmt.Println(utils.RenderInfo("正在通过 gRPC 提交平仓订单..."))
		message, orderID, err := grpcClient.CloseOrder(
			ctx.GetAccountInstance().Id,
			ctx.GetExchangeName(),
			symbolName,
			posSide,
			margin,
			strategy,
		)
		if err == nil {
			if orderID != "" {
				fmt.Println(utils.RenderSuccess(fmt.Sprintf("%s (订单号: %s)", message, orderID)))
			} else {
				fmt.Println(utils.RenderSuccess(message))
			}
			return nil
		}
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 提交平仓订单失败，回退到本地模式: %v", err)))
	}

	return c.executeLocal(ctx, symbolName, posSide, margin, strategy)
}

func (c *CloseCommand) executeLocal(ctx command.Context, symbolName, posSide, margin, strategy string) error {
	side := "sell"
	if posSide == "short" {
		side = "buy"
	}

	exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeName())
	if err != nil {
		return fmt.Errorf("get exchange client error: %w", err)
	}

	order := &model.FoxOrder{
		OrderID:    exchangeClient.GetClientOrderId(ctx.GetContext()),
		Exchange:   ctx.GetExchangeName(),
		AccountID:  ctx.GetAccountInstance().Id,
		Symbol:     symbolName,
		PosSide:    posSide,
		MarginType: margin,
		Side:       side,
		OrderType:  "market",
		Strategy:   strategy,
		Type:       "close",
		Status:     "waiting",
	}

	if err := database.Adapter().FoxOrder.Create(order); err != nil {
		return fmt.Errorf("create order error: %w", err)
	}

	fmt.Println(utils.RenderInfo("策略订单已创建，等待策略条件满足"))
	return nil
}
