package utils

import (
	"fmt"

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

// MessageRed 红色
func MessageRed(message string) string {
	return messageColor(91, message)
}

// MessageGreen 绿色
func MessageGreen(message string) string {
	return messageColor(92, message)
}

// MessageYellow 黄色
func MessageYellow(message string) string {
	return messageColor(93, message)
}

// MessageBlue 蓝色
func MessageBlue(message string) string {
	return messageColor(94, message)
}

// MessagePurple 紫色
func MessagePurple(message string) string {
	return messageColor(95, message)
}

// MessageCyan 青色
func MessageCyan(message string) string {
	return messageColor(96, message)
}

func messageColor(code int, message string) string {
	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 0, code, message, 0x1B)
}
