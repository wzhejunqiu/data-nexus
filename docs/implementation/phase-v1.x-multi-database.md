# v1.x — 多数据库（PostgreSQL / MySQL）

> **预估:** 按驱动分批，每个 7-14 天 · **前置:** v1.0

---

## 目标

在统一 Driver 抽象上接入 PostgreSQL、MySQL，连接 UI 按类型动态渲染。

---

## 架构任务

- [ ] `driver/postgres/` — `pgx` 或 `database/sql` + pg driver
- [ ] `driver/mysql/` — `go-sql-driver/mysql`
- [ ] `DriverFactory` — 按 `config.type` 创建实例
- [ ] Schema 差异吸收 — `TableInfo.Schema`、引号规则（见 [DATA_MODEL §6](../design/DATA_MODEL.md)）
- [ ] 连接测试 — `Ping` + 超时

---

## PostgreSQL（首个非 SQLite 驱动）

### 连接配置

```yaml
type: postgres
host, port, database, user, password, sslMode, schema
```

### 任务

- [ ] 连接表单 UI
- [ ] `ListTables` — `information_schema` 或 pg catalog
- [ ] `GetTableSchema` / `BrowseRows` / `Execute`
- [ ] SSL 模式支持
- [ ] 集成测试 — Docker postgres 或 testcontainers

---

## MySQL（第二个驱动）

- [ ] 同上模式实现
- [ ] `` `identifier` `` 引用
- [ ] Docker mysql 集成测试

---

## 可选：Headless 模式

- [ ] `--server` flag 启动 chi REST
- [ ] 映射 [API.md Headless 表](../design/API.md) 到同一 service 层
- [ ] 用于 CI / 脚本

---

## 平台能力（v2.0 候选）

- [ ] ER 图
- [ ] Basic Auth（内网共享）
- [ ] 插件扩展点

---

## 完成标准

- [ ] PostgreSQL 连接 → 浏览 → SQL 闭环
- [ ] MySQL 同上
- [ ] tag `v1.1.0` / `v1.2.0` 分批
