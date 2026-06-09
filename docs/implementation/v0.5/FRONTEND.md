# v0.5 前端实施方案

> **设计:** [DESIGN.md](./DESIGN.md) · **依赖:** Phase A 完成后再做 SqlExecutionHistoryDialog

---

## 1. AppMenuBar

**文件:** `frontend/src/components/AppMenuBar.tsx`

### Props

```typescript
interface AppMenuBarProps {
  openCount: number
  wizardTitle?: string
  activeConnectionId: string | null
  // Dialog 控制
  onNewConnection: () => void
  onNewGroup: () => void
  onOpenSettings: () => void
  onOpenSqlHistory: () => void
  onOpenAbout: () => void
  disabled?: boolean  // 向导模式部分 disabled
}
```

### 实现要点

1. 基于 `@radix-ui/react-menubar` + `components/ui/Menubar.tsx`（**仅 Win/Linux 渲染**）
2. **macOS：** `useIsMacOS()`（`lib/platform.ts` + `appApi.getPlatform()`）隐藏 Menubar；菜单由 `app_menu_darwin.go` 系统栏提供
3. 「打开 SQLite」：`dialogApi.openDatabaseFile()` → `connectionApi.openFromFile(...)`（快速新建/复用连接并 open；macOS 亦可通过 `app:open-sqlite` 触发）
4. 「新建分组」：`connectionGroupApi.createGroup('', t('connectionGroup.newGroupName'))`；macOS 经 `app:new-group` → `AppShell.createRootGroup`
5. 「关闭当前连接」：`activeConnectionId` 为空时 disabled
6. 右侧：`wizardTitle ?? t('app.openCount', { count: openCount })`

### openCount 数据来源

- 由 **`AppShell`** 从 `connectionApi.list()` 结果派生：`connections.items.filter(c => c.status === 'open').length`
- **非**后端单独接口；open/close 成功后 invalidate `['connections']` → 计数自动更新
- v0.5 变更：从 `ConnectionTree` 底部计数 **迁移至** MenuBar 右侧（Phase D S2）

### Header 重构

`Header.tsx` 变为薄包装：

```tsx
export function Header(props: AppMenuBarProps) {
  const isMacOS = useIsMacOS()
  return (
    <header className={`flex items-center border-b border-border px-4 ${isMacOS ? 'h-9' : 'h-12'}`}>
      <AppMenuBar {...props} />
    </header>
  )
}
```

移除内嵌 `SettingsDialogContainer` — 改由 `AppShell` 统一管理。

---

## 2. ConnectionTree（Group + 连接）

> **Group:** [CONNECTION_GROUPS_AND_STORAGE.md](./CONNECTION_GROUPS_AND_STORAGE.md) · **Schema 树:** [CONNECTION_TREE.md](./CONNECTION_TREE.md)

```typescript
useQuery({ queryKey: ['connectionSidebarTree'], queryFn: groupApi.getSidebarTree })
```

- `ConnectionTree` 根级：按 **`rootItems`** 混排渲染（fallback：`groups` + `freeConnections`）
- **根层空白右键** 或 **文件 → 新建分组…** → `createGroup('', name)`；无分组时显示 `connectionGroup.noGroupsHint`
- `ConnectionGroupNode` — 顶层/嵌套 📁；**inline 重命名**（F2/Enter）；单击设置 `selectedGroupId`；`data-group-drop-target`；右键 **新建/重命名/删除**；**draggable + sortable**
- `ConnectionTreeItem` — 连接行；**inline 重命名**；右键 **打开/关闭/重命名/编辑/删除**（SQLite 第五项「附加数据库…」）；**`data-sqlite-drop-target`** 供拖放 attach
- `placeConnectionInGroup.ts` — 新建/Cmd+O/拖 `.db` 入组 helper
- `sidebarRenameHandlers.ts` — F2/Enter handler 注册
- `SidebarDndContext` — `@dnd-kit` + **`@dnd-kit/sortable`**（改父级 + **`ReorderSidebarRoot` / `ReorderGroupMembers`**）
- `useSidebarFileDrop` — 监听 `app:file-drop`（Group 入组 / 空白游离 / open SQLite attach）
- `sidebarOrder.ts` — `rootItems` / `memberItems` 排序辅助
- `DeleteGroupDialog` — 未勾选时文案：变为 **游离连接**
- 连接 open 后挂载 `ConnectionSchemaTree`

### 2.1 Schema 子树

