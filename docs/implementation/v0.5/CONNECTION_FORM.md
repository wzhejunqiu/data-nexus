# v0.5 新建 / 编辑连接表单设计

> **呈现方式:** 浮动 Dialog，覆盖主界面；**不**开全屏子页面（与 CSV 导入/导出向导区分）  
> **范围:** `NewConnectionDialog` 全面 UI 重构 + `EditConnectionDialog` 共用表单组件 + 后端方言配置扩展

---

## 1. 设计目标

| 目标 | 说明 |
|------|------|
| 统一体验 | 新建与编辑共用同一套字段组件，减少重复与行为不一致 |
| 编辑前置条件 | **仅连接关闭时**可编辑配置；open 状态下后端拒绝 `UpdateConnection*` |
| 方言感知 | SQLite / PostgreSQL / MySQL 展示**各自相关**的配置项，隐藏无关字段 |
| 信息分层 | 常规 / 安全 / 高级 三区折叠，避免 v0.4 单列表单过长 |
| 保持 Modal | 宽 Dialog（约 720px），主工作台仍可见背景；Esc / 点击遮罩关闭 |

---

## 2. 布局线框

### 2.1 整体结构

```
┌─ 新建连接 ────────────────────────────────────────────────────────────────┐
│ ┌─────────────┐  ┌─────────────────────────────────────────────────────┐ │
│ │  SQLite  ●  │  │  显示名称                                            │ │
│ │  PostgreSQL │  │  [ local-app.db                              ]       │ │
│ │  MySQL      │  │                                                     │ │
│ │             │  │  ▼ 常规                                             │ │
│ │             │  │  ┌─────────────────────────────────────────────┐   │ │
│ │             │  │  │ （方言相关字段，见 §3）                        │   │ │
│ │             │  │  └─────────────────────────────────────────────┘   │ │
│ │             │  │  ▶ 安全与 TLS                                       │ │
│ │             │  │  ▶ 高级选项                                       │ │
│ └─────────────┘  └─────────────────────────────────────────────────────┘ │
│                                                                            │
│              [ 测试连接 ]                         [ 取消 ] [ 保存并打开 ] │
└────────────────────────────────────────────────────────────────────────────┘
```

- **左侧：** 方言选择（垂直 Radio / 列表项）；切换时保留已填写的通用字段（host/user 等）仅在同一方言 family 内不共享
- **右侧：** 可滚动表单区；默认展开「常规」，「安全」「高级」可折叠
- **底部：** 左「测试连接」（远程库）；右「取消」+ 主按钮「保存并打开」（新建）/「保存」（编辑）

### 2.2 尺寸与交互

