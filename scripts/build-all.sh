#!/bin/bash

set -e

PROJECT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$PROJECT_DIR"

VERSION=$(grep -o 'const Version = "[^"]*"' backend/upgrade/upgrade.go | cut -d'"' -f2)
[ -z "$VERSION" ] && echo "❌ 无法获取版本号" && exit 1

APP_NAME="techfunway-sqlite-manage"
RELEASE_DIR="release/v${VERSION}"

echo "============================================"
echo "  多平台打包 v${VERSION}"
echo "============================================"

mkdir -p "${RELEASE_DIR}"

echo ""
echo "📦 编译前端..."
cd frontend && npm run build && cd ..

compile() {
    GOOS=$1
    GOARCH=$2
    LABEL=$3
    echo -n "  📦 ${LABEL}... "
    
    DIR="${RELEASE_DIR}/${APP_NAME}-${VERSION}-${LABEL}"
    rm -rf "${DIR}"
    mkdir -p "${DIR}/www"
    
    cd backend
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags="-s -w" -o "../${DIR}/sqlite-manager" . 2>/dev/null
    cd ..
    
    if [ "${GOOS}" = "windows" ]; then
        mv "${DIR}/sqlite-manager" "${DIR}/sqlite-manager.exe"
        echo '@echo off
cd /d "%~dp0"
sqlite-manager.exe %*' > "${DIR}/start.bat"
    else
        echo '#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"
chmod +x sqlite-manager 2>/dev/null || true
./sqlite-manager "$@"' > "${DIR}/start.sh"
        chmod +x "${DIR}/start.sh"
    fi
    
    cp frontend/dist/index.html "${DIR}/www/"
    cp -r frontend/dist/sqlite-web "${DIR}/www/"
    
    echo "✅"
}

echo ""
echo "🔨 编译平台..."
echo ""
compile "linux" "amd64" "linux-amd64"
compile "linux" "arm64" "linux-arm64"
compile "darwin" "amd64" "macos-amd64"
compile "darwin" "arm64" "macos-arm64"
compile "windows" "amd64" "windows-amd64"

echo ""
echo "📦 创建压缩包..."
cd "${RELEASE_DIR}"
for dir in ${APP_NAME}-${VERSION}-*/; do
    name=$(basename "$dir")
    tar -czf "${name}.tar.gz" "$dir"
done
cd "$PROJECT_DIR"

echo ""
echo "✅ 完成!"
ls -lh "${RELEASE_DIR}"/*.tar.gz | awk '{print "  " $NF " (" $5 ")"}'
