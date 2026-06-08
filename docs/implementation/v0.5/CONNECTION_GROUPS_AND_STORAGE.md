# v0.5 连接分组与 SQLite 持久化

> **关联:** [CONNECTION_TREE.md](./CONNECTION_TREE.md) · [DATA_MODEL.md](../../design/DATA_MODEL.md) · [CONNECTION_UX.md](../../design/CONNECTION_UX.md)

v0.5 起，左侧连接列表通过 **可嵌套 Group** 组织连接；连接配置与分组元数据 **统一存入 SQLite**，废弃 `connections.json`。

---

## 1. 设计目标

| 目标 | 说明 |
|------|------|
| 分组管理 | Group 可嵌套；Group 内可含子 Group 与连接（仅引用 `connection_id`） |
| 数据分离 | 连接 **配置** 与 **分组结构** 分表存储；Group 不嵌入连接详情 |
| SQLite 持久化 | 替代 JSON 文件；schema 版本迁移、索引、事务 |
| 默认分组 | 首次启动创建顶层 Group **「我的连接」**（与普通 Group 同级，非全列表包裹层） |
| 游离连接 | 不属于任何 Group 的连接，与顶层 📁 **同级** 展示 |

**非目标：** **不**从 `connections.json` 导入或迁移；v0.5 起直接使用 `catalog.db`（旧 json 数据不自动搬迁）。

---

## 2. 侧边栏第一层结构

连接列表 **第一层**（同一缩进）仅两类节点：

1. **Group（📁）** — 含「我的连接」及用户创建的分组；可嵌套子 Group
2. **游离连接（○/●）** — 不属于任何 Group 的连接

```
┌─ 连接列表（第一层同级）───────────────────────────────────┐
│ ▼ 📁 我的连接              ← 默认顶层 Group（非全列表根）   │
│     ▼ 📁 生产环境                                           │
│         ▼ ● prod-mysql (MY)                                 │
│             ▼ myapp → users                                 │
│     ● app.db               ← Group 内连接                   │
│ ▼ 📁 归档                  ← 另一顶层 Group（parent_id 空） │
│ ○ staging.db               ← 游离连接，与 📁 我的连接 同级  │
│ ○ old.sqlite               ← 游离连接                       │
│ [SavedQueries]                                              │
└─────────────────────────────────────────────────────────────┘
```

> `[SavedQueries]` — **v0.3 已有功能，v0.5 保持不变**（侧边栏底部 canned queries 区块）。

**重要：** 「我的连接」**不是**包裹整个列表的根容器；**游离连接也不在被删 Group 的父 Group 里**，而是提升到 **连接列表第一层**，与「我的连接」**UI 同级**。

| 概念 | 说明 |
|------|------|
| **第一层 Group** | `connection_groups.parent_id IS NULL`；至少含默认「我的连接」 |
| **嵌套 Group** | `parent_id` 指向某 Group |
| **Group 内连接** | 出现在 `group_members` 中 |
| **游离连接** | **不在**任何 `group_members`；见 §4.2 `free_connections` |
| 库/表子树 | 连接展开后 namespace → 表（见 CONNECTION_TREE.md） |

---

## 3. Group 模型

### 3.1 规则

- 每个 Group：`id`（ULID）、`name`、可选 `parent_id`（`NULL` = **顶层 Group**，与「我的连接」同级）
- **「我的连接」**：**首次启动** catalog.db 时创建的 **默认顶层 Group**；名称 **随时可改**；删除规则与普通 Group 相同
- 子 Group：`parent_id` 指向父 Group（可嵌套）
- Group **不可循环**引用
- 连接 **要么** 在某一 Group 的 `group_members` 中，**要么** 为 **游离连接**（二者互斥）
- **重命名 Group：** **随时**可改
- **删除 Group：** 见 §3.4

### 3.2 排序字段（sort_order）

所有 Group 与连接在 **同级兄弟** 间均有 `sort_order`（整数，默认按创建顺序递增）。用户 **拖拽** 后批量更新序号。

