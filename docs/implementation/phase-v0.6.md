# v0.6 — 纯 Headless REST（`--api`）

> **预估:** 5–8 天 · **前置:** [v0.5.1](./phase-v0.5.1.md) · **下一版本:** [phase-v0.7.md](./phase-v0.7.md)

> **v0.5–v0.7 总览:** v0.5 桌面 UI → **v0.6 `--api`** → v0.7 `--server` 浏览器 UI

v0.6 交付 **无 UI、无 Wails** 的 HTTP JSON API，供 CI / 脚本 / 自动化集成。REST handler 由 v0.7 `--server` 复用。

设计见 [API.md §10](../design/API.md#10-http-rest-映射v05)、[ARCHITECTURE.md §2.3](../design/ARCHITECTURE.md#23-http-模式v05)。

---

## 1. 版本目标

| 项 | 说明 |
|----|------|
| 启动 | `./data-nexus --api`（默认 `127.0.0.1:8080`） |
| 与 Wails | **互斥**，不调用 `wails.Run()` |
| 路由 | **仅** `/api/v1/*`；`GET /` 不返回 Web UI |
| 业务层 | 与桌面共用 `internal/service` |

发布：**tag `v0.6.0`**

---

## 2. 架构

```
curl / 脚本  →  /api/v1/*  →  chi  →  internal/server  →  internal/service  →  driver
```

- [ ] `internal/server/`：`RegisterAPI`、`Run(Options{UI: false})`
- [ ] `main.go`：`--api`、`--listen`；与 `--server` 互斥（v0.7 再加 `--server`）
- [ ] 统一 JSON 错误体（对齐 Wails `AppError`）
- [ ] `GET /api/v1/health`、`GET /api/v1/version`
- [ ] 可选 Basic Auth（非 localhost 建议默认开启）

---

## 3. REST 映射

实现 [API.md §10](../design/API.md#10-http-rest-映射v05) 全部路由（至少覆盖核心路径）：

| 优先级 | 路由 | 用途 |
|--------|------|------|
| P0 | `GET /api/v1/connections` | 列表 |
| P0 | `POST .../open`、`POST .../close` | 会话 |
| P0 | `GET .../schema/tables` | Schema |
| P0 | `POST .../query` | SQL |
| P0 | `GET /api/v1/health` | 健康检查 |
| P1 | 连接 CRUD、BrowseRows、Export/Import、Config、SqlExecutions | 与桌面 parity |

---

## 4. 功能清单

- [ ] chi router + handler 分层（`internal/server/api/*.go` 或按 Service 分文件）
- [ ] 注入与 Wails 相同的 service 实例（`main.go` 重构：共享 bootstrap）
- [ ] 集成测试：list → open → tables → query 闭环
- [ ] 集成测试：错误码与 Wails 对照
- [ ] [API.md](../design/API.md) 补充 `curl` 示例
- [ ] 非 localhost `--listen` 时日志警告

**典型验收：**

```bash
./data-nexus --api --listen 127.0.0.1:8080

curl -s http://127.0.0.1:8080/api/v1/health
curl -s http://127.0.0.1:8080/api/v1/connections | jq
curl -s -X POST "http://127.0.0.1:8080/api/v1/connections/{id}/open"
curl -s -X POST "http://127.0.0.1:8080/api/v1/connections/{id}/query" \
  -H 'Content-Type: application/json' -d '{"sql":"SELECT 1"}'
```

---

## 5. 非目标（v0.6）

- 静态前端 / 浏览器 UI（[v0.7](./phase-v0.7.md)）
- 前端 Transport 层
- OpenAPI 文档站（P2）
- 驱动深化（[v0.8.2](./phase-v0.8.md#4-v082--驱动深化)）

---

## 6. 完成标准

- [ ] `--api` 核心 REST 与 Wails **同输入同输出**
- [ ] `make test` / server integration 通过
- [ ] tag **`v0.6.0`**

---

## 7. 相关文档

| 文档 | 用途 |
|------|------|
| [ARCHITECTURE.md §2.4–§2.5](../design/ARCHITECTURE.md) | 三运行时、server 包结构 |
| [API.md §10](../design/API.md#10-http-rest-映射v05) | REST 表与 CLI |
| [phase-v0.5.1.md](./phase-v0.5.1.md) | 上一版本 |
| [phase-v0.7.md](./phase-v0.7.md) | 下一版本 |
