package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/engine/syntax"
	"github.com/lemconn/foxflow/internal/utils"
	"github.com/shopspring/decimal"
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

	if posSide != "long" && posSide != "short" {
		return fmt.Errorf("direction 参数错误，只能为 long 或 short")
	}
	if margin != "isolated" && margin != "cross" {
		return fmt.Errorf("margin 参数错误，只能为 isolated 或 cross")
	}
	if amount == "" {
		return fmt.Errorf("amount 参数不能为空，例：100/100U")
	}

	amountType := ""
	if strings.HasSuffix(amount, "U") {
		amount = strings.TrimSuffix(amount, "U")
		amountType = "USDT"
	}

	amountDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		return fmt.Errorf("amount decimal error: %w", err)
	}

	strategy := ""
	if len(args) >= 6 {
		strategy = args[5]
		engineClient := syntax.NewEngine()
		node, err := engineClient.Parse(strategy)
		if err != nil {
			return fmt.Errorf("failed to parse strategy syntax: %w", err)
		}
		if err := engineClient.GetEvaluator().Validate(node); err != nil {
			return fmt.Errorf("failed to validate AST: %w", err)
		}
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	message, err := grpcClient.OpenOrder(
		ctx.GetAccountInstance().Id,
		ctx.GetExchangeName(),
		symbolName,
		posSide,
		margin,
		amountDecimal.String(),
		amountType,
		strategy,
	)
	if err != nil {
		return fmt.Errorf("提交订单失败: %v", err)
	}

	fmt.Println(utils.RenderSuccess(message))
	return nil
}
