package commands

import (
	"fmt"

	"github.com/lemconn/foxflow/internal/cli/command"
)

// CloseCommand 退出命令
type CloseCommand struct{}

func (c *CloseCommand) GetName() string        { return "close" }
func (c *CloseCommand) GetDescription() string { return "平仓" }
func (c *CloseCommand) GetUsage() string       { return "close <symbol> <direction> [amount] [with]" }

func (c *CloseCommand) Execute(ctx command.Context, args []string) error {
	if !ctx.IsReady() {
		return fmt.Errorf("请先选择交易所和用户")
	}

	fmt.Printf("----------[%+v]-------\n", args)

	// 为了保持与旧实现一致，这里直接返回特殊错误 "exit"
	return fmt.Errorf("exit")
}
