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
make test         # go test ./...
make build        # wails build → build/bin/
```

CI：推送到 `main` 或 PR 时运行 [GitHub Actions](.github/workflows/ci.yml)（**格式检查**、**漏洞扫描**、Go/前端测试；`main` 上额外三平台构建）。推送 `v*` tag 时 [Release](.github/workflows/release.yml) 会构建并上传产物。

本地与 CI 对齐：

```bash
make fmt-check    # gofmt
make lint         # golangci-lint（需已安装）
make vuln-check   # govulncheck + npm audit
make check        # 上述 + 测试构建
```

## 配置

用户配置：`~/.data-nexus/config.yaml`（示例见 [internal/config/config.yaml.example](internal/config/config.yaml.example)）

```bash
./build/bin/data-nexus.app/Contents/MacOS/data-nexus --log-level debug
./build/bin/data-nexus.app/Contents/MacOS/data-nexus --db /path/to/app.db
```

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
