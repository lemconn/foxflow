package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lemconn/foxflow/internal/cli/command"
	cliCmds "github.com/lemconn/foxflow/internal/cli/commands"
	"github.com/lemconn/foxflow/internal/cli/render"
	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/repository"
	"github.com/lemconn/foxflow/internal/utils"

	"github.com/c-bata/go-prompt"
)

// CLI 命令行界面
type CLI struct {
	ctx      *Context
	commands map[string]command.Command
}

// NewCLI 创建新的CLI实例
func NewCLI() (*CLI, error) {
	ctx := NewContext(context.Background())

	// 初始化交易所交易对数据
	InitExchangeSymbols()

	// 注册命令
	cmdMap := map[string]command.Command{
		"help":   &cliCmds.HelpCommand{},
		"show":   &cliCmds.ShowCommand{},
		"use":    &cliCmds.UseCommand{},
		"create": &cliCmds.CreateCommand{},
		"update": &cliCmds.UpdateCommand{},
		"open":   &cliCmds.OpenCommand{},
		"close":  &cliCmds.CloseCommand{},
		"cancel": &cliCmds.CancelCommand{},
		"delete": &cliCmds.DeleteCommand{},
		"exit":   &cliCmds.ExitCommand{},
		"quit":   &cliCmds.ExitCommand{},
	}

	return &CLI{
		ctx:      ctx,
		commands: cmdMap,
	}, nil
}

// Run 运行CLI
func (c *CLI) Run() error {
	fmt.Print(render.RenderWelcomeHints())

	// 显示操作指南
	fmt.Println("快捷键说明:")
	fmt.Println()

	// 设置默认交易所
	c.setDefaultExchange()

	p := prompt.New(
		c.executor,
		getCompleter(c.ctx),
		prompt.OptionTitle("foxflow"),
		prompt.OptionPrefix("> "),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionCompletionWordSeparator(" "),
		prompt.OptionSuggestionTextColor(prompt.LightGray),         // 设置下拉文字为白色，简洁明亮
		prompt.OptionSuggestionBGColor(prompt.DarkGray),            // 设置下拉背景为深灰
		prompt.OptionSelectedSuggestionTextColor(prompt.DarkGray),  // 选中项文字为白色
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),   // 选中项背景为深灰色
		prompt.OptionDescriptionTextColor(prompt.LightGray),        // 设置描述文字为深灰色
		prompt.OptionDescriptionBGColor(prompt.DarkGray),           // 设置描述背景为浅灰色
		prompt.OptionSelectedDescriptionTextColor(prompt.DarkGray), // 选中项描述文字为浅灰色
		prompt.OptionSelectedDescriptionBGColor(prompt.LightGray),  // 选中项描述背景为浅灰色
		prompt.OptionAddKeyBind(
			prompt.KeyBind{
				Key: prompt.ControlC,
				Fn: func(b *prompt.Buffer) {
					// 捕获 Ctrl+C：刷新一行彩色状态信息（当前已激活交易所/用户信息），给出退出操作的提示
					c.printStatus()
					fmt.Println(utils.RenderWarning("请输入 'exit' 或 'quit' 或 'Ctrl+D' 退出程序"))
				},
			},
		),
	)

	// 启动前刷新一行彩色状态信息（当前已激活交易所/用户信息）
	c.printStatus()

	p.Run()
	return nil
}

// executor 执行器：处理一行输入
func (c *CLI) executor(in string) {
	line := strings.TrimSpace(in)
	if line == "" {
		// 输入空数据回车刷新一行彩色状态信息（当前已激活交易所/用户信息）
		c.printStatus()
		return
	}

	args := parseArgs(line)
	if len(args) == 0 {
		return
	}

	if args[0] == "quit" || args[0] == "exit" {
		fmt.Println(utils.RenderInfo("再见！"))
		os.Exit(0)
		return
	}

	if err := c.executeCommand(args); err != nil {
		if err.Error() == "exit" {
			fmt.Println(utils.RenderInfo("再见！"))
			os.Exit(0)
			return
		}
		fmt.Println()
		fmt.Println(utils.RenderError(fmt.Sprintf("错误: %v", err)))
	}

	// 执行完成输出信息后刷新一行彩色状态（当前交易所和用户信息）
	fmt.Println()
	c.printStatus()
}

