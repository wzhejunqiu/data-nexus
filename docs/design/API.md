# Data Nexus — Service 绑定 API

> 版本: v0.3 · 通信方式: Wails in-memory Bridge · 最后更新: 2026-06-07

前端通过 `wailsjs/go/wails/*` 自动生成的绑定调用 Go Service。所有方法均为 **async**（返回 Promise）。时间戳使用 ISO 8601 UTC。

> **历史说明：** v0.1 曾设计 REST API（`localhost:8080`）。自 v0.2 起 MVP 改为 Wails 桌面绑定；方法语义与数据模型保持一致。未来 headless 模式（`--server`）可映射为 REST，见 [ARCHITECTURE.md](./ARCHITECTURE.md#23-可选扩展headless-模式v1x非-mvp)。

---

## 1. 通用约定

### 1.1 成功响应

Service 方法成功时直接返回领域对象（Go struct → JSON → TypeScript）。

### 1.2 错误响应

Go 方法返回 `*model.AppError`（实现 `error` 接口），Wails 将其抛至前端：

```typescript
interface AppError {
  code: string
  message: string
  details?: Record<string, unknown>
}
```

前端统一通过 `mapWailsError(err)` 解析。

### 1.3 错误码

| Code | 说明 |
|------|------|
| `INVALID_REQUEST` | 通用参数校验失败 |
| `INVALID_PATH` | 文件路径无效 |
| `NOT_CONNECTED` | 已废弃 → 使用 `CONNECTION_NOT_FOUND` |
| `CONNECTION_NOT_FOUND` | 连接 ID 不存在或未打开 |
| `CONNECTION_ALREADY_OPEN` | 连接已处于打开状态 |
| `CONNECTION_FAILED` | 无法打开数据库 |
| `DATABASE_LOCKED` | SQLite 数据库被锁定 |
| `TABLE_NOT_FOUND` | 表/视图不存在 |
| `SQL_ERROR` | SQL 执行失败 |
| `RESULT_TOO_LARGE` | 结果集超过行数上限 |
| `READ_ONLY` | 只读模式下拒绝写操作 |
| `DIALOG_CANCELLED` | 用户取消文件对话框 |
| `SAVED_NOT_FOUND` | 已保存连接 ID 不存在 |
| `INTERNAL_ERROR` | 未预期错误 |

---

## 2. DialogService

原生系统对话框，Wails `runtime` 封装。

### 2.1 OpenDatabaseFile

```go
func (s *DialogService) OpenDatabaseFile() (string, error)
```

**返回:** 用户选择的绝对路径；取消时返回 `DIALOG_CANCELLED`。

**Filters:** `*.db`, `*.sqlite`, `*.sqlite3`

### 2.2 SaveFile（v0.2）

```go
func (s *DialogService) SaveFile(defaultName string, filters []FileFilter) (string, error)
```

用于 CSV 导出等场景。

---

## 3. ConnectionService

管理**多条并存**的活跃连接 + 与 `connections.json` 联动。详见 [CONNECTION_UX.md](./CONNECTION_UX.md)。

### 3.1 ListConnections

```go
func (s *ConnectionService) ListConnections() (*ConnectionListView, error)
```

返回全部**已保存**连接，并标注是否已打开。

```json
{
  "items": [
    {
      "id": "01HX...",
      "name": "app.db",
      "type": "sqlite",
      "config": { "type": "sqlite", "sqlite": { "filePath": "...", "readOnly": false } },
      "status": "open",
      "connectedAt": "2026-06-07T08:00:00Z",
      "lastUsedAt": "2026-06-07T08:00:00Z"
    },
    {
      "id": "01HY...",
      "name": "staging.db",
      "status": "closed",
      "lastUsedAt": "2026-06-06T12:00:00Z"
    }
  ]
}
```

`status`: `open` | `closed`

### 3.2 CreateConnection

```go
func (s *ConnectionService) CreateConnection(req ConnectRequest) (*SavedConnection, error)
```

新建连接配置并 **upsert** 到 `connections.json`；**不自动打开**。用户需再调 `OpenConnection`。

### 3.3 OpenConnection

```go
func (s *ConnectionService) OpenConnection(connectionId string) (*Connection, error)
```

打开已保存连接；若已 `open` → `CONNECTION_ALREADY_OPEN`。成功后该连接加入活跃 map，其他连接保持原状。

### 3.4 OpenConnectionFromFile

```go
func (s *ConnectionService) OpenConnectionFromFile(req ConnectRequest) (*Connection, error)
```

`CreateConnection` + `OpenConnection` 组合；新建并立即打开（「新建连接」主路径）。

### 3.5 CloseConnection

```go
func (s *ConnectionService) CloseConnection(connectionId string) error
```

关闭指定连接，释放 Driver；**保留** saved 配置。

**Errors:** `CONNECTION_NOT_FOUND`（未打开时）

### 3.6 RemoveConnection

```go
func (s *ConnectionService) RemoveConnection(connectionId string) error
```

若仍打开则先 `CloseConnection`，再从 `connections.json` 删除。

### 3.7 RenameConnection（P1）

```go
func (s *ConnectionService) RenameConnection(connectionId string, name string) (*SavedConnection, error)
```

### 3.8 SetRestoreOpenOnStartup（P1）

```go
func (s *ConnectionService) SetRestoreOpenOnStartup(enabled bool) error
```

退出时持久化 `openConnectionIds`；下次启动自动 `OpenConnection`（默认 `false`）。

---

## 4. SavedConnectionService（可选合并）

> **实施建议:** MVP 可将 §4 合并进 `ConnectionService`，避免重复。下列方法可由 ConnectionService 直接提供。

若独立实现，职责仅为 `connections.json` CRUD；**打开/关闭** 仍在 ConnectionService。

~~原 ListSaved / RemoveSaved / SetAutoConnectLast~~ → 见 §3.1–3.8

---

## 5. SchemaService

> 参数 `connectionId` 必填；连接须为 `open`，否则 `CONNECTION_NOT_FOUND`。

### 5.1 ListTables

```go
func (s *SchemaService) ListTables(connectionId string) (*TableList, error)
```

**返回:**

```json
{
  "items": [
    { "name": "users", "type": "table", "rowCount": 1523 },
    { "name": "active_users", "type": "view", "rowCount": null }
  ]
}
```

### 5.2 GetTableSchema

```go
func (s *SchemaService) GetTableSchema(connectionId string, tableName string) (*TableSchema, error)
```

**返回:**

```json
{
  "name": "users",
  "type": "table",
  "columns": [
    {
      "name": "id",
      "dataType": "INTEGER",
      "nullable": false,
      "primaryKey": true,
      "defaultValue": null,
      "position": 1
    }
  ],
  "indexes": [
    {
      "name": "idx_users_email",
      "columns": ["email"],
      "unique": true,
      "primary": false
    }
  ]
}
```

**Errors:** `TABLE_NOT_FOUND`

---

## 6. TableService

### 6.1 BrowseRows

```go
func (s *TableService) BrowseRows(req BrowseRowsRequest) (*PaginatedTableData, error)
```

**BrowseRowsRequest:**

```json
{
  "connectionId": "01HX...",
  "tableName": "users",
  "page": 1,
  "pageSize": 50,
  "sort": "id",
  "order": "asc"
}
```

| 字段 | 默认 | 约束 |
|------|------|------|
| `page` | 1 | ≥ 1 |
| `pageSize` | 50 | 最大 200 |
| `order` | `asc` | `asc` \| `desc` |

**返回:**

```json
{
  "columns": [
    { "name": "id", "dataType": "INTEGER" },
    { "name": "email", "dataType": "TEXT" }
  ],
  "rows": [
    { "id": 1, "email": "a@example.com" }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 50,
    "totalRows": 1523,
    "totalPages": 31
  }
}
```

**Cell Value 序列化:**

| SQLite 类型 | JSON 表示 |
|-------------|-----------|
| NULL | `null` |
| INTEGER / REAL | number |
| TEXT | string |
| BLOB | `{ "type": "blob", "size": 1024 }` |

---

## 7. QueryService

### 7.1 Execute

```go
func (s *QueryService) Execute(req ExecuteQueryRequest) (*QueryResponse, error)
```

**ExecuteQueryRequest:**

```json
{
  "connectionId": "01HX...",
  "sql": "SELECT * FROM users WHERE id = ?",
  "params": [1],
  "maxRows": 1000
}
```

**返回 — SELECT:**

```json
{
  "kind": "result",
  "columns": [{ "name": "id", "dataType": "INTEGER" }],
  "rows": [{ "id": 1 }],
  "rowCount": 1,
  "truncated": false,
  "durationMs": 12
}
```

**返回 — INSERT/UPDATE/DELETE:**

```json
{
  "kind": "exec",
  "rowsAffected": 3,
  "lastInsertId": 42,
  "durationMs": 5
}
```

**Errors:** `SQL_ERROR`, `RESULT_TOO_LARGE`, `READ_ONLY`

SQL 类型检测：首关键字 `SELECT`/`WITH`/`PRAGMA`/`EXPLAIN` → result；其余 → exec。

---

## 8. AppService

### 8.1 GetVersion

```go
func (s *AppService) GetVersion() (*VersionInfo, error)
```

```json
{
  "version": "0.1.0",
  "platform": "darwin",
  "arch": "arm64"
}
```

---

## 9. 前端调用规范

### 8.1 封装层

所有 UI 组件 **不直接** import `wailsjs`，统一经 `frontend/src/lib/api/`：

```typescript
// frontend/src/lib/api/index.ts
export * from './connection'
export * from './schema'
export * from './table'
export * from './query'
export * from './dialog'
export * from './errors'
```

### 8.2 TanStack Query 集成

```typescript
export function useTables() {
  return useQuery({
    queryKey: ['schema', 'tables'],
    queryFn: () => schemaApi.listTables(),
    enabled: connectionStore.connected,
  })
}
```

### 8.3 类型来源

- 领域类型：手写 `frontend/src/lib/types/`（与 Go model 对齐）
- 绑定入口：`wailsjs/go/wails/*`（`wails dev` 自动生成）

---

## 10. 未来 Headless REST 映射（v1.x 预留）

| Service 方法 | 等价 REST |
|--------------|-----------|
| `ConnectionService.Connect` | `POST /api/v1/connection` |
| `SchemaService.ListTables` | `GET /api/v1/schema/tables` |
| `TableService.BrowseRows` | `GET /api/v1/tables/{name}/rows` |
| `QueryService.Execute` | `POST /api/v1/query` |

启用 `--server` 时由 chi router 包装同一 `internal/service` 层。
