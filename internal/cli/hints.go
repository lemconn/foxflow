package cli

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/repository"
)

// parseCommandInput 解析命令输入，返回字段列表
func parseCommandInput(w string) []string {
	return parseArgs(w)
}

// handleTopLevelCompletion 处理一级命令补全
func handleTopLevelCompletion(d prompt.Document, w string, fields []string) []prompt.Suggest {
	// 如果没有输入或正在输入第一个token，补全顶层命令
	if len(fields) == 0 || (len(fields) == 1 && !strings.HasSuffix(w, " ")) {
		return prompt.FilterHasPrefix(getTopLevelHelpSuggestions(), d.GetWordBeforeCursor(), true)
	}
	return nil
}

// handleSubcommandCompletion 处理二级命令补全
func handleSubcommandCompletion(d prompt.Document, w string, fields []string, first string) []prompt.Suggest {
	// 如果第一个token已完成(后面有空格)，根据命令补全子命令
	if len(fields) == 1 && strings.HasSuffix(w, " ") {
		if items, ok := getSubcommandSuggestions()[first]; ok {
			return items
		}
		// 对于没有子命令的命令，返回nil让后续处理
		return nil
	}

	// 正在输入第二个token（进行命令提示）
	if len(fields) == 2 && !strings.HasSuffix(w, " ") {
		secondPrefix := d.GetWordBeforeCursor()
		if items, ok := getSubcommandSuggestions()[first]; ok {
			return prompt.FilterHasPrefix(items, secondPrefix, true)
		}
	}
	return nil
}

// handleSpecialCommandCompletions 处理特殊命令的补全逻辑
func handleSpecialCommandCompletions(ctx *Context, d prompt.Document, w string, fields []string, first, second string) []prompt.Suggest {
	// use 命令的动态补全
	if result := handleUseCommandCompletion(d, w, fields, first, second); result != nil {
		return result
	}

	// create account 后的类型选择
	if result := handleCreateAccountCompletion(d, w, fields, first, second); result != nil {
		return result
	}

	// delete account 后的类型选择
	if result := handleDeleteAccountCompletion(ctx, d, w, fields, first, second); result != nil {
		return result
	}

	// update 命令的补全
	if result := handleUpdateCommandCompletion(ctx, d, w, fields, first, second); result != nil {
		return result
	}

	// cancel 命令的补全
	if result := handleCancelCommandCompletion(ctx, d, w, fields, first, second); result != nil {
		return result
	}

	return nil
}

// handleUseCommandCompletion 处理 use 命令的动态补全
func handleUseCommandCompletion(d prompt.Document, w string, fields []string, first, second string) []prompt.Suggest {
	if first != "use" {
		return nil
	}

	// 处理 use exchange/account 后的参数补全
	if len(fields) == 2 && strings.HasSuffix(w, " ") {
		// 输入 "use exchange " 或 "use account " 后显示所有选项
		prefix := d.GetWordBeforeCursor()
		switch second {
		case "exchange":
			return prompt.FilterHasPrefix(useExchangesList(), prefix, true)
		case "account":
			return prompt.FilterHasPrefix(useAccountsList(), prefix, true)
		}
	} else if len(fields) == 3 && !strings.HasSuffix(w, " ") {
		// 正在输入第三个参数时的模糊匹配，如 "use exchange ok"
		prefix := d.GetWordBeforeCursor()
		switch second {
		case "exchange":
			return prompt.FilterHasPrefix(useExchangesList(), prefix, true)
		case "account":
			return prompt.FilterHasPrefix(useAccountsList(), prefix, true)
		}
	}

	return nil
}

// handleCreateAccountCompletion 处理 create account 命令的补全
func handleCreateAccountCompletion(d prompt.Document, w string, fields []string, first, second string) []prompt.Suggest {
	if first != "create" || second != "account" {
		return nil
	}

	// 类型选择：第二个token已完成，在第三个token位置
	if len(fields) == 2 && strings.HasSuffix(w, " ") {
		return createAccountTradeTypeList()
	}

	// 正在输入类型选择（第三个token）
	if len(fields) == 3 && !strings.HasSuffix(w, " ") {
		prefix := d.GetWordBeforeCursor()
		return prompt.FilterHasPrefix(createAccountTradeTypeList(), prefix, true)
	}

	return nil
}

// handleDeleteAccountCompletion 处理 create account 命令的补全
func handleDeleteAccountCompletion(ctx *Context, d prompt.Document, w string, fields []string, first, second string) []prompt.Suggest {
	if first != "delete" || second != "account" {
		return nil
	}

	// 类型选择：第二个token已完成，在第三个token位置
	if len(fields) == 2 && strings.HasSuffix(w, " ") {
		return deleteAccountsList(ctx)
	}

	// 正在输入类型选择（第三个token）
	if len(fields) == 3 && !strings.HasSuffix(w, " ") {
		prefix := d.GetWordBeforeCursor()
		return prompt.FilterHasPrefix(deleteAccountsList(ctx), prefix, true)
	}

	return nil
}

