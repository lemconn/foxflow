package render

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/news"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"github.com/lemconn/foxflow/internal/utils"
	"github.com/shopspring/decimal"
)

// RenderExchangesWithStatus 渲染带状态的交易所列表
func RenderExchangesWithStatus(exchanges []*model.FoxExchange) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("可用交易所")
	pt.SetHeaders([]interface{}{"交易所名称", "状态"})

	for _, exchange := range exchanges {
		status := "非活跃"
		if exchange.IsActive == 1 {
			status = "激活"
		}

		pt.AddRow([]interface{}{
			exchange.Name,
			status,
		})
	}

	return pt.Render()
}

// RenderAccounts 渲染用户列表
func RenderAccounts(accounts []*model.FoxAccount) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("用户列表")
	pt.SetHeaders([]interface{}{"用户名", "交易所", "交易类型", "状态", "杠杆倍数", "代理地址"})

	for _, account := range accounts {
		status := "非活跃"
		if account.IsActive == 1 {
			status = "激活"
		}

		tradeType := "模拟"
		if account.TradeType == "live" {
			tradeType = "实盘"
		}

		leverSlice := make([]string, 0)
		for _, tradeConfig := range account.TradeConfigs {
			if tradeConfig.Margin == "cross" {
				leverSlice = append(leverSlice, fmt.Sprintf("%d(%s)", tradeConfig.Leverage, "全仓"))
			} else if tradeConfig.Margin == "isolated" {
				leverSlice = append(leverSlice, fmt.Sprintf("%d(%s)", tradeConfig.Leverage, "逐仓"))
			}
		}

		pt.AddRow([]interface{}{
			account.Name,
			account.Exchange,
			tradeType,
			status,
			strings.Join(leverSlice, " / "),
			account.Config.ProxyURL,
		})
	}

	return pt.Render()
}

// RenderAssets 渲染资产列表
func RenderAssets(assets []exchange.Asset) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("资产列表")
	pt.SetHeaders([]interface{}{"币种", "总余额", "可用余额", "冻结余额"})

	for _, asset := range assets {
		pt.AddRow([]interface{}{
			asset.Currency,
			fmt.Sprintf("%.4f", asset.Balance),
			fmt.Sprintf("%.4f", asset.Available),
			fmt.Sprintf("%.4f", asset.Frozen),
		})
	}

	return pt.Render()
}

// RenderPositions 渲染仓位列表
func RenderPositions(positions []exchange.Position) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("仓位列表")
	pt.SetHeaders([]interface{}{"交易对", "仓位方向", "保证金模式", "数量", "均价", "未实现盈亏"})

	for _, pos := range positions {
		var margin string
		if pos.MarginType == "isolated" {
			margin = fmt.Sprintf("%s（逐仓）", pos.MarginType)
		} else if pos.MarginType == "cross" {
			margin = fmt.Sprintf("%s（全仓）", pos.MarginType)
		}

		pt.AddRow([]interface{}{
			pos.Symbol,
			pos.PosSide,
			margin,
			fmt.Sprintf("%.4f", pos.Size),
			fmt.Sprintf("%.2f", pos.AvgPrice),
			fmt.Sprintf("%.2f", pos.UnrealPnl),
		})
	}

	return pt.Render()
}

// RenderStrategies 渲染策略列表
func RenderStrategies() string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("可用策略")
	pt.SetHeaders([]interface{}{"策略名称", "描述", "参数"})

	strategies := []struct {
		Name        string
		Description string
		Parameters  string
	}{
		{"volume", "成交量策略", "threshold: 成交量阈值"},
		{"macd", "MACD策略", "threshold: MACD阈值"},
		{"rsi", "RSI策略", "threshold: RSI阈值 (0-100)"},
	}

	for _, strategy := range strategies {
		pt.AddRow([]interface{}{
			strategy.Name,
			strategy.Description,
			strategy.Parameters,
		})
	}

	return pt.Render()
}

