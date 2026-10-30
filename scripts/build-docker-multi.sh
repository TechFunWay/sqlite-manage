#!/bin/bash
set -euo pipefail

# 多平台（linux/amd64 + linux/arm64）合并镜像，一个 tag 同时携带两种架构。
#
# 默认只导出本地 OCI 归档（release/v<版本>/techfunway-sqlite-manage-v<版本>-multiarch.oci.tar），
# 随发行目录分发，不碰 registry。PUSH=1（或 --push）时额外把同一份合并 manifest 推到
# Docker Hub（techfunways/sqlite-manage 的 v<版本> 与 latest 两个 tag），需已 docker login。
#
# 与账单 / 记事本的做法一致：用 build-all.sh 已交叉编译好的静态二进制 + FROM scratch 组镜像，
# 避免 buildx 在 qemu 下重跑 npm 与 go 编译。scratch 里没有的三样东西必须随镜像带上，
# 否则功能会坏：
#   - /etc/ssl/cert.pem：匿名使用统计走 https，缺 CA 证书整个上报失败；
#   - /usr/share/zoneinfo：应用按 TZ=Asia/Shanghai 解析本地时区（时间与统计按日归集），
#     缺了会退化成 UTC；
#   - /tmp：上传数据库与导入数据的 multipart 解析要在临时目录落盘，scratch 不建 /tmp，
#     缺目录会直接报「权限拒绝 / 没有该文件或目录」。

PUSH="${PUSH:-0}"
for arg in "$@"; do
  [ "$arg" = "--push" ] && PUSH=1
done

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

APP_NAME="techfunway-sqlite-manage"
BINARY_NAME="sqlite-manager"
IMAGE_NAME="techfunways/sqlite-manage"
BUILDER_NAME="${BUILDER_NAME:-techfunway-sqlite-manage-multiarch}"
VERSION=$(grep -o 'const Version = "[^"]*"' backend/upgrade/upgrade.go | cut -d'"' -f2)
[ -z "$VERSION" ] && echo "❌ 无法获取版本号" && exit 1
RELEASE_DIR="${ROOT_DIR}/release/v${VERSION}"
OCI_FILE="${RELEASE_DIR}/${APP_NAME}-v${VERSION}-multiarch.oci.tar"
SERVICE_PORT="8903"

if ! command -v docker >/dev/null 2>&1; then
  echo "Error: 构建多平台镜像需要 Docker。"
  exit 1
fi

if [ ! -d "$RELEASE_DIR" ]; then
  echo "缺少 ${RELEASE_DIR}，请先运行 bash scripts/build-all.sh"
  exit 1
fi

# build-all.sh 压缩后即清理中间平台目录；多平台构建需要 linux 平台目录，
# 缺失时从对应的 tar.gz 解包，用完在脚本末尾清理。
DOCKER_TMP_DIRS=()
for arch in amd64 arm64; do
  DIR_NAME="${APP_NAME}-${VERSION}-linux-${arch}"
  if [ ! -d "${RELEASE_DIR}/${DIR_NAME}" ]; then
    TAR_FILE="${RELEASE_DIR}/${DIR_NAME}.tar.gz"
    if [ ! -f "$TAR_FILE" ]; then
      echo "缺少 ${TAR_FILE}，请先运行 bash scripts/build-all.sh"
      exit 1
    fi
    tar -xzf "$TAR_FILE" -C "${RELEASE_DIR}"
    DOCKER_TMP_DIRS+=("${RELEASE_DIR}/${DIR_NAME}")
  fi
done

find "${RELEASE_DIR}" -name '.DS_Store' -delete 2>/dev/null || true
rm -f "${OCI_FILE}"

# 合并 manifest 的 OCI 归档必须用 docker-container 驱动的 builder，
# 默认驱动不推 registry 就导不出多平台合并结果。
if ! docker buildx inspect "${BUILDER_NAME}" >/dev/null 2>&1; then
  echo "Creating buildx builder: ${BUILDER_NAME}"
  docker buildx create --name "${BUILDER_NAME}" --driver docker-container --bootstrap >/dev/null
fi

CONTEXT_DIR="${ROOT_DIR}/tmp/docker-multi-context"
rm -rf "${CONTEXT_DIR}"
mkdir -p "${CONTEXT_DIR}/payload"
cleanup() {
  rm -rf "${CONTEXT_DIR}"
  for d in "${DOCKER_TMP_DIRS[@]:-}"; do
    [ -n "$d" ] && rm -rf "$d"
  done
}
trap cleanup EXIT

# 组装镜像内容：二进制与前端按 TARGETARCH 选，其余是平台无关的补充文件。
for arch in amd64 arm64; do
  SRC="${RELEASE_DIR}/${APP_NAME}-${VERSION}-linux-${arch}"
  DST="${CONTEXT_DIR}/payload/${arch}"
  mkdir -p "${DST}/www"
  cp "${SRC}/${BINARY_NAME}" "${DST}/${BINARY_NAME}"
  chmod 755 "${DST}/${BINARY_NAME}"
  cp -r "${SRC}/www/." "${DST}/www/"
