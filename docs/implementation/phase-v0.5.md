# v0.5 — 桌面 UI 重构

> **预估:** 8–11 天 · **前置:** v0.4 · **下一版本:** [phase-v0.5.1.md](./phase-v0.5.1.md)

> **v0.5.x 总览:** v0.5.0 桌面 UI 重构 → **v0.5.1 快捷键** → v0.6 `--api` → …

实施时 **以本文档为主清单**；详细方案见 **[v0.5/](./v0.5/README.md)**（实施方案、前后端设计、验收清单）。交互细节见 [UI_UX.md](../design/UI_UX.md)、[CONNECTION_UX.md](../design/CONNECTION_UX.md)。

---

## 1. 版本目标

v0.5 **仅交付 Wails 桌面端 UI 重构**，不引入 HTTP 模式（留 v0.6 / v0.7）。

| 主题 | 说明 |
|------|------|
| 应用内 MenuBar | 文件 / 视图 / 帮助；统一新建连接、设置、关于等入口 |
| 左连接列表 + 右展示区 | 移除侧边栏全局控件；**层级树**（连接→库→表）；右键 + 双击 |
| **连接分组 + 游离连接** | 顶层 Group（默认「我的连接」）与 **游离连接** 第一层同级；Group 可嵌套 |
| **SQLite 持久化** | 连接 + Group 存入 `catalog.db`；废弃 `connections.json`（**不迁移**） |
| 新建/编辑连接 Dialog | 浮动 Modal；方言分区表单；charset/TLS/存储引擎等 |
| 全局 SQL 历史 | MenuBar → 视图；需后端 `ListAllSqlExecutions` |
| Go 原生菜单 | 精简为 App / Edit / Window |

**不在 v0.5：** `--server`、`--api`、前端 Transport 层、驱动深化（见后续版本）。

发布：**tag `v0.5.0`**

---

## 2. 界面设计摘要

### 2.1 整体布局

```
┌──────────────────────────────────────────────────────────────────┐
│  [文件▾] [视图▾] [帮助▾]                         已打开 N 个连接 │
├─────────────────┬────────────────────────────────────────────────┤
│  连接列表        │  [表结构] [数据] [SQL]          连接: app.db ▾ │
│  ▼ 📁 我的连接   │                                                │
│    ▼ 📁 生产     │              Main Content                        │
│      ● mysql     │                                                │
│  ○ staging.db    │  ← 游离连接，与 📁 我的连接 同级                │
│  [SavedQueries]  │                                                │
├─────────────────┴────────────────────────────────────────────────┤
│  Status Bar                                                      │
└──────────────────────────────────────────────────────────────────┘
```

> `[SavedQueries]` — **v0.3 已有功能，v0.5 保持不变**（侧边栏底部 canned queries 区块）。

**Group 组织：** [CONNECTION_GROUPS_AND_STORAGE.md](./v0.5/CONNECTION_GROUPS_AND_STORAGE.md)  
**连接下库表层级：** [CONNECTION_TREE.md](./v0.5/CONNECTION_TREE.md)（连接 → database → [PG: schema] → 表）

### 2.2 MenuBar

| 菜单 | 项 | 快捷键 |
|------|-----|--------|
| 文件 | 新建连接、打开 SQLite、关闭当前连接、退出 | N / O / W |
| 视图 | SQL 执行历史、设置 | `,` |
| 帮助 | 关于 | — |

Dialog 状态提升到 `AppShell`；移除 Header 独立「设置」按钮。

### 2.3 连接列表

| 手势 | 行为 |
|------|------|
| 双击 | 打开连接（未 open → `OpenConnection`；已 open → `activeConnectionId`） |
| 单击 | 选中/高亮，不 open |
| 右键 | 打开/关闭、编辑（含显示名称）、删除；SQLite 已 open 时第五项「附加数据库…」 |

**迁出侧边栏：** 新建连接按钮、`readOnly`/`wal` → `NewConnectionDialog`；启动恢复连接 → `SettingsDialog`。

