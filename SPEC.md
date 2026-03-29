# SQLite Web Manager - 规范文档

## 1. 概念与愿景

一款现代化的SQLite数据库Web管理工具，灵感来源于DataGrip和Navicat的专业设计，但更注重简洁与直观。采用深色主题设计，配合流畅的动画和优雅的布局，为开发者提供专业级的数据库管理体验。支持响应式设计，在桌面和移动设备上都能流畅使用。

## 2. 设计语言

### 美学方向
- **风格**：现代深色主题，借鉴DataGrip的专业感与VS Code的简洁风格
- **质感**：玻璃拟态效果，微妙的阴影和模糊背景
- **氛围**：专业、冷静、高效

### 色彩系统
```
主色调 (Primary): #6366F1 (Indigo-500)
次要色 (Secondary): #8B5CF6 (Violet-500)
强调色 (Accent): #22D3EE (Cyan-400)
成功色 (Success): #10B981 (Emerald-500)
警告色 (Warning): #F59E0B (Amber-500)
错误色 (Error): #EF4444 (Red-500)

背景色:
- 深色背景: #0F172A (Slate-900)
- 卡片背景: #1E293B (Slate-800)
- 悬浮背景: #334155 (Slate-700)

文字色:
- 主文字: #F8FAFC (Slate-50)
- 次要文字: #94A3B8 (Slate-400)
- 禁用文字: #64748B (Slate-500)
```

### 字体系统
- **主字体**: Inter (Google Fonts) - 现代、清晰、专业
- **代码字体**: JetBrains Mono - 适合SQL和代码展示
- **字体大小**:
  - 标题: 24px / 20px / 16px
  - 正文: 14px
  - 小字: 12px
  - 代码: 13px

### 空间系统
- 基础单位: 4px
- 间距: 8px, 12px, 16px, 24px, 32px
- 圆角: 6px (小), 8px (中), 12px (大)
- 阴影: 0 4px 6px -1px rgba(0, 0, 0, 0.3)

### 动画哲学
- **过渡时长**: 200ms (快), 300ms (标准), 500ms (慢)
- **缓动函数**: cubic-bezier(0.4, 0, 0.2, 1)
- **动画场景**:
  - 页面切换: 淡入淡出 + 轻微上移
  - 侧边栏展开: 平滑滑动
  - 表格行悬浮: 背景色渐变
  - 按钮点击: 轻微缩放 + 颜色变化
  - 加载状态: 优雅的骨架屏或旋转动画

### 视觉资源
- **图标库**: Lucide Icons (现代化、一致性强)
- **空状态插图**: 简约的SVG插图
- **表格**: 使用虚拟滚动优化大数据量渲染

## 3. 布局与结构

### 整体架构
```
┌─────────────────────────────────────────────────────────────┐
│  Header (Logo + 数据库选择器 + 连接状态)                      │
├────────────────┬────────────────────────────────────────────┤
│                │                                            │
│   Sidebar      │         Main Content Area                  │
│   (数据库树)    │         (表数据/结构/查询结果)              │
│                │                                            │
│   - 数据库信息  │                                            │
│   - 表列表     │                                            │
│   - 视图       │                                            │
│                │                                            │
├────────────────┴────────────────────────────────────────────┤
│  Status Bar (连接信息 + 执行时间 + 行数统计)                  │
└─────────────────────────────────────────────────────────────┘
```

### 响应式断点
- **桌面**: ≥1024px - 完整三栏布局
- **平板**: 768px-1023px - 可折叠侧边栏
- **手机**: <768px - 单栏布局，底部导航

### 页面结构
1. **首页/连接页**: 数据库选择和上传
2. **数据库视图**: 侧边栏 + 主内容区
3. **表详情视图**: Tab切换 (结构/数据/SQL)

## 4. 功能与交互

### 4.1 数据库连接
**打开本地数据库**
- 点击"打开数据库"按钮 → 系统文件选择器 → 选择.db/.sqlite/.sqlite3文件
- 支持拖拽文件到上传区域
- 最近打开的数据库列表（LocalStorage存储路径）

