# Data Nexus — 数据模型

> 版本: v1.0 · 描述后端领域模型与多数据库扩展策略

---

## 1. 设计原则

1. **数据库无关的 Schema 抽象**: 前端只消费统一 JSON 结构，不感知 SQLite PRAGMA 细节
2. **Driver 内部转换**: 各数据库特有元数据查询封装在 Driver 实现内
3. **Optional 字段吸收差异**: PostgreSQL 的 `schema`、MySQL 的 `engine` 等用 optional 字段扩展

---

## 2. 连接模型

### 2.1 Connection（活跃连接）

```go
type Connection struct {
    ID          string       `json:"id"`
    Type        DriverType   `json:"type"`
    DisplayName string       `json:"displayName"`
    Config      DriverConfig `json:"config"`
    ConnectedAt time.Time    `json:"connectedAt"`
}
```

| 字段 | MVP | v1.0+ |
|------|-----|-------|
| `ID` | ULID | 同左 |
| `Type` | `"sqlite"` | `"postgres"`, `"mysql"` |
| `DisplayName` | 文件名 | 用户自定义别名 |
| `Config` | SQLiteConfig | 联合类型 |

> **活跃连接** 与 **已保存连接** 共用 `id`：保存于 `connections.json`；`OpenConnection` 后在内存中挂载 Driver。

**ConnectionManager（MVP）:**

```go
type ConnectionManager struct {
    mu      sync.RWMutex
    store   *ConnectionStore      // connections.json
    active  map[string]*Session   // id -> open driver session
}

type Session struct {
    Connection *Connection
    Driver     driver.Driver
}
```

- 支持 **多条** `active` 并存
- `OpenConnection(id)` 向 map 添加；`CloseConnection(id)` 移除并 `Driver.Close()`
- `ListConnections` 合并 store 与 active 状态 → `open` | `closed`

### 2.2 DriverConfig（联合配置）

```typescript
// 前端 TypeScript 类型示意
type DriverConfig =
  | { type: 'sqlite'; sqlite: SQLiteConfig }
  | { type: 'postgres'; postgres: PostgresConfig }
  | { type: 'mysql'; mysql: MySQLConfig }

interface SQLiteConfig {
  filePath: string
  readOnly?: boolean
  wal?: boolean
}

interface PostgresConfig {
  host: string
  port: number
  database: string
  user: string
  // password 仅运行时注入，不持久化 — 见 SECRETS.md
  sslMode?: 'disable' | 'require' | 'verify-full'
  schema?: string  // default "public"
  readOnly: boolean
}

interface MySQLConfig {
  host: string
  port: number
  database: string
  user: string
  tls: boolean
  readOnly: boolean
}
```

### 2.3 SavedConnection（MVP 持久化）

```go
type SavedConnection struct {
    ID              string          `json:"id"`
    Name            string          `json:"name"`
    Type            DriverType      `json:"type"`
    Config          DriverConfig    `json:"config"`
    SecretsBackend  SecretsBackend  `json:"secretsBackend,omitempty"` // keychain | vault；SQLite 为空
    CreatedAt       time.Time       `json:"createdAt"`
    UpdatedAt       time.Time       `json:"updatedAt"`
    LastUsedAt      time.Time       `json:"lastUsedAt"`
}
```

**密码边界:** 远程连接 password 存于 Keychain 或 `~/.data-nexus/vault/`，**不**出现在 `connections.json`。见 [SECRETS.md](./SECRETS.md).

**存储文件:** `~/.data-nexus/connections.json`

```json
{
  "version": 1,
  "restoreOpenOnStartup": false,
  "openConnectionIds": [],
  "items": [
    {
      "id": "01HX...",
      "name": "app.db",
      "type": "sqlite",
      "config": {
        "type": "sqlite",
        "sqlite": {
          "filePath": "/Users/dev/project/app.db",
          "readOnly": false
        }
      },
      "lastUsedAt": "2026-06-07T08:00:00Z"
    }
  ]
}
```

**规则:**
- `CreateConnection` / `OpenConnectionFromFile` → upsert `items`
- `OpenConnection` → 加入 `active`；退出时可选写入 `openConnectionIds`
- `CloseConnection` → 从 `active` 移除；`items` 保留
- `RemoveConnection` → 关闭 + 从 `items` 删除
- 同一路径新建 → 可复用已有 `id` 或提示已存在（MVP：复用并更新）

