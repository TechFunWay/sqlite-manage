#!/bin/bash

# Docker 多平台镜像合并脚本
# 需要在有网络的环境下运行

set -e

PROJECT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$PROJECT_DIR"

VERSION=$(grep -o 'const Version = "[^"]*"' backend/upgrade/upgrade.go | cut -d'"' -f2)
[ -z "$VERSION" ] && echo "❌ 无法获取版本号" && exit 1

IMAGE_NAME="techfunways/sqlite-manage"

echo "============================================"
echo "  Docker 多平台镜像合并 v${VERSION}"
echo "============================================"
echo ""

# 检查镜像是否存在
if ! docker images | grep -q "${IMAGE_NAME}:v${VERSION}-amd64"; then
    echo "❌ 未找到 ${IMAGE_NAME}:v${VERSION}-amd64"
    echo "请先运行: ./scripts/build-docker.sh"
    exit 1
fi

if ! docker images | grep -q "${IMAGE_NAME}:v${VERSION}-arm64"; then
    echo "❌ 未找到 ${IMAGE_NAME}:v${VERSION}-arm64"
    echo "请先运行: ./scripts/build-docker.sh"
    exit 1
fi

echo "📦 创建 manifest..."
docker manifest rm "${IMAGE_NAME}:v${VERSION}" 2>/dev/null || true
docker manifest rm "${IMAGE_NAME}:latest" 2>/dev/null || true

docker manifest create "${IMAGE_NAME}:v${VERSION}" \
    "${IMAGE_NAME}:v${VERSION}-amd64" \
    "${IMAGE_NAME}:v${VERSION}-arm64"

docker manifest create "${IMAGE_NAME}:latest" \
    "${IMAGE_NAME}:v${VERSION}-amd64" \
    "${IMAGE_NAME}:v${VERSION}-arm64"

echo ""
echo "✅ 完成!"
echo ""
echo "📦 本地 manifest:"
docker manifest inspect "${IMAGE_NAME}:v${VERSION}" | head -20
echo ""
echo "💡 使用: docker run -p 8080:8080 -v ./data:/data ${IMAGE_NAME}:v${VERSION}"
echo ""
echo "📤 推送 (可选):"
echo "   docker manifest push ${IMAGE_NAME}:v${VERSION}"
echo "   docker manifest push ${IMAGE_NAME}:latest"
