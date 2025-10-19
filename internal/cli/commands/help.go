package commands

import (
	"fmt"

	"github.com/lemconn/foxflow/internal/cli/command"
)

// getHelpCommandSuggestions 获取所有命令的帮助建议
func getHelpCommandSuggestions() []struct {
	Text        string
	Description string
} {
	return []struct {
		Text        string
		Description string
	}{
		{Text: "help", Description: "显示帮助信息 - 查看所有命令说明"},
		{Text: "show", Description: "查看数据列表 - 支持子命令：exchange(交易所)、account(账户)、balance(资产)、position(持仓)、symbol(交易对)、strategy(策略)、order(订单)、news(新闻)"},
		{Text: "use", Description: "激活上下文 - 支持子命令：exchange(激活交易所)、account(激活账户)"},
		{Text: "create", Description: "创建资源 - 支持子命令：users(用户)、symbols(交易对)、ss(策略订单)"},
		{Text: "update", Description: "更新配置 - 支持子命令：leverage(杠杆)"},
		{Text: "open", Description: "开仓/下单 - 执行交易开仓操作"},
		{Text: "close", Description: "平仓 - 执行交易平仓操作"},
		{Text: "cancel", Description: "取消订单 - 支持子命令：ss(策略订单)"},
		{Text: "delete", Description: "删除资源 - 支持子命令：users(用户)、symbols(交易对)"},
		{Text: "exit", Description: "退出系统"},
		{Text: "quit", Description: "退出系统"},
	}
}

// HelpCommand 帮助命令
type HelpCommand struct{}

func (c *HelpCommand) GetName() string        { return "help" }
func (c *HelpCommand) GetDescription() string { return "显示帮助信息" }
func (c *HelpCommand) GetUsage() string       { return "help [command]" }

func (c *HelpCommand) Execute(ctx command.Context, args []string) error {
	
	// 显示系统描述
	fmt.Println("欢迎使用 FoxFlow 交易系统")
	fmt.Println("================================================================")
	fmt.Println("FoxFlow 是一个强大的量化交易终端，支持多交易所、多策略的自动化交易。")
	fmt.Println("系统提供完整的交易生命周期管理，包括账户管理、策略部署、订单执行等功能。")
	fmt.Println("")
	fmt.Println("主要特性：")
	fmt.Println("• 多交易所支持：同时连接多个交易所进行交易")
	fmt.Println("• 策略管理：灵活的策略模板和订单管理系统")
	fmt.Println("• 风险控制：完善的杠杆和保证金管理机制")
	fmt.Println("• 实时监控：资产、持仓、订单状态实时监控")
	fmt.Println("• 金融资讯：集成实时金融新闻和市场数据")
	fmt.Println("")

	// 显示所有可用命令
	fmt.Println("可用命令列表：")
	fmt.Println("================================================================")
	commandSuggestions := getHelpCommandSuggestions()
	for _, suggestion := range commandSuggestions {
		fmt.Printf("  %-10s - %s\n", suggestion.Text, suggestion.Description)
	}
	fmt.Println("")
	fmt.Println("操作指南：")
	fmt.Println("• 输入命令时按 Tab 键可查看自动补全和下拉选择")
	fmt.Println("• 使用 ↑↓ 箭头键在历史记录中导航")
	fmt.Println("• 使用 Ctrl+R 进入历史记录搜索模式")
	fmt.Println("• 输入 'help <command>' 查看特定命令的详细帮助")

	return nil
}
