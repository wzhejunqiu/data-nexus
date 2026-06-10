# Data Nexus

一款轻量级 **桌面端数据库管理工具**，支持 **SQLite、PostgreSQL、MySQL**，可同时打开多个连接，浏览表结构、分页查看数据、执行 SQL。

当前版本：**v0.5.0**。更新记录见 [CHANGELOG.md](CHANGELOG.md)。

## 下载

从 [GitHub Releases](https://github.com/wzhejunqiu/data-nexus/releases) 下载对应平台的安装包或压缩包：

| 平台 | 文件名 |
|------|--------|
| Windows (64-bit) | `data-nexus-v0.5.0-windows-amd64-installer.exe` |
| Windows (ARM64) | `data-nexus-v0.5.0-windows-arm64-installer.exe` |
| Linux (64-bit) | `data-nexus-v0.5.0-linux-amd64.zip` |
| Linux (ARM64) | `data-nexus-v0.5.0-linux-arm64.zip` |

> macOS 预构建包暂未发布（待配置代码签名与公证后开启）。macOS 用户可 [从源码构建](#参与开发)。

## 快速开始

1. 安装或解压下载的产物并启动：
   - **Windows**：运行 `.exe` 安装包，从开始菜单或桌面快捷方式启动
   - **Linux**：解压 `.zip`，赋予可执行权限后运行 `./data-nexus`
   - **macOS**：从源码构建后打开 `data-nexus.app`（预构建包暂未发布）
2. 通过菜单 **文件 → 打开 SQLite 文件…**（`Ctrl/Cmd+O`）选择 `.db`，或将 `.db` 拖入窗口空白区；拖到分组上可入组，拖到已打开的 SQLite 连接可附加库。
3. 在左侧连接树中展开连接（连接 → 数据库 → [PostgreSQL: schema] → 表），选择表浏览数据，或在 SQL 编辑器中执行查询。

也可在启动时直接指定数据库：

```bash
# macOS
./data-nexus.app/Contents/MacOS/data-nexus --db /path/to/app.db

# Windows / Linux
./data-nexus --db /path/to/app.db
```

## 功能

### 连接与导航（v0.5）

- **应用菜单栏**：Win/Linux 顶栏 **文件 / 视图 / 帮助**；macOS 使用系统菜单栏（HIG），窗口内仅显示「已打开 N 个连接」
- **多连接**：同时管理多个 **SQLite / PostgreSQL / MySQL** 连接；顶栏显示已打开连接数
- **连接分组**：可嵌套 Group（默认「我的连接」）与 **游离连接** 同级展示；支持拖拽改父级与同级排序、inline 重命名（F2/Enter）
- **层级树**：连接展开后按方言懒加载 **database → [PG: schema] → 表/视图**（MySQL 多库、PG 多 schema 同一连接内浏览）
- **新建 / 编辑连接**：浮动 Dialog，方言侧栏 + 常规/安全/高级分区；MySQL charset/collation/存储引擎、PG 全量 sslMode 与客户端编码
- **远程连接**：host/port/database/user/password、SSL/TLS、**测试连接**；密码存 Keychain 或本地 Vault（见 [docs/design/SECRETS.md](docs/design/SECRETS.md)）
- **MySQL**：推荐 8.0+；5.7 best-effort 兼容。MySQL 8 默认 `caching_sha2_password` 时建议开启 TLS，或改用 `mysql_native_password`。启用 TLS 时**默认校验服务端证书**；自签名/内网 CA 可在连接表单勾选「不校验服务端证书」。PostgreSQL 使用 `sslMode`（6 档 libpq 模式）
- **SQLite 专属**：ATTACH / DETACH 附加库；右键或拖拽 `.db` 到已打开连接附加；L1 节点可取消附加、在 Finder/文件管理器中显示源文件
- **全局 SQL 执行历史**：**视图 → SQL 执行历史…** 跨连接查看记录（含影响行），点击可跳转对应连接的 SQL 编辑器
- **设置与关于**：**视图 → 设置…**（`Ctrl/Cmd+,`）含主题、语言、启动时恢复已打开连接；**帮助 → 关于 Data Nexus** 显示版本与平台信息

### 数据与 SQL

- **表浏览**：查看列定义、索引，分页浏览数据，支持列排序
- **行内编辑**：多单元格编辑、待提交栏、批量提交（事务回滚）
- **CSV 导入 / 导出**：数据 Tab 与 SQL 结果均支持，可配置分隔符、编码等格式
- **列数据画像**：Schema Tab 展示 distinct、NULL 占比、min/max、top-N 等统计（大表采样）
- **可视化筛选**：Filter Builder 构建条件，可「在 SQL 编辑器中编辑」；低基数列 Facet 分面
- **FTS 搜索**：自动检测 FTS 虚拟表，数据 Tab 提供全文搜索
- **SQL 编辑器**：语法高亮、表名/列名自动补全、格式化、EXPLAIN 计划；`Cmd/Ctrl + Enter` 执行
- **命名查询**：侧边栏 Saved Queries，持久化至 `~/.data-nexus/queries.json`
- **查询历史**：按连接下拉快捷入口 + 全局历史 Dialog；执行记录持久化至 `~/.data-nexus/sql-global.db`
- **工作区状态**：当前表、筛选、排序、Tab 等自动持久化（localStorage）
- **写操作确认**：`INSERT`、`UPDATE`、`DELETE` 需二次确认
- **只读模式**：连接时可开启，防止误改数据
- **结果导出**：查询结果一键复制为 CSV
- **PRAGMA 快捷**：常用 SQLite 诊断语句一键插入

### 快捷键（v0.5 固定）

| 快捷键 | 动作 |
|--------|------|
| `Cmd/Ctrl+N` | 新建连接 |
| `Cmd/Ctrl+O` | 打开 SQLite 文件 |
| `Cmd/Ctrl+W` | 关闭当前连接 |
| `Cmd/Ctrl+,` | 打开设置 |
| `Cmd+Q`（macOS） | 退出 |

## 系统要求

- **macOS** 11+（Apple Silicon 或 Intel）
- **Windows** 10+（64-bit 或 ARM64）
- **Linux** 带 GTK 3 与 WebKitGTK 4.1 的桌面环境（Ubuntu 22.04+、Fedora 等常见发行版）

## 配置

用户配置文件位于 `~/.data-nexus/config.yaml`（首次运行后按需创建）。示例：

```yaml
log:
  level: info
  output: auto
```

完整选项见 [internal/config/config.yaml.example](internal/config/config.yaml.example)。

### 命令行参数

| 参数 | 说明 |
|------|------|
| `--db <path>` | 启动时打开指定 SQLite 文件 |
| `--log-level` | 日志级别：`debug` / `info` / `warn` / `error` |
| `--log-output` | 输出目标：`auto` / `console` / `file` / `both` |
| `--log-file` | 自定义日志文件路径 |

Release 版本默认将日志写入文件（`log.output: auto`）：

- **macOS**：`~/Library/Logs/data-nexus/data-nexus.log`
- **Windows**：`%LOCALAPPDATA%\data-nexus\logs\data-nexus.log`
- **Linux**：`~/.local/share/data-nexus/logs/data-nexus.log`

连接、分组与启动恢复状态保存在 `~/.data-nexus/catalog.db`；命名查询保存在 `~/.data-nexus/queries.json`。

> **v0.5 起** 使用 `catalog.db` 替代旧版 `connections.json`。**不会自动导入**旧版 json 中的连接记录。

## 已知限制

- macOS 预构建包暂未发布（签名/公证就绪后通过 Release 提供）
- v0.5 **不迁移**旧版 `connections.json` 连接数据；升级后需重新添加连接或手动导入
- 快捷键暂不可配置（计划见 [v0.5.1](docs/implementation/phase-v0.5.1.md)）

### Driver 集成测试（可选）

```bash
make test-integration        # MySQL + PostgreSQL（进程内 DB，无需 Docker）
make test-mysql-integration  # 仅 MySQL
make test-postgres-integration  # 仅 PostgreSQL
```

## 参与开发

如需从源码构建或贡献代码，请参阅 [docs/implementation/README.md](docs/implementation/README.md)。v0.5 实施细节见 [docs/implementation/v0.5/](docs/implementation/v0.5/)。

```bash
make dev      # 开发模式
make build    # 构建当前平台产物
make test     # 运行测试
```

**CI：** push/PR 到 `main` 时运行测试与 `linux/amd64` smoke build（见 [`.github/workflows/ci.yml`](.github/workflows/ci.yml)）。四个平台的 Release 产物（Linux ×2 + Windows ×2；macOS 待签名后开启）通过 GitHub Actions **手动触发** Release workflow 构建（见 [`.github/workflows/release.yml`](.github/workflows/release.yml)）。

设计与架构文档见 [docs/](docs/) 目录。

## 许可证

本项目采用 [MIT License](LICENSE) 开源。
