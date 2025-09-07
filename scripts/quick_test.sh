#!/bin/bash

# FoxFlow 快速测试脚本

echo "=== FoxFlow 快速测试 ==="

# 检查构建文件
if [ ! -f "./bin/foxflow-cli" ]; then
    echo "CLI 程序不存在，正在构建..."
    ./scripts/build.sh
fi

if [ ! -f "./bin/foxflow-engine" ]; then
    echo "Engine 程序不存在，正在构建..."
    ./scripts/build.sh
fi

echo "✅ 程序构建完成"

# 检查数据库文件
if [ ! -f ".foxflow.db" ]; then
    echo "数据库文件不存在，程序启动时会自动创建"
else
    echo "✅ 数据库文件已存在"
fi

# 检查配置文件
if [ ! -f ".env" ]; then
    echo "⚠️  配置文件不存在，将使用默认配置"
    echo "   可以复制 config.env.example 为 .env 来自定义配置"
else
    echo "✅ 配置文件已存在"
fi

echo ""
echo "🎉 系统准备就绪！"
echo ""
echo "使用方法："
echo "1. 启动 CLI 程序: ./bin/foxflow-cli"
echo "2. 启动引擎程序: ./bin/foxflow-engine"
echo ""
echo "测试命令序列："
echo "  show exchanges"
echo "  use exchanges okx"
echo "  create users --username=testuser --ak=testkey --sk=testsecret --trade_type=mock"
echo "  use users testuser"
echo "  show users"
echo "  create ss --symbol=BTC/USDT --side=buy --posSide=long --px=50000 --sz=0.01 --strategy=volume:volume>100"
echo "  show ss"
echo "  exit"
