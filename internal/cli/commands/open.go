package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
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

	// 解析参数
	order := &models.FoxSS{
		UserID:    ctx.GetUserInstance().ID,
		Side:      "buy",
		OrderType: "limit",
		Type:      "open",
		Status:    "waiting",
	}

	order.Symbol = strings.ToUpper(args[0])
	order.PosSide = strings.ToLower(args[1])
	amount := strings.ToUpper(args[2])
	if amount != "" {
		// 校验是否存在U的后缀

	}

	if order.Symbol == "" || order.Sz == 0 {
		return fmt.Errorf("missing required parameters: symbol, side, size")
	}
	
	if err := repository.CreateSSOrder(order); err != nil {
		return fmt.Errorf("failed to create strategy order: %w", err)
	}

	fmt.Println(utils.RenderInfo("策略订单已创建，等待策略条件满足"))

	return nil
}