### 2.4 新建连接 Dialog（重构）

仍为 **浮动 Dialog**（非全屏子页）。UI 重新设计，详见 [v0.5/CONNECTION_FORM.md](./v0.5/CONNECTION_FORM.md)。

```
┌─ 新建连接 ────────────────────────────────────────────────┐
│ [SQLite●] [PostgreSQL] [MySQL]  │  显示名称 / 连接参数…  │
│                                 │  ▼ 常规  ▶ 安全  ▶ 高级 │
│                                 │  [测试连接]  [保存并打开] │
└───────────────────────────────────────────────────────────┘
```

| 方言 | 典型高级项 |
|------|------------|
| SQLite | 只读、WAL |
| PostgreSQL | SSL 模式（6 档）、客户端编码 |
| MySQL | TLS、charset、collation、默认存储引擎 |

编辑连接（右键 → 编辑）与新建 **共用** `ConnectionForm` 组件。

---

## 3. 功能清单

### 3.1 前端基元与 MenuBar

- [ ] `@radix-ui/react-menubar`、`@radix-ui/react-context-menu`
- [ ] `frontend/src/components/ui/Menubar.tsx`、`ContextMenu.tsx`
- [ ] `frontend/src/components/AppMenuBar.tsx`
- [ ] 重构 `Header.tsx`；`AppShell.tsx` Dialog 状态 + 快捷键

### 3.2 连接 Group 与 catalog.db

- [ ] `connection_groups.sort_order`、`sidebar_root_items` 表
- [ ] 默认根 Group「我的连接」；`ConnectionGroupService.GetSidebarTree`
- [ ] Group CRUD + **RenameGroup**（前端 inline，Enter 保存）
- [ ] **MoveGroup** / **ReorderSidebarRoot** / **ReorderGroupMembers**
- [ ] **free_connections** 表；`GetSidebarTree` 返回 `groups` + `freeConnections`
- [ ] 删除 Group 且不删连接 → **游离连接**（第一层，非父 Group）
- [ ] `CountConnectionsInGroup` API + 删除确认 Dialog
- [ ] `ConnectionStore` 改为 Catalog SQLite 实现（**移除** JSON 读写）
- [ ] 前端：`ConnectionGroupNode` + 右键 **新建/重命名/删除** + **DnD**
- [ ] 前端：`ConnectionTreeItem` + 右键 **打开/关闭/编辑/删除**（SQLite 第五项「附加数据库…」）
- [ ] 测试：嵌套 Group、拖拽排序/改父级、首次启动初始化、删除 Group 保留连接

### 3.3 连接列表与 schema 层级树

- [ ] `ConnectionSchemaTree` — 连接 → database → [PG: schema] → 表
- [ ] **后端：** `Driver.ListNamespaces` / `ListSchemas`；`ListTables` namespace 参数
- [ ] 移除 `RemoteNamespaceSwitch` / `SchemaSubtree`
- [ ] `workspaceStore`：`browseContext`
- [ ] **`AttachDatabaseDialog`** + SQLite 连接右键附加 + 拖拽 attach + Finder 显示
- [ ] 测试：SQLite / MySQL / PG 懒加载

### 3.3.1 设置

- [ ] `SettingsDialog`：迁入 `RestoreOnStartupToggle`

### 3.3.2 新建/编辑连接表单（UI 重构）

- [ ] `ConnectionForm/` 组件族：侧栏方言 + 常规/安全/高级分区
- [ ] 重构 `NewConnectionDialog`；`EditConnectionDialog` 共用表单
- [ ] 废弃/合并 `RemoteConnectionForm.tsx`
- [ ] **后端模型扩展：** MySQL `charset`/`collation`/`defaultStorageEngine`；PG `clientEncoding`、sslMode 全量
- [ ] Driver / DSN / Connect；配置写入 **`catalog.db`**
- [ ] i18n：`connectionForm.*`
- [ ] 测试：各方言字段、charset→collation 联动、编辑/password 留空、**open 连接拒绝 Update**

