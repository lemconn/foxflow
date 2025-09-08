package commands

import (
	"fmt"

	"foxflow/internal/cli/command"
	"foxflow/internal/exchange"
	"foxflow/internal/repository"
	"foxflow/pkg/utils"
)

// UseCommand 激活命令
type UseCommand struct{}

func (c *UseCommand) GetName() string        { return "use" }
func (c *UseCommand) GetDescription() string { return "激活交易所或用户" }
func (c *UseCommand) GetUsage() string {
	return `
Usage: use <type> <name>

Description:
  Can activate exchanges or users

Types:
  exchanges    - activate designated exchange
  users        - activate a specified user

Name：
  - If the type is 'exchanges', Currently supported options include 'okx', 'binance', and 'gate'.
  - If the type is 'users', the user needs to specify the user name.
`
}

func (c *UseCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("%s", c.GetUsage())
	}

	switch args[0] {
	case "exchanges":
		exchangeName := args[1]

		// 将所有交易所设置为非激活状态
		if err := repository.SetAllExchangesInactive(); err != nil {
			return fmt.Errorf("failed to deactivate exchanges: %w", err)
		}

		// 将所有用户设置为非激活状态
		if err := repository.SetAllUsersInactive(); err != nil {
			return fmt.Errorf("failed to deactivate users: %w", err)
		}

		// 断开当前交易所连接
		if ctx.GetExchange() != "" {
			exchange.GetManager().DisconnectUser(ctx.GetExchange())
		}

		// 激活选中的交易所
		if err := repository.ActivateExchange(exchangeName); err != nil {
			return fmt.Errorf("failed to activate exchange: %w", err)
		}

		// 设置新的交易所
		ex, err := exchange.GetManager().GetExchange(exchangeName)
		if err != nil {
			return fmt.Errorf("exchange not found: %s", exchangeName)
		}
		ctx.SetExchange(exchangeName)
		ctx.SetExchangeInstance(ex)
		ctx.SetUser(nil) // 清除当前用户
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("已激活交易所: %s", exchangeName)))

	case "users":
		username := args[1]
		user, err := repository.FindUserByUsername(username)
		if err != nil {
			return fmt.Errorf("user not found: %s", username)
		}

		// 将所有用户设置为非激活状态
		if err := repository.SetAllUsersInactive(); err != nil {
			return fmt.Errorf("failed to deactivate users: %w", err)
		}

		// 如果用户属于不同的交易所，需要切换交易所
		if ctx.GetExchange() != "" && ctx.GetExchange() != user.Exchange {
			// 将所有交易所设置为非激活状态
			if err := repository.SetAllExchangesInactive(); err != nil {
				return fmt.Errorf("failed to deactivate exchanges: %w", err)
			}

			// 断开当前交易所连接
			exchange.GetManager().DisconnectUser(ctx.GetExchange())

			// 切换到用户所属的交易所
			ex, err := exchange.GetManager().GetExchange(user.Exchange)
			if err != nil {
				return fmt.Errorf("failed to get exchange: %w", err)
			}

			// 激活用户所属的交易所
			if err := repository.ActivateExchange(user.Exchange); err != nil {
				return fmt.Errorf("failed to activate exchange: %w", err)
			}

			ctx.SetExchange(user.Exchange)
			ctx.SetExchangeInstance(ex)
		}

		// 激活选中的用户
		if err := repository.ActivateUserByUsername(username); err != nil {
			return fmt.Errorf("failed to activate user: %w", err)
		}

		// 连接用户到交易所
		if ctx.GetExchange() != "" {
			if err := exchange.GetManager().ConnectUser(ctx.GetContext(), ctx.GetExchange(), user); err != nil {
				return fmt.Errorf("failed to connect user: %w", err)
			}
		}

		ctx.SetUser(user)
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("已激活用户: %s", username)))

	default:
		return fmt.Errorf("unknown use type: %s", args[0])
	}

	return nil
}