**上传数据库**
- 拖拽上传或点击选择文件
- 显示上传进度条
- 上传后自动打开

### 4.2 数据库概览
显示选中数据库的详细信息：
- 数据库文件路径
- 文件大小
- SQLite版本
- 创建时间
- 最后修改时间
- 表数量
- 总行数

### 4.3 表管理
**表列表**
- 树形结构展示所有表
- 显示每张表的行数
- 支持搜索过滤
- 右键菜单：重命名、删除、导出

**表结构查看**
- 字段列表：列名、数据类型、可空、默认值、主键
- 支持添加、删除、修改字段
- 支持添加、删除索引
- 外键关系可视化

**表数据查看**
- 分页展示（每页100条）
- 列筛选和排序
- 数据编辑：
  - 双击单元格编辑
  - 新增行
  - 删除行
  - 保存修改
- 支持CSV/JSON导出

### 4.4 数据编辑
**单元格编辑**
- 双击进入编辑模式
- 支持字符串、数字、NULL、JSON编辑
- Tab键切换到下一个单元格
- Esc取消编辑

**新增/删除行**
- 工具栏按钮或快捷键
- 确认对话框
- 批量操作支持

**事务管理**
- 显示当前事务状态
- 提交/回滚按钮
- 自动保存前自动开启事务

### 4.5 SQL查询（可选扩展）
- SQL输入框（带语法高亮）
- 执行查询
- 结果表格展示
- 查询历史记录

### 4.6 错误处理
- Toast通知：成功/警告/错误
- 错误详情模态框
- 操作前的二次确认
- 空状态引导

## 5. 组件清单

### 5.1 布局组件

**AppHeader**
- Logo + 应用名称
- 数据库选择下拉框
- 连接状态指示器
- 状态：默认、已连接、断开

**AppSidebar**
- 数据库信息卡片
- 表树形列表
- 可折叠（移动端默认折叠）
- 状态：展开、折叠、加载中

**AppFooter/StatusBar**
- 连接信息
- 执行时间
- 数据统计

### 5.2 功能组件

**DatabaseSelector**
- 打开本地文件按钮
- 上传文件区域
- 最近文件列表
- 状态：默认、拖拽悬浮、上传中

**TableTree**
- 树形结构
- 表图标 + 表名 + 行数
- 搜索框
- 状态：默认、选中、悬浮、加载中、空

**DataTable**
- 虚拟滚动表格
- 表头排序指示器
- 行悬浮高亮
- 单元格可编辑
- 分页器
- 状态：加载中、空、错误、编辑中

**TableSchema**
- 字段卡片列表
- 拖拽排序（可选）
- 编辑模式切换
- 状态：查看、编辑、保存中

**Pagination**
- 页码导航
- 每页条数选择
- 总条数显示
- 状态：默认、禁用

**Modal**
- 标题栏
- 内容区
- 操作按钮
- 背景模糊
- 动画：缩放淡入

**Toast**
- 图标 + 消息 + 关闭按钮
- 自动消失（3秒）
- 堆叠显示
- 类型：success, warning, error, info

**ConfirmDialog**
- 图标 + 消息
- 取消/确认按钮
- 危险操作红色确认按钮

### 5.3 表单组件

**Input**
- 标签 + 输入框 + 错误提示
- 类型：text, number, textarea
- 状态：默认、聚焦、错误、禁用

**Select**
- 单选/多选下拉框
- 搜索过滤
- 状态：默认、展开、选中、禁用

**Button**
- 类型：primary, secondary, danger, ghost
- 尺寸：small, medium, large
- 状态：默认、悬浮、点击、加载中、禁用

**Checkbox/Switch**
- 标签
- 状态：未选中、选中、不确定、禁用

## 6. 技术方案

### 6.1 前端技术栈
- **框架**: Vue 3.4+ (Composition API + `<script setup>`)
- **构建工具**: Vite 5+
- **UI组件库**: 基于Tailwind CSS自定义组件
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **HTTP客户端**: Axios
- **图标**: Lucide Vue
- **日期处理**: Day.js
- **表格虚拟滚动**: vue-virtual-scroller

