# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-06-07

First public MVP release — a lightweight desktop SQLite manager for Windows and Linux.

### Added

- Multi-connection workspace: open and manage several SQLite databases at once
- Connection tree with saved connections persisted to `~/.data-nexus/connections.json`
- Table and view browser with schema details (columns and indexes)
- Paginated data grid with column sorting (25 / 50 / 100 / 200 rows per page)
- SQL editor powered by Monaco with `Cmd/Ctrl + Enter` to run queries
- Write-operation confirmation for `INSERT`, `UPDATE`, and `DELETE`
- Read-only connection mode to prevent accidental writes
- Session query history and one-click CSV copy for result sets
- PRAGMA shortcuts for common SQLite introspection
- Light / dark / system theme and Chinese / English UI (i18n)
- Startup options: `--db` to open a database on launch
- Configurable logging via `~/.data-nexus/config.yaml` and CLI flags

### Platforms

Pre-built binaries are published for:

- Windows (amd64 and arm64) — NSIS installers
- Linux (amd64 and arm64) — zip archives

macOS builds are temporarily omitted from Release until code signing and notarization are configured. macOS users can build from source.

[0.1.0]: https://github.com/wzhejunqiu/data-nexus/releases/tag/v0.1.0
