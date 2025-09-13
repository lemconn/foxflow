package cli

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/lemconn/foxflow/internal/cli/command"
	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/repository"
)

var currentExchange = ""

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

var createUserArgHints = map[string]map[string]map[string][]prompt.Suggest{
	"create": {
		"users": {
			"mock": {
				{Text: "--username==<username>", Description: "[必填]用户名称"},
				{Text: "--ak==<apiKey>", Description: "[必填]账户apiKey"},
				{Text: "--sk==<secretKey>", Description: "[必填]账户secretKey"},
			},
			"live": {
				{Text: "--username==<username>", Description: "[必填]用户名称"},
				{Text: "--ak==<apiKey>", Description: "[必填]账户apiKey"},
				{Text: "--sk==<secretKey>", Description: "[必填]账户secretKey"},
			},
		},
	},
}

// 各命令的参数提示（按位置给出占位符）
var argHints = map[string]map[string]map[string][]prompt.Suggest{
	"create": {
		"users": {
			"mock": {
				{Text: "--username=<username>", Description: "[必填]用户名称"},
				{Text: "--ak=<apiKey>", Description: "[必填]账户apiKey"},
				{Text: "--sk=<secretKey>", Description: "[必填]账户secretKey"},
			},
			"live": {
				{Text: "--username=<username>", Description: "[必填]用户名称"},
				{Text: "--ak=<apiKey>", Description: "[必填]账户apiKey"},
				{Text: "--sk=<secretKey>", Description: "[必填]账户secretKey"},
			},
		},
	},
}

// 动态生成 user add 的参数提示
func getCreateUsersArgHints(exchangeName string) []prompt.Suggest {
	baseArgs := []prompt.Suggest{
		{Text: "--username=<name>", Description: "[必填]用户名称"},
		{Text: "--ak=<apiKey>", Description: "[必填]账户apiKey"},
		{Text: "--sk=<secretKey>", Description: "[必填]账户secretKey"},
	}

	// 根据动态参数决定是否添加 passphrase
	if exchangeName == config.DefaultExchange {
		baseArgs = append(baseArgs, prompt.Suggest{
			Text:        "--passphrase=<passphrase>",
			Description: "[必填]生成apiKey时的passphrase",
		})
	}

	return baseArgs
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

func createUsersTradeTypeList() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "mock", Description: "- 模拟盘"},
		{Text: "live", Description: "- 实盘"},
	}
}

// getCompleter 获取命令补全器（go-prompt）
func getCompleter(ctx *Context, commands map[string]command.Command) prompt.Completer {
	// 命令集合
	var cmdNames []string
	for name := range commands {
		cmdNames = append(cmdNames, name)
	}

	// 获取交易所信息
	currentExchange = ctx.GetExchangeName()

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
			if items, ok := getSubcommandSuggestions()[first]; ok {
				return items
			}
			return []prompt.Suggest{}
		}

		// 正在输入第二个token（进行命令提示）
		second := fields[1]
		if len(fields) == 2 && !strings.HasSuffix(w, " ") {
			secondPrefix := d.GetWordBeforeCursor()
			if items, ok := getSubcommandSuggestions()[first]; ok {
				return prompt.FilterHasPrefix(items, secondPrefix, true)
			}
		}

		// use exchanges 的第一个参数输入过程中（未以空格结束）动态过滤交易所信息
		if len(fields) >= 2 && strings.HasSuffix(w, " ") && first == "use" && second == "exchanges" {
			prefix := d.GetWordBeforeCursor()
			return prompt.FilterHasPrefix(useExchangesList(), prefix, true)
		}

		// create users 后的模式选择：第二个token已完成，在第三个token位置
		if len(fields) == 2 && strings.HasSuffix(w, " ") && first == "create" && second == "users" {
			return createUsersTradeTypeList()
		}

		// 正在输入create users后的模式选择（第三个token）
		if len(fields) == 3 && !strings.HasSuffix(w, " ") && first == "create" && second == "users" {
			prefix := d.GetWordBeforeCursor()
			return prompt.FilterHasPrefix(createUsersTradeTypeList(), prefix, true)
		}

		// 参数位置提示：在子命令之后，为每个参数位置给出占位符
		if len(fields) >= 2 {
			var mode string
			var argOffset int

			// 对于user add，如果已经指定了模式，则获取模式信息
			if first == "create" && second == "users" && len(fields) >= 3 {
				mode = fields[2]
				argOffset = 1 // create users 模式后的参数从第4个token开始
			}

			// 仅在新参数开始时提示（即末尾有空格），避免在输入值时干扰
			argIndex := -1
			if len(fields) == 2+argOffset && strings.HasSuffix(w, " ") {
				argIndex = 0
			} else if len(fields) >= 3+argOffset && strings.HasSuffix(w, " ") {
				argIndex = len(fields) - 2 - argOffset // 计算实际参数索引
			}

			if argIndex >= 0 {
				// 获取参数提示
				if first == "create" && second == "users" {
					// 对于 user add 命令，使用动态生成的参数提示
					hints := getCreateUsersArgHints(ctx.GetExchangeName())
					if argIndex < len(hints) {
						return hints[argIndex:]
					}
				} else {
					// 其他命令使用静态配置的参数提示
					if group, ok := argHints[first]; ok {
						if subGroup, ok := group[second]; ok {
							if hints, ok := subGroup[mode]; ok && argIndex < len(hints) {
								return hints[argIndex:]
							}
						}
					}
				}
			}
		}

		return []prompt.Suggest{}
	}
}
