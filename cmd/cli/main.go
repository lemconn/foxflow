package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lemconn/foxflow/internal/cli/render"
	"github.com/lemconn/foxflow/internal/utils"

	"github.com/lemconn/foxflow/internal/cli"
	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/grpc"
)

func main() {
	// 解析命令行参数
	var (
		host     = flag.String("H", "127.0.0.1", "服务端地址")
		port     = flag.Int("P", 1259, "服务端端口")
		username = flag.String("u", "foxflow", "用户名")
		password = flag.String("p", "foxflow", "密码")
		dbFile   = flag.String("db", "", "SQLite数据库文件路径（例如：./foxflow-1.db 或 /var/lib/foxflow/foxflow-1.db）")
	)
	flag.Parse()

	// 输出产品
	fmt.Println(render.RenderBanner(config.Version))

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *dbFile != "" {
		config.GlobalConfig.DBFile = *dbFile
	}

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println(utils.RenderSuccess("数据库初始化完成"))

	// 尝试连接gRPC服务端
	var grpcClient *grpc.Client
	if *host != "" && *port > 0 {
		client, err := grpc.NewClient(*host, *port)
		if err != nil {
			fmt.Println(utils.RenderWarning(fmt.Sprintf("无法连接到gRPC服务端 (%s:%d), 将以本地模式运行", *host, *port)))
		} else {
			// 尝试认证
			if err := client.Authenticate(*username, *password); err != nil {
				// 检查是否是连接错误
				if strings.Contains(err.Error(), "connection failed") {
					fmt.Println(utils.RenderWarning(fmt.Sprintf("无法连接到gRPC服务端 (%s:%d), 将以本地模式运行", *host, *port)))
				} else {
					fmt.Println(utils.RenderWarning(fmt.Sprintf("用户 %s 认证失败, 将以本地模式运行", *username)))
				}
				client.Close()
			} else {
				grpcClient = client
				fmt.Println(utils.RenderSuccess(fmt.Sprintf("已连接到gRPC服务端 %s:%d", *host, *port)))
			}
		}
	}

	// 创建CLI实例
	cliInstance, err := cli.NewCLI()
	if err != nil {
		log.Fatalf("Failed to create CLI: %v", err)
	}

	// 设置gRPC客户端
	if grpcClient != nil {
		cliInstance.SetGRPCClient(grpcClient)
	}

	// 运行CLI
	if err := cliInstance.Run(); err != nil {
		log.Printf("CLI error: %v", err)
		os.Exit(1)
	}

	// 关闭gRPC连接
	if grpcClient != nil {
		grpcClient.Close()
	}
}
