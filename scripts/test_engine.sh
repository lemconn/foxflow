#!/bin/bash

# FoxFlow Engine 测试脚本

echo "=== FoxFlow Engine 测试脚本 ==="

# 检查是否已编译
if [ ! -f "./bin/foxflow-engine" ]; then
    echo "编译 Engine 程序..."
    mkdir -p bin
    go build -o bin/foxflow-engine ./cmd/engine
    if [ $? -ne 0 ]; then
        echo "编译失败"
        exit 1
    fi
fi

echo "启动 Engine 程序..."
echo "引擎将每5秒检查一次策略订单"
echo "按 Ctrl+C 停止引擎"
echo ""

# 启动引擎程序
./bin/foxflow-engine
