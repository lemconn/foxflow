package server

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/exchange"
	pb "github.com/lemconn/foxflow/proto/generated"
	"gorm.io/gorm"
)

type BalanceServer struct{}

func NewBalanceServer() *BalanceServer {
	return &BalanceServer{}
}

// GetBalance 获取资产列表
func (s *BalanceServer) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	if req.AccountId <= 0 {
		return &pb.GetBalanceResponse{
			Success: false,
			Message: "account_id 是必填参数，且必须大于 0",
		}, nil
	}

	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.ID.Eq(req.AccountId),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &pb.GetBalanceResponse{
				Success: false,
				Message: fmt.Sprintf("账户 %d 不存在", req.AccountId),
			}, nil
		}
		log.Printf("查询账户失败: %v", err)
		return &pb.GetBalanceResponse{
			Success: false,
			Message: fmt.Sprintf("查询账户失败: %v", err),
		}, nil
	}

	exchangeClient, err := exchange.GetManager().GetExchange(account.Exchange)
	if err != nil {
		log.Printf("获取交易所客户端失败: %v", err)
		return &pb.GetBalanceResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所客户端失败: %v", err),
		}, nil
	}

	if err := exchangeClient.SetAccount(ctx, account); err != nil {
		log.Printf("设置账户失败: %v", err)
		return &pb.GetBalanceResponse{
			Success: false,
			Message: fmt.Sprintf("设置账户失败: %v", err),
		}, nil
	}

	assets, err := exchangeClient.GetBalance(ctx)
	if err != nil {
		log.Printf("获取资产失败: %v", err)
		return &pb.GetBalanceResponse{
			Success: false,
			Message: fmt.Sprintf("获取资产失败: %v", err),
		}, nil
	}

	var pbAssets []*pb.BalanceItem
	for _, asset := range assets {
		pbAssets = append(pbAssets, &pb.BalanceItem{
			Currency:  asset.Currency,
			Balance:   asset.Balance,
			Available: asset.Available,
			Frozen:    asset.Frozen,
		})
	}

	return &pb.GetBalanceResponse{
		Success: true,
		Message: fmt.Sprintf("成功获取 %d 个资产", len(pbAssets)),
		Assets:  pbAssets,
	}, nil
}
