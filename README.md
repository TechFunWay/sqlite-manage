# SQLite Manager

一个现代化的 SQLite 数据库 Web 管理工具，类似于 DataGrip 和 Navicat，但更加简洁直观。

## 功能特性

- 📂 **多数据库管理**: 同时打开和管理多个数据库
- 🗂️ **数据库切换**: 在多个数据库之间快速切换
- 📊 **数据库概览**: 查看数据库文件大小、表数量、总行数等详细信息
- 📋 **表管理**: 查看所有表、支持搜索、创建和删除表
- 🔍 **表结构**: 查看和修改表结构（添加/删除字段、索引管理）
- ✏️ **数据编辑**: 分页查看、双击编辑、新增/删除行
- 📤 **数据导出**: 支持导出为 CSV 格式
- 🌙 **深色主题**: 专业级的深色界面设计
- 📱 **响应式**: 完美支持桌面和移动设备

## 技术栈

- **前端**: Vue 3 + Vite + Tailwind CSS + Pinia
- **后端**: Go + Gin
- **数据库**: SQLite3

## 快速开始

### 方式一：使用 Make (推荐)

```bash
# 安装依赖并启动
make dev

# 或者分别执行
make install  # 安装所有依赖
make run      # 启动应用
```

### 方式二：手动启动

**1. 安装 Go 依赖**
```bash
cd backend
go mod tidy
go build -o sqlite-manager .
```

**2. 安装前端依赖**
```bash
cd frontend
npm install
```

**3. 启动后端**
```bash
cd backend
./sqlite-manager
```
后端服务将在 http://localhost:8903 启动，并自动打开浏览器。

**4. 启动前端开发服务器 (可选)**
```bash
cd frontend
npm run dev
```
前端开发服务器将在 http://localhost:5173 启动，并代理 API 请求到后端。

### 方式三：构建生产版本

```bash
# 构建后端
cd backend
go build -o sqlite-manager .

# 构建前端
cd frontend
npm run build

# 将 dist 目录中的前端文件复制到后端的 public 目录
cp -r dist ../backend/public
```

## 使用说明

### 多数据库管理

1. 可以同时打开多个数据库文件
2. 在顶栏点击数据库名称切换当前活跃数据库
3. 点击 X 按钮关闭不需要的数据库
4. 活跃数据库会在列表中标记为"活跃"

### 打开数据库

1. **方式一**: 在首页输入数据库文件路径，点击"打开"
2. **方式二**: 拖拽数据库文件到上传区域
3. **方式三**: 点击上传区域选择文件

### 管理表

- 点击左侧边栏的表名查看表数据和结构
- 使用 Tab 切换查看"数据"、"结构"、"信息"
- 在结构视图中可以添加/删除字段和索引

### 编辑数据

- 双击单元格进行编辑
- 使用"新增"按钮添加新行
- 点击行尾的删除按钮删除行
- 编辑完成后自动保存

### 导出数据

在数据视图中点击"导出"按钮可将当前页数据导出为 CSV 格式。

## 项目结构

```
sqlite-manage/
├── backend/                 # Go 后端
│   ├── main.go             # 主程序入口
│   ├── handlers/           # HTTP 处理器
│   │   ├── database.go     # 数据库相关 API
│   │   ├── tables.go       # 表管理 API
│   │   └── data.go         # 数据操作 API
│   ├── database/           # 数据库操作层
│   │   └── sqlite.go       # SQLite 封装
│   ├── models/             # 数据模型
│   │   └── models.go
│   └── public/             # 前端静态文件
│
├── frontend/               # Vue 前端
│   ├── src/
│   │   ├── components/     # Vue 组件
│   │   ├── views/          # 页面视图
│   │   ├── stores/         # Pinia 状态管理
│   │   ├── api/            # API 请求封装
│   │   └── router/         # 路由配置
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
│
├── SPEC.md                 # 项目规范文档
└── README.md               # 项目说明文档
```

## API 接口

### 数据库操作
- `POST /api/database/open` - 打开数据库
- `POST /api/database/create` - 创建数据库
- `POST /api/database/upload` - 上传数据库
- `GET /api/databases` - 获取所有已打开的数据库
- `PUT /api/databases/:id/activate` - 切换活跃数据库
- `DELETE /api/databases/:id` - 关闭数据库
- `GET /api/database/info` - 获取当前活跃数据库信息

### 表操作
- `GET /api/tables` - 获取所有表
- `GET /api/tables/:name/schema` - 获取表结构
- `POST /api/tables` - 创建表
- `DELETE /api/tables/:name` - 删除表

### 数据操作
- `GET /api/tables/:name/data` - 获取表数据（支持分页）
- `POST /api/tables/:name/data` - 新增数据
- `PUT /api/tables/:name/data` - 更新数据
- `DELETE /api/tables/:name/data` - 删除数据

## 开发说明

### 前端开发
```bash
cd frontend
npm install
npm run dev
```

### 后端开发
```bash
cd backend
go mod tidy
go run main.go
```

## 云端部署 (CloudBase)

### 腾讯云 CloudBase 部署

**环境信息:**
- 环境 ID: `code-buddy-wei-1gpibyda029a37b5`
- 区域: `ap-shanghai`
- 版本: v1.1.0

**服务地址:**

| 服务 | 类型 | 地址 |
|------|------|------|
| 前端 | 静态托管 | https://code-buddy-wei-1gpibyda029a37b5-1251585624.tcloudbaseapp.com/ |
| 后端 API | 云托管 Cloud Run | https://sqlite-manager-backend-244964-5-1251585624.sh.run.tcloudbase.com |

**部署架构:**
```
用户浏览器
    ↓
静态托管 (前端 Vue)
    ↓ API 请求
云托管 Cloud Run (后端 Go/Gin)
    ↓
SQLite 数据库 (本地文件)
```

**后端配置:**
- 服务名称: `sqlite-manager-backend`
- 端口: `8903`
- 资源规格: 0.5 CPU / 1GB 内存
- 最小实例数: 1
- 配置文件: `backend/Dockerfile`

**控制台管理:**
- [云托管控制台](https://tcb.cloud.tencent.com/dev#/platform-run/service/detail?id=sqlite-manager-backend&NameSpace=code-buddy-wei-1gpibyda029a37b5)
- [静态托管控制台](https://tcb.cloud.tencent.com/dev#/static-hosting)

## License

MIT License
