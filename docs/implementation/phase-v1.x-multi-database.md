# v1.x — Headless 与平台能力

> **预估:** 按特性分批 · **前置:** v1.0（PostgreSQL / MySQL 已在 v1.0 落地）

---

## 目标

v1.0 完成远程 SQL 库驱动后，v1.x 聚焦 **非桌面 UI** 能力与 **平台级** 扩展。

> **说明:** PostgreSQL / MySQL Driver、连接表单、Schema/SQL 闭环见 [phase-v1.0-multi-connection.md](./phase-v1.0-multi-connection.md)。

---

## Headless 模式（优先）

- [ ] `--server` flag 启动 chi REST
- [ ] 映射 [API.md Headless 表](../design/API.md) 到同一 service 层
- [ ] 用于 CI / 脚本 / 内网共享
- [ ] 可选 Basic Auth

---

## 驱动深化（按需）

- [ ] 连接池与超时策略统一
- [x] SQLite 整表 CSV keyset 流式导出（v0.2+ / Phase 1 cursor 重构）
- [ ] PostgreSQL / MySQL 整表导出（见 [EXPORT_MULTI_DIALECT.md](../design/EXPORT_MULTI_DIALECT.md)）
- [ ] 更多 SSL / 认证方式（证书、IAM 等，按需求）

---

## 平台能力（v2.0 候选）

- [ ] ER 图
- [ ] 插件扩展点
- [ ] 连接配置 keychain 集成（跨平台）

---

## 完成标准

- [ ] Headless：`--server` 下 REST 与桌面 Service 行为一致
- [ ] tag `v1.1.0` 起按特性分批发布

---

## 相关文档

- [ARCHITECTURE.md §2.3](../design/ARCHITECTURE.md) — Headless 扩展说明
- [DATA_MODEL.md §6](../design/DATA_MODEL.md) — 跨库 Schema 差异
