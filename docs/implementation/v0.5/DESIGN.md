# v0.5 UI/UX 设计规范

> **全局规范:** [UI_UX.md](../../design/UI_UX.md) · **连接交互:** [CONNECTION_UX.md](../../design/CONNECTION_UX.md)  
> 本文档为 v0.5 实施专用摘要，与全局设计文档一致；冲突时以本文档 + phase-v0.5 为准。

---

## 1. 整体布局

> **Group:** [CONNECTION_GROUPS_AND_STORAGE.md](./CONNECTION_GROUPS_AND_STORAGE.md) · **Schema 树:** [CONNECTION_TREE.md](./CONNECTION_TREE.md)

```
┌──────────────────────────────────────────────────────────────────┐
│  Win/Linux: [文件▾][视图▾][帮助▾]  已打开 N 个连接              │
│  macOS: 系统菜单栏（App/文件/编辑/视图/窗口/帮助）+ 仅状态行     │
├─────────────────┬────────────────────────────────────────────────┤
│  连接列表        │  [表结构] [数据] [SQL]          连接: app.db ▾ │
│  ▼ 📁 我的连接   │                                                │
│    ▼ 📁 生产     │              Main Content                        │
│      ● mysql     │                                                │
│  ○ staging.db    │  ← 游离连接                                    │
│  [SavedQueries]  │                                                │
├─────────────────┴────────────────────────────────────────────────┤
│  Status Bar                                                      │
└──────────────────────────────────────────────────────────────────┘
```

> `[SavedQueries]` — **v0.3 已有功能，v0.5 保持不变**（侧边栏底部 canned queries 区块，无 UI 变更）。

| 区域 | 职责 | v0.5 变更 |
|------|------|-----------|
| 顶栏 MenuBar | 新建连接、打开 SQLite、关闭连接、**新建分组**、SQL 历史、设置、关于 | **新增**；macOS 走系统菜单栏，应用内仅状态行 |
| 左侧连接列表 | 顶层 **Group** + **游离连接**（同级）+ schema 子树 | catalog.db |
| 右侧展示区 | Tab + 主内容 | 不变 |
| Status Bar | 行数、耗时、版本 | 不变 |

**窗口:** 默认 1280×800，最小 960×600（沿用 MVP）。

---

## 2. AppMenuBar

### 2.1 结构（平台差异）

**Win / Linux — 应用内 Menubar：**

```
[文件▾] [视图▾] [帮助▾]                    已打开 2 个连接
```

**macOS — 系统菜单栏 + 应用内状态行：**

```
（屏幕顶部系统栏）Data Nexus | 文件 | 编辑 | 视图 | 窗口 | 帮助
（窗口内顶栏）                              已打开 2 个连接
```

