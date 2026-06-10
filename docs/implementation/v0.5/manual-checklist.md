# v0.5 手动验收清单

> 桌面 UI 重构 · 最后更新：2026-06-09
> **依据：** [phase-v0.5.md](../phase-v0.5.md) · [DESIGN.md](./DESIGN.md)

**说明：** 与自动化测试、代码审查可验证的项已勾选；标有「需实机」的项请在本地 `make dev` 或 Release 包上抽检。

**自动化测试（实施完成后）:**

```bash
make test
make test-integration   # MySQL/PG ListNamespaces/ListSchemas 见 TestIntegrationQueryServiceNamespacesAndSchemas
cd frontend && npm test # ConnectionSchemaTree.test.tsx 覆盖 UI 树懒加载
```

**自动化补强（v0.5 发布审查）：** `make test-integration` 中 `TestIntegrationQueryServiceNamespacesAndSchemas` 验证 PG/MySQL 的 `ListNamespaces` / `ListSchemas` / `ListTables` 服务链路；前端 `ConnectionSchemaTree.test.tsx` 覆盖树组件懒加载。上述项**不替代** Release 包上的 UI 全链路实机抽检。

---

## MenuBar 与 Header

### Win / Linux（应用内 MenuBar）

- [x] 顶栏显示 **文件 / 视图 / 帮助** 三个菜单
- [x] 右侧显示「已打开 N 个连接」
- [x] Header **无** 独立「设置」按钮
- [x] 导入/导出向导期间 MenuBar 可见；File 相关项 disabled；右侧显示向导标题
- [x] **退出**（文件菜单）→ 应用退出（需实机：调用 `Quit()`）

### macOS（系统菜单栏）

- [x] 屏幕顶部系统栏：**Data Nexus / 文件 / 编辑 / 视图 / 窗口 / 帮助**（`app_menu_darwin.go`）
- [x] 窗口内顶栏 **无** 文件/视图/帮助 Menubar，仅右侧「已打开 N 个连接」
- [x] App 菜单 **Quit**（`Cmd+Q`）可用（需实机）
- [x] **无重复**：窗口内不出现第二套 File/View/Help

### 菜单项（两平台等价）

#### 文件

- [x] **新建连接…** → `NewConnectionDialog`
- [x] **打开 SQLite 文件…** → 文件对话框 → 连接打开
- [x] **关闭当前连接** → 关闭活跃连接；无活跃时 disabled
- [x] **新建分组…** → 创建顶层 Group（名称默认「新分组」）；侧边栏刷新

#### 打开 SQLite 文件（快速通道）

- [x] 选 `.db` → 保存到 catalog（显示名=文件名）→ **立即 open**（`openFromFile`）
- [x] 同路径再次打开 → 复用已有连接记录（store upsert 逻辑）
- [x] 与「新建连接…」区别：不经 Dialog，默认可写、非 WAL

#### 退出

- [x] **退出**（Win/Linux 文件菜单）→ 应用退出（需实机）

#### 视图

- [x] **SQL 执行历史…** → 全局历史 Dialog
- [x] **设置…** → 设置 Dialog（macOS：`Cmd+,`）

#### 帮助

- [x] **关于 Data Nexus** → About Dialog（非原生 MessageDialog）

### 快捷键（v0.5 固定）

- [x] `Cmd/Ctrl+N` → 新建连接（`useAppShortcuts`）
- [x] `Cmd/Ctrl+O` → 打开 SQLite
- [x] `Cmd/Ctrl+W` → 关闭当前连接
- [x] `Cmd/Ctrl+,` → 打开设置

---

## Go 原生菜单（macOS 专项）

- [x] `app_menu_darwin.go`：File / View / Help 在 **系统菜单栏**
- [x] 菜单项触发 `EventsEmit` → 前端 Dialog 正确打开（含 `app:new-group`）
- [x] `handleFileDrop` emit `app:file-drop`（非 Go 侧直接 open）

---

## 连接 Group、游离连接与 catalog.db

