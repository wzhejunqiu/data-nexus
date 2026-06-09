# v0.5.1 — 快捷键与 macOS 设置入口

> **预估:** 2–4 天 · **前置:** [v0.5](./phase-v0.5.md) · **下一版本:** [v0.5.2](./phase-v0.5.2.md)

v0.5 完成 MenuBar 与设置 Dialog 后，本版本补齐 **macOS 标准 `⌘,` 打开设置**，并在设置中提供 **可配置的快捷键**（「打开设置」除外）。

---

## 1. 版本目标

| 主题 | 说明 |
|------|------|
| macOS `⌘,` | 系统/应用内均可靠打开 `SettingsDialog`（与 MenuBar → 视图 → 设置一致） |
| 快捷键设置页 | 设置 Dialog 新增「快捷键」区块，可改绑所有**可配置**快捷键 |
| 固定快捷键 | **`⌘,` / `Ctrl+,` 打开设置** — 不可改绑，避免无法再次进入设置 |

**范围：** 仅 **Wails 桌面**；`--server` / `--api` 不适用（v0.7 浏览器模式另议）。

发布：**tag `v0.5.1`**

---

## 2. 打开设置（固定）

| 平台 | 快捷键 | 行为 |
|------|--------|------|
| macOS | `⌘,` | 打开设置 Dialog |
| Windows / Linux | `Ctrl+,` | 同上 |

**实现要点：**

- [ ] 前端 `AppShell` 全局监听 `Cmd/Ctrl+,`（v0.5 已有则验收 macOS WebView 焦点下仍生效）
- [x] macOS：`Cmd+,` 已在 **系统菜单栏 → 视图 → 设置**（`app_menu_darwin.go` → `app:settings`）；v0.5.1 验收 WebView 焦点下仍生效
- [ ] macOS：可选将 Settings 移至 App 菜单（HIG 推荐位置）；原生菜单文案随 i18n 动态更新
- [ ] MenuBar「视图 → 设置」与上述快捷键 **同一入口**，显示固定文案 `⌘,` / `Ctrl+,`

---

## 3. 可配置快捷键清单

以下快捷键可在设置中修改；冲突检测、恢复默认需支持。

### 3.1 MenuBar / 全局

| ID | 默认（macOS） | 默认（Win/Linux） | 动作 |
|----|---------------|-------------------|------|
| `shortcut.newConnection` | `⌘N` | `Ctrl+N` | 新建连接 |
| `shortcut.openSQLite` | `⌘O` | `Ctrl+O` | 打开 SQLite |
| `shortcut.closeConnection` | `⌘W` | `Ctrl+W` | 关闭当前连接 |
| `shortcut.openSqlHistory` | — | — | SQL 执行历史（可选默认 `⌘⇧H`） |

**不可配置：**

| ID | 快捷键 | 原因 |
|----|--------|------|
| `shortcut.openSettings` | `⌘,` / `Ctrl+,` | 设置页内无法改「打开设置」本身 |

**说明：** `Quit`（`⌘Q`）保留 macOS App 菜单系统项，不纳入用户自定义表。

### 3.2 SQL 编辑器

| ID | 默认 | 动作 |
|----|------|------|
| `shortcut.runQuery` | `⌘↵` / `Ctrl+Enter` | 执行 SQL |
| `shortcut.formatSql` | `⇧⌥F` / `Shift+Alt+F` | 格式化 SQL |

Monaco 内命令需与全局 shortcut 注册同步（改绑后 `editor.addCommand` / 全局 handler 更新）。

---

## 4. 设置 UI

在 [SettingsDialog](../design/UI_UX.md#47-设置v02v05-入口变更) 增加 **快捷键** 分区：

```
┌─ 设置 ─────────────────────────────────────────────┐
│  … 常规 / 外观 / 日志 …                             │
│  快捷键                                            │
│    新建连接      [ ⌘ N        ] [恢复默认]         │
│    打开 SQLite   [ ⌘ O        ]                    │
│    关闭连接      [ ⌘ W        ]                    │
│    SQL 执行历史  [ —          ]                    │
│    执行 SQL      [ ⌘ ↵        ]                    │
│    格式化 SQL    [ ⇧ ⌥ F      ]                    │
│    打开设置      ⌘ ,  （不可修改）                  │
│  [取消]                              [保存]         │
└───────────────────────────────────────────────────┘
```

**交互：**

- [ ] 点击输入框 → 录制按键组合（忽略纯修饰键）
- [ ] 冲突：与另一已绑快捷键冲突时 inline 错误，禁止保存
- [ ] 「恢复默认」：单项或整页
- [ ] 保存后立即生效（无需重启）；MenuBar 菜单项旁显示当前快捷键

---

## 5. 持久化

- [ ] 扩展 `AppConfig`（`~/.data-nexus/config.yaml`）或 `shortcuts.json`（与 config 同目录）；推荐 **`config.yaml` 下 `shortcuts:`** 节，经 `ConfigService` 读写
- [ ] 缺省字段使用上表默认值
- [ ] Go / TS 类型对齐见 [DATA_MODEL.md §2.4](../design/DATA_MODEL.md#24-appconfig应用配置含日志)（实施时补充 `ShortcutBindings`）

**序列化格式（示例）：**

```yaml
shortcuts:
  newConnection: "Meta+N"      # macOS；Windows 存 "Control+N"
  openSQLite: "Meta+O"
  closeConnection: "Meta+W"
  runQuery: "Meta+Enter"
  formatSql: "Shift+Alt+KeyF"
```

---

## 6. 功能清单

- [ ] `ShortcutRegistry`（或 `useShortcuts`）：加载配置、注册/注销全局键盘事件
- [ ] `SettingsDialog` 快捷键编辑 UI + 校验
- [ ] `ConfigService` / `config.yaml` 扩展 + 默认值
- [ ] macOS `⌘,` 与 App 菜单 Settings 联调
- [ ] `AppMenuBar` / Menu 项显示动态快捷键文案
- [ ] `SqlEditor` 执行/格式化与 registry 同步
- [ ] i18n：`settings.shortcuts.*`
- [ ] 测试：录制、冲突、恢复默认、macOS `⌘,` 打开设置

---

## 7. 手动验收

- [ ] macOS：`⌘,` 打开设置（焦点在连接树 / SQL 编辑器 / MenuBar 均有效）
- [ ] 修改「新建连接」为 `⌘⇧N` 后 MenuBar 与快捷键均生效
- [ ] 「打开设置」行不可编辑，始终显示 `⌘,`
- [ ] 冲突组合无法保存
- [ ] 重启应用后快捷键仍保留

---

## 8. 非目标

- 自定义 Quit（`⌘Q`）
- 连接树鼠标手势快捷键
- Server / API 模式快捷键（v0.7）

---

## 9. 完成标准

- [ ] checklist 勾选 + 手动验收
- [ ] tag **`v0.5.1`**

---

## 10. 相关文档

| 文档 | 用途 |
|------|------|
| [UI_UX.md §4.7](../design/UI_UX.md#47-设置v02v05-入口变更) | 设置页布局 |
| [phase-v0.5.md](./phase-v0.5.md) | 上一版本 |
| [phase-v0.5.2.md](./phase-v0.5.2.md) | 下一版本（PG 多库浏览） |
| [phase-v0.6.md](./phase-v0.6.md) | 后续版本 |
