# v0.5 连接列表层级树设计

> **关联:** [DESIGN.md](./DESIGN.md) · [CONNECTION_UX.md](../../design/CONNECTION_UX.md) · [API.md](../../design/API.md)

> **Group 组织（上层）：** [CONNECTION_GROUPS_AND_STORAGE.md](./CONNECTION_GROUPS_AND_STORAGE.md) — 连接列表先按 Group 嵌套，其下才是本节 schema 树。

v0.5 连接 **展开后** 的子树由「连接 → 扁平表列表」升级为 **可展开层级树**，统一抽象下兼容 SQLite 单文件、MySQL 多库、PostgreSQL 库内多 schema，并支持 PG/MySQL **同一连接下浏览多个 database**。

---

## 1. 设计目标

| 目标 | 说明 |
|------|------|
| 多方言统一 | 同一套树组件 + 节点类型，按 `DriverType` 决定层级深度 |
| 多 database | MySQL / PG 连接展开后列出服务器上可访问的 database（懒加载） |
| SQLite 兼容 | 单文件视为一个逻辑库 `main`；ATTACH 库作为同级子节点 |
| 取代文本切换 | 移除 v0.4 `RemoteNamespaceSwitch` 手输 database/schema |
| 懒加载 | 展开 database / schema 节点时才拉表列表，避免一次加载全实例 |

---

## 2. 统一树模型

### 2.1 节点类型

```typescript
type TreeNodeKind =
  | 'connection_group'  // 用户分组（可嵌套）
  | 'connection'        // 已保存/已打开连接
  | 'namespace'         // L1：database / SQLite main|attach
  | 'schema'            // L2：仅 PostgreSQL
  | 'table_group'       // TABLES / VIEWS 分组标题
  | 'table'
  | 'view'
```

### 2.2 各方言层级映射

| 方言 | L0 连接 | L1 namespace | L2 schema | L3 叶子 |
|------|---------|--------------|-----------|---------|
| **SQLite** | `app.db` | `main`、attach alias | — | table / view |
| **MySQL** | `prod` | `myapp`、`analytics`… | — | table / view |
| **PostgreSQL** | `prod-pg` | `production`、`staging`… | `public`、`analytics`… | table / view |

**命名约定（UI 文案）：**

- L1 统一称 **「数据库」**（MySQL database / PG database / SQLite 逻辑库）
- L2 仅 PG 显示 **「Schema」**
- SQLite 的 attach alias 在 L1 显示为 alias 名，副标题标注「已附加」

### 2.3 线框（多方言并存）

```
┌─ 连接列表 ────────────────────────────────────────────────┐
│ ▼ ● prod-mysql (MY)          user@host:3306               │
│     ▼ myapp                                               │
│         TABLES (8)                                        │
│           users                                           │
│           orders                                          │
│         VIEWS (1)                                         │
│           active_users                                    │
│     ▶ analytics                                           │
│     ▶ information_schema        ← 系统库，弱化样式        │
│ ▼ ● app.db (SQL)             /Users/.../app.db          │
│     ▼ main                                                │
│         users                                             │
│         posts                                             │
│     ▶ logs (attached)                                     │
│ ○ staging-pg (PG)            user@host:5432               │
│ [SavedQueries]                                            │
└───────────────────────────────────────────────────────────┘
```

> `[SavedQueries]` — **v0.3 已有功能，v0.5 保持不变**（侧边栏底部 canned queries 区块）。

PostgreSQL 展开示例：

```
▼ ● prod-pg (PG)
    ▼ production
        ▼ public
            users
            orders
        ▶ analytics
    ▶ staging
```

---

## 3. 交互规则

### 3.1 连接节点（L0）

| 手势 | 行为 |
|------|------|
| 单击 | 选中连接（高亮），不 open |
| 双击 | 未 open → `OpenConnection`；已 open → 设为 `activeConnectionId` |
| 右键 | 打开/关闭/编辑/删除（SQLite 已 open 时第五项「附加数据库…」，见 §3.6） |
| 点击 ▶/▼ | 展开/折叠；**仅 open 连接**可展开子树 |