| 存储 | sort_order 作用域 |
|------|-------------------|
| `connection_groups.sort_order` | 同一 `parent_id` 下的 **Group 兄弟**（含顶层 `parent_id IS NULL`） |
| `group_members.sort_order` | 同一 Group 内的 **子 Group + 连接** |
| `free_connections.sort_order` | 游离连接之间 |
| `sidebar_root_items.sort_order` | **根层** 顶层 Group 与游离连接的 **混合排序**（见 §4.2） |

新建 Group / 连接时：`sort_order = max(同级) + 1`（或末尾插入）。

### 3.3 操作（CRUD / 移动）

| 操作 | 行为 |
|------|------|
| 新建顶层 Group | `parent_id = NULL`；与「我的连接」同级 |
| 新建子 Group | `parent_id = 当前选中 Group` |
| **重命名 Group** | **原地 inline** 编辑（见 §3.6）；非 Dialog |
| **删除 Group** | 见 §3.4 |
| 移入 Group | 从 `free_connections` 删除 → 写入 `group_members` |
| **移为游离** | 从 `group_members` 删除 → 写入 `free_connections` |
| 新建连接 | 选中某 Group → 入该 Group；**未选中 Group** → **游离连接**（第一层） |

### 3.4 删除 Group（递归校验）

**步骤 1 — 递归统计：** 收集待删 Group **自身及全部子孙 Group** 内的 **连接 ID**（去重，不含子 Group 本身）。

**步骤 2 — 确认 UI：**

| 递归连接数 | 对话框 |
|------------|--------|
| **0** | 简单确认：「删除分组「{{name}}」及其子分组？」 |
| **≥ 1** | 扩展确认（见下） |

**含连接时的删除对话框：**

```
┌─ 删除分组 ─────────────────────────────────────────┐
│  将删除分组「生产环境」及其 N 个子分组。              │
│  内含 M 个连接（含子分组）。                          │
│                                                     │
│  ☐ 同时删除这 M 个连接   ← 默认 **不勾选**          │
│                                                     │
│  未勾选时：连接将变为 **游离连接**，与分组（如「我的连接」）同级。 │
│                                                     │
│              [ 取消 ]  [ 删除 ]                      │
└─────────────────────────────────────────────────────┘
```

**步骤 3 — 执行逻辑（事务）：**

| `deleteConnections` | 行为 |
|---------------------|------|
| **false**（默认） | 1. 递归收集的连接：从 `group_members` **移除** → 插入 **`free_connections`**（**第一层游离**，与「我的连接」同级）<br>2. 删除该 Group 及全部子孙 Group<br>3. 清理相关 `group_members` 中 `member_type=group` 行 |
| **true** | 1. 对每个递归连接 `RemoveConnection`（已 open 先 close）<br>2. 删除 Group 子树 |

**示例：** 删除「我的连接」下的「生产环境」（含 2 条连接）、不删连接 → 2 条连接出现在 **第一层**，与剩余顶层 📁 **并列**（**不会**自动进入「我的连接」或任何父 Group）。

**嵌套 Group 内连接：** 删除任意 Group 时，未勾选「删除连接」则 **一律变为游离连接**（第一层），**与**被删 Group 所在层级 **无关**。

### 3.5 默认「我的连接」

- 首次启动 catalog.db：若不存在默认顶层 Group，创建 **「我的连接」**（`parent_id = NULL`）
- i18n：`connectionGroup.defaultName` / `"My Connections"`
- 与普通顶层 Group 相同：可重命名、可删除、可拖动

### 3.6 Group / 连接交互（右键 · 拖拽）

#### 3.6.1 Group 原地重命名

- **不弹出 Dialog**；在树节点 **原地** 变为 `<input>`
- 触发：右键「重命名」、**F2**、或慢双击标题（可选）
- **Enter** → 调用 `RenameGroup(id, name)` 持久化并退出编辑
- **Esc** → 取消，恢复原名
- 空名称：拒绝保存，保持编辑态或回滚

