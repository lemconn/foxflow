#!/bin/bash

# FoxFlow CLI 测试脚本

echo "=== FoxFlow CLI 测试脚本 ==="

# 检查是否已编译
if [ ! -f "./bin/foxflow-cli" ]; then
    echo "编译 CLI 程序..."
    mkdir -p bin
    go build -o bin/foxflow-cli ./cmd/cli
    if [ $? -ne 0 ]; then
        echo "编译失败"
        exit 1
    fi
fi

echo "启动 CLI 程序..."
echo "测试命令序列："
echo "1. show exchanges"
echo "2. use exchanges okx"
echo "3. create users --username=testuser --ak=testkey --sk=testsecret --trade_type=mock"
echo "4. use users testuser"
echo "5. show users"
echo "6. show strategies"
echo "7. create ss --symbol=BTC/USDT --side=buy --posSide=long --px=50000 --sz=0.01 --strategy=volume:volume>100"
echo "8. show ss"
echo "9. exit"

echo ""
echo "请手动执行上述命令进行测试"
echo ""

# 启动CLI程序
./bin/foxflow-cli
