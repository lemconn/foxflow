package cli

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/repository"
)

var topLevel = []prompt.Suggest{
	{Text: command.TopCommandHelp, Description: "- 显示帮助信息"},
	{Text: command.TopCommandShow, Description: "- 查看数据列表"},
	{Text: command.TopCommandUse, Description: "- 激活交易所或用户"},
	{Text: command.TopCommandCreate, Description: "- 创建用户、标的或策略订单"},
	{Text: command.TopCommandUpdate, Description: "- 设置杠杆或保证金模式"},
	{Text: command.TopCommandCancel, Description: "- 取消策略订单"},
	{Text: command.TopCommandDelete, Description: "- 删除用户或标的"},
	{Text: command.TopCommandExit, Description: "- 退出系统"},
	{Text: command.TopCommandQuit, Description: "- 退出系统"},
}

// getSubcommandSuggestions 返回各命令的子命令建议及说明文案
func getSubcommandSuggestions() map[string][]prompt.Suggest {
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

// useExchangesList 激活交易所列表
func useExchangesList() []prompt.Suggest {
	// 获取所有交易所列表
	exchangeList, err := repository.ListExchanges()
	if err != nil {
		return []prompt.Suggest{}
	}

	exchanges := make([]prompt.Suggest, 0, len(exchangeList))
	for _, exchange := range exchangeList {
		exchanges = append(exchanges, prompt.Suggest{Text: exchange.Name})
	}

	return exchanges
}

// getCompleter 获取命令补全器（go-prompt）
func getCompleter(commands map[string]command.Command) prompt.Completer {
	// 命令集合
	var cmdNames []string
	for name := range commands {
		cmdNames = append(cmdNames, name)
	}

	// 子命令提示集合
	sub := getSubcommandSuggestions()

	return func(d prompt.Document) []prompt.Suggest {
		// 按空格分割，决定补全上下文
		w := d.TextBeforeCursor()
		fields := parseArgs(w)

		// 如果没有输入或正在输入第一个token，补全顶层命令
		if len(fields) == 0 || (len(fields) == 1 && !strings.HasSuffix(w, " ")) {
			return prompt.FilterHasPrefix(topLevel, d.GetWordBeforeCursor(), true)
		}

		// 如果第一个token已完成(后面有空格)，根据命令补全子命令
		first := fields[0]
		if len(fields) == 1 && strings.HasSuffix(w, " ") {
			if items, ok := sub[first]; ok {
				return items
			}
			return []prompt.Suggest{}
		}

		// 正在输入第二个token（进行命令提示）
		if len(fields) >= 2 && !strings.HasSuffix(w, " ") {
			secondPrefix := d.GetWordBeforeCursor()
			if items, ok := sub[first]; ok {
				return prompt.FilterHasPrefix(items, secondPrefix, true)
			}
		}

		// use exchanges 的第一个参数输入过程中（未以空格结束）动态过滤交易所信息
		if len(fields) == 2 && strings.HasSuffix(w, " ") && first == "use" && fields[1] == "exchanges" {
			prefix := d.GetWordBeforeCursor()
			return prompt.FilterHasPrefix(useExchangesList(), prefix, true)
		}

		return []prompt.Suggest{}
	}
}
