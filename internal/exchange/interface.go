package exchange

import (
	"context"

	"github.com/lemconn/foxflow/internal/models"
)

// Order 订单信息
type Order struct {
	ID      string  `json:"id"`
	Symbol  string  `json:"symbol"`
	Side    string  `json:"side"`
	PosSide string  `json:"pos_side"`
	Price   float64 `json:"price"`
	Size    float64 `json:"size"`
	Type    string  `json:"type"`
	Status  string  `json:"status"`
	Filled  float64 `json:"filled"`
	Remain  float64 `json:"remain"`
}

// Position 仓位信息
type Position struct {
	Symbol    string  `json:"symbol"`
	PosSide   string  `json:"pos_side"`
	Size      float64 `json:"size"`
	AvgPrice  float64 `json:"avg_price"`
	UnrealPnl float64 `json:"unreal_pnl"`
}

// Asset 资产信息
type Asset struct {
	Currency  string  `json:"currency"`
	Balance   float64 `json:"balance"`
	Frozen    float64 `json:"frozen"`
	Available float64 `json:"available"`
}

// Ticker 行情信息
type Ticker struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
}

// Exchange 交易所接口
type Exchange interface {
	// 基础信息
	GetName() string
	GetAPIURL() string
	GetProxyURL() string

	// 连接管理
	Connect(ctx context.Context, user *models.FoxUser) error
	Disconnect() error

	// 账户信息
	GetBalance(ctx context.Context) ([]Asset, error)
	GetPositions(ctx context.Context) ([]Position, error)

	// 订单管理
	GetOrders(ctx context.Context, symbol string, status string) ([]Order, error)
	CreateOrder(ctx context.Context, order *Order) (*Order, error)
	CancelOrder(ctx context.Context, orderID string) error

	// 行情数据
	GetTicker(ctx context.Context, symbol string) (*Ticker, error)
	GetTickers(ctx context.Context) ([]Ticker, error)

	// 标的配置
	GetSymbols(ctx context.Context) ([]string, error)
	SetLeverage(ctx context.Context, symbol string, leverage int) error
	SetMarginType(ctx context.Context, symbol string, marginType string) error
}
