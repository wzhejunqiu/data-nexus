# Data Nexus — UI/UX 设计

> 版本: v0.5 · 状态: 草案 · 最后更新: 2026-06-08  
> **v0.5–v0.8 实施:** [phase-v0.5.md](../implementation/phase-v0.5.md) · [phase-v0.5.1.md](../implementation/phase-v0.5.1.md) · [phase-v0.6.md](../implementation/phase-v0.6.md) · [phase-v0.7.md](../implementation/phase-v0.7.md) · [phase-v0.8.md](../implementation/phase-v0.8.md)

---

## 1. 设计原则

| 原则 | 说明 |
|------|------|
| **效率优先** | 核心路径（连接 → 看表 → 写 SQL）不超过 3 次点击 |
| **信息密度适中** | 开发者工具风格，紧凑但不拥挤 |
| **状态可见** | 连接状态、加载、错误始终有明确反馈 |
| **安全提示** | 写操作必须有确认，危险操作视觉区分 |
| **桌面原生** | Wails 窗口 + 文件对话框；或 v0.5 **Server 模式** 下浏览器访问同一 UI |
| **可扩展布局** | 侧边栏结构预留多连接树形导航 |

### 1.1 桌面壳（Wails）

| 元素 | 说明 |
|------|------|
| 窗口 | 默认 1280×800，最小 960×600 |
| 标题栏 | `Data Nexus — app.db`（连接后显示文件名） |
| 应用内 MenuBar（v0.5） | **Win/Linux：** 顶栏 **文件**（新建连接 / 打开 SQLite / 关闭 / **新建分组** / 退出）、**视图**、**帮助**；**macOS：** 应用内 **不渲染** Menubar，等价项在系统菜单栏 |
| 原生菜单（v0.5） | **macOS：** `app_menu_darwin.go` — App / 文件 / 编辑 / 视图 / 窗口 / 帮助（`EventsEmit` → 前端 Dialog）；**Win/Linux：** 仅 App / Edit / Window，File/View/Help 由应用内 MenuBar 提供 |
| 文件打开 | 系统原生对话框（MenuBar「打开 SQLite」或新建连接 Dialog） |
| 拖拽 | 拖 `.db` 到窗口：**空白区** → 新建连接；**已 open 的 SQLite 连接** → `AttachDatabaseDialog` 预填路径（`app:file-drop` + 前端命中检测） |

### 1.2 HTTP 模式 UI（v0.7，仅 `--server`）

> v0.6 `--api` 无 Web UI。本节适用于 [phase-v0.7.md](../implementation/phase-v0.7.md)。

| 元素 | 说明 |
|------|------|
| 启动 | `./data-nexus --server` — **不** 打开 Wails 窗口 |
| 访问 | 浏览器打开 `http://127.0.0.1:8080/` |
| 布局 | 与桌面相同的 React UI（MenuBar + 左连接列表 + 右展示区） |
| API | HTTP REST；前端 Transport 层替代 Wails Bridge |

**`--api` 模式无 Web UI**（v0.6），不适用本节。见 [phase-v0.6.md](../implementation/phase-v0.6.md)。

---

## 2. 视觉风格

### 2.1 设计语言

- **风格**: 现代开发者工具风，参考 VS Code / Linear 的克制美学
- **组件库**: shadcn/ui（Radix 原语 + Tailwind）
- **图标**: Lucide React
- **字体**: 
  - UI: `Inter`, system-ui
  - 代码/SQL: `JetBrains Mono`, `Fira Code`, monospace

### 2.2 色彩（CSS Variables）

```css
/* 浅色主题 */
--background: 0 0% 100%;
--foreground: 222 47% 11%;
--muted: 210 40% 96%;
--accent: 221 83% 53%;        /* 主色：蓝 */
--destructive: 0 84% 60%;     /* 危险：红 */
--success: 142 76% 36%;
--border: 214 32% 91%;

/* 深色主题 */
--background: 222 47% 11%;
--foreground: 210 40% 98%;
--muted: 217 33% 17%;
--accent: 217 91% 60%;
```

### 2.3 间距与圆角

- 基础间距: 4px 网格（4, 8, 12, 16, 24, 32）
- 圆角: `sm=4px`, `md=6px`, `lg=8px`
- 侧边栏宽度: 260px（可拖拽调整，P1）

