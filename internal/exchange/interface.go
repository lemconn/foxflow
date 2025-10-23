package exchange

import (
	"context"
	"time"

	"github.com/lemconn/foxflow/internal/models"
)

const (
	// 保证金模式
	MarginTypeCross    = "cross"    // 全仓
	MarginTypeIsolated = "isolated" // 逐仓
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
	TpTriggerPx          string  `json:"tpTriggerPx,omitempty"`          // 止盈触发价
	TpOrdPx              string  `json:"tpOrdPx,omitempty"`              // 止盈委托价
	TpOrdKind            string  `json:"tpOrdKind,omitempty"`            // 止盈订单类型: condition(条件单), limit(限价单)
	SlTriggerPx          string  `json:"slTriggerPx,omitempty"`          // 止损触发价
	SlOrdPx              string  `json:"slOrdPx,omitempty"`              // 止损委托价
	TpTriggerPxType      string  `json:"tpTriggerPxType,omitempty"`      // 止盈触发价类型: last(最新价格), index(指数价格), mark(标记价格)
	SlTriggerPxType      string  `json:"slTriggerPxType,omitempty"`      // 止损触发价类型: last(最新价格), index(指数价格), mark(标记价格)
	Size                 float64 `json:"size,omitempty"`                 // 数量 (适用于"多笔止盈")
	AmendPxOnTriggerType string  `json:"amendPxOnTriggerType,omitempty"` // 是否启用开仓价止损: "0"(不开启), "1"(开启)
}

type OrderCostReq struct {
	Symbol     string  `json:"symbol"`      // 标的
	Side       string  `json:"side"`        // 方向
	Amount     float64 `json:"amount"`      // 购买数量（标的数量）
	AmountType string  `json:"amount_type"` // 数量类型：coin(标的数量) / usdt(USDT数量)
	MarginType string  `json:"margin_type"` // 保证金模式：isolated(逐仓) / cross(全仓)
	LimitPrice float64 `json:"limit_price"` // 限价（目前不需要传递）
}

type OrderCostResp struct {
	Symbol          string  `json:"symbol"`             // 标的
	MarkPrice       float64 `json:"mark_price"`         // 标的价格
	MarginType      string  `json:"margin_type"`        // 保证金模式：isolated(逐仓) / cross(全仓)
	Lever           float64 `json:"lever"`              // 标的杠杆倍数
	Contracts       float64 `json:"contracts"`          // 购买张数
	AvailableFunds  float64 `json:"available_funds"`    // 可用资金
	MarginRequired  float64 `json:"margin_required"`    // 保证金
	Fee             float64 `json:"fee"`                // 手续费
	TotalRequired   float64 `json:"total_required"`     // 需求总资金
	CanBuyWithTaker bool    `json:"can_buy_with_taker"` // 是否可以购买
}

// Position 仓位信息
type Position struct {
	Symbol     string  `json:"symbol"`
	PosSide    string  `json:"pos_side"`
	MarginType string  `json:"margin_type"` // 保证金模式：isolated(逐仓) / cross(全仓)
	Size       float64 `json:"size"`
	AvgPrice   float64 `json:"avg_price"`
	UnrealPnl  float64 `json:"unreal_pnl"`
}

// Asset 资产信息
type Asset struct {
	Currency  string  `json:"currency"`
	Balance   float64 `json:"balance"`
	Frozen    float64 `json:"frozen"`
	Available float64 `json:"available"`
}

type Symbol struct {
	Type          string  `json:"type"`
	Name          string  `json:"name"`
	Base          string  `json:"base"`
	Quote         string  `json:"quote"`
	MaxLever      float64 `json:"max_lever"`
	MinSize       float64 `json:"min_size"`       // 最小下单（合约：张，现货：交易货币）
	ContractValue float64 `json:"contract_value"` // 张/标的的换算单位（1张=0.01个BTC，这里是0.01）
}

// Ticker 行情信息
type Ticker struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
}

// KlineData K线数据
type KlineData struct {
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

type SymbolLeverageMarginType struct {
	Symbol  string  `json:"symbol"`
	PosSide string  `json:"posSide"`
	Lever   float64 `json:"lever"`
	Margin  string  `json:"margin"`
}

// Exchange 交易所接口
type Exchange interface {
	// 基础信息
	GetName() string
	GetAPIURL() string
	GetProxyURL() string

	// 连接管理
	Connect(ctx context.Context, account *models.FoxAccount) error
	Disconnect() error

	// 设置用户信息
	SetAccount(ctx context.Context, account *models.FoxAccount) error
	GetAccount(ctx context.Context) (*models.FoxAccount, error)

	// 账户信息
	GetBalance(ctx context.Context) ([]Asset, error)
	GetPositions(ctx context.Context) ([]Position, error)

	// 订单管理
	GetOrders(ctx context.Context, symbol string, status string) ([]Order, error)
	CreateOrder(ctx context.Context, order *Order) (*Order, error)
	CancelOrder(ctx context.Context, order *Order) error
	CalcOrderCost(ctx context.Context, req *OrderCostReq) (*OrderCostResp, error) // 计算order成本（手续费+可买价格，是否可成交等等）

	// 行情数据
	GetTicker(ctx context.Context, symbol string) (*Ticker, error)
	GetTickers(ctx context.Context) ([]Ticker, error)

	// 标的配置
	GetSymbols(ctx context.Context, symbol string) (*Symbol, error)
	GetAllSymbols(ctx context.Context, instType string) ([]Symbol, error)
	SetLeverage(ctx context.Context, symbol string, leverage int, marginType string) error
	SetMarginType(ctx context.Context, symbol string, marginType string) error
	GetLeverageMarginType(ctx context.Context, margin, symbol string) ([]SymbolLeverageMarginType, error)
	GetSizeByQuote(ctx context.Context, symbol string, amount float64) (float64, error) // 根据报价数量获取可买张数
	GetSizeByBase(ctx context.Context, symbol string, amount float64) (float64, error)  // 根据标的数量获取可买张数

	// 币种名称转换
	ConvertToExchangeSymbol(accountSymbol string) string
	ConvertFromExchangeSymbol(exchangeSymbol string) string

	// K线数据
	GetKlineData(ctx context.Context, symbol, interval string, limit int) ([]KlineData, error)

	// 币种名称转换
	GetSwapSymbolByName(ctx context.Context, coinName string) string

	// 时间间隔格式转换
	ConvertIntervalFormat(interval string) string
}