### 2.4 AppConfig（应用配置，含日志）

```yaml
# ~/.data-nexus/config.yaml
log:
  level: info          # debug | info | warn | error，默认 info
  output: auto         # auto | console | file | both
  file:
    path: ""           # 空则平台默认，见 ARCHITECTURE.md §3.5
    max_size_mb: 10
    max_backups: 5
    max_age_days: 30
    compress: true
```

```go
type LogConfig struct {
    Level  string         `yaml:"level"`  // default: "info"
    Output string         `yaml:"output"` // default: "auto"
    File   LogFileConfig  `yaml:"file"`
}

type LogFileConfig struct {
    Path       string `yaml:"path"`
    MaxSizeMB  int    `yaml:"max_size_mb"`  // default: 10
    MaxBackups int    `yaml:"max_backups"`  // default: 5
    MaxAgeDays int    `yaml:"max_age_days"` // default: 30
    Compress   bool   `yaml:"compress"`     // default: true
}
```

**优先级：** CLI flags > 配置文件 > 内置默认值。`output=auto` 时，`wails dev` / debug build 写控制台，release build 写文件。

---

## 3. Schema 模型

### 3.1 TableInfo（表/视图列表项）

```go
type TableInfo struct {
    Name     string    `json:"name"`
    Type     TableType `json:"type"`     // "table" | "view"
    Schema   *string   `json:"schema"`   // PostgreSQL: "public"; SQLite: null
    RowCount *int64    `json:"rowCount"` // optional, P1
    Comment  *string   `json:"comment"`  // optional, future
}

type TableType string
const (
    TableTypeTable TableType = "table"
    TableTypeView  TableType = "view"
)
```

### 3.2 TableSchema（表结构详情）

```go
type TableSchema struct {
    Name    string       `json:"name"`
    Type    TableType    `json:"type"`
    Schema  *string      `json:"schema"`
    Columns []ColumnInfo `json:"columns"`
    Indexes []IndexInfo  `json:"indexes"`
}

type ColumnInfo struct {
    Name         string  `json:"name"`
    DataType     string  `json:"dataType"`     // 规范化类型名
    NativeType   *string `json:"nativeType"`   // 原始类型，如 "VARCHAR(255)"
    Nullable     bool    `json:"nullable"`
    PrimaryKey   bool    `json:"primaryKey"`
    DefaultValue *string `json:"defaultValue"`
    Position     int     `json:"position"`
    Comment      *string `json:"comment"`
}

type IndexInfo struct {
    Name    string   `json:"name"`
    Columns []string `json:"columns"`
    Unique  bool     `json:"unique"`
    Primary bool     `json:"primary"`
}
```

### 3.3 TableProfile（v0.3 列画像）

```go
type TableProfile struct {
    TableName   string          `json:"tableName"`
    SampledRows int64           `json:"sampledRows"`
    TotalRows   *int64          `json:"totalRows"`
    IsSampled   bool            `json:"isSampled"`
    Columns     []ColumnProfile `json:"columns"`
}

type ColumnProfile struct {
    Name             string       `json:"name"`
    DistinctCount    *int64       `json:"distinctCount"`
    NullPercent      *float64     `json:"nullPercent"`
    MinValue         *string      `json:"minValue"`
    MaxValue         *string      `json:"maxValue"`
    TopValues        []ValueCount `json:"topValues"`
    IsLowCardinality bool         `json:"isLowCardinality"`
}
```

大表（>10,000 行）基于 `LIMIT 10000` 子查询采样。

### 3.4 RowFilter（v0.3 结构化筛选）

```go
type RowFilter struct {
    Column   string         `json:"column"`
    Operator FilterOperator `json:"operator"`
    Value    *string        `json:"value,omitempty"`
    Values   []string       `json:"values,omitempty"`
}
```

### 3.5 SQLite → 统一模型映射

| 统一字段 | SQLite 来源 |
|----------|-------------|
| `TableInfo.name` | `sqlite_master.name` WHERE type IN ('table','view') |
| `TableInfo.type` | `sqlite_master.type` |
| `ColumnInfo.*` | `PRAGMA table_info({table})` |
| `IndexInfo.*` | `PRAGMA index_list` + `PRAGMA index_info` |
| `ColumnInfo.dataType` | `PRAGMA table_info.type` 原样或规范化 |

**SQLite 系统表过滤:**

```sql
SELECT name, type FROM sqlite_master
WHERE type IN ('table', 'view')
  AND name NOT LIKE 'sqlite_%'
ORDER BY name
```

