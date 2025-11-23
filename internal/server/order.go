package server

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/engine/syntax"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	pb "github.com/lemconn/foxflow/proto/generated"
	"github.com/shopspring/decimal"
)

type OrderServer struct{}

func NewOrderServer() *OrderServer {
	return &OrderServer{}
}

// GetOrders 获取订单列表
func (s *OrderServer) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersResponse, error) {
	// 验证必填字段：account_id
	if req.AccountId <= 0 {
		log.Printf("account_id 参数无效: %d", req.AccountId)
		return &pb.GetOrdersResponse{
			Success: false,
			Message: "account_id 是必填参数，且必须大于 0",
		}, nil
	}

	// 构建查询
	tx := database.Adapter().FoxOrder.Where(database.Adapter().FoxOrder.AccountID.Eq(req.AccountId))

	// 如果指定了状态过滤，则应用状态过滤
	if len(req.Status) > 0 {
		tx = tx.Where(database.Adapter().FoxOrder.Status.In(req.Status...))
	}

	// 按创建时间倒序排列，并限制数量
	tx = tx.Order(database.Adapter().FoxOrder.CreatedAt.Desc())

	// 执行查询
	orders, err := tx.Find()
	if err != nil {
		log.Printf("获取订单列表失败: %v", err)
		return &pb.GetOrdersResponse{
			Success: false,
			Message: fmt.Sprintf("获取订单列表失败: %v", err),
		}, nil
	}

	// 转换为 protobuf 格式
	var pbOrders []*pb.OrderItem
	for _, order := range orders {
		pbOrders = append(pbOrders, &pb.OrderItem{
			Id:         order.ID,
			Exchange:   order.Exchange,
			AccountId:  order.AccountID,
			Symbol:     order.Symbol,
			Side:       order.Side,
			PosSide:    order.PosSide,
			MarginType: order.MarginType,
			Price:      order.Price,
			Size:       order.Size,
			SizeType:   order.SizeType,
			OrderType:  order.OrderType,
			Strategy:   order.Strategy,
			OrderId:    order.OrderID,
			Type:       order.Type,
			Status:     order.Status,
			Msg:        order.Msg,
			CreatedAt:  order.CreatedAt.Unix(),
			UpdatedAt:  order.UpdatedAt.Unix(),
		})
	}

	return &pb.GetOrdersResponse{
		Success: true,
		Message: fmt.Sprintf("成功获取 %d 个订单", len(pbOrders)),
		Orders:  pbOrders,
	}, nil
}

