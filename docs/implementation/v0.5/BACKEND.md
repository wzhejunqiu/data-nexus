# v0.5 后端实施方案

> **API 契约:** [API.md §14](../../design/API.md#14-sqlexecutionservicev03) · **数据模型:** [DATA_MODEL.md §9](../../design/DATA_MODEL.md)

---

## 1. ListAllSqlExecutions

### 1.1 需求

v0.5 新增跨连接 SQL 执行历史查询，供 `SqlExecutionHistoryDialog` 使用。

| 方法 | 说明 |
|------|------|
| `ListAllSqlExecutions(limit int)` | 全部连接的执行记录，按 `executed_at DESC`；默认 limit=50，上限 200 |

现有 per-connection 方法不变：

- `ListQueryHistory(connectionID)` — SQL Tab 下拉
- `ListSqlExecutions(connectionID, limit)` — 单连接完整记录

### 1.2 Store 层

**接口扩展** — `internal/executionlog/store.go`:

```go
type Store interface {
    // ... existing ...
    ListAllExecutions(ctx context.Context, limit int) ([]model.SqlExecutionRecord, error)
}
```

**SQLite 实现** — `internal/executionlog/sqlite/store.go`:

```sql
SELECT id, connection_id, sql_text, kind, effect_rows, duration_ms, executed_at
FROM sql_executions
ORDER BY executed_at DESC
LIMIT ?
```

- 与 `ListExecutions` 共用行扫描逻辑，可提取 `scanExecutionRows`
- `limit <= 0` → 使用默认 50
- `limit > 200` → 截断为 200

### 1.3 Service 层

**文件:** `internal/service/sql_execution_service.go`

```go
func (s *SqlExecutionService) ListAllSqlExecutions(ctx context.Context, limit int) (*model.SqlExecutionList, error) {
    if s.store == nil {
        return &model.SqlExecutionList{Items: []model.SqlExecutionRecord{}}, nil
    }
    limit = clampLimit(limit, 50, 200)
    items, err := s.store.ListAllExecutions(ctx, limit)
    if err != nil {
        return nil, err
    }
    return &model.SqlExecutionList{Items: items}, nil
}
```

- 复用已有 `clampLimit`  helper（若 `ListSqlExecutions` 已有则共用）

### 1.4 Wails 绑定

**文件:** `internal/wails/sql_execution.go`

```go
func (s *SqlExecutionService) ListAllSqlExecutions(limit int) (*model.SqlExecutionList, error) {
    return call(s.log, "SqlExecutionService.ListAllSqlExecutions", func() (*model.SqlExecutionList, error) {
        return s.svc.ListAllSqlExecutions(context.Background(), limit)
    })
}
```

**生成绑定:**

```bash
wails generate module
# 或项目 Makefile 中的 generate 目标
```

前端调用路径: `wailsjs/go/wails/SqlExecutionService.ListAllSqlExecutions`

### 1.5 测试

| 文件 | 用例 |
|------|------|
| `internal/executionlog/sqlite/store_test.go` | 插入 conn-A、conn-B 各多条；`ListAllExecutions(3)` 返回全局最新 3 条且顺序正确 |
| `internal/service/sql_execution_service_test.go` | nil store 返回空；limit 0 → 50；limit 999 → 200 |
| `internal/wails/services_test.go` | Wails wrapper smoke test |

运行:

```bash
make test
```

---

## 2. Go 原生菜单

### 2.1 平台策略（v0.5 定稿）

| 平台 | 文件 | 策略 |
|------|------|------|
| **darwin** | `app_menu_darwin.go` | 系统菜单栏提供 File/View/Help；`EventsEmit` 驱动 React Dialog |
| **!darwin** | `app_menu.go` | 仅 App / Edit / Window；菜单功能由前端 `AppMenuBar` 提供 |

`ApplicationMenu()` 委托 `platformApplicationMenu()`。

### 2.2 macOS 菜单结构

```go
// app_menu_darwin.go — 顺序遵循 HIG：App 必须首位
appMenu.Append(menu.AppMenu())
appMenu.Append(menu.SubMenu("文件", fileSub))   // N/O/W + 新建分组 → app:new-group 等
appMenu.Append(menu.EditMenu())
appMenu.Append(menu.SubMenu("视图", viewSub))   // SQL 历史、设置（Cmd+,）
appMenu.Append(menu.WindowMenu())
appMenu.Append(menu.SubMenu("帮助", helpSub))   // 关于 → app:about
```

菜单文案按 `LANG` 环境变量中/英（默认中文）；与前端 i18n 完全同步待 v0.5.1。

### 2.3 Win/Linux 菜单结构

```go
// app_menu.go
appMenu.Append(menu.AppMenu())
appMenu.Append(menu.EditMenu())
appMenu.Append(menu.WindowMenu())
```

### 2.4 文件拖放

`handleFileDrop(x, y, paths)` **不再**在 Go 侧直接 `OpenConnectionFromFile`，改为：

```go
runtime.EventsEmit(a.ctx, "app:file-drop", map[string]interface{}{
    "x": x, "y": y, "paths": dbPaths,
})
```

前端 `useSidebarFileDrop`：

- `elementFromPoint(x,y)` 命中已 open SQLite 连接 → `AttachDatabaseDialog` 预填路径
- 否则 → `connectionApi.openFromFile`（空白区新建连接）

### 2.5 已移除的 Go handler

| Handler | 替代 |
|---------|------|
| `handleOpenDatabase()` | macOS：`app:open-sqlite`；Win/Linux：`AppMenuBar` |
| `handleCloseConnection()` | `app:close-connection` → 前端 `connectionApi.close` |
| 新建顶层 Group | `app:new-group` → `AppShell.createRootGroup` |
| View 主题/语言原生项 | SettingsDialog |
| `ShowAbout()` 原生对话框 | `app:about` → `AboutDialog` |

### 2.6 macOS 注意事项

- `menu.AppMenu()` 含 **Quit**（`Cmd+Q`）；应用内 MenuBar **不显示**「退出」
- 应用内顶栏无 Menubar，避免与系统栏重复（HIG）
- v0.5.1：原生菜单文案随 `app:language` 动态更新、`MenuUpdateApplicationMenu`

---

## 3. 无需变更的部分

| 模块 | 原因 |
|------|------|
| `QueryService.Execute` | 已自动写入 execution log |
| `ConnectionService.Attach` / `Detach` / `ListAttached` | v0.3 已有；v0.5 仅 UI 重构 |
| `ConnectionService` restore on startup API | 已有 |
| `AppService.GetVersion` | AboutDialog 直接复用 |
| `ConfigService` | 设置 Dialog 已有 |

---

## 4. 连接树 API（ListNamespaces / ListSchemas）

> 完整交互见 [CONNECTION_TREE.md](./CONNECTION_TREE.md)

### 4.1 SchemaService 扩展

```go
func (s *SchemaService) ListNamespaces(ctx context.Context, connectionID string) (*model.NamespaceList, error)
func (s *SchemaService) ListSchemas(ctx context.Context, connectionID, database string) (*model.SchemaList, error)
func (s *SchemaService) ListTables(ctx context.Context, connectionID string, opts model.ListTablesOptions) (*model.TableList, error)
```

`SchemaService` 按 `connectionID` 取 driver 实例并委托调用，**不做 Type switch 分发**。

### 4.2 Driver 接口扩展

**文件:** `internal/driver/driver.go`

```go
type Driver interface {
    // ... existing ...
    ListNamespaces(ctx context.Context) ([]model.NamespaceInfo, error)
    ListSchemas(ctx context.Context, database string) ([]model.SchemaInfo, error) // PG only
}
```

`ListTables` 签名扩展为接受 `ListTablesOptions{Database, Schema}`。

| 方法 | SQLite | MySQL | PostgreSQL |
|------|--------|-------|------------|
| ListNamespaces | `main` + `ListAttached` | SHOW DATABASES | SELECT datname FROM pg_database |
| ListSchemas | 返回空 / ErrNotSupported | 返回空 | schemata for database |
| ListTables | 按 schema=main/alias 过滤 | qualified by database | by database + schema |

SQLite attach 逻辑保留在 `internal/driver/sqlite/` 内；MySQL/PG 实现在各自 `namespaces.go`。

### 4.3 Wails 绑定 + 测试

- `SchemaService.ListNamespaces` / `ListSchemas` / `ListTables` 扩展
- 单元 + integration：三方言至少 smoke

---

## 5. catalog.db 与 ConnectionGroupService

> 完整 Schema 见 [CONNECTION_GROUPS_AND_STORAGE.md](./CONNECTION_GROUPS_AND_STORAGE.md)

### 5.1 路径

`~/.data-nexus/catalog.db` — `config.CatalogDBPath()`

### 5.2 核心表

- `connections` — 连接配置（`config_json`，无 password）
- `connection_groups` — 嵌套分组（含 `parent_id`、`sort_order`）
- `group_members` — group_id + member_type + member_id + sort_order
- `free_connections` — 游离连接 + sort_order
- `sidebar_root_items` — 根层 Group 与游离连接混合排序
- `app_connection_state` — restore_open_on_startup、open_connection_ids

### 5.3 ConnectionGroupService API

```go
func (s *ConnectionGroupService) GetSidebarTree() (*model.ConnectionSidebarTree, error)
func (s *ConnectionGroupService) CreateGroup(parentID, name string) (*ConnectionGroup, error)
func (s *ConnectionGroupService) RenameGroup(id, name string) (*ConnectionGroup, error)
func (s *ConnectionGroupService) CountConnectionsInGroup(id string) (int, error)
func (s *ConnectionGroupService) DeleteGroup(req DeleteGroupRequest) error
func (s *ConnectionGroupService) MoveGroup(req MoveGroupRequest) error
func (s *ConnectionGroupService) MoveConnectionToGroup(connectionID, groupID string, sortOrder int) error
func (s *ConnectionGroupService) ReleaseConnection(connectionID string, sortOrder int) error
func (s *ConnectionGroupService) ReorderGroupMembers(groupID string, ordered []GroupMemberRef) error
func (s *ConnectionGroupService) ReorderSidebarRoot(ordered []SidebarRootItemRef) error
```

**`MoveGroup` 错误：**

- `newParentId` 与被移动 Group 的 `id` 相同 → `INVALID_REQUEST`（`cannot move group into itself`）
- `newParentId` 为目标子孙 Group → `INVALID_REQUEST`（`cannot move group into its descendant`）
- 实现：`internal/catalog/sqlite/groups.go` `ensureNotDescendantLocked`；单元测试见 `store_test.go` `TestStoreMoveGroupIntoItself` / `TestStoreMoveGroupIntoDescendant`

Wails 绑定：`ConnectionGroupService.GetSidebarTree` 等（见 `internal/wails/connection_group.go`）。

### 5.4 首次启动

若不存在 `catalog.db` → 创建 schema + 默认顶层 Group「我的连接」。**不**读取或导入 `connections.json`。

### 5.5 废弃

- 移除 `config.ConnectionsPath()`、`connection_store.go` JSON 读写
- `model.ConnectionsFile` 逐步移除

---

## 6. 连接配置模型扩展

> 完整字段矩阵见 [CONNECTION_FORM.md §3 & §6](./CONNECTION_FORM.md)

### 6.1 Go 模型

**文件:** `internal/model/connection.go`

```go
type PostgresConfig struct {
    // ... existing ...
    ClientEncoding string `json:"clientEncoding,omitempty"` // default "UTF8"
}

type MySQLConfig struct {
    // ... existing ...
    Charset              string `json:"charset,omitempty"`              // default "utf8mb4"
    Collation            string `json:"collation,omitempty"`            // default "utf8mb4_unicode_ci"
    DefaultStorageEngine string `json:"defaultStorageEngine,omitempty"` // default "InnoDB"
}
```

`PostgresSettingsUpdate` / `MySQLSettingsUpdate` 同步新增字段。

### 6.2 密码更新语义

`UpdateConnectionPostgresSettings` / `UpdateConnectionMySQLSettings`：

- `password` 字段为 **空字符串** 或 **JSON omitted** → **跳过密码更新**，保留 Vault 中现有值
- 新建连接时 password **仍必填**（与 [CONNECTION_FORM §4.3](./CONNECTION_FORM.md#43-校验规则) 一致）
- password **不进** `catalog.db`（SECRETS 不变）

### 6.3 MySQL Driver

- `buildMySQLDSN`: 使用配置的 `charset` / `collation`，空则回退 `utf8mb4` / `utf8mb4_unicode_ci`
- `Connect` 成功后: 若 `DefaultStorageEngine != ""`，执行 `SET SESSION default_storage_engine = ?`

### 6.4 PostgreSQL Driver

- DSN query 增加 `client_encoding`（来自 `ClientEncoding`，默认 UTF8）
- `sslmode` 支持 libpq 全量值（disable / allow / prefer / require / verify-ca / verify-full）

### 6.5 ConnectionStore

`UpdatePostgresSettings` / `UpdateMySQLSettings` 合并新字段；写入 **`catalog.db`**。

### 6.6 测试

| 文件 | 用例 |
|------|------|
| `model/connection_test.go` | 默认值 normalization |
| `mysql/dsn_test.go` | 自定义 charset/collation 出现在 DSN |
| `connection_store_test.go` | 读写新字段；password 留空不覆盖 Vault |

---

## 7. AppService 扩展（About + Reveal）

### 7.1 GetPlatform

供 `AboutDialog` 展示平台信息：

```go
func (s *AppService) GetPlatform() string {
    return runtime.GOOS + "/" + runtime.GOARCH  // e.g. "darwin/arm64"
}
```

### 7.2 RevealFileInExplorer

供 SQLite attach L1 节点右键「在 Finder 中显示」：

```go
func (s *AppService) RevealFileInExplorer(filePath string) error
```

| 平台 | 实现 |
|------|------|
| macOS | `open -R <path>` |
| Windows | `explorer /select,<path>` |
| Linux | `xdg-open` 父目录或等效 |

文件不存在 → 返回 `INVALID_REQUEST`；前端 Toast 错误。

---

## 8. HTTP 模式（非 v0.5）

[API.md](../../design/API.md) 已预留 REST 映射:

```
GET /api/v1/sql-executions  →  ListAllSqlExecutions
```

v0.6 `--api` 实施时再添加 HTTP handler；v0.5 仅 Wails 绑定。

---

## 9. 实施检查清单

- [x] `Store.ListAllExecutions` 接口 + SQLite 实现
- [x] `SqlExecutionService.ListAllSqlExecutions`
- [x] Wails `ListAllSqlExecutions` 绑定
- [x] **`Driver.ListNamespaces` / `ListSchemas`** + `SchemaService` 扩展
- [x] **`ListTables` 扩展** + Wails 绑定
- [x] **`GetSidebarTree`** 返回 `rootItems` + Group `memberItems`
- [x] **`AppService.GetPlatform`** + **`RevealFileInExplorer`**
- [x] 单元测试通过
- [x] **`app_menu_darwin.go`** macOS 系统菜单 + EventsEmit
- [x] **`app_menu.go`** Win/Linux 精简菜单
- [x] **`handleFileDrop`** → `app:file-drop`
- [x] `wails generate` 更新前端绑定
