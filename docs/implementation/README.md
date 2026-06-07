# Data Nexus — 分阶段实施指南

> 实施时 **直接打开当前阶段对应文件**，按清单逐项勾选。路线图概览见 [ROADMAP.md](../ROADMAP.md)。

## 阶段索引

| 阶段 | 文件 | 目标 | 状态 | 预估 |
|------|------|------|------|------|
| Phase 0 | [phase-0-review.md](./phase-0-review.md) | 文档评审，确认后开工 | ✅ 已完成 | 1-2 天 |
| Phase 1 | [phase-1-wails-backend.md](./phase-1-wails-backend.md) | Wails 骨架 + Driver + Services | ✅ 已完成 | 3-5 天 |
| Phase 2 | [phase-2-frontend-shell.md](./phase-2-frontend-shell.md) | 前端骨架 + 连接 + Schema 侧边栏 | ✅ 已完成 | 3-4 天 |
| Phase 3 | [phase-3-feature-integration.md](./phase-3-feature-integration.md) | 表浏览 + SQL 编辑器 + 联调 | ✅ 已完成 | 4-6 天 |
| Phase 4 | [phase-4-release.md](./phase-4-release.md) | 测试 + 打包 + v0.1.0 发布 | 🔄 进行中 | 2-3 天 |

## MVP 后版本

| 版本 | 文件 | 主题 |
|------|------|------|
| v0.2 | [phase-v0.2-enhancements.md](./phase-v0.2-enhancements.md) | 体验增强（CSV、行内编辑、自动补全等） |
| v0.3 | [phase-v0.3-exploration.md](./phase-v0.3-exploration.md) | 探索型差异化（列画像、Filter→SQL 等） |
| v1.0 | [phase-v1.0-multi-connection.md](./phase-v1.0-multi-connection.md) | PostgreSQL / MySQL |
| v1.x | [phase-v1.x-multi-database.md](./phase-v1.x-multi-database.md) | Headless REST、ER 等 |

## 设计文档速查

| 文档 | 实施时查阅 |
|------|------------|
| [ARCHITECTURE.md](../design/ARCHITECTURE.md) | 目录结构、Wails、zap 日志 |
| [API.md](../design/API.md) | Service 方法签名与错误码 |
| [DATA_MODEL.md](../design/DATA_MODEL.md) | Go/TS 数据结构 |
| [UI_UX.md](../design/UI_UX.md) | 组件与交互 |
| [PRD.md](../product/PRD.md) | 功能优先级与验收标准 |

## 推荐实施顺序

```
Phase 0 ✓ → Phase 1 ✓ → Phase 2 ✓ → Phase 3 ✓ → Phase 4（进行中）→ tag v0.1.0
                                                          ↓
                                                v0.2 → v0.3 → v1.0 → v1.x
```
