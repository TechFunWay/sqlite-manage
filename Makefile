.PHONY: all install backend frontend run dev build clean release help

all: install build

# 参数传递支持: make run ARGS="-port 3000 -data ./mydata"
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
	@echo "后端 API: http://localhost:8080"
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
	./release.sh

help:
	@echo "用法:"
	@echo "  make build              构建项目 (生产版本)"
	@echo "  make run                构建并运行 (生产模式)"
	@echo "  make run ARGS=\"-port 3000\"    指定端口运行"
	@echo "  make dev                开发模式 (前端热更新)"
	@echo "  make release            构建发布包"
	@echo "  make clean              清理构建文件"
	@echo ""
	@echo "模式说明:"
	@echo "  run  - 生产模式: 前端构建后打包，单端口访问"
	@echo "         访问: http://localhost:8080"
	@echo "  dev  - 开发模式: 前端开发服务器+后端API"
	@echo "         前端: http://localhost:5173 (修改自动刷新)"
	@echo "         后端: http://localhost:8080 (API)"
	@echo ""
	@echo "参数选项:"
	@echo "  -port string     服务端口 (默认: 8080)"
	@echo "  -data string     数据目录 (默认: ./data)"
	@echo "  -public string   静态资源目录 (默认: ./public)"
	@echo "  -upload string   上传目录 (默认: ./upload)"

clean:
	rm -rf backend/sqlite-manager
	rm -rf backend/public
	rm -rf backend/data
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf release
	cd backend && go clean
