package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/lemconn/foxflow/proto/generated"
	"google.golang.org/grpc"
)

// Server gRPC服务端
type Server struct {
	pb.UnimplementedFoxFlowServiceServer
	port int
}

// NewServer 创建新的gRPC服务端
func NewServer(port int) *Server {
	return &Server{
		port: port,
	}
}

// Start 启动gRPC服务端
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFoxFlowServiceServer(grpcServer, s)

	log.Printf("gRPC服务端启动在端口 %d", s.port)

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Authenticate 认证方法
func (s *Server) Authenticate(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	// 硬编码的认证信息
	validUsername := "foxflow"
	validPassword := "foxflow"

	if req.Username == validUsername && req.Password == validPassword {
		log.Printf("用户认证成功: %s", req.Username)
		return &pb.AuthResponse{
			Success: true,
			Message: "认证成功",
		}, nil
	}

	log.Printf("用户认证失败: %s", req.Username)
	return &pb.AuthResponse{
		Success: false,
		Message: "用户名或密码错误",
	}, nil
}

// SendCommand 接收命令方法
func (s *Server) SendCommand(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	// 打印接收到的命令到标准输出
	log.Printf("收到命令: %s %v (交易所: %s, 账户: %s)",
		req.Command, req.Args, req.Exchange, req.Account)

	// 暂时不做任何操作，只打印
	fmt.Printf("[gRPC服务端] 收到命令: %s %v (交易所: %s, 账户: %s)\n",
		req.Command, req.Args, req.Exchange, req.Account)

	return &pb.CommandResponse{
		Success: true,
		Message: "命令已接收",
	}, nil
}
