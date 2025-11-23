package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"

	"github.com/lemconn/foxflow/internal/cli/command"
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

	account := &model.FoxAccount{}

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

	exchangeName := ctx.GetExchangeName()
	if exchangeName == "" {
		exchangeName = config.DefaultExchange
	}

	if grpcClient := ctx.GetGRPCClient(); grpcClient != nil {
		fmt.Println("正在通过 gRPC 创建账户...")
		accountItem, err := grpcClient.CreateAccount(
			exchangeName,
			account.TradeType,
			account.Name,
			account.AccessKey,
			account.SecretKey,
			account.Passphrase,
		)
		if err == nil {
			useCommand := UseCommand{}
			if err := useCommand.HandleAccountCommand(ctx, accountItem.Name); err != nil {
				return fmt.Errorf("create account & use error: %w", err)
			}
			fmt.Println("账户创建成功并已激活")
			return nil
		}
		fmt.Println("gRPC 创建账户失败，回退到本地模式:", err)
	}

	return c.createAccountLocal(ctx, account)
}

func (c *CreateCommand) createAccountLocal(ctx command.Context, account *model.FoxAccount) error {
	accountInfo, err := repository.FindAccountByName(account.Name)
	if err != nil {
		return fmt.Errorf("find username err: %w", err)
	}

	if accountInfo != nil && accountInfo.ID > 0 {
		return fmt.Errorf("account already exists, name: %s", accountInfo.Name)
	}

	account.Exchange = ctx.GetExchangeName()
	if account.Exchange == "" {
		account.Exchange = config.DefaultExchange
	}

	exchangeInfo, err := repository.GetExchange(account.Exchange)
	if err != nil {
		return fmt.Errorf("get exchange error: %w", err)
	}
	if exchangeInfo == nil || exchangeInfo.Name == "" {
		return fmt.Errorf("exchange is not found")
	}

	if exchangeInfo.Name == config.DefaultExchange && account.Passphrase == "" {
		return fmt.Errorf("okx exchange passphrase is required")
	}

	exchangeClient, err := exchange.GetManager().GetExchange(exchangeInfo.Name)
	if err != nil {
		return fmt.Errorf("get exchange client error: %w", err)
	}
	if err := exchangeClient.Connect(ctx.GetContext(), account); err != nil {
		return fmt.Errorf("connect exchange error: %w", err)
	}

	accountConfig, err := exchangeClient.GetAccountConfig(ctx.GetContext())
	if err != nil {
		return fmt.Errorf("get account config err: %w", err)
	}
	if accountConfig == nil {
		return fmt.Errorf("account config is nil")
	}

	if accountConfig.AccountMode <= 1 {
		return fmt.Errorf("当前账户不支持合约交易，请前往“交易设置 > 账户模式”进行切换")
	}

	if !strings.Contains(strings.ToLower(accountConfig.Permission), "trade") {
		return fmt.Errorf("当前账户不支持交易，请重新生成API key，生成时请注意权限选择“交易”")
	}

	if accountConfig.PositionMode != "long_short_mode" {
		if err := exchangeClient.SetPositionMode(ctx.GetContext(), "long_short_mode"); err != nil {
			return fmt.Errorf("账户双向仓位切换失败，请前往“交易设置 > 仓位模式”进行手动切换，异常信息: %w", err)
		}
	}

	if err := repository.CreateAccount(account); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	useCommand := UseCommand{}
	if err := useCommand.HandleAccountCommand(ctx, account.Name); err != nil {
		return fmt.Errorf("create account & use error: %w", err)
	}

	return nil
}
