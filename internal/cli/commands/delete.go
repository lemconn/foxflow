package commands

import (
	"fmt"

	"foxflow/internal/cli/command"
	"foxflow/internal/repository"
	"foxflow/pkg/utils"
)

// DeleteCommand 删除命令
type DeleteCommand struct{}

func (c *DeleteCommand) GetName() string        { return "delete" }
func (c *DeleteCommand) GetDescription() string { return "删除用户或标的" }
func (c *DeleteCommand) GetUsage() string       { return "delete <type> <name>" }

func (c *DeleteCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "users":
		username := args[1]
		if err := repository.DeleteUserByUsername(username); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("用户已删除: %s", username)))
	case "symbols":
		if !ctx.IsReady() {
			return fmt.Errorf("请先选择交易所和用户")
		}
		symbolName := args[1]
		if err := repository.DeleteSymbolByNameForUser(ctx.GetUser().ID, symbolName); err != nil {
			return fmt.Errorf("failed to delete symbol: %w", err)
		}
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("标的已删除: %s", symbolName)))
	default:
		return fmt.Errorf("unknown delete type: %s", args[0])
	}

	return nil
}
