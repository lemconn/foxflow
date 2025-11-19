package grpc

// ShowExchangeItem 交易所展示项
type ShowExchangeItem struct {
	Name        string `json:"name"`
	APIUrl      string `json:"api_url"`
	ProxyUrl    string `json:"proxy_url"`
	StatusValue int64  `json:"status_value"`
}

// ShowAssetItem 资产展示项
type ShowAssetItem struct {
	Currency  string `json:"currency"`
	Balance   string `json:"balance"`
	Available string `json:"available"`
	Frozen    string `json:"frozen"`
}

// ShowPositionItem 仓位展示项
type ShowPositionItem struct {
	Symbol     string `json:"symbol"`
	PosSide    string `json:"pos_side"`
	MarginType string `json:"margin_type"`
	Size       string `json:"size"`
	AvgPrice   string `json:"avg_price"`
	UnrealPnl  string `json:"unreal_pnl"`
}

// ShowSymbolItem 交易对展示项
type ShowSymbolItem struct {
	Exchange    string `json:"exchange"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	Volume      string `json:"volume"`
	High        string `json:"high"`
	Low         string `json:"low"`
	Base        string `json:"base"`
	Quote       string `json:"quote"`
	MaxLeverage int64  `json:"max_leverage"`
	MinSize     string `json:"min_size"`
	Contract    string `json:"contract"`
}

// ShowAccountItem 账户展示项
type ShowAccountItem struct {
	Id               int64  `json:"id"`
	Name             string `json:"name"`
	Exchange         string `json:"exchange"`
	TradeTypeValue   string `json:"trade_type_value"`
	StatusValue      int64  `json:"status_value"`
	IsolatedLeverage int64  `json:"isolated_leverage"`
	CrossLeverage    int64  `json:"cross_leverage"`
	ProxyUrl         string `json:"proxy_url"`
}

// ShowOrderItem 订单展示项
type ShowOrderItem struct {
	ID         int64  `json:"id"`          // 订单ID
	Exchange   string `json:"exchange"`    // 交易所
	AccountID  int64  `json:"account_id"`  // 账户ID
	Symbol     string `json:"symbol"`      // 交易对符号
	Side       string `json:"side"`        // 买卖方向 (buy/sell)
	PosSide    string `json:"pos_side"`    // 持仓方向 (long/short)
	MarginType string `json:"margin_type"` // 保证金类型 (isolated/cross)
	Price      string `json:"price"`       // 价格
	Size       string `json:"size"`        // 数量
	SizeType   string `json:"size_type"`   // 数量类型
	OrderType  string `json:"order_type"`  // 订单类型 (limit/market)
	Strategy   string `json:"strategy"`    // 策略名称
	OrderID    string `json:"order_id"`    // 交易所订单ID
	Type       string `json:"type"`        // 订单类型 (open/close)
	Status     string `json:"status"`      // 订单状态
	Msg        string `json:"msg"`         // 订单消息/描述
	CreatedAt  int64  `json:"created_at"`  // 创建时间
	UpdatedAt  int64  `json:"updated_at"`  // 更新时间
}
