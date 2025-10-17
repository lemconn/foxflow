package commands

import (
	"fmt"

	"github.com/lemconn/foxflow/internal/cli/command"
)

// OpenCommand 退出命令
type OpenCommand struct{}

func (c *OpenCommand) GetName() string        { return "open" }
func (c *OpenCommand) GetDescription() string { return "开仓/下单" }
func (c *OpenCommand) GetUsage() string       { return "open <symbol> <direction> <amount> [with]" }

func (c *OpenCommand) Execute(_ command.Context, _ []string) error {
	// 为了保持与旧实现一致，这里直接返回特殊错误 "exit"
	return fmt.Errorf("exit")
}
