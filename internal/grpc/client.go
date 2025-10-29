package grpc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lemconn/foxflow/internal/news"
	pb "github.com/lemconn/foxflow/proto/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client gRPC客户端
type Client struct {
	conn   *grpc.ClientConn
	client pb.FoxFlowServiceClient
}

// NewClient 创建新的gRPC客户端
func NewClient(host string, port int) (*Client, error) {
	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := pb.NewFoxFlowServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Authenticate 认证
func (c *Client) Authenticate(username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.client.Authenticate(ctx, &pb.AuthRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		// 检查是否是连接错误
		if isConnectionError(err) {
			return fmt.Errorf("connection failed: %w", err)
		}
		return fmt.Errorf("authentication failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("authentication failed: %s", resp.Message)
	}

	log.Printf("gRPC认证成功: %s", username)
	return nil
}

// isConnectionError 检查是否是连接错误
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "unreachable") ||
		strings.Contains(errStr, "refused")
}

// SendCommand 发送命令
func (c *Client) SendCommand(command string, args []string, exchange, account string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.SendCommand(ctx, &pb.CommandRequest{
		Command:  command,
		Args:     args,
		Exchange: exchange,
		Account:  account,
	})
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("command failed: %s", resp.Message)
	}

	return nil
}

// GetNews 获取新闻
func (c *Client) GetNews(count int, source string) ([]news.NewsItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.GetNews(ctx, &pb.GetNewsRequest{
		Count:  int32(count),
		Source: source,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get news: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get news failed: %s", resp.Message)
	}

	// 转换为内部格式
	var newsList []news.NewsItem
	for _, item := range resp.News {
		newsList = append(newsList, news.NewsItem{
			ID:          item.Id,
			Title:       item.Title,
			Content:     item.Content,
			URL:         item.Url,
			Source:      item.Source,
			PublishedAt: time.Unix(item.PublishedAt, 0),
			Tags:        item.Tags,
			ImageURL:    item.ImageUrl,
		})
	}

	return newsList, nil
}