```
▼ [ 生产环境____ ]   ← 编辑态 inline input
```

#### 3.6.2 Group 右键菜单

| 菜单项 | 行为 |
|--------|------|
| **新建** | 在该 Group **下** 创建子 Group（inline 新建或默认名「新分组」并进入重命名态） |
| **重命名** | 进入 §3.6.1 原地编辑 |
| **删除** | §3.4 删除流程（含连接时带 checkbox） |

顶层 Group 与嵌套 Group **同一套** 右键菜单。根层空白处右键可选「新建分组」（创建 `parent_id NULL` 的顶层 Group）。

#### 3.6.3 连接右键菜单

连接 item **无行内按钮**；`打开` / `关闭` 按状态 **互斥显示**（未 open 仅「打开连接」，已 open 仅「关闭连接」）。

**PG / MySQL（四项）：**

| 菜单项 | 显示条件 | 行为 |
|--------|----------|------|
| **打开连接** | `status !== 'open'` | `OpenConnection(id)` |
| **关闭连接** | `status === 'open'` | `CloseConnection(id)` |
| **编辑连接** | 始终显示；**仅** `status !== 'open'` 时可点 | 打开 `EditConnectionDialog`（含 **显示名称** 与全部连接配置）；已 open 时 **disabled**，Tooltip「请先关闭连接」 |
| **删除** | 始终（danger） | `RemoveConnection`；已 open 需 confirm |

**SQLite 额外第五项：**

| 菜单项 | 显示条件 | 行为 |
|--------|----------|------|
| **附加数据库…** | `type === 'sqlite'`；**仅** `status === 'open'` 时可点 | 打开 `AttachDatabaseDialog` |
| （未 open 时） | `type === 'sqlite'` **且** `status !== 'open'` | **disabled** + Tooltip「请先打开连接」 |

**显示名称（重命名）：** **不**单独提供右键「重命名」或树内 inline 编辑；与 host/端口等 **一并** 在「编辑连接」Dialog 的「显示名称」字段修改，保存时随 `UpdateConnection*` 持久化。

**边界：** 连接树中所有节点均来自 `catalog.db` 持久化记录；新建连接在 Dialog 保存成功后才出现在树中，**无**「未保存连接」右键场景。

**移入/移出 Group** 不放在右键菜单，由 **拖拽**（§3.6.5）完成。

**编辑连接约束：** 仅 **连接已关闭** 时可编辑（含 displayName 与连接配置）；后端 `UpdateConnection*` 若连接处于 open 状态 → 返回 `CONNECTION_OPEN`（前端不应打开 Dialog）。