// handleUpdateCommandCompletion 处理 update 命令的补全
func handleUpdateCommandCompletion(ctx *Context, d prompt.Document, w string, fields []string, first, second string) []prompt.Suggest {
	if first != "update" {
		return nil
	}

	// 处理 update symbol
	if result := handleUpdateSymbolCompletion(ctx, d, w, fields, first, second); result != nil {
		return result
	}

	// 处理 update account
	if result := handleUpdateAccountCompletion(ctx, d, w, fields, first, second); result != nil {
		return result
	}

	return nil
}

// handleUpdateSymbolCompletion 处理 update symbol 命令的补全
func handleUpdateSymbolCompletion(ctx *Context, d prompt.Document, w string, fields []string, first, second string) []prompt.Suggest {
	if second != "symbol" {
		return nil
	}

	// 选择symbol：第三个token位置
	if len(fields) == 2 && strings.HasSuffix(w, " ") {
		return getDynamicSymbolList(ctx)
	}

	// 正在输入symbol（第三个token）
	if len(fields) == 3 && !strings.HasSuffix(w, " ") {
		prefix := strings.ToLower(d.GetWordBeforeCursor()) // 忽略大小写
		symbols := getDynamicSymbolList(ctx)
		var filtered []prompt.Suggest
		for _, symbol := range symbols {
			if strings.Contains(strings.ToLower(symbol.Text), prefix) {
				filtered = append(filtered, symbol)
			}
		}
		return filtered
	}

	// 选择symbol后，显示参数（从第4个字段开始处理参数）
	if len(fields) >= 3 {
		// 获取所有可用参数
		allArgs := getUpdateSymbolArgHints()

		// 如果刚选择完类型（第3个token后有空格），显示所有参数
		if len(fields) == 3 && strings.HasSuffix(w, " ") {
			return allArgs
		}

		// 如果正在输入参数（len(fields) >= 4），处理参数输入
		if len(fields) >= 4 {
			// 获取所有已输入的参数名（已经完成的参数）
			var usedParams []string
			for i := 3; i < len(fields); i++ {
				paramInput := fields[i]
				// 检查是否已经输入了完整的参数（包含=且有值）
				if strings.Contains(paramInput, "=") {
					parts := strings.Split(paramInput, "=")
					if len(parts) >= 1 && parts[0] != "" {
						usedParams = append(usedParams, parts[0])
					}
				}
			}

			// 过滤掉已使用的参数
			var availableArgs []prompt.Suggest
			for _, arg := range allArgs {
				argName := strings.Split(arg.Text, "=")[0]
				isUsed := false
				for _, used := range usedParams {
					if argName == used {
						isUsed = true
						break
					}
				}
				if !isUsed {
					availableArgs = append(availableArgs, arg)
				}
			}

			// 获取当前正在输入的参数（最后一个字段）
			currentInput := fields[len(fields)-1]

			// 如果命令以空格结尾，表示要开始输入新参数
			if strings.HasSuffix(w, " ") {
				return availableArgs
			}

			// 如果当前输入不包含=，表示正在输入参数名，支持模糊匹配
			if !strings.Contains(currentInput, "=") {
				prefix := strings.ToLower(currentInput)
				var filtered []prompt.Suggest
				for _, arg := range availableArgs {
					argName := strings.Split(arg.Text, "=")[0]
					if strings.Contains(strings.ToLower(argName), prefix) {
						filtered = append(filtered, arg)
					}
				}
				return filtered
			}

			// 如果当前输入包含=但没有完整（可能正在输入值），显示剩余参数
			if strings.Contains(currentInput, "=") && !strings.HasSuffix(currentInput, " ") {
				return availableArgs
			}
		}
	}

	return nil
}

// handleUpdateAccountCompletion 处理 update account 命令的补全
func handleUpdateAccountCompletion(ctx *Context, d prompt.Document, w string, fields []string, first, second string) []prompt.Suggest {
	if second != "account" {
		return nil
	}

	// 选择账户：第三个token位置
	if len(fields) == 2 && strings.HasSuffix(w, " ") {
		return getDynamicAccountList(ctx)
	}

	// 正在输入账户名（第三个token）
	if len(fields) == 3 && !strings.HasSuffix(w, " ") {
		prefix := strings.ToLower(d.GetWordBeforeCursor()) // 忽略大小写
		accounts := getDynamicAccountList(ctx)
		var filtered []prompt.Suggest
		for _, account := range accounts {
			if strings.Contains(strings.ToLower(account.Text), prefix) {
				filtered = append(filtered, account)
			}
		}
		return filtered
	}

	// 选择账户后，显示类型选择
	if len(fields) == 3 && strings.HasSuffix(w, " ") {
		return updateAccountTradeTypeList()
	}

	// 正在输入类型（第四个token）
	if len(fields) == 4 && !strings.HasSuffix(w, " ") {
		prefix := d.GetWordBeforeCursor()
		return prompt.FilterHasPrefix(updateAccountTradeTypeList(), prefix, true)
	}

	// 选择类型后，显示参数提示
	if len(fields) == 4 && strings.HasSuffix(w, " ") {
		return getUpdateAccountArgHints(ctx.GetExchangeName())
	}

	return nil
}

