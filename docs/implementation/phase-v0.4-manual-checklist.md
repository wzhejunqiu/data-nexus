# v0.4 手动验收清单

> PostgreSQL / MySQL 远程连接 · 最后更新: 2026-06-07

Driver 自动化集成测试（进程内 DB，无需 Docker）：

```bash
make test-integration
```

手动验收请连接真实 PostgreSQL / MySQL 实例；`data/testdb/*.sql` 提供与 SQLite demo 同结构的种子数据（users、posts、orders、tags、active_users 视图等），可按需导入。

---

## 连接与 Secrets

- [ ] 新建连接对话框可切换 **SQLite / PostgreSQL / MySQL** Tab
- [ ] 远程表单：host、port、database、user、password、SSL/TLS、只读
- [ ] **测试连接** 成功/失败提示正确（网络/认证/库不存在）
- [ ] **保存并打开** 后连接出现在左侧树，副标题显示 `user@host:port/database`
- [ ] `connections.json` 中**无**明文 password 字段
- [ ] Keychain 可用环境：密码存入 OS Keychain
- [ ] Linux CI / 无 Secret Service：fallback vault，首次保存弹出 **InitVault**
- [ ] 打开远程连接时 vault locked → **UnlockVault** 对话框 → 解锁后自动 retry open
- [ ] 应用**启动时不**弹出 vault 对话框
- [ ] 同会话解锁后，打开第二个远程连接不再询问主密码

## 多库并存

- [ ] 同时打开 SQLite + PostgreSQL + MySQL 各至少 1 个
- [ ] 切换活跃连接后 Schema / 数据 / SQL Tab 路由正确

## Schema 与数据

- [ ] PostgreSQL：侧边栏可修改 **schema**（默认 `public`）并刷新表列表
- [ ] MySQL：侧边栏可切换 **database** 并重连
- [ ] SQLite **Attach** 按钮仅 SQLite 连接显示
- [ ] 表结构、分页数据、SQL Execute 在 PG/MySQL 上可用

## 导出

- [ ] 有主键的 PG/MySQL 表可整表 CSV 导出
- [ ] 无主键表可流式整表 CSV 导出（行序不保证）

## 编辑连接

- [ ] 编辑远程连接可改 host/port/用户/schema 等
- [ ] 密码留空表示不修改

## 回归

- [ ] SQLite 原有功能（Attach、WAL、只读、FTS 等）未回归
