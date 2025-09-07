package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"foxflow/internal/models"
)

// GateExchange Gate交易所实现
type GateExchange struct {
	name     string
	apiURL   string
	proxyURL string
	client   *http.Client
	user     *models.FoxUser
}

// NewGateExchange 创建Gate交易所实例
func NewGateExchange(apiURL, proxyURL string) *GateExchange {
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

	return &GateExchange{
		name:     "gate",
		apiURL:   apiURL,
		proxyURL: proxyURL,
		client:   client,
	}
}

func (e *GateExchange) GetName() string {
	return e.name
}

func (e *GateExchange) GetAPIURL() string {
	return e.apiURL
}

func (e *GateExchange) GetProxyURL() string {
	return e.proxyURL
}

func (e *GateExchange) Connect(ctx context.Context, user *models.FoxUser) error {
	e.user = user
	return nil
}

func (e *GateExchange) Disconnect() error {
	e.user = nil
	return nil
}

func (e *GateExchange) GetBalance(ctx context.Context) ([]Asset, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/api/v4/futures/usdt/accounts"
	result, err := e.sendRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	var assets []Asset
	if data, ok := result.(map[string]interface{}); ok {
		if total, ok := data["total"].(string); ok {
			if totalFloat, err := strconv.ParseFloat(total, 64); err == nil && totalFloat > 0 {
				assets = append(assets, Asset{
					Currency:  "USDT",
					Balance:   totalFloat,
					Available: totalFloat, // Gate.io期货账户通常只有USDT
					Frozen:    0,
				})
			}
		}
	}

	return assets, nil
}

func (e *GateExchange) GetPositions(ctx context.Context) ([]Position, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/api/v4/futures/usdt/positions"
	result, err := e.sendRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var positions []Position
	if data, ok := result.([]interface{}); ok {
		for _, item := range data {
			if posMap, ok := item.(map[string]interface{}); ok {
				size, _ := strconv.ParseFloat(posMap["size"].(string), 64)
				if size != 0 {
					entryPrice, _ := strconv.ParseFloat(posMap["entry_price"].(string), 64)
					unrealizedPnl, _ := strconv.ParseFloat(posMap["unrealised_pnl"].(string), 64)

					posSide := "long"
					if size < 0 {
						posSide = "short"
						size = -size
					}

					positions = append(positions, Position{
						Symbol:    posMap["contract"].(string),
						PosSide:   posSide,
						Size:      size,
						AvgPrice:  entryPrice,
						UnrealPnl: unrealizedPnl,
					})
				}
			}
		}
	}

	return positions, nil
}

func (e *GateExchange) GetOrders(ctx context.Context, symbol string, status string) ([]Order, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/api/v4/futures/usdt/orders"
	params := make(map[string]interface{})
	if symbol != "" {
		params["contract"] = symbol
	}
	params["status"] = "open" // 只获取未成交订单

	result, err := e.sendRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var orders []Order
	if data, ok := result.([]interface{}); ok {
		for _, item := range data {
			if orderMap, ok := item.(map[string]interface{}); ok {
				price, _ := strconv.ParseFloat(orderMap["price"].(string), 64)
				size, _ := strconv.ParseFloat(orderMap["size"].(string), 64)
				filled, _ := strconv.ParseFloat(orderMap["fill_price"].(string), 64)
				remain := size - filled

				orders = append(orders, Order{
					ID:      fmt.Sprintf("%.0f", orderMap["id"].(float64)),
					Symbol:  orderMap["contract"].(string),
					Side:    orderMap["side"].(string),
					PosSide: orderMap["side"].(string), // Gate.io中side就是posSide
					Price:   price,
					Size:    size,
					Type:    orderMap["type"].(string),
					Status:  orderMap["status"].(string),
					Filled:  filled,
					Remain:  remain,
				})
			}
		}
	}

	return orders, nil
}

func (e *GateExchange) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/api/v4/futures/usdt/orders"
	params := map[string]interface{}{
		"contract": order.Symbol,
		"side":     order.Side,
		"type":     order.Type,
		"size":     order.Size,
	}

	if order.Type == "limit" {
		params["price"] = order.Price
	}

	result, err := e.sendRequest(ctx, "POST", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if orderData, ok := result.(map[string]interface{}); ok {
		order.ID = fmt.Sprintf("%.0f", orderData["id"].(float64))
		order.Status = orderData["status"].(string)
	}

	return order, nil
}

