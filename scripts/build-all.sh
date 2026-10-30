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
    cp frontend/dist/manifest.webmanifest "${DIR}/www/"
    cp -r frontend/dist/sqlite-web "${DIR}/www/"
    cp -r frontend/dist/icons "${DIR}/www/"
    
    find "${DIR}" -name ".DS_Store" -delete 2>/dev/null || true
    
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
    # 清掉旧格式的包，Windows 用 zip，其余用 tar.gz
    rm -f "${name}.tar.gz" "${name}.zip"
    if [[ "${name}" == *-windows-* ]]; then
        zip -qr "${name}.zip" "$dir" -x '*.DS_Store'
    else
        COPYFILE_DISABLE=1 tar -czf "${name}.tar.gz" --exclude='.DS_Store' "$dir"
    fi
    # 发行目录只留最终产物，未压缩的中间目录随手清掉
    rm -rf "$dir"
done
cd "$PROJECT_DIR"

# 更新日志、截图、compose 文件随发行目录分发，版本目录自包含
[ -f CHANGELOG.md ] && cp CHANGELOG.md "${RELEASE_DIR}/CHANGELOG.md"

if [ -d docs/screenshots ]; then
    rm -rf "${RELEASE_DIR}/screenshots"
    mkdir -p "${RELEASE_DIR}/screenshots"
    cp docs/screenshots/*.png "${RELEASE_DIR}/screenshots/"
    find "${RELEASE_DIR}/screenshots" -name '.DS_Store' -delete 2>/dev/null || true
    (cd "${RELEASE_DIR}" && zip -qr "screenshots-v${VERSION}.zip" screenshots)
    echo "📸 已打包截图 screenshots-v${VERSION}.zip"
fi

# 版本文件中的镜像 tag 钉死为当前版本，latest 文件原样分发；
# 拷贝到目标主机后 docker compose up -d 即可运行
sed "s|__VERSION__|v${VERSION}|" docker-compose.version.yaml > "${RELEASE_DIR}/docker-compose.yml"
cp docker-compose.yaml "${RELEASE_DIR}/docker-compose.latest.yml"

echo ""
echo "✅ 完成!"
ls -lh "${RELEASE_DIR}"/ | awk 'NR>1 {print "  " $NF " (" $5 ")"}'
