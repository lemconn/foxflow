package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/engine"
	"github.com/lemconn/foxflow/internal/grpc"
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

	// 启动gRPC服务端
	grpcServer := grpc.NewServer(1259)   // 默认端口1259
	grpcServer.SetEngine(engineInstance) // 设置引擎实例
	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Printf("gRPC服务端启动失败: %v", err)
		}
	}()

	log.Println("策略引擎已启动，gRPC服务端已启动，按 Ctrl+C 停止")

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
