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
- [ ] 字体：Inter（UI）+ JetBrains Mono（代码区预留）
- [ ] TanStack Query Provider
- [ ] Zustand：`connectionStore`（connected、connection、readOnly）
- [ ] 主题：`light` / `dark` / `system`（CSS variables 见 UI_UX §2.2）

### API 封装层

```
frontend/src/lib/
├── api/
│   ├── connection.ts
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

```
frontend/src/
├── app/
│   └── AppShell.tsx
├── features/connection/
│   ├── WelcomePage.tsx
│   ├── ConnectionForm.tsx
│   ├── OpenDatabaseButton.tsx
│   └── ConnectionStatus.tsx
└── components/
    ├── Header.tsx
    ├── StatusBar.tsx
    └── ThemeToggle.tsx
```

### 任务清单

- [ ] `AppShell` — Header + Sidebar + Main + StatusBar
- [ ] 未连接 → `WelcomePage`；已连接 → 工作台布局
- [ ] `OpenDatabaseButton` → `DialogService.OpenDatabaseFile()` → `Connect`
- [ ] `ConnectionForm` — 路径输入 + 只读勾选 + 连接按钮
- [ ] 菜单 **File → Open** 触发同上流程
- [ ] `ConnectionStatus` — 文件名、路径 truncate、断开按钮
- [ ] 断开需确认；断开后清空 Sidebar
- [ ] `ThemeToggle` + 跟随系统
- [ ] Loading / 错误 inline + Toast

### 验收

- [ ] 文件对话框选库 → 连接成功 → Header 显示状态
- [ ] 无效路径 → 明确错误
- [ ] 只读模式勾选 → 后端 `readOnly: true`

---

## 2.3 Schema 侧边栏

```
frontend/src/features/schema/
├── SchemaSidebar.tsx
└── useTables.ts          # TanStack Query
```

### 任务清单

- [ ] `useTables` — `SchemaService.ListTables`，连接后自动 fetch
- [ ] 分组：**TABLES** / **VIEWS**
- [ ] 搜索框 — 前端 filter
- [ ] 点击表名 — 选中高亮，写入 store（`selectedTable`）
- [ ] 空库 Empty State
- [ ] Sidebar skeleton loading

### 主内容区 Tab 栏（骨架）

- [ ] Tabs：`表结构` | `数据` | `SQL`（Phase 3 填内容）
- [ ] 未选表时 placeholder 文案

---

## 完成标准

- [ ] `wails dev` 完整连接流程无阻断
- [ ] 侧边栏表列表正确（含 views）
- [ ] 主题切换正常
- [ ] 进入 [phase-3-feature-integration.md](./phase-3-feature-integration.md)
