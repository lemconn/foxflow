package main

import (
	"fmt"
	"foxflow/internal/cli/render"
	"log"
	"os"

	"foxflow/internal/cli"
	"foxflow/internal/config"
	"foxflow/internal/database"
)

func main() {

	// 输出产品名称
	fmt.Println(render.RenderBanner(config.Version))

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

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
