# FoxFlow Makefile

.PHONY: build clean test cli engine help

# 默认目标
all: build

# 构建所有程序
build:
	@echo "构建 FoxFlow 程序..."
	@./scripts/build.sh

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
	@echo "运行测试..."
	@./scripts/quick_test.sh

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf bin/
	@rm -f .foxflow.db
	@rm -f .foxflow.history

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
	@echo "  test       - 运行测试"
	@echo "  clean      - 清理构建文件"
	@echo "  deps       - 安装依赖"
	@echo "  fmt        - 格式化代码"
	@echo "  vet        - 检查代码"
	@echo "  help       - 显示此帮助信息"
