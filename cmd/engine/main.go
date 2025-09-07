package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"foxflow/internal/config"
	"foxflow/internal/database"
	"foxflow/internal/engine"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 创建策略引擎
	engineInstance := engine.NewEngine()

	// 启动引擎
	if err := engineInstance.Start(); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}

	log.Println("策略引擎已启动，按 Ctrl+C 停止")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("收到停止信号，正在关闭引擎...")

	// 停止引擎
	if err := engineInstance.Stop(); err != nil {
		log.Printf("Failed to stop engine: %v", err)
		os.Exit(1)
	}

	log.Println("引擎已安全关闭")
}
