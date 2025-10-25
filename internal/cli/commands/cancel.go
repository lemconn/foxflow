package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/models"
	"github.com/lemconn/foxflow/internal/repository"
	"github.com/lemconn/foxflow/internal/utils"
)

// CancelCommand 取消命令
type CancelCommand struct{}

func (c *CancelCommand) GetName() string        { return "cancel" }
func (c *CancelCommand) GetDescription() string { return "取消订单" }
func (c *CancelCommand) GetUsage() string       { return "cancel order <symbol>:<direction>:<amount>" }

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
	if len(orderParts) != 3 {
		return fmt.Errorf("invalid order format, expected: symbol:direction:amount")
	}

	symbol := strings.ToUpper(orderParts[0])
	direction := orderParts[1]
	amountStr := orderParts[2]

	amountType := ""
	if strings.HasSuffix(amountStr, "U") {
		amountStr = strings.TrimSuffix(amountStr, "U")
		amountType = "USDT"
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %s", amountStr)
	}

	// 根据 symbol、direction、amount 查找订单
	orders, err := repository.ListSSOrders(ctx.GetAccountInstance().ID, nil)
	if err != nil {
		return fmt.Errorf("get orders error: %s", err)
	}

	var targetOrder *models.FoxOrder
	for _, order := range orders {
		if order.Symbol == symbol && order.Side == direction && order.Size == amount && order.SizeType == amountType {
			targetOrder = order
			break
		}
	}

	if targetOrder == nil {
		return fmt.Errorf("order not found: %s:%s:%s", symbol, direction, orderParts[2])
	}

	if targetOrder.Status != "waiting" {
		return fmt.Errorf("order %s status is not waiting", symbol)
	}

	// 更新订单状态
	targetOrder.Status = "cancelled"
	if err := repository.SaveSSOrder(targetOrder); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("订单已取消: %s:%s:%.4f", targetOrder.Symbol, targetOrder.Side, targetOrder.Size)))
	return nil
}
