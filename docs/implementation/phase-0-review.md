# Phase 0 — 文档与设计评审

> **状态:** 进行中 · **预估:** 1-2 天 · **前置:** 无 · **下一阶段:** [phase-1-wails-backend.md](./phase-1-wails-backend.md)

---

## 目标

产品与技术方案对齐，评审通过后进入编码。

---

## 交付物检查

- [x] [PRD.md](../product/PRD.md)
- [x] [COMPETITIVE_ANALYSIS.md](../product/COMPETITIVE_ANALYSIS.md)
- [x] [ARCHITECTURE.md](../design/ARCHITECTURE.md)
- [x] [API.md](../design/API.md)
- [x] [DATA_MODEL.md](../design/DATA_MODEL.md)
- [x] [UI_UX.md](../design/UI_UX.md)
- [x] [ROADMAP.md](../ROADMAP.md)
- [x] [implementation/](./README.md) 分阶段实施清单
- [ ] **评审签字 / 确认**（团队或本人确认可开工）

---

## 评审检查项

- [ ] MVP 范围合理（P0 可在一个迭代内闭环）
- [ ] Wails Service 绑定覆盖全部前端场景（对照 [API.md](../design/API.md)）
- [ ] Driver 抽象可支撑 v1.x 多数据库（对照 [ARCHITECTURE.md §11](../design/ARCHITECTURE.md)）
- [ ] zap 日志方案可接受（dev 控制台 / release 文件 / 用户 YAML）
- [ ] 开放问题已决策（见 [PRD §8](../product/PRD.md)）

### 已决策项（无需再讨论）

| 问题 | 决策 |
|------|------|
| 只读模式 | 连接时可勾选，默认读写 |
| 文件选择 | MVP 原生对话框 + 可选路径输入 |
| 桌面 vs Web | Wails v2 桌面 |
| 日志 | zap，默认 INFO |

---

## 开工前环境准备

```bash
# Go 1.22+
go version

# Node 20+
node -v

# Wails v2 CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

- [ ] `wails doctor` 全部通过

---

## 完成标准

- [ ] 所有评审检查项已勾选
- [ ] 环境就绪
- [ ] 打开 [phase-1-wails-backend.md](./phase-1-wails-backend.md) 开始实施
