#!/bin/bash

# FoxFlow 构建脚本

echo "=== FoxFlow 构建脚本 ==="

# 创建构建目录
mkdir -p bin

echo "编译 CLI 程序..."
go build -o bin/foxflow-cli ./cmd/cli
if [ $? -ne 0 ]; then
    echo "CLI 编译失败"
    exit 1
fi

echo "编译 Engine 程序..."
go build -o bin/foxflow-engine ./cmd/engine
if [ $? -ne 0 ]; then
    echo "Engine 编译失败"
    exit 1
fi

echo "构建完成！"
echo "可执行文件位于 bin/ 目录："
echo "  - bin/foxflow-cli    (CLI 程序)"
echo "  - bin/foxflow-engine (引擎程序)"