---

## 3. 页面结构

### 3.1 整体布局（v0.5）

主工作台为 **左连接列表 + 右展示区** 两栏；全局操作经 MenuBar（Win/Linux 应用内顶栏，macOS 系统菜单栏）。

**Win / Linux：**

```
┌──────────────────────────────────────────────────────────────────┐
│  MenuBar: [文件▾] [视图▾] [帮助▾]              已打开 N 个连接   │
├─────────────────┬────────────────────────────────────────────────┤
```

**macOS：** 窗口内顶栏仅「已打开 N 个连接」；菜单在屏幕顶部系统栏。

```
├─────────────────┬────────────────────────────────────────────────┤
│  连接列表        │  Tab: [表结构] [数据] [SQL]    连接: app.db ▾  │
│  ▼ 📁 我的连接   │                                                │
│    ● mysql       │              Main Content Area                 │
│  ○ staging.db    │  ← 游离连接                                    │
│  [SavedQueries]  │                                                │
├─────────────────┴────────────────────────────────────────────────┤
│  Status Bar: 行数 | 耗时 | 版本                                  │
└──────────────────────────────────────────────────────────────────┘
```

| 区域 | 职责 |
|------|------|
| 顶栏 MenuBar | 新建连接、打开 SQLite、关闭连接、**新建分组**、SQL 历史、设置、关于（macOS：系统菜单栏） |
| 左侧连接列表 | 顶层 Group + 游离连接（同级）；已 open 连接展开 schema 子树；**文件 → 新建分组** 或 **根层空白右键** |
| 右侧展示区 | Tab + 主内容（结构 / 数据 / SQL） |

### 3.2 路由

| 路径 | 页面 | MVP |
|------|------|-----|
| `/` | 主工作台（连接树 + 主内容区） | ✓ |

MVP 采用单页应用；**无 Welcome 页**，启动即进入主工作台。见 [CONNECTION_UX.md](./CONNECTION_UX.md)。

---

## 4. 关键页面与交互

### 4.1 主界面 — 连接列表（Navicat 模式，v0.5）

**启动即显示本布局**（无 Welcome）。左侧为紧凑 **连接列表**，管理操作经 **右键菜单**；**双击** 打开连接。

```
├─ 连接列表 ───────────────────┬─ 展示区 ─────────────────────────┤
│  ● app.db                    │  [结构|数据|SQL]  连接: app.db ▾│
│    └ users                   ├──────────────────────────────────┤
│  ○ staging.db                │         Content                  │
│  [SavedQueries]              │                                  │
└──────────────────────────────┴──────────────────────────────────┘
```

**交互（v0.5）：**

| 手势 | 行为 |
|------|------|
| **双击** 连接 item | **打开连接**：未打开 → `OpenConnection`；已打开 → 设为 `activeConnectionId` |
| **单击** 连接 item | 仅选中/高亮，不 open |
| **右键** 连接 item | 上下文菜单：打开/关闭、编辑连接、删除 |
| 点击表名 | 主区绑定 `connectionId + tableName` |
| SQL Tab | 连接下拉切换 SQL 上下文；保留 per-connection 历史下拉 |

**右键菜单项（按状态）：**

| 菜单项 | 条件 |
|--------|------|
| 打开 | `status !== 'open'` |
| 关闭 | `status === 'open'` |
| 重命名 | —（displayName 在 **编辑连接** Dialog） |
| 编辑连接 | **仅** `status !== 'open'`（已 open 时 disabled；含显示名称） |
| 删除 | 始终（danger；打开中需 confirm） |

**自左侧移除（v0.5）：** 「新建连接」按钮、`readOnly`/`wal` 勾选（迁入新建连接 Dialog）、「启动时恢复已打开连接」（迁入设置）。

