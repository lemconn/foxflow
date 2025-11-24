package commands

import (
	"fmt"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/utils"
)

// DeleteCommand 删除命令
type DeleteCommand struct{}

func (c *DeleteCommand) GetName() string        { return "delete" }
func (c *DeleteCommand) GetDescription() string { return "删除用户" }
func (c *DeleteCommand) GetUsage() string       { return "delete <type> <name>" }

func (c *DeleteCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "account":
		return c.handleAccountCommand(ctx, args[1])
	default:
		return fmt.Errorf("unknown delete type: %s", args[0])
	}
}

func (c *DeleteCommand) handleAccountCommand(ctx command.Context, name string) error {
	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	err := grpcClient.DeleteAccount(name)
	if err != nil {
		return fmt.Errorf("删除账户失败: %v", err)
	}

	fmt.Println(utils.RenderSuccess(fmt.Sprintf("用户已删除: %s", name)))
	return nil
}