---

## 4. 查询结果模型

### 4.1 QueryResult

```go
type QueryResult struct {
    Columns   []ColumnMeta     `json:"columns"`
    Rows      []map[string]any `json:"rows"`
    RowCount  int              `json:"rowCount"`
    Truncated bool             `json:"truncated"`
    Duration  time.Duration    `json:"-"` // API 输出 durationMs
}

type ColumnMeta struct {
    Name     string `json:"name"`
    DataType string `json:"dataType"`
}
```

### 4.2 ExecResult

```go
type ExecResult struct {
    RowsAffected int64 `json:"rowsAffected"`
    LastInsertID int64 `json:"lastInsertId"`
    Duration     time.Duration
}
```

### 4.3 BrowseOptions

```go
type BrowseOptions struct {
    Page     int
    PageSize int
    Sort     string
    Order    SortOrder
}

type SortOrder string
const (
    SortAsc  SortOrder = "asc"
    SortDesc SortOrder = "desc"
)
```

### 4.4 PaginatedTableData

```go
type PaginatedTableData struct {
    Columns    []ColumnMeta       `json:"columns"`
    Rows       []map[string]any   `json:"rows"`
    Pagination PaginationMeta     `json:"pagination"`
}

type PaginationMeta struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"pageSize"`
    TotalRows  int64 `json:"totalRows"`
    TotalPages int   `json:"totalPages"`
}
```

---

## 5. CSV 与批量编辑模型

> v0.2 新增；ImportService / ExportService / TableService.UpdateCellsBatch 共用。

### 5.1 CSVFormatOptions

导入与导出共用同一结构，保证往返一致。

```go
type CSVFormatOptions struct {
    Delimiter     string `json:"delimiter"`     // 默认 ","
    QuoteChar     string `json:"quoteChar"`     // 默认 "\""
    IncludeHeader bool   `json:"includeHeader"` // 默认 true
    NullValue     string `json:"nullValue"`     // NULL 的 CSV 表示，默认 ""
    LineEnding    string `json:"lineEnding"`    // "lf" | "crlf"，默认 "lf"
    Encoding      string `json:"encoding"`      // "utf-8" | "utf-8-bom"，默认 "utf-8"
}
```

| 字段 | 说明 |
|------|------|
| `Delimiter` | 单字符；`,` `\t` `;` 等 |
| `QuoteChar` | 包裹含分隔符/换行的字段 |
| `IncludeHeader` | 首行是否为列名 |
| `NullValue` | 导出时 JSON `null` 的文本；导入时识别为 NULL |
| `Encoding` | `utf-8-bom` 便于 Excel 直接打开中文 CSV |

**默认来源:** `AppConfig.defaults.csv`（可选，见 ConfigService）；UI ExportDialog / ImportWizard 可覆盖。

### 5.2 CellChange（单格变更）

```go
type CellChange struct {
    RowKey   map[string]any `json:"rowKey"`   // 主键列 → 值；或 rowid
    Column   string         `json:"column"`
    OldValue any            `json:"oldValue"`
    NewValue any            `json:"newValue"`
}
```

`RowKey` 示例：

```json
{ "id": 42 }
{ "rowid": 7 }
```

### 5.3 UpdateCellsBatchRequest / Result

```go
type UpdateCellsBatchRequest struct {
    ConnectionID string       `json:"connectionId"`
    TableName    string       `json:"tableName"`
    Changes      []CellChange `json:"changes"` // 1–200
}

type UpdateCellsBatchResult struct {
    UpdatedCount int `json:"updatedCount"`
    DurationMs   int `json:"durationMs"`
}
```

**约束:** `len(Changes)` ∈ [1, 200]；Driver 单事务执行。

### 5.4 ImportTarget / ImportMode

```go
type ImportTargetKind string
const (
    ImportTargetNew      ImportTargetKind = "new"
    ImportTargetExisting ImportTargetKind = "existing"
)

type ImportMode string
const (
    ImportModeAppend ImportMode = "append" // INSERT
    ImportModeUpdate ImportMode = "update" // UPSERT，需主键
)

type ImportTarget struct {
    Kind          ImportTargetKind   `json:"kind"`
    TableName     string             `json:"tableName,omitempty"`     // existing
    NewTableName  string             `json:"newTableName,omitempty"`  // new
    Mode          ImportMode         `json:"mode,omitempty"`          // existing only
    ColumnMapping map[string]string  `json:"columnMapping"`           // csvCol → dbCol
    Columns       []ImportColumnDef  `json:"columns,omitempty"`       // new table
}

