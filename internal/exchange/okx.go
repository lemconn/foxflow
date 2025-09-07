package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"foxflow/internal/models"
)

// OKXExchange OKX交易所实现
type OKXExchange struct {
	name     string
	apiURL   string
	proxyURL string
	client   *http.Client
	user     *models.FoxUser
}

// NewOKXExchange 创建OKX交易所实例
func NewOKXExchange(apiURL, proxyURL string) *OKXExchange {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 如果设置了代理
	if proxyURL != "" {
		proxyURLParsed, err := url.Parse(proxyURL)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURLParsed),
			}
		}
	}

	return &OKXExchange{
		name:     "okx",
		apiURL:   apiURL,
		proxyURL: proxyURL,
		client:   client,
	}
}

func (e *OKXExchange) GetName() string {
	return e.name
}

func (e *OKXExchange) GetAPIURL() string {
	return e.apiURL
}

func (e *OKXExchange) GetProxyURL() string {
	return e.proxyURL
}

func (e *OKXExchange) Connect(ctx context.Context, user *models.FoxUser) error {
	e.user = user
	// 这里可以添加连接测试逻辑
	return nil
}

func (e *OKXExchange) Disconnect() error {
	e.user = nil
	return nil
}

func (e *OKXExchange) GetBalance(ctx context.Context) ([]Asset, error) {
	// 实现获取余额逻辑
	// 这里返回模拟数据，实际应该调用OKX API
	return []Asset{
		{Currency: "USDT", Balance: 10000.0, Frozen: 0.0, Available: 10000.0},
		{Currency: "BTC", Balance: 0.1, Frozen: 0.0, Available: 0.1},
	}, nil
}

func (e *OKXExchange) GetPositions(ctx context.Context) ([]Position, error) {
	// 实现获取仓位逻辑
	// 这里返回模拟数据，实际应该调用OKX API
	return []Position{
		{Symbol: "BTC/USDT", PosSide: "long", Size: 0.01, AvgPrice: 50000.0, UnrealPnl: 100.0},
	}, nil
}

func (e *OKXExchange) GetOrders(ctx context.Context, symbol string, status string) ([]Order, error) {
	// 实现获取订单逻辑
	// 这里返回模拟数据，实际应该调用OKX API
	return []Order{
		{
			ID:      "123456789",
			Symbol:  "BTC/USDT",
			Side:    "buy",
			PosSide: "long",
			Price:   50000.0,
			Size:    0.01,
			Type:    "limit",
			Status:  "pending",
			Filled:  0.0,
			Remain:  0.01,
		},
	}, nil
}

func (e *OKXExchange) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
	// 实现创建订单逻辑
	// 这里返回模拟数据，实际应该调用OKX API
	order.ID = fmt.Sprintf("order_%d", time.Now().Unix())
	order.Status = "pending"
	return order, nil
}

func (e *OKXExchange) CancelOrder(ctx context.Context, orderID string) error {
	// 实现取消订单逻辑
	// 实际应该调用OKX API
	return nil
}

func (e *OKXExchange) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	// 实现获取单个行情逻辑
	// 这里返回模拟数据，实际应该调用OKX API
	return &Ticker{
		Symbol: symbol,
		Price:  50000.0,
		Volume: 1000.0,
		High:   51000.0,
		Low:    49000.0,
	}, nil
}

func (e *OKXExchange) GetTickers(ctx context.Context) ([]Ticker, error) {
	// 实现获取所有行情逻辑
	// 这里返回模拟数据，实际应该调用OKX API
	return []Ticker{
		{Symbol: "BTC/USDT", Price: 50000.0, Volume: 1000.0, High: 51000.0, Low: 49000.0},
		{Symbol: "ETH/USDT", Price: 3000.0, Volume: 2000.0, High: 3100.0, Low: 2900.0},
	}, nil
}

func (e *OKXExchange) GetSymbols(ctx context.Context) ([]string, error) {
	// 实现获取交易对列表逻辑
	// 这里返回模拟数据，实际应该调用OKX API
	return []string{
		"BTC/USDT",
		"ETH/USDT",
		"BNB/USDT",
		"ADA/USDT",
		"SOL/USDT",
	}, nil
}

func (e *OKXExchange) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	// 实现设置杠杆逻辑
	// 实际应该调用OKX API
	return nil
}

func (e *OKXExchange) SetMarginType(ctx context.Context, symbol string, marginType string) error {
	// 实现设置保证金模式逻辑
	// 实际应该调用OKX API
	return nil
}

// 辅助方法：构建请求头
func (e *OKXExchange) buildHeaders(method, path, body string) map[string]string {
	// 这里应该实现OKX的签名逻辑
	// 为了简化，这里只返回基本的请求头
	return map[string]string{
		"Content-Type":         "application/json",
		"OK-ACCESS-KEY":        e.user.AccessKey,
		"OK-ACCESS-SIGN":       "mock_signature", // 实际应该计算签名
		"OK-ACCESS-TIMESTAMP":  strconv.FormatInt(time.Now().Unix(), 10),
		"OK-ACCESS-PASSPHRASE": "mock_passphrase", // 实际应该从用户配置获取
	}
}

// 辅助方法：发送HTTP请求
func (e *OKXExchange) sendRequest(ctx context.Context, method, path string, params map[string]interface{}) (map[string]interface{}, error) {
	// 构建完整URL
	fullURL := e.apiURL + path

	// 构建请求体
	var body []byte
	if method == "POST" || method == "PUT" {
		bodyBytes, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		body = bodyBytes
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	headers := e.buildHeaders(method, path, string(body))
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
