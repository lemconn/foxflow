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

func (c *CloseCommand) Execute(_ command.Context, _ []string) error {
	// 为了保持与旧实现一致，这里直接返回特殊错误 "exit"
	return fmt.Errorf("exit")
}