type ImportColumnDef struct {
    Name     string `json:"name"`
    DataType string `json:"dataType"` // INTEGER | TEXT | REAL | BLOB
}
```

### 5.5 ExportScope

```go
type ExportScope string
const (
    ExportScopePage ExportScope = "page" // 当前分页
    ExportScopeAll  ExportScope = "all"  // 全表
)

type ExportTableRequest struct {
    ConnectionID string            `json:"connectionId"`
    TableName    string            `json:"tableName"`
    Scope        ExportScope       `json:"scope"`
    Page         int               `json:"page,omitempty"`
    PageSize     int               `json:"pageSize,omitempty"`
    Sort         string            `json:"sort,omitempty"`
    Order        SortOrder         `json:"order,omitempty"`
    Format       CSVFormatOptions  `json:"format"`
    OutputPath   string            `json:"outputPath,omitempty"`
}

type ExportResult struct {
    Content    string `json:"content,omitempty"`
    RowCount   int    `json:"rowCount"`
    FilePath   string `json:"filePath,omitempty"`
    DurationMs int    `json:"durationMs"`
}

// 桌面端全表导出（ExportTableCSV）
type ExportTableCSVRequest struct {
    ConnectionID string           `json:"connectionId"`
    TableName    string           `json:"tableName"`
    Format       CSVFormatOptions `json:"format"`
    DefaultPath  string           `json:"defaultPath,omitempty"`
    ExportID     string           `json:"exportId,omitempty"`
}

type ExportProgress struct {
    ExportID string `json:"exportId"`
    Exported int    `json:"exported"`
}

const CSVExportWarnRows = 10_000 // UI 警告阈值，非硬限制
```

**整表导出 Driver 契约（内部，不暴露 Wails）：**

```go
type StableRowKey struct {
    Columns []string // ORDER BY 列（复合 PK 多列）
    Source  string   // "primary_key" | "unique_index" | "rowid"（仅 SQLite）
}

type TableExportOptions struct {
    BatchSize int // 默认 1000
}

type TableExportBatch struct {
    Rows    []map[string]any
    HasMore bool
}

// driver/export.TableExportCursor
//   Columns() / NextBatch(ctx) / Close()
```

详见 [EXPORT_MULTI_DIALECT.md](./EXPORT_MULTI_DIALECT.md)。

### 5.6 ImportPreview / ImportResult

```go
type ImportPreviewColumn struct {
    Name          string   `json:"name"`
    InferredType  string   `json:"inferredType"`
    SampleValues  []string `json:"sampleValues"`
}

type ImportPreview struct {
    Columns   []ImportPreviewColumn `json:"columns"`
    TotalRows int                   `json:"totalRows"`
    Warnings  []string              `json:"warnings,omitempty"`
}

type ImportResult struct {
    RowsImported int    `json:"rowsImported"`
    RowsUpdated  int    `json:"rowsUpdated"`
    TableName    string `json:"tableName"`
    DurationMs   int    `json:"durationMs"`
}
```

### 5.7 前端 PendingEdits 状态

```typescript
interface PendingEdit {
  rowKey: Record<string, unknown>
  column: string
  oldValue: unknown
  newValue: unknown
}

