package grpc

// ShowAccountItem 账户展示项
type ShowAccountItem struct {
	Id               int    `json:"id"`
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
