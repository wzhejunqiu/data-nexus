# Data Nexus — 远程连接密码存储（Secrets）

> 版本: v1.0 · 最后更新: 2026-06-07

---

## 1. 目标

- 远程连接（PostgreSQL / MySQL）密码**绝不**写入 `catalog.db`
- 优先使用操作系统 Keychain / Secret Service
- Keychain 不可用时 fallback 到本地 **Vault**（RSA-4096 + AES-256），由用户主密码保护
- **按需解锁**：仅打开/保存/编辑远程连接密码时询问主密码；应用启动不弹窗

---

## 2. Backend 选择

启动时 `secrets.NewStore()` **Probe** 环境：

| 顺序 | Backend | 条件 |
|------|---------|------|
| 1 | `keychain` | `zalando/go-keyring` 可用（macOS Keychain、Windows Credential Manager、Linux Secret Service） |
| 2 | `vault` | Probe 失败 → 文件 vault |

用户**不能**手动选择 backend（v1.0 自动决策）。

`SavedConnection.secretsBackend` 记录创建/更新时使用的 backend，便于 UI 提示。

---

## 3. Keychain 路径

优先 backend。完整设计（架构图、Probe 流程、密码生命周期、前端感知）见 **[OS_KEYCHAIN.md](./OS_KEYCHAIN.md)**。

摘要：

- 库：`zalando/go-keyring`（macOS Keychain / Windows Credential Manager / Linux Secret Service）
- Service name: `data-nexus`；Account: 连接 `id`（ULID）
- `SetPassword` / `GetPassword` / `DeletePassword` 按连接 ID 读写
- Vault 相关接口为 no-op；无需主密码，启动即可 `GetPassword`

---

## 4. Vault 路径（fallback）

目录：`~/.data-nexus/vault/`

| 文件 | 内容 |
|------|------|
| `vault.json` | RSA 公钥（PEM）、AES 数据密钥（RSA-OAEP 加密）、各连接密码条目（AES-GCM） |
| `rsa_private.enc` | RSA 私钥（AES-GCM 加密，密钥由用户主密码 PBKDF2 派生） |

### 密钥层次

```
用户主密码 (PBKDF2-SHA256)
    └── 加密 RSA 私钥 (rsa_private.enc)
RSA 私钥
    └── 解密 AES 数据密钥 (vault.json)
AES 数据密钥
    └── 加密各连接 password 条目
```

### 状态机

| 状态 | GetPassword | SetPassword |
|------|-------------|-------------|
| 未初始化 | `SECRETS_VAULT_NOT_INITIALIZED` | 需先 `InitVault` |
| 已锁定 | `SECRETS_VAULT_LOCKED` | `SECRETS_VAULT_LOCKED` |
| 已解锁 | 正常 | 正常 |

同会话解锁后，打开多个远程连接不再重复询问。

---

## 5. 与 ConnectionManager 集成

| 操作 | Secrets 行为 |
|------|--------------|
| `CreateRemoteConnection` | `SetPassword(connID, password)` |
| `OpenConnection`（远程） | `GetPassword(connID)` 注入 `DriverConfig` |
| `RemoveConnection`（远程） | `DeletePassword(connID)` |
| `UpdatePostgres/MySQLSettings`（含新密码） | `SetPassword` |
| `TestConnection` | **不使用** vault；请求内 `password` 直接 Connect+Ping |
| `RestoreConnectionsOnStartup` | vault locked 时**跳过**远程连接 |

---

## 6. Wails API（SecretsService）

| 方法 | 说明 |
|------|------|
| `GetSecretsBackend()` | `keychain` \| `vault` |
| `VaultInitialized()` | vault 是否已创建 |
| `VaultUnlocked()` | vault 是否已解锁 |
| `InitVault(masterPassword)` | 首次创建 vault |
| `UnlockVault(masterPassword)` | 解锁 |
| `LockVault()` | 锁定（清内存私钥） |
| `ChangeVaultPassword(old, new)` | 更换主密码 |
| `IsVaultRequiredForRemote()` | 当前 backend 是否为 vault |

---

## 7. 错误码

| Code | 说明 |
|------|------|
| `SECRETS_VAULT_LOCKED` | vault 已锁定 |
| `SECRETS_VAULT_NOT_INITIALIZED` | vault 未初始化 |
| `SECRETS_VAULT_WRONG_PASSWORD` | 主密码错误 |

前端 `VaultDialog` 拦截上述错误并 retry 原操作。

---

## 8. 边界与非目标

- 密码不进 `catalog.db`、不进日志、不进 Wails 持久化绑定缓存
- 不跨设备同步 vault（换机需重新输入连接密码）
- v1.0 不支持 Touch ID / 生物识别解锁
- v1.0 不支持证书/IAM 等替代认证

交叉引用：[OS_KEYCHAIN.md](./OS_KEYCHAIN.md) · [DATA_MODEL.md](./DATA_MODEL.md) · [API.md](./API.md) · [CONNECTION_UX.md](./CONNECTION_UX.md)
