package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/api/v5/account/balance"
	result, err := e.sendRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	var assets []Asset
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if details, ok := data[0].(map[string]interface{})["details"].([]interface{}); ok {
			for _, detail := range details {
				if detailMap, ok := detail.(map[string]interface{}); ok {
					currency := detailMap["ccy"].(string)
					balance, _ := strconv.ParseFloat(detailMap["eq"].(string), 64)
					available, _ := strconv.ParseFloat(detailMap["availEq"].(string), 64)
					frozen := balance - available

					if balance > 0 {
						assets = append(assets, Asset{
							Currency:  currency,
							Balance:   balance,
							Available: available,
							Frozen:    frozen,
						})
					}
				}
			}
		}
	}

	return assets, nil
}

func (e *OKXExchange) GetPositions(ctx context.Context) ([]Position, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/api/v5/account/positions"
	result, err := e.sendRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var positions []Position
	if data, ok := result["data"].([]interface{}); ok {
		for _, item := range data {
			if posMap, ok := item.(map[string]interface{}); ok {
				pos, _ := strconv.ParseFloat(posMap["pos"].(string), 64)
				if pos != 0 {
					avgPx, _ := strconv.ParseFloat(posMap["avgPx"].(string), 64)
					upl, _ := strconv.ParseFloat(posMap["upl"].(string), 64)

					positions = append(positions, Position{
						Symbol:    posMap["instId"].(string),
						PosSide:   posMap["posSide"].(string),
						Size:      pos,
						AvgPrice:  avgPx,
						UnrealPnl: upl,
					})
				}
			}
		}
	}

	return positions, nil
}

func (e *OKXExchange) GetOrders(ctx context.Context, symbol string, status string) ([]Order, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/api/v5/trade/orders-pending"
	params := make(map[string]interface{})
	if symbol != "" {
		params["instId"] = symbol
	}

	result, err := e.sendRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var orders []Order
	if data, ok := result["data"].([]interface{}); ok {
		for _, item := range data {
			if orderMap, ok := item.(map[string]interface{}); ok {
				price, _ := strconv.ParseFloat(orderMap["px"].(string), 64)
				size, _ := strconv.ParseFloat(orderMap["sz"].(string), 64)
				filled, _ := strconv.ParseFloat(orderMap["fillSz"].(string), 64)
				remain := size - filled

				orders = append(orders, Order{
					ID:      orderMap["ordId"].(string),
					Symbol:  orderMap["instId"].(string),
					Side:    orderMap["side"].(string),
					PosSide: orderMap["posSide"].(string),
					Price:   price,
					Size:    size,
					Type:    orderMap["ordType"].(string),
					Status:  orderMap["state"].(string),
					Filled:  filled,
					Remain:  remain,
				})
			}
		}
	}

	return orders, nil
}

func (e *OKXExchange) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/api/v5/trade/order"
	params := map[string]interface{}{
		"instId":  order.Symbol,
		"tdMode":  "cross", // 全仓模式
		"side":    order.Side,
		"posSide": order.PosSide,
		"ordType": order.Type,
		"sz":      fmt.Sprintf("%.8f", order.Size),
	}

	if order.Type == "limit" {
		params["px"] = fmt.Sprintf("%.8f", order.Price)
	}

	result, err := e.sendRequest(ctx, "POST", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if orderData, ok := data[0].(map[string]interface{}); ok {
			order.ID = orderData["ordId"].(string)
			order.Status = "pending"
		}
	}

	return order, nil
}

