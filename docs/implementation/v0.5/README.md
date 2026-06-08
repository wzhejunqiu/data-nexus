# v0.5 — 桌面 UI 重构 · 文档索引

> **版本:** v0.5.0 · **预估:** 8–11 天 · **前置:** [v0.4](../phase-v0.4-multi-connection.md) · **下一版本:** [v0.5.1](../phase-v0.5.1.md)

本目录为 **v0.5 专用实施包**：主清单仍以 [phase-v0.5.md](../phase-v0.5.md) 为准；此处提供可执行的实施方案、设计细节与验收标准。

---

## 文档结构

| 文档 | 用途 | 实施时优先阅读 |
|------|------|----------------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | **主实施方案** — 任务分解、实施顺序、文件清单、工期 | ⭐ 开工必读 |
| [CONNECTION_GROUPS_AND_STORAGE.md](./CONNECTION_GROUPS_AND_STORAGE.md) | **Group 嵌套 + catalog.db 持久化** | 后端 + 连接树 |
| [CONNECTION_TREE.md](./CONNECTION_TREE.md) | 连接下 database/schema 层级树 | 连接树开发 |
| [CONNECTION_FORM.md](./CONNECTION_FORM.md) | **新建/编辑连接表单** — 方言配置、UI 线框、模型扩展 | 连接表单开发 |
| [DESIGN.md](./DESIGN.md) | v0.5 UI/UX 设计规范（布局、MenuBar、连接列表、Dialog） | 前端开发 |
| [BACKEND.md](./BACKEND.md) | 后端变更：`ListAllExecutions`、Go 原生菜单精简 | 后端开发 |
| [FRONTEND.md](./FRONTEND.md) | 前端组件实施细节、状态迁移、测试要点 | 前端开发 |
| [manual-checklist.md](./manual-checklist.md) | 手动验收清单 | 发布前 |

---

## 版本目标（摘要）

v0.5 **仅交付 Wails 桌面端 UI 重构**，不引入 HTTP 模式（留 v0.6 / v0.7）。

| 主题 | 说明 |
|------|------|
| 应用内 MenuBar | **Win/Linux：** 顶栏 Menubar；**macOS：** 系统菜单栏（HIG），应用内仅状态行 |
| 连接分组 + 游离连接 | 顶层 Group 与游离连接同级；catalog.db；**DnD 改父级/同级排序**；**文件 → 新建分组** 或 **根层空白右键** |
| SQLite Attach 完善 | 拖 `.db` attach、L1 Detach/Finder、`SchemaSubtree` 已删 |
| 左连接列表 + 右展示区 | Group 树 + 连接下 schema 层级；Group inline 重命名、右键、拖拽 |
| 新建/编辑连接 Dialog | 浮动 Modal 重构；各库 charset、TLS、存储引擎等方言配置 |
| 全局 SQL 历史 | Menu → 视图（macOS 系统菜单 / Win·Linux 应用内） |
| Go 原生菜单 | **macOS：** App/文件/编辑/视图/窗口/帮助；**Win/Linux：** App/Edit/Window |

**发布 tag:** `v0.5.0`

---

## 上游设计文档

| 文档 | 关系 |
|------|------|
| [UI_UX.md](../../design/UI_UX.md) | 全局 UI 规范（v0.5 章节已对齐） |
| [CONNECTION_UX.md](../../design/CONNECTION_UX.md) | 连接交互（§5 右键菜单） |
| [API.md §14](../../design/API.md#14-sqlexecutionservicev03) | `ListAllSqlExecutions` 契约 |
| [DATA_MODEL.md §9](../../design/DATA_MODEL.md) | SQL 执行历史数据模型 |

---

## 推荐实施顺序

```
Phase 0  Day 1–3   catalog.db + Group 服务（无 json 迁移）
Phase A  Day 2–3   后端 ListAllExecutions + Go 菜单精简
Phase B  Day 2–3   Radix + AppMenuBar
Phase C  Day 3–5   ConnectionGroup 树 UI
Phase D  Day 5–7   ConnectionSchemaTree + ListNamespaces + Attach UI
Phase E  Day 6–8   ConnectionForm 重构（可与 D 部分并行）
Phase F  Day 8–9   SqlHistory + About Dialog
Phase G  Day 9–10  测试 + 验收
```

详细任务见 [IMPLEMENTATION.md](./IMPLEMENTATION.md)。

---

## 非目标

- **`connections.json` → `catalog.db` 数据迁移**（v0.5 直接使用新存储；旧 json 不自动导入）
- `--server` / `--api`（[v0.6](../phase-v0.6.md) / [v0.7](../phase-v0.7.md)）
- 快捷键可配置（[v0.5.1](../phase-v0.5.1.md)）
- ER 图、插件、驱动深化（[v0.8](../phase-v0.8.md)）
