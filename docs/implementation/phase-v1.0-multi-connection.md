# v1.0 — 远程数据库（PostgreSQL / MySQL）

> **预估:** 10-14 天 · **前置:** v0.3 · **参考:** [ROADMAP](../ROADMAP.md)

---

## 目标

在 MVP 已具备的 **Navicat 式多 SQLite 连接** 基础上，扩展 PostgreSQL / MySQL 驱动与连接表单。

> **说明:** 多连接并存、连接树 UI、`connectionId` 路由已在 MVP（v0.1）实现；本版本不再重复「多连接」主题。

---

## 功能清单

### Driver 层

- [ ] `PostgresDriver` — `pgx` 或 `database/sql` + `pgx/stdlib`
- [ ] `MySQLDriver` — `go-sql-driver/mysql`
- [ ] `OpenTableExport` — 各方言 `TableExportCursor`（PG/MySQL 见 [EXPORT_MULTI_DIALECT.md](../design/EXPORT_MULTI_DIALECT.md)）
- [ ] Driver 工厂按 `DriverType` 分发

### 连接配置

- [ ] `NewConnectionDialog` 支持类型切换（SQLite / PostgreSQL / MySQL）
- [ ] 远程连接表单：host、port、database、user、password、SSL
- [ ] 密码加密存储（`connections.json` 或 keychain，待定）
- [ ] 连接测试（Ping）按钮

### Schema 差异

- [ ] PostgreSQL schema 列表（`public` 等）
- [ ] MySQL database 切换
- [ ] 类型映射扩展（见 [DATA_MODEL.md](../design/DATA_MODEL.md)）

### UI

- [ ] 连接树图标区分数据库类型
- [ ] 远程连接错误提示（网络、认证失败）

---

## 非目标（v1.0）

- Headless REST（可并行规划 v1.x）
- Redis / Mongo 等非 SQL 引擎

---

## 完成标准

- [ ] 同时打开 SQLite + PostgreSQL（或 MySQL）各至少 1 个
- [ ] Schema 浏览与 SQL 执行在两种远程库上可用
- [ ] tag `v1.0.0`

---

## 下一阶段

[phase-v1.x-multi-database.md](./phase-v1.x-multi-database.md)