- **Win/Linux：** 左侧 Radix Menubar（`AppMenuBar.tsx`）+ 右侧状态
- **macOS：** 菜单在 **系统菜单栏**（`app_menu_darwin.go`）；应用内 **不渲染** Menubar（`useIsMacOS()`）；顶栏仅右侧 `openCount` / 向导标题（`Header` 高度 36px）
- **`openCount` 数据来源：** 由 `AppShell` 从 `connectionApi.list()` 派生；详见 [FRONTEND.md §1](./FRONTEND.md#1-appmenubar)

### 2.2 文件菜单

| 项 | 行为 | 快捷键（v0.5 固定） |
|----|------|----------------------|
| 新建连接… | 打开 `NewConnectionDialog` | `Cmd/Ctrl+N` |
| 打开 SQLite 文件… | 原生文件对话框 → `OpenConnectionFromFile`（保存到 catalog 并**立即打开**；默认可写、非 WAL） | `Cmd/Ctrl+O` |
| 关闭当前连接 | `CloseConnection(activeConnectionId)`；无活跃连接时 disabled | `Cmd/Ctrl+W` |
| 新建分组… | `CreateGroup('', '新分组')` 创建顶层 Group（`parent_id NULL`） | — |
| 退出 | Wails Quit | macOS：`Cmd+Q`（App 菜单）；Win/Linux：可选 MenuBar 项 |

**「打开 SQLite 文件…」与「新建连接…」：** 前者为**快速通道**——选 `.db` 即用默认参数新建/复用连接并 open，不经 Dialog；后者可配置显示名称、只读/WAL、分组及 PG/MySQL。

**远程连接:** 「新建连接」Dialog 内 Tab 切换 SQLite / PostgreSQL / MySQL（v0.4 已有）。

### 2.3 视图菜单

| 项 | 行为 | 快捷键 |
|----|------|--------|
| SQL 执行历史… | `SqlExecutionHistoryDialog` | —（v0.5.1 可选默认 `Cmd+Shift+H`） |
| 设置… | `SettingsDialog` | `Cmd/Ctrl+,` |

**迁出 Header:** 主题、语言切换仅在设置 Dialog 内（v0.2 已有分区）。

### 2.4 帮助菜单

| 项 | 行为 |
|----|------|
| 关于 Data Nexus | `AboutDialog` |

### 2.5 向导模式

导入/导出 CSV 全屏向导期间：

- 顶栏状态行 **保持可见**（macOS 系统菜单仍可用）
- 「新建连接」「打开 SQLite」「关闭当前连接」「新建分组」→ **disabled**（macOS 由 Go 菜单项同步 disabled 策略待 v0.5.1；v0.5 前端快捷键仍 respect `disabled`）
- 右侧文案：向导标题（如「导出 CSV」），非 openCount

### 2.6 Go 原生菜单（平台差异）

| 平台 | 实现 | 菜单结构 |
|------|------|----------|
| **macOS** | `app_menu_darwin.go` | **App**（系统标准 + Quit）→ **文件** → **Edit** → **视图** → **Window** → **帮助** |
| **Win/Linux** | `app_menu.go` | **App** / **Edit** / **Window**（无 File/View/Help；由应用内 MenuBar 提供） |

**macOS 菜单项 → 前端事件：**

| 菜单项 | `EventsEmit` |
|--------|----------------|
| 新建连接… | `app:new-connection` |
| 打开 SQLite 文件… | `app:open-sqlite` |
| 关闭当前连接 | `app:close-connection` |
| 新建分组… | `app:new-group` |
| SQL 执行历史… | `app:sql-history` |
| 设置… | `app:settings` |
| 关于 Data Nexus | `app:about` |

**macOS HIG：** 不在窗口内重复 File/View/Help；**Quit** 仅 App 菜单（`Cmd+Q`）；Win/Linux MenuBar 保留「退出」项。

**Win/Linux：** 与 v0.5 初版一致 — 原生菜单精简，功能全部由 `AppMenuBar` 承担。

---

## 3. 连接 Group 与游离连接

> **完整规范:** [CONNECTION_GROUPS_AND_STORAGE.md](./CONNECTION_GROUPS_AND_STORAGE.md)

- **第一层同级：** 顶层 📁 Group（含默认「我的连接」）与 **游离连接** 同一缩进
- 「我的连接」是 **默认顶层 Group**，不是包裹全列表的根
- **游离连接：** 不在任何 Group；删除 Group（不删连接）时 **一律变为游离**，不进入父 Group
- Group 名称随时可改；删除含连接时「同时删除连接」默认不勾选

### 3.1 Group / 连接交互

| 能力 | Group | 连接 |
|------|-------|------|
| 重命名 | **原地 inline**，Enter 保存；**无 Dialog** | **无**；displayName 在 **编辑连接** Dialog |
| 右键 | **新建 / 重命名 / 删除**；**根层空白 → 新建分组** | 见 §4.4（PG/MySQL 四项；SQLite 五项） |
| 文件菜单 | **新建分组…**（顶层 Group，与空白右键等价） | — |
| 编辑 | — | **仅 closed**；含显示名称 + 连接配置 |
| 拖改父级 | 拖到 Group 上 / 根层 | 拖到 Group 上 / 根层游离 |
| 拖改顺序 | 同级 `sort_order` | 同级 `sort_order` |

根层 Group 与游离连接通过 `sidebar_root_items` **混合排序**。详见 CONNECTION_GROUPS_AND_STORAGE §3.6。

---

## 4. 连接下 Schema 层级树

> **完整规范:** [CONNECTION_TREE.md](./CONNECTION_TREE.md) · **SQLite Attach:** [CONNECTION_TREE.md §3.6](./CONNECTION_TREE.md#36-sqlite-attach--detach)

### 4.1 统一层级

| 方言 | L1 数据库 | L2 Schema | 叶子 |
|------|-----------|-----------|------|
| SQLite | `main`、attach alias | — | table / view |
| MySQL | 各 database | — | table / view |
| PostgreSQL | 各 database | `public` 等 | table / view |

### 4.2 视觉

- 宽度：260px 默认，可拖拽 200–480px（保留）
- 连接状态：`●` 已打开 / `○` 未打开
- 树节点：缩进 + ▶/▼ 展开；懒加载
- 类型标签：SQL / PG / MY；系统库弱化样式
- SQLite attach L1：badge「已附加」；hover 显示源文件路径

### 4.3 手势

| 手势 | 行为 |
|------|------|
| **双击** 连接 | 未 open → `OpenConnection`；已 open → `activeConnectionId` |
| **单击** 连接 | 选中/高亮，不 open |
| **右键** 连接 | ContextMenu（见 §4.4） |
| **展开** database/schema | 懒加载子节点 |
| **单击** database/schema | 设为 `browseContext` |
| **单击** 表名 | 主区「数据」Tab，`connectionId + tableKey` |
| **拖拽 `.db`** 到已 open SQLite 连接 | 打开 `AttachDatabaseDialog` 并预填路径 |
| **右键** attach L1 | 取消附加（Detach）；在 Finder/文件管理器中显示源文件 |

### 4.4 右键菜单（连接 L0）

**PG / MySQL（四项）：**

| 菜单项 | 显示条件 | API / 行为 |
|--------|----------|------------|
| 打开连接 | `status !== 'open'` | `OpenConnection` |
| 关闭连接 | `status === 'open'` | `CloseConnection` |
| **重命名** | 始终 | inline 编辑显示名（`RenameConnection`）；**open / closed 均可** |
| 编辑连接 | 始终显示；**仅** `status !== 'open'` 可用 | `EditConnectionDialog` + `ConnectionForm`（含 **显示名称** + 全部配置） |
| 删除 | 始终（danger 样式） | `RemoveConnection`；打开中需 confirm |

**SQLite 额外第五项：**

| 菜单项 | 显示条件 | API / 行为 |
|--------|----------|------------|
| 附加数据库… | `type === 'sqlite'`；**仅** `status === 'open'` 可用 | `AttachDatabaseDialog` |
| （未 open 时） | `type === 'sqlite'` **且** `status !== 'open'` | **disabled** + Tooltip「请先打开连接」 |

**编辑连接：** 仅 **连接关闭** 时可编辑连接 **配置**（host/端口/库/密码等）。已 open 时菜单项 **disabled**，Tooltip `connection.editRequiresClosed`。**显示名称** 另有 **inline 重命名**（F2/Enter/右键「重命名」），**open / closed 均可**；与 Edit Dialog 内改显示名 **并存、不冲突**。

**边界：** 连接树中所有节点均来自 `catalog.db` 持久化记录；新建连接在 Dialog 保存成功后才出现在树中，**无**「未保存连接」右键场景。

**移除:** 行内 Open/Close 按钮；`RenameConnectionDialog`；`RemoteNamespaceSwitch` 手输切换。移入/移出 Group 除 **拖拽** 外，新建/Cmd+O/拖 `.db` 到 Group 亦可通过 `selectedGroupId` / drop target 入组。

### 4.5 自左侧迁出

| 控件 | 迁入位置 |
|------|----------|
| 「新建连接」按钮 | MenuBar → 文件 → 新建连接 |
| SQLite `readOnly` / `wal` 勾选 | `NewConnectionDialog` 内（SQLite Tab） |
| 「启动时恢复已打开连接」 | `SettingsDialog` → 常规分区 |
| v0.4 Attach 顶部按钮 | SQLite 连接右键「附加数据库…」+ `AttachDatabaseDialog` |

### 4.6 空状态

**无连接：**

- zh-CN：「暂无连接。使用 文件 → 新建连接 开始。」
- en: "No connections yet. Use File → New Connection to get started."

**无顶层 Group（已删除默认「我的连接」等）：** 列表区显示 `connectionGroup.noGroupsHint`（「在空白处右键可新建分组」）；可用 **文件 → 新建分组…** 或根层空白 **ContextMenu → 新建分组**。

---

## 5. Dialog

### 5.1 新建 / 编辑连接（ConnectionForm 重构）

**完整规范:** [CONNECTION_FORM.md](./CONNECTION_FORM.md)

**呈现:** 浮动 Dialog（约 720px 宽），覆盖主界面；**非**全屏子页。

**布局:** 左侧方言选择（SQLite / PostgreSQL / MySQL）+ 右侧分区表单（常规 / 安全与 TLS / 高级）。

```
┌─ 新建连接 ────────────────────────────────────────────────────────────────┐
│ ┌─────────────┐  ┌─────────────────────────────────────────────────────┐ │
│ │  SQLite  ●  │  │  显示名称  [________________________]               │ │
│ │  PostgreSQL │  │  ▼ 常规  — 主机/端口/库/用户/密码 或 文件路径        │ │
│ │  MySQL      │  │  ▶ 安全与 TLS  — SSL/TLS、只读                      │ │
│ │             │  │  ▶ 高级  — charset、collation、存储引擎、编码…       │ │
│ └─────────────┘  └─────────────────────────────────────────────────────┘ │
│              [ 测试连接 ]                         [ 取消 ] [ 保存并打开 ] │
└────────────────────────────────────────────────────────────────────────────┘
```

**各方言要点:**

| 方言 | 安全 | 高级 |
|------|------|------|
| SQLite | — | 只读、WAL |
| PostgreSQL | SSL 模式（6 档 libpq） | Schema、客户端编码 |
| MySQL | TLS、跳过证书校验 | charset、collation、默认存储引擎 |

- 新建与编辑（右键 → 编辑连接）**共用** `ConnectionForm` 组件
- `readOnly` / `wal` 为 Dialog 内部 state（SQLite 高级区）
- MenuBar「打开 SQLite 文件…」仍为快速通道，不经本 Dialog

### 5.2 SettingsDialog（调整）

**常规分区新增:**

| 项 | 说明 |
|----|------|
| 启动时恢复已打开连接 | checkbox；读写 `ConnectionService` restore 配置 |

i18n key: `settings.restoreOnStartup`

**v0.5.1 预留:** 快捷键分区（本版本不实现）

### 5.3 SqlExecutionHistoryDialog（新增）

**入口:** MenuBar → 视图 → SQL 执行历史

```
┌─ SQL 执行历史 ─────────────────────────────────────────────┐
│  执行时间      连接        SQL 摘要           类型   耗时  │
│  2026-06-08   app.db      SELECT * FROM ...   result 12ms │
│  2026-06-08   staging.db  DELETE FROM ...     exec  5ms   │
│                                              [关闭]        │
└────────────────────────────────────────────────────────────┘
```

| 列 | 字段 |
|----|------|
| 执行时间 | `executedAt` |
| 连接 | `connectionId` → 连接 display name |
| SQL | 截断展示，hover 全文 |
| 类型 | `result` / `exec` |
| 耗时 / 影响行 | `durationMs` / `effectRows` |

**交互:**

1. 打开时调用 `ListAllSqlExecutions(50)`
2. 点击行 → 关闭 Dialog
3. 若目标连接未 open → 先 `OpenConnection`
4. `setActiveConnectionId` + `setActiveTab('sql')`
5. 将 SQL 填入编辑器（不自动执行）

**保留:** SQL Tab 内 per-connection `QueryHistory` 下拉（快捷入口，不重复全局能力）。

### 5.4 AboutDialog（新增）

**入口:** MenuBar → 帮助 → 关于 Data Nexus

```
┌─ 关于 Data Nexus ──────────────────┐
│         [Logo / 应用名]            │
│         版本 0.5.0                 │
│         macOS arm64                │
│                    [关闭]          │
└────────────────────────────────────┘
```

- 版本：`AppService.GetVersion()`
- 平台：**Go 侧** `AppService.GetPlatform()` → `runtime.GOOS` + `runtime.GOARCH`（如 `darwin/arm64`）
- 替代原生 `ShowAbout` MessageDialog
- 全 i18n：`about.title`、`about.version`、`about.platform`

### 5.5 AttachDatabaseDialog（SQLite 附加库）

**入口:** SQLite 连接 L0 右键「附加数据库…」；或拖拽 `.db` 到已 open SQLite 连接。

**完整规范:** [CONNECTION_TREE.md §3.6](./CONNECTION_TREE.md#36-sqlite-attach--detach)

---

## 6. 状态管理

### 6.1 Dialog 状态（AppShell）

```typescript
// AppShell 集中管理
newConnectionOpen: boolean
settingsOpen: boolean
sqlHistoryOpen: boolean
aboutOpen: boolean
attachDatabaseOpen: { connectionId: string; prefilledPath?: string } | null
```

- `EventsOn('app:settings')` 仍可用（Go App 菜单 → 打开设置）
- MenuBar 与快捷键共用同一 setter
- **设置：** AppShell 渲染 `SettingsDialogContainer`（内部包装 `SettingsDialog`）；见 [FRONTEND.md §3.2 / §4](./FRONTEND.md#32-settingsdialog--settingsdialogcontainer)
- **Attach：** `attachDatabaseOpen` 非 null 时渲染 `AttachDatabaseDialog`；右键与拖拽 `.db` 共用此状态

### 6.2 侧边栏选中态（workspaceStore，非持久化）

```typescript
selectedGroupId: string | null   // 新建连接 / Cmd+O / 拖 .db 入组目标
setSelectedGroupId: (id: string | null) => void

sidebarFocus: { kind: 'group' | 'connection'; id: string } | null  // F2/Enter inline 重命名焦点
setSidebarFocus: (focus: ...) => void
```

- 单击 **Group** → 设置 `selectedGroupId` + `sidebarFocus`
- 单击 **连接** → 仅更新 `sidebarFocus`；**保留** `selectedGroupId`
- `ConnectionTree` 监听 **F2 / Enter** → `invokeSidebarRename(sidebarFocus)`（Dialog/Monaco/input 焦点时忽略）

**与 activeConnectionId 区别：**

- `activeConnectionId`：当前工作区绑定的连接（可未打开时亦被单击选中）
- `sidebarFocus`：键盘重命名与行高亮（Group 或连接）

---

## 7. i18n 键清单

| 前缀 | 示例 key |
|------|----------|
| `menu.file.*` | `newConnection`, `openSQLite`, `closeConnection`, `newGroup`, `quit` |
| `menu.view.*` | `sqlHistory`, `settings` |
| `menu.help.*` | `about` |
| `sqlHistory.*` | `title`, `executedAt`, `connection`, `sql`, `kind`, `duration`, `empty` |
| `about.*` | `title`, `version`, `platform` |
| `settings.restoreOnStartup` | 启动恢复开关标签 |
| `connection.emptyHint` | 空列表引导文案 |
| `connectionGroup.*` | `defaultName`, `new`, `createRoot`, `noGroupsHint`, `rename`, `delete`, …（完整表见 [CONNECTION_GROUPS_AND_STORAGE.md §8](./CONNECTION_GROUPS_AND_STORAGE.md#8-i18n)） |
| `connection.attachedBadge` | attach 节点 L1 徽标文案 |
| `connectionForm.*` | 见 [CONNECTION_FORM.md §7](./CONNECTION_FORM.md#7-i18n-键新增) |
| `attach.*` | `dialogTitle`, `attach`, `detach`, `showInFinder`, `sessionHint` 等（见 [CONNECTION_TREE.md §3.6](./CONNECTION_TREE.md#36-sqlite-attach--detach)） |

---

## 8. 组件清单（v0.5 新增/变更）

| 组件 | 状态 | 说明 |
|------|------|------|
| `AppMenuBar` | 新增 | Win/Linux 顶栏菜单；macOS 隐藏（系统菜单栏） |
| `lib/platform.ts` | 新增 | `useIsMacOS()` 平台检测 |
| `ConnectionSchemaTree` | 新增 | 连接下 schema 层级树；attach L1 右键 |
| `SidebarDndContext` | 新增 | `@dnd-kit` 改父级 + 同级排序 |
| `useSidebarFileDrop` | 新增 | `app:file-drop` 分流 attach / 新建连接 |
| `Menubar` | 新增 | Radix 基元 |
| `ContextMenu` | 新增 | Radix 基元 |
| `SqlExecutionHistoryDialog` | 新增 | 全局 SQL 历史 |
| `AboutDialog` | 新增 | 关于 |
| `AttachDatabaseDialog` | 新增 | SQLite 附加库 |
| `Header` | 变更 | 组合 MenuBar，移除设置按钮 |
| `AppShell` | 变更 | Dialog 状态 + 快捷键 |
| `ConnectionTree` | 变更 | `rootItems` 混排 + 根层空白右键新建分组 |
| `ConnectionGroupNode` | 新增 | 嵌套分组 + DnD |
| `ConnectionTreeItem` | 变更 | Group 下连接行 + ContextMenu + sortable |
| `ConnectionForm` | 新增 | 新建/编辑连接统一表单 |
| `NewConnectionDialog` | 变更 | 壳 + mode=create |
| `EditConnectionDialog` | 变更 | 壳 + mode=edit，共用 ConnectionForm |
| `RemoteConnectionForm` | 废弃 | 逻辑迁入 ConnectionForm |
| `SchemaSubtree` | **已删除** | 逻辑迁入 `ConnectionSchemaTree` |
| `SettingsDialog` | 变更 | 设置 Dialog 展示壳 |
| `SettingsDialogContainer` | 变更 | 设置 wrapper（config query）；**AppShell 挂载** |