// handleArgumentHints 处理参数提示
func handleArgumentHints(ctx *Context, d prompt.Document, w string, fields []string, first, second string) []prompt.Suggest {
	if len(fields) < 2 {
		return nil
	}

	var mode string
	var argOffset int

	// 对于 update account，在选择完类型后显示参数
	if first == "update" && second == "account" && len(fields) >= 4 {
		// 获取所有可用参数
		allArgs := getUpdateAccountArgHints(ctx.GetExchangeName())

		// 如果刚选择完类型（第4个token后有空格），显示所有参数
		if len(fields) == 4 && strings.HasSuffix(w, " ") {
			return allArgs
		}

		// 如果正在输入参数（len(fields) >= 5），处理参数输入
		if len(fields) >= 5 {
			// 获取所有已输入的参数名（已经完成的参数）
			var usedParams []string
			for i := 4; i < len(fields); i++ {
				paramInput := fields[i]
				// 检查是否已经输入了完整的参数（包含=且有值）
				if strings.Contains(paramInput, "=") {
					parts := strings.Split(paramInput, "=")
					if len(parts) >= 1 && parts[0] != "" {
						usedParams = append(usedParams, parts[0])
					}
				}
			}

			// 过滤掉已使用的参数
			var availableArgs []prompt.Suggest
			for _, arg := range allArgs {
				argName := strings.Split(arg.Text, "=")[0]
				isUsed := false
				for _, used := range usedParams {
					if argName == used {
						isUsed = true
						break
					}
				}
				if !isUsed {
					availableArgs = append(availableArgs, arg)
				}
			}

			// 获取当前正在输入的参数（最后一个字段）
			currentInput := fields[len(fields)-1]

			// 如果命令以空格结尾，表示要开始输入新参数
			if strings.HasSuffix(w, " ") {
				return availableArgs
			}

			// 如果当前输入不包含=，表示正在输入参数名，支持模糊匹配
			if !strings.Contains(currentInput, "=") {
				prefix := strings.ToLower(currentInput)
				var filtered []prompt.Suggest
				for _, arg := range availableArgs {
					argName := strings.Split(arg.Text, "=")[0]
					if strings.Contains(strings.ToLower(argName), prefix) {
						filtered = append(filtered, arg)
					}
				}
				return filtered
			}

			// 如果当前输入包含=但没有完整（可能正在输入值），显示剩余参数
			if strings.Contains(currentInput, "=") && !strings.HasSuffix(currentInput, " ") {
				return availableArgs
			}
		}
	}

	// 对于 show symbol，在输入子命令后显示标的补全，支持自定义输入
	if first == "show" && second == "symbol" {
		// 如果刚输入 "show symbol "（后面有空格），显示所有可用标的作为建议
		if len(fields) == 2 && strings.HasSuffix(w, " ") {
			return getDynamicSymbolList(ctx)
		}

		// 如果正在输入第三个参数（标的名称），支持模糊匹配，同时允许自定义输入
		if len(fields) == 3 && !strings.HasSuffix(w, " ") {
			prefix := strings.ToLower(d.GetWordBeforeCursor())
			symbols := getDynamicSymbolList(ctx)
			var filtered []prompt.Suggest
			for _, symbol := range symbols {
				if strings.Contains(strings.ToLower(symbol.Text), prefix) {
					filtered = append(filtered, symbol)
				}
			}
			return filtered
		}
	}

	// 对于 create account，在选择完类型后显示参数
	if first == "create" && second == "account" && len(fields) >= 3 {
		// 获取所有可用参数
		allArgs := getCreateAccountArgHints(ctx.GetExchangeName())

		// 如果刚选择完类型（第3个token后有空格），显示所有参数
		if len(fields) == 3 && strings.HasSuffix(w, " ") {
			return allArgs
		}

		// 如果正在输入参数（len(fields) >= 4），处理参数输入
		if len(fields) >= 4 {
			// 获取所有已输入的参数名（已经完成的参数）
			var usedParams []string
			for i := 3; i < len(fields); i++ {
				paramInput := fields[i]
				// 检查是否已经输入了完整的参数（包含=且有值）
				if strings.Contains(paramInput, "=") {
					parts := strings.Split(paramInput, "=")
					if len(parts) >= 1 && parts[0] != "" {
						usedParams = append(usedParams, parts[0])
					}
				}
			}

			// 过滤掉已使用的参数
			var availableArgs []prompt.Suggest
			for _, arg := range allArgs {
				argName := strings.Split(arg.Text, "=")[0]
				isUsed := false
				for _, used := range usedParams {
					if argName == used {
						isUsed = true
						break
					}
				}
				if !isUsed {
					availableArgs = append(availableArgs, arg)
				}
			}

			// 获取当前正在输入的参数（最后一个字段）
			currentInput := fields[len(fields)-1]

			// 如果命令以空格结尾，表示要开始输入新参数
			if strings.HasSuffix(w, " ") {
				return availableArgs
			}

			// 如果当前输入不包含=，表示正在输入参数名，支持模糊匹配
			if !strings.Contains(currentInput, "=") {
				prefix := strings.ToLower(currentInput)
				var filtered []prompt.Suggest
				for _, arg := range availableArgs {
					argName := strings.Split(arg.Text, "=")[0]
					if strings.Contains(strings.ToLower(argName), prefix) {
						filtered = append(filtered, arg)
					}
				}
				return filtered
			}

			// 如果当前输入包含=但没有完整（可能正在输入值），显示剩余参数
			if strings.Contains(currentInput, "=") && !strings.HasSuffix(currentInput, " ") {
				return availableArgs
			}
		}
	}

	// 仅在新参数开始时提示（即末尾有空格），避免在输入值时干扰
	argIndex := -1
	if len(fields) == 2+argOffset && strings.HasSuffix(w, " ") {
		argIndex = 0
	} else if len(fields) >= 3+argOffset && strings.HasSuffix(w, " ") {
		argIndex = len(fields) - 2 - argOffset // 计算实际参数索引
	}

	if argIndex >= 0 {
		return getArgumentSuggestions(ctx, fields, first, second, mode, argIndex)
	}

	return nil
}