未打开连接：显示 `○`，子树隐藏或 disabled，提示「双击打开」。

### 3.2 Namespace / Schema 节点（L1/L2）

| 手势 | 行为 |
|------|------|
| 点击 ▶/▼ | 懒加载子节点（首次展开 fetch API） |
| 单击 | 设为当前 **browse 上下文**（`activeNamespace` / `activeSchema`） |
| 双击 | 展开并设为 browse 上下文 |

**Browse 上下文** 决定右侧 Tab 与 SQL 的默认 namespace；点击表名后进一步绑定 `tableName`。

### 3.3 表 / 视图节点（叶子）

| 手势 | 行为 |
|------|------|
| 单击 | `selectTable(connectionId, tableKey)` → 主区「数据」Tab |

`tableKey` 规则不变（SQLite: `alias.table`；MySQL: 当前 database 下表名；PG: `schema.table` 或带 search_path 限定）。

### 3.4 搜索

- 连接列表顶部 **搜索框** 过滤：连接名、database 名、表名（已展开子树内）
- 前端 filter；大库可后续做服务端 prefix 搜索（v0.5 非必须）

### 3.5 移除 v0.4 控件

| 移除 | 替代 |
|------|------|
| `RemoteNamespaceSwitch` 文本框 + 应用 | 点击树中 database/schema 节点切换上下文 |
| SchemaSubtree 内扁平 `groups` 渲染 | `ConnectionTree` 内嵌 `ConnectionSchemaTree` |
| SchemaSubtree 顶部 Attach 按钮 + `window.prompt` | SQLite 连接 L0 右键「附加数据库…」+ `AttachDatabaseDialog`（§3.6） |

### 3.6 SQLite Attach / Detach

> **API：** v0.3 已有 `ConnectionService.Attach` / `Detach` / `ListAttached`；v0.5 重构 UI。

#### 3.6.1 入口（ContextMenu）

**SQLite 连接 L0 右键（v0.5 必做）** — 在四项基础上增加第五项「附加数据库…」：

| 菜单项 | 条件 | 行为 |
|--------|------|------|
| 附加数据库… | `type === 'sqlite'` **且** `status === 'open'` | 打开 `AttachDatabaseDialog` |
| （未 open） | `type === 'sqlite'` **且** `status !== 'open'` | **disabled** + Tooltip「请先打开连接」 |

PG/MySQL **不出现**此项。`main` L1 **不提供** Attach 入口；无 L1 ghost 行。

**attach 命名空间 L1 节点右键：**

| 菜单项 | 行为 |
|--------|------|
| 取消附加 | danger 样式；见 §3.6.4 Detach |
| 在 Finder 中显示 / 在文件管理器中显示 | `AppService.RevealFileInExplorer(filePath)` |

#### 3.6.2 AttachDatabaseDialog（~480px）

```
┌─ 附加数据库 ─────────────────────────────────────┐
│  文件路径   [ /Users/.../logs.db    ] [ 浏览… ]   │
│  别名       [ logs_________________ ]              │
│             默认 basename 去 .db；不可 main/temp   │
│  ⓘ 附加为当前会话临时挂载，关闭连接后失效           │
│                        [ 取消 ]  [ 附加 ]          │
└────────────────────────────────────────────────────┘
```

- 只读连接：Dialog 顶部提示「附加库将以只读模式挂载」
- alias 实时校验：reserved（`main`/`temp`）、重复、非法字符（对齐 `IsSafeQuotedIdentifier`）
- 成功 → invalidate `['namespaces', connectionId]`、`['tables', connectionId]`、`['attached', connectionId]` → 展开新 alias 节点

#### 3.6.3 拖拽 `.db` 到已 open SQLite 连接（v0.5 必做）

