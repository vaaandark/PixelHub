#!/bin/bash

# Docker 构建脚本

IMAGE_NAME="pixelhub"
VERSION=${1:-latest}

echo "🐳 构建 Docker 镜像: ${IMAGE_NAME}:${VERSION}"

docker build -t ${IMAGE_NAME}:${VERSION} .

echo "✅ 构建完成！"
echo ""
echo "运行容器："
echo "  docker run -d -p 8080:8080 -v \$(pwd)/config.toml:/app/config.toml -v \$(pwd)/data:/app/data ${IMAGE_NAME}:${VERSION}"