### 6.2 后端技术栈
- **语言**: Go 1.21+
- **Web框架**: Gin
- **数据库驱动**: go-sqlite3 (CGO)
- **文件上传**: 标准库multipart
- **CORS**: gin-contrib/cors

### 6.3 API设计

#### 数据库操作
```
POST   /api/database/open           # 打开本地数据库
POST   /api/database/upload         # 上传数据库文件
GET    /api/database/info           # 获取数据库信息
DELETE /api/database/close          # 关闭当前连接
```

#### 表操作
```
GET    /api/tables                   # 获取所有表
GET    /api/tables/:name/schema      # 获取表结构
PUT    /api/tables/:name/schema      # 修改表结构
POST   /api/tables                   # 创建表
DELETE /api/tables/:name             # 删除表
```

#### 数据操作
```
GET    /api/tables/:name/data        # 获取表数据（分页）
POST   /api/tables/:name/data        # 新增数据
PUT    /api/tables/:name/data/:id    # 更新数据
DELETE /api/tables/:name/data/:id    # 删除数据
```

#### 索引操作
```
GET    /api/tables/:name/indexes     # 获取索引列表
POST   /api/tables/:name/indexes     # 创建索引
DELETE /api/tables/:name/indexes/:name # 删除索引
```

### 6.4 数据模型

#### DatabaseInfo
```json
{
  "path": "string",
  "name": "string",
  "size": "number",
  "sqliteVersion": "string",
  "createdAt": "string",
  "modifiedAt": "string",
  "tableCount": "number",
  "totalRows": "number"
}
```

#### TableSchema
```json
{
  "name": "string",
  "columns": [
    {
      "name": "string",
      "type": "string",
      "nullable": "boolean",
      "defaultValue": "any",
      "primaryKey": "boolean"
    }
  ],
  "indexes": [
    {
      "name": "string",
      "columns": ["string"],
      "unique": "boolean"
    }
  ]
}
```

#### TableData (分页)
```json
{
  "data": [
    { "column1": "value1", "column2": "value2" }
  ],
  "total": "number",
  "page": "number",
  "pageSize": "number"
}
```

### 6.5 项目结构

```
sqlite-manage/
├── backend/                 # Go后端
│   ├── main.go
│   ├── handlers/
│   │   ├── database.go
│   │   ├── tables.go
│   │   └── data.go
│   ├── models/
│   │   └── models.go
│   ├── database/
│   │   └── sqlite.go
│   └── go.mod
│
├── frontend/                # Vue前端
│   ├── public/
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   │   ├── layout/
│   │   │   ├── database/
│   │   │   ├── table/
│   │   │   └── common/
│   │   ├── views/
│   │   │   ├── HomeView.vue
│   │   │   └── DatabaseView.vue
│   │   ├── stores/
│   │   │   └── database.js
│   │   ├── api/
│   │   │   └── index.js
│   │   ├── router/
│   │   │   └── index.js
│   │   ├── App.vue
│   │   └── main.js
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   └── tailwind.config.js
│
├── SPEC.md
└── README.md
```

## 7. 实现优先级

### Phase 1 - 核心功能
1. ✅ 项目初始化（Vite + Vue + Tailwind + Go）
2. ✅ 数据库连接（打开/上传）
3. ✅ 数据库信息展示
4. ✅ 表列表展示
5. ✅ 表结构查看

### Phase 2 - 数据操作
6. ✅ 表数据分页展示
7. ✅ 数据编辑（新增/修改/删除）
8. ✅ 表结构修改（新增/删除字段）

### Phase 3 - 增强功能
9. ✅ 表创建和删除
10. ✅ 索引管理
11. ✅ 数据导出（CSV/JSON）
12. ✅ 响应式优化

### Phase 4 - 优化完善
13. ✅ SQL查询界面
14. ✅ 查询历史
15. ✅ 快捷键支持
16. ✅ 错误处理优化
