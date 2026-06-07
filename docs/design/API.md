# Data Nexus — Service 绑定 API

> 版本: v1.0 · 通信方式: Wails in-memory Bridge · 最后更新: 2026-06-07

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
| `EXPORT_CANCELLED` | 用户取消 CSV 全表导出 |
| `EXPORT_NO_STABLE_KEY` | 表无法确定稳定排序键，无法整表导出 |
| `SAVED_NOT_FOUND` | 已保存连接 ID 不存在 |
| `SECRETS_VAULT_LOCKED` | Vault 已锁定，需解锁后才能访问远程连接密码 |
| `SECRETS_VAULT_NOT_INITIALIZED` | Vault 未初始化，需设置主密码 |
| `SECRETS_VAULT_WRONG_PASSWORD` | Vault 主密码错误 |
| `BATCH_TOO_LARGE` | 批量编辑超过 200 条变更上限 |
| `INVALID_CSV` | CSV 解析或格式校验失败 |
| `IMPORT_FAILED` | CSV 导入执行失败 |
| `EXPORT_FAILED` | CSV 导出失败 |
| `PRIMARY_KEY_REQUIRED` | update 导入模式需要主键列 |
| `FILE_WRITE_FAILED` | 写入文件失败 |
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

**Filters 默认:** `*.csv`

### 2.3 OpenCSVFile（v0.2）

```go
func (s *DialogService) OpenCSVFile() (string, error)
```

**返回:** 用户选择的 CSV 绝对路径；取消时返回 `DIALOG_CANCELLED`。

**Filters:** `*.csv`

用于 CSV 导入向导的文件选择入口。

### 2.4 WriteTextFile — FileService（v0.2）

> **职责分离:** 对话框选路径（DialogService）与文件写入（FileService）分开，便于 headless 模式复用同一写入逻辑。

```go
func (s *FileService) WriteTextFile(path string, content string) error
```

| 参数 | 约束 |
|------|------|
| `path` | 非空绝对路径；须为用户通过 `SaveFile` 或 CLI 显式提供的合法路径 |
| `content` | UTF-8 文本；ExportService 生成的 CSV 字符串 |

**Errors:** `INVALID_PATH`, `FILE_WRITE_FAILED`

**典型流程:** `SaveFile` → 获得路径 → `ExportService.ExportTable` 生成内容 → `WriteTextFile` 落盘。

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

### 3.2a CreateRemoteConnection（v1.0）

```go
func (s *ConnectionService) CreateRemoteConnection(req RemoteConnectRequest) (*SavedConnection, error)
```

创建 PostgreSQL / MySQL 连接：校验 → `SecretsStore.SetPassword` → upsert store → 可选 `OpenConnection`（`req.Open`）。

密码**不**写入 JSON。Vault backend 未初始化/未解锁时返回 `SECRETS_VAULT_*`。

### 3.2b TestConnection（v1.0）

```go
func (s *ConnectionService) TestConnection(req TestConnectionRequest) error
```

使用请求内 `password` 临时 Connect + Ping；**不**读写 vault/keychain。

连接失败时 `CONNECTION_FAILED`，`details.reason`: `network` | `auth` | `database`。

### 3.2c UpdateConnectionPostgresSettings / UpdateConnectionMySQLSettings（v1.0）

```go
func (s *ConnectionService) UpdateConnectionPostgresSettings(connectionID string, update PostgresSettingsUpdate) (*SavedConnection, error)
func (s *ConnectionService) UpdateConnectionMySQLSettings(connectionID string, update MySQLSettingsUpdate) (*SavedConnection, error)
```

更新远程连接配置；`password` 可选（空表示不修改）。若连接已打开则自动重连。

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

### 3.9 Attach / Detach（v0.3）

```go
func (s *ConnectionService) Attach(connectionId string, filePath string, alias string) error
func (s *ConnectionService) Detach(connectionId string, alias string) error
func (s *ConnectionService) ListAttached(connectionId string) ([]AttachedDatabase, error)
```

- 会话级 `ATTACH DATABASE`；关闭连接时自动 `DETACH`
- 只读连接以 `mode=ro` attach
- `ListTables` 返回 `schema` 字段区分 `main` 与 attached alias
- **SQLite only**；PG/MySQL 调用返回 `INVALID_REQUEST`

---

## 3.10 SecretsService（v1.0）

远程连接密码存储。详见 [SECRETS.md](./SECRETS.md)。

