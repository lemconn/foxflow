package server

import (
	"context"
	"fmt"
	"log"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"github.com/lemconn/foxflow/internal/repository"
	pb "github.com/lemconn/foxflow/proto/generated"
)

type AccountServer struct{}

func NewAccountServer() *AccountServer {
	return &AccountServer{}
}

func (s *AccountServer) GetAccounts(ctx context.Context, req *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	// 从数据库获取账户列表
	accounts, err := database.Adapter().FoxAccount.Find()
	if err != nil {
		log.Printf("获取账户列表失败: %v", err)
		return &pb.GetAccountsResponse{
			Success: false,
			Message: fmt.Sprintf("获取账户列表失败: %v", err),
		}, nil
	}

	// 转换为 protobuf 格式
	var pbAccounts []*pb.AccountsItem
	for _, account := range accounts {
		pbAccounts = append(pbAccounts, buildPBAccountItem(account))
	}

	return &pb.GetAccountsResponse{
		Success:  true,
		Message:  fmt.Sprintf("成功获取 %d 个账户", len(pbAccounts)),
		Accounts: pbAccounts,
	}, nil
}

// UseAccount 激活账户
func (s *AccountServer) UseAccount(ctx context.Context, req *pb.UseAccountRequest) (*pb.UseAccountResponse, error) {
	if req.Account == "" {
		return &pb.UseAccountResponse{
			Success: false,
			Message: "account 是必填参数",
		}, nil
	}

	account, err := repository.FindAccountByName(req.Account)
	if err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("获取账户失败: %v", err),
		}, nil
	}
	if account == nil || account.ID == 0 {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("account `%s` not found", req.Account),
		}, nil
	}

	if err := repository.SetAllAccountInactive(); err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("重置账户状态失败: %v", err),
		}, nil
	}

	if err := repository.SetAllExchangesInactive(); err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("重置交易所状态失败: %v", err),
		}, nil
	}

	ex, err := repository.GetExchange(account.Exchange)
	if err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所失败: %v", err),
		}, nil
	}
	if ex == nil || ex.ID == 0 {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("exchange `%s` not found", account.Exchange),
		}, nil
	}

	if err := repository.ActivateExchange(account.Exchange); err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("激活交易所失败: %v", err),
		}, nil
	}

	if err := repository.ActivateAccountByName(account.Name); err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("激活账户失败: %v", err),
		}, nil
	}

	if exchangeClient, err := exchange.GetManager().GetExchange(account.Exchange); err == nil {
		if err := exchangeClient.SetAccount(ctx, account); err != nil {
			log.Printf("设置账户失败: %v", err)
		}
	} else {
		log.Printf("获取交易所客户端失败: %v", err)
	}

	return &pb.UseAccountResponse{
		Success: true,
		Message: fmt.Sprintf("已激活用户: %s", account.Name),
		Account: buildPBAccountItem(account),
		Exchange: &pb.ExchangesItem{
			Name:        ex.Name,
			ApiUrl:      ex.APIURL,
			ProxyUrl:    ex.ProxyURL,
			StatusValue: ex.IsActive,
		},
	}, nil
}

func buildPBAccountItem(account *model.FoxAccount) *pb.AccountsItem {
	// 处理杠杆倍数
	var crossLeverage, isolatedLeverage int64
	for _, tradeConfig := range account.TradeConfigs {
		if tradeConfig.Margin == "cross" {
			crossLeverage = tradeConfig.Leverage
		} else if tradeConfig.Margin == "isolated" {
			isolatedLeverage = tradeConfig.Leverage
		}
	}

	// 处理代理地址
	proxyURL := ""
	if account.Config.ProxyURL != "" {
		proxyURL = account.Config.ProxyURL
	}

	return &pb.AccountsItem{
		Id:               account.ID,
		Name:             account.Name,
		Exchange:         account.Exchange,
		TradeTypeValue:   account.TradeType,
		StatusValue:      account.IsActive,
		CrossLeverage:    crossLeverage,
		IsolatedLeverage: isolatedLeverage,
		ProxyUrl:         proxyURL,
	}
}
