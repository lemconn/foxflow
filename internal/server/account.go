package server

import (
	"context"
	"fmt"
	"log"

	"github.com/lemconn/foxflow/internal/database"
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

		pbAccounts = append(pbAccounts, &pb.AccountsItem{
			Name:             account.Name,
			Exchange:         account.Exchange,
			TradeTypeValue:   account.TradeType,
			StatusValue:      account.IsActive,
			CrossLeverage:    crossLeverage,
			IsolatedLeverage: isolatedLeverage,
			ProxyUrl:         proxyURL,
		})
	}

	return &pb.GetAccountsResponse{
		Success:  true,
		Message:  fmt.Sprintf("成功获取 %d 个账户", len(pbAccounts)),
		Accounts: pbAccounts,
	}, nil
}
