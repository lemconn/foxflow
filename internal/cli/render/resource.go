package render

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/models"
	"github.com/lemconn/foxflow/internal/news"
	"github.com/lemconn/foxflow/internal/utils"
)

// RenderExchangesWithStatus 渲染带状态的交易所列表
func RenderExchangesWithStatus(exchanges []*models.FoxExchange) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("🏦 可用交易所")
	pt.SetHeaders([]interface{}{"#", "交易所名称", "状态"})

	for i, exchange := range exchanges {
		status := "❌ 非活跃"
		if exchange.IsActive {
			status = "✅ 激活"
		}

		pt.AddRow([]interface{}{
			i + 1,
			exchange.Name,
			status,
		})
	}

	return pt.Render()
}

// RenderUsers 渲染用户列表
func RenderUsers(users []models.FoxUser) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("👥 用户列表")
	pt.SetHeaders([]interface{}{"ID", "用户名", "交易所", "交易类型", "状态"})

	for _, user := range users {
		status := "❌ 非活跃"
		if user.IsActive {
			status = "✅ 激活"
		}

		tradeType := "🎯 模拟"
		if user.TradeType == "real" {
			tradeType = "💰 实盘"
		}

		pt.AddRow([]interface{}{
			user.ID,
			user.Username,
			user.Exchange,
			tradeType,
			status,
		})
	}

	return pt.Render()
}

// RenderAssets 渲染资产列表
func RenderAssets(assets []exchange.Asset) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("💰 资产列表")
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

// RenderOrders 渲染订单列表
func RenderOrders(orders []exchange.Order) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("📋 订单列表")
	pt.SetHeaders([]interface{}{"订单ID", "交易对", "方向", "仓位", "价格", "数量", "状态"})

	for _, order := range orders {
		side := "🟢 买入"
		if order.Side == "sell" {
			side = "🔴 卖出"
		}

		posSide := order.PosSide
		if posSide == "long" {
			posSide = "📈 多头"
		} else if posSide == "short" {
			posSide = "📉 空头"
		}

		status := "⏳ 等待中"
		switch order.Status {
		case "pending":
			status = "🔄 处理中"
		case "filled":
			status = "✅ 已成交"
		case "cancelled":
			status = "❌ 已取消"
		}

		pt.AddRow([]interface{}{
			order.ID,
			order.Symbol,
			side,
			posSide,
			fmt.Sprintf("%.2f", order.Price),
			fmt.Sprintf("%.4f", order.Size),
			status,
		})
	}

	return pt.Render()
}

// RenderPositions 渲染仓位列表
func RenderPositions(positions []exchange.Position) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("📊 仓位列表")
	pt.SetHeaders([]interface{}{"交易对", "仓位方向", "数量", "均价", "未实现盈亏"})

	for _, pos := range positions {
		posSide := pos.PosSide
		if posSide == "long" {
			posSide = "📈 多头"
		} else if posSide == "short" {
			posSide = "📉 空头"
		}

		pnlColor := "🟢"
		if pos.UnrealPnl < 0 {
			pnlColor = "🔴"
		} else if pos.UnrealPnl == 0 {
			pnlColor = "⚪"
		}

		pt.AddRow([]interface{}{
			pos.Symbol,
			posSide,
			fmt.Sprintf("%.4f", pos.Size),
			fmt.Sprintf("%.2f", pos.AvgPrice),
			fmt.Sprintf("%s %.2f", pnlColor, pos.UnrealPnl),
		})
	}

	return pt.Render()
}

// RenderStrategies 渲染策略列表
func RenderStrategies() string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("🎯 可用策略")
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

// RenderSymbols 渲染交易对列表
func RenderSymbols(symbols []*models.FoxSymbol) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("💱 交易对列表")
	pt.SetHeaders([]interface{}{"#", "交易对", "交易所", "杠杆倍数", "保证金模式"})

	for i, symbol := range symbols {
		pt.AddRow([]interface{}{
			i + 1,
			symbol.Name,
			symbol.Exchange,
			symbol.Leverage,
			symbol.MarginType,
		})
	}

	return pt.Render()
}

// RenderStrategyOrders 渲染策略订单列表
func RenderStrategyOrders(orders []*models.FoxSS) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("🎯 策略订单列表")
	pt.SetHeaders([]interface{}{"ID", "交易对", "方向", "仓位", "价格", "数量", "策略", "状态"})

	for _, order := range orders {
		side := "🟢 买入"
		if order.Side == "sell" {
			side = "🔴 卖出"
		}

		posSide := order.PosSide
		if posSide == "long" {
			posSide = "📈 多头"
		} else if posSide == "short" {
			posSide = "📉 空头"
		}

		status := "⏳ 等待中"
		switch order.Status {
		case "pending":
			status = "🔄 处理中"
		case "filled":
			status = "✅ 已成交"
		case "cancelled":
			status = "❌ 已取消"
		}

		pt.AddRow([]interface{}{
			order.ID,
			order.Symbol,
			side,
			posSide,
			fmt.Sprintf("%.2f", order.Px),
			fmt.Sprintf("%.4f", order.Sz),
			order.Strategy,
			status,
		})
	}

	return pt.Render()
}

// RenderTickers 渲染行情列表
func RenderTickers(tickers []exchange.Ticker) string {
	pt := utils.NewPrettyTable()
	pt.SetTitle("📈 行情列表")
	pt.SetHeaders([]interface{}{"交易对", "价格", "成交量", "最高价", "最低价"})

	for _, ticker := range tickers {
		pt.AddRow([]interface{}{
			ticker.Symbol,
			fmt.Sprintf("%.2f", ticker.Price),
			fmt.Sprintf("%.0f", ticker.Volume),
			fmt.Sprintf("%.2f", ticker.High),
			fmt.Sprintf("%.2f", ticker.Low),
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