Attach/Detach 完整 UI 见 [CONNECTION_TREE.md §3.6](./CONNECTION_TREE.md#36-sqlite-attach--detach)。

#### 3.6.4 拖拽 — 改变 Group 父层级

- **拖 Group → 另一 Group 上**：成为目标 Group 的 **子 Group**（更新 `parent_id`；写入目标 `group_members`；不可拖入自身或子孙）
- **拖 Group → 根层空白/根 drop 区**：`parent_id = NULL`，进入顶层（与「我的连接」同级）
- **同级排序**：在同一父下拖动 → 仅更新 `sort_order` / `group_members.sort_order`

#### 3.6.5 拖拽 — 连接移入 Group / 游离

- **拖连接 → Group 上**：`MoveConnectionToGroup(connectionId, groupId)`；从 `free_connections` 或原 `group_members` 移出
- **拖连接 → 根层空白（非 Group）**：`ReleaseConnection` → **游离连接**
- **同级排序**：同一 Group 内或根层游离区拖动 → 更新 `sort_order`

#### 3.6.6 拖拽 — 根层混合排序

根层展示顺序 = `sidebar_root_items`（顶层 Group + 游离连接 **交错**）。拖动后调用 `ReorderSidebarRoot(ordered[])` 重写序号。

#### 3.6.7 实现要点（前端）

- DnD 库：`@dnd-kit/core` + `@dnd-kit/sortable`（或项目内等价方案）
- Drop 指示：高亮目标 Group；根层显示插入线
- 拖拽中禁用 Group **inline 重命名**

---

## 4. SQLite 存储

### 4.1 文件位置

```
~/.data-nexus/catalog.db
```

与 `sql-global.db`（执行历史）、`config.yaml`、`vault/` 并列。路径经 `config.CatalogDBPath()` 暴露。

**驱动：** `modernc.org/sqlite`（与 execution log 一致，纯 Go）。

### 4.2 Schema（v1）

```sql
PRAGMA user_version = 1;

-- 已保存连接
CREATE TABLE connections (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    driver_type      TEXT NOT NULL,  -- sqlite | postgres | mysql
    config_json      TEXT NOT NULL,  -- DriverConfig JSON，无 password
    secrets_backend  TEXT,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL,
    last_used_at     TEXT NOT NULL
);

CREATE INDEX idx_connections_last_used ON connections (last_used_at DESC);

-- 分组（parent_id NULL = 顶层，与「我的连接」同级）
CREATE TABLE connection_groups (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    parent_id  TEXT REFERENCES connection_groups(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_connection_groups_parent ON connection_groups (parent_id);
CREATE INDEX idx_connection_groups_sort ON connection_groups (parent_id, sort_order);

-- 分组成员（仅 Group 内部：子 group + 连接）
CREATE TABLE group_members (
    group_id     TEXT NOT NULL REFERENCES connection_groups(id) ON DELETE CASCADE,
    member_type  TEXT NOT NULL,  -- 'group' | 'connection'
    member_id    TEXT NOT NULL,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, member_type, member_id)
);

CREATE INDEX idx_group_members_sort ON group_members (group_id, sort_order);

-- 游离连接（第一层，与顶层 Group 同级；不在任何 group_members 中）
CREATE TABLE free_connections (
    connection_id TEXT PRIMARY KEY REFERENCES connections(id) ON DELETE CASCADE,
    sort_order    INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_free_connections_sort ON free_connections (sort_order);

-- 根层混合排序：顶层 Group + 游离连接
CREATE TABLE sidebar_root_items (
    item_type  TEXT NOT NULL,  -- 'group' | 'connection'
    item_id    TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    PRIMARY KEY (item_type, item_id)
);

CREATE INDEX idx_sidebar_root_sort ON sidebar_root_items (sort_order);

-- 应用级连接状态（原 ConnectionsFile 顶层字段）
CREATE TABLE app_connection_state (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
-- keys: restore_open_on_startup (bool JSON), open_connection_ids (JSON array)
```

**密码：** 仍 **不** 写入 `catalog.db`；远程密码走 Keychain / Vault（[SECRETS.md](../../design/SECRETS.md)）。

### 4.3 连接、Group、游离 三者关系

```
connections 表          group_members              free_connections
┌─────────────┐        ┌────────────────────┐     ┌──────────────────┐
│ id: c1      │◄───────│ group: 我的连接      │     │                  │
└─────────────┘        │ member: c1         │     │                  │
┌─────────────┐        └────────────────────┘     │ connection: c2   │
│ id: c2      │◄────────────────────────────────│ sort_order       │
└─────────────┘                                   └──────────────────┘
  c1 ∈ Group「我的连接」     c2 为游离连接（第一层）
```

删除连接：`connections` + 清理 `group_members` + 清理 `free_connections`。

### 4.4 Store 层

```
internal/catalog/
├── store.go           # CatalogStore 接口
├── sqlite/
│   ├── store.go       # 实现 + migrate
│   ├── connections.go
│   ├── groups.go
│   └── migrate.go
└── sqlite/store_test.go
```

**替代：** `internal/service/connection_store.go` 的 JSON 读写 → 委托 `CatalogStore`。

### 4.5 首次启动初始化

```
1. 若不存在 catalog.db → 创建文件 + 执行 schema（PRAGMA user_version）
2. 若不存在默认顶层 Group → 插入「我的连接」（parent_id NULL）
3. 写入 sidebar_root_items（默认 Group 排序）
```

**不读取、不导入** `connections.json`；旧版 json 中的连接 **不会** 自动出现在 v0.5 列表中。

---

## 5. API（Wails / Service）

### 5.1 ConnectionGroupService（新增）

```go
func (s *ConnectionGroupService) GetSidebarTree() (*model.ConnectionSidebarTree, error)
func (s *ConnectionGroupService) CreateGroup(parentID, name string) (*ConnectionGroup, error)
func (s *ConnectionGroupService) RenameGroup(id, name string) (*ConnectionGroup, error)

// CountConnectionsInGroup 递归统计 Group 子树内连接数（不含子 Group 自身）
func (s *ConnectionGroupService) CountConnectionsInGroup(id string) (int, error)

type DeleteGroupRequest struct {
    ID                string `json:"id"`
    DeleteConnections bool   `json:"deleteConnections"` // 默认 false
}
func (s *ConnectionGroupService) DeleteGroup(req DeleteGroupRequest) error

func (s *ConnectionGroupService) MoveGroup(req MoveGroupRequest) error
// NewParentID 空 = 顶层；SortOrder 为目标兄弟位置

func (s *ConnectionGroupService) MoveConnectionToGroup(connectionID, groupID string, sortOrder int) error
func (s *ConnectionGroupService) ReleaseConnection(connectionID string, sortOrder int) error

func (s *ConnectionGroupService) ReorderGroupMembers(groupID string, ordered []GroupMemberRef) error
func (s *ConnectionGroupService) ReorderSidebarRoot(ordered []SidebarRootItemRef) error
```

**GetSidebarTree 返回结构：**

```typescript
interface ConnectionSidebarTree {
  /** 顶层 Group（parent_id 空），各节点可嵌套 */
  groups: ConnectionGroupNode[]
  /** 游离连接 — 与 groups 第一层同级 */
  freeConnections: ConnectionListItem[]
}

interface ConnectionGroupNode {
  id: string
  name: string
  childGroups: ConnectionGroupNode[]
  connections: ConnectionListItem[]  // 仅本 Group 直属连接
}
```

前端 `ConnectionTree` 根级渲染：

```tsx
{tree.groups.map((g) => <ConnectionGroupNode key={g.id} node={g} />)}
{tree.freeConnections.map((c) => <ConnectionTreeItem key={c.id} item={c} depth={0} />)}
```

连接节点下再挂 `ConnectionSchemaTree`（见 CONNECTION_TREE.md）。

### 5.2 ConnectionService 变更

- `CreateRemote` / `UpsertSQLite`：`groupId` 有值 → `group_members`；**无** → `free_connections`（游离）
- 删除连接：清理 `group_members` + `free_connections`

---

## 6. 前端

### 6.1 组件

```
features/connection/
├── ConnectionTree.tsx
├── ConnectionGroupNode.tsx      # inline 重命名 + 右键菜单 + draggable
├── ConnectionTreeItem.tsx       # 右键菜单 + draggable（无 inline 重命名）
├── dnd/
│   ├── SidebarDndContext.tsx    # @dnd-kit 根 context
│   ├── useGroupDragDrop.ts
│   └── useConnectionDragDrop.ts
├── DeleteGroupDialog.tsx
└── tree/ConnectionSchemaTree.tsx
```

### 6.2 交互摘要

| 能力 | Group | 连接 |
|------|-------|------|
| 原地重命名 | Enter 保存；**无 Dialog** | **无**；displayName 在 **编辑连接** Dialog |
| 右键 | **新建 / 重命名 / 删除** | **打开连接 / 关闭连接 / 编辑连接 / 删除** |
| 编辑 | — | **仅 closed**；含显示名称 + 连接配置 |
| 拖改父级 | 拖到 Group 上 / 根层 | 拖到 Group 上 / 根层游离 |
| 拖改顺序 | 同级 `sort_order` | 同级 `sort_order` |

### 6.3 Query

```typescript
useQuery({ queryKey: ['connectionSidebarTree'], queryFn: groupApi.getSidebarTree })
// 连接 open 状态变更时 invalidate
```

---

## 7. 废弃与兼容

| 废弃 | 替代 |
|------|------|
| `~/.data-nexus/connections.json` | **废弃**；v0.5 **不迁移**，改用 `catalog.db` |
| `ConnectionStore` JSON 读写 | `catalog/sqlite.Store` |
| `model.ConnectionsFile` | DB 表 + `GetSidebarTree` API |
| `config.ConnectionsPath()` | **移除**；统一 `config.CatalogDBPath()` |

**文档/代码引用：** PRD、SECRETS、API 等全局文档在 v0.5 实施时批量改为 catalog.db 表述；密码边界不变。

---

## 8. i18n

| Key | zh-CN |
|-----|-------|
| `connectionGroup.defaultName` | 我的连接 |
| `connectionGroup.new` | 新建 |
| `connectionGroup.rename` | 重命名 |
| `connectionGroup.delete` | 删除分组 |
| `connectionGroup.deleteSimpleConfirm` | 删除分组「{{name}}」及其子分组？ |
| `connectionGroup.deleteWithConnections` | 将删除分组「{{name}}」及其 {{subCount}} 个子分组。内含 {{connCount}} 个连接。 |
| `connectionGroup.deleteConnectionsToo` | 同时删除这 {{connCount}} 个连接 |
| `connectionGroup.deleteConnectionsHint` | 未勾选时，连接将变为游离连接，与分组同级显示。 |
| `connectionGroup.freeConnection` | 游离连接 |
| `connectionGroup.moveTo` | （v0.5 不用菜单；DnD 移入 Group） |
| `connectionGroup.empty` | 此分组暂无连接 |
| `connection.open` | 打开连接 |
| `connection.close` | 关闭连接 |
| `connection.edit` | 编辑连接 |
| `connection.editRequiresClosed` | 请先关闭连接 |
| `connection.delete` | 删除 |

---

## 9. 非目标（v0.5）

- **`connections.json` 数据迁移**
- Group 跨设备同步
- 连接在多个 Group 间 **链接**（alias，非移动）
- 加密 catalog.db 文件

---

## 10. 验收要点

- [ ] **游离连接** 与顶层 📁「我的连接」**第一层同级**（相同缩进）
- [ ] 删除 Group 且不删连接 → 连接变 **游离**，**不**进入父 Group 或「我的连接」
- [ ] Group 名称随时可改
- [ ] 删除空 Group / 含连接 Group（「同时删除连接」默认不勾选）
- [ ] 不勾选删除连接 → **游离连接**（第一层）
- [ ] 勾选 → 连接配置删除
- [ ] 支持无限嵌套子 Group
- [ ] 默认顶层 Group「我的连接」
- [ ] Group 内连接仅引用 ID；编辑连接后树仍正确
- [ ] 首次启动创建 catalog.db + 默认「我的连接」；**无** connections.json 导入
- [ ] catalog.db 无 password 字段
- [ ] Group **原地重命名**：Enter 生效，无 Dialog
- [ ] Group 右键：**新建 / 重命名 / 删除**
- [ ] 连接右键：**打开连接 / 关闭连接 / 编辑连接 / 删除**（**无**单独重命名）
- [ ] **已 open 连接**「编辑连接」disabled；关闭后在 Dialog 内可改 **显示名称** 与配置
- [ ] 拖 Group 改 **父层级**；拖连接进 Group 或根层游离
- [ ] 拖 Group / 连接改 **同级顺序**（`sort_order`）
