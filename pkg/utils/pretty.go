package utils

import (
	"fmt"

	"foxflow/internal/exchange"
	"foxflow/internal/models"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// PrettyTable 美化表格输出
type PrettyTable struct {
	writer table.Writer
}

// NewPrettyTable 创建新的美化表格
func NewPrettyTable() *PrettyTable {
	t := table.NewWriter()
	t.SetStyle(table.StyleColoredBright)
	t.Style().Options.SeparateRows = true
	t.Style().Color.Header = []text.Color{text.FgHiCyan, text.Bold}
	t.Style().Color.Row = []text.Color{text.FgHiWhite}
	t.Style().Color.RowAlternate = []text.Color{text.FgWhite}

	return &PrettyTable{writer: t}
}

// SetTitle 设置表格标题
func (pt *PrettyTable) SetTitle(title string) {
	pt.writer.SetTitle(title)
}

// SetHeaders 设置表头
func (pt *PrettyTable) SetHeaders(headers []interface{}) {
	pt.writer.AppendHeader(headers)
}

// AddRow 添加行
func (pt *PrettyTable) AddRow(row []interface{}) {
	pt.writer.AppendRow(row)
}

// Render 渲染表格
func (pt *PrettyTable) Render() string {
	return pt.writer.Render()
}

// RenderExchanges 渲染交易所列表
func RenderExchanges(exchanges []string) string {
	pt := NewPrettyTable()
	pt.SetTitle("🏦 可用交易所")
	pt.SetHeaders([]interface{}{"#", "交易所名称", "状态"})

	for i, exchange := range exchanges {
		pt.AddRow([]interface{}{
			i + 1,
			exchange,
			"✅ 活跃",
		})
	}

	return pt.Render()
}

// RenderUsers 渲染用户列表
func RenderUsers(users []models.FoxUser) string {
	pt := NewPrettyTable()
	pt.SetTitle("👥 用户列表")
	pt.SetHeaders([]interface{}{"ID", "用户名", "交易所", "交易类型", "状态"})

	for _, user := range users {
		status := "✅ 活跃"
		if user.Status == "inactive" {
			status = "❌ 非活跃"
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
	pt := NewPrettyTable()
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
	pt := NewPrettyTable()
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
	pt := NewPrettyTable()
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
	pt := NewPrettyTable()
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
func RenderSymbols(symbols []string) string {
	pt := NewPrettyTable()
	pt.SetTitle("💱 交易对列表")
	pt.SetHeaders([]interface{}{"#", "交易对"})

	for i, symbol := range symbols {
		pt.AddRow([]interface{}{
			i + 1,
			symbol,
		})
	}

	return pt.Render()
}

// RenderStrategyOrders 渲染策略订单列表
func RenderStrategyOrders(orders []models.FoxSS) string {
	pt := NewPrettyTable()
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
	pt := NewPrettyTable()
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

// RenderSuccess 渲染成功消息
func RenderSuccess(message string) string {
	return fmt.Sprintf("✅ %s", message)
}

// RenderError 渲染错误消息
func RenderError(message string) string {
	return fmt.Sprintf("❌ %s", message)
}

// RenderInfo 渲染信息消息
func RenderInfo(message string) string {
	return fmt.Sprintf("ℹ️  %s", message)
}

// RenderWarning 渲染警告消息
func RenderWarning(message string) string {
	return fmt.Sprintf("⚠️  %s", message)
}