```go
func (s *SecretsService) GetSecretsBackend() (string, error)
func (s *SecretsService) VaultInitialized() (bool, error)
func (s *SecretsService) VaultUnlocked() (bool, error)
func (s *SecretsService) InitVault(masterPassword string) error
func (s *SecretsService) UnlockVault(masterPassword string) error
func (s *SecretsService) LockVault() error
func (s *SecretsService) ChangeVaultPassword(oldPassword, newPassword string) error
func (s *SecretsService) IsVaultRequiredForRemote() (bool, error)
```

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

### 5.3 GetTableProfile（v0.3）

```go
func (s *SchemaService) GetTableProfile(connectionId string, tableName string) (*TableProfile, error)
```

列级采样统计：`distinctCount`、`nullPercent`、数值/日期 `min`/`max`、低基数列 `topValues`。超过 10,000 行时 `isSampled=true`。

### 5.4 DetectFTSTable（v0.3）

```go
func (s *SchemaService) DetectFTSTable(connectionId string, tableName string) (*FTSInfo, error)
```

检测 FTS4/FTS5 虚拟表（`{table}_fts`、`fts_{table}` 等命名）。

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
  "order": "asc",
  "filters": [{ "column": "status", "operator": "eq", "value": "active" }],
  "search": "keyword"
}
```

| 字段 | 默认 | 约束 |
|------|------|------|
| `page` | 1 | ≥ 1 |
| `pageSize` | 50 | 最大 200 |
| `order` | `asc` | `asc` \| `desc` |
| `filters` | — | 结构化 WHERE（列名白名单 + 值参数化） |
| `search` | — | FTS `MATCH`（检测到 FTS 表时） |

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

### 6.2 UpdateCellsBatch（v0.2）

```go
func (s *TableService) UpdateCellsBatch(req UpdateCellsBatchRequest) (*UpdateCellsBatchResult, error)
```

在同一数据库事务内批量更新单元格。**全部成功或全部回滚**（all-or-nothing）。

**UpdateCellsBatchRequest:**

```json
{
  "connectionId": "01HX...",
  "tableName": "users",
  "changes": [
    {
      "rowKey": { "id": 1 },
      "column": "email",
      "oldValue": "a@example.com",
      "newValue": "new@example.com"
    }
  ]
}
```

| 字段 | 约束 |
|------|------|
| `changes` | 1–200 条；超出 → `BATCH_TOO_LARGE` |
| `rowKey` | 主键列名→值 map；表无主键时可用 `{ "rowid": 42 }` |
| `column` | 合法标识符；BLOB 列拒绝编辑 |
| `oldValue` / `newValue` | 与 BrowseRows 序列化规则一致；`null` 表示 NULL |

**返回:**

```json
{
  "updatedCount": 3,
  "durationMs": 8
}
```

**Errors:** `CONNECTION_NOT_FOUND`, `TABLE_NOT_FOUND`, `READ_ONLY`, `BATCH_TOO_LARGE`, `SQL_ERROR`

**事务语义:**
- Driver 层 `BEGIN` → 逐条 `UPDATE ... WHERE pk = ? AND column = oldValue`（乐观校验）→ `COMMIT`
- 任一条失败 → `ROLLBACK`，返回 `SQL_ERROR`，`updatedCount` 不返回部分成功

**前端配合:** UI 维护 `pendingEdits` 集合，用户确认后一次性提交；见 [UI_UX.md §4.5.1](./UI_UX.md#451-行内编辑与批量提交)。

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

### 9.1 封装层

所有 UI 组件 **不直接** import `wailsjs`，统一经 `frontend/src/lib/api/`：

```typescript
// frontend/src/lib/api/index.ts
export * from './connection'
export * from './schema'
export * from './table'
export * from './query'
export * from './dialog'
export * from './file'
export * from './export'
export * from './import'
export * from './config'
export * from './errors'
```

### 9.2 TanStack Query 集成

```typescript
export function useTables() {
  return useQuery({
    queryKey: ['schema', 'tables'],
    queryFn: () => schemaApi.listTables(),
    enabled: connectionStore.connected,
  })
}
```

### 9.3 类型来源

- 领域类型：手写 `frontend/src/lib/types/`（与 Go model 对齐）
- 绑定入口：`wailsjs/go/wails/*`（`wails dev` 自动生成）

---

## 10. 未来 Headless REST 映射（v1.x 预留）

| Service 方法 | 等价 REST |
|--------------|-----------|
| `ConnectionService.OpenConnection` | `POST /api/v1/connections/{id}/open` |
| `ConnectionService.CreateConnection` | `POST /api/v1/connections` |
| `SchemaService.ListTables` | `GET /api/v1/connections/{id}/schema/tables` |
| `TableService.BrowseRows` | `GET /api/v1/connections/{id}/tables/{name}/rows` |
| `TableService.UpdateCellsBatch` | `PATCH /api/v1/connections/{id}/tables/{name}/cells` |
| `QueryService.Execute` | `POST /api/v1/connections/{id}/query` |
| `ExportService.ExportTable` | `POST /api/v1/connections/{id}/tables/{name}/export` |
| `ExportService.ExportQueryResult` | `POST /api/v1/export/query-result` |
| `ImportService.PreviewImport` | `POST /api/v1/import/preview` |
| `ImportService.ImportCSV` | `POST /api/v1/connections/{id}/import` |
| `ConfigService.GetConfig` | `GET /api/v1/config` |
| `ConfigService.UpdateConfig` | `PATCH /api/v1/config` |
| `FileService.WriteTextFile` | `PUT /api/v1/files`（headless 受路径白名单约束） |

启用 `--server` 时由 chi router 包装同一 `internal/service` 层。

---

## 11. ExportService（v0.2）

CSV 导出；与 ImportService 共用 [CSVFormatOptions](./DATA_MODEL.md#51-csvformatoptions)。

### 11.1 ExportTable

```go
func (s *ExportService) ExportTable(req ExportTableRequest) (*ExportResult, error)
```

**ExportTableRequest:**

```json
{
  "connectionId": "01HX...",
  "tableName": "users",
  "scope": "page",
  "page": 1,
  "pageSize": 50,
  "sort": "id",
  "order": "asc",
  "format": {
    "delimiter": ",",
    "quoteChar": "\"",
    "includeHeader": true,
    "nullValue": "",
    "encoding": "utf-8"
  },
  "outputPath": "/Users/dev/users.csv"
}
```

| 字段 | 说明 |
|------|------|
| `scope` | `page` — 当前页（配合 page/pageSize/sort/order）；`all` — 全表 |
| `format` | `CSVFormatOptions`；省略时使用应用默认 |
| `outputPath` | 可选；若提供则调用 `FileService.WriteTextFile` 落盘；否则仅返回 `content` |

**返回:**

```json
{
  "content": "id,email\n1,a@example.com\n",
  "rowCount": 1,
  "filePath": "/Users/dev/users.csv",
  "durationMs": 15
}
```

**Errors:** `CONNECTION_NOT_FOUND`, `TABLE_NOT_FOUND`, `EXPORT_FAILED`, `EXPORT_CANCELLED`, `EXPORT_NO_STABLE_KEY`

**全表导出（`ExportTableCSV`）补充:**

- 请求增加 `exportId`（前端生成 UUID）、`defaultPath`（SaveFile 后传入）。
- 无行数硬上限；后端通过 Driver `OpenTableExport` 打开 `TableExportCursor`，按稳定键 **keyset 分页**（`WHERE key > lastKey ORDER BY … LIMIT 1000`）流式写盘，保证顺序稳定、不丢不重。
- SQLite：稳定键为 PK，无主键时 fallback 到 `rowid`；`WITHOUT ROWID` 且无 PK → `EXPORT_NO_STABLE_KEY`。
- PostgreSQL / MySQL（Phase 2/3）：必须 PK，无主键 → `EXPORT_NO_STABLE_KEY`。详见 [EXPORT_MULTI_DIALECT.md](./EXPORT_MULTI_DIALECT.md)。
- 进度事件 `export:progress`：`{ exportId, exported }`（Wails `EventsEmit`）；不预先 `COUNT(*)`，进度仅展示已导出行数。
- 取消：`ExportService.CancelExportTableCSV(exportId)`，删除未完成文件，返回 `EXPORT_CANCELLED`。
- SQL 查询结果导出仍受 `MaxQueryRows`（10,000）约束，可能返回 `RESULT_TOO_LARGE`。

### 11.2 ExportQueryResult

```go
func (s *ExportService) ExportQueryResult(req ExportQueryResultRequest) (*ExportResult, error)
```

将 `QueryService.Execute` 返回的结果集（已在内存中）序列化为 CSV。请求体含 `columns`、`rows` 与 `format`；可选 `outputPath`。

**典型 UI 流程:** 用户执行 SELECT → 结果 Tab 点击「导出 CSV」→ `SaveFile` 选路径 → `ExportQueryResult` + `WriteTextFile`。

---

## 12. ImportService（v0.2）

CSV 导入；解析与 ExportService 对称。

### 12.1 PreviewImport

```go
func (s *ImportService) PreviewImport(req PreviewImportRequest) (*ImportPreview, error)
```

**PreviewImportRequest:**

```json
{
  "filePath": "/Users/dev/import.csv",
  "format": { "delimiter": ",", "quoteChar": "\"", "includeHeader": true },
  "previewRows": 20
}
```

**返回:**

```json
{
  "columns": [
    { "name": "id", "inferredType": "INTEGER", "sampleValues": ["1", "2"] },
    { "name": "email", "inferredType": "TEXT", "sampleValues": ["a@example.com"] }
  ],
  "totalRows": 1523,
  "warnings": ["Column 'note' has mixed types"]
}
```

**Errors:** `INVALID_PATH`, `INVALID_CSV`

### 12.2 ImportCSV

```go
func (s *ImportService) ImportCSV(req ImportCSVRequest) (*ImportResult, error)
```

**ImportCSVRequest:**

```json
{
  "connectionId": "01HX...",
  "filePath": "/Users/dev/import.csv",
  "format": { "delimiter": ",", "quoteChar": "\"", "includeHeader": true },
  "target": {
    "kind": "existing",
    "tableName": "users",
    "mode": "append",
    "columnMapping": { "id": "id", "email": "email_address" }
  }
}
```

**target.kind:**

| 值 | 行为 |
|----|------|
| `new` | `CREATE TABLE` + `INSERT`；需提供 `newTableName`、`columns`（名+类型） |
| `existing` | 写入已有表；需 `tableName` + `columnMapping` |

**target.mode（`existing` 时）:**

| 值 | 行为 |
|----|------|
| `append` | `INSERT` 追加行 |
| `update` | 按主键 `UPSERT`；表无主键 → `PRIMARY_KEY_REQUIRED` |

**返回:**

```json
{
  "rowsImported": 1523,
  "rowsUpdated": 0,
  "tableName": "users",
  "durationMs": 420
}
```

**Errors:** `CONNECTION_NOT_FOUND`, `READ_ONLY`, `TABLE_NOT_FOUND`, `PRIMARY_KEY_REQUIRED`, `INVALID_CSV`, `IMPORT_FAILED`

**事务语义:** 单事务批量写入；失败整批回滚。

---

## 13. CannedQueryService（v0.3）

持久化至 `~/.data-nexus/queries.json`。

```go
func (s *CannedQueryService) ListCannedQueries() (*CannedQueryList, error)
func (s *CannedQueryService) SaveCannedQuery(req SaveCannedQueryRequest) (*CannedQuery, error)
func (s *CannedQueryService) DeleteCannedQuery(id string) error
```

---

## 14. SqlExecutionService（v0.3）

执行日志存储于 `~/.data-nexus/sql-global.db`；由 `QueryService.Execute` 成功时自动写入，无单独 Insert API。

```go
func (s *SqlExecutionService) ListQueryHistory(connectionID string) ([]string, error)
func (s *SqlExecutionService) ListSqlExecutions(connectionID string, limit int) (*SqlExecutionList, error)
```

| 方法 | 说明 |
|------|------|
| `ListQueryHistory` | 该连接最近执行的 SQL，按文本去重，最多 50 条（SQL Tab 下拉） |
| `ListSqlExecutions` | 完整 `SqlExecutionRecord` 列表，按时间倒序；`limit` 默认 50，上限 200 |

记录字段见 [DATA_MODEL.md §9](./DATA_MODEL.md#9-sql-执行历史v03)。

---

## 15. ConfigService（v0.2）

读写 `~/.data-nexus/config.yaml`；模型见 [DATA_MODEL.md §2.4](./DATA_MODEL.md#24-appconfig应用配置含日志)。

### 15.1 GetConfig

```go
func (s *ConfigService) GetConfig() (*AppConfig, error)
```

返回当前生效配置（含默认值填充后的完整视图）。

### 15.2 UpdateConfig

```go
func (s *ConfigService) UpdateConfig(req UpdateConfigRequest) (*AppConfig, error)
```

**UpdateConfigRequest:** 部分字段 PATCH；仅更新提供的键。

```json
{
  "log": {
    "level": "debug",
    "output": "both"
  }
}
```

**行为:**
- 持久化到 `config.yaml`
- 日志级别 / 输出目标热更新（调用 `logger.Reload`）
- CLI flags 仍优先于配置文件（启动时生效，运行时 PATCH 不覆盖 CLI）

**Errors:** `INVALID_REQUEST`
