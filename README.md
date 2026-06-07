# Data Nexus

一款轻量级 **桌面端** 数据库管理应用（Wails + Go + React），MVP 阶段聚焦 SQLite，后续扩展多数据库连接支持。

## 文档索引

| 文档 | 说明 |
|------|------|
| [产品需求文档 (PRD)](docs/product/PRD.md) | 目标用户、功能范围、用户故事、验收标准 |
| [竞品调研与差异化](docs/product/COMPETITIVE_ANALYSIS.md) | 市场分析、功能借鉴、卖点定位 |
| [技术设计文档](docs/design/ARCHITECTURE.md) | 系统架构、技术选型、模块划分 |
| [API 设计](docs/design/API.md) | Wails Service 绑定契约与错误码 |
| [数据模型](docs/design/DATA_MODEL.md) | 连接配置、Schema 抽象、多数据库扩展策略 |
| [UI/UX 设计](docs/design/UI_UX.md) | 页面结构、交互流程、组件规范 |
| [连接交互（Navicat 模式）](docs/design/CONNECTION_UX.md) | 启动即主界面、多连接并存 |
| [实施路线图](docs/ROADMAP.md) | 分阶段交付计划与里程碑 |
| [分阶段实施指南](docs/implementation/README.md) | **实施时直接查阅** — 各 Phase 详细任务清单 |

## 技术栈

- **桌面壳**: [Wails v2](https://wails.io/)（系统 WebView，单二进制）
- **后端**: Go 1.22+ · 日志 [zap](https://github.com/uber-go/zap)（默认 INFO，dev 控制台 / release 文件）
- **前端**: TypeScript + React + Vite + **i18next**（默认 zh-CN，含 en）
- **通信**: Wails in-memory Bridge（Go Service ↔ TypeScript 绑定）
- **MVP 数据库**: SQLite（`database/sql` + `modernc.org/sqlite`）

## 运行方式（实施后）

```bash
wails dev          # 开发
wails build        # 构建桌面应用 → build/bin/
```

## 状态

当前阶段：**文档设计** — 待评审通过后进入实施。
