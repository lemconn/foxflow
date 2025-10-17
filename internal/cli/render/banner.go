package render

import (
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/utils"
)

// RenderBanner 渲染启动横幅（右下角显示版本）
func RenderBanner(version string) string {

	lines := `
================================================================
  ███████╗ ██████╗ ██╗  ██╗███████╗██╗      ██████╗ ██╗    ██╗
  ██╔════╝██╔═══██╗╚██╗██╔╝██╔════╝██║     ██╔═══██╗██║    ██║
  █████╗  ██║   ██║ ╚███╔╝ █████╗  ██║     ██║   ██║██║ █╗ ██║
  ██╔══╝  ██║   ██║ ██╔██╗ ██╔══╝  ██║     ██║   ██║██║███╗██║
  ██║     ╚██████╔╝██╔╝ ██╗██║     ███████╗╚██████╔╝╚███╔███╔╝
  ╚═╝      ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚══════╝ ╚═════╝  ╚══╝╚══╝
                                               version: %s
  %s
================================================================
`

	return fmt.Sprintf(lines, utils.MessageGreen(version), utils.MessageBlue("欢迎使用 FoxFlow 策略下单系统"))
}

// RenderWelcomeHints 渲染 CLI 启动提示文案
func RenderWelcomeHints() string {
	lines := []string{
		utils.RenderInfo("输入 'help' 查看可用命令"),
		utils.RenderInfo("输入命令时按 Tab 键可查看自动补全和选项"),
		utils.RenderInfo("输入 'exit' 或 'quit' 或 'Ctrl-D' 退出程序"),
	}
	return strings.Join(lines, "\n") + "\n\n"
}