- 触发：`.db` / `.sqlite` / `.sqlite3` 拖到侧边栏 **已 open 的 SQLite 连接 L0**
- 行为：打开 `AttachDatabaseDialog`，**预填文件路径**；alias 默认 basename
- 与窗口级 `handleFileDrop` 区分：空白区 drop 仍为新建连接；连接节点 drop 由前端拦截
- 未 open / 非 SQLite：Toast「请先打开 SQLite 连接」
- 多文件：逐个 Dialog，不 silent 批量 attach

#### 3.6.4 Detach 行为

- 无活跃浏览：**直接 Detach**
- 若 `browseContext.namespace === alias` 或 `selectedTable` 以该 alias 为前缀：**confirm** 后 Detach 并重置 browseContext

#### 3.6.5 L1 节点视觉

- `main`：默认库，无 Detach
- attach 节点：badge「已附加」（`connectionTree.attached`）；hover 显示源文件完整路径

#### 3.6.6 browseContext 联动

- 单击 attach L1 → `browseContext.namespace = alias`
- Attach 成功 **不自动切换** browseContext

#### 3.6.7 i18n（`attach.*`）

| Key | 说明 |
|-----|------|
| `attach.dialogTitle` | 附加数据库 |
| `attach.filePath` | 文件路径 |
| `attach.alias` | 别名 |
| `attach.sessionHint` | 会话级提示 |
| `attach.readOnlyHint` | 只读 attach 提示 |
| `attach.attach` | 附加（按钮） |
| `attach.detach` | 取消附加 |
| `attach.detachConfirm` | Detach 确认文案 |
| `attach.showInFinder` | macOS「在 Finder 中显示」 |
| `attach.showInFileManager` | Win/Linux「在文件管理器中显示」 |
| `attach.dropRequiresOpen` | 拖拽时需先 open |
| `attach.dropNotSQLite` | 非 SQLite 连接 |
| `attach.aliasInvalid` / `aliasReserved` / `aliasDuplicate` | 校验错误 |

---

## 4. 工作区状态

扩展 `workspaceStore`（或等价 context）：

```typescript
interface BrowseContext {
  connectionId: string
  namespace: string      // L1: mysql db / pg database / sqlite main|alias
  schema?: string        // L2: pg schema only
}

interface WorkspaceState {
  activeConnectionId: string | null
  browseContext: BrowseContext | null   // 当前浏览 namespace
  selectedTable: string | null          // tableKey
  // ...
}
```

**切换规则：**

- 切换 `activeConnectionId` → 若新连接已 open，保留上次 browseContext 或回退到连接配置默认库
- 切换 namespace/schema → 清空 `selectedTable` 或保留同表名（若新 namespace 存在）
- 右侧「连接: app.db ▾」下拉仍只列 **已 open 连接**；namespace 由左侧树驱动

---

## 5. 后端 API

### 5.1 新增

```go
// SchemaService — 供树 L1 懒加载
func (s *SchemaService) ListNamespaces(connectionId string) (*NamespaceList, error)

type NamespaceItem struct {
    Name     string `json:"name"`
    Kind     string `json:"kind"`     // "database" | "attached" | "system"
    Default  bool   `json:"default"`  // 连接配置中的默认库
}

type NamespaceList struct {
    Items []NamespaceItem `json:"items"`
}

// PostgreSQL L2
func (s *SchemaService) ListSchemas(connectionId string, database string) (*SchemaList, error)

// ListTables 扩展（向后兼容：无参数 = 当前 browse 上下文 / 连接默认库）
func (s *SchemaService) ListTables(connectionId string, opts ListTablesOptions) (*TableList, error)

type ListTablesOptions struct {
    Database string `json:"database,omitempty"`
    Schema   string `json:"schema,omitempty"`
}
```

### 5.2 各方言实现要点

