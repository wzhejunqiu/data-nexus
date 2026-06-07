# v0.2 — 体验增强

> **预估:** 5-7 天 · **前置:** v0.1.0 发布 · **参考:** [COMPETITIVE_ANALYSIS §4.2](../product/COMPETITIVE_ANALYSIS.md)

---

## 目标

补齐竞品标配能力，提升日常开发效率。

---

## 功能清单

### 连接与文件

- [ ] 双击 `.db` 文件关联 — 安装器 / 启动参数
- [ ] 拖拽 `.db` 到窗口打开（Wails OnFileDrop）

### 数据编辑

- [ ] Spreadsheet 行内编辑 — 双击单元格 → `UPDATE`
- [ ] 编辑确认 + 只读模式拦截

### 导入导出

- [ ] 表数据导出 CSV — `DialogService.SaveFile`
- [ ] 查询结果导出 CSV
- [ ] CSV 导入表 — 新建表或 append（设计导入 UI）

### SQL 编辑器

- [ ] Monaco 表名/列名自动补全 — 基于当前 Schema
- [ ] SQL 格式化（beautify）
- [ ] EXPLAIN 可视化 — 查询计划树

### 配置

- [ ] 设置页或配置文件 UI — 日志级别/输出路径
- [ ] 完整 `~/.data-nexus/config.yaml` 管理

---

## 后端任务

- [ ] `TableService.UpdateCell` — 行内编辑（或通用 Exec 封装）
- [ ] `ExportService` — CSV 生成（可选独立 Service）
- [ ] `ImportService` — CSV → INSERT（v0.2 末）

---

## 前端任务

- [ ] Settings 面板
- [ ] DataGrid inline edit 模式
- [ ] Monaco completion provider
- [ ] Export 按钮（表 Tab + SQL 结果）

---

## 完成标准

- [ ] CSV 导出/导入基本流程可用
- [ ] 行内编辑 + 自动补全可用
- [ ] tag `v0.2.0`

---

## 下一阶段

[phase-v0.3-exploration.md](./phase-v0.3-exploration.md)