**目录:** `frontend/src/features/connection/tree/`

- `NamespaceNode` — L1 database（SQLite: main/attach；MySQL/PG: 各库）
- `SchemaNode` — L2 仅 PostgreSQL
- 懒加载：`listNamespaces` → 展开 → `listSchemas`（PG）→ `listTables`
- `browseContext` 写入 `workspaceStore`

### 2.2 SQLite Attach UI

**目录:** `frontend/src/features/connection/AttachDatabaseDialog.tsx`

- **入口：** SQLite 连接 L0 ContextMenu「附加数据库…」；拖拽 `.db` 到已 open SQLite 连接（预填路径）
- **attach L1 右键：** Detach + `appApi.revealFileInExplorer(filePath)`
- **invalidate：** `['namespaces', id]`、`['tables', id]`、`['attached', id]`
- 完整规格：[CONNECTION_TREE.md §3.6](./CONNECTION_TREE.md#36-sqlite-attach--detach)

### 2.3 移除（已完成）

- **`SchemaSubtree.tsx`**、**`SchemaSubtree.test.tsx`** — 已删除
- `RemoteNamespaceSwitch` — 未保留
- v0.4 Attach 顶部按钮与 `window.prompt` — 迁至 `AttachDatabaseDialog`

### 2.4 Vault

右键「打开」流程不变。

---

## 3. Dialog 组件

> **新建/编辑连接完整设计:** [CONNECTION_FORM.md](./CONNECTION_FORM.md)

### 3.1 ConnectionForm（新建/编辑，重构）

**目录:** `frontend/src/features/connection/ConnectionForm/`

**壳组件:**

| 组件 | mode | 说明 |
|------|------|------|
| `NewConnectionDialog` | `create` | AppShell 控制 open 状态 |
| `EditConnectionDialog` | `edit` | 右键「编辑连接」；**仅** `status !== 'open'` 时打开 |

**ConnectionForm Props:**

```typescript
interface ConnectionFormProps {
  mode: 'create' | 'edit'
  initialDialect?: DriverType
  initialValues?: Partial<ConnectionFormState>  // edit 时从 item 填充
  connectionId?: string                        // edit 时
  onSuccess: (connectionId: string) => void
  onCancel: () => void
  onNeedVault: (mode: VaultDialogMode, retry: () => void) => void
}
```

**布局:**

- `DialectSidebar` — 垂直列表；create 可切换；edit 时 dialect 固定 disabled
- `FormField` — 标签 + 必填 `*` / 可选提示 + inline 错误（见 [CONNECTION_FORM §4](./CONNECTION_FORM.md#4-表单校验与必填标识)）
- `connectionFormValidation.ts` — 提交 / 测试前校验；create·edit、三方言规则矩阵
- `GeneralSection` / `SecuritySection` / `AdvancedSection` — Collapsible（Radix Collapsible 或 `<details>`）
- 底部：`TestConnectionButton`（远程）+ `Cancel` + `Submit`

**MySQL charset→collation 联动:**

```typescript
const COLLATIONS_BY_CHARSET: Record<string, string[]> = {
  utf8mb4: ['utf8mb4_unicode_ci', 'utf8mb4_general_ci', 'utf8mb4_0900_ai_ci'],
  utf8: ['utf8_general_ci', 'utf8_unicode_ci'],
  // ...
}
// charset 变更时，若当前 collation 不在列表中则重置为第一项
```

**SQLite create 提交:**

```typescript
connectionApi.openFromFile({ filePath, readOnly, wal: readOnly ? false : wal })
// 或 save-only 变体（若产品需要「仅保存不打开」— 当前主按钮仍为保存并打开）
```

**废弃:** `RemoteConnectionForm.tsx` — 逻辑全部迁入 `PostgresFields` / `MySQLFields`。

### 3.2 SettingsDialog / SettingsDialogContainer

**文件:** `frontend/src/features/settings/SettingsDialog.tsx`

- `SettingsDialog` — 展示壳（Dialog + `SettingsForm`）
- `SettingsDialogContainer` — 数据 wrapper（`useQuery` config）；**由 AppShell 挂载**（v0.5 从 `Header` 迁出）

从 `ConnectionTree.RestoreOnStartupToggle` 迁移逻辑至 `SettingsForm`「常规」分区：

```tsx
<section>
  <h3>{t('settings.general')}</h3>
  <label>
    <input type="checkbox" checked={restoreOnStartup} onChange={...} />
    {t('settings.restoreOnStartup')}
  </label>
</section>
```

- 加载：`connectionApi.getRestoreOpenOnStartup()`
- 保存：与现有 config 保存流程合并，或 toggle 时即时调用 `setRestoreOpenOnStartup`

### 3.3 SqlExecutionHistoryDialog

**文件:** `frontend/src/features/sql-history/SqlExecutionHistoryDialog.tsx`

**API 封装** — `frontend/src/lib/api/queryHistory.ts`:

```typescript
export async function listAllExecutions(limit = 50): Promise<SqlExecutionList> {
  const res = await ListAllSqlExecutions(limit)
  return { items: res.items.map(mapRecord) }
}
```

**组件结构:**

- `Dialog` + 表格（可用现有 `Table` 或简单 `<table>`）
- `useQuery(['sql-executions-all'], () => listAllExecutions(50))`
- 连接名：从 `useQuery(['connections'])` 建 `id → name` map

**点击行 handler:**

```typescript
async function handleRowClick(record: SqlExecutionRecord) {
  onOpenChange(false)
  const conn = connections.find(c => c.id === record.connectionId)
  if (conn?.status !== 'open') {
    await connectionApi.open(record.connectionId)
  }
  setActiveConnectionId(record.connectionId)
  setActiveTab('sql')
  // 通过 callback 或 event 填充 SQL 编辑器
  onFillSql?.(record.connectionId, record.sql)
}
```

**SQL 填充:** 可通过 `workspaceStore` 新增 `pendingSql: { connectionId, sql } | null`，`SqlEditor` mount 时消费并清空。

### 3.4 AboutDialog

**文件:** `frontend/src/components/AboutDialog.tsx`

```tsx
export function AboutDialog({ open, onOpenChange }: DialogProps) {
  const { data: version } = useQuery({ queryKey: ['version'], queryFn: appApi.getVersion })
  const { data: platform } = useQuery({ queryKey: ['platform'], queryFn: appApi.getPlatform })
  // 展示 app.title, version.version, platform (GOOS/GOARCH from Go)
}
```

---

## 4. AppShell 集成

**文件:** `frontend/src/app/AppShell.tsx`

### Dialog 状态

与 [DESIGN.md §6.1](./DESIGN.md#61-dialog-状态appshell) 一致：

```typescript
const [newConnectionOpen, setNewConnectionOpen] = useState(false)
const [settingsOpen, setSettingsOpen] = useState(false)
const [sqlHistoryOpen, setSqlHistoryOpen] = useState(false)
const [aboutOpen, setAboutOpen] = useState(false)
const [attachDatabaseOpen, setAttachDatabaseOpen] = useState<{
  connectionId: string
  prefilledPath?: string
} | null>(null)
```

- **右键「附加数据库…」：** `setAttachDatabaseOpen({ connectionId })`
- **拖拽 `.db` 到 SQLite 连接：** `setAttachDatabaseOpen({ connectionId, prefilledPath: path })`
- Attach 成功或取消 → `setAttachDatabaseOpen(null)`

### 快捷键（v0.5 固定）

```typescript
useEffect(() => {
  const onKeyDown = (e: KeyboardEvent) => {
    const mod = e.metaKey || e.ctrlKey
    if (!mod) return
    switch (e.key.toLowerCase()) {
      case 'n': e.preventDefault(); setNewConnectionOpen(true); break
      case 'o': e.preventDefault(); handleOpenSQLite(); break
      case 'w': e.preventDefault(); handleCloseConnection(); break
      case ',': e.preventDefault(); setSettingsOpen(true); break
    }
  }
  window.addEventListener('keydown', onKeyDown)
  return () => window.removeEventListener('keydown', onKeyDown)
}, [activeConnectionId, ...])
```

保留 macOS 原生菜单事件，含 `EventsOn('app:settings', ...)`、`EventsOn('app:new-group', () => createRootGroup())` 等。`createRootGroup` 调用 `connectionGroupApi.createGroup('', t('connectionGroup.newGroupName'))` 并 invalidate 侧边栏树。

### 渲染

```tsx
<Header
  openCount={openCount}
  wizardTitle={...}
  activeConnectionId={activeConnectionId}
  onNewConnection={() => setNewConnectionOpen(true)}
  onNewGroup={() => void createRootGroup()}
  onOpenSettings={() => setSettingsOpen(true)}
  onOpenSqlHistory={() => setSqlHistoryOpen(true)}
  onOpenAbout={() => setAboutOpen(true)}
  disabled={!!exportSession || !!importSession}
/>
<NewConnectionDialog open={newConnectionOpen} onOpenChange={setNewConnectionOpen} />
<SettingsDialogContainer open={settingsOpen} onOpenChange={setSettingsOpen} />
<SqlExecutionHistoryDialog open={sqlHistoryOpen} onOpenChange={setSqlHistoryOpen} ... />
<AboutDialog open={aboutOpen} onOpenChange={setAboutOpen} />
{attachDatabaseOpen && (
  <AttachDatabaseDialog
    connectionId={attachDatabaseOpen.connectionId}
    prefilledPath={attachDatabaseOpen.prefilledPath}
    onOpenChange={(open) => !open && setAttachDatabaseOpen(null)}
  />
)}
```

**设置 Dialog 组件关系：** `SettingsDialog` 为展示壳（`open` / `config` / `onOpenChange`）；`SettingsDialogContainer` 为 **AppShell 挂载的 wrapper**，内部 `useQuery(['config'])` 后传入 `SettingsDialog`。MenuBar / 快捷键只 toggle `settingsOpen`，不直接引用 `SettingsDialog`。

---

## 5. UI 基元

### Menubar.tsx

参考 shadcn/ui Menubar 组件，导出:

- `Menubar`, `MenubarMenu`, `MenubarTrigger`, `MenubarContent`, `MenubarItem`, `MenubarSeparator`, `MenubarShortcut`

### ContextMenu.tsx

参考 shadcn/ui Context Menu，导出:

- `ContextMenu`, `ContextMenuTrigger`, `ContextMenuContent`, `ContextMenuItem`, `ContextMenuSeparator`

样式与现有 `Dialog`、`Button` 一致（Tailwind + CSS variables）。

---

## 6. 依赖安装

```bash
cd frontend
npm install @radix-ui/react-menubar @radix-ui/react-context-menu
```

---

## 7. 测试要点

### ConnectionTree.test.tsx

| 用例 | 断言 |
|------|------|
| 无顶部新建按钮 | `queryByRole('button', { name: /新建/ })` 为 null |
| 双击未打开连接 | mock `open` 被调用 |
| 双击已打开连接 | `setActiveConnectionId` 被调用 |
| 右键菜单打开 | ContextMenu 渲染打开项 |

### NewConnectionDialog.test.tsx

- 移除 readOnly/wal props；在 Dialog 内勾选后 `openFromFile` 参数正确

### SettingsDialog.test.tsx（新增）

- 启动恢复 checkbox 加载与 toggle

### SqlExecutionHistoryDialog.test.tsx（新增）

- mock `ListAllSqlExecutions` 渲染行
- 点击行触发连接切换与 SQL 填充

### AppMenuBar.test.tsx（可选）

- 各菜单项 callback 触发

---

## 8. i18n 变更

**zh-CN.json / en.json** 新增键（示例）:

```json
{
  "menu": {
    "file": {
      "label": "文件",
      "newConnection": "新建连接…",
      "openSQLite": "打开 SQLite 文件…",
      "closeConnection": "关闭当前连接",
      "newGroup": "新建分组…",
      "quit": "退出"
    },
    "view": {
      "label": "视图",
      "sqlHistory": "SQL 执行历史…",
      "settings": "设置…"
    },
    "help": {
      "label": "帮助",
      "about": "关于 Data Nexus"
    }
  },
  "settings": {
    "restoreOnStartup": "启动时恢复已打开连接"
  },
  "sqlHistory": {
    "title": "SQL 执行历史",
    "empty": "暂无执行记录"
  },
  "about": {
    "title": "关于 Data Nexus",
    "version": "版本 {{version}}",
    "platform": "平台 {{platform}}"
  },
  "connection": {
    "emptyHint": "暂无连接。使用 文件 → 新建连接 开始。"
  }
}
```

可删除或保留但不再使用的 key: `connection.sqliteOptionsHint`、`connection.restoreOnStartup`（迁至 settings）。

---

## 9. v0.5.1 预留（勿提前实现）

- `ShortcutRegistry` / 可配置快捷键
- MenuBar 菜单项旁动态快捷键文案
- App 菜单 Settings 与 `⌘,` 全平台验收

v0.5 仅需固定快捷键 + `AppShell` 内 `keydown` listener，便于 v0.5.1 替换为 registry。
