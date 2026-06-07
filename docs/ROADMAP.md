# Data Nexus — 实施路线图

> 版本: v0.3 · 最后更新: 2026-06-07

> **实施时请打开 [docs/implementation/](./implementation/README.md)**，各阶段有独立任务清单文件，本文件仅作概览与决策记录。

---

## 阶段概览

```
Phase 0 ──► Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4
 文档评审     Wails+Driver   前端骨架      功能集成      桌面端发布
           → phase-1        → phase-2     → phase-3     → phase-4
```

| 阶段 | 实施文档 | 预估 |
|------|----------|------|
| [Phase 0](./implementation/phase-0-review.md) | 文档评审 | 1-2 天 |
| [Phase 1](./implementation/phase-1-wails-backend.md) | Wails + Driver + Services | 3-5 天 |
| [Phase 2](./implementation/phase-2-frontend-shell.md) | 连接树 + Schema 子树 + 主区骨架 | 3-4 天 |
| [Phase 3](./implementation/phase-3-feature-integration.md) | MVP 功能闭环 | 4-6 天 |
| [Phase 4](./implementation/phase-4-release.md) | 测试 + v0.1.0 | 2-3 天 |

**MVP 合计:** 13-20 天（1 人全职）

---

## 后续版本

| 版本 | 实施文档 | 主题 |
|------|----------|------|
| v0.2 | [phase-v0.2-enhancements.md](./implementation/phase-v0.2-enhancements.md) | 体验增强 |
| v0.3 | [phase-v0.3-exploration.md](./implementation/phase-v0.3-exploration.md) | Filter→SQL、列画像 |
| v1.0 | [phase-v1.0-multi-connection.md](./implementation/phase-v1.0-multi-connection.md) | PostgreSQL / MySQL |
| v1.x | [phase-v1.x-multi-database.md](./implementation/phase-v1.x-multi-database.md) | Headless REST、ER 等 |

---

## 决策日志

| 日期 | 决策 | 理由 |
|------|------|------|
| 2026-06-07 | MVP 多连接 Navicat 模式 | 启动即主界面；`map[id]*Session`；Schema/SQL 带 `connectionId` |
| 2026-06-07 | v1.0 改为远程库 | 多连接已在 MVP；v1.0 聚焦 PostgreSQL/MySQL |
| 2026-06-07 | SQLite 驱动选 modernc.org/sqlite | 纯 Go，跨平台编译无 CGO |
| 2026-06-07 | 前端选 React + shadcn/ui | 生态成熟，定制性强 |
| 2026-06-07 | 不做 MVP 持久化 | ~~已废止~~ → 见下行 |
| 2026-06-07 | MVP 连接信息持久化 | `connections.json` + 连接树；P1 启动恢复已打开连接 |
| 2026-06-07 | 默认绑定 127.0.0.1 | 本地工具，无需认证 |
| 2026-06-07 | 不做 AI SQL 助手 | 市场拥挤、与本地优先冲突（竞品调研） |
| 2026-06-07 | v0.3 聚焦 Filter→SQL + 列画像 | Datasette/DuckDB UI 验证的需求，SQLite 工具空白区 |
| 2026-06-07 | 只读模式纳入 MVP P0 | sqlite-web/Tabulita 标配，降低误操作风险 |
| 2026-06-07 | 采用 Wails v2 桌面端 | 原生文件对话框、单二进制、对标 TablePlus；弃 HTTP embed 方案 |
| 2026-06-07 | MVP 不做 headless REST | Bridge 足够；`internal/service` 保留便于 v1.x `--server` |
| 2026-06-07 | 日志框架用 zap | 默认 INFO；dev 控制台、release 文件；用户 YAML/CLI 可配置 |
| 2026-06-07 | MVP 含全部 P1 | 索引、主题、PRAGMA、CSV 复制等均纳入 v0.1.0 |
| 2026-06-07 | 需要 i18n | react-i18next，默认 zh-CN，MVP 含 en 语言包 |
| 2026-06-07 | 首发三平台 | v0.1.0 同时 build macOS / Windows / Linux |
| 2026-06-07 | v0.3 确认做 | Filter→SQL、列画像、Canned Queries 等探索型差异化 |

> 开放问题决策后追加到此表。
