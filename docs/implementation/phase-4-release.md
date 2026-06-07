# Phase 4 — 测试 & MVP 发布（v0.1.0）

> **预估:** 2-3 天 · **前置:** [phase-3-feature-integration.md](./phase-3-feature-integration.md) · **下一阶段:** [phase-v0.2-enhancements.md](./phase-v0.2-enhancements.md)

---

## 目标

稳定可发布的 **v0.1.0** 桌面应用。

---

## 4.1 自动化测试

### Go

```bash
go test ./internal/... -cover
```

- [ ] `internal/driver/sqlite` — metadata、BrowseTable、标识符校验
- [ ] `internal/service` — ConnectionManager 生命周期
- [ ] `internal/wails` — 错误映射（可选 mock driver）
- [ ] 覆盖率 **driver + service > 70%**

### 前端

```bash
cd frontend && npm test
```

- [ ] `ConnectionForm` — 渲染、只读勾选
- [ ] `mapWailsError` — 各错误码
- [ ] `DataGrid` — NULL/BLOB 渲染（可选）

---

## 4.2 手动测试清单

逐项勾选：

### 启动与连接

- [ ] 冷启动窗口可见 < 1s
- [ ] File → Open 选有效 `.db` → 连接成功
- [ ] 连接树「新建连接」+ 路径输入（高级）
- [ ] 无效路径 → 错误提示
- [ ] 只读模式 → 写 SQL 被拒绝
- [ ] 断开 → 重连
- [ ] 连接成功后出现在最近连接；重启应用仍在
- [ ] 最近连接一键打开，无需输入路径
- [ ] 删除最近连接条目
- [ ] P1：启动自动连接上次库（文件存在时）

### Schema & 数据

- [ ] 表/视图列表完整（无 sqlite_ 系统表）
- [ ] 表结构：列 + 索引（P1）
- [ ] 分页：翻页、pageSize 25/50/100/200
- [ ] 列排序 asc/desc
- [ ] 空表 Empty State
- [ ] 大表（10 万行）首页 < 500ms

### SQL

- [ ] SELECT 结果 + 耗时
- [ ] INSERT/UPDATE/DELETE + 确认框
- [ ] 语法错误提示
- [ ] 查询历史（P1）
- [ ] 结果 CSV 复制（P1）

### 桌面 & 日志

- [ ] 深色/浅色/跟随系统
- [ ] `wails build` 当前平台产物可运行
- [ ] release 构建日志写入文件（非控制台）
- [ ] `--log-level debug` 生效

---

## 4.3 发布准备

### 文档

- [ ] [README.md](../../README.md) — 安装、运行、构建、配置说明
- [ ] `config.yaml.example` 复制说明

### 构建

#### CI（main / PR）

[`.github/workflows/ci.yml`](../../.github/workflows/ci.yml) 在每次 push/PR 时运行：format、security、test，以及 **`smoke-build`（仅 `linux/amd64`）**。不在 main 上构建 macOS / Windows 产物。

#### Release（打 tag）

推送 `v*` tag 时，[`.github/workflows/release.yml`](../../.github/workflows/release.yml) 先复用 CI，再矩阵构建 **六个平台** 并上传 GitHub Release：

| 产物 | platform |
|------|----------|
| linux-amd64 / linux-arm64 | `linux/amd64` · `linux/arm64` |
| macos-amd64 / macos-arm64 | `darwin/amd64` · `darwin/arm64` |
| windows-amd64 / windows-arm64 | `windows/amd64` · `windows/arm64` |

本地交叉编译（可选，发布前自测）：

```bash
wails build                          # 当前平台（开发机）
wails build -platform darwin/arm64   # macOS Apple Silicon
wails build -platform darwin/amd64   # macOS Intel（按需）
wails build -platform windows/amd64
wails build -platform linux/amd64
```

- [ ] 打 tag 后 Release workflow 六平台构建全部成功
- [ ] macOS `.app` smoke test（本机或 Release 产物）
- [ ] Windows `.exe` smoke test（本机或 Release 产物）
- [ ] Linux binary smoke test（CI smoke-build 或 Release 产物）
- [ ] 各平台 i18n：zh-CN / en 切换正常

### Git

```bash
git tag -a v0.1.0 -m "MVP: SQLite desktop manager"
```

- [ ] tag `v0.1.0` 打在 release commit
- [ ] CHANGELOG 或 Release Notes（可选）

---

## 完成标准

- [ ] 手动测试清单 100% 通过
- [ ] Go 测试覆盖率达标
- [ ] v0.1.0 tag 已打
- [ ] 进入 [phase-v0.2-enhancements.md](./phase-v0.2-enhancements.md)（按需）