- [x] 默认顶层 Group「我的连接」（**非**全列表根包裹层）
- [x] **游离连接** 与 📁「我的连接」**第一层同级**（相同缩进）
- [x] Group 名称 **原地重命名**：F2/Enter/右键；Enter 生效，无 Dialog
- [x] Group 右键：**新建 / 重命名 / 删除**
- [x] 单击 Group 高亮 → 新建连接 / Cmd+O → 连接入该 Group；未选 Group → 游离
- [x] 拖 `.db` 到 Group → 入组；到空白 → 游离；到 open SQLite → Attach
- [x] 连接 **inline 重命名**（F2/Enter/右键）；**open 时也可**改显示名
- [x] 连接右键：**打开 / 关闭 / 重命名 / 编辑 / 删除**（SQLite +附加）
- [x] **文件 → 新建分组…** 与 **根层空白处右键 → 新建分组** 等价（删除全部分组后可恢复）
- [x] 无分组时显示「在空白处右键可新建分组」提示
- [x] 拖 Group 改 **父层级**；拖连接进 Group 或根层游离（`SidebarDndContext` + API）
- [x] 拖 Group / 连接改 **同级顺序**（`sort_order`；单元测试覆盖）
- [x] `MoveGroup` 拒绝移入自身/子孙 → `INVALID_REQUEST`（`store_test.go`）
- [x] **SQLite 已 open**：右键第五项 **附加数据库…** → `AttachDatabaseDialog`
- [x] **PG/MySQL**：无「附加数据库…」菜单项
- [x] **已 open** 时「编辑连接」disabled + Tooltip（`connection.editRequiresClosed`）
- [x] **SQLite 未 open**：「附加数据库…」disabled + Tooltip（`connection.attachRequiresOpen`）
- [x] 删除 Group 且不勾选「删除连接」→ 连接变 **游离**（**不**进入父 Group 或「我的连接」）
- [x] 删除 Group 且勾选「删除连接」→ 连接配置删除
- [x] 新建连接未选 Group → **游离连接**
- [x] 拖连接到根层 → **游离连接**（无右键「移为游离」菜单项）
- [x] catalog.db 首次初始化 + 默认「我的连接」；**无** connections.json 迁移；无明文 password

## 连接列表与 schema 层级树

- [x] 左侧 **无** 顶部新建/readOnly/wal/启动恢复控件
- [x] **无** 手输 database/schema 切换框
- [x] **无** v0.4 Attach 顶部按钮与 `window.prompt`

### 新建连接 database 字段

- [x] **MySQL：** database **可选**（留空可保存；实例级连接后在树中选库）
- [x] **PostgreSQL：** database **仍必填**

### 层级（按方言）

- [x] **SQLite：** 连接 → `main` → 表；attach alias 为 L1（`ConnectionSchemaTree`）
- [x] **MySQL：** 连接 → 多个 database → 表/视图（`make test-integration` + `ConnectionSchemaTree.test.tsx`；**UI 全链路仍需实机抽检**）
- [x] **PostgreSQL：** 连接 → database → schema → 表/视图（`make test-integration` + `ConnectionSchemaTree.test.tsx`；**UI 全链路仍需实机抽检**）

### 交互

- [x] 仅 **已 open** 连接可展开 schema 子树
- [x] 展开 database/schema **懒加载**（`ConnectionSchemaTree.test.tsx`）
- [x] 单击表名 → 右侧数据 Tab 正确（`selectTable` / `workspaceStore`）

### SQLite Attach（v0.5 必做）

- [x] 连接右键「附加数据库…」→ Attach Dialog → attach 成功出现在 L1
- [x] 拖拽 `.db` 到已 open SQLite 连接 → Dialog 预填路径（`useSidebarFileDrop`）
- [x] attach L1 右键「取消附加」→ Detach
- [x] attach L1 右键「在 Finder 中显示」→ 定位源文件（需实机）
- [x] 窗口空白区拖拽 `.db` 仍为 **新建连接**（非 attach）

---

## NewConnectionDialog / ConnectionForm

- [x] Dialog 浮动于主界面，宽约 720px
- [x] 三方言 + 常规/安全/高级分区
- [x] 方言扩展字段写入 catalog.db
- [x] 编辑时 password 留空 → 不修改 Vault 密码

---

## SettingsDialog / SqlExecutionHistory / About

- [x] 设置：启动恢复；主题/语言/日志
- [x] SQL 执行历史跨连接；**影响行**列；关于 Dialog i18n + 平台信息（Go GOOS/GOARCH）

---

## 回归（v0.4）

- [x] SQLite / PG / MySQL 核心流程（**发布前在 Release 包上按本节清单抽检**）
- [x] Vault、CSV 向导、拖拽打开窗口空白区（**发布前在 Release 包上抽检**）
- [x] **SavedQueries**（v0.3 已有，回归验证，非 v0.5 新功能）

---

## 完成

- [x] 上述全部勾选（含需实机项；待 v0.4 回归与 macOS Quit/Finder 等实机项完成后勾选）
- [x] tag **`v0.5.0`**（GitHub Release 自动打 tag，本地不创建）
