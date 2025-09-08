package commands

import (
	"fmt"
	"strconv"

	"foxflow/internal/cli/command"
	"foxflow/pkg/utils"
)

// UpdateCommand 设置命令
type UpdateCommand struct{}

func (c *UpdateCommand) GetName() string        { return "update" }
func (c *UpdateCommand) GetDescription() string { return "设置杠杆或保证金模式" }
func (c *UpdateCommand) GetUsage() string       { return "update <type> <value>" }

func (c *UpdateCommand) Execute(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "leverage":
		leverage, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid leverage value")
		}
		// 这里应该调用交易所API设置杠杆
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("杠杆设置为: %d", leverage)))
	case "margin-type":
		marginType := args[1]
		// 这里应该调用交易所API设置保证金模式
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("保证金模式设置为: %s", marginType)))
	default:
		return fmt.Errorf("unknown update type: %s", args[0])
	}

	return nil
}
