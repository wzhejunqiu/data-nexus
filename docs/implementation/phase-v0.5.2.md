# v0.5.2 — PostgreSQL 多库树浏览

> **预估:** 3–5 天 · **前置:** [v0.5](./phase-v0.5.md) / [v0.5.1](./phase-v0.5.1.md) · **下一版本:** [v0.6](./phase-v0.6.md)

v0.5 连接树已支持 MySQL 实例级连接后在树中浏览多个 database。PostgreSQL 因协议限制，当前 Driver 仅对**当前会话连接的 database** 执行 `ListSchemas` / `ListTables`。本版本补齐 PG 多库浏览，并将 PG 连接表单的 database 字段改为可选（与 MySQL 对齐）。

---

## 1. 版本目标

| 主题 | 说明 |
|------|------|
| PG 按库重连 | 展开非当前会话 database 时，Driver 懒重连（或 per-database 连接池） |
| 树完整展开 | 任意可访问 PG database → schema → 表/视图 |
| 表单对齐 | PG 新建/编辑连接：database 改为**可选**（留空默认连 `postgres` 维护库） |
| SQL 上下文 | 树节点选择与 SQL 执行 catalog/schema 上下文一致 |

发布：**tag `v0.5.2`**

---

## 2. 背景：当前限制

PostgreSQL 无 `USE database`；`internal/driver/postgres/namespaces.go` 中 `ListSchemas` 在 `database != d.database` 时返回空。连接表单虽可列出全部 database（`ListNamespaces`），但展开其他库无法加载 schema/表。

MySQL 已在 v0.5 补丁中支持 database 可选 + `information_schema` 跨库查表。

---

## 3. 实施方案（草案）

### 3.1 PG Driver

- `Connect` / 内部会话：database 为空时默认 `postgres`
- `ListSchemas(connectionId, database)` / `ListTables(..., database)`：若 `database != 当前会话库`，对该库建立临时连接或从池中取用
- `GetTableSchema`、查询执行：使用与树节点一致的 database 上下文
- 连接关闭时释放附加会话

### 3.2 后端校验

- `validateRemoteConnectRequest`：删除 PostgreSQL `database is required`
- `PostgresConfig.NormalizedDatabase()`：空 → `postgres`

### 3.3 前端

- [`PostgresFields.tsx`](../frontend/src/features/connection/ConnectionForm/fields/PostgresFields.tsx)：database 改为可选（与 MySQL 相同 UI）
- [`connectionFormValidation.ts`](../frontend/src/features/connection/ConnectionForm/connectionFormValidation.ts)：PG 不再校验 database 非空
- [`connectionDisplay.ts`](../frontend/src/lib/connectionDisplay.ts)：PG database 为空时副标题省略 `/database`

### 3.4 文档

- 更新 [CONNECTION_FORM.md](./v0.5/CONNECTION_FORM.md)、[CONNECTION_TREE.md](./v0.5/CONNECTION_TREE.md)、[CONNECTION_UX.md](../design/CONNECTION_UX.md)

---

## 4. 验收要点

- [ ] PG 连接树：展开任意可访问 database → 可见 schema → 表/视图
- [ ] 点击表名后右侧 Tab 数据正确（catalog/schema 与树节点一致）
- [ ] PG 新建连接可不填 database，仍能连接并在树中浏览
- [ ] MySQL 行为不退化（database 仍可选）
- [ ] `make test` + `make test-integration` + 前端测试通过

---

## 5. 相关文档

| 文档 | 用途 |
|------|------|
| [CONNECTION_TREE.md](./v0.5/CONNECTION_TREE.md) | 层级树设计 |
| [CONNECTION_FORM.md](./v0.5/CONNECTION_FORM.md) | 连接表单字段 |
| [phase-v0.5.1.md](./phase-v0.5.1.md) | 上一版本 |
