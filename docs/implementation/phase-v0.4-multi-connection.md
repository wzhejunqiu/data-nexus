# v0.4 — 远程数据库（PostgreSQL / MySQL）

> **预估:** 10-14 天 · **前置:** v0.3 · **参考:** [ROADMAP](../ROADMAP.md)

---

## 目标

在 MVP 已具备的 **Navicat 式多 SQLite 连接** 基础上，扩展 PostgreSQL / MySQL 驱动与连接表单。

> **说明:** 多连接并存、连接树 UI、`connectionId` 路由已在 MVP（v0.1）实现；本版本不再重复「多连接」主题。

---

## 功能清单

### Driver 层

- [x] `PostgresDriver` — `database/sql` + `pgx/stdlib`
- [x] `MySQLDriver` — `go-sql-driver/mysql`
- [x] `OpenTableExport` — PG/MySQL `TableExportCursor`（streaming `SELECT *`，行序不保证）
- [x] Driver 工厂按 `DriverType` 分发

### 连接配置

- [x] `NewConnectionDialog` 支持类型切换（SQLite / PostgreSQL / MySQL）
- [x] 远程连接表单：host、port、database、user、password、SSL
- [x] 密码存储：**Keychain 优先**，不可用时 **Vault fallback**（RSA/AES + 用户主密码），见 [SECRETS.md](../design/SECRETS.md)
- [x] 连接测试（Ping）按钮

### Schema 差异

- [x] PostgreSQL schema 切换（`UpdateConnectionPostgresSettings` + 侧边栏）
- [x] MySQL database 切换
- [x] 类型映射扩展（见 [DATA_MODEL.md](../design/DATA_MODEL.md)）

### UI

- [x] 连接树图标/标签区分数据库类型（SQL / PG / MY）
- [x] 远程连接错误提示（网络、认证失败）
- [x] `VaultDialog` 按需解锁（非启动弹窗）

---

## 非目标（v0.4）

- Headless REST（见 [phase-v0.6.md](./phase-v0.6.md)）；浏览器 Server（见 [phase-v0.7.md](./phase-v0.7.md)）
- Redis / Mongo 等非 SQL 引擎

---

## 完成标准

- [x] 同时打开 SQLite + PostgreSQL（或 MySQL）各至少 1 个（见 [phase-v0.4-manual-checklist.md](./phase-v0.4-manual-checklist.md)）
- [x] Schema 浏览与 SQL 执行在两种远程库上可用
- [ ] tag `v0.4.0`（待手动发布）

---

## 下一阶段

[phase-v0.5.md](./phase-v0.5.md) → … → [phase-v0.8.md](./phase-v0.8.md)
