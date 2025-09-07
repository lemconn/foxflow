package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"foxflow/pkg/utils"

	"github.com/chzyer/readline"
)

// CLI 命令行界面
type CLI struct {
	ctx      *Context
	commands map[string]Command
	rl       *readline.Instance
}

// NewCLI 创建新的CLI实例
func NewCLI() (*CLI, error) {
	ctx := NewContext(context.Background())

	// 注册命令
	commands := map[string]Command{
		"show":   &ShowCommand{},
		"use":    &UseCommand{},
		"create": &CreateCommand{},
		"update": &UpdateCommand{},
		"cancel": &CancelCommand{},
		"delete": &DeleteCommand{},
		"help":   &HelpCommand{},
		"exit":   &ExitCommand{},
		"quit":   &ExitCommand{},
	}

	// 创建readline实例
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          ctx.GetPrompt(),
		HistoryFile:     ".foxflow.history",
		AutoComplete:    getCompleter(commands),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create readline: %w", err)
	}

	return &CLI{
		ctx:      ctx,
		commands: commands,
		rl:       rl,
	}, nil
}

// Run 运行CLI
func (c *CLI) Run() error {
	defer c.rl.Close()

	fmt.Println(utils.RenderSuccess("数据库初始化完成"))
	fmt.Println(utils.RenderInfo("输入 'help' 查看可用命令"))
	fmt.Println(utils.RenderInfo("输入 'exit' 或 'quit' 退出程序"))
	fmt.Println()

	for {
		// 更新提示符
		c.rl.SetPrompt(c.ctx.GetPrompt())

		line, err := c.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				fmt.Println(utils.RenderInfo("提示: 请输入 'exit' 或 'quit' 退出程序"))
				continue
			} else if err == io.EOF {
				fmt.Println(utils.RenderInfo("再见！"))
				break
			}
			fmt.Println(utils.RenderError(fmt.Sprintf("读取输入错误: %v\n", err)))
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析命令
		args := parseArgs(line)
		if len(args) == 0 {
			continue
		}

		// 执行命令
		if err := c.executeCommand(args); err != nil {
			if err.Error() == "exit" {
				break
			}
			fmt.Println(utils.RenderError(fmt.Sprintf("错误: %v", err)))
		}
	}

	return nil
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

// getCompleter 获取命令补全器
func getCompleter(commands map[string]Command) readline.AutoCompleter {
	var items []readline.PrefixCompleterInterface

	// 添加命令补全
	for name := range commands {
		items = append(items, readline.PcItem(name))
	}

	// 添加子命令补全
	showItems := []readline.PrefixCompleterInterface{
		readline.PcItem("exchanges"),
		readline.PcItem("users"),
		readline.PcItem("assets"),
		readline.PcItem("orders"),
		readline.PcItem("positions"),
		readline.PcItem("strategies"),
		readline.PcItem("symbols"),
		readline.PcItem("ss"),
	}
	items = append(items, readline.PcItem("show", showItems...))

	useItems := []readline.PrefixCompleterInterface{
		readline.PcItem("exchanges"),
		readline.PcItem("users"),
	}
	items = append(items, readline.PcItem("use", useItems...))

	createItems := []readline.PrefixCompleterInterface{
		readline.PcItem("users"),
		readline.PcItem("symbols"),
		readline.PcItem("ss"),
	}
	items = append(items, readline.PcItem("create", createItems...))

	updateItems := []readline.PrefixCompleterInterface{
		readline.PcItem("leverage"),
		readline.PcItem("margin-type"),
	}
	items = append(items, readline.PcItem("update", updateItems...))

	cancelItems := []readline.PrefixCompleterInterface{
		readline.PcItem("ss"),
	}
	items = append(items, readline.PcItem("cancel", cancelItems...))

	deleteItems := []readline.PrefixCompleterInterface{
		readline.PcItem("users"),
		readline.PcItem("symbols"),
	}
	items = append(items, readline.PcItem("delete", deleteItems...))

	return readline.NewPrefixCompleter(items...)
}
