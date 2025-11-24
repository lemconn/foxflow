package commands

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"

	"github.com/lemconn/foxflow/internal/cli/command"
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

	grpcClient := ctx.GetGRPCClient()
	if grpcClient == nil {
		return fmt.Errorf("gRPC 客户端初始化异常")
	}

	accountItem, err := grpcClient.CreateAccount(
		exchangeName,
		account.TradeType,
		account.Name,
		account.AccessKey,
		account.SecretKey,
		account.Passphrase,
	)
	if err != nil {
		return fmt.Errorf("创建账户失败: %w", err)
	}

	useCommand := UseCommand{}
	if err := useCommand.HandleAccountCommand(ctx, accountItem.Name); err != nil {
		return fmt.Errorf("create account & use error: %w", err)
	}
	
	fmt.Println("账户创建成功并已激活")
	return nil
}
