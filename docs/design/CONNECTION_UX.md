# Data Nexus — 连接与主界面交互（Navicat 模式）

> 版本: v0.5 · 最后更新: 2026-06-08  
> **v0.5–v0.8 实施:** [phase-v0.5.md](../implementation/phase-v0.5.md) · [phase-v0.5.1.md](../implementation/phase-v0.5.1.md) · [phase-v0.6.md](../implementation/phase-v0.6.md) · [phase-v0.7.md](../implementation/phase-v0.7.md) · [phase-v0.8.md](../implementation/phase-v0.8.md)

---

## 1. 设计原则

| 原则 | 说明 |
|------|------|
| 启动即主界面 | 无 Welcome 门闸；冷启动直接进入工作台 |
| 连接列表常驻 | 左侧展示全部已保存连接及打开状态 |
| 多库并存 | 可同时打开多个 SQLite 文件，互不影响 |
| 用户主动打开 | 默认不自动连接；双击或右键「打开」才建立会话 |
| 类 Navicat | 连接 → 展开 Schema 树 → 主区 Tab 操作 |

---

## 2. 启动流程

```
启动应用
  → 打开 catalog.db（不存在则创建 schema + 默认 Group）
  → 加载 Group 树 + 连接列表
  → 显示主界面
  → 不自动 OpenConnection（除非用户开启「启动恢复已打开连接」）
```

**P1 可选：** 退出时记录 `openConnectionIds`，下次启动自动 `OpenConnection` 恢复（设置中开关）。

---

## 3. 主界面布局（v0.5）

主界面为 **左连接列表 + 右展示区**；全局操作（新建连接、设置等）在顶栏 MenuBar，不在左侧。

```
┌─────────────────────────────────────────────────────────────────┐
│ Data Nexus   [文件▾] [视图▾] [帮助▾]        已打开 N 个连接    │
├──────────────────────┬──────────────────────────────────────────┤
│ 连接列表              │  [ 表结构 | 数据 | SQL ]    app.db ▾   │
│ ────────────────────  ├──────────────────────────────────────────┤
│ ▼ ● app.db           │                                          │
│     Tables            │         Main Content                     │
│       users           │                                          │
│       orders          │                                          │
│ ○ staging.sqlite3     │                                          │
│ [SavedQueries]        │                                          │
├──────────────────────┴──────────────────────────────────────────┤
│ StatusBar                                                       │
└─────────────────────────────────────────────────────────────────┘
```

| 符号 | 含义 |
|------|------|
| `●` | 已打开（活跃连接） |
| `○` | 已保存但未打开 |
| `▼/▶` | 展开/折叠 Schema 树（仅已打开连接可展开） |

---

## 4. 用户操作

| 操作 | 行为 |
|------|------|
| **新建连接** | MenuBar → 文件 → 新建连接 → Dialog → 保存到列表 → 可选立即打开 |
| **打开 SQLite** | MenuBar → 文件 → 打开 SQLite → 选文件 → 保存并立即 open（快速通道） |
| **新建顶层分组** | MenuBar → 文件 → 新建分组；或侧边栏根层空白右键 |
| **打开**（○ 连接） | 双击连接 item，或右键 → 打开 → `OpenConnection(savedId)` |
| **关闭**（● 连接） | 右键 → 关闭，或 MenuBar → 文件 → 关闭当前连接 |
| **删除** | 右键 → 删除 → `RemoveConnection`（打开中需 confirm） |
| **重命名 / 编辑** | 右键 → **重命名**（inline，随时）或 **编辑连接**（closed，含显示名称 + 配置） |
| **单击 Group** | 设置 `selectedGroupId`（新建连接/Cmd+O 入组目标）并高亮 |
| **单击连接** | 选中/高亮；**保留** `selectedGroupId` |
| **双击 ○ 或 ● 连接** | 打开：未打开则 `OpenConnection`；已打开则设为当前 `activeConnectionId` |
| **单击连接** | 仅选中/高亮，不 open |
| **点击表名** | 主区切到「数据」Tab，上下文为当前连接 |
| **SQL Tab** | 绑定当前选中的连接 ID；切换连接时切换编辑器上下文 |

---

## 5. v0.5 连接列表与右键菜单

v0.5 起，连接 item **不再** 显示行内「打开 / 关闭 / 编辑 / 删除」按钮；改为 **右键 ContextMenu**。

