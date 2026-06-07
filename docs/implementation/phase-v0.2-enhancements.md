# v0.2 — 体验增强

> **预估:** 5-7 天 · **前置:** v0.1.0 发布 · **参考:** [COMPETITIVE_ANALYSIS §4.2](../product/COMPETITIVE_ANALYSIS.md)

---

## 目标

补齐竞品标配能力，提升日常开发效率：行内批量编辑、CSV 导入导出、SQL 编辑器增强、应用配置 UI。

---

## 设计文档速查

| 主题 | 文档 |
|------|------|
| Service 方法签名与错误码 | [API.md §6.2 / §11–§13](../design/API.md) |
| CSV / 批量编辑数据模型 | [DATA_MODEL.md §5](../design/DATA_MODEL.md#5-csv-与批量编辑模型) |
| 行内编辑 / 导入导出 UI | [UI_UX.md §4.5.1–§4.7](../design/UI_UX.md) |
| FileService / ExportService 架构 | [ARCHITECTURE.md §5.1 / §5.6](../design/ARCHITECTURE.md) |
| 手动验收清单 | [phase-v0.2-manual-checklist.md](./phase-v0.2-manual-checklist.md) |

---

## 功能清单

### 连接与文件

- [x] CLI / 拖拽打开 `.db` — positional args + OnFileDrop；安装器**不**注册文件关联，避免覆盖用户默认应用
- [x] 拖拽 `.db` 到窗口打开（Wails OnFileDrop）— **已在 MVP 实现**

### 数据编辑（批量）

- [x] Spreadsheet 行内编辑 — 双击单元格进入编辑态
- [x] **多单元格待提交编辑** — 前端维护 `pendingEdits` 集合，高亮已改单元格
- [x] **批量提交** — 确认后调用 `TableService.UpdateCellsBatch`（1–200 条变更，事务 all-or-nothing）
- [x] 编辑确认 Modal + 只读模式拦截（`READ_ONLY`）
- [x] 提交失败整批回滚，Toast 展示 SQL 错误

> 设计要点见 [API.md §6.2](../design/API.md#62-updatecellsbatch) · [UI_UX.md §4.5.1](../design/UI_UX.md#451-行内编辑与批量提交)

### 导入导出

- [x] 表数据导出 CSV — 当前页 / 全表，可配置 [CSVFormatOptions](../design/DATA_MODEL.md#51-csvformatoptions)
- [x] 查询结果导出 CSV — SQL 结果 Tab，复用同一格式选项
- [x] CSV 导入 — 新建表或写入已有表；模式 `append` / `update`
- [x] 导入预览 — 列映射、前 N 行预览

> 设计要点见 [API.md §11–§12](../design/API.md#11-exportservice) · [UI_UX.md §4.6–§4.7](../design/UI_UX.md#46-导出-csv)

### SQL 编辑器

- [x] Monaco 表名/列名自动补全 — 基于当前 Schema
- [x] SQL 格式化（beautify）
- [x] EXPLAIN 可视化 — 查询计划树

### 配置

- [x] 设置页 — 日志级别 / 输出路径
- [x] `ConfigService` — 读写 `~/.data-nexus/config.yaml`
- [x] 完整 `config.yaml` 管理（含日志轮转参数）

> 设计要点见 [API.md §13](../design/API.md#13-configservice) · [UI_UX.md §4.7](../design/UI_UX.md#47-设置页)

---

## 后端任务

### TableService

- [x] `UpdateCellsBatch` — 批量 UPDATE，单事务，1–200 条变更上限
  - 每条变更：`{ rowKey, column, oldValue, newValue }`
  - 行定位：主键列组合 或 SQLite `rowid`（无主键时）
  - 失败 → 整批回滚，返回 `SQL_ERROR` + 详情

### FileService（v0.2 新增）

- [x] `WriteTextFile` — 将 CSV 文本写入用户选定路径（配合 `DialogService.SaveFile`）
- [x] 路径校验：仅允许 `.csv/.tsv/.txt` 扩展名

### ExportService（v0.2 新增）

- [x] `ExportTableCSV` — 全表导出 → SaveFile + 写盘
- [x] 共用 `CSVFormatOptions`（分隔符、引号、表头、NULL 表示、编码）

### ImportService（v0.2 末）

- [x] `ParseCSVPreview` — 解析 CSV 头 + 前 N 行
- [x] `ImportCSV` — append / update，新建或已有表

### DialogService 扩展

- [x] `OpenCSVFile` — CSV 文件选择器（导入入口）
- [x] `SaveFile` — 导出路径选择

### ConfigService（v0.2 新增）

- [x] `GetConfig` / `UpdateConfig` / `GetConfigPath`
- [x] 日志级别保存后立即生效；输出路径 / 轮转参数需重启

---

## 前端任务

- [x] `DataGrid` 行内编辑模式 + `pendingEdits` 状态条
- [x] `BatchEditConfirmDialog` — 批量确认
- [x] `ExportWizardPage` + `CsvFormatFields`
- [x] `ImportWizard` — 选文件 → 预览 → 目标表/模式 → 执行
- [x] `SettingsDialog` — 绑定 `ConfigService`
- [x] Monaco completion provider（表名/列名）
- [x] Export 按钮（数据 Tab 工具栏 + SQL 结果工具栏）
- [x] `lib/api/` 封装：`table.updateCellsBatch`、`export.*`、`import.*`、`config.*`、`file.*`

---

## 完成标准

- [x] CSV 导出/导入基本流程可用（含格式选项与 append/update 模式）
- [x] 行内编辑 + 批量提交可用（含只读拦截与事务回滚）
- [x] Monaco 表名/列名自动补全可用
- [x] 设置页可修改日志配置并持久化
- [ ] tag `v0.2.0`

---

## 下一阶段

[phase-v0.3-exploration.md](./phase-v0.3-exploration.md)
