# Phase 4 — 测试 & MVP 发布（v0.1.0）

> **预估：** 2-3 天 · **状态：** 进行中 · **前置：** [phase-3-feature-integration.md](./phase-3-feature-integration.md) · **下一阶段：** [phase-v0.2-enhancements.md](./phase-v0.2-enhancements.md)

---

## 目标

稳定可发布的 **v0.1.0** 桌面应用。

---

## 4.1 自动化测试

### Go

```bash
go test ./internal/... -cover
```

- [x] `internal/driver/sqlite` — metadata、BrowseTable、标识符校验
- [x] `internal/service` — ConnectionManager 生命周期
- [x] `internal/wails` — 错误映射（可选 mock driver）
- [x] 覆盖率 **driver + service > 70%**（driver 85.2%、service 89.3%）

### 前端

```bash
cd frontend && npm test
```

- [x] `ConnectionForm` — 渲染、只读勾选
- [x] `mapWailsError` — 各错误码
- [x] `DataGrid` — NULL/BLOB 渲染（可选）

---

## 4.2 手动测试清单

> 详细状态见 [phase-4-manual-checklist.md](./phase-4-manual-checklist.md)；本节保留完整清单供发布走查。

逐项勾选：

### 启动与连接

- [x] 冷启动窗口可见 < 1s
- [x] File → Open 选有效 `.db` → 连接成功
- [x] 连接树「新建连接」+ 路径输入（高级）
- [x] 无效路径 → 错误提示
- [x] 只读模式 → 写 SQL 被拒绝
- [x] 断开 → 重连
- [x] 连接成功后出现在最近连接；重启应用仍在
- [x] 最近连接一键打开，无需输入路径
- [x] 删除最近连接条目
- [x] P1：启动自动连接上次库（文件存在时）

### Schema & 数据

- [x] 表/视图列表完整（无 sqlite_ 系统表）
- [x] 表结构：列 + 索引（P1）
- [x] 分页：翻页、pageSize 25/50/100/200
- [x] 列排序 asc/desc
- [x] 空表 Empty State（`DataGrid` / `SchemaSubtree`）
- [x] 大表（10 万行）首页 < 500ms（`query_service_perf_test.go`）

### SQL

- [x] SELECT 结果 + 耗时
- [x] INSERT/UPDATE/DELETE + 确认框
- [x] 语法错误提示
- [x] 查询历史（P1）
- [x] 结果 CSV 复制（P1）

### 桌面 & 日志

- [x] 深色/浅色/跟随系统
- [x] `wails build` 当前平台产物可运行
- [x] release 构建日志写入文件（非控制台）
- [x] `--log-level debug` 生效（CLI + `config_test.go` / `logger_test.go`）

---

## 4.3 发布准备

### 文档

- [x] [README.md](../../README.md) — 安装、运行、构建、配置说明
- [x] `config.yaml.example` 复制说明

### 构建

#### CI（main / PR）

[`.github/workflows/ci.yml`](../../.github/workflows/ci.yml) 在每次 push/PR 时运行：format、security、test，以及 **`smoke-build`（仅 `linux/amd64`）**。不在 main 上构建 macOS / Windows 产物。

#### Release（手动触发）

在 GitHub **Actions → Release → Run workflow** 中填写版本号（如 `v0.1.0`），[`.github/workflows/release.yml`](../../.github/workflows/release.yml) 会先复用 CI，再矩阵构建 **四个平台** 并创建 GitHub Release（含 tag，无需本地打 tag）：

| 产物                                    | platform                          |
| --------------------------------------- | --------------------------------- |
| linux-amd64 / linux-arm64 zip           | `linux/amd64` · `linux/arm64`     |
| windows-amd64 / windows-arm64 installer | `windows/amd64` · `windows/arm64` |

> macOS 矩阵在 release.yml 中已注释，待代码签名/公证后恢复。

本地交叉编译（可选，发布前自测）：

```bash
wails build                          # 当前平台（开发机）
wails build -platform darwin/arm64   # macOS Apple Silicon
wails build -platform darwin/amd64   # macOS Intel（按需）
wails build -platform windows/amd64
wails build -platform linux/amd64
```

- [ ] 手动触发 Release workflow 后四平台构建全部成功
- [x] macOS 本机 `wails build` smoke test（Release 暂无 macOS 产物）
- [x] Windows `.exe` smoke test（本机或 Release 产物）
- [x] Linux binary smoke test（CI `smoke-build` linux/amd64）
- [x] 各平台 i18n：zh-CN / en 切换正常

### Git

Release workflow 会在发布时自动创建 `v*` tag（指向触发时所选分支的 HEAD）。本地打 tag 为可选：

```bash
# 可选：与 workflow 输入的版本号一致
git tag -a v0.1.0 -m "MVP: SQLite desktop manager"
git push origin v0.1.0
```

- [ ] Release workflow 已成功创建 `v0.1.0` tag 与 GitHub Release
- [x] CHANGELOG 或 Release Notes（可选）

---

## 完成标准

- [x] 手动测试清单 100% 通过
- [x] Go 测试覆盖率达标
- [ ] v0.1.0 Release 已发布（含 tag）
- [ ] 进入 [phase-v0.2-enhancements.md](./phase-v0.2-enhancements.md)（按需）
