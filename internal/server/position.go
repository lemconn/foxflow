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

type PositionServer struct{}

func NewPositionServer() *PositionServer {
	return &PositionServer{}
}

// GetPositions 获取仓位列表
func (s *PositionServer) GetPositions(ctx context.Context, req *pb.GetPositionsRequest) (*pb.GetPositionsResponse, error) {
	if req.AccountId <= 0 {
		return &pb.GetPositionsResponse{
			Success: false,
			Message: "account_id 是必填参数，且必须大于 0",
		}, nil
	}

	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.ID.Eq(req.AccountId),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &pb.GetPositionsResponse{
				Success: false,
				Message: fmt.Sprintf("账户 %d 不存在", req.AccountId),
			}, nil
		}
		log.Printf("查询账户失败: %v", err)
		return &pb.GetPositionsResponse{
			Success: false,
			Message: fmt.Sprintf("查询账户失败: %v", err),
		}, nil
	}

	exchangeClient, err := exchange.GetManager().GetExchange(account.Exchange)
	if err != nil {
		log.Printf("获取交易所客户端失败: %v", err)
		return &pb.GetPositionsResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所客户端失败: %v", err),
		}, nil
	}

	if err := exchangeClient.SetAccount(ctx, account); err != nil {
		log.Printf("设置账户失败: %v", err)
		return &pb.GetPositionsResponse{
			Success: false,
			Message: fmt.Sprintf("设置账户失败: %v", err),
		}, nil
	}

	positions, err := exchangeClient.GetPositions(ctx)
	if err != nil {
		log.Printf("获取仓位失败: %v", err)
		return &pb.GetPositionsResponse{
			Success: false,
			Message: fmt.Sprintf("获取仓位失败: %v", err),
		}, nil
	}

	var pbPositions []*pb.PositionItem
	for _, pos := range positions {
		pbPositions = append(pbPositions, &pb.PositionItem{
			Symbol:     pos.Symbol,
			PosSide:    pos.PosSide,
			MarginType: pos.MarginType,
			Size:       pos.Size,
			AvgPrice:   pos.AvgPrice,
			UnrealPnl:  pos.UnrealPnl,
		})
	}

	return &pb.GetPositionsResponse{
		Success:   true,
		Message:   fmt.Sprintf("成功获取 %d 个仓位", len(pbPositions)),
		Positions: pbPositions,
	}, nil
}
