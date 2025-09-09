package commands

import (
	"fmt"

	"github.com/lemconn/foxflow/internal/cli/command"
)

// ExitCommand 退出命令
type ExitCommand struct{}

func (c *ExitCommand) GetName() string        { return "exit" }
func (c *ExitCommand) GetDescription() string { return "退出程序" }
func (c *ExitCommand) GetUsage() string       { return "exit" }

func (c *ExitCommand) Execute(_ command.Context, _ []string) error {
	// 为了保持与旧实现一致，这里直接返回特殊错误 "exit"
	return fmt.Errorf("exit")
}
