package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lemconn/foxflow/internal/models"
)

// BinanceExchange Binance交易所实现
type BinanceExchange struct {
	name     string
	apiURL   string
	proxyURL string
	client   *http.Client
	user     *models.FoxUser
}

// NewBinanceExchange 创建Binance交易所实例
func NewBinanceExchange(apiURL, proxyURL string) *BinanceExchange {
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

	return &BinanceExchange{
		name:     "binance",
		apiURL:   apiURL,
		proxyURL: proxyURL,
		client:   client,
	}
}

func (e *BinanceExchange) GetName() string {
	return e.name
}

func (e *BinanceExchange) GetAPIURL() string {
	return e.apiURL
}

func (e *BinanceExchange) GetProxyURL() string {
	return e.proxyURL
}

func (e *BinanceExchange) Connect(ctx context.Context, user *models.FoxUser) error {
	e.user = user
	return nil
}

func (e *BinanceExchange) Disconnect() error {
	e.user = nil
	return nil
}

func (e *BinanceExchange) SetUSer(ctx context.Context, user *models.FoxUser) error {
	e.user = user
	return nil
}

func (e *BinanceExchange) GetBalance(ctx context.Context) ([]Asset, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/fapi/v2/balance"
	result, err := e.sendRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	var assets []Asset
	if data, ok := result.([]interface{}); ok {
		for _, item := range data {
			if assetMap, ok := item.(map[string]interface{}); ok {
				balance, _ := strconv.ParseFloat(assetMap["balance"].(string), 64)
				available, _ := strconv.ParseFloat(assetMap["availableBalance"].(string), 64)
				frozen := balance - available

				if balance > 0 {
					assets = append(assets, Asset{
						Currency:  assetMap["asset"].(string),
						Balance:   balance,
						Available: available,
						Frozen:    frozen,
					})
				}
			}
		}
	}

	return assets, nil
}

func (e *BinanceExchange) GetPositions(ctx context.Context) ([]Position, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/fapi/v2/positionRisk"
	result, err := e.sendRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var positions []Position
	if data, ok := result.([]interface{}); ok {
		for _, item := range data {
			if posMap, ok := item.(map[string]interface{}); ok {
				positionAmt, _ := strconv.ParseFloat(posMap["positionAmt"].(string), 64)
				if positionAmt != 0 {
					entryPrice, _ := strconv.ParseFloat(posMap["entryPrice"].(string), 64)
					unRealizedProfit, _ := strconv.ParseFloat(posMap["unRealizedProfit"].(string), 64)

					posSide := "long"
					if positionAmt < 0 {
						posSide = "short"
						positionAmt = -positionAmt
					}

					positions = append(positions, Position{
						Symbol:    posMap["symbol"].(string),
						PosSide:   posSide,
						Size:      positionAmt,
						AvgPrice:  entryPrice,
						UnrealPnl: unRealizedProfit,
					})
				}
			}
		}
	}

	return positions, nil
}

func (e *BinanceExchange) GetOrders(ctx context.Context, symbol string, status string) ([]Order, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/fapi/v1/openOrders"
	params := make(map[string]interface{})
	if symbol != "" {
		params["symbol"] = symbol
	}

	result, err := e.sendRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var orders []Order
	if data, ok := result.([]interface{}); ok {
		for _, item := range data {
			if orderMap, ok := item.(map[string]interface{}); ok {
				price, _ := strconv.ParseFloat(orderMap["price"].(string), 64)
				origQty, _ := strconv.ParseFloat(orderMap["origQty"].(string), 64)
				executedQty, _ := strconv.ParseFloat(orderMap["executedQty"].(string), 64)
				remain := origQty - executedQty

				orders = append(orders, Order{
					ID:      fmt.Sprintf("%.0f", orderMap["orderId"].(float64)),
					Symbol:  orderMap["symbol"].(string),
					Side:    orderMap["side"].(string),
					PosSide: orderMap["side"].(string), // Binance期货中side就是posSide
					Price:   price,
					Size:    origQty,
					Type:    orderMap["type"].(string),
					Status:  orderMap["status"].(string),
					Filled:  executedQty,
					Remain:  remain,
				})
			}
		}
	}

	return orders, nil
}

func (e *BinanceExchange) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	path := "/fapi/v1/order"
	params := map[string]interface{}{
		"symbol":   order.Symbol,
		"side":     order.Side,
		"type":     order.Type,
		"quantity": fmt.Sprintf("%.8f", order.Size),
	}

	if order.Type == "LIMIT" {
		params["price"] = fmt.Sprintf("%.8f", order.Price)
		params["timeInForce"] = "GTC"
	}

	result, err := e.sendRequest(ctx, "POST", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if orderData, ok := result.(map[string]interface{}); ok {
		order.ID = fmt.Sprintf("%.0f", orderData["orderId"].(float64))
		order.Status = orderData["status"].(string)
	}

	return order, nil
}