type RenderSymbolsInfo struct {
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

// RenderSymbols 渲染交易对列表
func RenderSymbols(symbols []RenderSymbolsInfo) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("💱 交易对列表")
	pt.SetHeaders([]interface{}{"#", "交易对", "最新价格", "24小时最高价", "24小时最低价", "24小时成交量（标的）", "最大杠杆倍数", "最小下单张数", "最小下单标的数量"})

	for i, symbol := range symbols {
		symbolContract, _ := decimal.NewFromString(symbol.Contract)
		symbolMinSize, _ := decimal.NewFromString(symbol.MinSize)

		pt.AddRow([]interface{}{
			i + 1,
			symbol.Name,
			symbol.Price,
			symbol.High,
			symbol.Low,
			symbol.Volume,
			symbol.MaxLeverage,
			symbol.MinSize,
			symbolContract.Mul(symbolMinSize),
		})
	}

	return pt.Render()
}

// RenderStrategyOrders 渲染策略订单列表
func RenderStrategyOrders(orders []*model.FoxOrder) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("策略订单列表")
	pt.SetHeaders([]interface{}{"ID", "交易对", "方向", "仓位", "数量/金额", "价格", "状态", "策略", "异常结果"})

	for _, order := range orders {

		fmt.Printf("---------[%+v]--------\n", order)

		side := ""
		if order.Side == "buy" {
			side = fmt.Sprintf("%s(买入)", order.Side)
		} else if order.Side == "sell" {
			side = fmt.Sprintf("%s(卖出)", order.Side)
		}

		posSide := ""
		if order.PosSide == "long" {
			posSide = fmt.Sprintf("%s(多头)", order.PosSide)
		} else if order.PosSide == "short" {
			posSide = fmt.Sprintf("%s(空头)", order.PosSide)
		}

		status := "等待中"
		switch order.Status {
		case "opened":
			status = "开仓成功"
		case "closed":
			status = "平仓成功"
		case "cancelled":
			status = "已取消"
		case "failed":
			status = "失败"
		}

		var amount string
		switch order.SizeType {
		case "USDT":
			amount = fmt.Sprintf("%sU", order.Size)
		default:
			amount = order.Size
		}

		price := "-"
		if order.Price != "" {
			orderPriceDecimal := decimal.RequireFromString(order.Price)
			if orderPriceDecimal.GreaterThan(decimal.Zero) {
				price = order.Price
			}
		}

		strategy := "-"
		if len(order.Strategy) > 0 {
			strategy = order.Strategy
		}

		msg := "-"
		if len(order.Msg) > 0 {
			msg = order.Msg
		}

		pt.AddRow([]interface{}{
			order.OrderID,
			order.Symbol,
			side,
			posSide,
			amount,
			price,
			status,
			strategy,
			msg,
		})
	}

	return pt.Render()
}

// RenderNews 渲染新闻列表
func RenderNews(newsList []news.NewsItem) string {
	if len(newsList) == 0 {
		return utils.RenderWarning("暂无新闻数据")
	}

	// 按时间正序排列（最新的在下面）
	// 由于 newsList 已经是按时间倒序排列的，我们需要反转它
	reversedList := make([]news.NewsItem, len(newsList))
	for i, item := range newsList {
		reversedList[len(newsList)-1-i] = item
	}

	// 使用表格格式显示
	pt := utils.NewPrettyTable()
	pt.SetTitle(fmt.Sprintf("📰 最新新闻 (共 %d 条)", len(newsList)))
	pt.SetHeaders([]interface{}{"#", "标题", "时间", "来源", "链接"})

	// 设置列配置：优化列宽和对齐
	pt.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: 4, Align: text.AlignCenter},  // 序号列，居中对齐
		{Number: 2, WidthMax: 60, Align: text.AlignLeft},   // 标题列，左对齐，增加宽度
		{Number: 3, WidthMax: 12, Align: text.AlignCenter}, // 时间列，居中对齐
		{Number: 4, WidthMax: 12, Align: text.AlignCenter}, // 来源列，居中对齐
		{Number: 5, WidthMax: 60, Align: text.AlignLeft},   // 链接列，左对齐，增加宽度
	})

	for i, item := range reversedList {
		pt.AddRow([]interface{}{
			i + 1,
			item.Title, // 显示完整标题，不截断
			item.PublishedAt.Format("01-02 15:04"),
			item.Source,
			item.URL,
		})
	}

	return pt.Render()
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
