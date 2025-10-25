package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/exchange"

	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/models"
	"github.com/lemconn/foxflow/internal/repository"
)

// CreateCommand 创建命令
type CreateCommand struct{}

func (c *CreateCommand) GetName() string        { return "create" }
func (c *CreateCommand) GetDescription() string { return "创建用户、标的或策略订单" }
func (c *CreateCommand) GetUsage() string       { return "create <type> [options]" }

func (c *CreateCommand) Execute(ctx command.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	switch args[0] {
	case "account":
		return c.createAccount(ctx, args[1:])
	default:
		return fmt.Errorf("unknown create type: %s", args[0])
	}
}

func (c *CreateCommand) createAccount(ctx command.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: create account <trade_type> name=<name> apiKey=<key> secretKey=<secret> [passphrase=<passphrase>]")
	}

	account := &models.FoxAccount{}

	account.TradeType = args[0]
	for _, arg := range args {
		if strings.HasPrefix(arg, "name=") {
			account.Name = strings.TrimPrefix(arg, "name=")
		} else if strings.HasPrefix(arg, "apiKey=") {
			account.AccessKey = strings.TrimPrefix(arg, "apiKey=")
		} else if strings.HasPrefix(arg, "secretKey=") {
			account.SecretKey = strings.TrimPrefix(arg, "secretKey=")
		} else if strings.HasPrefix(arg, "passphrase=") {
			account.Passphrase = strings.TrimPrefix(arg, "passphrase=")
		}
	}

	if account.TradeType == "" || account.Name == "" || account.AccessKey == "" || account.SecretKey == "" {
		return fmt.Errorf("missing required parameters")
	}

	// 根据用户名获取用户信息
	accountInfo, err := repository.FindAccountByName(account.Name)
	if err != nil {
		return fmt.Errorf("find username err: %w", err)
	}

	// 用户存在则不允许创建
	if accountInfo != nil && accountInfo.ID > 0 {
		return fmt.Errorf("account already exists, name: %s", accountInfo.Name)
	}

	account.Exchange = ctx.GetExchangeName()
	if account.Exchange == "" {
		account.Exchange = config.DefaultExchange // 默认交易所
	}

	exchangeInfo, err := repository.GetExchange(account.Exchange)
	if err != nil {
		return fmt.Errorf("get exchange error: %w", err)
	}

	if exchangeInfo.Name == "" {
		return fmt.Errorf("exchange is not found")
	}

	if exchangeInfo.Name == config.DefaultExchange && account.Passphrase == "" {
		return fmt.Errorf("okx exchange passphrase is required")
	}

	// 到指定交易交易所验证当前用户
	exchangeClient, err := exchange.GetManager().GetExchange(exchangeInfo.Name)
	if err != nil {
		return fmt.Errorf("get exchange client error: %w", err)
	}
	err = exchangeClient.Connect(ctx.GetContext(), account)
	if err != nil {
		return fmt.Errorf("connect exchange error: %w", err)
	}

	if err = repository.CreateAccount(account); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	useCommand := UseCommand{}
	err = useCommand.HandleAccountCommand(ctx, account.Name)
	if err != nil {
		return fmt.Errorf("create account & use error: %w", err)
	}

	return nil
}