| 属性 | 值 |
|------|-----|
| 宽度 | `max-w-[720px]`，最小 560px |
| 最大高度 | `max-h-[85vh]`，内容区 `overflow-y-auto` |
| 遮罩 | 半透明；向导模式时新建 Dialog 仍可用 |
| 校验 | 见 [§4 表单校验与必填标识](#4-表单校验与必填标识)；提交 / 测试连接前客户端校验；必填项 inline 错误 |
| 密码 | 新建必填；编辑留空表示不修改 |

### 2.3 编辑模式约束

| 规则 | 说明 |
|------|------|
| **仅关闭可编辑** | 连接 `status === 'open'` 时 **不得** 打开 `EditConnectionDialog`；右键菜单项 disabled |
| 显示名称 | Edit Dialog 保留「显示名称」字段；与 **inline 重命名**（F2/Enter/右键）**并存**；两条路径均 `RenameConnection`；Edit 仅 **closed** 时可改配置 |
| 后端校验 | `UpdateConnection*` / `UpdateConnectionPostgresSettings` / `UpdateConnectionMySQLSettings` / SQLite 路径更新：若连接在活跃 map 中 → `CONNECTION_OPEN` |
| 密码留空 | `UpdateConnection*` 中 password 为空或 omitted → **跳过密码更新**，保留 Vault 现有值（见 [BACKEND.md §6.2](./BACKEND.md#62-密码更新语义)） |
| 用户提示 | disabled 菜单 Tooltip：`connection.editRequiresClosed` →「请先关闭连接」 |

### 2.4 SQLite 快捷路径

MenuBar → **打开 SQLite 文件…**（`Cmd+O`）仍为快速通道：文件对话框 + 默认选项（读写、无 WAL）直接打开，**不**经过本 Dialog。

本 Dialog 用于需要配置只读/WAL 或先保存再打开的 SQLite，以及全部远程连接。

---

## 3. 方言字段矩阵

### 3.1 SQLite

| 分区 | 字段 | 类型 | 默认 | 必填 | 说明 |
|------|------|------|------|------|------|
| 常规 | 显示名称 | text | 文件名 | 否 | 空则取 basename；标签旁标注「可选」 |
| 常规 | 文件路径 | path + 浏览 | — | **是** | |
| 高级 | 只读模式 | checkbox | off | 否 | 同 v0.4 |
| 高级 | WAL 模式 | checkbox | off | 否 | 只读时 disabled |

**无** TLS / charset / 存储引擎（不适用）。

### 3.2 PostgreSQL

| 分区 | 字段 | 类型 | 默认 | 必填 | 说明 |
|------|------|------|------|------|------|
| 常规 | 显示名称 | text | — | **是** | trim 后非空 |
| 常规 | 主机 | text | `localhost` | 否 | 有默认值；仍校验 trim 非空（防用户清空） |
| 常规 | 端口 | number | `5432` | 否 | 1–65535 |
| 常规 | 数据库 | text | — | **是** | |
| 常规 | 用户 | text | — | **是** | |
| 常规 | 密码 | password | — | **新建** | 编辑留空 = 不修改 |
| 安全 | SSL 模式 | select | `disable` | 否 | 见 §3.5 |
| 安全 | 只读模式 | checkbox | off | 否 | |
| 高级 | Schema | text | `public` | 否 | 连接后默认 schema |
| 高级 | 客户端编码 | select | `UTF8` | 否 | 映射 `client_encoding` |

### 3.3 MySQL

| 分区 | 字段 | 类型 | 默认 | 必填 | 说明 |
|------|------|------|------|------|------|
| 常规 | 显示名称 | text | — | **是** | trim 后非空 |
| 常规 | 主机 | text | `localhost` | 否 | 有默认值；仍校验 trim 非空 |
| 常规 | 端口 | number | `3306` | 否 | 1–65535 |
| 常规 | 数据库 | text | — | **是** | |
| 常规 | 用户 | text | — | **是** | |
| 常规 | 密码 | password | — | **新建** | 编辑留空 = 不修改 |
| 安全 | 启用 TLS | checkbox | off | 否 | MySQL 8 `caching_sha2_password` 建议开启 |
| 安全 | 不校验服务端证书 | checkbox | off | 否 | 仅 TLS 开启时显示；自签名/内网 CA |
| 安全 | 只读模式 | checkbox | off | 否 | |
| 高级 | 字符集 (charset) | select | `utf8mb4` | 否 | DSN `charset` |
| 高级 | 排序规则 (collation) | select | `utf8mb4_unicode_ci` | 否 | DSN `collation`；选项随 charset 过滤 |
| 高级 | 默认存储引擎 | select | `InnoDB` | 否 | 连接后 `SET SESSION default_storage_engine` |

### 3.4 字段可见性规则

```
切换方言 → 仅渲染对应 DriverConfig 字段
折叠分区 → 未展开时不渲染复杂控件（可选 lazy）
编辑模式 → SQLite 文件路径只读；远程 host 等可改
```

### 3.5 PostgreSQL SSL 模式

| 值 | 用户文案 (zh-CN) | 行为 |
|----|------------------|------|
| `disable` | 禁用 | 不加密 |
| `allow` | 允许 | 优先非 SSL，可升级 |
| `prefer` | 优先 | 优先 SSL，可降级 |
| `require` | 必需 | 加密，不校验证书 |
| `verify-ca` | 验证 CA | 校验 CA，不校验主机名 |
| `verify-full` | 完全验证 | 校验 CA + 主机名 |

v0.4 仅 expose `disable` / `require` / `verify-full`；v0.5 **补齐** libpq 全量模式。

### 3.6 MySQL charset / collation 预设

**Charset 选项（首版）:**

| charset | 说明 |
|---------|------|
| `utf8mb4` | 推荐，全 Unicode |
| `utf8` | 兼容旧库 |
| `latin1` | 西欧 |
| `gbk` | 中文 Windows 常用 |

**Collation（随 charset 联动，示例）:**

| charset | collation 选项 |
|---------|----------------|
| utf8mb4 | `utf8mb4_unicode_ci`, `utf8mb4_general_ci`, `utf8mb4_0900_ai_ci` |
| utf8 | `utf8_general_ci`, `utf8_unicode_ci` |
| latin1 | `latin1_swedish_ci` |
| gbk | `gbk_chinese_ci` |

**默认存储引擎:**

| 值 | 说明 |
|----|------|
| `InnoDB` | 默认，推荐 |
| `MyISAM` |  legacy |
| （空） | 不 SET，使用服务端默认 |

---

## 4. 表单校验与必填标识

v0.4 `RemoteConnectionForm` 仅用 `placeholder`、无客户端校验；v0.5 **统一**为带标签的字段 + 必填星号 + 提交前校验。

### 4.1 字段 UI 规范

每个 text / number / password / select / path 控件使用共享 **`FormField`** 包装（可放在 `ConnectionForm/FormField.tsx`）：

```
显示名称 *
[ local-app.db                              ]
请输入显示名称                    ← 仅校验失败时显示，text-red-500 text-xs
```

| 元素 | 规范 |
|------|------|
| **标签** | 字段上方 `<label>`，文案来自 i18n；**不再**仅依赖 `placeholder` |
| **必填星号** | 必填字段标签后加 `*`，`text-red-500`，`aria-hidden="true"` |
| **可选提示** | 非必填且易误解的字段（如 SQLite 显示名称）标签后加灰色 `(可选)` |
| **输入框** | 复用 `components/ui/Input`；错误态追加 `border-red-500 focus:ring-red-500/40` |
| **无障碍** | 必填：`aria-required="true"`；错误：`aria-invalid="true"` + `aria-describedby` 指向错误文案 `id` |
| **Checkbox / Select** | 同行 label；checkbox 无星号（均有明确默认） |

**编辑模式差异：**

| 字段 | UI |
|------|-----|
| 密码 | 标签无星号；placeholder / hint：`connectionForm.passwordKeep` →「留空表示不修改」 |
| SQLite 文件路径 | 只读展示，不参与校验 |

### 4.2 校验时机

| 时机 | 行为 |
|------|------|
| **保存 / 保存并打开** | 全量校验；有错误则 **阻止提交**，不调用 API |
| **测试连接**（远程） | 与保存相同的远程必填规则（含新建密码）；SQLite 不适用 |
| **字段 blur** | 仅对该字段 re-validate（已 touch 过的字段） |
| **输入中** | 不实时报错；若该字段已有错误且用户修改了值，可清除该字段错误 |

校验失败时：

1. 展开含第一个错误字段的 Collapsible 分区（常规 / 安全 / 高级）
2. `scrollIntoView` 滚至首个错误字段
3. 可选：对首个错误 input `focus()`

### 4.3 校验规则

**文件:** `ConnectionForm/connectionFormValidation.ts`

```typescript
export type ConnectionFormErrors = Partial<
  Record<'displayName' | 'filePath' | 'host' | 'port' | 'database' | 'user' | 'password', string>
>

export function validateConnectionForm(
  state: ConnectionFormState,
  mode: 'create' | 'edit',
  dialect: DriverType,
): ConnectionFormErrors
```

| 字段 | SQLite | PostgreSQL / MySQL |
|------|--------|-------------------|
| `displayName` | 始终通过（空 → 后端 basename） | `trim()` 非空，否则 `required` |
| `filePath` | `trim()` 非空 | — |
| `host` | — | `trim()` 非空 |
| `port` | — | 整数 1–65535，否则 `invalidPort` |
| `database` | — | `trim()` 非空 |
| `user` | — | `trim()` 非空 |
| `password` | — | **create:** 非空；**edit:** 跳过（留空合法） |

有默认值的下拉（sslMode、charset、collation、clientEncoding、storageEngine）**不做**空值校验。

### 4.4 测试连接 / 后端错误反馈

| 来源 | UI |
|------|-----|
| 客户端校验失败 | 字段旁 inline 错误；**无** Toast |
| `CONNECTION_FAILED` + `reason: auth` | Toast + 高亮 `user` / `password` |
| `CONNECTION_FAILED` + `reason: network` | Toast + 高亮 `host` / `port` |
| `CONNECTION_FAILED` + `reason: database` | Toast + 高亮 `database` |
| 其他 API 错误 | Toast（`formatError`）；不强制映射字段 |
| Vault 锁定 | 现有 `onNeedVault` 流程，不变 |

### 4.5 与 §3 字段矩阵对应

```
必填列「是」     → 标签带 *
必填列「新建」   → create 带 *；edit 密码带「留空不修改」hint、无 *
必填列「否」     → 无 *；SQLite 显示名称带「可选」
host / port      → 无 *（有默认）；port 仍做范围校验
```

---

## 5. 组件架构

```
features/connection/
├── NewConnectionDialog.tsx          # 壳：open/onOpenChange，mode=create
├── EditConnectionDialog.tsx         # 壳：mode=edit，复用 ConnectionForm
├── ConnectionForm/
│   ├── ConnectionForm.tsx           # 左栏方言 + 右栏分区表单；持有 errors state
│   ├── FormField.tsx                # label + 必填 * + error + a11y
│   ├── connectionFormValidation.ts  # validateConnectionForm
│   ├── DialectSidebar.tsx           # SQLite / PG / MySQL 选择
│   ├── sections/
│   │   ├── GeneralSection.tsx       # 按 dialect 分发
│   │   ├── SecuritySection.tsx
│   │   └── AdvancedSection.tsx
│   ├── fields/
│   │   ├── SQLiteFields.tsx
│   │   ├── PostgresFields.tsx
│   │   └── MySQLFields.tsx
│   └── connectionFormDefaults.ts    # 各方言 default state
├── RemoteConnectionForm.tsx         # 废弃，逻辑迁入 ConnectionForm
```

**共享:**

- `ConnectionForm` 接受 `mode: 'create' | 'edit'`、`initialValues?`、`connectionId?`
- 底部按钮区：`onTest`、`onSubmit`、`submitLabel`；二者提交前均调用 `validateConnectionForm`
- Vault 解锁流程不变（`onNeedVault` callback）

---

## 6. 后端与数据模型

### 6.1 模型扩展

**Go** — `internal/model/connection.go`:

```go
type PostgresConfig struct {
    // ... existing ...
    ClientEncoding string `json:"clientEncoding,omitempty"` // default "UTF8"
}

type MySQLConfig struct {
    // ... existing ...
    Charset              string `json:"charset,omitempty"`              // default "utf8mb4"
    Collation            string `json:"collation,omitempty"`            // default "utf8mb4_unicode_ci"
    DefaultStorageEngine string `json:"defaultStorageEngine,omitempty"` // default "InnoDB"; empty = server default
}
```

**SettingsUpdate** 同步扩展 `PostgresSettingsUpdate` / `MySQLSettingsUpdate`。

**TypeScript** — `frontend/src/lib/types/index.ts` 对齐。

### 6.2 Driver 行为

| 方言 | 配置项 | Driver 实现 |
|------|--------|-------------|
| MySQL | charset, collation | `buildMySQLDSN` 使用配置值，缺省回退 utf8mb4 |
| MySQL | defaultStorageEngine | `Connect` 成功后 `SET SESSION default_storage_engine=?`（非空时） |
| PostgreSQL | clientEncoding | DSN query `client_encoding=UTF8` |
| PostgreSQL | sslMode 扩展 | 直接传 libpq `sslmode` |
| SQLite | readOnly, wal | 不变 |

### 6.3 持久化

新字段写入 `catalog.db` `connections.config_json`；**password 仍不进库**（SECRETS 不变）。

### 6.4 测试

| 层 | 用例 |
|----|------|
| `model` | 默认值 normalization |
| `mysql/dsn_test.go` | charset/collation 出现在 DSN |
| `mysql` integration | default_storage_engine SET 生效（可选） |
| `postgres` | client_encoding query param |
| `connection_store_test` | 更新/读取新字段 |
| 前端 | 方言切换、折叠区、charset→collation 联动、编辑密码留空、**必填星号**、**提交拦截**、**inline 错误** |
| `connectionFormValidation.test.ts` | 各方言 / create·edit 规则矩阵 |

---

## 7. i18n 键（新增）

| 前缀 | 示例 |
|------|------|
| `connectionForm.titleNew` | 新建连接 |
| `connectionForm.titleEdit` | 编辑连接 |
| `connectionForm.section.general` | 常规 |
| `connectionForm.section.security` | 安全与 TLS |
| `connectionForm.section.advanced` | 高级选项 |
| `connectionForm.charset` | 字符集 |
| `connectionForm.collation` | 排序规则 |
| `connectionForm.storageEngine` | 默认存储引擎 |
| `connectionForm.clientEncoding` | 客户端编码 |
| `connectionForm.sslMode.*` | 各 SSL 模式文案 |
| `connectionForm.testConnection` | 测试连接 |
| `connectionForm.saveAndOpen` | 保存并打开 |
| `connectionForm.optional` | （可选） |
| `connectionForm.passwordKeep` | 留空表示不修改 |
| `connectionForm.validation.required` | 请输入{{field}} |
| `connectionForm.validation.invalidPort` | 端口须为 1–65535 的整数 |
| `connectionForm.fields.displayName` | 显示名称 |
| `connectionForm.fields.filePath` | 文件路径 |
| `connectionForm.fields.host` | 主机 |
| `connectionForm.fields.port` | 端口 |
| `connectionForm.fields.database` | 数据库 |
| `connectionForm.fields.user` | 用户 |
| `connectionForm.fields.password` | 密码 |

---

## 8. 与 v0.4 差异

| v0.4 | v0.5 |
|------|------|
| 顶部 Tab 按钮切换方言 | 左侧方言侧栏 |
| 远程表单扁平堆叠 | 常规 / 安全 / 高级 折叠分区 |
| MySQL charset 写死 DSN | 用户可配 charset / collation / 引擎 |
| PG 仅 3 种 sslMode | 6 种 libpq sslMode |
| `RemoteConnectionForm` 独立 | 统一 `ConnectionForm` |
| 编辑 Dialog 单独实现 | 与新建共用字段组件 |
| 仅 placeholder、无客户端校验 | `FormField` + 必填 `*` + `validateConnectionForm` |

---

## 9. 非目标（v0.5）

- SSH 隧道、Socket 路径连接
- 客户端证书 / 自定义 CA 文件（留 [v0.8](../phase-v0.8.md) 驱动深化）
- 连接 URL 粘贴解析
- 最近使用连接列表

---

## 10. 验收要点

- [ ] Dialog 浮动于主界面，非全屏子页
- [ ] 三方言字段矩阵正确显示/隐藏
- [ ] **必填字段标签带红色 `*`；SQLite 显示名称为「可选」**
- [ ] **保存 / 测试连接：缺必填项时 inline 错误、不发起 API**
- [ ] **编辑模式密码留空可保存；新建远程连接密码必填**
- [ ] **校验失败时自动展开含错误字段的分区并滚动到位**
- [ ] MySQL：charset、collation、TLS、默认存储引擎可保存并生效
- [ ] PostgreSQL：sslMode 全量、clientEncoding 可保存
- [ ] SQLite：只读/WAL 在高级区
- [ ] 编辑连接复用同一表单 UI
- [ ] 测试连接 + Vault 解锁流程正常
