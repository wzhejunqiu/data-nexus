# Data Nexus — 连接与主界面交互（Navicat 模式）

> 版本: v1.0 · 最后更新: 2026-06-07

---

## 1. 设计原则

| 原则 | 说明 |
|------|------|
| 启动即主界面 | 无 Welcome 门闸；冷启动直接进入工作台 |
| 连接列表常驻 | 左侧展示全部已保存连接及打开状态 |
| 多库并存 | 可同时打开多个 SQLite 文件，互不影响 |
| 用户主动打开 | 默认不自动连接；用户点击「打开」才建立会话 |
| 类 Navicat | 连接 → 展开 Schema 树 → 主区 Tab 操作 |

---

## 2. 启动流程

```
启动应用
  → 加载 connections.json
  → 显示主界面（连接列表 + 空主区或上次 Tab 布局）
  → 不自动 OpenConnection（除非用户开启「启动恢复已打开连接」）
```

**P1 可选：** 退出时记录 `openConnectionIds`，下次启动自动 `OpenConnection` 恢复（设置中开关）。

---

## 3. 主界面布局

```
┌─────────────────────────────────────────────────────────────────┐
│ Data Nexus          [+ 新建连接]  [主题] [语言]                  │
├──────────────────────┬──────────────────────────────────────────┤
│ 连接                  │  [ 表结构 | 数据 | SQL ]    app.db ▾   │
│ ────────────────────  ├──────────────────────────────────────────┤
│ ▼ ● app.db           │                                          │
│     Tables            │         Main Content                     │
│       users           │                                          │
│       orders          │                                          │
│     Views             │                                          │
│       v_active        │                                          │
│ ○ staging.sqlite3     │                                          │
│ ○ analytics.db        │                                          │
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
| **新建连接** | 文件对话框 / 表单 → 保存到列表 → 可选立即打开 |
| **打开**（○ 连接） | `OpenConnection(savedId)`，侧边栏变为 ●，加载 Schema |
| **关闭**（● 连接） | `CloseConnection(id)`，释放 Driver；保存配置保留 |
| **断开并删除** | 关闭 + `RemoveConnection` |
| **双击 ○ 连接** | 打开 |
| **点击表名** | 主区切到「数据」Tab，上下文为当前连接 |
| **SQL Tab** | 绑定当前选中的连接 ID；切换连接时切换编辑器上下文 |
| **重命名** | `RenameConnection`（P1） |

---

## 5. 与旧方案差异

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
- 远程：host / port / database / user / password / SSL / 只读
- **[测试连接]**：临时 Connect+Ping，不触发 vault
- **[保存并打开]**：写入 `connections.json` + Secrets + 可选立即 Open

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
| MySQL | MY | `user@host:port/database` |

### 7.4 Schema / Database 切换

- **PostgreSQL**：侧边栏 schema 输入 + 应用 → `UpdateConnectionPostgresSettings` → 重连 → 刷新表列表
- **MySQL**：database 切换同理
- **Attach**：仅 SQLite 连接显示

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