详见 [CONNECTION_UX.md §5](./CONNECTION_UX.md#5-v05-连接列表与右键菜单)。

### 4.2 顶栏 — AppMenuBar（v0.5，平台差异）

**Win / Linux — 应用内：**

```
[文件▾] [视图▾] [帮助▾]                    已打开 2 个连接
```

**macOS — 系统菜单栏 + 应用内状态行：**

- 菜单在屏幕顶部系统栏（`app_menu_darwin.go`）
- 窗口内顶栏 **无** Menubar，仅「已打开 N 个连接」

**文件**

| 项 | 行为 | 快捷键 |
|----|------|--------|
| 新建连接… | `NewConnectionDialog` | `Cmd/Ctrl+N` |
| 打开 SQLite 文件… | 文件对话框 → `OpenConnectionFromFile`（快速保存并 open，默认非只读） | `Cmd/Ctrl+O` |
| 关闭当前连接 | `CloseConnection(activeConnectionId)` | `Cmd/Ctrl+W` |
| 新建分组… | `CreateGroup('', name)` 创建顶层 Group | — |
| 退出 | Quit（macOS 亦可经 App 菜单） | `Cmd/Ctrl+Q`（macOS App 菜单） |

**视图**

| 项 | 行为 | 快捷键 |
|----|------|--------|
| SQL 执行历史… | `SqlExecutionHistoryDialog`（跨连接） | — |
| 设置… | `SettingsDialog`（含主题、语言、日志、启动恢复、**快捷键 v0.5.1**） | `Cmd/Ctrl+,` **固定** |

**帮助**

| 项 | 行为 |
|----|------|
| 关于 Data Nexus | `AboutDialog`（版本、平台；i18n） |

- 主题 / 语言：迁入设置 Dialog（不再单独 Header 按钮）
- 导出/导入向导期间：MenuBar 可见；部分 File 项 disabled；右侧显示向导标题

### 4.3 Sidebar — Schema 树

```
搜索表...  🔍
─────────────
TABLES (12)
  orders
  products
  users
VIEWS (2)
  active_users
  order_summary
```

**交互:**
- 点击表名 → 主区域切到「数据」Tab 并加载第一页
- 右键菜单（P1）：查看结构 / 查看数据 / 复制表名
- 搜索实时过滤（前端 filter，无需 API）
- 当前选中表高亮

### 4.4 Tab: 表结构

选中表 `users` 时展示：

| 列名 | 类型 | 主键 | 可空 | 默认值 |
|------|------|------|------|--------|
| id | INTEGER | ✓ | | |
| email | TEXT | | | |
| created_at | TEXT | | ✓ | CURRENT_TIMESTAMP |

下方 **索引** 区块（P1）：

| 名称 | 列 | 唯一 |
|------|-----|------|
| sqlite_autoindex_users_1 | email | ✓ |

### 4.5 Tab: 数据浏览

```
┌─────────────────────────────────────────────────────────────┐
│ users                                    < 1 2 3 ... 31 >   │
│ 每页 [50 ▾]  共 1,523 行                                      │
├────┬──────────────────┬─────────────────────┬───────────────┤
│ id ↑│ email            │ created_at          │               │
├────┼──────────────────┼─────────────────────┼───────────────┤
│ 1  │ a@example.com    │ 2026-01-01T00:00:00Z│               │
│ 2  │ b@example.com    │ NULL                │               │
└────┴──────────────────┴─────────────────────┴───────────────┘
```

**交互:**
- 列头点击切换排序（↑ ↓ ↕ 三态）
- 分页器：首页/末页/页码
- NULL 单元格：灰色斜体
- BLOB：`[BLOB 1024 bytes]`
- Loading：表格 skeleton
- 空表：Empty state 插画 + 「此表暂无数据」

#### 4.5.1 行内编辑与批量提交（v0.2）

```
┌─────────────────────────────────────────────────────────────┐
│ users  [导出 CSV ▾]              < 1 2 3 ... 31 >           │
│ 每页 [50 ▾]  共 1,523 行                                      │
├────┬──────────────────┬─────────────────────┬───────────────┤
│ id │ email *          │ created_at          │               │  ← * 表示已修改
├────┼──────────────────┼─────────────────────┼───────────────┤
│ 1  │ [new@example.com]│ 2026-01-01T...      │               │  ← 编辑中 input
│ 2  │ b@example.com    │ NULL                │               │
└────┴──────────────────┴─────────────────────┴───────────────┘
┌─ 待提交 2 处变更 ────────────────── [放弃] [提交变更 ▶] ─────┐
└─────────────────────────────────────────────────────────────┘
```

**交互:**
- 双击单元格 → 原地 `<input>` / 类型适配编辑器（TEXT/INTEGER/REAL）
- 修改后不立即写库；加入 `pendingEdits`，单元格左侧或背景高亮
- 底部 **EditBatchBar** 显示待提交数量；「放弃」清空 pending
- 「提交变更」→ ConfirmDialog 摘要（表名、变更条数、样例行）→ `UpdateCellsBatch`
- 成功：Toast + 刷新当前页 + 清空 pending
- 失败：Toast + SQL 错误；pending 保留供用户修正
- 只读连接：双击无反应或 Tooltip「只读模式」
- BLOB 列：禁用编辑，Tooltip 说明
- Esc：取消当前格编辑（未提交到 pending 则丢弃；已在 pending 则恢复原值）
- 切换表 / 分页 / 关闭连接：若有 pending → 确认是否放弃

### 4.6 导出 CSV（v0.2+）

**入口:**
- 数据 Tab 工具栏「导出 CSV」→ 当前页 / 全表
- SQL 结果区工具栏「导出 CSV」

**布局:** 全屏子页面（覆盖连接树与主内容，仅保留顶栏），多步骤向导 `ExportWizardPage`。

**步骤:** 导出选项 → 保存路径 → 导出中 → **结果**（步数指示器仅计前两步配置步骤）

| 步骤 | 内容 |
|------|------|
| 导出选项 | 表导出：范围（当前页/全表，>1 万行警告）+ 列选择 + 格式；SQL 结果：列选择 + 格式 |
| 保存路径 | `SaveFile` 选路径 + 摘要；点「开始导出」 |
| 导出中 | 全表：不确定进度条 + 已导出行数 + 可取消 |
| 结果 | 路径、行数、列数、耗时；**不自动关闭**，用户点「关闭」回主界面；失败时可「上一步」回到保存路径修改后重试 |

**流程:** 向导配置 → 选路径 → 开始导出 → 结果页确认 → 关闭回主界面（成功不再 Toast）

**全表导出:** 不预先 `COUNT(*)`；取消调用 `CancelExportTableCSV` 删除半成品文件，结果页展示「已取消」。

### 4.7 设置（v0.2，v0.5 入口变更）

**入口:** MenuBar → 视图 → 设置（`Cmd/Ctrl+,`）

```
┌─ 设置 ─────────────────────────────────────────────┐
│  常规                                             │
│    启动时恢复已打开连接  [ ]   （v0.5 自左侧迁入）  │
│  外观                                             │
│    主题    [ system ▾ ]                           │
│    语言    [ zh-CN ▾ ]                            │
│  日志                                             │
│    级别    [ info ▾ ]                             │
│    输出    [ auto ▾ ]  console | file | both      │
│    文件路径 [ ~/Library/Logs/...        ]         │
│  [取消]                              [保存]       │
└───────────────────────────────────────────────────┘
```

- 保存调用 `ConfigService.UpdateConfig`
- 日志热更新；无效值 inline 校验

**v0.5.1 快捷键分区：** 见 [phase-v0.5.1.md](../implementation/phase-v0.5.1.md)。`⌘,` / `Ctrl+,` 固定打开设置，不可改绑；其余 MenuBar / SQL 快捷键可在设置页录制修改。

### 4.8 全局 SQL 执行历史（v0.5）

**入口:** MenuBar → 视图 → SQL 执行历史

**组件:** `SqlExecutionHistoryDialog`

| 列 | 说明 |
|----|------|
| 执行时间 | `executedAt` |
| 连接 | 由 `connectionId` 映射连接名 |
| SQL | 摘要（可点击整行） |
| 类型 | result / exec |
| 耗时 / 影响行 | `durationMs` / `effectRows` |

**交互:** 点击行 → 关闭 Dialog → 切换 `activeConnectionId` → SQL Tab → 填充编辑器。

**API:** `SqlExecutionService.ListAllSqlExecutions(limit?)`（默认 50，上限 200）。

SQL Tab 内 per-connection `QueryHistory` 下拉 **保留**（快捷入口，不重复全局能力）。

### 4.9 关于（v0.5）

**入口:** MenuBar → 帮助 → 关于 Data Nexus

**组件:** `AboutDialog` — 应用名、版本（`AppService.GetVersion`）、平台/架构；中英文 i18n。替代原生 `ShowAbout` MessageDialog。

### 4.8 Tab: SQL 编辑器

```
┌─ 历史 ▾ ──────────────────────────────────── [执行 ▶] ────────┐
│  1 │ SELECT * FROM users LIMIT 10;                           │
│  2 │                                                          │
│  3 │                                                          │
└──────────────────────────────────────────────────────────────┘
┌─ 结果 ───────────────────────────────────────────────────────┐
│  1 row · 12ms                                                │
│  ┌────┬──────────────────┐                                   │
│  │ id │ email            │                                   │
│  └────┴──────────────────┘                                   │
└──────────────────────────────────────────────────────────────┘
```

**交互:**
- `Cmd/Ctrl + Enter` 执行
- 执行写操作 → Modal 确认：「将执行 DELETE，是否继续？」
- 错误：结果区顶部红色 alert，展示 SQL 错误原文
- 历史下拉：点击填充到编辑器
- Monaco 配置：最小 8 行，自动增高至 20 行
- **v0.2:** 表名/列名自动补全（CompletionItem）
- **v0.2:** 工具栏「格式化」「EXPLAIN」

#### 4.8.1 CSV 导入向导（v0.2）

**入口:** 数据 Tab 工具栏「导入 CSV」

**布局:** 全屏子页面（覆盖连接树与主内容，仅保留顶栏），多步骤向导 `ImportWizardPage`。

**步骤:** 文件与格式 → 目标与模式 → 列映射与确认 → 导入中 → **结果**（步数指示器仅计前三步配置步骤）

| 步骤 | 内容 |
|------|------|
| 文件与格式 | `OpenCSVFile` 选文件 + `CSVFormatOptions`；点「预览」解析前 20 行 |
| 目标与模式 | 数据预览；新建表 / 已有表；`append` / `update`（需主键） |
| 列映射与确认 | CSV 列映射；新建表含列类型/主键、`CREATE TABLE` 预览（`SqlCodeView`）、导入摘要；点「开始导入」 |
| 导入中 | loading 文案（同步导入） |
| 结果 | 插入/更新行数；**不自动关闭**，用户点「关闭」回主界面；失败时可「上一步」回到列映射与确认修改后重试 |

**流程:** 向导配置 → 开始导入 → 结果页确认 → 关闭回主界面（成功不再 Toast）；关闭后刷新 Schema / 数据 Tab

---

## 5. 组件清单

| 组件 | 用途 | 优先级 |
|------|------|--------|
| `AppShell` | 整体布局框架 | P0 |
| `AppMenuBar` | Win/Linux 应用内顶栏菜单；macOS 隐藏 Menubar（系统菜单栏承担） | v0.5 |
| `ConnectionSchemaTree` | 连接下 database/schema 层级树 + attach L1 右键 | v0.5 |
| `SidebarDndContext` | 连接/Group 拖拽改父级与同级排序 | v0.5 |
| `ConnectionTree` | 左连接列表 + Schema 子树 | P0 |
| `ConnectionTreeItem` | 连接行 + 右键 ContextMenu | v0.5 |
| `NewConnectionDialog` | 新建连接（含 SQLite 只读/WAL） | P0 |
| `SqlExecutionHistoryDialog` | 跨连接 SQL 执行历史 | v0.5 |
| `AboutDialog` | 关于（版本、平台） | v0.5 |
| `ConnectionStatus` | Header 状态 Chip | P0 |
| `SchemaSidebar` | 表/视图列表 | P0 |
| `SchemaTable` | 列定义表格 | P0 |
| `DataGrid` | 分页数据表格 | P0 |
| `EditBatchBar` | 待提交编辑条 + 提交/放弃 | v0.2 |
| `ExportWizardPage` | CSV 导出全屏多步向导 + 结果页 | v0.2+ |
| `ImportWizardPage` | CSV 导入全屏多步向导 + 结果页 | v0.2 |
| `SettingsPanel` | 应用配置（日志等） | v0.2 |
| `SqlEditor` | Monaco 包装 | P0 |
| `QueryResultPanel` | 结果/错误展示 | P0 |
| `ConfirmDialog` | 写操作确认 | P0 |
| `Pagination` | 分页控件 | P0 |
| `EmptyState` | 空状态 | P0 |
| `ThemeToggle` | 主题切换 | P1 |
| `LanguageToggle` | 语言切换 zh-CN / en | P1（i18n 必做） |
| `IndexList` | 索引展示 | P1 |
| `QueryHistory` | SQL Tab 当前连接历史下拉 | P1 |
| `ContextMenu` | Radix 右键菜单基元 | v0.5 |
| `Menubar` | Radix 顶栏菜单基元 | v0.5 |

---

## 6. 状态与反馈

### 6.1 Loading 状态

| 场景 | 表现 |
|------|------|
| 连接中 | 按钮 spinner + 「连接中...」 |
| 加载表列表 | Sidebar skeleton（3-5 行） |
| 加载表数据 | DataGrid skeleton |
| 执行 SQL | 执行按钮 spinner + 结果区 loading |
| 批量提交 | EditBatchBar 按钮 spinner + 「提交中…」 |
| CSV 全表导出 | 进度条 + 可取消；>1 万行非阻塞警告 |
| CSV 导入 | ImportWizardPage 步骤内 loading |

### 6.2 错误状态

| 场景 | 表现 |
|------|------|
| 连接失败 | 表单 inline error |
| API 错误 | Toast + 可展开详情（Wails Service 抛错） |
| SQL 错误 | 结果区 Alert， monospace 展示错误信息 |
| 未打开连接访问表/SQL | 提示「请先打开该连接」 |
| 无已打开连接 | 主区 Empty：「新建或打开一个连接开始」 |
| 批量提交失败 | Toast + 可展开 SQL 详情；pending 保留 |
| CSV 导入失败 | ImportWizardPage 结果页展示错误；事务回滚说明 |
| 只读拦截编辑/导入 | inline 提示 + `READ_ONLY` Toast |

### 6.3 空状态

| 场景 | 文案 |
|------|------|
| 无表 | 「数据库中没有用户表或视图」 |
| 未选表 | 「从左侧选择一个表，或切换到 SQL 编辑器」 |
| 无查询结果 | 「查询成功，返回 0 行」 |

---

## 7. 响应式策略

MVP 目标：**桌面窗口**（Wails 默认窗口，≥ 960px 宽）

| 断点 | 行为 |
|------|------|
| < 768px | 侧边栏 collapsible drawer（P2，非 MVP 必须） |
| ≥ 1024px | 标准三栏布局 |

---

## 8. 无障碍 (A11y)

- 所有交互元素可键盘聚焦
- Modal 焦点 trap
- 表格支持方向键导航（P2）
- 颜色对比度 WCAG AA
- `aria-label` 用于图标按钮

---

## 9. 线框图参考

MVP 核心态 — 多连接 + 数据浏览（v0.5 布局）：

```
┌─────────────────────────────────────────────────────────────────┐
│ ■ Data Nexus  [文件▾][视图▾][帮助▾]      已打开 2 个连接       │
├──────────┬──────────────────────────────────────────────────────┤
│ ● app.db │  users @ app.db  │ 表结构 │ 数据 │ SQL              │
│   users ◀│──────────────────────────────────────────────────────│
│ ○ staging│  id ▲ │ email          │ created_at                    │
│          │  ─────┼────────────────┼──────────────                 │
│          │    1  │ a@example.com  │ 2026-01-01...                 │
├──────────┴──────────────────────────────────────────────────────┤
│ 1,523 rows · page 1/31                          data-nexus v0.5 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 10. 与 PRD 功能映射

| PRD 功能 | UI 入口 |
|----------|---------|
| 连接 SQLite | MenuBar 文件 → **打开 SQLite**（快速）或 **新建连接**（完整 Dialog） |
| 远程连接 | MenuBar 文件 → 新建连接 → PG/MySQL Tab |
| 连接分组 | MenuBar 文件 → **新建分组**；或侧边栏根层空白右键 |
| Schema 浏览 | 左连接列表 Schema 子树 + 表结构 Tab |
| 表数据浏览 | 数据 Tab + DataGrid |
| SQL 编辑器 | SQL Tab |
| SQL 执行历史 | MenuBar 视图 → SQL 执行历史；SQL Tab 内历史下拉 |
| 应用设置 | MenuBar 视图 → 设置 |
| 关于 | MenuBar 帮助 → 关于 |
| 写操作确认 | ConfirmDialog |
| 批量编辑确认 | ConfirmDialog（变更摘要） |
| CSV 导出/导入 | ExportWizardPage / ImportWizardPage |
