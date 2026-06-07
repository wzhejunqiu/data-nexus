# Phase 2 — 前端骨架 + 核心 UI

> **预估:** 3-4 天 · **状态:** 已完成 · **前置:** [phase-1-wails-backend.md](./phase-1-wails-backend.md) · **下一阶段:** [phase-3-feature-integration.md](./phase-3-feature-integration.md)

---

## 目标

完整 React 桌面 UI 骨架；连接流程与 Schema 侧边栏可用。

## 里程碑

桌面窗口内打开 `.db`，侧边栏展示表/视图列表。

---

## 2.1 前端工程

**参考:** [UI_UX.md](../design/UI_UX.md) · [ARCHITECTURE §3.3](../design/ARCHITECTURE.md)

### 任务清单

- [x] Tailwind CSS 配置
- [x] shadcn/ui 初始化（Button、Input、Tabs、Toast、Dialog、Table 等）
- [x] ESLint + Prettier
- [x] **i18n**：`react-i18next` + `i18next`，`locales/zh-CN` + `locales/en`，默认 zh-CN
- [x] `LanguageToggle` — Header 语言切换，持久化 `localStorage`
- [x] 所有 UI 文案走 `t('key')`，禁止硬编码中文/英文
- [x] 字体：Inter（UI）+ JetBrains Mono（代码区预留）
- [x] TanStack Query Provider
- [x] Zustand：`connectionStore`（connected、connection、readOnly）
- [x] 主题：`light` / `dark` / `system`（CSS variables 见 UI_UX §2.2）

### API 封装层

```
frontend/src/lib/
├── api/
│   ├── connection.ts
│   ├── saved-connection.ts
│   ├── schema.ts
│   ├── dialog.ts
│   ├── app.ts
│   └── errors.ts      # mapWailsError
└── types/             # 与 Go model 对齐
```

- [x] **禁止**组件直接 import `wailsjs`，统一走 `lib/api/`
- [x] `mapWailsError` 解析 `AppError` code/message

---

## 2.2 布局与连接

**参考:** [UI_UX §3-4](../design/UI_UX.md)

### 组件

## 2.1 连接 UI（Navicat 模式）

**参考:** [CONNECTION_UX.md](../design/CONNECTION_UX.md) · [UI_UX §4.1](../design/UI_UX.md)

```
frontend/src/
├── app/
│   └── AppShell.tsx
├── features/connection/
│   ├── ConnectionTree.tsx
│   ├── ConnectionTreeItem.tsx
│   ├── NewConnectionDialog.tsx
│   └── useConnections.ts       # TanStack Query
└── components/
    ├── Header.tsx
    ├── StatusBar.tsx
    └── ThemeToggle.tsx
```

### 任务清单

- [x] `AppShell` — 启动即主工作台（**无 WelcomePage**）
- [x] `ConnectionTree` — `ListConnections`；展示 open/closed 状态
- [x] `NewConnectionDialog` → `OpenConnectionFromFile`（文件对话框 + 只读）
- [x] 节点「打开」→ `OpenConnection(id)`；「关闭」→ `CloseConnection(id)`
- [x] 删除 → `RemoveConnection`（已打开需确认）
- [x] 已打开连接下嵌套 Schema 子树（见 §2.3）
- [x] 菜单 **File → Open** → 新建并打开连接
- [x] **P1:** `SetRestoreOpenOnStartup` — 启动恢复上次已打开连接
- [x] `ThemeToggle` + 跟随系统
- [x] Loading / 错误 inline + Toast

### 验收

- [x] 启动即见连接列表（可为空）
- [x] 新建连接 → 打开 → Schema 展开
- [x] 同时打开 ≥2 个 SQLite，互不影响
- [x] 关闭单个连接后列表条目仍在

---

## 2.3 Schema 侧边栏（嵌入连接树）

```
frontend/src/features/schema/
├── SchemaSubtree.tsx      # 某 connectionId 下的表/视图
└── useTables.ts           # TanStack Query，key 含 connectionId
```

### 任务清单

- [x] `useTables(connectionId)` — 仅对已 `open` 连接 fetch
- [x] 分组：**TABLES** / **VIEWS**
- [x] 搜索框 — 前端 filter（按当前连接）
- [x] 点击表名 — 写入 store（`activeConnectionId` + `selectedTable`）
- [x] 空库 Empty State
- [x] 加载 skeleton

### 主内容区 Tab 栏（骨架）

- [x] Tabs：`表结构` | `数据` | `SQL`（Phase 3 填内容）
- [x] 未选表时 placeholder 文案

---

## 完成标准

- [x] `wails dev` 完整连接流程无阻断
- [x] 侧边栏表列表正确（含 views）
- [x] 主题切换正常
- [x] 进入 [phase-3-feature-integration.md](./phase-3-feature-integration.md)
