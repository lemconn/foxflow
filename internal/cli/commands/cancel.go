package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/utils"
	"github.com/shopspring/decimal"
)

// CancelCommand 取消命令
type CancelCommand struct{}

func (c *CancelCommand) GetName() string        { return "cancel" }
func (c *CancelCommand) GetDescription() string { return "取消订单" }
func (c *CancelCommand) GetUsage() string       { return "cancel order <symbol>:<side>:<posSide>:<amount>" }

func (c *CancelCommand) Execute(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}
	if args[0] != "order" {
		return fmt.Errorf("only support cancel order")
	}

	// 解析订单标识：symbol:direction:amount
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	orderParts := strings.Split(args[1], ":")
	if len(orderParts) != 4 {
		return fmt.Errorf("invalid order format, expected: symbol:side:posSide:amount")
	}

	symbol := strings.ToUpper(orderParts[0])
	side := orderParts[1]
	posSide := orderParts[2]
	amountStr := orderParts[3]

	amountType := ""
	if strings.HasSuffix(amountStr, "U") {
		amountStr = strings.TrimSuffix(amountStr, "U")
		amountType = "USDT"
	}

	amountDecimal, err := decimal.NewFromString(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount decimal: %w", err)
	}

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	message, err := grpcClient.CancelOrder(
		ctx.GetAccountInstance().Id,
		ctx.GetExchangeName(),
		symbol,
		side,
		posSide,
		amountDecimal.String(),
		amountType,
	)
	if err != nil {
		return fmt.Errorf("取消订单失败: %v", err)
	}

	fmt.Println(utils.RenderSuccess(message))
	return nil
}
