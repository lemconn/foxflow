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

	"github.com/lemconn/foxflow/internal/models"
)

const (
	okxPublicUriInstruments = "/api/v5/public/instruments"
	okxUriSetLeverage       = "/api/v5/account/set-leverage"

	okxUriUserAssetValuation   = "/api/v5/asset/asset-valuation"
	okxUriUserBalance          = "/api/v5/account/balance"
	okxUriUserPositions        = "/api/v5/account/positions"
	okxUriUserCurrentPositions = "/api/v5/copytrading/current-subpositions"
	okxUriUserOrderPending     = "/api/v5/trade/orders-pending"
	okxUriUserTradeOrder       = "/api/v5/trade/order"
)

const (
	UserTradeTypeMock = "mock"
	UserTradeTypeLive = "live"
)

type okxResponse struct {
	Code         string      `json:"code"`
	Msg          string      `json:"msg"`
	ErrorCode    string      `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
	Data         interface{} `json:"data"`
}

type okxAssetValuation struct {
	TotalBal string `json:"totalBal"` // 账户总资产估值
	Ts       string `json:"ts"`       // 数据更新时间，Unix时间戳的毫秒数格式，如 1597026383085
	Details  struct {
		Earn    string `json:"earn"`    // 金融账户
		Funding string `json:"funding"` // 资金账户
		Trading string `json:"trading"` // 交易账户
	} `json:"details"` // 各个账户的资产估值
}

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

	// 这里可以添加连接测试逻辑（获取当前用户的资产估值数据）
	_, err := e.getAssetValuation(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (e *OKXExchange) Disconnect() error {
	e.user = nil
	return nil
}

func (e *OKXExchange) SetUSer(ctx context.Context, user *models.FoxUser) error {
	e.user = user
	return nil
}

// getAssetValuation 获取用户的资产估值（可以作为试探连接使用）
func (e *OKXExchange) getAssetValuation(ctx context.Context) (float64, error) {
	if e.user == nil || e.user.AccessKey == "" || e.user.SecretKey == "" || e.user.Passphrase == "" {
		return 0, fmt.Errorf("user information is missing, user: %+v ", e.user)
	}

	result, err := e.sendRequest(ctx, "GET", okxUriUserAssetValuation, nil)
	if err != nil {
		return 0, fmt.Errorf("okx getAssetValuation err: %w", err)
	}

	assetValuation := make([]okxAssetValuation, 0)
	resultBytes, _ := json.Marshal(result.Data)
	err = json.Unmarshal(resultBytes, &assetValuation)
	if err != nil {
		return 0, fmt.Errorf("okx getAssetValuation jsonDecode result err: %w", err)
	}

	floatNum, err := strconv.ParseFloat(assetValuation[0].TotalBal, 64)
	if err != nil {
		return 0, fmt.Errorf("okx getAssetValuation result ParseFloat err: %w", err)
	}

	return floatNum, nil
}

type okxAccountBalance struct {
	UTime                 string `json:"uTime"`                 // 账户信息的更新时间，Unix时间戳的毫秒数格式，如 1597026383085
	TotalEq               string `json:"totalEq"`               // 美金层面权益
	IsoEq                 string `json:"isoEq"`                 // 美金层面逐仓仓位权益。适用于合约模式/跨币种保证金模式/组合保证金模式
	AdjEq                 string `json:"adjEq"`                 // 美金层面有效保证金。适用于现货模式/跨币种保证金模式/组合保证金模式
	AvailEq               string `json:"availEq"`               // 账户美金层面可用保证金，排除因总质押借币上限而被限制的币种。适用于跨币种保证金模式/组合保证金模式
	OrdFroz               string `json:"ordFroz"`               // 美金层面全仓挂单占用保证金。仅适用于现货模式/跨币种保证金模式/组合保证金模式
	Imr                   string `json:"imr"`                   // 美金层面占用保证金。适用于现货模式/跨币种保证金模式/组合保证金模式
	Mmr                   string `json:"mmr"`                   // 美金层面维持保证金。适用于现货模式/跨币种保证金模式/组合保证金模式
	BorrowFroz            string `json:"borrowFroz"`            // 账户美金层面潜在借币占用保证金。仅适用于现货模式/跨币种保证金模式/组合保证金模式。在其他账户模式下为""。
	MgnRatio              string `json:"mgnRatio"`              // 美金层面维持保证金率。适用于现货模式/跨币种保证金模式/组合保证金模式
	NotionalUsd           string `json:"notionalUsd"`           // 以美金价值为单位的持仓数量，即仓位美金价值。适用于现货模式/跨币种保证金模式/组合保证金模式
	NotionalUsdForBorrow  string `json:"notionalUsdForBorrow"`  // 借币金额（美元价值）。适用于现货模式/跨币种保证金模式/组合保证金模式
	NotionalUsdForSwap    string `json:"notionalUsdForSwap"`    // 永续合约持仓美元价值。适用于跨币种保证金模式/组合保证金模式
	NotionalUsdForFutures string `json:"notionalUsdForFutures"` // 交割合约持仓美元价值。适用于跨币种保证金模式/组合保证金模式
	NotionalUsdForOption  string `json:"notionalUsdForOption"`  // 期权持仓美元价值。适用于现货模式/跨币种保证金模式/组合保证金模式
	Upl                   string `json:"upl"`                   // 账户层面全仓未实现盈亏（美元单位）。适用于跨币种保证金模式/组合保证金模式
	Details               []struct {
		Ccy                   string `json:"ccy"`                   // 币种
		Eq                    string `json:"eq"`                    // 币种总权益
		CashBal               string `json:"cashBal"`               // 币种余额
		UTime                 string `json:"uTime"`                 // 币种余额信息的更新时间，Unix时间戳的毫秒数格式，如 1597026383085
		IsoEq                 string `json:"isoEq"`                 // 币种逐仓仓位权益。适用于合约模式/跨币种保证金模式/组合保证金模式
		AvailEq               string `json:"availEq"`               // 可用保证金。适用于合约模式/跨币种保证金模式/组合保证金模式
		DisEq                 string `json:"disEq"`                 // 美金层面币种折算权益。适用于现货模式(开通了借币功能)/跨币种保证金模式/组合保证金模式
		FixedBal              string `json:"fixedBal"`              // 抄底宝、逃顶宝功能的币种冻结金额
		AvailBal              string `json:"availBal"`              // 可用余额
		FrozenBal             string `json:"frozenBal"`             // 币种占用金额
		OrdFrozen             string `json:"ordFrozen"`             // 挂单冻结数量。适用于现货模式/合约模式/跨币种保证金模式
		Liab                  string `json:"liab"`                  // 币种负债额。值为正数，如 "21625.64"。适用于现货模式/跨币种保证金模式/组合保证金模式
		Upl                   string `json:"upl"`                   // 未实现盈亏。适用于合约模式/跨币种保证金模式/组合保证金模式
		UplLiab               string `json:"uplLiab"`               // 由于仓位未实现亏损导致的负债。适用于跨币种保证金模式/组合保证金模式
		CrossLiab             string `json:"crossLiab"`             // 币种全仓负债额。适用于现货模式/跨币种保证金模式/组合保证金模式
		IsoLiab               string `json:"isoLiab"`               // 币种逐仓负债额。适用于跨币种保证金模式/组合保证金模式
		RewardBal             string `json:"rewardBal"`             // 体验金余额
		MgnRatio              string `json:"mgnRatio"`              // 币种全仓维持保证金率，衡量账户内某项资产风险的指标。适用于合约模式且有全仓仓位时
		Imr                   string `json:"imr"`                   // 币种维度全仓占用保证金。适用于合约模式且有全仓仓位时
		Mmr                   string `json:"mmr"`                   // 币种维度全仓维持保证金。适用于合约模式且有全仓仓位时
		Interest              string `json:"interest"`              // 计息，应扣未扣利息。值为正数，如 9.01。适用于现货模式/跨币种保证金模式/组合保证金模式
		Twap                  string `json:"twap"`                  // 当前负债币种触发自动换币的风险。0、1、2、3、4、5其中之一，数字越大代表您的负债币种触发自动换币概率越高。适用于现货模式/跨币种保证金模式/组合保证金模式
		FrpType               string `json:"frpType"`               // 自动换币类型。0：未发生自动换币；1：基于用户的自动换币；2：基于平台借币限额的自动换币。当twap>=1时返回1或2代表自动换币风险类型，适用于现货模式/跨币种保证金模式/组合保证金模式
		MaxLoan               string `json:"maxLoan"`               // 币种最大可借。适用于现货模式/跨币种保证金模式/组合保证金模式 的全仓
		EqUsd                 string `json:"eqUsd"`                 // 币种权益美金价值
		BorrowFroz            string `json:"borrowFroz"`            // 币种美金层面潜在借币占用保证金。仅适用于现货模式/跨币种保证金模式/组合保证金模式。在其他账户模式下为""。
		NotionalLever         string `json:"notionalLever"`         // 币种杠杆倍数。适用于合约模式
		StgyEq                string `json:"stgyEq"`                // 策略权益
		IsoUpl                string `json:"isoUpl"`                // 逐仓未实现盈亏。适用于合约模式/跨币种保证金模式/组合保证金模式
		SpotInUseAmt          string `json:"spotInUseAmt"`          // 现货对冲占用数量。适用于组合保证金模式
		ClSpotInUseAmt        string `json:"clSpotInUseAmt"`        // 用户自定义现货占用数量。适用于组合保证金模式
		MaxSpotInUse          string `json:"maxSpotInUse"`          // 系统计算得到的最大可能现货占用数量。适用于组合保证金模式
		SpotIsoBal            string `json:"spotIsoBal"`            // 现货逐仓余额。仅适用于现货带单/跟单。适用于现货模式/合约模式
		SmtSyncEq             string `json:"smtSyncEq"`             // 合约智能跟单权益。默认为0，仅适用于跟单人。
		SpotCopyTradingEq     string `json:"spotCopyTradingEq"`     // 现货智能跟单权益。默认为0，仅适用于跟单人。
		SpotBal               string `json:"spotBal"`               // 现货余额 ，单位为 币种，比如 BTC。
		OpenAvgPx             string `json:"openAvgPx"`             // 现货开仓成本价 单位 USD。
		AccAvgPx              string `json:"accAvgPx"`              // 现货累计成本价 单位 USD。
		SpotUpl               string `json:"spotUpl"`               // 现货未实现收益，单位 USD。
		SpotUplRatio          string `json:"spotUplRatio"`          // 现货未实现收益率。
		TotalPnl              string `json:"totalPnl"`              // 现货累计收益，单位 USD。
		TotalPnlRatio         string `json:"totalPnlRatio"`         // 现货累计收益率。
		ColRes                string `json:"colRes"`                // 平台维度质押限制状态。0：限制未触发；1：限制未触发，但该币种接近平台质押上限；2：限制已触发。该币种不可用作新订单的保证金，这可能会导致下单失败。但它仍会被计入账户有效保证金，保证金率不会收到影响。
		ColBorrAutoConversion string `json:"colBorrAutoConversion"` // 基于平台质押借币限额的自动换币风险指标。分为1-5多个等级，数字越大，触发自动换币的可能性越大。默认值为0，表示当前无风险。5表示该用户正在进行自动换币，4代表该用户即将被进行自动换币，1/2/3表示存在自动换币风险。适用于现货模式/合约模式/跨币种保证金模式/组合保证金模式
		CollateralEnabled     bool   `json:"collateralEnabled"`     // true：质押币；false：非质押币。适用于跨币种保证金模式
		AutoLendStatus        string `json:"autoLendStatus"`        // 自动借出状态。unsupported：该币种不支持自动借出；off：自动借出功能关闭；pending：自动借出功能开启但未匹配；active：自动借出功能开启且已匹配
		AutoLendMtAmt         string `json:"autoLendMtAmt"`         // 自动借出已匹配量。当 autoLendStatus 为 unsupported/off/pending 时返回 0；当 autoLendStatus 为 active 时返回已匹配量
	} `json:"details"` // 各币种资产详细信息
}

func (e *OKXExchange) GetBalance(ctx context.Context) ([]Asset, error) {
	if e.user == nil || e.user.AccessKey == "" || e.user.SecretKey == "" || e.user.Passphrase == "" {
		return nil, fmt.Errorf("user information is missing, user: %+v ", e.user)
	}

	result, err := e.sendRequest(ctx, "GET", okxUriUserBalance, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if result.Code != "0" {
		return nil, fmt.Errorf("okx GetBalance error: %s", result.Msg)
	}

	accountBalance := make([]okxAccountBalance, 0)
	resultBytes, _ := json.Marshal(result.Data)
	err = json.Unmarshal(resultBytes, &accountBalance)
	if err != nil {
		return nil, fmt.Errorf("accountBalance json.Decode err: %w", err)
	}

	if len(accountBalance) == 0 || len(accountBalance[0].Details) == 0 {
		return []Asset{}, nil
	}

	var assets []Asset
	for _, detail := range accountBalance[0].Details {
		balance, _ := strconv.ParseFloat(detail.CashBal, 64)
		frozen, _ := strconv.ParseFloat(detail.FrozenBal, 64)
		available, _ := strconv.ParseFloat(detail.AvailBal, 64)

		assets = append(assets, Asset{
			Currency:  detail.Ccy,
			Balance:   balance,
			Frozen:    frozen,
			Available: available,
		})
	}

	return assets, nil
}

func (e *OKXExchange) GetPositions(ctx context.Context) ([]Position, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	return nil, nil
}

func (e *OKXExchange) GetOrders(ctx context.Context, symbol string, status string) ([]Order, error) {
	if e.user == nil {
		return nil, fmt.Errorf("user not connected")
	}

	return nil, nil
}

// oxkOrderRequest 主订单结构体
type oxkOrderRequest struct {
	InstID         string             `json:"instId"`                   // 产品ID，如 BTC-USDT
	TdMode         string             `json:"tdMode"`                   // 交易模式: isolated(逐仓), cross(全仓), cash(非保证金), spot_isolated(现货逐仓)
	Ccy            string             `json:"ccy,omitempty"`            // 保证金币种，适用于逐仓杠杆及合约模式下的全仓杠杆订单
	ClOrdID        string             `json:"clOrdId,omitempty"`        // 客户自定义订单ID (1-32位)
	Tag            string             `json:"tag,omitempty"`            // 订单标签 (1-16位)
	Side           string             `json:"side"`                     // 订单方向: buy(买), sell(卖)
	PosSide        string             `json:"posSide,omitempty"`        // 持仓方向: long(多), short(空) (开平仓模式必填)
	OrdType        string             `json:"ordType"`                  // 订单类型: market, limit, post_only, fok, ioc, optimal_limit_ioc, mmp, mmp_and_post_only
	Sz             string             `json:"sz"`                       // 委托数量
	Px             string             `json:"px,omitempty"`             // 委托价格 (限价单等类型需要)
	PxUsd          string             `json:"pxUsd,omitempty"`          // 以USD价格进行期权下单
	PxVol          string             `json:"pxVol,omitempty"`          // 以隐含波动率进行期权下单
	ReduceOnly     bool               `json:"reduceOnly,omitempty"`     // 是否只减仓: true 或 false
	TgtCcy         string             `json:"tgtCcy,omitempty"`         // 市价单委托数量单位: base_ccy(交易货币), quote_ccy(计价货币)
	BanAmend       bool               `json:"banAmend,omitempty"`       // 是否禁止币币市价改单: true 或 false
	PxAmendType    string             `json:"pxAmendType,omitempty"`    // 订单价格修正类型: "0"(不允许修改), "1"(允许修改)
	TradeQuoteCcy  string             `json:"tradeQuoteCcy,omitempty"`  // 用于交易的计价币种
	StpMode        string             `json:"stpMode,omitempty"`        // 自成交保护模式: cancel_maker, cancel_taker, cancel_both
	AttachAlgoOrds []oxkAttachAlgoOrd `json:"attachAlgoOrds,omitempty"` // 下单附带止盈止损信息数组
}

// 止盈止损附加订单结构体
type oxkAttachAlgoOrd struct {
	AttachAlgoClOrdId    string `json:"attachAlgoClOrdId,omitempty"`    // 客户自定义的策略订单ID
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

func (e *OKXExchange) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
	if e.user == nil || e.user.AccessKey == "" || e.user.SecretKey == "" || e.user.Passphrase == "" {
		return nil, fmt.Errorf("user information is missing, user: %+v ", e.user)
	}

	reqBody := oxkOrderRequest{
		InstID:        e.ConvertToExchangeSymbol(order.Symbol),
		TdMode:        order.MarginType,
		Side:          order.Side,
		OrdType:       order.Type,
		TradeQuoteCcy: "USDT", // tradeQuoteCcy 对于特定国家和地区的用户，下单成功需要填写该参数，否则会取 `instId` 的计价币种为默认值，报错 51000。
	}

	// 按类型填充价格与数量
	if strings.ToLower(order.Type) == "limit" {
		reqBody.Px = fmt.Sprintf("%f", order.Price)
	}
	// OKX合约下单数量字段为 sz，单位张。此处直接使用传入数量
	reqBody.Sz = fmt.Sprintf("%f", order.Size)

	// 增加订单条件
	if len(order.OrderCondition) > 0 {
		for _, cond := range order.OrderCondition {
			reqBody.AttachAlgoOrds = append(reqBody.AttachAlgoOrds, oxkAttachAlgoOrd{
				TpTriggerPx:          cond.TpTriggerPx,
				TpOrdPx:              cond.TpOrdPx,
				TpOrdKind:            cond.TpOrdKind,
				SlTriggerPx:          cond.SlTriggerPx,
				SlOrdPx:              cond.SlOrdPx,
				TpTriggerPxType:      cond.TpTriggerPxType,
				SlTriggerPxType:      cond.SlTriggerPxType,
				Sz:                   cond.Sz,
				AmendPxOnTriggerType: cond.AmendPxOnTriggerType,
			})
		}
	}

	reqBodyByte, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	body := make(map[string]interface{})
	err = json.Unmarshal(reqBodyByte, &body)
	if err != nil {
		return nil, err
	}

	result, err := e.sendRequest(ctx, "POST", okxUriUserTradeOrder, body)
	if err != nil {
		return nil, fmt.Errorf("okx create order err: %w", err)
	}
	if result.Code != "0" {
		return nil, fmt.Errorf("okx create order error: %s", result.Msg)
	}

	// 解析返回，写回订单ID
	var dataArr []map[string]interface{}
	bytes, _ := json.Marshal(result.Data)
	if err := json.Unmarshal(bytes, &dataArr); err != nil {
		return nil, fmt.Errorf("okx create order json decode err: %w", err)
	}
	if len(dataArr) == 0 {
		return nil, fmt.Errorf("okx create order empty data")
	}
	if val, ok := dataArr[0]["ordId"].(string); ok {
		order.ID = val
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

	return nil, fmt.Errorf("no ticker data found")
}

func (e *OKXExchange) GetTickers(ctx context.Context) ([]Ticker, error) {

	return nil, nil
}

type okxSymbol struct {
	InstType          string   `json:"instType,omitempty"`          // 产品类型
	InstId            string   `json:"instId,omitempty"`            // 产品ID，如 BTC-USDT
	Uly               string   `json:"uly,omitempty"`               // 标的指数，如 BTC-USD，仅适用于杠杆/交割/永续/期权
	InstFamily        string   `json:"instFamily,omitempty"`        // 交易品种，如 BTC-USD，仅适用于杠杆/交割/永续/期权
	BaseCcy           string   `json:"baseCcy,omitempty"`           // 交易货币币种，如 BTC-USDT 中的 BTC，仅适用于币币/币币杠杆
	QuoteCcy          string   `json:"quoteCcy,omitempty"`          // 计价货币币种，如 BTC-USDT 中的USDT，仅适用于币币/币币杠杆
	SettleCcy         string   `json:"settleCcy,omitempty"`         // 盈亏结算和保证金币种，如 BTC，仅适用于交割/永续/期权
	CtVal             string   `json:"ctVal,omitempty"`             // 合约面值，仅适用于交割/永续/期权
	CtMult            string   `json:"ctMult,omitempty"`            // 合约乘数，仅适用于交割/永续/期权
	CtValCcy          string   `json:"ctValCcy,omitempty"`          // 合约面值计价币种，仅适用于交割/永续/期权
	OptType           string   `json:"optType,omitempty"`           // 期权类型，C或P，仅适用于期权
	Stk               string   `json:"stk,omitempty"`               // 行权价格，仅适用于期权
	ListTime          string   `json:"listTime,omitempty"`          // 上线时间，Unix时间戳的毫秒数格式，如 1597026383085
	AuctionEndTime    string   `json:"auctionEndTime,omitempty"`    // 集合竞价结束时间，Unix时间戳的毫秒数格式，如 1597026383085（已废弃，请使用contTdSwTime）
	ContTdSwTime      string   `json:"contTdSwTime,omitempty"`      // 连续交易开始时间，从集合竞价、提前挂单切换到连续交易的时间，Unix时间戳格式，单位为毫秒，仅适用于通过集合竞价或提前挂单上线的SPOT/MARGIN
	PreMktSwTime      string   `json:"preMktSwTime,omitempty"`      // 盘前永续合约转为普通永续合约的时间，Unix时间戳的毫秒数格式，如 1597026383085，仅适用于盘前SWAP
	OpenType          string   `json:"openType,omitempty"`          // 开盘类型：fix_price（定价开盘）、pre_quote（提前挂单）、call_auction（集合竞价），只适用于SPOT/MARGIN
	ExpTime           string   `json:"expTime,omitempty"`           // 产品下线时间，适用于币币/杠杆/交割/永续/期权
	Lever             string   `json:"lever,omitempty"`             // 该instId支持的最大杠杆倍数，不适用于币币、期权
	TickSz            string   `json:"tickSz,omitempty"`            // 下单价格精度，如 0.0001
	LotSz             string   `json:"lotSz,omitempty"`             // 下单数量精度。合约的数量单位是张，现货的数量单位是交易货币
	MinSz             string   `json:"minSz,omitempty"`             // 最小下单数量。合约的数量单位是张，现货的数量单位是交易货币
	CtType            string   `json:"ctType,omitempty"`            // 合约类型：linear（正向合约）、inverse（反向合约），仅适用于交割/永续
	Alias             string   `json:"alias,omitempty"`             // 合约日期别名（如this_week、next_week等），仅适用于交割，不建议使用
	State             string   `json:"state,omitempty"`             // 产品状态：live（交易中）、suspend（暂停中）、preopen（预上线）、test（测试中）
	RuleType          string   `json:"ruleType,omitempty"`          // 交易规则类型：normal（普通交易）、pre_market（盘前交易）
	MaxLmtSz          string   `json:"maxLmtSz,omitempty"`          // 限价单的单笔最大委托数量。合约单位张，现货单位交易货币
	MaxMktSz          string   `json:"maxMktSz,omitempty"`          // 市价单的单笔最大委托数量。合约单位张，现货单位USDT
	MaxLmtAmt         string   `json:"maxLmtAmt,omitempty"`         // 限价单的单笔最大美元价值，仅适用于币币/币币杠杆
	MaxMktAmt         string   `json:"maxMktAmt,omitempty"`         // 市价单的单笔最大美元价值，仅适用于币币/币币杠杆
	MaxTwapSz         string   `json:"maxTwapSz,omitempty"`         // 时间加权单的单笔最大委托数量。合约单位张，现货单位交易货币
	MaxIcebergSz      string   `json:"maxIcebergSz,omitempty"`      // 冰山委托的单笔最大委托数量。合约单位张，现货单位交易货币
	MaxTriggerSz      string   `json:"maxTriggerSz,omitempty"`      // 计划委托的单笔最大委托数量。合约单位张，现货单位交易货币
	MaxStopSz         string   `json:"maxStopSz,omitempty"`         // 止盈止损市价委托的单笔最大委托数量。合约单位张，现货单位USDT
	FutureSettlement  bool     `json:"futureSettlement,omitempty"`  // 交割合约是否支持每日结算，适用于全仓交割
	TradeQuoteCcyList []string `json:"tradeQuoteCcyList,omitempty"` // 可用于交易的计价币种列表，如 ["USD", "USDC"]
	InstIdCode        int64    `json:"instIdCode,omitempty"`        // 产品唯一标识代码
}

func (e *OKXExchange) GetSymbols(ctx context.Context, userSymbol string) (*Symbol, error) {

	exchangeSymbol := e.ConvertToExchangeSymbol(userSymbol)

	// 使用OKX公共接口获取交易对信息
	params := url.Values{}
	params.Set("instType", "SWAP")
	params.Set("instId", exchangeSymbol)

	result, err := e.sendRequest(ctx, "GET", fmt.Sprintf("%s?%s", okxPublicUriInstruments, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get instruments [%s] err: %w", userSymbol, err)
	}

	if result.Code != "0" {
		return nil, fmt.Errorf("okx GetSymbols [%s] error: %s", userSymbol, result.Msg)
	}

	okxSymbolInfos := make([]okxSymbol, 0)
	resultBytes, _ := json.Marshal(result.Data)
	err = json.Unmarshal(resultBytes, &okxSymbolInfos)
	if err != nil {
		return nil, fmt.Errorf("symbols [%s] json.Decode err: %w", userSymbol, err)
	}

	symbol := &Symbol{
		Type:     okxSymbolInfos[0].InstType,
		Name:     okxSymbolInfos[0].InstId,
		Base:     okxSymbolInfos[0].BaseCcy,
		Quote:    okxSymbolInfos[0].QuoteCcy,
		MaxLever: okxSymbolInfos[0].Lever,
		MinSize:  okxSymbolInfos[0].MinSz,
	}

	return symbol, nil
}

var marginTypeMap = map[string]string{
	"isolated": "isolated",
	"cross":    "cross",
}

type okxSetLeverageBody struct {
	InstId  string `json:"instId"`
	Lever   string `json:"lever"`
	MgnMode string `json:"mgnMode"`
	Ccy     string `json:"ccy"`
	PosSide string `json:"posSide"`
}

func (e *OKXExchange) SetLeverage(ctx context.Context, symbol string, leverage int, marginType string) error {
	if e.user == nil || e.user.AccessKey == "" || e.user.SecretKey == "" || e.user.Passphrase == "" {
		return fmt.Errorf("user information is missing, user: %+v ", e.user)
	}

	reqBody := okxSetLeverageBody{
		InstId:  e.ConvertToExchangeSymbol(symbol),
		Lever:   strconv.Itoa(leverage),
		MgnMode: marginTypeMap[marginType],
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	var bodyMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &bodyMap)
	if err != nil {
		return err
	}

	result, err := e.sendRequest(ctx, "POST", okxUriSetLeverage, bodyMap)
	if err != nil {
		return fmt.Errorf("failed to set leverage [%s] err: %w", bodyMap, err)
	}

	if result.Code != "0" {
		return fmt.Errorf("okx SetLeverage [%s] error: %s", bodyMap, result.Msg)
	}

	resultData := make([]okxSetLeverageBody, 0)
	resultByte, _ := json.Marshal(result.Data)
	err = json.Unmarshal(resultByte, &resultData)
	if err != nil {
		return fmt.Errorf("failed to set leverage [%s] json.Decode err: %w", string(resultByte), err)
	}

	if len(resultData) == 0 || resultData[0].Lever != reqBody.Lever {
		return fmt.Errorf("set lever exception, resultData: %+v", resultData)
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

// ConvertToExchangeSymbol 将用户输入的币种名称转换为OKX交易所格式
// 例如：BTC -> BTC-USDT-SWAP
func (e *OKXExchange) ConvertToExchangeSymbol(userSymbol string) string {
	// OKX期货使用币种-USDT-SWAP的格式
	return userSymbol + "-USDT-SWAP"
}

// ConvertFromExchangeSymbol 将OKX交易所格式的币种名称转换为用户格式
// 例如：BTC-USDT-SWAP -> BTC
func (e *OKXExchange) ConvertFromExchangeSymbol(exchangeSymbol string) string {
	// 去除-USDT-SWAP后缀
	if len(exchangeSymbol) > 10 && exchangeSymbol[len(exchangeSymbol)-10:] == "-USDT-SWAP" {
		return exchangeSymbol[:len(exchangeSymbol)-10]
	}
	// 去除-USD-SWAP后缀
	if len(exchangeSymbol) > 9 && exchangeSymbol[len(exchangeSymbol)-9:] == "-USD-SWAP" {
		return exchangeSymbol[:len(exchangeSymbol)-9]
	}
	// 去除-USDT后缀（现货格式）
	if len(exchangeSymbol) > 5 && exchangeSymbol[len(exchangeSymbol)-5:] == "-USDT" {
		return exchangeSymbol[:len(exchangeSymbol)-5]
	}
	// 如果没有匹配的后缀，返回原符号
	return exchangeSymbol
}

// 辅助方法：构建请求头
func (e *OKXExchange) buildHeaders(method, uri, body string) map[string]string {
	if e.user == nil {
		return map[string]string{
			"Content-Type":        "application/json",
			"x-simulated-trading": "0",
		}
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if e.user.TradeType == UserTradeTypeMock {
		headers["x-simulated-trading"] = "1"
	} else { // 默认为实盘
		headers["x-simulated-trading"] = "0"
	}

	// 用户的信息中未设置ak/sk 表示为公共接口
	if e.user.AccessKey == "" || e.user.SecretKey == "" {
		return headers
	}

	if e.user.AccessKey != "" {
		headers["OK-ACCESS-KEY"] = e.user.AccessKey
	}
	if e.user.Passphrase != "" {
		headers["OK-ACCESS-PASSPHRASE"] = e.user.Passphrase
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	message := timestamp + strings.ToUpper(method) + uri + body
	signature := e.buildSignature(message)

	headers["OK-ACCESS-SIGN"] = signature
	headers["OK-ACCESS-TIMESTAMP"] = timestamp

	return headers
}

// 辅助方法：构建OKX签名
func (e *OKXExchange) buildSignature(message string) string {
	h := hmac.New(sha256.New, []byte(e.user.SecretKey))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// 辅助方法：发送HTTP请求
func (e *OKXExchange) sendRequest(ctx context.Context, method, uri string, params map[string]interface{}) (*okxResponse, error) {
	// 构建完整URL
	fullURL := e.apiURL + uri

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
	headers := e.buildHeaders(method, uri, string(body))
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
	var result okxResponse
	if err = json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
