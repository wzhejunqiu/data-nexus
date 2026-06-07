# v0.5 — 桌面 UI 重构

> **预估:** 5–8 天 · **前置:** v0.4 · **下一版本:** [phase-v0.5.1.md](./phase-v0.5.1.md)

> **v0.5.x 总览:** v0.5.0 桌面 UI 重构 → **v0.5.1 快捷键** → v0.6 `--api` → …

实施时 **以本文档为主清单**。交互细节见 [UI_UX.md](../design/UI_UX.md)、[CONNECTION_UX.md](../design/CONNECTION_UX.md)。

---

## 1. 版本目标

v0.5 **仅交付 Wails 桌面端 UI 重构**，不引入 HTTP 模式（留 v0.6 / v0.7）。

| 主题 | 说明 |
|------|------|
| 应用内 MenuBar | 文件 / 视图 / 帮助；统一新建连接、设置、关于等入口 |
| 左连接列表 + 右展示区 | 移除侧边栏全局控件；连接项右键菜单 + 双击打开 |
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
│  ● app.db       ├────────────────────────────────────────────────┤
│    └ users      │              Main Content                        │
│  ○ staging.db   │                                                │
│  [SavedQueries] │                                                │
├─────────────────┴────────────────────────────────────────────────┤
│  Status Bar                                                      │
└──────────────────────────────────────────────────────────────────┘
```

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
| 右键 | 打开/关闭、重命名、编辑、删除 |

**迁出侧边栏：** 新建连接按钮、`readOnly`/`wal` → `NewConnectionDialog`；启动恢复连接 → `SettingsDialog`。

---

## 3. 功能清单

### 3.1 前端基元与 MenuBar

- [ ] `@radix-ui/react-menubar`、`@radix-ui/react-context-menu`
- [ ] `frontend/src/components/ui/Menubar.tsx`、`ContextMenu.tsx`
- [ ] `frontend/src/components/AppMenuBar.tsx`
- [ ] 重构 `Header.tsx`；`AppShell.tsx` Dialog 状态 + 快捷键

### 3.2 连接列表

- [ ] `ConnectionTree.tsx`：去顶部控件；ContextMenu；双击/单击
- [ ] `NewConnectionDialog`：内置 `readOnly` / `wal`
- [ ] `SettingsDialog`：迁入 `RestoreOnStartupToggle`
- [ ] 测试：`ConnectionTree`、`NewConnectionDialog`、`SettingsDialog`

### 3.3 Dialog

- [ ] `SqlExecutionHistoryDialog.tsx` + `listAllExecutions` API
- [ ] `AboutDialog.tsx`（i18n，替代原生 `ShowAbout`）
- [ ] i18n：`menu.*`、`sqlHistory.*`、`about.*`、`settings.restoreOnStartup`

### 3.4 后端（全局 SQL 历史）

- [ ] `ListAllExecutions` store / service / Wails 绑定
- [ ] 单元测试

### 3.5 Go 原生菜单

- [ ] `ApplicationMenu()` 仅 App / Edit / Window
- [ ] 文件/设置/关于改由前端 MenuBar；清理 `handleOpenDatabase` 等（前端接管）

---

## 4. 手动验收

- [ ] MenuBar 各菜单可用；无侧边栏/Header 重复入口
- [ ] 双击/右键连接操作正常
- [ ] 新建连接 Dialog：只读/WAL；设置：启动恢复
- [ ] 视图 → SQL 执行历史、帮助 → 关于
- [ ] macOS App 菜单 Quit；无重复 File/View/Help

---

## 5. 非目标

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
| [UI_UX.md](../design/UI_UX.md) | MenuBar、两栏布局 |
| [CONNECTION_UX.md §5](../design/CONNECTION_UX.md) | 右键菜单 |
| [API.md §14](../design/API.md#14-sqlexecutionservicev03) | `ListAllSqlExecutions` |
| [phase-v0.5.1.md](./phase-v0.5.1.md) | 下一版本：快捷键与 macOS ⌘, |
| [phase-v0.4-multi-connection.md](./phase-v0.4-multi-connection.md) | 上一版本 |
