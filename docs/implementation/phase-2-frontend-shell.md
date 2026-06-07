# Phase 2 — 前端骨架 + 核心 UI

> **预估:** 3-4 天 · **前置:** [phase-1-wails-backend.md](./phase-1-wails-backend.md) · **下一阶段:** [phase-3-feature-integration.md](./phase-3-feature-integration.md)

---

## 目标

完整 React 桌面 UI 骨架；连接流程与 Schema 侧边栏可用。

## 里程碑

桌面窗口内打开 `.db`，侧边栏展示表/视图列表。

---

## 2.1 前端工程

**参考:** [UI_UX.md](../design/UI_UX.md) · [ARCHITECTURE §3.3](../design/ARCHITECTURE.md)

### 任务清单

- [ ] Tailwind CSS 配置
- [ ] shadcn/ui 初始化（Button、Input、Tabs、Toast、Dialog、Table 等）
- [ ] ESLint + Prettier
- [ ] **i18n**：`react-i18next` + `i18next`，`locales/zh-CN` + `locales/en`，默认 zh-CN
- [ ] `LanguageToggle` — Header 语言切换，持久化 `localStorage`
- [ ] 所有 UI 文案走 `t('key')`，禁止硬编码中文/英文
- [ ] 字体：Inter（UI）+ JetBrains Mono（代码区预留）
- [ ] TanStack Query Provider
- [ ] Zustand：`connectionStore`（connected、connection、readOnly）
- [ ] 主题：`light` / `dark` / `system`（CSS variables 见 UI_UX §2.2）

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

- [ ] **禁止**组件直接 import `wailsjs`，统一走 `lib/api/`
- [ ] `mapWailsError` 解析 `AppError` code/message

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

- [ ] `AppShell` — 启动即主工作台（**无 WelcomePage**）
- [ ] `ConnectionTree` — `ListConnections`；展示 open/closed 状态
- [ ] `NewConnectionDialog` → `OpenConnectionFromFile`（文件对话框 + 只读）
- [ ] 节点「打开」→ `OpenConnection(id)`；「关闭」→ `CloseConnection(id)`
- [ ] 删除 → `RemoveConnection`（已打开需确认）
- [ ] 已打开连接下嵌套 Schema 子树（见 §2.3）
- [ ] 菜单 **File → Open** → 新建并打开连接
- [ ] **P1:** `SetRestoreOpenOnStartup` — 启动恢复上次已打开连接
- [ ] `ThemeToggle` + 跟随系统
- [ ] Loading / 错误 inline + Toast

### 验收

- [ ] 启动即见连接列表（可为空）
- [ ] 新建连接 → 打开 → Schema 展开
- [ ] 同时打开 ≥2 个 SQLite，互不影响
- [ ] 关闭单个连接后列表条目仍在

---

## 2.3 Schema 侧边栏（嵌入连接树）

```
frontend/src/features/schema/
├── SchemaSubtree.tsx      # 某 connectionId 下的表/视图
└── useTables.ts           # TanStack Query，key 含 connectionId
```

### 任务清单

- [ ] `useTables(connectionId)` — 仅对已 `open` 连接 fetch
- [ ] 分组：**TABLES** / **VIEWS**
- [ ] 搜索框 — 前端 filter（按当前连接）
- [ ] 点击表名 — 写入 store（`activeConnectionId` + `selectedTable`）
- [ ] 空库 Empty State
- [ ] 加载 skeleton

### 主内容区 Tab 栏（骨架）

- [ ] Tabs：`表结构` | `数据` | `SQL`（Phase 3 填内容）
- [ ] 未选表时 placeholder 文案

---

## 完成标准

- [ ] `wails dev` 完整连接流程无阻断
- [ ] 侧边栏表列表正确（含 views）
- [ ] 主题切换正常
- [ ] 进入 [phase-3-feature-integration.md](./phase-3-feature-integration.md)
