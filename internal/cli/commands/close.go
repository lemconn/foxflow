package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/engine/syntax"
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

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	message, err := grpcClient.CloseOrder(
		ctx.GetAccountInstance().Id,
		ctx.GetExchangeName(),
		symbolName,
		posSide,
		margin,
		strategy,
	)
	if err != nil {
		return fmt.Errorf("提交平仓订单失败: %v", err)
	}

	fmt.Println(utils.RenderSuccess(message))
	return nil
}
