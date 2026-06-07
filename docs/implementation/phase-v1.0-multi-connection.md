# v1.0 — 多连接

> **预估:** 7-10 天 · **前置:** v0.3 · **参考:** [ROADMAP](../ROADMAP.md)

---

## 目标

同时管理多个数据库连接，Tab 化工作区。

---

## 功能清单

### 连接管理

- [ ] `ConnectionManager` 重构 — `map[id]*Connection` 多实例
- [ ] 持久化 — `~/.data-nexus/connections.json` 升级 schema
- [ ] 连接别名 `displayName`
- [ ] Sidebar 顶部连接切换 dropdown
- [ ] 「+ 新建连接」— 不关闭已有连接

### API 变更

- [ ] `ConnectionService.List` — 全部连接
- [ ] `ConnectionService.Activate(id)` — 切换活跃连接
- [ ] `ConnectionService.Remove(id)`
- [ ] Schema/Table/Query 方法增加 `connectionId` 或 implicit active

### UI

- [ ] 连接列表面板
- [ ] 多 Tab SQL 编辑器（每 Tab 绑定 connection + sql）
- [ ] 窗口标题反映当前连接

### 迁移

- [ ] v0.x 单连接配置自动迁移

---

## 非目标（v1.0）

- 多数据库类型（仍 SQLite only）
- Headless REST（可并行规划）

---

## 完成标准

- [ ] 同时保存 ≥3 个 SQLite 连接并切换
- [ ] tag `v1.0.0`

---

## 下一阶段

[phase-v1.x-multi-database.md](./phase-v1.x-multi-database.md)
