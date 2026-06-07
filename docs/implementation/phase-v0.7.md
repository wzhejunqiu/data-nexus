# v0.7 — Server 模式（`--server`）

> **预估:** 6–10 天 · **前置:** [v0.6](./phase-v0.6.md) · **下一版本:** [phase-v0.8.md](./phase-v0.8.md)

> **v0.5–v0.7 总览:** v0.5 桌面 UI → v0.6 `--api` → **v0.7 `--server` 浏览器 UI**

v0.7 在 v0.6 REST 之上挂载静态 React UI，并引入前端 **Transport 层**（Wails vs HTTP）。

ER 图、插件、驱动深化见 [phase-v0.8.md](./phase-v0.8.md)（v0.8.0 / v0.8.1 / v0.8.2）。

---

## 1. 版本目标

| 项 | 说明 |
|----|------|
| 启动 | `./data-nexus --server`（默认 `127.0.0.1:8080`） |
| 与 Wails | **互斥** |
| 浏览器 | `GET /` → `frontend/dist`；`/api/v1/*` 复用 v0.6 handler |
| 桌面 | 仍用 Wails + v0.5 MenuBar UI（默认启动方式不变） |

发布：**tag `v0.7.0`**

---

## 2. 架构

```
浏览器  →  GET /              →  embed frontend/dist（SPA fallback）
        →  /api/v1/*         →  v0.6 共用 REST（internal/server）
```

| 组件 | 工作 |
|------|------|
| `internal/server/static.go` | 仅 `--server` 注册 |
| `server.Run({UI: true})` | `--server` 分支 |
| `frontend/.../transport.ts` | `desktop` → wailsjs；`server` → fetch |

---

## 3. 功能清单

### 3.1 后端

- [ ] `main.go`：`--server` → `Run({UI: true})`；与 `--api` 互斥
- [ ] 静态资源 embed + SPA `index.html` fallback
- [ ] 构建流水线：`wails build` 或独立 `frontend build` 嵌入 server 二进制

### 3.2 前端 Transport

- [ ] `frontend/src/lib/api/transport.ts`
- [ ] 运行时检测 server 模式；各 `lib/api/*.ts` 经 Transport
- [ ] Server 能力降级（见下表）

| 能力 | Wails 桌面 | `--server` 浏览器 |
|------|------------|-------------------|
| 打开 SQLite | 文件对话框 | 路径输入 / 上传；或 `--db` |
| CSV 导出 | SaveFile | 下载响应 / 服务端路径 |
| Keychain | 可用 | 不可用；Vault 策略见 SECRETS |
| 事件 | `EventsOn` | refetch |

### 3.3 验收

- [ ] 浏览器：连接 → Schema → 数据 → SQL（与 v0.5 桌面核心路径一致）
- [ ] Transport 单测；server integration：`GET /` 返回 HTML

---

## 4. 非目标（v0.7）

- 重复实现 v0.6 REST（应复用 handler）
- ER 图、插件、驱动深化（[v0.8](./phase-v0.8.md)）

---

## 5. 完成标准

- [ ] `--server` + 浏览器完整 UI 可用
- [ ] REST 行为与 v0.6 / Wails 一致
- [ ] tag **`v0.7.0`**

---

## 6. 相关文档

| 文档 | 用途 |
|------|------|
| [UI_UX.md §1.2](../design/UI_UX.md) | Server 模式 UI |
| [API.md §10](../design/API.md#10-http-rest-映射v05) | REST + CLI |
| [ARCHITECTURE.md](../design/ARCHITECTURE.md) | 三运行时 |
| [phase-v0.6.md](./phase-v0.6.md) | 上一版本 |
| [phase-v0.8.md](./phase-v0.8.md) | 下一版本（平台能力） |
