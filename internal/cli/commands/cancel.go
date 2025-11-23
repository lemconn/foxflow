package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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

	if grpcClient := ctx.GetGRPCClient(); grpcClient != nil {
		fmt.Println(utils.RenderInfo("正在通过 gRPC 取消订单..."))
		message, orderID, err := grpcClient.CancelOrder(
			ctx.GetAccountInstance().Id,
			ctx.GetExchangeName(),
			symbol,
			side,
			posSide,
			amountDecimal.String(),
			amountType,
		)
		if err == nil {
			if orderID != "" {
				fmt.Println(utils.RenderSuccess(fmt.Sprintf("%s (订单号: %s)", message, orderID)))
			} else {
				fmt.Println(utils.RenderSuccess(message))
			}
			return nil
		}
		fmt.Println(utils.RenderWarning(fmt.Sprintf("gRPC 取消订单失败，回退到本地模式: %v", err)))
	}

	return c.cancelOrderLocal(symbol, side, posSide, amountDecimal.String(), amountType, args[1])
}

func (c *CancelCommand) cancelOrderLocal(symbol, side, posSide, amount, amountType, raw string) error {
	targetOrder, err := database.Adapter().FoxOrder.Where(
		database.Adapter().FoxOrder.Status.Eq("waiting"),
		database.Adapter().FoxOrder.Symbol.Eq(symbol),
		database.Adapter().FoxOrder.Side.Eq(side),
		database.Adapter().FoxOrder.PosSide.Eq(posSide),
		database.Adapter().FoxOrder.Size.Eq(amount),
		database.Adapter().FoxOrder.SizeType.Eq(amountType),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("find order error: %w", err)
	}
	if targetOrder == nil {
		return fmt.Errorf("order not found: %s", raw)
	}

	if targetOrder.Status != "waiting" {
		return fmt.Errorf("order %s status is not waiting", symbol)
	}

	targetOrder.Status = "cancelled"
	if err := database.Adapter().FoxOrder.Save(targetOrder); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("订单（%s）取消成功", raw)))
	return nil
}