// getArgumentSuggestions 获取参数建议
func getArgumentSuggestions(ctx *Context, fields []string, first, second, mode string, argIndex int) []prompt.Suggest {
	// create symbols 的动态提示
	if first == "create" && second == "symbols" {
		return getCreateSymbolsArgSuggestions(fields)
	}

	// 其他命令使用静态配置的参数提示
	if group, ok := argHints[first]; ok {
		if subGroup, ok := group[second]; ok {
			if hints, ok := subGroup[mode]; ok && argIndex < len(hints) {
				return hints[argIndex:]
			}
		}
	}

	return []prompt.Suggest{}
}

// getCreateSymbolsArgSuggestions 获取 create symbols 的参数建议
func getCreateSymbolsArgSuggestions(fields []string) []prompt.Suggest {
	// 已输入的参数（去掉前两个token: create symbols）
	extra := []string{}
	if len(fields) > 2 {
		extra = fields[2:]
	}

	// 情况A：正在输入第一个参数（symbol）
	if len(extra) == 0 {
		return []prompt.Suggest{
			{Text: "<symbol>", Description: "交易对名称，例如：BTC"},
		}
	}

	// 检查是否已包含可选参数
	hasLeverage := false
	hasMarginType := false
	marginPrefixEntered := ""
	for i, t := range extra {
		// 第一个为必填 symbol，不参与判断
		if i == 0 {
			continue
		}
		if strings.HasPrefix(t, "--leverage=") {
			hasLeverage = true
		}
		// 仅当已完整选择其中一个模式时，才视为 hasMarginType
		if t == "--margin-type=isolated" || t == "--margin-type=cross" {
			hasMarginType = true
		} else if strings.HasPrefix(t, "--margin-type=") {
			marginPrefixEntered = t
		}
	}

	// 情况B：已输入了 <symbol>，还未输入任何可选项
	if len(extra) == 1 {
		return []prompt.Suggest{
			{Text: "--leverage=<num>", Description: "[选填] 杠杆倍数，范围 1-100"},
			{Text: "--margin-type=isolated", Description: "[选填] 保证金模式：逐仓"},
			{Text: "--margin-type=cross", Description: "[选填] 保证金模式：全仓"},
		}
	}

	// 情况C：已输入了 --leverage，仅提示保证金模式两个选项
	if hasLeverage && !hasMarginType {
		opts := []prompt.Suggest{
			{Text: "--margin-type=isolated", Description: "保证金模式：逐仓"},
			{Text: "--margin-type=cross", Description: "保证金模式：全仓"},
		}
		// 如果用户已输入了 --margin-type= 前缀，则使用该前缀进行过滤
		if marginPrefixEntered != "" {
			return prompt.FilterHasPrefix(opts, marginPrefixEntered, true)
		}
		return opts
	}

	// 情况D：已输入了保证金模式，但还没输入杠杆时，提示杠杆参数
	if !hasLeverage && hasMarginType {
		return []prompt.Suggest{
			{Text: "--leverage=<num>", Description: "杠杆倍数，范围 1-100"},
		}
	}

	// 情况E：均已输入或无法判断，返回空
	return []prompt.Suggest{}
}

