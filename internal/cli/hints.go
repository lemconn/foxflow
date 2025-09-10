package cli

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/lemconn/foxflow/internal/cli/command"
)

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

// getCompleter 获取命令补全器（go-prompt）
func getCompleter(commands map[string]command.Command) prompt.Completer {
	// 命令集合
	var cmdNames []string
	for name := range commands {
		cmdNames = append(cmdNames, name)
	}

	// 子命令集合（迁移至 render 包统一维护文案）
	sub := GetSubcommandSuggestions()

	// 顶层命令建议
	var top []prompt.Suggest
	for _, n := range cmdNames {
		top = append(top, prompt.Suggest{Text: n})
	}

	return func(d prompt.Document) []prompt.Suggest {
		text := d.TextBeforeCursor()
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			return top
		}

		args := parseArgs(text)
		// 情况1：只有一个 token，且原始输入以空格结尾，表示用户已输入命令并在敲空格后等待二级建议
		if len(args) == 1 && strings.HasSuffix(text, " ") {
			if items, ok := sub[args[0]]; ok {
				return items
			}
			return []prompt.Suggest{}
		}

		if len(args) <= 1 {
			// 输入在首个 token：补全命令
			return prompt.FilterHasPrefix(top, d.GetWordBeforeCursor(), true)
		}

		// 第二个 token：如果是已知命令提供子命令建议
		first := args[0]
		if items, ok := sub[first]; ok {
			return prompt.FilterHasPrefix(items, d.GetWordBeforeCursor(), true)
		}

		return []prompt.Suggest{}
	}
}
