package render

import (
	"foxflow/pkg/utils"
	"strings"

	"github.com/c-bata/go-prompt"
)

// RenderWelcomeHints 渲染 CLI 启动提示文案
func RenderWelcomeHints() string {
	lines := []string{
		utils.RenderInfo("输入 'help' 查看可用命令"),
		utils.RenderInfo("输入 'exit' 或 'quit' 或 'Ctrl-D' 退出程序"),
	}
	return strings.Join(lines, "\n") + "\n\n"
}

// GetSubcommandSuggestions 返回各命令的子命令建议及说明文案
func GetSubcommandSuggestions() map[string][]prompt.Suggest {
	return map[string][]prompt.Suggest{
		"show": {
			{Text: "exchanges", Description: "展示交易所"},
			{Text: "users", Description: "展示用户"},
			{Text: "assets", Description: "展示资产"},
			{Text: "orders", Description: "展示订单"},
			{Text: "positions", Description: "展示持仓"},
			{Text: "strategies", Description: "展示当前可用策略"},
			{Text: "symbols", Description: "展示交易对"},
			{Text: "ss", Description: "展示策略订单"},
		},
		"use": {
			{Text: "exchanges", Description: "选择交易所"},
			{Text: "users", Description: "选择用户"},
		},
		"create": {
			{Text: "users", Description: "创建用户"},
			{Text: "symbols", Description: "创建交易对"},
			{Text: "ss", Description: "创建策略订单"},
		},
		"update": {
			{Text: "leverage", Description: "调整交易对杠杆系数"},
			{Text: "margin-type", Description: "调整交易对保证金模式"},
		},
		"cancel": {
			{Text: "ss", Description: "取消策略订单"},
		},
		"delete": {
			{Text: "users", Description: "删除用户"},
			{Text: "symbols", Description: "删除交易对"},
		},
	}
}