// OpenOrder 创建开仓订单
func (s *OrderServer) OpenOrder(ctx context.Context, req *pb.OpenOrderRequest) (*pb.OpenOrderResponse, error) {
	if req.AccountId <= 0 {
		return &pb.OpenOrderResponse{Success: false, Message: "account_id 是必填参数"}, nil
	}
	if req.Exchange == "" {
		return &pb.OpenOrderResponse{Success: false, Message: "exchange 是必填参数"}, nil
	}
	if req.Symbol == "" {
		return &pb.OpenOrderResponse{Success: false, Message: "symbol 是必填参数"}, nil
	}
	if req.PosSide != "long" && req.PosSide != "short" {
		return &pb.OpenOrderResponse{Success: false, Message: "pos_side 只能为 long 或 short"}, nil
	}
	if req.Margin != "isolated" && req.Margin != "cross" {
		return &pb.OpenOrderResponse{Success: false, Message: "margin 只能为 isolated 或 cross"}, nil
	}
	if req.Amount == "" {
		return &pb.OpenOrderResponse{Success: false, Message: "amount 是必填参数"}, nil
	}

	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.ID.Eq(req.AccountId),
	).Preload(database.Adapter().FoxAccount.Config).First()
	if err != nil {
		return &pb.OpenOrderResponse{
			Success: false,
			Message: fmt.Sprintf("获取账户失败: %v", err),
		}, nil
	}

	if account.Exchange != req.Exchange {
		return &pb.OpenOrderResponse{
			Success: false,
			Message: "账户所属交易所与请求不一致",
		}, nil
	}

	exchangeClient, err := exchange.GetManager().GetExchange(req.Exchange)
	if err != nil {
		return &pb.OpenOrderResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所客户端失败: %v", err),
		}, nil
	}

	if err := exchangeClient.SetAccount(ctx, account); err != nil {
		return &pb.OpenOrderResponse{
			Success: false,
			Message: fmt.Sprintf("设置账户失败: %v", err),
		}, nil
	}

	symbolList := NewSymbolServer().getSymbolList(ctx, req.Exchange, exchangeClient)
	var symbolInfo config.SymbolInfo
	for _, symbol := range symbolList {
		if symbol.Name == strings.ToUpper(req.Symbol) {
			symbolInfo = symbol
			break
		}
	}
	if symbolInfo.Name == "" {
		return &pb.OpenOrderResponse{
			Success: false,
			Message: fmt.Sprintf("交易对 %s 不存在", req.Symbol),
		}, nil
	}

	amountDecimal, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return &pb.OpenOrderResponse{
			Success: false,
			Message: fmt.Sprintf("amount 解析失败: %v", err),
		}, nil
	}

	side := req.Side
	if side == "" {
		if req.PosSide == "long" {
			side = "buy"
		} else {
			side = "sell"
		}
	}

	orderCostReq := &exchange.OrderCostReq{
		Side:       side,
		Symbol:     req.Symbol,
		Amount:     amountDecimal.String(),
		AmountType: req.AmountType,
		MarginType: req.Margin,
	}

	costRes, err := exchangeClient.CalcOrderCost(ctx, orderCostReq)
	if err != nil {
		return &pb.OpenOrderResponse{
			Success: false,
			Message: fmt.Sprintf("计算下单成本失败: %v", err),
		}, nil
	}

	if !costRes.CanBuyWithTaker {
		return &pb.OpenOrderResponse{
			Success: false,
			Message: fmt.Sprintf("当前不可提交订单，标的价格: %s，可用资金: %s，手续费: %s", costRes.MarkPrice, costRes.AvailableFunds, costRes.Fee),
		}, nil
	}

	strategy := req.Strategy
	if strings.TrimSpace(strategy) != "" {
		engineClient := syntax.NewEngine()
		node, err := engineClient.Parse(strategy)
		if err != nil {
			return &pb.OpenOrderResponse{
				Success: false,
				Message: fmt.Sprintf("解析策略失败: %v", err),
			}, nil
		}
		if err := engineClient.GetEvaluator().Validate(node); err != nil {
			return &pb.OpenOrderResponse{
				Success: false,
				Message: fmt.Sprintf("策略校验失败: %v", err),
			}, nil
		}
	}

	order := &model.FoxOrder{
		OrderID:    exchangeClient.GetClientOrderId(ctx),
		Exchange:   req.Exchange,
		AccountID:  req.AccountId,
		Symbol:     req.Symbol,
		PosSide:    req.PosSide,
		MarginType: req.Margin,
		Size:       amountDecimal.String(),
		SizeType:   req.AmountType,
		Side:       side,
		OrderType:  "market",
		Strategy:   strategy,
		Type:       "open",
		Status:     "waiting",
	}

	if err := database.Adapter().FoxOrder.Create(order); err != nil {
		return &pb.OpenOrderResponse{
			Success: false,
			Message: fmt.Sprintf("创建订单失败: %v", err),
		}, nil
	}

	pbOrder := &pb.OrderItem{
		Id:         order.ID,
		Exchange:   order.Exchange,
		AccountId:  order.AccountID,
		Symbol:     order.Symbol,
		Side:       order.Side,
		PosSide:    order.PosSide,
		MarginType: order.MarginType,
		Price:      order.Price,
		Size:       order.Size,
		SizeType:   order.SizeType,
		OrderType:  order.OrderType,
		Strategy:   order.Strategy,
		OrderId:    order.OrderID,
		Type:       order.Type,
		Status:     order.Status,
		Msg:        order.Msg,
		CreatedAt:  order.CreatedAt.Unix(),
		UpdatedAt:  order.UpdatedAt.Unix(),
	}

	return &pb.OpenOrderResponse{
		Success: true,
		Message: fmt.Sprintf("策略订单已创建，订单号: %s", order.OrderID),
		Order:   pbOrder,
	}, nil
}

