# Data Nexus — 实施路线图

> 版本: v0.2 · 最后更新: 2026-06-07

---

## 阶段概览

```
Phase 0 ──► Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4
 文档评审     Wails+Driver   前端骨架      功能集成      桌面端发布
 (当前)      + Services     + 布局        联调测试      v0.1.0
```

---

## Phase 0: 文档与设计评审（当前阶段）

**目标:** 产品与技术方案对齐，评审通过后进入编码。

**交付物:**
- [x] PRD（产品需求文档）
- [x] 技术设计文档（架构、API、数据模型）
- [x] UI/UX 设计规范
- [x] 实施路线图
- [ ] 文档评审确认

**评审检查项:**
- [ ] MVP 范围是否合理（不过大/过小）
- [ ] API 契约是否满足前端所有场景（Wails Service 绑定）
- [ ] Driver 抽象是否足够支撑 v1.x 多数据库
- [ ] 开放问题（PRD §8）是否已有决策

**预计耗时:** 1-2 天（评审讨论）

---

## Phase 1: Wails 骨架 + Driver + Services

**目标:** 可独立测试的 Wails 桌面应用骨架，Driver 与 Service 层支持连接、Schema、查询。

### 1.1 项目初始化

- [ ] `wails init`（React-TS 模板）+ 调整目录结构
- [ ] `wails.json` 配置（窗口尺寸、标题、图标）
- [ ] `internal/service` + `internal/driver` 目录
- [ ] Makefile/Taskfile: `dev`, `build`, `test`, `lint`
- [ ] 应用菜单：File → Open / Quit

### 1.2 Driver 层

- [ ] `driver.Driver` 接口定义
- [ ] `sqlite` 实现：Connect / Close / Ping
- [ ] Schema：`ListTables`, `GetTableSchema`（PRAGMA）
- [ ] Data：`BrowseTable`（分页+排序）
- [ ] Query：`QueryRows`, `Exec`
- [ ] 标识符校验 + SQL 注入防护
- [ ] 单元测试（内存 SQLite + 临时文件）

### 1.3 Wails Services 层

- [ ] `ConnectionManager`（单连接）
- [ ] `ConnectionService` / `SchemaService` / `TableService` / `QueryService`
- [ ] `DialogService`（OpenDatabaseFile）
- [ ] `AppService`（GetVersion）
- [ ] `model.AppError` 统一错误
- [ ] Service 层集成测试（不启动 UI）

### 1.4 Wails 绑定验证

- [ ] `main.go` 注册所有 Services
- [ ] 前端最小页面调用 `GetVersion` + `OpenDatabaseFile` + `Connect`
- [ ] 验证 `wailsjs/` 自动生成绑定

**里程碑:** `wails dev` 下可通过文件对话框连接 SQLite 并 ListTables

**预计耗时:** 3-5 天

---

## Phase 2: 前端骨架 + 核心 UI

**目标:** 完整 React 桌面 UI 骨架，连接与 Schema 浏览可用。

### 2.1 项目初始化

- [ ] 在 Wails `frontend/` 上集成 shadcn/ui + Tailwind
- [ ] ESLint + Prettier
- [ ] `lib/api/` 封装 wailsjs 调用 + 错误映射
- [ ] TanStack Query + Zustand 配置

### 2.2 布局与连接

- [ ] AppShell（Header + Sidebar + Main + StatusBar）
- [ ] Welcome 页：「打开数据库」按钮 → DialogService
- [ ] 可选路径输入 + 只读模式勾选
- [ ] 连接状态管理
- [ ] 主题切换（light/dark/system）

### 2.3 Schema 侧边栏

- [ ] SchemaSidebar 组件
- [ ] 表/视图分组列表
- [ ] 搜索过滤
- [ ] 选中态 + 点击联动

**里程碑:** 桌面窗口内可打开 `.db` 并展示表列表

**预计耗时:** 3-4 天

---

## Phase 3: 功能集成

**目标:** 完整 MVP 功能闭环。

### 3.1 表结构 & 数据浏览

- [ ] SchemaTable（列定义）
- [ ] IndexList（P1）
- [ ] DataGrid + Pagination + 排序
- [ ] NULL / BLOB 渲染
- [ ] Loading / Empty / Error 状态

### 3.2 SQL 工作台

- [ ] Monaco Editor 集成
- [ ] 执行 + 快捷键
- [ ] QueryResultPanel（结果集 + exec 结果）
- [ ] 写操作 ConfirmDialog
- [ ] QueryHistory（P1）

### 3.3 联调与 polish

- [ ] 前后端 Service 联调
- [ ] 错误场景覆盖（文件不存在、锁定、SQL 错误、对话框取消）
- [ ] StatusBar 信息展示
- [ ] 窗口最小尺寸、启动参数 `--db`
- [ ] `wails build` 三平台验证

**里程碑:** 完整用户流程可用

**预计耗时:** 4-6 天

---

## Phase 4: 测试 & MVP 发布

**目标:** 稳定可发布的 v0.1.0。

### 4.1 测试

