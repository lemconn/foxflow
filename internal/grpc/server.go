package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/lemconn/foxflow/internal/engine"
	pb "github.com/lemconn/foxflow/proto/generated"
	"google.golang.org/grpc"
)

// Server gRPC服务端
type Server struct {
	pb.UnimplementedFoxFlowServiceServer
	port        int
	engine      *engine.Engine
	authManager *AuthManager
}

// NewServer 创建新的gRPC服务端
func NewServer(port int) *Server {
	return &Server{
		port:        port,
		authManager: NewAuthManager(),
	}
}

// SetEngine 设置引擎实例
func (s *Server) SetEngine(engine *engine.Engine) {
	s.engine = engine
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
		// 生成 JWT token
		token, expiresAt, err := s.authManager.GenerateToken(req.Username)
		if err != nil {
			log.Printf("生成 token 失败: %v", err)
			return &pb.AuthResponse{
				Success: false,
				Message: "生成 token 失败",
			}, nil
		}

		log.Printf("用户认证成功: %s", req.Username)
		return &pb.AuthResponse{
			Success:     true,
			Message:     "认证成功",
			AccessToken: token,
			ExpiresAt:   expiresAt,
		}, nil
	}

	log.Printf("用户认证失败: %s", req.Username)
	return &pb.AuthResponse{
		Success: false,
		Message: "用户名或密码错误",
	}, nil
}

// RefreshToken Token 刷新方法
func (s *Server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if req.AccessToken == "" {
		return &pb.RefreshTokenResponse{
			Success: false,
			Message: "缺少 access token",
		}, nil
	}

	// 刷新 token
	newToken, expiresAt, err := s.authManager.RefreshToken(req.AccessToken)
	if err != nil {
		log.Printf("刷新 token 失败: %v", err)
		return &pb.RefreshTokenResponse{
			Success: false,
			Message: fmt.Sprintf("刷新 token 失败: %v", err),
		}, nil
	}

	log.Printf("Token 刷新成功")
	return &pb.RefreshTokenResponse{
		Success:     true,
		Message:     "Token 刷新成功",
		AccessToken: newToken,
		ExpiresAt:   expiresAt,
	}, nil
}

// SendCommand 接收命令方法
func (s *Server) SendCommand(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	// 验证 access token
	if err := s.validateToken(req.AccessToken); err != nil {
		log.Printf("Token 验证失败: %v", err)
		return &pb.CommandResponse{
			Success: false,
			Message: fmt.Sprintf("认证失败: %v", err),
		}, nil
	}

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

// GetNews 获取新闻方法
func (s *Server) GetNews(ctx context.Context, req *pb.GetNewsRequest) (*pb.GetNewsResponse, error) {
	// 验证 access token
	if err := s.validateToken(req.AccessToken); err != nil {
		log.Printf("Token 验证失败: %v", err)
		return &pb.GetNewsResponse{
			Success: false,
			Message: fmt.Sprintf("认证失败: %v", err),
		}, nil
	}

	if s.engine == nil {
		return &pb.GetNewsResponse{
			Success: false,
			Message: "引擎未初始化",
		}, nil
	}

	// 设置默认值
	count := int(req.Count)
	if count <= 0 {
		count = 10
	}
	if count > 100 {
		count = 100 // 限制最大数量
	}

	source := req.Source
	if source == "" {
		source = "blockbeats" // 默认使用 blockbeats
	}

	// 创建带超时的上下文
	newsCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 获取新闻
	newsManager := s.engine.GetNewsManager()
	newsList, err := newsManager.GetNewsFromSource(newsCtx, source, count)
	if err != nil {
		log.Printf("获取新闻失败: %v", err)
		return &pb.GetNewsResponse{
			Success: false,
			Message: fmt.Sprintf("获取新闻失败: %v", err),
		}, nil
	}

	// 转换为 protobuf 格式
	var pbNews []*pb.NewsItem
	for _, item := range newsList {
		pbNews = append(pbNews, &pb.NewsItem{
			Id:          item.ID,
			Title:       item.Title,
			Content:     item.Content,
			Url:         item.URL,
			Source:      item.Source,
			PublishedAt: item.PublishedAt.Unix(),
			Tags:        item.Tags,
			ImageUrl:    item.ImageURL,
		})
	}

	log.Printf("成功获取 %d 条新闻", len(pbNews))
	return &pb.GetNewsResponse{
		Success: true,
		Message: fmt.Sprintf("成功获取 %d 条新闻", len(pbNews)),
		News:    pbNews,
	}, nil
}

// validateToken 验证 access token
func (s *Server) validateToken(token string) error {
	if token == "" {
		return fmt.Errorf("缺少 access token")
	}

	_, err := s.authManager.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("token 验证失败: %w", err)
	}

	return nil
}
