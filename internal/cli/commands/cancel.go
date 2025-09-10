package commands

import (
	"fmt"
	"strconv"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/repository"
	"github.com/lemconn/foxflow/internal/utils"
)

// CancelCommand 取消命令
type CancelCommand struct{}

func (c *CancelCommand) GetName() string        { return "cancel" }
func (c *CancelCommand) GetDescription() string { return "取消策略订单" }
func (c *CancelCommand) GetUsage() string       { return "cancel ss <id>" }

func (c *CancelCommand) Execute(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}
	if args[0] != "ss" {
		return fmt.Errorf("only support cancel ss")
	}

	orderID, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid order ID")
	}

	order, err := repository.FindSSOrderByIDForUser(ctx.GetUserInstance().ID, orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// 如果订单已提交到交易所，需要取消远程订单
	if order.Status == "pending" && order.OrderID != "" {
		exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeInstance().Name)
		if err != nil {
			return fmt.Errorf("failed to get exchange client: %w", err)
		}
		if err := exchangeClient.CancelOrder(ctx.GetContext(), order.OrderID); err != nil {
			return fmt.Errorf("failed to cancel remote order: %w", err)
		}
	}

	// 更新订单状态
	order.Status = "cancelled"
	if err := repository.SaveSSOrder(order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("订单已取消: ID=%d", order.ID)))
	return nil
}
