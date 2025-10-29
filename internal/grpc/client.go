package grpc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/lemconn/foxflow/internal/news"
	pb "github.com/lemconn/foxflow/proto/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client gRPC客户端
type Client struct {
	conn        *grpc.ClientConn
	client      pb.FoxFlowServiceClient
	accessToken string
	expiresAt   int64
	mu          sync.RWMutex
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

	// 保存 token 信息
	c.mu.Lock()
	c.accessToken = resp.AccessToken
	c.expiresAt = resp.ExpiresAt
	c.mu.Unlock()

	log.Printf("gRPC认证成功: %s, token 过期时间: %s", username, time.Unix(resp.ExpiresAt, 0).Format("2006-01-02 15:04:05"))
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

// RefreshToken 刷新 token
func (c *Client) RefreshToken() error {
	c.mu.RLock()
	currentToken := c.accessToken
	c.mu.RUnlock()

	if currentToken == "" {
		return fmt.Errorf("没有可用的 token 进行刷新")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.RefreshToken(ctx, &pb.RefreshTokenRequest{
		AccessToken: currentToken,
	})
	if err != nil {
		return fmt.Errorf("刷新 token 失败: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("刷新 token 失败: %s", resp.Message)
	}

	// 更新 token 信息
	c.mu.Lock()
	c.accessToken = resp.AccessToken
	c.expiresAt = resp.ExpiresAt
	c.mu.Unlock()

	log.Printf("Token 刷新成功, 新过期时间: %s", time.Unix(resp.ExpiresAt, 0).Format("2006-01-02 15:04:05"))
	return nil
}

// isTokenExpired 检查 token 是否即将过期（1小时内）
func (c *Client) isTokenExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.accessToken == "" || c.expiresAt == 0 {
		return true
	}

	// 检查是否在 1 小时内过期
	expiresAt := time.Unix(c.expiresAt, 0)
	oneHourFromNow := time.Now().Add(1 * time.Hour)

	return expiresAt.Before(oneHourFromNow)
}

// ensureValidToken 确保 token 有效，如果即将过期则自动刷新
func (c *Client) ensureValidToken() error {
	if c.isTokenExpired() {
		if err := c.RefreshToken(); err != nil {
			return fmt.Errorf("token 刷新失败: %w", err)
		}
	}
	return nil
}

// getAccessToken 获取当前的 access token
func (c *Client) getAccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken
}

// GetTokenExpiry 获取 token 过期时间
func (c *Client) GetTokenExpiry() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.expiresAt
}

// SendCommand 发送命令
func (c *Client) SendCommand(command string, args []string, exchange, account string) error {
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return fmt.Errorf("token 验证失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.SendCommand(ctx, &pb.CommandRequest{
		Command:     command,
		Args:        args,
		Exchange:    exchange,
		Account:     account,
		AccessToken: c.getAccessToken(),
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
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.GetNews(ctx, &pb.GetNewsRequest{
		Count:       int32(count),
		Source:      source,
		AccessToken: c.getAccessToken(),
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