// CloseOrder 创建平仓订单
func (s *OrderServer) CloseOrder(ctx context.Context, req *pb.CloseOrderRequest) (*pb.CloseOrderResponse, error) {
	if req.AccountId <= 0 {
		return &pb.CloseOrderResponse{Success: false, Message: "account_id 是必填参数"}, nil
	}
	if req.Exchange == "" {
		return &pb.CloseOrderResponse{Success: false, Message: "exchange 是必填参数"}, nil
	}
	if req.Symbol == "" {
		return &pb.CloseOrderResponse{Success: false, Message: "symbol 是必填参数"}, nil
	}
	if req.PosSide != "long" && req.PosSide != "short" {
		return &pb.CloseOrderResponse{Success: false, Message: "pos_side 只能为 long 或 short"}, nil
	}
	if req.Margin != "isolated" && req.Margin != "cross" {
		return &pb.CloseOrderResponse{Success: false, Message: "margin 只能为 isolated 或 cross"}, nil
	}

	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.ID.Eq(req.AccountId),
	).Preload(database.Adapter().FoxAccount.Config).First()
	if err != nil {
		return &pb.CloseOrderResponse{
			Success: false,
			Message: fmt.Sprintf("获取账户失败: %v", err),
		}, nil
	}

	if account.Exchange != req.Exchange {
		return &pb.CloseOrderResponse{
			Success: false,
			Message: "账户所属交易所与请求不一致",
		}, nil
	}

	exchangeClient, err := exchange.GetManager().GetExchange(req.Exchange)
	if err != nil {
		return &pb.CloseOrderResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所客户端失败: %v", err),
		}, nil
	}

	if err := exchangeClient.SetAccount(ctx, account); err != nil {
		return &pb.CloseOrderResponse{
			Success: false,
			Message: fmt.Sprintf("设置账户失败: %v", err),
		}, nil
	}

	symbolList := NewSymbolServer().getSymbolList(ctx, req.Exchange, exchangeClient)
	var symbolExists bool
	for _, symbol := range symbolList {
		if symbol.Name == strings.ToUpper(req.Symbol) {
			symbolExists = true
			break
		}
	}
	if !symbolExists {
		return &pb.CloseOrderResponse{
			Success: false,
			Message: fmt.Sprintf("交易对 %s 不存在", req.Symbol),
		}, nil
	}

	strategy := strings.TrimSpace(req.Strategy)
	if strategy != "" {
		engineClient := syntax.NewEngine()
		node, err := engineClient.Parse(strategy)
		if err != nil {
			return &pb.CloseOrderResponse{
				Success: false,
				Message: fmt.Sprintf("解析策略失败: %v", err),
			}, nil
		}
		if err := engineClient.GetEvaluator().Validate(node); err != nil {
			return &pb.CloseOrderResponse{
				Success: false,
				Message: fmt.Sprintf("策略校验失败: %v", err),
			}, nil
		}
	}

	side := "sell"
	if req.PosSide == "short" {
		side = "buy"
	}

	order := &model.FoxOrder{
		OrderID:    exchangeClient.GetClientOrderId(ctx),
		Exchange:   req.Exchange,
		AccountID:  req.AccountId,
		Symbol:     req.Symbol,
		PosSide:    req.PosSide,
		MarginType: req.Margin,
		Side:       side,
		OrderType:  "market",
		Strategy:   strategy,
		Type:       "close",
		Status:     "waiting",
	}

	if err := database.Adapter().FoxOrder.Create(order); err != nil {
		return &pb.CloseOrderResponse{
			Success: false,
			Message: fmt.Sprintf("创建订单失败: %v", err),
		}, nil
	}

	pbOrder := &pb.OrderItem{
		Id:         order.ID,
		Exchange:   order.Exchange,
		AccountId:  order.AccountID,
		Symbol:     order.Symbol,
		Side:       order.Side,
		PosSide:    order.PosSide,
		MarginType: order.MarginType,
		Price:      order.Price,
		Size:       order.Size,
		SizeType:   order.SizeType,
		OrderType:  order.OrderType,
		Strategy:   order.Strategy,
		OrderId:    order.OrderID,
		Type:       order.Type,
		Status:     order.Status,
		Msg:        order.Msg,
		CreatedAt:  order.CreatedAt.Unix(),
		UpdatedAt:  order.UpdatedAt.Unix(),
	}

	return &pb.CloseOrderResponse{
		Success: true,
		Message: "策略订单已创建，等待策略条件满足",
		Order:   pbOrder,
	}, nil
}