- [ ] Go 单元/集成测试覆盖率 > 70%（driver + service）
- [ ] 前端关键组件测试
- [ ] 手动测试清单执行（见下）

### 4.2 发布准备

- [ ] README 使用说明
- [ ] `wails build` 产物：macOS / Windows / Linux
- [ ] Git tag `v0.1.0`

### 4.3 手动测试清单

- [ ] 应用启动 < 1s 窗口可见
- [ ] File → Open 选择有效 SQLite 文件
- [ ] 连接无效路径 → 错误提示
- [ ] 断开并重连
- [ ] 浏览所有表/视图
- [ ] 查看表结构（列 + 索引）
- [ ] 分页浏览数据（翻页、改 pageSize）
- [ ] 列排序
- [ ] 执行 SELECT 并查看结果
- [ ] 执行 INSERT/UPDATE/DELETE（含确认）
- [ ] SQL 语法错误提示
- [ ] 深色/浅色主题切换
- [ ] 大表（10 万行）首页加载 < 500ms

**预计耗时:** 2-3 天

---

## 总工期估算

| 阶段 | 预估 |
|------|------|
| Phase 0 | 1-2 天 |
| Phase 1 | 3-5 天 |
| Phase 2 | 3-4 天 |
| Phase 3 | 4-6 天 |
| Phase 4 | 2-3 天 |
| **合计** | **13-20 天**（1 人全职） |

---

## v0.2 预览（体验增强 — 竞品标配）

| 功能 | 说明 | 竞品参考 |
|------|------|----------|
| 连接历史持久化 | `~/.data-nexus/connections.json` | Beekeeper |
| CSV 导入/导出 | 表数据 + 查询结果 | sqlite-web, Beekeeper |
| SQL 自动补全 | 表名/列名 Monaco completion | Beekeeper, DuckDB UI |
| SQL 格式化 | 编辑器内 beautify | TablePlus |
| 表数据 inline 编辑 | 双击单元格 → UPDATE | DB4S, Tabulita |
| Native 文件选择器 | Wails OpenFileDialog（MVP 已有） |
| 双击 `.db` 关联打开 | 安装器 + 启动参数 | TablePlus |
| EXPLAIN 可视化 | 查询计划树形展示 | DB4S |

---

## v0.3 预览（探索型差异化 — 核心卖点）

| 功能 | 说明 | 竞品参考 |
|------|------|----------|
| 列数据画像 | 每列 distinct/NULL%/min/max（采样） | DuckDB UI |
| Filter Builder → SQL | 可视化筛选，生成可编辑 WHERE | Datasette |
| URL 状态同步 | 应用内路由状态持久化（非 HTTP URL） | Datasette 思路适配桌面 |
| Canned Queries | 命名保存查询，侧边栏快捷入口 | Datasette |
| FTS 检测与搜索 | 自动识别 FTS 表，提供搜索框 | Datasette |
| Facet 分面筛选 | 低基数列分面（可选） | Datasette |
| SQLite ATTACH | 多 `.db` 文件 attach，分组展示 | DuckDB UI, DB4S |

**预计耗时:** 5-8 天

---

## v1.0 预览（多连接）

| 功能 | 说明 |
|------|------|
| 连接列表 | 保存多个 SQLite 连接 |
| 连接切换 | Sidebar 顶部 dropdown |
| API 升级 | `/connections` 复数端点 |
| 连接别名 | 自定义 displayName |

---

## v1.x 预览（多数据库）

| 功能 | 说明 |
|------|------|
| PostgreSQL | 首个非 SQLite 驱动 |
| MySQL | 第二个驱动 |
| 连接配置 UI | 表单按 type 动态渲染 |
| SSL / 连接测试 | 远程数据库支持 |

---

## 决策日志

| 日期 | 决策 | 理由 |
|------|------|------|
| 2026-06-07 | MVP 单连接，API 用 `/connection` 单数 | 简化实现，v1.0 再扩展复数端点 |
| 2026-06-07 | SQLite 驱动选 modernc.org/sqlite | 纯 Go，跨平台编译无 CGO |
| 2026-06-07 | 前端选 React + shadcn/ui | 生态成熟，定制性强 |
| 2026-06-07 | 不做 MVP 持久化 | 降低复杂度，v0.2 再加 |
| 2026-06-07 | 默认绑定 127.0.0.1 | 本地工具，无需认证 |
| 2026-06-07 | 不做 AI SQL 助手 | 市场拥挤、与本地优先冲突（竞品调研） |
| 2026-06-07 | v0.3 聚焦 Filter→SQL + 列画像 | Datasette/DuckDB UI 验证的需求，SQLite 工具空白区 |
| 2026-06-07 | 只读模式纳入 MVP P0 | sqlite-web/Tabulita 标配，降低误操作风险 |
| 2026-06-07 | 采用 Wails v2 桌面端 | 原生文件对话框、单二进制、对标 TablePlus；弃 HTTP embed 方案 |
| 2026-06-07 | MVP 不做 headless REST | Bridge 足够；`internal/service` 保留便于 v1.x `--server` |

> 开放问题决策后追加到此表。