func (e *BinanceExchange) CancelOrder(ctx context.Context, orderID string) error {
	if e.user == nil {
		return fmt.Errorf("user not connected")
	}

	path := "/fapi/v1/order"
	params := map[string]interface{}{
		"orderId": orderID,
	}

	_, err := e.sendRequest(ctx, "DELETE", path, params)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

func (e *BinanceExchange) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	path := "/fapi/v1/ticker/24hr"
	params := map[string]interface{}{
		"symbol": symbol,
	}

	result, err := e.sendRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	if tickerMap, ok := result.(map[string]interface{}); ok {
		price, _ := strconv.ParseFloat(tickerMap["lastPrice"].(string), 64)
		volume, _ := strconv.ParseFloat(tickerMap["volume"].(string), 64)
		high, _ := strconv.ParseFloat(tickerMap["highPrice"].(string), 64)
		low, _ := strconv.ParseFloat(tickerMap["lowPrice"].(string), 64)

		return &Ticker{
			Symbol: symbol,
			Price:  price,
			Volume: volume,
			High:   high,
			Low:    low,
		}, nil
	}

	return nil, fmt.Errorf("no ticker data found")
}

func (e *BinanceExchange) GetTickers(ctx context.Context) ([]Ticker, error) {
	path := "/fapi/v1/ticker/24hr"
	result, err := e.sendRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickers: %w", err)
	}

	var tickers []Ticker
	if data, ok := result.([]interface{}); ok {
		for _, item := range data {
			if tickerMap, ok := item.(map[string]interface{}); ok {
				price, _ := strconv.ParseFloat(tickerMap["lastPrice"].(string), 64)
				volume, _ := strconv.ParseFloat(tickerMap["volume"].(string), 64)
				high, _ := strconv.ParseFloat(tickerMap["highPrice"].(string), 64)
				low, _ := strconv.ParseFloat(tickerMap["lowPrice"].(string), 64)

				tickers = append(tickers, Ticker{
					Symbol: tickerMap["symbol"].(string),
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

func (e *BinanceExchange) GetSymbols(ctx context.Context, userSymbol string) (*Symbol, error) {

	return nil, nil
}

func (e *BinanceExchange) SetLeverage(ctx context.Context, symbol string, leverage int, marginType string) error {
	if e.user == nil {
		return fmt.Errorf("user not connected")
	}

	path := "/fapi/v1/leverage"
	params := map[string]interface{}{
		"symbol":   symbol,
		"leverage": leverage,
	}

	_, err := e.sendRequest(ctx, "POST", path, params)
	if err != nil {
		return fmt.Errorf("failed to set leverage: %w", err)
	}

	return nil
}

func (e *BinanceExchange) SetMarginType(ctx context.Context, symbol string, marginType string) error {
	if e.user == nil {
		return fmt.Errorf("user not connected")
	}

	path := "/fapi/v1/marginType"
	params := map[string]interface{}{
		"symbol":     symbol,
		"marginType": marginType,
	}

	_, err := e.sendRequest(ctx, "POST", path, params)
	if err != nil {
		return fmt.Errorf("failed to set margin type: %w", err)
	}

	return nil
}

// ConvertToExchangeSymbol 将用户输入的币种名称转换为Binance交易所格式
// 例如：BTC -> BTCUSDT
func (e *BinanceExchange) ConvertToExchangeSymbol(userSymbol string) string {
	// Binance期货通常使用币种+USDT的格式
	return userSymbol + "USDT"
}

// ConvertFromExchangeSymbol 将Binance交易所格式的币种名称转换为用户格式
// 例如：BTCUSDT -> BTC
func (e *BinanceExchange) ConvertFromExchangeSymbol(exchangeSymbol string) string {
	// 去除USDT后缀
	if len(exchangeSymbol) > 4 && exchangeSymbol[len(exchangeSymbol)-4:] == "USDT" {
		return exchangeSymbol[:len(exchangeSymbol)-4]
	}
	// 去除BUSD后缀
	if len(exchangeSymbol) > 4 && exchangeSymbol[len(exchangeSymbol)-4:] == "BUSD" {
		return exchangeSymbol[:len(exchangeSymbol)-4]
	}
	// 去除USDC后缀
	if len(exchangeSymbol) > 4 && exchangeSymbol[len(exchangeSymbol)-4:] == "USDC" {
		return exchangeSymbol[:len(exchangeSymbol)-4]
	}
	// 如果没有匹配的后缀，返回原符号
	return exchangeSymbol
}

// 辅助方法：构建签名
func (e *BinanceExchange) buildSignature(queryString string) string {
	h := hmac.New(sha256.New, []byte(e.user.SecretKey))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

// 辅助方法：发送HTTP请求
func (e *BinanceExchange) sendRequest(ctx context.Context, method, path string, params map[string]interface{}) (interface{}, error) {
	// 构建完整URL
	fullURL := e.apiURL + path

	// 构建查询参数
	query := url.Values{}
	if e.user != nil {
		query.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	}

	// 添加其他参数
	for key, value := range params {
		query.Set(key, fmt.Sprintf("%v", value))
	}

	// 构建请求体
	var body []byte
	var err error
	if method == "POST" || method == "PUT" || method == "DELETE" {
		body, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
	}

	// 添加签名
	if e.user != nil {
		queryString := query.Encode()
		signature := e.buildSignature(queryString)
		query.Set("signature", signature)
		fullURL += "?" + query.Encode()
	} else {
		if len(params) > 0 {
			fullURL += "?" + query.Encode()
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if e.user != nil {
		req.Header.Set("X-MBX-APIKEY", e.user.AccessKey)
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
		if code, ok := resultMap["code"].(float64); ok && code != 200 {
			return nil, fmt.Errorf("API error: %s", resultMap["msg"])
		}
	}

	return result, nil
}