func (e *GateExchange) CancelOrder(ctx context.Context, orderID string) error {
	if e.user == nil {
		return fmt.Errorf("user not connected")
	}

	path := "/api/v4/futures/usdt/orders/" + orderID
	_, err := e.sendRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

func (e *GateExchange) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	path := "/api/v4/futures/usdt/tickers"
	params := map[string]interface{}{
		"contract": symbol,
	}

	result, err := e.sendRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	if data, ok := result.([]interface{}); ok && len(data) > 0 {
		if tickerMap, ok := data[0].(map[string]interface{}); ok {
			price, _ := strconv.ParseFloat(tickerMap["last"].(string), 64)
			volume, _ := strconv.ParseFloat(tickerMap["volume_24h"].(string), 64)
			high, _ := strconv.ParseFloat(tickerMap["high_24h"].(string), 64)
			low, _ := strconv.ParseFloat(tickerMap["low_24h"].(string), 64)

			return &Ticker{
				Symbol: symbol,
				Price:  price,
				Volume: volume,
				High:   high,
				Low:    low,
			}, nil
		}
	}

	return nil, fmt.Errorf("no ticker data found")
}

func (e *GateExchange) GetTickers(ctx context.Context) ([]Ticker, error) {
	path := "/api/v4/futures/usdt/tickers"
	result, err := e.sendRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickers: %w", err)
	}

	var tickers []Ticker
	if data, ok := result.([]interface{}); ok {
		for _, item := range data {
			if tickerMap, ok := item.(map[string]interface{}); ok {
				price, _ := strconv.ParseFloat(tickerMap["last"].(string), 64)
				volume, _ := strconv.ParseFloat(tickerMap["volume_24h"].(string), 64)
				high, _ := strconv.ParseFloat(tickerMap["high_24h"].(string), 64)
				low, _ := strconv.ParseFloat(tickerMap["low_24h"].(string), 64)

				tickers = append(tickers, Ticker{
					Symbol: tickerMap["contract"].(string),
					Price:  price,
					Volume: volume,
					High:   high,
					Low:    low,
				})
			}
		}
	}

	return tickers, nil
}

func (e *GateExchange) GetSymbols(ctx context.Context) ([]string, error) {
	path := "/api/v4/futures/usdt/contracts"
	result, err := e.sendRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}

	var symbols []string
	if data, ok := result.([]interface{}); ok {
		for _, item := range data {
			if contractMap, ok := item.(map[string]interface{}); ok {
				if tradeStatus := contractMap["trade_status"].(string); tradeStatus == "tradable" {
					symbols = append(symbols, contractMap["name"].(string))
				}
			}
		}
	}

	return symbols, nil
}

func (e *GateExchange) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	if e.user == nil {
		return fmt.Errorf("user not connected")
	}

	path := "/api/v4/futures/usdt/positions/" + symbol
	params := map[string]interface{}{
		"leverage": fmt.Sprintf("%d", leverage),
	}

	_, err := e.sendRequest(ctx, "POST", path, params)
	if err != nil {
		return fmt.Errorf("failed to set leverage: %w", err)
	}

	return nil
}

func (e *GateExchange) SetMarginType(ctx context.Context, symbol string, marginType string) error {
	// Gate.io期货默认使用全仓模式，这里返回成功
	return nil
}

// 辅助方法：构建签名
func (e *GateExchange) buildSignature(method, path, queryString, body string) string {
	message := method + "\n" + path + "\n" + queryString + "\n" + body
	h := hmac.New(sha512.New, []byte(e.user.SecretKey))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// 辅助方法：发送HTTP请求
func (e *GateExchange) sendRequest(ctx context.Context, method, path string, params map[string]interface{}) (interface{}, error) {
	// 构建完整URL
	fullURL := e.apiURL + path

	// 构建查询参数
	query := url.Values{}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	query.Set("timestamp", timestamp)

	// 添加其他参数
	for key, value := range params {
		query.Set(key, fmt.Sprintf("%v", value))
	}

	// 构建请求体
	var body []byte
	var err error
	if method == "POST" || method == "PUT" {
		body, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
	}

	// 构建签名
	queryString := query.Encode()
	signature := e.buildSignature(method, path, queryString, string(body))

	// 添加签名到查询参数
	query.Set("signature", signature)
	fullURL += "?" + query.Encode()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if e.user != nil {
		req.Header.Set("KEY", e.user.AccessKey)
		req.Header.Set("SIGN", signature)
		req.Header.Set("Timestamp", timestamp)
	}

	// 发送请求
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	// 检查API错误
	if resultMap, ok := result.(map[string]interface{}); ok {
		if label, ok := resultMap["label"].(string); ok && label != "OK" {
			return nil, fmt.Errorf("API error: %s", resultMap["message"])
		}
	}

	return result, nil
}
