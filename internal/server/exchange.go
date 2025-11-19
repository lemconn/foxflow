package server

import (
	"context"
	"fmt"
	"log"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/repository"
	pb "github.com/lemconn/foxflow/proto/generated"
)

type ExchangeServer struct{}

func NewExchangeServer() *ExchangeServer {
	return &ExchangeServer{}
}

// GetExchanges 获取交易所列表
func (s *ExchangeServer) GetExchanges(ctx context.Context, req *pb.GetExchangesRequest) (*pb.GetExchangesResponse, error) {
	exchanges, err := database.Adapter().FoxExchange.Find()
	if err != nil {
		log.Printf("获取交易所列表失败: %v", err)
		return &pb.GetExchangesResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所列表失败: %v", err),
		}, nil
	}

	var pbExchanges []*pb.ExchangesItem
	for _, exchange := range exchanges {
		pbExchanges = append(pbExchanges, &pb.ExchangesItem{
			Name:        exchange.Name,
			ApiUrl:      exchange.APIURL,
			ProxyUrl:    exchange.ProxyURL,
			StatusValue: exchange.IsActive,
		})
	}

	return &pb.GetExchangesResponse{
		Success:   true,
		Message:   fmt.Sprintf("成功获取 %d 个交易所", len(pbExchanges)),
		Exchanges: pbExchanges,
	}, nil
}

// UseExchange 激活交易所
func (s *ExchangeServer) UseExchange(ctx context.Context, req *pb.UseExchangeRequest) (*pb.UseExchangeResponse, error) {
	if req.Exchange == "" {
		return &pb.UseExchangeResponse{
			Success: false,
			Message: "exchange 是必填参数",
		}, nil
	}

	// 将所有交易所和账户置为未激活状态
	if err := repository.SetAllExchangesInactive(); err != nil {
		log.Printf("重置交易所状态失败: %v", err)
		return &pb.UseExchangeResponse{
			Success: false,
			Message: fmt.Sprintf("重置交易所状态失败: %v", err),
		}, nil
	}
	if err := repository.SetAllAccountInactive(); err != nil {
		log.Printf("重置账户状态失败: %v", err)
		return &pb.UseExchangeResponse{
			Success: false,
			Message: fmt.Sprintf("重置账户状态失败: %v", err),
		}, nil
	}

	// 断开当前交易所连接
	exchange.GetManager().DisconnectAccount(req.Exchange)

	// 获取并激活指定交易所
	exchangeInfo, err := repository.GetExchange(req.Exchange)
	if err != nil {
		log.Printf("获取交易所 %s 失败: %v", req.Exchange, err)
		return &pb.UseExchangeResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所失败: %v", err),
		}, nil
	}
	if exchangeInfo == nil || exchangeInfo.ID == 0 {
		return &pb.UseExchangeResponse{
			Success: false,
			Message: fmt.Sprintf("exchange `%s` not found", req.Exchange),
		}, nil
	}

	if err := repository.ActivateExchange(req.Exchange); err != nil {
		log.Printf("激活交易所 %s 失败: %v", req.Exchange, err)
		return &pb.UseExchangeResponse{
			Success: false,
			Message: fmt.Sprintf("激活交易所失败: %v", err),
		}, nil
	}

	return &pb.UseExchangeResponse{
		Success: true,
		Message: fmt.Sprintf("已激活交易所: %s", req.Exchange),
		Exchange: &pb.ExchangesItem{
			Name:        exchangeInfo.Name,
			ApiUrl:      exchangeInfo.APIURL,
			ProxyUrl:    exchangeInfo.ProxyURL,
			StatusValue: 1,
		},
	}, nil
}