// getCompleter 获取命令补全器（go-prompt）
func getCompleter(ctx *Context) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		w := d.TextBeforeCursor()
		fields := parseCommandInput(w)

		// 处理一级命令补全
		if result := handleTopLevelCompletion(d, w, fields); result != nil {
			return result
		}

		if len(fields) == 0 {
			return []prompt.Suggest{}
		}

		first := fields[0]

		// 处理二级命令补全
		if result := handleSubcommandCompletion(d, w, fields, first); result != nil {
			return result
		}

		// 特殊处理 open 命令（因为它只需要一个参数）
		if first == "open" {
			if result := handleOpenCommandCompletion(ctx, d, w, fields, first); result != nil {
				return result
			}
		}

		// 特殊处理 close 命令（因为它只需要一个参数）
		if first == "close" {
			if result := handleCloseCommandCompletion(ctx, d, w, fields, first); result != nil {
				return result
			}
		}

		if len(fields) >= 2 {
			second := fields[1]

			// 处理特殊命令补全
			if result := handleSpecialCommandCompletions(ctx, d, w, fields, first, second); result != nil {
				return result
			}

			// 处理参数提示
			if result := handleArgumentHints(ctx, d, w, fields, first, second); result != nil {
				return result
			}
		}

		return []prompt.Suggest{}
	}
}

// getTopLevelHelpSuggestions 获取所有命令的建议（实际上是所有一级命令的详细说明）
func getTopLevelHelpSuggestions() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "help", Description: "显示帮助信息 - 查看所有命令说明"},
		{Text: "show", Description: "查看数据列表 - 支持子命令：exchange(交易所)、account(账户)、balance(资产)、position(持仓)、symbol(交易对)、strategy(策略)、order(订单)、news(新闻)"},
		{Text: "use", Description: "激活上下文 - 支持子命令：exchange(交易所)、account(交易账户)"},
		{Text: "create", Description: "创建资源 - 支持子命令：account(交易账户)"},
		{Text: "update", Description: "更新配置 - 支持子命令：symbol(交易对)、account(交易账户)"},
		{Text: "open", Description: "开仓/下单 - 执行交易开仓操作"},
		{Text: "close", Description: "平仓 - 执行交易平仓操作"},
		{Text: "cancel", Description: "取消订单 - 支持子命令：order(策略订单)"},
		{Text: "delete", Description: "删除资源 - 支持子命令：account(交易账户)"},
		{Text: "exit", Description: "退出系统"},
		{Text: "quit", Description: "退出系统"},
	}
}

// getShowSuggestions 获取show命令的子命令建议
func getShowSuggestions() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "exchange", Description: "查看所有交易所"},
		{Text: "account", Description: "查看所有账户"},
		{Text: "balance", Description: "查看个人资产"},
		{Text: "position", Description: "查看持仓"},
		{Text: "symbol", Description: "查看可用交易对"},
		{Text: "strategy", Description: "查看策略模板"},
		{Text: "order", Description: "查看订单列表"},
		{Text: "news", Description: "查看金融新闻"},
	}
}

// getUseSuggestions 获取use命令的子命令建议
func getUseSuggestions() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "exchange", Description: "激活交易所"},
		{Text: "account", Description: "激活账户"},
	}
}

// getSubcommandSuggestions 返回各命令的子命令建议及说明文案
func getSubcommandSuggestions() map[string][]prompt.Suggest {
	return map[string][]prompt.Suggest{
		"show": getShowSuggestions(),
		"use":  getUseSuggestions(),
		"create": {
			{Text: "account", Description: "创建交易账户"},
		},
		"update": {
			{Text: "symbol", Description: "更新交易对配置"},
			{Text: "account", Description: "更新账户配置"},
		},
		"cancel": {
			{Text: "order", Description: "取消订单"},
		},
		"delete": {
			{Text: "account", Description: "删除交易账户"},
		},
	}
}

// 各命令的参数提示（按位置给出占位符）
var argHints = map[string]map[string]map[string][]prompt.Suggest{
	"create": {
		"account": {},
	},
}

// 动态生成 create account 的参数提示
func getCreateAccountArgHints(exchangeName string) []prompt.Suggest {
	baseArgs := []prompt.Suggest{
		{Text: "name=", Description: "[必填] 用户账户名称"},
		{Text: "apiKey=", Description: "[必填] API访问密钥"},
		{Text: "secretKey=", Description: "[必填] API密钥密钥"},
	}

	// 根据交易所决定是否添加 passphrase（仅OKX需要）
	if exchangeName == config.DefaultExchange {
		baseArgs = append(baseArgs, prompt.Suggest{
			Text:        "passphrase=",
			Description: "[选填] API密钥密码（如OKX交易所）",
		})
	}

	return baseArgs
}

