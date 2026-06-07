# v0.3 — 探索型差异化

> **预估:** 5-8 天 · **前置:** v0.2 · **参考:** [COMPETITIVE_ANALYSIS §4.3](../product/COMPETITIVE_ANALYSIS.md)

---

## 目标

列画像、Filter→SQL 等探索能力，形成与 Tabulita/DB4S 的差异化。

---

## 功能清单

### 列数据画像

- [x] `SchemaService.GetTableProfile(tableName)` — 采样统计
- [x] 每列：distinct 数、NULL 占比、min/max（数值/日期）、top-N 值（低基数）
- [x] 大表采样 — `LIMIT 10000` 或 TABLESAMPLE 策略
- [x] UI：表结构 Tab 下方「数据画像」区块

### Filter Builder → SQL

- [x] 数据 Tab 可视化筛选 UI（列 / 操作符 / 值）
- [x] 生成 WHERE 子句，可「在 SQL 编辑器中编辑」
- [x] 后端 `BrowseRows` 扩展 `filters[]` 或前端生成 SQL

### 保存与分享

- [x] Canned Queries — 命名 SQL，侧边栏快捷入口，存 `~/.data-nexus/queries.json`
- [x] 应用内视图状态 — 当前表/筛选/排序序列化（localStorage 或配置文件）

### FTS & Facet

- [x] 检测 FTS 虚拟表 — `sqlite_master` + 命名规则
- [x] 表页搜索框 — `?_search=` 等价逻辑
- [x] Facet 分面 — 低基数列可选开启（自动 suggest）

### SQLite ATTACH

- [x] `ConnectionService.Attach(filePath, alias)`
- [x] Sidebar 按 attached 库分组
- [x] `Detach(alias)`

---

## 后端 API 扩展

| 新方法 | 说明 |
|--------|------|
| `GetTableProfile` | 列级统计 |
| `BrowseRows` + filters | 结构化筛选 |
| `Attach` / `Detach` | 多 db 文件 |
| `ListCannedQueries` / `SaveCannedQuery` | 命名查询 |

详见实施时更新 [API.md](../design/API.md)。

---

## 前端任务

- [x] `ColumnProfile` 组件
- [x] `FilterBuilder` 组件
- [x] `SavedQueries` 侧边栏区块
- [x] FTS 搜索框（条件渲染）

---

## 完成标准

- [x] 选表后可看列画像（无需写 SQL）
- [x] Filter UI 生成可执行 SQL
- [ ] tag `v0.3.0`（暂缓）

---

## 下一阶段

[phase-v1.0-multi-connection.md](./phase-v1.0-multi-connection.md)
