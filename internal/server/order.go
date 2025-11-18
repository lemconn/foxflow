package server

import (
	"context"
	"fmt"
	"log"

	"github.com/lemconn/foxflow/internal/database"
	pb "github.com/lemconn/foxflow/proto/generated"
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
