.PHONY: all install backend frontend run dev build clean release fnpack docker help

all: install build

ARGS ?=

install:
	cd backend && go mod tidy
	cd frontend && npm install

backend:
	cd backend && go build -o sqlite-manager .

frontend:
	cd frontend && npm run build

build: backend frontend
	rm -rf backend/public
	mkdir -p backend/public/sqlite-web
	cp frontend/dist/index.html backend/public/
	cp -r frontend/dist/sqlite-web/* backend/public/sqlite-web/

run: build
	cd backend && ./sqlite-manager $(ARGS)

dev: backend
	@echo "============================================"
	@echo "  开发模式启动"
	@echo "============================================"
	@echo "后端 API: http://localhost:8903"
	@echo "前端页面: http://localhost:5173"
	@echo "============================================"
	@echo ""
	@echo "按 Ctrl+C 停止所有服务"
	@echo ""
	@trap 'kill $$BACKEND_PID $$FRONTEND_PID 2>/dev/null; exit' INT; \
	cd backend && ./sqlite-manager -no-browser $(ARGS) & BACKEND_PID=$$!; \
	sleep 1; \
	cd frontend && npm run dev & FRONTEND_PID=$$!; \
	wait

release:
	./build.sh

platforms:
	./scripts/build-all.sh

fnpack:
	./scripts/build-fnpack.sh

docker:
	./scripts/build-docker.sh

help:
	@echo "用法:"
	@echo "  make build        构建项目 (本地运行)"
	@echo "  make run          运行应用"
	@echo "  make dev          开发模式"
	@echo ""
	@echo "  make release      一键打包所有 (平台+飞牛+Docker)"
	@echo "  make platforms    只打包各平台"
	@echo "  make fnpack       只打包飞牛应用"
	@echo "  make docker       构建 Docker 多平台镜像"
	@echo ""
	@echo "  make clean        清理构建文件"

clean:
	rm -rf backend/sqlite-manager
	rm -rf backend/public
	rm -rf backend/data
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf release
	cd backend && go clean
