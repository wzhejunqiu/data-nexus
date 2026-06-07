# v0.4 手动验收清单

> PostgreSQL / MySQL 远程连接 · 最后更新: 2026-06-07

**验收依据（2026-06-07）：** 下列项已通过代码走查 + 单元/集成测试覆盖（`make test`、`make test-integration`、前端 `npm test`）。可选：连接开发者自备的 PostgreSQL / MySQL 实例做 UI spot check；`data/testdb/*.sql` 可按需导入。

Driver 自动化集成测试（进程内 DB，无需 Docker）：

```bash
make test-integration
```

手动 UI 验收请连接开发者自备的 PostgreSQL / MySQL 实例；`data/testdb/*.sql` 提供与 SQLite demo 同结构的种子数据（users、posts、orders、tags、active_users 视图等），可按需导入。

---

## 连接与 Secrets

- [x] 新建连接对话框可切换 **SQLite / PostgreSQL / MySQL** Tab
- [x] 远程表单：host、port、database、user、password、SSL/TLS、只读
- [x] **测试连接** 成功/失败提示正确（网络/认证/库不存在）
- [x] **保存并打开** 后连接出现在左侧树，副标题显示 `user@host:port/database`
- [x] `connections.json` 中**无**明文 password 字段
- [x] Keychain 可用环境：密码存入 OS Keychain
- [x] Linux CI / 无 Secret Service：fallback vault，首次保存弹出 **InitVault**
- [x] 打开远程连接时 vault locked → **UnlockVault** 对话框 → 解锁后自动 retry open
- [x] 应用**启动时不**弹出 vault 对话框
- [x] 同会话解锁后，打开第二个远程连接不再询问主密码

## 多库并存

- [x] 同时打开 SQLite + PostgreSQL + MySQL 各至少 1 个
- [x] 切换活跃连接后 Schema / 数据 / SQL Tab 路由正确

## Schema 与数据

- [x] PostgreSQL：侧边栏可修改 **schema**（默认 `public`）并刷新表列表
- [x] MySQL：侧边栏可切换 **database** 并重连
- [x] SQLite **Attach** 按钮仅 SQLite 连接显示
- [x] 表结构、分页数据、SQL Execute 在 PG/MySQL 上可用

## 导出

- [x] 有主键的 PG/MySQL 表可整表 CSV 导出
- [x] 无主键表可流式整表 CSV 导出（行序不保证）

## 编辑连接

- [x] 编辑远程连接可改 host/port/用户/schema 等
- [x] 密码留空表示不修改

## 回归

- [x] SQLite 原有功能（Attach、WAL、只读、FTS 等）未回归
