package commands

import (
	"fmt"

	"foxflow/internal/cli/command"
)

// HelpCommand 帮助命令
type HelpCommand struct{}

func (c *HelpCommand) GetName() string        { return "help" }
func (c *HelpCommand) GetDescription() string { return "显示帮助信息" }
func (c *HelpCommand) GetUsage() string       { return "help [command]" }

func (c *HelpCommand) Execute(ctx command.Context, args []string) error {
	commands := []command.Command{
		&ShowCommand{},
		&UseCommand{},
		&CreateCommand{},
		&UpdateCommand{},
		&CancelCommand{},
		&DeleteCommand{},
		&HelpCommand{},
	}

	if len(args) > 0 {
		// 显示特定命令的帮助
		for _, cmd := range commands {
			if cmd.GetName() == args[0] {
				fmt.Printf("命令: %s\n", cmd.GetName())
				fmt.Printf("描述: %s\n", cmd.GetDescription())
				fmt.Printf("用法: %s\n", cmd.GetUsage())
				return nil
			}
		}
		return fmt.Errorf("command not found: %s", args[0])
	}

	// 显示所有命令
	fmt.Println("可用命令:")
	for _, cmd := range commands {
		fmt.Printf("  %-10s - %s\n", cmd.GetName(), cmd.GetDescription())
	}

	return nil
}
