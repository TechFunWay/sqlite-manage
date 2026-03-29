#!/bin/bash

# ============================================================
# SQLite Manager 发布构建脚本
# ============================================================

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

VERSION=$(grep -o 'const Version = "[^"]*"' backend/upgrade/upgrade.go | cut -d'"' -f2)
if [ -z "$VERSION" ]; then
    echo "❌ 无法获取版本号"
    exit 1
fi

PROJECT_NAME="techfunway-sqlite-manage"
RELEASE_NAME="${PROJECT_NAME}-v${VERSION}"
RELEASE_DIR="release/v${VERSION}"
BUILD_DIR="${RELEASE_DIR}/${RELEASE_NAME}"

CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH=$(uname -m)
if [ "$CURRENT_ARCH" = "x86_64" ]; then
    CURRENT_ARCH="x64"
elif [ "$CURRENT_ARCH" = "aarch64" ] || [ "$CURRENT_ARCH" = "arm64" ]; then
    CURRENT_ARCH="arm64"
fi

echo "============================================"
echo "  SQLite Manager 发布构建工具"
echo "============================================"
echo "📦 项目: ${PROJECT_NAME}"
echo "🏷️  版本: v${VERSION}"
echo "📁 目录: ${BUILD_DIR}"
echo "============================================"

rm -rf "${RELEASE_DIR:?}" 2>/dev/null || true
mkdir -p "${BUILD_DIR}"

echo ""
echo "📦 编译前端..."
cd frontend
npm run build
cd ..

create_platform_files() {
    local DIR=$1
    local GOOS=$2
    
    mkdir -p "${DIR}/public/sqlite-web"
    cp frontend/dist/index.html "${DIR}/public/"
    cp -r frontend/dist/sqlite-web/* "${DIR}/public/sqlite-web/"
    
    if [ "$GOOS" = "windows" ]; then
        cat > "${DIR}/start.bat" << 'EOF'
@echo off
cd /d "%~dp0"
set SQLITE_DATA_DIR=%~dp0data
set SQLITE_UPLOAD_DIR=%~dp0upload
sqlite-manager.exe %*
EOF
    else
        cat > "${DIR}/start.sh" << 'EOF'
#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"
chmod +x sqlite-manager 2>/dev/null || true

# 设置默认目录
export SQLITE_DATA_DIR="${SQLITE_DATA_DIR:-$SCRIPT_DIR/data}"
export SQLITE_UPLOAD_DIR="${SQLITE_UPLOAD_DIR:-$SCRIPT_DIR/upload}"

./sqlite-manager "$@"
EOF
        chmod +x "${DIR}/start.sh"
    fi
    
    cat > "${DIR}/README.md" << EOF
# SQLite Manager v${VERSION}

## 启动方式
\`\`\`bash
# 使用默认配置
./start.sh

# 指定端口
./start.sh -port 3000

# 指定数据目录
./start.sh -data /path/to/data

# 查看所有参数
./start.sh help
\`\`\`

## 命令行参数
| 参数 | 说明 | 默认值 |
|------|------|--------|
| -port | 服务端口 | 8080 |
| -data | 数据目录 | ./data |
| -public | 静态资源目录 | ./public |
| -upload | 上传目录 | ./upload |

## 环境变量
| 变量 | 说明 |
|------|------|
| PORT | 服务端口 |
| SQLITE_DATA_DIR | 数据目录 |
| SQLITE_PUBLIC_DIR | 静态资源目录 |
| SQLITE_UPLOAD_DIR | 上传目录 |

## 子命令
\`\`\`bash
./sqlite-manager version          # 查看版本
./sqlite-manager reset-password   # 重置密码
./sqlite-manager upgrade-status   # 升级状态
./sqlite-manager help             # 帮助信息
\`\`\`

构建时间: $(date "+%Y-%m-%d %H:%M:%S")
EOF
}

compile_platform() {
    local GOOS=$1
    local GOARCH=$2
    local ARCH_LABEL=$3
    local USE_CGO=${4:-0}
    
    local PLATFORM_DIR="${BUILD_DIR}/${ARCH_LABEL}"
    create_platform_files "${PLATFORM_DIR}" "${GOOS}"
    
    local BINARY="sqlite-manager"
    [ "$GOOS" = "windows" ] && BINARY="sqlite-manager.exe"
    
    echo -n "  📦 ${ARCH_LABEL}... "
    cd backend
    CGO_ENABLED=${USE_CGO} GOOS="${GOOS}" GOARCH="${GOARCH}" \
        go build -ldflags="-s -w" -o "../${PLATFORM_DIR}/${BINARY}" . 2>/dev/null
    cd ..
    echo "✅"
}

echo ""
echo "🔨 编译各平台版本..."
echo ""

compile_platform "${CURRENT_OS}" "${CURRENT_ARCH}" "${CURRENT_OS}-${CURRENT_ARCH}" "1"
compile_platform "darwin"  "amd64" "mac-x64"
compile_platform "darwin"  "arm64" "mac-arm64"
compile_platform "linux"   "amd64" "linux-x64"
compile_platform "linux"   "arm64" "linux-arm64"
compile_platform "windows" "amd64" "win-x64"

echo ""
echo "📄 创建测试数据库..."
sqlite3 "${BUILD_DIR}/${CURRENT_OS}-${CURRENT_ARCH}/test.db" \
    "CREATE TABLE demo (id INTEGER PRIMARY KEY, name TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP); 
     INSERT INTO demo (name) VALUES ('测试数据1'), ('测试数据2');"

echo ""
echo "📦 创建压缩包..."
cd "${RELEASE_DIR}"

for DIR in "${RELEASE_NAME}"/*/; do
    if [ -d "$DIR" ]; then
        PLATFORM=$(basename "$DIR")
        ARCHIVE="${RELEASE_NAME}-${PLATFORM}"
        
        if [[ "$PLATFORM" == *"windows"* ]] || [[ "$PLATFORM" == *"win-"* ]]; then
            echo "  📦 ${PLATFORM} (.zip)..."
            zip -r "${ARCHIVE}.zip" "${RELEASE_NAME}/${PLATFORM}/" > /dev/null 2>&1
        else
            echo "  📦 ${PLATFORM} (.tar.gz)..."
            tar -czf "${ARCHIVE}.tar.gz" "${RELEASE_NAME}/${PLATFORM}/"
        fi
    fi
done

cd - > /dev/null

echo ""
echo "============================================"
echo "  ✅ 构建完成!"
echo "============================================"
echo ""
echo "📁 发布目录: ${BUILD_DIR}"
echo ""
echo "📦 压缩包:"
ls -lh "${RELEASE_DIR}"/*.{tar.gz,zip} 2>/dev/null | awk '{print "  " $NF " (" $5 ")"}'
echo ""
echo "🚀 测试当前平台:"
echo "   cd ${BUILD_DIR}/${CURRENT_OS}-${CURRENT_ARCH} && ./start.sh"
echo ""
echo "============================================"
