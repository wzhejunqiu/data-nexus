# Phase 3 — 功能集成（MVP 闭环）

> **预估:** 4-6 天 · **前置:** [phase-2-frontend-shell.md](./phase-2-frontend-shell.md) · **下一阶段:** [phase-4-release.md](./phase-4-release.md)

---

## 目标

完成 MVP 全部 P0 + P1 功能，前后端联调通过。

## 里程碑

连接 → 浏览表结构/数据 → 执行 SQL（含写操作确认）全流程可用。

---

## 3.1 表结构 & 数据浏览

**参考:** [PRD §2.1](../product/PRD.md) · [UI_UX §4.4-4.5](../design/UI_UX.md) · [API §4-5](../design/API.md)

### 组件

```
frontend/src/features/
├── schema/
│   ├── SchemaTable.tsx      # 列定义
│   └── IndexList.tsx        # P1
└── table-browser/
    ├── DataGrid.tsx
    ├── Pagination.tsx
    └── useTableRows.ts
```

### 表结构 Tab

- [ ] `SchemaTable` — 列名、类型、PK、nullable、default
- [ ] `IndexList`（P1）— 索引名、列、unique
- [ ] 复制表名/列名（P1）— 按钮或右键

### 数据 Tab

- [ ] `BrowseRows` — page / pageSize(25|50|100|200) / sort / order
- [ ] 列头点击排序（三态：无/asc/desc）
- [ ] `Pagination` — 首页/末页/页码/总行数
- [ ] NULL — 灰色斜体 `NULL`
- [ ] BLOB — `[BLOB N bytes]`
- [ ] Loading skeleton / Empty / Error
- [ ] P2：NULL 高亮强化

### 后端（若 Phase 1 未做全）

- [ ] `GetTableSchema` 返回 indexes
- [ ] `BrowseRows` totalRows COUNT
- [ ] 表名/列名校验错误 → `TABLE_NOT_FOUND`

---

## 3.2 SQL 工作台

**参考:** [UI_UX §4.6](../design/UI_UX.md) · [API §6](../design/API.md)

### 组件

```
frontend/src/features/sql-editor/
├── SqlEditor.tsx           # Monaco
├── QueryResultPanel.tsx
├── QueryHistory.tsx        # P1，会话内 max 50
└── ConfirmDialog.tsx       # 写操作确认
```

### 任务清单

- [ ] Monaco Editor — 最小 8 行，最高 20 行 auto-grow
- [ ] `Cmd/Ctrl + Enter` 执行
- [ ] `QueryService.Execute` — SELECT 结果表格 + rowCount + durationMs
- [ ] INSERT/UPDATE/DELETE — `ConfirmDialog` 二次确认
- [ ] 只读模式 — 写操作后端 `READ_ONLY` + 前端禁用提示
- [ ] SQL 错误 — Alert 展示数据库错误原文
- [ ] `QueryHistory`（P1）— 下拉填充 SQL
- [ ] 结果集 CSV 复制（P1）—  clipboard
- [ ] PRAGMA 快捷入口（P1）— 侧边或 SQL Tab 工具栏

### 写操作检测

前端或后端：首关键字非 SELECT/WITH/PRAGMA/EXPLAIN → 需确认

---

## 3.3 联调与 Polish

### 错误场景

- [ ] 文件不存在 → `CONNECTION_FAILED`
- [ ] 数据库锁定 → `DATABASE_LOCKED`
- [ ] 未连接调 Schema → `NOT_CONNECTED`
- [ ] 对话框取消 → 静默，不 Toast
- [ ] SQL 语法错误 → `SQL_ERROR`
- [ ] 结果超 maxRows → `RESULT_TOO_LARGE`

### 桌面体验

- [ ] StatusBar — 行数、耗时、版本（`AppService.GetVersion`）
- [ ] 窗口标题 — `Data Nexus — {filename}`
- [ ] CLI `--db /path/to.db` 启动时自动 Connect
- [ ] `wails build` 本地构建通过
- [ ] release 模式日志写文件（非控制台）

### 性能抽查

- [ ] 10 万行表首页 pageSize=50 < 500ms
- [ ] 1 万行 SELECT < 200ms

---

## MVP 功能对照（实施勾选）

| 优先级 | 功能 | Phase 3 验收 |
|--------|------|--------------|
| P0 | 连接 + 断开 | ✓ |
| P0 | 只读模式 | ✓ |
| P0 | Schema 列表 + 表结构 | ✓ |
| P0 | 分页数据 + 排序 | ✓ |
| P0 | SQL SELECT | ✓ |
| P0 | SQL 写 + 确认 | ✓ |
| P1 | 索引 | ✓ |
| P1 | 查询历史 | ✓ |
| P1 | 主题 | ✓ |
| P1 | PRAGMA | ✓ |
| P1 | CSV 复制 | ✓ |
| P1 | 复制表名/列名 | ✓ |

---

## 完成标准

- [ ] 上述 MVP 对照表全部勾选
- [ ] `wails dev` 全流程手动走通
- [ ] 进入 [phase-4-release.md](./phase-4-release.md)
