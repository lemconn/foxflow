package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lemconn/foxflow/internal/cli/render"
	"github.com/lemconn/foxflow/internal/utils"

	"github.com/lemconn/foxflow/internal/cli"
	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/database"
)

func main() {

	// 输出产品
	fmt.Println(render.RenderBanner(config.Version))

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println(utils.RenderSuccess("数据库初始化完成"))

	// 创建CLI实例
	cliInstance, err := cli.NewCLI()
	if err != nil {
		log.Fatalf("Failed to create CLI: %v", err)
	}

	// 运行CLI
	if err := cliInstance.Run(); err != nil {
		log.Printf("CLI error: %v", err)
		os.Exit(1)
	}
}