func createAccountTradeTypeList() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "mock", Description: "模拟盘 - 用于测试和学习"},
		{Text: "live", Description: "实盘 - 真实交易环境"},
	}
}

func updateAccountTradeTypeList() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "mock", Description: "模拟盘 - 用于测试和学习"},
		{Text: "live", Description: "实盘 - 真实交易环境"},
	}
}

// getDynamicSymbolList 获取动态symbol列表
func getDynamicSymbolList(ctx *Context) []prompt.Suggest {
	if ctx.GetExchangeName() == "" {
		return []prompt.Suggest{}
	}

	symbols, exist := config.ExchangeSymbolList[ctx.GetExchangeName()]
	if !exist {
		return []prompt.Suggest{}
	}

	var suggestions []prompt.Suggest
	for _, symbol := range symbols {
		suggestions = append(suggestions, prompt.Suggest{
			Text: symbol.Name,
		})
	}

	return suggestions
}

// getDynamicAccountList 获取动态账户列表
func getDynamicAccountList(ctx *Context) []prompt.Suggest {
	accountList, err := repository.ExchangeAccountList(ctx.GetExchangeName())
	if err != nil {
		return []prompt.Suggest{}
	}

	var suggestions []prompt.Suggest
	for _, account := range accountList {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        account.Name,
			Description: fmt.Sprintf("%s：%s", account.TradeType, account.Exchange),
		})
	}

	return suggestions
}

// getUpdateSymbolArgHints 获取update symbol的参数提示
func getUpdateSymbolArgHints() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "margin=", Description: "[必填] 保证金模式：isolated(逐仓)/cross(全仓)"},
		{Text: "leverage=", Description: "[必填] 杠杆倍数"},
	}
}

// getUpdateAccountArgHints 获取update account的参数提示
func getUpdateAccountArgHints(exchangeName string) []prompt.Suggest {
	baseArgs := []prompt.Suggest{
		{Text: "name=", Description: "[必填] 用户账户名称"},
		{Text: "apiKey=", Description: "[必填] API访问密钥"},
		{Text: "secretKey=", Description: "[必填] API密钥密钥"},
	}

	// 根据交易所决定是否添加 passphrase（仅OKX需要）
	if exchangeName == config.DefaultExchange {
		baseArgs = append(baseArgs, prompt.Suggest{
			Text:        "passphrase=",
			Description: "[选填] API密钥密码（如OKX交易所）",
		})
	}

	return baseArgs
}

// handleOpenCommandCompletion 处理 open 命令的补全
func handleOpenCommandCompletion(ctx *Context, d prompt.Document, w string, fields []string, first string) []prompt.Suggest {
	if first != "open" {
		return nil
	}

	// 选择symbol：第二个token位置
	if len(fields) == 1 && strings.HasSuffix(w, " ") {
		return getOpenSymbolList(ctx)
	}

	// 正在输入symbol（第二个token）
	if len(fields) == 2 && !strings.HasSuffix(w, " ") {
		prefix := strings.ToLower(d.GetWordBeforeCursor()) // 忽略大小写
		symbols := getOpenSymbolList(ctx)
		var filtered []prompt.Suggest
		for _, symbol := range symbols {
			if strings.Contains(strings.ToLower(symbol.Text), prefix) {
				filtered = append(filtered, symbol)
			}
		}
		return filtered
	}

	// 选择symbol后，选择direction
	if len(fields) == 2 && strings.HasSuffix(w, " ") {
		return getDirectionList()
	}

	// 正在输入direction（第三个token）
	if len(fields) == 3 && !strings.HasSuffix(w, " ") {
		prefix := strings.ToLower(d.GetWordBeforeCursor()) // 忽略大小写
		directions := getDirectionList()
		var filtered []prompt.Suggest
		for _, direction := range directions {
			if strings.Contains(strings.ToLower(direction.Text), prefix) {
				filtered = append(filtered, direction)
			}
		}
		return filtered
	}

	// 选择direction后，选择保证金模式
	if len(fields) == 3 && strings.HasSuffix(w, " ") {
		return getMarginModeList(ctx, fields[1])
	}

	// 正在输入保证金模式（第四个token）
	if len(fields) == 4 && !strings.HasSuffix(w, " ") {
		prefix := strings.ToLower(d.GetWordBeforeCursor())
		marginModes := getMarginModeList(ctx, fields[1])
		var filtered []prompt.Suggest
		for _, mode := range marginModes {
			if strings.Contains(strings.ToLower(mode.Text), prefix) {
				filtered = append(filtered, mode)
			}
		}
		return filtered
	}

	// 选择保证金模式后，输入amount（不显示下拉提示，等待用户输入）
	if len(fields) == 4 && strings.HasSuffix(w, " ") {
		// 返回nil表示没有补全建议，让用户直接输入amount
		return nil
	}

	// 输入amount后，可以选择with策略（选填）
	if len(fields) == 5 && strings.HasSuffix(w, " ") {
		return []prompt.Suggest{
			{Text: "with", Description: "[选填] 添加策略条件"},
		}
	}

	// 输入with后，显示策略提示
	if len(fields) == 6 && strings.HasSuffix(w, " ") && fields[5] == "with" {
		return getStrategyList()
	}

	// 正在输入策略名称时显示提示
	if len(fields) == 7 && !strings.HasSuffix(w, " ") && fields[5] == "with" {
		prefix := strings.ToLower(d.GetWordBeforeCursor())
		strategies := getStrategyList()
		var filtered []prompt.Suggest
		for _, strategy := range strategies {
			if strings.Contains(strings.ToLower(strategy.Text), prefix) {
				filtered = append(filtered, strategy)
			}
		}
		return filtered
	}

	return nil
}

