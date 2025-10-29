# FoxFlow Makefile

.PHONY: build clean test cli engine help proto test-grpc

# 默认目标
all: build

# 构建所有程序
build: proto
	@echo "构建 FoxFlow 程序..."
	@echo "=== FoxFlow 构建脚本 ==="
	@echo "编译 CLI 程序..."
	@mkdir -p bin
	@go build -o bin/foxflow-cli ./cmd/cli
	@echo "编译 Engine 程序..."
	@go build -o bin/foxflow-engine ./cmd/engine
	@echo "构建完成！"
	@echo "可执行文件位于 bin/ 目录："
	@echo "  - bin/foxflow-cli    (CLI 程序)"
	@echo "  - bin/foxflow-engine (引擎程序)"

# 构建 CLI 程序
cli:
	@echo "构建 CLI 程序..."
	@mkdir -p bin
	@go build -o bin/foxflow-cli ./cmd/cli

# 构建引擎程序
engine:
	@echo "构建引擎程序..."
	@mkdir -p bin
	@go build -o bin/foxflow-engine ./cmd/engine

# 运行 CLI 程序
run-cli: cli
	@echo "启动 CLI 程序..."
	@./bin/foxflow-cli

# 运行引擎程序
run-engine: engine
	@echo "启动引擎程序..."
	@./bin/foxflow-engine

# 测试程序
test:
	@echo "运行所有测试..."
	@go test -v ./internal/grpc/...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf bin/
	@rm -f .foxflow.db
	@rm -f .foxflow.history

# 生成 proto 代码
proto:
	@echo "生成 gRPC 代码..."
	@if ! command -v protoc &> /dev/null; then \
		echo "错误: protoc 未安装，请先安装 Protocol Buffers 编译器"; \
		echo "安装方法:"; \
		echo "  macOS: brew install protobuf"; \
		echo "  Ubuntu: sudo apt-get install protobuf-compiler"; \
		echo "  或访问: https://grpc.io/docs/protoc-installation/"; \
		exit 1; \
	fi
	@if ! command -v protoc-gen-go &> /dev/null; then \
		echo "错误: protoc-gen-go 未安装"; \
		echo "安装方法: go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0"; \
		exit 1; \
	fi
	@if ! command -v protoc-gen-go-grpc &> /dev/null; then \
		echo "错误: protoc-gen-go-grpc 未安装"; \
		echo "安装方法: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0"; \
		exit 1; \
	fi
	@mkdir -p proto/generated
	@echo "正在生成 gRPC Go 代码..."
	@protoc \
		--proto_path=proto \
		--go_out=proto/generated \
		--go_opt=paths=source_relative \
		--go-grpc_out=proto/generated \
		--go-grpc_opt=paths=source_relative \
		proto/foxflow.proto
	@echo "gRPC Go 代码生成完成！"
	@echo "生成的文件:"
	@echo "  - proto/generated/foxflow.pb.go"
	@echo "  - proto/generated/foxflow_grpc.pb.go"

# 安装依赖
deps:
	@echo "安装依赖..."
	@go mod tidy
	@go mod download

# 格式化代码
fmt:
	@echo "格式化代码..."
	@go fmt ./...

# 检查代码
vet:
	@echo "检查代码..."
	@go vet ./...

# 显示帮助
help:
	@echo "FoxFlow 可用命令："
	@echo "  build      - 构建所有程序"
	@echo "  cli        - 构建 CLI 程序"
	@echo "  engine     - 构建引擎程序"
	@echo "  run-cli    - 运行 CLI 程序"
	@echo "  run-engine - 运行引擎程序"
	@echo "  test       - 运行所有测试"
	@echo "  proto      - 生成 gRPC 代码"
	@echo "  clean      - 清理构建文件"
	@echo "  deps       - 安装依赖"
	@echo "  fmt        - 格式化代码"
	@echo "  vet        - 检查代码"
	@echo "  help       - 显示此帮助信息"

