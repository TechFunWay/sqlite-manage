#!/bin/bash

# ============================================================
# 飞牛 fnOS 应用打包脚本
# ============================================================

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 从 Go 代码获取版本号
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

# 备份原始 manifest
cp "${FNPACK_DIR}/manifest" "${FNPACK_DIR}/manifest.bak"

# 更新 manifest 版本号
echo ""
echo "📝 更新 manifest 版本号..."
sed -i '' "s/^version.*/version               = ${VERSION}/" "${FNPACK_DIR}/manifest"

# 编译前端
echo ""
echo "📦 编译前端..."
cd frontend
npm run build
cd ..

# 编译函数
compile_and_pack() {
    local GOARCH=$1
    local FNPLATFORM=$2
    local ARCH_LABEL=$3
    
    echo ""
    echo "🔨 编译 ${ARCH_LABEL} 版本 (GOARCH=${GOARCH}, platform=${FNPLATFORM})..."
    
    # 创建临时打包目录
    local BUILD_DIR="${RELEASE_DIR}/${APP_NAME}-${ARCH_LABEL}"
    rm -rf "${BUILD_DIR}" 2>/dev/null || true
    mkdir -p "${BUILD_DIR}"
    
    # 复制飞牛模板文件
    cp -r "${FNPACK_DIR}/"* "${BUILD_DIR}/"
    
    # 更新 manifest 的 platform
    sed -i '' "s/^platform.*/platform              = ${FNPLATFORM}/" "${BUILD_DIR}/manifest"
    
    # 创建 app 目录结构
    mkdir -p "${BUILD_DIR}/app/server"
    mkdir -p "${BUILD_DIR}/app/www"
    
    # 编译 Go 程序
    cd backend
    CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} \
        go build -ldflags="-s -w" -o "../${BUILD_DIR}/app/server/sqlite-manager" . 2>/dev/null
    cd ..
    
    # 复制前端静态资源
    cp frontend/dist/index.html "${BUILD_DIR}/app/www/"
    cp -r frontend/dist/sqlite-web "${BUILD_DIR}/app/www/"
    
    echo "✅ ${ARCH_LABEL} 编译完成"
    
    # 使用 fnpack build 打包
    echo "📦 使用 fnpack build 打包 ${ARCH_LABEL} fpk..."
    cd "${BUILD_DIR}"
    fnpack build
    cd "${SCRIPT_DIR}"
    
    # 移动 fpk 文件到 release 目录
    local FPK_NAME="${APP_NAME}-v${VERSION}-${ARCH_LABEL}.fpk"
    mv "${BUILD_DIR}/${APP_NAME}.fpk" "${RELEASE_DIR}/${FPK_NAME}" 2>/dev/null || true
    
    # 清理临时目录
    rm -rf "${BUILD_DIR}"
    
    echo "✅ ${FPK_NAME} 打包完成"
}

# 创建 release 目录
mkdir -p "${RELEASE_DIR}"

# 编译并打包 arm 版本
compile_and_pack "arm64" "arm" "arm"

# 编译并打包 x86 版本
compile_and_pack "amd64" "x86" "x86"

# 恢复原始 manifest
mv "${FNPACK_DIR}/manifest.bak" "${FNPACK_DIR}/manifest"

# 显示结果
echo ""
echo "============================================"
echo "  ✅ 飞牛应用打包完成!"
echo "============================================"
echo ""
echo "📦 打包文件:"
ls -lh "${RELEASE_DIR}"/*.fpk 2>/dev/null | awk '{print "  " $NF " (" $5 ")"}'
echo ""
echo "============================================"