| 菜单项 | 显示条件 | API / 行为 |
|--------|----------|------------|
| 打开连接 | `status !== 'open'` | `OpenConnection` |
| 关闭连接 | `status === 'open'` | `CloseConnection` |
| **重命名** | 始终 | 树内 **inline** 编辑显示名（`RenameConnection`）；**open / closed 均可** |
| 编辑连接 | 始终显示；**仅** `status !== 'open'` 可用 | `EditConnectionDialog` + `ConnectionForm`（含 **显示名称** + 全部配置） |
| 删除 | 始终（danger） | `RemoveConnection` |

**键盘：** 侧边栏内最后一次单击选中的 Group 或连接，按 **F2** 或 **Enter** 进入 **inline 重命名**（Dialog 打开或 Monaco 焦点时忽略）。

**显示名称 — 双入口（并存）：** inline（F2/Enter/右键「重命名」）随时改显示名；**编辑连接** Dialog（**仅 closed**）可同时改显示名与连接配置。两条路径均调用 `RenameConnection`。

**编辑连接配置：** 仅 **连接关闭** 时可编辑；已 open 时右键「编辑连接」**disabled**（Tooltip：`connection.editRequiresClosed`）。

**自左侧移除：** 「新建连接」按钮、SQLite `readOnly`/`wal` 勾选（迁入 `NewConnectionDialog`）、「启动时恢复已打开连接」（迁入 `SettingsDialog`）。

---

## 6. 与旧方案差异

| 旧方案 | 新方案 |
|--------|--------|
| Welcome 页 | 取消 |
| 单活跃连接 | **多活跃连接** map |
| Connect 前先 Disconnect | Open 互不影响 |
| 启动自动连上次库 | 默认不自动；列表中选择打开 |
| v1.0 才多连接 | **MVP 即多连接** |

---

## 7. 远程连接（v1.0）

### 7.1 新建连接 Tab

`NewConnectionDialog`：**SQLite | PostgreSQL | MySQL**

- SQLite：文件路径 + 浏览（不变）
- PostgreSQL：host / port / **database（必填）** / user / password / SSL / 只读
- MySQL：host / port / **database（可选）** / user / password / TLS / 只读
- **[测试连接]**：临时 Connect+Ping，不触发 vault
- **[保存并打开]**：写入 `catalog.db` + Secrets + 可选立即 Open

### 7.2 Vault 按需解锁

当 Secrets backend 为 **vault** 时：

- **不**在应用启动时弹窗
- 打开远程连接 / 保存远程连接 / 编辑密码 → 若 locked → `VaultDialog`（Init 或 Unlock）→ retry 原操作
- 同会话解锁后打开多个远程连接不再重复询问

详见 [SECRETS.md](./SECRETS.md)。

### 7.3 连接树展示

| 类型 | 标签 | 副标题 |
|------|------|--------|
| SQLite | SQL | 文件路径 |
| PostgreSQL | PG | `user@host:port/database` |
| MySQL | MY | `user@host:port/database`（database 为空时省略 `/database`） |

### 7.4 Schema / Database 浏览（v0.5 层级树）

v0.5 起，database/schema **不再**用手输框切换，改为连接列表内 **可展开层级树**。详见 [v0.5/CONNECTION_TREE.md](../implementation/v0.5/CONNECTION_TREE.md)。

| 类型 | 树层级 |
|------|--------|
| SQLite | 连接 → `main` / attach alias → 表 |
| MySQL | 连接 → database → 表 |
| PostgreSQL | 连接 → database → schema → 表 |

**Attach：** 仅 SQLite；attach 库为 L1 节点，与 `main` 同级。入口：连接 L0 右键「附加数据库…」、拖拽 `.db` 到已 open SQLite 连接。L1 attach 节点右键：**取消附加**、**在 Finder/文件管理器中显示**（`RevealFileInExplorer`）。

### 7.5 连接失败文案

`CONNECTION_FAILED` + `details.reason`：

| reason | 用户文案 |
|--------|----------|
| `network` | 网络连接失败 |
| `auth` | 认证失败 |
| `database` | 数据库不存在 |

---

## 8. 相关文档

- API：[API.md](./API.md) §3–§7（均带 `connectionId`）
- 密码存储：[SECRETS.md](./SECRETS.md)
- 数据模型：[DATA_MODEL.md](./DATA_MODEL.md) §2
- UI 组件：[UI_UX.md](./UI_UX.md)