// handleCloseCommandCompletion 处理 close 命令的补全
func handleCloseCommandCompletion(ctx *Context, d prompt.Document, w string, fields []string, first string) []prompt.Suggest {
	if first != "close" {
		return nil
	}

	// 选择symbol：第二个token位置
	if len(fields) == 1 && strings.HasSuffix(w, " ") {
		return getOpenSymbolList(ctx)
	}

	// 正在输入symbol（第二个token）
	if len(fields) == 2 && !strings.HasSuffix(w, " ") {
		prefix := strings.ToLower(d.GetWordBeforeCursor()) // 忽略大小写
		symbols := getOpenSymbolList(ctx)
		var filtered []prompt.Suggest
		for _, symbol := range symbols {
			if strings.Contains(strings.ToLower(symbol.Text), prefix) {
				filtered = append(filtered, symbol)
			}
		}
		return filtered
	}

	// 选择symbol后，选择direction
	if len(fields) == 2 && strings.HasSuffix(w, " ") {
		return getDirectionList()
	}

	// 正在输入direction（第三个token）
	if len(fields) == 3 && !strings.HasSuffix(w, " ") {
		prefix := strings.ToLower(d.GetWordBeforeCursor()) // 忽略大小写
		directions := getDirectionList()
		var filtered []prompt.Suggest
		for _, direction := range directions {
			if strings.Contains(strings.ToLower(direction.Text), prefix) {
				filtered = append(filtered, direction)
			}
		}
		return filtered
	}

	// 选择direction后，选择保证金模式
	if len(fields) == 3 && strings.HasSuffix(w, " ") {
		return getMarginModeList(ctx, fields[1])
	}

	// 正在输入保证金模式（第四个token）
	if len(fields) == 4 && !strings.HasSuffix(w, " ") {
		prefix := strings.ToLower(d.GetWordBeforeCursor())
		marginModes := getMarginModeList(ctx, fields[1])
		var filtered []prompt.Suggest
		for _, mode := range marginModes {
			if strings.Contains(strings.ToLower(mode.Text), prefix) {
				filtered = append(filtered, mode)
			}
		}
		return filtered
	}

	// 选择保证金模式后，输入amount（不显示下拉提示，等待用户输入）
	if len(fields) == 4 && strings.HasSuffix(w, " ") {
		// 返回nil表示没有补全建议，让用户直接输入amount
		return nil
	}

	// 输入amount后，可以选择with策略（选填）
	if len(fields) == 5 && strings.HasSuffix(w, " ") {
		return []prompt.Suggest{
			{Text: "with", Description: "[选填] 添加策略条件"},
		}
	}

	// 输入with后，显示策略提示
	if len(fields) == 6 && strings.HasSuffix(w, " ") && fields[5] == "with" {
		return getStrategyList()
	}

	// 正在输入策略名称时显示提示
	if len(fields) == 7 && !strings.HasSuffix(w, " ") && fields[5] == "with" {
		prefix := strings.ToLower(d.GetWordBeforeCursor())
		strategies := getStrategyList()
		var filtered []prompt.Suggest
		for _, strategy := range strategies {
			if strings.Contains(strings.ToLower(strategy.Text), prefix) {
				filtered = append(filtered, strategy)
			}
		}
		return filtered
	}

	return nil
}