done

# CA 证书：本机（macOS）的系统信任库。
CA_SRC="/etc/ssl/cert.pem"
if [ ! -f "$CA_SRC" ]; then
  echo "本机缺少 ${CA_SRC}，无法构建带 TLS 能力的镜像。"
  exit 1
fi
mkdir -p "${CONTEXT_DIR}/payload/etc/ssl"
cp "$CA_SRC" "${CONTEXT_DIR}/payload/etc/ssl/cert.pem"

# 时区数据：Go 按 TZ 名字到 /usr/share/zoneinfo 下找同名文件。
ZONEINFO_SRC="/usr/share/zoneinfo"
for z in Asia/Shanghai UTC GMT; do
  if [ -f "${ZONEINFO_SRC}/${z}" ]; then
    mkdir -p "$(dirname "${CONTEXT_DIR}/payload/usr/share/zoneinfo/${z}")"
    cp "${ZONEINFO_SRC}/${z}" "${CONTEXT_DIR}/payload/usr/share/zoneinfo/${z}"
  else
    echo "警告：本机缺少 ${ZONEINFO_SRC}/${z}，该时区会退化为 UTC。"
  fi
done

# 可写的临时目录与数据目录（scratch 里什么都没有，两个都要显式准备）。
mkdir -p "${CONTEXT_DIR}/payload/tmp" "${CONTEXT_DIR}/payload/data"
touch "${CONTEXT_DIR}/payload/tmp/.keep" "${CONTEXT_DIR}/payload/data/.keep"
chmod 1777 "${CONTEXT_DIR}/payload/tmp"

cat > "${CONTEXT_DIR}/Dockerfile" <<EOF
FROM scratch

ARG TARGETARCH
ARG VERSION

LABEL org.opencontainers.image.title="${APP_NAME}"
LABEL org.opencontainers.image.description="简洁高效的 SQLite 数据库 Web 管理工具"
LABEL org.opencontainers.image.version="\${VERSION}"
LABEL org.opencontainers.image.source="https://github.com/TechFunWay/${APP_NAME#techfunway-}"

WORKDIR /app

COPY payload/\${TARGETARCH}/${BINARY_NAME} /app/${BINARY_NAME}
COPY payload/\${TARGETARCH}/www /app/www
COPY payload/etc/ssl/cert.pem /etc/ssl/cert.pem
COPY payload/usr/share/zoneinfo /usr/share/zoneinfo
COPY --chmod=1777 payload/tmp /tmp
COPY --chown=1000:1000 payload/data /data

EXPOSE ${SERVICE_PORT}
VOLUME ["/data"]

ENV PORT=${SERVICE_PORT}
ENV SQLITE_DATA_DIR=/data
ENV SQLITE_UPLOAD_DIR=/data/upload
ENV SQLITE_WEB_DIR=/app/www
ENV TZ=Asia/Shanghai

USER 1000

ENTRYPOINT ["/app/${BINARY_NAME}"]
CMD ["-port", "${SERVICE_PORT}", "-data-dir", "/data", "-upload-dir", "/data/upload", "-web-dir", "/app/www", "-no-browser", "-device-type", "docker"]
EOF

echo "构建 ${IMAGE_NAME}:v${VERSION} 的 linux/amd64,linux/arm64 两种架构..."
echo "导出本地 OCI 归档（不推仓库）：${OCI_FILE}"

docker buildx build \
  --builder "${BUILDER_NAME}" \
  --platform linux/amd64,linux/arm64 \
  --build-arg "VERSION=${VERSION}" \
  -t "${IMAGE_NAME}:v${VERSION}" \
  -t "${IMAGE_NAME}:latest" \
  --output "type=oci,dest=${OCI_FILE}" \
  "${CONTEXT_DIR}"

echo "多平台 OCI 归档完成：${OCI_FILE}"
echo "目标机器载入：docker load -i ${APP_NAME}-v${VERSION}-multiarch.oci.tar"

# PUSH=1：把同一份多平台合并 manifest 推到 Docker Hub（需已 docker login）
if [ "${PUSH}" = "1" ]; then
  echo ""
  echo "推送多平台镜像到 Docker Hub：${IMAGE_NAME}:v${VERSION} / :latest"
  docker buildx build \
    --builder "${BUILDER_NAME}" \
    --platform linux/amd64,linux/arm64 \
    --push \
    --build-arg "VERSION=${VERSION}" \
    -t "${IMAGE_NAME}:v${VERSION}" \
    -t "${IMAGE_NAME}:latest" \
    "${CONTEXT_DIR}"
  echo "已推送 ${IMAGE_NAME}:v${VERSION} 与 ${IMAGE_NAME}:latest（amd64+arm64 合并 manifest）"
fi