### 3.4 Dialog

- [ ] `SqlExecutionHistoryDialog.tsx` + `listAllExecutions` API
- [ ] `AboutDialog.tsx`（i18n + `AppService.GetPlatform()`，替代原生 `ShowAbout`）
- [ ] i18n：`menu.*`、`sqlHistory.*`、`about.*`、`settings.restoreOnStartup`

### 3.5 后端（全局 SQL 历史）

- [ ] `ListAllExecutions` store / service / Wails 绑定
- [ ] 单元测试

### 3.6 Go 原生菜单

- [ ] `ApplicationMenu()` 仅 App / Edit / Window
- [ ] 文件/设置/关于改由前端 MenuBar；清理 `handleOpenDatabase` 等（前端接管）

---

## 4. 手动验收

- [ ] MenuBar 各菜单可用；无侧边栏/Header 重复入口
- [ ] 双击/右键连接操作正常
- [ ] Group 嵌套 + 默认「我的连接」；`catalog.db` 持久化（**无 json 迁移**）
- [ ] Group **inline 重命名**（Enter）；右键 **新建/重命名/删除**；**DnD** 改父级与顺序
- [ ] 连接右键 **打开/关闭/编辑/删除**；**编辑仅 closed**（含显示名称）
- [ ] **schema 层级树：** MySQL/PG 多 database；PG schema；SQLite main+attach + Attach UI
- [ ] 新建/编辑连接 Dialog：三方言字段正确
- [ ] 设置：启动恢复
- [ ] 视图 → SQL 执行历史、帮助 → 关于
- [ ] macOS App 菜单 Quit；无重复 File/View/Help

---

## 5. 非目标

- **`connections.json` 迁移**（不读取、不导入、不备份旧 json）
- `--server` / `--api`（[v0.6](./phase-v0.6.md) / [v0.7](./phase-v0.7.md)）
- ER 图、插件、驱动深化（[v0.8](./phase-v0.8.md)）

---

## 6. 完成标准

- [ ] 上述 checklist 全部勾选
- [ ] tag **`v0.5.0`**

---

## 7. 相关文档

| 文档 | 用途 |
|------|------|
| **[v0.5/IMPLEMENTATION.md](./v0.5/IMPLEMENTATION.md)** | **主实施方案**（任务分解、顺序、文件清单） |
| [v0.5/DESIGN.md](./v0.5/DESIGN.md) | v0.5 UI/UX 设计规范 |
| [v0.5/BACKEND.md](./v0.5/BACKEND.md) | 后端：`ListAllExecutions`、Go 菜单 |
| [v0.5/CONNECTION_GROUPS_AND_STORAGE.md](./v0.5/CONNECTION_GROUPS_AND_STORAGE.md) | **Group 嵌套 + catalog.db** |
| [v0.5/CONNECTION_TREE.md](./v0.5/CONNECTION_TREE.md) | 连接下 schema 层级树 |
| [v0.5/CONNECTION_FORM.md](./v0.5/CONNECTION_FORM.md) | 新建/编辑连接表单 |
| [v0.5/FRONTEND.md](./v0.5/FRONTEND.md) | 前端组件实施细节 |
| [v0.5/manual-checklist.md](./v0.5/manual-checklist.md) | 手动验收清单 |
| [UI_UX.md](../design/UI_UX.md) | MenuBar、两栏布局 |
| [CONNECTION_UX.md §5](../design/CONNECTION_UX.md) | 右键菜单 |
| [API.md §14](../design/API.md#14-sqlexecutionservicev03) | `ListAllSqlExecutions` |
| [phase-v0.5.1.md](./phase-v0.5.1.md) | 下一版本：快捷键与 macOS ⌘, |
| [phase-v0.4-multi-connection.md](./phase-v0.4-multi-connection.md) | 上一版本 |
