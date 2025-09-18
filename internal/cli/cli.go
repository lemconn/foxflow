package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/repository"

	"github.com/lemconn/foxflow/internal/cli/command"
	cliCmds "github.com/lemconn/foxflow/internal/cli/commands"
	"github.com/lemconn/foxflow/internal/cli/render"
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

	// 注册命令
	cmdMap := map[string]command.Command{
		"show":   &cliCmds.ShowCommand{},
		"use":    &cliCmds.UseCommand{},
		"create": &cliCmds.CreateCommand{},
		"update": &cliCmds.UpdateCommand{},
		"cancel": &cliCmds.CancelCommand{},
		"delete": &cliCmds.DeleteCommand{},
		"help":   &cliCmds.HelpCommand{},
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

	// 设置默认交易所
	c.setDefaultExchange()

	p := prompt.New(
		c.executor,
		getCompleter(c.ctx, c.commands),
		prompt.OptionTitle("foxflow"),
		prompt.OptionPrefix("> "),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionCompletionWordSeparator(" "),
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
	commandName := args[0]
	command, exists := c.commands[commandName]
	if !exists {
		return fmt.Errorf("未知命令: %s", commandName)
	}

	return command.Execute(c.ctx, args[1:])
}

// parseArgs 解析命令行参数
func parseArgs(line string) []string {
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
	err = useCommand.Execute(c.ctx, []string{"exchanges", exchangeName})
	if err != nil {
		log.Printf("set default exchange execute error: %v\n", err)
	}
}

// 额外：在提示行上方打印一行彩色状态，作为多色前缀替代
func (c *CLI) printStatus() {
	exchangeName := c.ctx.GetExchangeName()
	user := c.ctx.GetUserInstance()

	if exchangeName == "" {
		fmt.Println(utils.MessageGreen("foxflow ") +
			utils.MessagePurple("["+time.Now().Format(config.DateFormat)+"]"))
		return
	}

	if user == nil || user.Username == "" {
		fmt.Println(utils.MessageGreen("foxflow ") +
			utils.MessageYellow("["+exchangeName+"] ") +
			utils.MessagePurple("["+time.Now().Format(config.DateFormat)+"]"))
		return
	}

	fmt.Println(utils.MessageGreen("foxflow ") +
		utils.MessageYellow("["+exchangeName+":"+user.TradeType+"@"+user.Username+"] ") +
		utils.MessagePurple("["+time.Now().Format(config.DateFormat)+"]"))
}
