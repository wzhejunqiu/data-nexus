# Data Nexus

一款轻量级 **桌面端 SQLite 管理工具**，支持同时打开多个数据库，浏览表结构、分页查看数据、执行 SQL。

当前版本：**v0.1.0**（MVP）。更新记录见 [CHANGELOG.md](CHANGELOG.md)。

## 下载

从 [GitHub Releases](https://github.com/wzhejunqiu/data-nexus/releases) 下载对应平台的安装包或压缩包：

| 平台 | 文件名 |
|------|--------|
| Windows (64-bit) | `data-nexus-v0.1.0-windows-amd64-installer.exe` |
| Windows (ARM64) | `data-nexus-v0.1.0-windows-arm64-installer.exe` |
| Linux (64-bit) | `data-nexus-v0.1.0-linux-amd64.zip` |
| Linux (ARM64) | `data-nexus-v0.1.0-linux-arm64.zip` |

> macOS 预构建包暂未发布（待配置代码签名与公证后开启）。macOS 用户可 [从源码构建](#参与开发)。

## 快速开始

1. 安装或解压下载的产物并启动：
   - **Windows**：运行 `.exe` 安装包，从开始菜单或桌面快捷方式启动
   - **Linux**：解压 `.zip`，赋予可执行权限后运行 `./data-nexus`
   - **macOS**：从源码构建后打开 `data-nexus.app`（预构建包暂未发布）
2. 通过 **File → Open** 选择 `.db` 文件，或将 `.db` 文件拖入窗口。
3. 在左侧连接树中选择数据库，浏览表结构或数据，或在 SQL 编辑器中执行查询。

也可在启动时直接指定数据库：

```bash
# macOS
./data-nexus.app/Contents/MacOS/data-nexus --db /path/to/app.db

# Windows / Linux
./data-nexus --db /path/to/app.db
```

## 功能

- **多连接**：像 Navicat 一样同时管理多个 SQLite 文件
- **表浏览**：查看列定义、索引，分页浏览数据，支持列排序
- **SQL 编辑器**：语法高亮，`Cmd/Ctrl + Enter` 执行；写操作需二次确认
- **只读模式**：连接时可开启，防止误改数据
- **查询历史**：当前会话内的 SQL 可快速复用
- **结果导出**：查询结果一键复制为 CSV
- **PRAGMA 快捷**：常用 SQLite 诊断语句一键插入
- **主题与语言**：深色 / 浅色 / 跟随系统；界面支持中文与英文

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

连接信息保存在 `~/.data-nexus/connections.json`。

## 已知限制（v0.1.0）

- 仅支持 **SQLite** 本地文件（PostgreSQL / MySQL 计划在后续版本）
- macOS 预构建包暂未发布（签名/公证就绪后通过 Release 提供）

## 参与开发

如需从源码构建或贡献代码，请参阅 [docs/implementation/README.md](docs/implementation/README.md)。

```bash
make dev      # 开发模式
make build    # 构建当前平台产物
make test     # 运行测试
```

**CI：** push/PR 到 `main` 时运行测试与 `linux/amd64` smoke build（见 [`.github/workflows/ci.yml`](.github/workflows/ci.yml)）。四个平台的 Release 产物（Linux ×2 + Windows ×2；macOS 待签名后开启）仅在推送 `v*` tag 时构建（见 [`.github/workflows/release.yml`](.github/workflows/release.yml)）。

设计与架构文档见 [docs/](docs/) 目录。

## 许可证

本项目采用 [MIT License](LICENSE) 开源。