// executeCommand 执行命令
func (c *CLI) executeCommand(args []string) error {
	commandName := strings.ToLower(args[0]) // 忽略大小写
	command, exists := c.commands[commandName]
	if !exists {
		return fmt.Errorf("未知命令: %s", args[0])
	}

	return command.Execute(c.ctx, args[1:])
}

// parseArgs 解析命令行参数
func parseArgs(line string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := '"'
	withFound := false

	// 首先检查是否包含 "with" 关键字
	words := strings.Fields(line)
	for _, word := range words {
		if strings.ToLower(word) == "with" {
			withFound = true
			break
		}
	}

	// 如果没有找到 "with"，使用原来的解析逻辑
	if !withFound {
		return parseArgsOriginal(line)
	}

	// 如果找到 "with"，特殊处理
	for i, char := range line {
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			} else {
				current.WriteRune(char)
			}
		case ' ':
			if !inQuotes {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
					
					// 如果当前参数是 "with"，则后续所有内容作为一个整体
					if strings.ToLower(args[len(args)-1]) == "with" {
						// 获取 "with" 后面的所有内容
						remaining := strings.TrimSpace(line[i+1:])
						if remaining != "" {
							args = append(args, remaining)
						}
						return args
					}
				}
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// parseArgsOriginal 原始的参数解析逻辑
func parseArgsOriginal(line string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := '"'

	for _, char := range line {
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			} else {
				current.WriteRune(char)
			}
		case ' ':
			if !inQuotes {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func (c *CLI) setDefaultExchange() {

	err := c.useActiveAccount()
	if err == nil {
		return
	}

	exchangesList, err := repository.ListExchanges()
	if err != nil {
		log.Printf("set default exchanges list error: %v\n", err)
		return
	}

	// 没有默认值则直接使用第一个交易所
	exchangeName := exchangesList[0].Name
	for _, exchange := range exchangesList {

		// 默认交易所优先级次之
		if exchange.Name == config.DefaultExchange {
			exchangeName = exchange.Name
		}

		// 已经激活的交易所优先级最高
		if exchange.IsActive {
			exchangeName = exchange.Name
			break
		}
	}

	// 初始化设置默认交易所
	useCommand := &cliCmds.UseCommand{}
	err = useCommand.Execute(c.ctx, []string{"exchange", exchangeName})
	if err != nil {
		log.Printf("set default exchange execute error: %v\n", err)
	}
}

func (c *CLI) useActiveAccount() error {

	activeAccount, err := repository.ActiveAccount()
	if err != nil {
		log.Printf("Failed to obtain activation account: %v\n", err)
		return err
	}

	if activeAccount.Name == "" {
		log.Printf("No active account found")
		return nil
	}

	// 初始化设置默认交易所
	useCommand := &cliCmds.UseCommand{}
	err = useCommand.Execute(c.ctx, []string{"account", activeAccount.Name})
	if err != nil {
		log.Printf("set default exchange execute error: %v\n", err)
		return err
	}

	return nil
}

// 额外：在提示行上方打印一行彩色状态，作为多色前缀替代
func (c *CLI) printStatus() {
	exchangeName := c.ctx.GetExchangeName()
	account := c.ctx.GetAccountInstance()

	if exchangeName == "" {
		fmt.Println(utils.MessageGreen("foxflow ") +
			utils.MessagePurple("["+time.Now().Format(config.DateFormat)+"]"))
		return
	}

	if account == nil || account.Name == "" {
		fmt.Println(utils.MessageGreen("foxflow ") +
			utils.MessageYellow("["+exchangeName+"] ") +
			utils.MessagePurple("["+time.Now().Format(config.DateFormat)+"]"))
		return
	}

	fmt.Println(utils.MessageGreen("foxflow ") +
		utils.MessageYellow("["+exchangeName+":"+account.TradeType+"@"+account.Name+"] ") +
		utils.MessagePurple("["+time.Now().Format(config.DateFormat)+"]"))
}
