# Phase 1 — Wails 骨架 + Driver + Services

> **预估:** 3-5 天 · **前置:** [phase-0-review.md](./phase-0-review.md) · **下一阶段:** [phase-2-frontend-shell.md](./phase-2-frontend-shell.md)

---

## 目标

可独立测试的 Wails 桌面应用骨架；Driver 与 Service 层支持连接、Schema、查询。

## 里程碑

`wails dev` 下通过文件对话框连接 SQLite，`ListTables` 返回表列表（前端最小验证页即可）。

---

## 1.1 项目初始化

### 命令

```bash
cd /path/to/data-nexus
wails init -n data-nexus -t react-ts
# 若已有 docs 目录，在空 Go 模块上 init 或手动合并目录
```

### 任务清单

- [ ] 初始化 Wails React-TS 项目
- [ ] 调整目录为 [ARCHITECTURE.md §4](../design/ARCHITECTURE.md) 结构：
  - `internal/wails/` — 绑定层
  - `internal/service/` — 业务层
  - `internal/driver/sqlite/`
  - `internal/model/`
  - `internal/config/`
  - `internal/logger/`
- [ ] `wails.json`：窗口 1280×800，最小 960×600，标题 `Data Nexus`
- [ ] `build/appicon.png` 占位图标
- [ ] 应用菜单：**File → Open**、**File → Quit**
- [ ] Makefile 或 Taskfile：
  - `make dev` → `wails dev`
  - `make build` → `wails build`
  - `make test` → `go test ./...`
  - `make lint` → `golangci-lint run`

### 配置与日志

- [ ] `internal/config/config.go` — 加载 `~/.data-nexus/config.yaml`（不存在则用默认值）
- [ ] `internal/config/config.yaml.example` — 提交到仓库
- [ ] `internal/logger/logger.go` — zap 初始化（见 [ARCHITECTURE §3.5](../design/ARCHITECTURE.md)）
  - 默认级别 **INFO**
  - `auto`：dev → 控制台；release → 文件
  - lumberjack 轮转
- [ ] CLI flags（覆盖配置文件）：
  - `--log-level`
  - `--log-output`（auto|console|file|both）
  - `--log-file`
- [ ] `main.go` 启动时初始化 logger，注入各 Service

### 依赖

```
go.uber.org/zap
gopkg.in/natefinch/lumberjack.v2
modernc.org/sqlite
gopkg.in/yaml.v3          # 配置
github.com/google/uuid    # 或 ulid — connection ID
```

---

## 1.2 Driver 层

**参考:** [DATA_MODEL.md](../design/DATA_MODEL.md) · [ARCHITECTURE §5.5](../design/ARCHITECTURE.md)

### 文件

```
internal/driver/
├── driver.go          # Driver 接口 + Type 常量
└── sqlite/
    ├── sqlite.go      # Connect/Close/Ping/Query/Exec/BrowseTable
    └── metadata.go    # ListTables/GetTableSchema (PRAGMA)
```

### 任务清单

- [ ] 定义 `driver.Driver` 接口
- [ ] `sqlite` 实现 `modernc.org/sqlite`
- [ ] `Connect` / `Close` / `Ping`
- [ ] `ListTables` — 过滤 `sqlite_%` 系统表
- [ ] `GetTableSchema` — `PRAGMA table_info` + 索引
- [ ] `BrowseTable` — 分页 + 单列排序 + 标识符白名单
- [ ] `QueryRows` / `Exec` — 结果集上限 10000
- [ ] `SerializeCellValue` — BLOB → `{type,size}`
- [ ] 单元测试：临时 `.db` 文件 + 内存库

### 验收

```bash
go test ./internal/driver/... -v
```

---

## 1.3 业务层 + Wails Services

**参考:** [API.md](../design/API.md)

### 业务层

```
internal/service/
├── connection_manager.go
└── query_service.go     # 可选，或逻辑放 manager
```

- [ ] `ConnectionManager` — 单连接，Connect 前先 Disconnect
- [ ] `model.AppError` — 错误码与 [API.md §1.3](../design/API.md) 一致
- [ ] 集成测试（不启动 Wails UI）

### Wails 绑定层

```
internal/wails/
├── connection.go   # ConnectionService
├── schema.go       # SchemaService
├── table.go        # TableService
├── query.go        # QueryService
├── dialog.go       # DialogService
└── app.go          # AppService
```

| Service | 方法 | 说明 |
|---------|------|------|
| `ConnectionService` | `GetStatus`, `Connect`, `Disconnect` | [API §3](../design/API.md) |
| `SchemaService` | `ListTables`, `GetTableSchema` | [API §4](../design/API.md) |
| `TableService` | `BrowseRows` | [API §5](../design/API.md) |
| `QueryService` | `Execute` | [API §6](../design/API.md) |
| `DialogService` | `OpenDatabaseFile` | 过滤器 `*.db;*.sqlite;*.sqlite3` |
| `AppService` | `GetVersion` | version + platform + arch |

- [ ] 所有 Service 构造函数接收 `*zap.Logger`
- [ ] 绑定层记录：`method`、`duration_ms`、错误码
- [ ] `OpenDatabaseFile` 取消 → `DIALOG_CANCELLED`

---

## 1.4 Wails 绑定验证

### main.go

- [ ] 创建 `ConnectionManager` + 各 Service
- [ ] `wails.Run` 注册 `Bind: []interface{}{...}`
- [ ] `OnStartup` 保存 context（Dialog 需要）

### 前端最小验证页（临时，Phase 2 替换）

```
frontend/src/App.tsx  — 三个按钮：
  GetVersion / OpenDatabaseFile+Connect / ListTables
```

- [ ] `wails dev` 可启动
- [ ] `wailsjs/go/wails/*` 绑定自动生成
- [ ] 选 `.db` → Connect → ListTables 有数据

---

## 完成标准

- [ ] `go test ./internal/...` 通过
- [ ] `wails dev` 全流程：Open → Connect → ListTables
- [ ] 日志：dev 模式控制台可见 INFO
- [ ] 进入 [phase-2-frontend-shell.md](./phase-2-frontend-shell.md)
