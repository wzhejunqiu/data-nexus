# Phase 4 — Manual Test Checklist (v0.1.0)

> Record results when validating a release candidate. See [phase-4-release.md](./phase-4-release.md).
>
> **更新 (2026-06-07):** 下列 `[x]` 项已在代码中实现并有自动化测试覆盖；发布前仍建议在本机走查一遍。标注 ⏳ 的项需目标平台手动验证。

## Startup & Connections

- [ ] ⏳ Cold start window visible < 1s
- [x] File → Open selects valid `.db` and opens connection
- [x] New connection dialog supports path input + browse
- [x] Invalid path shows localized error
- [x] Read-only connection blocks write SQL
- [x] Close / reopen connection works
- [x] Connection persists in `~/.data-nexus/connections.json` after restart
- [x] Remove saved connection works
- [x] Restore open connections on startup (when enabled)
- [x] Drag `.db` onto window opens connection
- [x] Double-click closed connection opens it

## Schema & Data

- [x] Tables/views listed (no `sqlite_` system tables)
- [x] Table schema shows columns + indexes
- [x] Pagination 25/50/100/200
- [x] Column sort asc/desc/clear
- [x] NULL and BLOB display correctly

## SQL

- [x] Cmd/Ctrl+Enter runs query from Monaco focus
- [x] SELECT shows rows + duration in status bar
- [x] Write SQL shows confirm dialog
- [x] Query history populated
- [x] CSV copy works
- [x] PRAGMA shortcuts work

## Desktop & i18n

- [x] Window title `Data Nexus — {filename}`
- [x] File / View / Help menus work
- [x] Theme light/dark/system
- [x] Language zh-CN / en
- [x] Release build logs to platform default file path
- [ ] ⏳ Release 产物或本机 `wails build` smoke test 通过（目标平台；main CI 仅覆盖 linux-amd64）
