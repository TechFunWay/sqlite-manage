#!/bin/bash

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

VERSION=$(grep -o 'const Version = "[^"]*"' backend/upgrade/upgrade.go | cut -d'"' -f2)
[ -z "$VERSION" ] && echo "❌ 无法获取版本号" && exit 1
APP_NAME="techfunway-sqlite-manage"
RELEASE_DIR="release/v${VERSION}"
GOCACHE_DIR="${GOCACHE:-/tmp/sqlite-manage-fnos-go-cache}"
WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/sqlite-manage-fnpack.XXXXXX")
trap 'rm -rf "$WORK_DIR"' EXIT

command -v fnpack >/dev/null 2>&1 || { echo "❌ 未找到 fnpack"; exit 1; }
mkdir -p "$RELEASE_DIR"

echo "============================================"
echo "  仅构建飞牛 fnOS 安装包 v${VERSION}"
echo "============================================"

echo "📦 构建飞牛前端..."
VITE_FNOS_APP=true npm --prefix frontend run build -- --base="/app/${APP_NAME}/"

build_fnpack() {
    local arch=$1 platform=$2 label=$3
    local binary="${WORK_DIR}/sqlite-manager-linux-${arch}"
    local package_dir="${WORK_DIR}/package-${arch}"

    echo "🔨 编译 linux/${arch}..."
    (
        cd backend
        GOCACHE="$GOCACHE_DIR" CGO_ENABLED=0 GOOS=linux GOARCH="$arch" \
            go build -trimpath -ldflags="-s -w" -o "$binary" .
    )

    mkdir -p "$package_dir"
    cp -R fnpack/. "$package_dir/"
    rm -rf "$package_dir/app/server" "$package_dir/app/www"
    mkdir -p "$package_dir/app/server" "$package_dir/app/www"
    cp "$binary" "$package_dir/app/server/sqlite-manager"
    chmod +x "$package_dir/app/server/sqlite-manager"
    cp -R frontend/dist/. "$package_dir/app/www/"
    sed -e "s/^version.*/version               = ${VERSION}/" \
        -e "s/^platform.*/platform              = ${platform}/" \
        "$package_dir/manifest" > "$package_dir/manifest.tmp"
    mv "$package_dir/manifest.tmp" "$package_dir/manifest"
    find "$package_dir" -name '.DS_Store' -delete

    (cd "$package_dir" && fnpack build)
    # 命名对齐家族规范：包名_v版本_架构.fpk（架构标 x86，不写 amd64）
    mv "$package_dir/${APP_NAME}.fpk" "$RELEASE_DIR/${APP_NAME}_v${VERSION}_${label}.fpk"
    echo "✅ ${APP_NAME}_v${VERSION}_${label}.fpk"
}

build_fnpack amd64 x86 x86
build_fnpack arm64 arm arm64

echo "✅ 飞牛安装包已输出到 ${RELEASE_DIR}/"
ls -lh "$RELEASE_DIR"/*.fpk
