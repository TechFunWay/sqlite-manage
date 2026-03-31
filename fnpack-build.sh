#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

VERSION=$(grep -o 'const Version = "[^"]*"' backend/upgrade/upgrade.go | cut -d'"' -f2)
if [ -z "$VERSION" ]; then
    echo "❌ 无法获取版本号"
    exit 1
fi

APP_NAME="techfunway-sqlite-manage"
FNPACK_DIR="fnpack"
RELEASE_DIR="release/v${VERSION}"

echo "============================================"
echo "  飞牛 fnOS 应用打包工具"
echo "============================================"
echo "📦 应用名称: ${APP_NAME}"
echo "🏷️  版本号: v${VERSION}"
echo "📁 发布目录: ${RELEASE_DIR}"
echo "============================================"

cp "${FNPACK_DIR}/manifest" "${FNPACK_DIR}/manifest.bak"
sed -i '' "s/^version.*/version               = ${VERSION}/" "${FNPACK_DIR}/manifest"

echo ""
echo "📦 编译前端..."
cd frontend
npm run build
cd ..

mkdir -p "${RELEASE_DIR}"

compile_platform() {
    local GOOS=$1
    local GOARCH=$2
    local PLATFORM_LABEL=$3
    
    echo -n "  📦 ${PLATFORM_LABEL}... "
    
    local PLATFORM_DIR="${RELEASE_DIR}/${APP_NAME}-${PLATFORM_LABEL}"
    rm -rf "${PLATFORM_DIR}" 2>/dev/null || true
    mkdir -p "${PLATFORM_DIR}/www"
    
    cd backend
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} \
        go build -ldflags="-s -w" -o "../${PLATFORM_DIR}/sqlite-manager" . 2>/dev/null
    cd ..
    
    if [ "${GOOS}" = "windows" ]; then
        mv "${PLATFORM_DIR}/sqlite-manager" "${PLATFORM_DIR}/sqlite-manager.exe"
    fi
    
    cp frontend/dist/index.html "${PLATFORM_DIR}/www/"
    cp -r frontend/dist/sqlite-web "${PLATFORM_DIR}/www/"
    
    if [ "${GOOS}" = "windows" ]; then
        cat > "${PLATFORM_DIR}/start.bat" << 'BATEOF'
@echo off
cd /d "%~dp0"
sqlite-manager.exe %*
BATEOF
    else
        cat > "${PLATFORM_DIR}/start.sh" << 'SHEOF'
#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"
chmod +x sqlite-manager 2>/dev/null || true
./sqlite-manager "$@"
SHEOF
        chmod +x "${PLATFORM_DIR}/start.sh"
    fi
    
    echo "✅"
}

build_fnpack() {
    local GOARCH=$1
    local FNPLATFORM=$2
    local ARCH_LABEL=$3
    
    echo -n "  📦 飞牛 ${ARCH_LABEL}.fpk... "
    
    local BUILD_DIR="${RELEASE_DIR}/${APP_NAME}-fnos-${ARCH_LABEL}"
    rm -rf "${BUILD_DIR}" 2>/dev/null || true
    mkdir -p "${BUILD_DIR}"
    
    cp -r "${FNPACK_DIR}/"* "${BUILD_DIR}/"
    sed -i '' "s/^platform.*/platform              = ${FNPLATFORM}/" "${BUILD_DIR}/manifest"
    
    mkdir -p "${BUILD_DIR}/app/server"
    mkdir -p "${BUILD_DIR}/app/www"
    
    cp "${RELEASE_DIR}/${APP_NAME}-linux-${ARCH_LABEL}/sqlite-manager" "${BUILD_DIR}/app/server/"
    cp -r "${RELEASE_DIR}/${APP_NAME}-linux-${ARCH_LABEL}/www/"* "${BUILD_DIR}/app/www/"
    
    cd "${BUILD_DIR}"
    fnpack build > /dev/null 2>&1
    cd "${SCRIPT_DIR}"
    
    mv "${BUILD_DIR}/${APP_NAME}.fpk" "${RELEASE_DIR}/${APP_NAME}-v${VERSION}-${ARCH_LABEL}.fpk"
    rm -rf "${BUILD_DIR}"
    
    echo "✅"
}

echo ""
echo "🔨 编译所有平台..."
echo ""

compile_platform "linux"   "amd64" "linux-amd64"
compile_platform "linux"   "arm64" "linux-arm64"
compile_platform "darwin"  "amd64" "macos-amd64"
compile_platform "darwin"  "arm64" "macos-arm64"
compile_platform "windows" "amd64" "windows-amd64"

echo ""
echo "📦 打包飞牛应用..."
build_fnpack "amd64" "x86" "amd64"
build_fnpack "arm64" "arm" "arm64"

mv "${FNPACK_DIR}/manifest.bak" "${FNPACK_DIR}/manifest"

echo ""
echo "📦 创建压缩包..."
cd "${RELEASE_DIR}"

for dir in ${APP_NAME}-linux-* ${APP_NAME}-macos-* ${APP_NAME}-windows-*; do
    if [ -d "$dir" ]; then
        tar -czf "${dir}.tar.gz" "$dir/"
    fi
done

tar -czf "${APP_NAME}-v${VERSION}-all.tar.gz" ${APP_NAME}-linux-* ${APP_NAME}-macos-* ${APP_NAME}-windows-*

cd "${SCRIPT_DIR}"

echo ""
echo "============================================"
echo "  ✅ 打包完成!"
echo "============================================"
echo ""
echo "📁 平台包:"
ls -lh "${RELEASE_DIR}"/*.tar.gz 2>/dev/null | grep -v "all" | awk '{print "  " $NF " (" $5 ")"}'
echo ""
echo "📦 飞牛应用:"
ls -lh "${RELEASE_DIR}"/*.fpk 2>/dev/null | awk '{print "  " $NF " (" $5 ")"}'
echo ""
echo "📦 全平台包:"
ls -lh "${RELEASE_DIR}"/*all.tar.gz 2>/dev/null | awk '{print "  " $NF " (" $5 ")"}'
echo ""
echo "============================================"
