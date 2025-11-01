package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"github.com/lemconn/foxflow/internal/repository"
	"github.com/lemconn/foxflow/internal/utils"
)

// UseCommand 激活命令
type UseCommand struct{}

func (c *UseCommand) GetName() string        { return "use" }
func (c *UseCommand) GetDescription() string { return "激活交易所或用户" }
func (c *UseCommand) GetUsage() string {
	return `use <type> <name> \n  types: exchange（激活交易所）, account（激活交易账户）\n `
}

func (c *UseCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("%s", c.GetUsage())
	}

	switch strings.ToLower(args[0]) {
	case "exchange":
		return c.HandleExchangeCommand(ctx, args[1])
	case "account":
		return c.HandleAccountCommand(ctx, args[1])
	default:
		return fmt.Errorf("unknown use type: %s", args[0])
	}
}

func (c *UseCommand) HandleExchangeCommand(ctx command.Context, exchangeName string) error {
	// 将所有交易所设置为非激活状态
	if err := repository.SetAllExchangesInactive(); err != nil {
		return fmt.Errorf("failed to deactivate exchanges: %w", err)
	}

	// 将所有用户设置为非激活状态
	if err := repository.SetAllAccountInactive(); err != nil {
		return fmt.Errorf("failed to deactivate accounts: %w", err)
	}

	// 断开当前交易所连接
	if ctx.GetExchangeName() != "" {
		exchange.GetManager().DisconnectAccount(ctx.GetExchangeName())
	}

	// 获取交易所信息
	exchangeInfo, err := repository.GetExchange(exchangeName)
	if err != nil {
		return fmt.Errorf("failed to get exchange %s: %w", exchangeName, err)
	}
	if exchangeInfo == nil || exchangeInfo.ID == 0 {
		return fmt.Errorf("exchange `%s` not found", exchangeName)
	}

	// 激活指定交易所
	if err = repository.ActivateExchange(exchangeName); err != nil {
		return fmt.Errorf("failed to activate exchange: %w", err)
	}

	// 设置新的交易所
	ctx.SetExchangeName(exchangeName)
	ctx.SetExchangeInstance(exchangeInfo)
	ctx.SetAccountName("")                      // 清除当前用户
	ctx.SetAccountInstance(&model.FoxAccount{}) // 清楚当前用户信息
	fmt.Println(utils.RenderSuccess(fmt.Sprintf("已激活交易所: %s", utils.MessageGreen(exchangeName))))

	return nil
}

func (c *UseCommand) HandleAccountCommand(ctx command.Context, name string) error {
	account, err := repository.FindAccountByName(name)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	if account == nil || account.ID == 0 {
		return fmt.Errorf("account `%s` not found", name)
	}

	// 将所有用户设置为非激活状态
	if err = repository.SetAllAccountInactive(); err != nil {
		return fmt.Errorf("failed to deactivate accounts: %w", err)
	}

	// 将所有交易所设置为非激活状态
	if err = repository.SetAllExchangesInactive(); err != nil {
		return fmt.Errorf("failed to deactivate exchanges: %w", err)
	}

	// 切换到用户所属的交易所
	ex, err := repository.GetExchange(account.Exchange)
	if err != nil {
		return fmt.Errorf("failed to get exchange: %w", err)
	}

	if ex == nil || ex.ID == 0 {
		return fmt.Errorf("exchange `%s` not found", account.Exchange)
	}

	// 激活用户所属的交易所
	if err = repository.ActivateExchange(account.Exchange); err != nil {
		return fmt.Errorf("failed to activate exchange: %w", err)
	}

	ctx.SetExchangeName(account.Exchange)
	ctx.SetExchangeInstance(ex)

	// 激活选中的用户
	if err = repository.ActivateAccountByName(name); err != nil {
		return fmt.Errorf("failed to activate account: %w", err)
	}

	// 连接用户到交易所
	if ctx.GetExchangeName() != "" {
		exchangeClient, err := exchange.GetManager().GetExchange(ctx.GetExchangeName())
		if err != nil {
			return fmt.Errorf("failed to get exchange client: %w", err)
		}

		if err = exchangeClient.SetAccount(ctx.GetContext(), account); err != nil {
			return fmt.Errorf("failed to connect account: %w", err)
		}
	}

	ctx.SetAccountName(account.Name)
	ctx.SetAccountInstance(account)
	fmt.Println(utils.RenderSuccess(fmt.Sprintf("已激活用户: %s", utils.MessageGreen(name))))

	return nil
}
