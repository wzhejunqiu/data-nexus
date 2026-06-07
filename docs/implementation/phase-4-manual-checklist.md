# Phase 4 — Manual Test Checklist (v0.1.0)

> Record results when validating a release candidate. See [phase-4-release.md](./phase-4-release.md).

## Startup & Connections

- [ ] Cold start window visible < 1s
- [ ] File → Open selects valid `.db` and opens connection
- [ ] New connection dialog supports path input + browse
- [ ] Invalid path shows localized error
- [ ] Read-only connection blocks write SQL
- [ ] Close / reopen connection works
- [ ] Connection persists in `~/.data-nexus/connections.json` after restart
- [ ] Remove saved connection works
- [ ] Restore open connections on startup (when enabled)
- [ ] Drag `.db` onto window opens connection
- [ ] Double-click closed connection opens it

## Schema & Data

- [ ] Tables/views listed (no `sqlite_` system tables)
- [ ] Table schema shows columns + indexes
- [ ] Pagination 25/50/100/200
- [ ] Column sort asc/desc/clear
- [ ] NULL and BLOB display correctly

## SQL

- [ ] Cmd/Ctrl+Enter runs query from Monaco focus
- [ ] SELECT shows rows + duration in status bar
- [ ] Write SQL shows confirm dialog
- [ ] Query history populated
- [ ] CSV copy works
- [ ] PRAGMA shortcuts work

## Desktop & i18n

- [ ] Window title `Data Nexus — {filename}`
- [ ] File / View / Help menus work
- [ ] Theme light/dark/system
- [ ] Language zh-CN / en
- [ ] Release build logs to platform default file path
- [ ] `wails build` smoke test passes on target platform
