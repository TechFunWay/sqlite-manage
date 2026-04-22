# 运行时镜像 - 包含预编译的 Go 二进制
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir -p /app/data /app/public /app/upload

WORKDIR /app

# 二进制文件将在构建时复制
COPY sqlite-manager /app/sqlite-manager

ENV PORT=8903
ENV SQLITE_DATA_DIR=/app/data
ENV SQLITE_WEB_DIR=/app/public
ENV SQLITE_UPLOAD_DIR=/app/upload

EXPOSE 8903

CMD ["./sqlite-manager", "-device-type", "docker"]
