# Data Nexus — 数据模型

> 版本: v0.3 · 描述后端领域模型与多数据库扩展策略

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
| `ID` | ULID | 持久化连接 ID |
| `Type` | `"sqlite"` | `"postgres"`, `"mysql"` |
| `DisplayName` | 文件名 | 用户自定义别名 |
| `Config` | SQLiteConfig | 联合类型 |

### 2.2 DriverConfig（联合配置）

```typescript
// 前端 TypeScript 类型示意
type DriverConfig =
  | { type: 'sqlite'; sqlite: SQLiteConfig }
  // | { type: 'postgres'; postgres: PostgresConfig }
  // | { type: 'mysql'; mysql: MySQLConfig }

interface SQLiteConfig {
  filePath: string
  readOnly?: boolean
}

// v1.x 预留
interface PostgresConfig {
  host: string
  port: number
  database: string
  user: string
  password?: string
  sslMode?: 'disable' | 'require' | 'verify-full'
  schema?: string  // default "public"
}
```

### 2.3 SavedConnection（v1.0 持久化，MVP 不实现）

```go
type SavedConnection struct {
    ID        string
    Name      string
    Type      DriverType
    Config    DriverConfig  // 敏感字段加密存储
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

存储：本地 JSON 文件 `~/.data-nexus/connections.json` 或 SQLite 配置库。

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

### 3.3 SQLite → 统一模型映射

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

## 5. 值序列化

### 5.1 Go → JSON

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

### 5.2 前端 Cell 渲染

| JSON 值 | 渲染 |
|---------|------|
| `null` | 灰色斜体 `NULL` |
| `{ type: "blob", size: N }` | `[BLOB N bytes]` |
| number / string | 原样 |
| boolean | `true` / `false` |

---

## 6. 多数据库差异矩阵（设计参考）

| 能力 | SQLite | PostgreSQL | MySQL |
|------|--------|------------|-------|
| 连接方式 | 文件路径 | host+port+db | host+port+db |
| Schema 层级 | 无（单 schema） | `schema.table` | `database.table` |
| 视图 | ✓ | ✓ | ✓ |
| 自增 ID | ROWID / INTEGER PK | SERIAL / IDENTITY | AUTO_INCREMENT |
| 类型系统 | 动态 | 丰富 | 丰富 |
| 只读连接 | `?mode=ro` | `default_transaction_read_only` | `READ ONLY` |

**抽象策略:**
- `TableInfo.Schema` 在 SQLite 为 `null`，Postgres 为 `"public"` 等
- `BrowseTable` 内部构建方言特定的 quoted identifier
- `DataType` 统一为大写字符串；前端不依赖特定 DB 类型语义

---

## 7. 标识符校验

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

## 8. 会话状态（前端）

```typescript
interface AppState {
  connection: Connection | null
  selectedTable: string | null
  activeTab: 'schema' | 'data' | 'sql'
  queryHistory: QueryHistoryItem[]  // max 50
  theme: 'light' | 'dark' | 'system'
}

interface QueryHistoryItem {
  sql: string
  executedAt: string
  durationMs: number
  success: boolean
}
```

Zustand store + TanStack Query 管理后端状态；通过 `frontend/src/lib/api/` 调用 Wails Service 绑定（非 HTTP fetch）。MVP 不持久化到 localStorage。
