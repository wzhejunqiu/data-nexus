# Data Nexus — 分阶段实施指南

> 实施时 **直接打开当前阶段对应文件**，按清单逐项勾选。路线图概览见 [ROADMAP.md](../ROADMAP.md)。

## 阶段索引

| 阶段 | 文件 | 目标 | 状态 | 预估 |
|------|------|------|------|------|
| Phase 0 | [phase-0-review.md](./phase-0-review.md) | 文档评审，确认后开工 | ✅ 已完成 | 1-2 天 |
| Phase 1 | [phase-1-wails-backend.md](./phase-1-wails-backend.md) | Wails 骨架 + Driver + Services | ✅ 已完成 | 3-5 天 |
| Phase 2 | [phase-2-frontend-shell.md](./phase-2-frontend-shell.md) | 前端骨架 + 连接 + Schema 侧边栏 | ✅ 已完成 | 3-4 天 |
| Phase 3 | [phase-3-feature-integration.md](./phase-3-feature-integration.md) | 表浏览 + SQL 编辑器 + 联调 | ✅ 已完成 | 4-6 天 |
| Phase 4 | [phase-4-release.md](./phase-4-release.md) | 测试 + 打包 + v0.1.0 发布 | ✅ 已完成 | 2-3 天 |

## MVP 后版本

| 版本 | 文件 | 主题 | 状态 |
|------|------|------|------|
| v0.2 | [phase-v0.2-enhancements.md](./phase-v0.2-enhancements.md) | 体验增强（CSV、批量编辑、自动补全等） | ✅ 已完成 |
| | [phase-v0.2-manual-checklist.md](./phase-v0.2-manual-checklist.md) | v0.2 手动验收清单 | |
| v0.3 | [phase-v0.3-exploration.md](./phase-v0.3-exploration.md) | 探索型差异化（列画像、Filter→SQL 等） | ✅ 已完成 |
| v0.4 | [phase-v0.4-multi-connection.md](./phase-v0.4-multi-connection.md) | PostgreSQL / MySQL（以 **v0.4.0** 发布） | ✅ 已验收（待 tag） |
| | [phase-v0.4-manual-checklist.md](./phase-v0.4-manual-checklist.md) | v0.4 手动验收清单 | ✅ |
| v0.5 | [phase-v0.5.md](./phase-v0.5.md) | 桌面 UI 重构（MenuBar、连接右键、连接表单、SQL 历史） | |
| | [v0.5/](./v0.5/README.md) | v0.5 实施方案 · 设计 · 验收 | |
| v0.5.1 | [phase-v0.5.1.md](./phase-v0.5.1.md) | macOS ⌘, 设置；快捷键可配置 | |
| v0.6 | [phase-v0.6.md](./phase-v0.6.md) | 纯 Headless REST（`--api`） | |
| v0.7 | [phase-v0.7.md](./phase-v0.7.md) | Server 模式（`--server` 浏览器 UI） | |
| v0.8 | [phase-v0.8.md](./phase-v0.8.md) | ER 图（v0.8.0）· 插件（v0.8.1）· 驱动深化（v0.8.2） | |

## 设计文档速查

| 文档 | 实施时查阅 |
|------|------------|
| [ARCHITECTURE.md](../design/ARCHITECTURE.md) | 目录结构、Wails、zap 日志 |
| [API.md](../design/API.md) | Service 方法签名与错误码 |
| [DATA_MODEL.md](../design/DATA_MODEL.md) | Go/TS 数据结构 |
| [SECRETS.md](../design/SECRETS.md) | 远程连接密码（Keychain / Vault） |
| [OS_KEYCHAIN.md](../design/OS_KEYCHAIN.md) | OS Keychain 路径详细设计 |
| [UI_UX.md](../design/UI_UX.md) | 组件与交互 |
| [PRD.md](../product/PRD.md) | 功能优先级与验收标准 |

## 推荐实施顺序

```
Phase 0 ✓ → Phase 1 ✓ → Phase 2 ✓ → Phase 3 ✓ → Phase 4 ✓ → tag v0.1.0
                                                          ↓
                                                v0.2 ✓ → v0.3 ✓ → v0.4 ✓ → tag v0.4.0 → v0.5 → v0.5.1 → v0.6 → v0.7 → v0.8
```
