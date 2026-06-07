# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-06-07

### Added

- PostgreSQL and MySQL drivers (`internal/driver/postgres`, `internal/driver/mysql`) with schema browse, data grid, SQL, profiling, and table CSV export
- Remote connection form in `NewConnectionDialog` (SQLite / PostgreSQL / MySQL tabs) with **Test connection**
- Secrets storage: OS Keychain preferred, RSA/AES file vault fallback with user master password ([docs/design/SECRETS.md](docs/design/SECRETS.md))
- `SecretsService` Wails bindings and on-demand `VaultDialog` (including edit-connection save retry)
- `CreateRemoteConnection`, `TestConnection`, `UpdateConnectionPostgresSettings`, `UpdateConnectionMySQLSettings`
- PostgreSQL schema / MySQL database switcher in connection tree sidebar
- In-process driver integration tests (`go-mysql-server` + `embedded-postgres`; `make test-integration`)

### Changed

- **整表 CSV 导出**：SQLite / PostgreSQL / MySQL 统一为流式单查询（`SELECT *` + `NextBatch` 读 `sql.Rows`），移除 keyset 分页；无 PK 表亦可导出（行序不保证）
- Connection tree shows type badge (SQL/PG/MY) and remote connection subtitle
- Attach database UI limited to SQLite connections
- Product version bumped to 0.4.0

## [0.3.0] - 2026-06-07

### Added

- Column data profiling (`GetTableProfile`) with sampling for large tables and `ColumnProfile` UI in Schema tab
- Filter Builder with backend `BrowseRows.filters[]` and「在 SQL 编辑器中编辑」hybrid flow
- Canned Queries (`CannedQueryService`, `~/.data-nexus/queries.json`, sidebar `SavedQueries`)
- Workspace state persistence (table filters, sort, tab) via `localStorage`
- FTS table detection and conditional full-text search on data tab
- Facet panel for low-cardinality columns
- SQLite `ATTACH` / `DETACH` with sidebar schema grouping
- SQL execution history persistence (`ExecutionLogStore`, `~/.data-nexus/sql-global.db`, `SqlExecutionService`); SQL editor history loaded per `connectionId` via React Query

### Changed

- Product version bumped to 0.3.0
- Single-source versioning: `wails.json` `productVersion` synced to Go (`internal/version`) and frontend via `make sync-version`

## [0.2.0] - 2026-06-07

### Added

- Multi-cell inline table editing with pending changes bar and batch submit (`UpdateCellsBatch`, transactional)
- CSV export from data grid (current page / full table) and SQL results with configurable `CSVFormatOptions`
- CSV import wizard: new or existing table, append or update modes, column mapping, preview
- Monaco SQL autocomplete for tables/columns, SQL formatting (`sql-formatter`), EXPLAIN query plan tree
- Settings dialog for `~/.data-nexus/config.yaml` (log level, output, file path)
- `FileService`, `ExportService`, `ImportService`, `ConfigService` Wails bindings
- Positional CLI args and drag-and-drop to open `.db` files (installer does **not** register file association)
- `DialogService.SaveFile`, `OpenCSVFile`

### Changed

- Product version bumped to 0.2.0

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

[0.3.0]: https://github.com/wzhejunqiu/data-nexus/releases/tag/v0.3.0
[0.2.0]: https://github.com/wzhejunqiu/data-nexus/releases/tag/v0.2.0
[0.1.0]: https://github.com/wzhejunqiu/data-nexus/releases/tag/v0.1.0