// handleCancelCommandCompletion 处理 cancel 命令的补全
func handleCancelCommandCompletion(ctx *Context, d prompt.Document, w string, fields []string, first, second string) []prompt.Suggest {
	if first != "cancel" {
		return nil
	}

	// cancel order 的订单选择
	if second == "order" {
		// 选择订单：第三个token位置
		if len(fields) == 2 && strings.HasSuffix(w, " ") {
			return getCancelOrderList(ctx)
		}

		// 正在输入订单（第三个token）
		if len(fields) == 3 && !strings.HasSuffix(w, " ") {
			prefix := strings.ToLower(d.GetWordBeforeCursor())
			orders := getCancelOrderList(ctx)
			var filtered []prompt.Suggest
			for _, order := range orders {
				if strings.Contains(strings.ToLower(order.Text), prefix) {
					filtered = append(filtered, order)
				}
			}
			return filtered
		}
	}

	return nil
}

// getCancelOrderList 获取可取消的订单列表（暂时使用mock数据）
func getCancelOrderList(ctx *Context) []prompt.Suggest {

	// 暂时使用mock数据，确保有数据显示
	mockOrders := []struct {
		symbol   string
		side     string
		amount   float64
		strategy string
		status   string
	}{
		{"BTC-USDT", "long", 10.5, "avg", "waiting"},
		{"ETH-USDT", "short", 5.0, "", "waiting"},
		{"ADA-USDT", "long", 1000.0, "limit", "pending"},
		{"SOL-USDT", "short", 50.0, "stop", "waiting"},
		{"DOGE-USDT", "long", 50000.0, "", "pending"},
	}

	var suggestions []prompt.Suggest
	for _, order := range mockOrders {
		// 只显示未完成和未取消的订单
		if order.status == "cancelled" || order.status == "completed" {
			continue
		}

		// 构建订单标识：symbol:direction:amount
		orderText := fmt.Sprintf("%s:%s:%.4f", order.symbol, order.side, order.amount)

		// 构建描述信息
		var description string
		if order.strategy != "" {
			description = fmt.Sprintf("策略: %s", order.strategy)
		} else {
			description = "暂无订单策略"
		}

		suggestions = append(suggestions, prompt.Suggest{
			Text:        orderText,
			Description: description,
		})
	}

	return suggestions
}

// getOpenSymbolList 获取open命令的symbol列表（暂时使用mock数据）
func getOpenSymbolList(ctx *Context) []prompt.Suggest {
	exchangeName := ctx.GetExchangeName()
	symbolList, exist := config.ExchangeSymbolList[exchangeName]
	if !exist {
		return []prompt.Suggest{}
	}

	var suggestions []prompt.Suggest
	for _, symbol := range symbolList {
		suggestions = append(suggestions, prompt.Suggest{
			Text: symbol.Name,
			Description: fmt.Sprintf("最大杠杆:%.0fx 最小购买量:%.4f",
				symbol.MaxLever,
				symbol.MinSize),
		})
	}

	return suggestions
}

// getDirectionList 获取方向选择列表
func getDirectionList() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "long", Description: "多头 - 买入做多"},
		{Text: "short", Description: "空头 - 卖出做空"},
	}
}

// getMarginModeList 获取保证金模式列表
func getMarginModeList(ctx *Context, symbol string) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "isolated", Description: "逐仓模式 - 仅使用仓位保证金"},
		{Text: "cross", Description: "全仓模式 - 使用账户全部可用余额"},
	}
}

// getStrategyList 获取策略列表
func getStrategyList() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "avg", Description: "均价策略 - 在指定价格区间内分批买入"},
		{Text: "limit", Description: "限价策略 - 在指定价格时执行"},
		{Text: "stop", Description: "止损策略 - 价格触发时执行"},
		{Text: "time", Description: "时间策略 - 在指定时间执行"},
	}
}

// useExchangesList 获取交易所列表（用于use exchange命令）
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

// useAccountsList 获取账户列表（用于use account命令）
func useAccountsList() []prompt.Suggest {
	// 获取所有用户列表
	accountList, err := repository.ListAccount()
	if err != nil {
		return []prompt.Suggest{}
	}
	accounts := make([]prompt.Suggest, 0)
	for _, account := range accountList {
		accounts = append(accounts, prompt.Suggest{
			Text:        account.Name,
			Description: fmt.Sprintf("%s：%s", account.TradeType, account.Exchange),
		})
	}

	return accounts
}

// deleteAccountsList 获取账户列表（用于delete account命令）
func deleteAccountsList(ctx *Context) []prompt.Suggest {

	// 获取指定交易所的所有用户列表
	accountList, err := repository.ExchangeAccountList(ctx.GetExchangeName())
	if err != nil {
		return []prompt.Suggest{}
	}

	accounts := make([]prompt.Suggest, 0)
	for _, account := range accountList {
		accounts = append(accounts, prompt.Suggest{
			Text:        account.Name,
			Description: fmt.Sprintf("%s：%s", account.TradeType, account.Exchange),
		})
	}

	return accounts
}
