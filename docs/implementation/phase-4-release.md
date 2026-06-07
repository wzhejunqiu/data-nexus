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
- [ ] Welcome 页路径输入连接
- [ ] 无效路径 → 错误提示
- [ ] 只读模式 → 写 SQL 被拒绝
- [ ] 断开 → 重连
- [ ] `--db /path/to.db` 启动自动连接

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

```bash
wails build                          # 当前平台
wails build -platform darwin/arm64   # 按需
wails build -platform windows/amd64
wails build -platform linux/amd64
```

- [ ] `build/bin/` 产物可双击运行
- [ ] macOS `.app` / Windows `.exe`  smoke test

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