func (e *OKXExchange) CancelOrder(ctx context.Context, orderID string) error {
	if e.user == nil {
		return fmt.Errorf("user not connected")
	}

	path := "/api/v5/trade/cancel-order"
	params := map[string]interface{}{
		"ordId": orderID,
	}

	_, err := e.sendRequest(ctx, "POST", path, params)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

func (e *OKXExchange) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	path := "/api/v5/market/ticker"
	params := map[string]interface{}{
		"instId": symbol,
	}

	result, err := e.sendRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if tickerMap, ok := data[0].(map[string]interface{}); ok {
			price, _ := strconv.ParseFloat(tickerMap["last"].(string), 64)
			volume, _ := strconv.ParseFloat(tickerMap["vol24h"].(string), 64)
			high, _ := strconv.ParseFloat(tickerMap["high24h"].(string), 64)
			low, _ := strconv.ParseFloat(tickerMap["low24h"].(string), 64)

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

func (e *OKXExchange) GetTickers(ctx context.Context) ([]Ticker, error) {
	path := "/api/v5/market/tickers"
	params := map[string]interface{}{
		"instType": "SWAP", // 永续合约
	}

	result, err := e.sendRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickers: %w", err)
	}

	var tickers []Ticker
	if data, ok := result["data"].([]interface{}); ok {
		for _, item := range data {
			if tickerMap, ok := item.(map[string]interface{}); ok {
				price, _ := strconv.ParseFloat(tickerMap["last"].(string), 64)
				volume, _ := strconv.ParseFloat(tickerMap["vol24h"].(string), 64)
				high, _ := strconv.ParseFloat(tickerMap["high24h"].(string), 64)
				low, _ := strconv.ParseFloat(tickerMap["low24h"].(string), 64)

				tickers = append(tickers, Ticker{
					Symbol: tickerMap["instId"].(string),
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

func (e *OKXExchange) GetSymbols(ctx context.Context) ([]string, error) {
	path := "/api/v5/public/instruments"
	params := map[string]interface{}{
		"instType": "SWAP", // 永续合约
	}

	result, err := e.sendRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}

	var symbols []string
	if data, ok := result["data"].([]interface{}); ok {
		for _, item := range data {
			if instMap, ok := item.(map[string]interface{}); ok {
				if state := instMap["state"].(string); state == "live" {
					symbols = append(symbols, instMap["instId"].(string))
				}
			}
		}
	}

	return symbols, nil
}

func (e *OKXExchange) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	if e.user == nil {
		return fmt.Errorf("user not connected")
	}

	path := "/api/v5/account/set-leverage"
	params := map[string]interface{}{
		"instId":  symbol,
		"lever":   fmt.Sprintf("%d", leverage),
		"mgnMode": "cross", // 全仓模式
	}

	_, err := e.sendRequest(ctx, "POST", path, params)
	if err != nil {
		return fmt.Errorf("failed to set leverage: %w", err)
	}

	return nil
}

func (e *OKXExchange) SetMarginType(ctx context.Context, symbol string, marginType string) error {
	if e.user == nil {
		return fmt.Errorf("user not connected")
	}

	path := "/api/v5/account/set-margin-mode"
	params := map[string]interface{}{
		"instId":  symbol,
		"mgnMode": marginType,
	}

	_, err := e.sendRequest(ctx, "POST", path, params)
	if err != nil {
		return fmt.Errorf("failed to set margin type: %w", err)
	}

	return nil
}

// 辅助方法：构建请求头
func (e *OKXExchange) buildHeaders(method, path, body string) map[string]string {
	if e.user == nil {
		return map[string]string{
			"Content-Type": "application/json",
		}
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	message := timestamp + strings.ToUpper(method) + path + body
	signature := e.buildSignature(message)

	return map[string]string{
		"Content-Type":         "application/json",
		"OK-ACCESS-KEY":        e.user.AccessKey,
		"OK-ACCESS-SIGN":       signature,
		"OK-ACCESS-TIMESTAMP":  timestamp,
		"OK-ACCESS-PASSPHRASE": e.user.TradeType, // 使用trade_type作为passphrase
	}
}

// 辅助方法：构建OKX签名
func (e *OKXExchange) buildSignature(message string) string {
	h := hmac.New(sha256.New, []byte(e.user.SecretKey))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
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

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	// 检查API错误
	if code, ok := result["code"].(string); ok && code != "0" {
		return nil, fmt.Errorf("API error: %s", result["msg"])
	}

	return result, nil
}
