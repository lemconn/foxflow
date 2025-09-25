package exchange

import (
	"context"

	"github.com/lemconn/foxflow/internal/models"
)

// Order 订单信息
type Order struct {
	ID             string           `json:"id"`
	Symbol         string           `json:"symbol"`
	Side           string           `json:"side"`
	PosSide        string           `json:"pos_side"`
	MarginType     string           `json:"margin_type"`
	Price          float64          `json:"price"`
	Size           float64          `json:"size"`
	Type           string           `json:"type"`
	Status         string           `json:"status"`
	Filled         float64          `json:"filled"`
	Remain         float64          `json:"remain"`
	OrderCondition []OrderCondition `json:"order_condition"`
}

type OrderCondition struct {
	TpTriggerPx          string `json:"tpTriggerPx,omitempty"`          // 止盈触发价
	TpOrdPx              string `json:"tpOrdPx,omitempty"`              // 止盈委托价
	TpOrdKind            string `json:"tpOrdKind,omitempty"`            // 止盈订单类型: condition(条件单), limit(限价单)
	SlTriggerPx          string `json:"slTriggerPx,omitempty"`          // 止损触发价
	SlOrdPx              string `json:"slOrdPx,omitempty"`              // 止损委托价
	TpTriggerPxType      string `json:"tpTriggerPxType,omitempty"`      // 止盈触发价类型: last(最新价格), index(指数价格), mark(标记价格)
	SlTriggerPxType      string `json:"slTriggerPxType,omitempty"`      // 止损触发价类型: last(最新价格), index(指数价格), mark(标记价格)
	Sz                   string `json:"sz,omitempty"`                   // 数量 (适用于"多笔止盈")
	AmendPxOnTriggerType string `json:"amendPxOnTriggerType,omitempty"` // 是否启用开仓价止损: "0"(不开启), "1"(开启)
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

type Symbol struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Base     string `json:"base"`
	Quote    string `json:"quote"`
	MaxLever string `json:"max_lever"`
	MinSize  string `json:"min_size"` // 最小下单（合约：张，现货：交易货币）
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

	// 设置用户信息
	SetUSer(ctx context.Context, user *models.FoxUser) error

	// 账户信息
	GetBalance(ctx context.Context) ([]Asset, error)
	GetPositions(ctx context.Context) ([]Position, error)

	// 订单管理
	GetOrders(ctx context.Context, symbol string, status string) ([]Order, error)
	CreateOrder(ctx context.Context, order *Order) (*Order, error)
	CancelOrder(ctx context.Context, order *Order) error

	// 行情数据
	GetTicker(ctx context.Context, symbol string) (*Ticker, error)
	GetTickers(ctx context.Context) ([]Ticker, error)

	// 标的配置
	GetSymbols(ctx context.Context, userSymbol string) (*Symbol, error)
	SetLeverage(ctx context.Context, symbol string, leverage int, marginType string) error
	SetMarginType(ctx context.Context, symbol string, marginType string) error

	// 币种名称转换
	ConvertToExchangeSymbol(userSymbol string) string
	ConvertFromExchangeSymbol(exchangeSymbol string) string
}