interface PendingEditsState {
  edits: Map<string, PendingEdit> // key: `${rowKeyHash}:${column}`
  tableName: string
  connectionId: string
}
```

提交时将 `edits` 转为 `CellChange[]` 调用 `UpdateCellsBatch`；放弃时清空 Map。

---

## 6. 值序列化

### 6.1 Go → JSON

```go
func SerializeCellValue(v any, colType string) any {
    if v == nil {
        return nil
    }
    switch val := v.(type) {
    case []byte:
        return map[string]any{"type": "blob", "size": len(val)}
    case time.Time:
        return val.Format(time.RFC3339Nano)
    default:
        return val
    }
}
```

### 6.2 前端 Cell 渲染

| JSON 值 | 渲染 |
|---------|------|
| `null` | 灰色斜体 `NULL` |
| `{ type: "blob", size: N }` | `[BLOB N bytes]` |
| number / string | 原样 |
| boolean | `true` / `false` |

---

## 7. 多数据库差异矩阵（设计参考）

| 能力 | SQLite | PostgreSQL | MySQL |
|------|--------|------------|-------|
| 连接方式 | 文件路径 | host+port+db | host+port+db |
| Schema 层级 | 无（单 schema） | `schema.table` | `database.table` |
| 视图 | ✓ | ✓ | ✓ |
| 自增 ID | ROWID / INTEGER PK | SERIAL / IDENTITY | AUTO_INCREMENT |
| 类型系统 | 动态 | 丰富 | 丰富 |
| 只读连接 | `?mode=ro` | `default_transaction_read_only` | `READ ONLY` |
| 整表 CSV 导出 | `OpenTableExport` 流式 | 同左 | 同左 |
| 扫描方式 | `SELECT *` 单查询 | 同左 | 同左 |
| 无主键表导出 | 允许（无序） | 允许（无序） | 允许（无序） |
| 导出分批 | NextBatch，batch=1000 | 同左 | 同左 |
| 导出行数上限 | 无 | 无 | 无 |
| CSV 格式选项 | `CSVFormatOptions` 共用 | 同左 | 同左 |
| 最低版本 | 3.15+（row value） | 9.4+ | **best-effort 5.7**，推荐 **8.0+** |

**MySQL 版本说明：** 驱动生成的 SQL（浏览、schema、导出）按 5.7 兼容编写；用户自行执行的 CTE、窗口函数等需 MySQL 8.0+。MySQL 8 默认 `caching_sha2_password` 认证建议开启 TLS，或改用 `mysql_native_password`。

**抽象策略:**
- `TableInfo.Schema` 在 SQLite 为 `null`，Postgres 为 `"public"` 等
- `BrowseTable` 内部构建方言特定的 quoted identifier
- `DataType` 统一为大写字符串；前端不依赖特定 DB 类型语义
- 整表导出走 `OpenTableExport` / `TableExportCursor`，与 `BrowseTable` 分离；详见 [EXPORT_MULTI_DIALECT.md](./EXPORT_MULTI_DIALECT.md)

---

## 8. 标识符校验

防止 SQL 注入（动态表名/列名场景）：

```go
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func ValidateIdentifier(name string) error {
    if !identifierRegex.MatchString(name) {
        return ErrInvalidIdentifier
    }
    return nil
}
```

SQLite 引用：`"identifier"`（双引号）
Postgres：`"identifier"`
MySQL：`` `identifier` ``

各 Driver 各自实现 `QuoteIdentifier(name string) string`。

---

## 9. SQL 执行历史（v0.3）

持久化至 `~/.data-nexus/sql-global.db`（元数据 SQLite，与用户打开的 `.db` 分离）。由 `QueryService.Execute` 成功时自动写入；与 Canned Queries（`queries.json`）无关。

### 9.1 表 `sql_executions`

| 列 | 类型 | 说明 |
|----|------|------|
| `id` | INTEGER PK | 自增 |
| `connection_id` | TEXT | 执行时连接 ID |
| `sql_text` | TEXT | 规范化后的 SQL |
| `kind` | INTEGER | `0` = 只读查询（`result`），`1` = 写操作（`exec`） |
| `effect_rows` | INTEGER | result → `rowCount`；exec → `rowsAffected` |
| `duration_ms` | INTEGER | 耗时 |
| `executed_at` | TEXT | UTC RFC3339 |

### 9.2 Go 模型

```go
type SqlExecutionKind int

const (
    SqlExecutionResult SqlExecutionKind = 0
    SqlExecutionExec   SqlExecutionKind = 1
)

type SqlExecutionRecord struct {
    ID           int64
    ConnectionID string
    SQL          string
    Kind         SqlExecutionKind  // JSON: "result" | "exec"
    EffectRows   int64
    DurationMs   int64
    ExecutedAt   time.Time
}
```

### 9.3 前端

SQL 编辑器历史下拉：`SqlExecutionService.ListQueryHistory(connectionId)` → `string[]`（按 `sql_text` 去重，最多 50 条）。TanStack Query key：`['query-history', connectionId]`；Execute 成功后 `invalidateQueries`。

完整审计列表：`ListSqlExecutions(connectionId, limit?)` → `SqlExecutionList`（默认 50，上限 200；本次 UI 不展示）。

### 9.4 其它会话状态

`workspaceStore`（localStorage）仍持久化 Tab、筛选、排序等；见 v0.3 workspace 设计。连接列表由 `connections.json` 持久化。