| 方言 | ListNamespaces | ListSchemas | ListTables |
|------|----------------|-------------|------------|
| SQLite | `[{name:"main", default:true}]` + `ListAttached` | 不适用 | 按 `schema`=main/alias 过滤（已有） |
| MySQL | `SHOW DATABASES` | 不适用 | `USE db` 或 qualified `` `db`.`tbl` `` |
| PostgreSQL | `SELECT datname FROM pg_database WHERE ...` | `information_schema.schemata` | `table_schema` + `table_catalog` 过滤 |

**系统库：** `information_schema`、`mysql`、`performance_schema`、`pg_catalog` 等标记 `kind: "system"`，UI 弱化、默认折叠。

**权限：** 无权限的 database 不出现在列表或显示为 disabled + tooltip。

### 5.3 连接配置与 browse 上下文

- **新建连接** 仍指定**默认 database**（PG/MySQL 连接串所需）
- 打开连接后，树列出**实例上所有可访问 database**，不限于默认库
- 在非常规 database 下查表：Driver 会话需支持 **catalog 切换** 或 **qualified 查询**（实施时选一种，MySQL 常用 USE / 三段名）

---

## 6. 前端组件

```
features/connection/
├── ConnectionTree.tsx           # L0 连接列表 + SavedQueries
├── ConnectionTreeItem.tsx       # ContextMenu（四项）+ draggable
├── tree/
│   ├── ConnectionSchemaTree.tsx # 已 open 连接下挂载
│   ├── NamespaceNode.tsx        # L1
│   ├── SchemaNode.tsx           # L2 PG only
│   ├── TableGroupNode.tsx       # TABLES / VIEWS 分组（可折叠）
│   ├── TreeNodeRow.tsx          # 通用行：缩进、展开箭头、图标
│   └── useConnectionTree.ts     # 懒加载 query keys、展开状态
```

**Query keys 示例：**

```
['namespaces', connectionId]
['schemas', connectionId, database]
['tables', connectionId, database, schema?]
```

**废弃：** `SchemaSubtree.tsx` 逻辑迁入 `ConnectionSchemaTree`；`RemoteNamespaceSwitch` 删除。

---

## 7. SQLite 兼容说明

| 场景 | 树表现 |
|------|--------|
| 单文件无 ATTACH | L1 仅 `main`，其下直接挂表 |
| ATTACH 额外文件 | 每个 alias 为 L1  sibling，与 `main` 同级 |
| 只读 / WAL | 连接级配置，不影响树结构 |
| `tableKey` | `alias.table` 不变，API 兼容 v0.4 |

SQLite **无** L2 schema 层（除非未来 PRAGMA database_list 统一，当前 attach alias 即 L1）。

---

## 8. 与新建连接 Dialog 的关系

- 新建 PG/MySQL 连接时填写的 **database** = 连接默认库 → 树展开后该节点标记 `default`，首次 open 自动展开并选中
- 浏览其他 database **不需要** 修改已保存连接配置（与 v0.4「切换 database 要 UpdateMySQLSettings 重连」不同）

---

## 9. i18n

| Key | 说明 |
|-----|------|
| `connectionTree.namespaces` | 数据库（L1 分组标题，可选） |
| `connectionTree.schemas` | Schema |
| `connectionTree.systemDb` | 系统数据库 |
| `connectionTree.attached` | 已附加 |
| `connectionTree.expandToLoad` | 展开以加载… |
| `connectionTree.openToBrowse` | 双击打开连接以浏览 |

---

## 10. 非目标（v0.5）

- 跨连接 JOIN / 联邦查询
- 树内拖拽排序
- 服务端表名 prefix 搜索（可 P2）
- ER 图从树节点生成（v0.8）

---

## 11. 验收要点

- [ ] MySQL 连接下可见多个 database 节点，展开后出现表
- [ ] PostgreSQL：database → schema → 表 三级
- [ ] SQLite：`main` + attach alias 为 L1；Attach Dialog、连接右键附加、拖拽 attach、Finder 显示
- [ ] 无手输 database/schema 切换框
- [ ] 懒加载：未展开节点不请求 ListTables
- [ ] 点击表名后右侧 Tab 数据正确