// CancelOrder 取消策略订单
func (s *OrderServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	if req.AccountId <= 0 {
		return &pb.CancelOrderResponse{Success: false, Message: "account_id 是必填参数"}, nil
	}
	if req.Exchange == "" {
		return &pb.CancelOrderResponse{Success: false, Message: "exchange 是必填参数"}, nil
	}
	if req.Symbol == "" {
		return &pb.CancelOrderResponse{Success: false, Message: "symbol 是必填参数"}, nil
	}
	if req.Side == "" || req.PosSide == "" || req.Amount == "" {
		return &pb.CancelOrderResponse{Success: false, Message: "side、pos_side、amount 均为必填参数"}, nil
	}

	order, err := database.Adapter().FoxOrder.Where(
		database.Adapter().FoxOrder.AccountID.Eq(req.AccountId),
		database.Adapter().FoxOrder.Exchange.Eq(req.Exchange),
		database.Adapter().FoxOrder.Symbol.Eq(strings.ToUpper(req.Symbol)),
		database.Adapter().FoxOrder.Side.Eq(req.Side),
		database.Adapter().FoxOrder.PosSide.Eq(req.PosSide),
		database.Adapter().FoxOrder.Size.Eq(req.Amount),
		database.Adapter().FoxOrder.SizeType.Eq(req.AmountType),
		database.Adapter().FoxOrder.Status.Eq("waiting"),
	).First()
	if err != nil {
		return &pb.CancelOrderResponse{
			Success: false,
			Message: fmt.Sprintf("查询订单失败: %v", err),
		}, nil
	}

	order.Status = "cancelled"
	if err := database.Adapter().FoxOrder.Save(order); err != nil {
		return &pb.CancelOrderResponse{
			Success: false,
			Message: fmt.Sprintf("更新订单失败: %v", err),
		}, nil
	}

	return &pb.CancelOrderResponse{
		Success: true,
		Message: fmt.Sprintf("订单（%s:%s:%s:%s）取消成功", req.Symbol, req.Side, req.PosSide, req.Amount),
		Order: &pb.OrderItem{
			Id:         order.ID,
			Exchange:   order.Exchange,
			AccountId:  order.AccountID,
			Symbol:     order.Symbol,
			Side:       order.Side,
			PosSide:    order.PosSide,
			MarginType: order.MarginType,
			Price:      order.Price,
			Size:       order.Size,
			SizeType:   order.SizeType,
			OrderType:  order.OrderType,
			Strategy:   order.Strategy,
			OrderId:    order.OrderID,
			Type:       order.Type,
			Status:     order.Status,
			Msg:        order.Msg,
			CreatedAt:  order.CreatedAt.Unix(),
			UpdatedAt:  order.UpdatedAt.Unix(),
		},
	}, nil
}
