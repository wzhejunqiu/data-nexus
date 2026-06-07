# Data Nexus

一款轻量级 **桌面端** 数据库管理应用（Wails + Go + React），MVP 聚焦 SQLite，支持 Navicat 式多连接并存。

## 功能（v0.1.0）

- 启动即主界面，连接树管理多个 SQLite
- 表结构 / 分页数据 / SQL 编辑器（Monaco）
- 连接持久化（`~/.data-nexus/connections.json`）
- 只读模式、查询历史、PRAGMA 快捷、CSV 复制
- 中英文 i18n、深色/浅色/跟随系统主题

## 环境要求

- Go 1.25+
- Node.js 20+
- [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation)

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

## 开发

```bash
make dev          # wails dev
make test         # go test -short ./...（CI 等价，跳过性能测试）
make test-cover   # 带覆盖率
make build        # wails build → build/bin/
```

发布前本地性能验收：

```bash
make test-perf    # 10 万行 Browse <500ms、1 万行 SELECT <200ms
make bench        # Go benchmark（可选 profiling）
make gen-test-db  # 含 data/large.db（10 万行，手动验 UI）
```

CI：推送到 `main` 或 PR 时运行 [GitHub Actions](.github/workflows/ci.yml)（**格式检查**、**漏洞扫描**、Go/前端测试；`main` 上额外三平台构建）。推送 `v*` tag 时 [Release](.github/workflows/release.yml) 会构建并上传产物。

本地与 CI 对齐：

```bash
make fmt-check    # gofmt + frontend prettier
make lint         # golangci-lint（需已安装）
make vuln-check   # govulncheck + npm audit
make check        # 上述 + 测试构建
make pre-push     # push 前检查：format + lint + vuln（与 CI 格式/安全 job 对齐）
make install-hooks # 安装 pre-push hook，失败时阻止 push
```

## 配置

用户配置：`~/.data-nexus/config.yaml`（示例见 [internal/config/config.yaml.example](internal/config/config.yaml.example)）

Release 构建默认日志路径（`log.output: auto`）：

- macOS: `~/Library/Logs/data-nexus/data-nexus.log`
- Windows: `%LOCALAPPDATA%\data-nexus\logs\data-nexus.log`
- Linux: `~/.local/share/data-nexus/logs/data-nexus.log`

```bash
./build/bin/data-nexus.app/Contents/MacOS/data-nexus --log-level debug
./build/bin/data-nexus.app/Contents/MacOS/data-nexus --db /path/to/app.db
```

## 发布

MVP 版本 **v0.1.0**。打 tag 前请完成 [手动测试清单](docs/implementation/phase-4-manual-checklist.md)：

```bash
git tag -a v0.1.0 -m "MVP: SQLite desktop manager"
git push origin v0.1.0
```

推送 `v*` tag 后 [Release workflow](.github/workflows/release.yml) 会构建并上传三平台产物。

## 文档

| 文档 | 说明 |
|------|------|
| [PRD](docs/product/PRD.md) | 产品需求 |
| [ARCHITECTURE](docs/design/ARCHITECTURE.md) | 技术架构 |
| [API](docs/design/API.md) | Wails Service 契约 |
| [CONNECTION_UX](docs/design/CONNECTION_UX.md) | 多连接交互 |
| [实施指南](docs/implementation/README.md) | 分阶段清单 |

## 状态

**v0.1.0 MVP** — Wails 桌面端可构建运行。
