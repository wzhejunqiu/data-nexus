# v0.8 — 平台能力（ER 图 · 插件 · 驱动深化）

> **预估:** 按子版本分批 · **前置:** [v0.7](./phase-v0.7.md)

> **子版本:** **v0.8.0** ER 图 → **v0.8.1** 插件扩展 → **v0.8.2** 驱动深化

承接原 v1.x / v2.0 候选中的平台级能力；与 v0.5–v0.7 的 UI / HTTP 主线独立，可并行规划但建议 **v0.7 完成后再开工**。

---

## 1. 版本路线

| Tag | 主题 | 文档章节 |
|-----|------|----------|
| **v0.8.0** | ER 图 | §2 |
| **v0.8.1** | 插件扩展点 | §3 |
| **v0.8.2** | 驱动深化 | §4 |

各子版本可独立 tag、独立验收；无强依赖顺序，但 ER 图与 Schema 服务耦合较紧，建议先于插件。

---

## 2. v0.8.0 — ER 图

### 目标

可视化表间关系（外键 / 推断关系），对标 Beekeeper、DBeaver 的 ER 视图；非 MVP 核心路径。

### 功能清单

- [ ] 从当前连接 Schema 提取表、列、外键（含 PG/MySQL/SQLite 差异）
- [ ] 后端：`SchemaService` 或专用 `ERService` 返回图模型（节点 + 边）
- [ ] 前端：ER 视图入口（MenuBar 或 Schema Tab 子页）；平移/缩放；表节点点击跳转表结构
- [ ] 大库性能：限制初始渲染表数或按 schema 过滤
- [ ] 仅读展示（v0.8.0 不做 ER 上改表结构）

### 完成标准

- [ ] 至少 1 个远程库 + SQLite 上 ER 可浏览
- [ ] tag **`v0.8.0`**

---

## 3. v0.8.1 — 插件扩展点

### 目标

定义插件加载与扩展边界，为社区/内部扩展留接口；v0.8.1 以 **框架 + 示例插件** 为主，不做插件市场。

### 功能清单

- [ ] 插件 manifest 格式（id、version、hooks）
- [ ] 扩展点设计（初版）：菜单项注入、SQL 编辑器命令、连接类型（评估后裁剪）
- [ ] 桌面（Wails）加载路径：`~/.data-nexus/plugins/`；Server 模式策略文档化
- [ ] 沙箱与安全：仅本地可信插件；禁止任意 Go 动态库（若用 WASM/JS 插件需单独 ADR）
- [ ] 示例插件 1 个 + 开发者文档

### 完成标准

- [ ] 示例插件可在桌面端加载并执行一个可见扩展
- [ ] tag **`v0.8.1`**

---

## 4. v0.8.2 — 驱动深化

### 目标

统一远程连接运行时行为，扩展认证方式；导出能力已在 v0.2–v0.4 完成。

### 功能清单

- [ ] 连接池与超时策略统一（SQLite / PG / MySQL）
- [x] SQLite 整表 CSV keyset 流式导出（v0.2+）
- [x] PostgreSQL / MySQL 整表导出（v0.4，见 [EXPORT_MULTI_DIALECT.md](../design/EXPORT_MULTI_DIALECT.md)）
- [ ] 更多 SSL 模式（证书校验、客户端证书）
- [ ] 更多认证方式（按需求：IAM、Kerberos 等 — 实施前 ADR）

跨库 Schema 策略见 [DATA_MODEL.md §6](../design/DATA_MODEL.md)。

### 完成标准

- [ ] 连接池/超时在三种驱动上可配置且文档化
- [ ] 至少 1 项 SSL/认证增强落地
- [ ] tag **`v0.8.2`**

---

## 5. 非目标（v0.8 全系）

- 完整 BI / 图表
- 插件商店、在线安装
- SQLCipher 专用管理（可后续社区插件）
- 重复 v0.5–v0.7 已有能力

---

## 6. 相关文档

| 文档 | 用途 |
|------|------|
| [PRD.md](../product/PRD.md) | 平台能力优先级 |
| [COMPETITIVE_ANALYSIS.md](../product/COMPETITIVE_ANALYSIS.md) | ER / 插件竞品 |
| [DATA_MODEL.md §6](../design/DATA_MODEL.md) | 跨库 Schema |
| [phase-v0.7.md](./phase-v0.7.md) | 上一版本 |
