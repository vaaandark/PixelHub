#!/bin/bash

# PixelHub 初始化脚本

set -e

echo "🚀 PixelHub 初始化脚本"
echo "===================="

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误：未安装 Go"
    echo "请访问 https://golang.org/dl/ 安装 Go 1.21 或更高版本"
    exit 1
fi

echo "✅ 检测到 Go 版本：$(go version)"

# 下载依赖
echo ""
echo "📦 下载依赖..."
go mod download
go mod tidy

# 创建配置文件
echo ""
if [ ! -f "config.toml" ]; then
    echo "📝 创建配置文件..."
    cp config.example.toml config.toml
    echo "✅ 已创建 config.toml，请编辑此文件填入你的配置"
else
    echo "ℹ️  config.toml 已存在，跳过创建"
fi

# 创建数据目录
echo ""
echo "📁 创建数据目录..."
mkdir -p data

# 编译项目
echo ""
echo "🔨 编译项目..."
mkdir -p bin
go build -o bin/pixelhub cmd/server/main.go

echo ""
echo "✅ 初始化完成！"
echo ""
echo "下一步："
echo "1. 编辑 config.toml 文件，填入你的配置信息"
echo "2. 运行 ./bin/pixelhub 或 make run 启动服务"
echo ""
echo "开发模式（需要安装 air）："
echo "  go install github.com/cosmtrek/air@latest"
echo "  air"
echo ""

