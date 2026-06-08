# v0.5 手动验收清单

> 桌面 UI 重构 · 最后更新: 2026-06-08  
> **依据:** [phase-v0.5.md](../phase-v0.5.md) · [DESIGN.md](./DESIGN.md)

**自动化测试（实施完成后）:**

```bash
make test
make test-integration   # 可选回归
cd frontend && npm test
```

---

## MenuBar 与 Header

- [ ] 顶栏显示 **文件 / 视图 / 帮助** 三个菜单
- [ ] 右侧显示「已打开 N 个连接」（由 `AppShell` 从 `connectionApi.list()` 派生，非后端接口）
- [ ] Header **无** 独立「设置」按钮
- [ ] 导入/导出向导期间 MenuBar 可见；File 相关项 disabled；右侧显示向导标题

### 文件菜单

- [ ] **新建连接…** → 打开 `NewConnectionDialog`
- [ ] **打开 SQLite 文件…** → 文件对话框 → 连接打开并出现在列表
- [ ] **关闭当前连接** → 关闭当前活跃连接；无活跃连接时 disabled
- [ ] **退出**（Win/Linux MenuBar 或 macOS App 菜单）→ 应用退出

### 视图菜单

- [ ] **SQL 执行历史…** → 打开全局历史 Dialog
- [ ] **设置…** → 打开设置 Dialog

### 帮助菜单

- [ ] **关于 Data Nexus** → 打开 About Dialog（非原生 MessageDialog）；平台信息来自 Go `GetPlatform()`

### 快捷键（v0.5 固定）

- [ ] `Cmd/Ctrl+N` → 新建连接
- [ ] `Cmd/Ctrl+O` → 打开 SQLite
- [ ] `Cmd/Ctrl+W` → 关闭当前连接
- [ ] `Cmd/Ctrl+,` → 打开设置

---

## Go 原生菜单

- [ ] macOS：菜单栏 **仅** App / Edit / Window（无 File / View / Help）
- [ ] macOS App 菜单 **Quit**（`Cmd+Q`）可用
- [ ] **无重复** 的 Open / Settings / About / Theme / Language 原生项

---

## 连接 Group、游离连接与 catalog.db

- [ ] 默认顶层 Group「我的连接」（**非**全列表根包裹层）
- [ ] **游离连接** 与 📁「我的连接」**第一层同级**（相同缩进）
- [ ] Group 名称 **原地重命名**：Enter 生效，无 Dialog
- [ ] Group 右键：**新建 / 重命名 / 删除**
- [ ] 拖 Group 改 **父层级**；拖连接进 Group 或根层游离
- [ ] 拖 Group / 连接改 **同级顺序**（`sort_order`）
- [ ] 连接右键：**打开连接 / 关闭连接 / 编辑连接 / 删除**（**无**单独重命名）
- [ ] **SQLite 已 open**：右键第五项 **附加数据库…** → `AttachDatabaseDialog`
- [ ] **SQLite 未 open**：「附加数据库…」disabled
- [ ] **PG/MySQL**：无「附加数据库…」菜单项
- [ ] **已 open** 时「编辑连接」disabled；关闭后在 Dialog 改 **显示名称** 与配置
- [ ] 删除 Group 且不勾选「删除连接」→ 连接变 **游离**（**不**进入父 Group 或「我的连接」）
- [ ] 删除 Group 且勾选「删除连接」→ 连接配置删除
- [ ] 新建连接未选 Group → **游离连接**
- [ ] 拖连接到根层 → **游离连接**（无右键「移为游离」菜单项）
- [ ] catalog.db 首次初始化 + 默认「我的连接」；**无** connections.json 迁移；无明文 password

## 连接列表与 schema 层级树

- [ ] 左侧 **无** 顶部新建/readOnly/wal/启动恢复控件
- [ ] **无** 手输 database/schema 切换框
- [ ] **无** v0.4 Attach 顶部按钮与 `window.prompt`

### 层级（按方言）

- [ ] **SQLite：** 连接 → `main` → 表；attach alias 为 L1
- [ ] **MySQL：** 连接 → 多个 database → 表/视图
- [ ] **PostgreSQL：** 连接 → database → schema → 表/视图

### 交互

- [ ] 仅 **已 open** 连接可展开 schema 子树
- [ ] 展开 database/schema **懒加载**
- [ ] 单击表名 → 右侧数据 Tab 正确

### SQLite Attach（v0.5 必做）

- [ ] 连接右键「附加数据库…」→ Attach Dialog → attach 成功出现在 L1
- [ ] 拖拽 `.db` 到已 open SQLite 连接 → Dialog 预填路径
- [ ] attach L1 右键「取消附加」→ Detach
- [ ] attach L1 右键「在 Finder 中显示」→ 定位源文件
- [ ] 窗口空白区拖拽 `.db` 仍为 **新建连接**（非 attach）

---

## NewConnectionDialog / ConnectionForm

- [ ] Dialog 浮动于主界面，宽约 720px
- [ ] 三方言 + 常规/安全/高级分区
- [ ] 方言扩展字段写入 catalog.db
- [ ] 编辑时 password 留空 → 不修改 Vault 密码

---

## SettingsDialog / SqlExecutionHistory / About

- [ ] 设置：启动恢复；主题/语言/日志
- [ ] SQL 执行历史跨连接；关于 Dialog i18n + 平台信息（Go GOOS/GOARCH）

---

## 回归（v0.4）

- [ ] SQLite / PG / MySQL 核心流程
- [ ] Vault、CSV 向导、拖拽打开窗口空白区
- [ ] **SavedQueries**（v0.3 已有，回归验证，非 v0.5 新功能）

---

## 完成

- [ ] 上述全部勾选
- [ ] tag **`v0.5.0`**
