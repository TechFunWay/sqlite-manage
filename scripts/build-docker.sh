#!/bin/bash

set -e

PROJECT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$PROJECT_DIR"

VERSION=$(grep -o 'const Version = "[^"]*"' backend/upgrade/upgrade.go | cut -d'"' -f2)
[ -z "$VERSION" ] && echo "❌ 无法获取版本号" && exit 1

APP_NAME="techfunway-sqlite-manage"
IMAGE_NAME="techfunways/sqlite-manage"
RELEASE_DIR="release/v${VERSION}"

[ ! -d "${RELEASE_DIR}/${APP_NAME}-${VERSION}-linux-amd64" ] && echo "❌ 请先运行 ./scripts/build-all.sh" && exit 1

echo "============================================"
echo "  Docker 多平台镜像构建 v${VERSION}"
echo "============================================"

# 创建 Dockerfile
cat > "${RELEASE_DIR}/Dockerfile" << 'EOF'
FROM scratch

ARG TARGETARCH
ARG VERSION

# 复制可执行文件
COPY sqlite-manager /app/sqlite-manager

# 复制静态资源
COPY www/ /app/www/

# 创建数据目录结构
VOLUME ["/data"]

EXPOSE 8080

ENV SQLITE_DATA_DIR=/data
ENV SQLITE_UPLOAD_DIR=/data/upload
ENV SQLITE_WEB_DIR=/app/www

WORKDIR /app

ENTRYPOINT ["/app/sqlite-manager"]
CMD ["-port", "8080", "-data-dir", "/data", "-upload-dir", "/data/upload", "-web-dir", "/app/www"]
EOF

echo ""
echo "🔨 使用 Docker Buildx 构建多平台镜像..."
echo ""

# 切换到 default context
docker context use default 2>/dev/null || true

cd "${RELEASE_DIR}"

# 构建 amd64
echo -n "  📦 amd64... "
mkdir -p .build-amd64/www
cp Dockerfile .build-amd64/
cp ${APP_NAME}-${VERSION}-linux-amd64/sqlite-manager .build-amd64/
cp -r ${APP_NAME}-${VERSION}-linux-amd64/www/* .build-amd64/www/
cd .build-amd64
docker build --platform linux/amd64 -t "${IMAGE_NAME}:v${VERSION}-amd64" . > /dev/null 2>&1
cd ..
rm -rf .build-amd64
echo "✅"

# 构建 arm64
echo -n "  📦 arm64... "
mkdir -p .build-arm64/www
cp Dockerfile .build-arm64/
cp ${APP_NAME}-${VERSION}-linux-arm64/sqlite-manager .build-arm64/
cp -r ${APP_NAME}-${VERSION}-linux-arm64/www/* .build-arm64/www/
cd .build-arm64
docker build --platform linux/arm64 -t "${IMAGE_NAME}:v${VERSION}-arm64" . > /dev/null 2>&1
cd ..
rm -rf .build-arm64
echo "✅"

# 合并为多平台镜像
echo ""
echo "📦 合并多平台镜像..."
docker manifest rm "${IMAGE_NAME}:v${VERSION}" 2>/dev/null || true
docker manifest rm "${IMAGE_NAME}:latest" 2>/dev/null || true

docker manifest create "${IMAGE_NAME}:v${VERSION}" \
    "${IMAGE_NAME}:v${VERSION}-amd64" \
    "${IMAGE_NAME}:v${VERSION}-arm64" 2>/dev/null || true

docker manifest create "${IMAGE_NAME}:latest" \
    "${IMAGE_NAME}:v${VERSION}-amd64" \
    "${IMAGE_NAME}:v${VERSION}-arm64" 2>/dev/null || true

# 清理
rm -f Dockerfile
cd "${PROJECT_DIR}"

echo ""
echo "✅ 完成!"
echo ""
echo "📦 本地镜像:"
docker images "${IMAGE_NAME}" --format "  {{.Repository}}:{{.Tag}}" | sort -u
echo ""
echo "💡 运行: docker run -p 8080:8080 -v ./data:/data ${IMAGE_NAME}:v${VERSION}-amd64"
