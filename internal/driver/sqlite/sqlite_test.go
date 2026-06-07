package sqlite_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/driver/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	_ "modernc.org/sqlite"
)

func TestDriverListTablesAndBrowse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT); INSERT INTO users (email) VALUES ('a@example.com');`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	cfg := model.DriverConfig{
		Type:   model.DriverTypeSQLite,
		SQLite: &model.SQLiteConfig{FilePath: path},
	}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	tables, err := drv.ListTables(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 1 || tables[0].Name != "users" {
		t.Fatalf("unexpected tables: %+v", tables)
	}
	if tables[0].RowCount == nil || *tables[0].RowCount != 1 {
		t.Fatalf("expected rowCount 1, got %+v", tables[0].RowCount)
	}

	schema, err := drv.GetTableSchema(context.Background(), "users")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(schema.Columns))
	}

	data, err := drv.BrowseTable(context.Background(), "users", model.BrowseOptions{Page: 1, PageSize: 50, Order: model.SortAsc})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 1 {
		t.Fatalf("expected 1 row, got %d", data.Pagination.TotalRows)
	}
}

func TestSerializeBlob(t *testing.T) {
	val := sqlite.SerializeCellValue([]byte{1, 2, 3})
	m, ok := val.(map[string]any)
	if !ok || m["type"] != "blob" || m["size"] != 3 {
		t.Fatalf("unexpected blob serialization: %#v", val)
	}
}

func TestDriverQueryRowsTooLarge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER);`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if _, err := db.Exec(`INSERT INTO t (id) VALUES (?)`, i); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.Close()

	drv := sqlite.New()
	cfg := model.DriverConfig{
		Type:   model.DriverTypeSQLite,
		SQLite: &model.SQLiteConfig{FilePath: path},
	}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	_, err = drv.QueryRows(context.Background(), "SELECT id FROM t", nil, 3)
	if err == nil {
		t.Fatal("expected RESULT_TOO_LARGE")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "RESULT_TOO_LARGE" {
		t.Fatalf("expected RESULT_TOO_LARGE, got %v", err)
	}
}

func TestDriverBrowseSort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sort.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER, name TEXT);
		INSERT INTO t VALUES (2,'b'),(1,'a');`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	data, err := drv.BrowseTable(context.Background(), "t", model.BrowseOptions{
		Page: 1, PageSize: 10, Sort: "id", Order: model.SortAsc,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(data.Rows))
	}
}

func TestDriverGetTableSchemaIndexes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "idx.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE users (email TEXT UNIQUE);`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	schema, err := drv.GetTableSchema(context.Background(), "users")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Indexes) == 0 {
		t.Fatal("expected at least one index")
	}
}

func TestDriverReadOnlyExecRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ro.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	drv := sqlite.New()
	cfg := model.DriverConfig{
		Type:   model.DriverTypeSQLite,
		SQLite: &model.SQLiteConfig{FilePath: path, ReadOnly: true},
	}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	_, err = drv.Exec(context.Background(), "CREATE TABLE t(id INTEGER)", nil)
	if err == nil {
		t.Fatal("expected read-only error")
	}
}

func TestListTablesExcludesSystemTables(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sys.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE users (id INTEGER); INSERT INTO users (id) VALUES (1); ANALYZE users;`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	tables, err := drv.ListTables(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, tbl := range tables {
		if tbl.Name == "sqlite_stat1" {
			t.Fatal("sqlite_ system table should be excluded")
		}
	}
	if len(tables) != 1 || tables[0].Name != "users" {
		t.Fatalf("unexpected tables: %+v", tables)
	}
}

func TestListTablesIncludesView(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "view.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER); CREATE VIEW v AS SELECT id FROM t;`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	tables, err := drv.ListTables(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var hasView bool
	for _, tbl := range tables {
		if tbl.Name == "v" && tbl.Type == model.TableTypeView {
			hasView = true
		}
	}
	if !hasView {
		t.Fatalf("expected view in list: %+v", tables)
	}
}

func TestSerializeNull(t *testing.T) {
	if sqlite.SerializeCellValue(nil) != nil {
		t.Fatal("expected nil for null value")
	}
}

func TestBrowseInvalidSortColumn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sort.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER);`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	_, err = drv.BrowseTable(context.Background(), "t", model.BrowseOptions{
		Page: 1, PageSize: 50, Sort: `bad"name`,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}

func TestDriverTypeAndReadOnly(t *testing.T) {
	drv := sqlite.New()
	if drv.Type() != model.DriverTypeSQLite {
		t.Fatalf("unexpected type %s", drv.Type())
	}
	if drv.ReadOnly() {
		t.Fatal("expected read-write by default")
	}
}

func TestDriverConnectWALMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wal.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite,
		SQLite: &model.SQLiteConfig{
			FilePath: path,
			WAL:      true,
		},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	res, err := drv.QueryRows(context.Background(), "PRAGMA journal_mode", nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 1 {
		t.Fatalf("expected 1 row, got %d", res.RowCount)
	}
	mode, ok := res.Rows[0]["journal_mode"].(string)
	if !ok || !strings.EqualFold(mode, "wal") {
		t.Fatalf("expected wal journal mode, got %+v", res.Rows[0])
	}
}

func TestDriverConnectNilSQLiteConfig(t *testing.T) {
	drv := sqlite.New()
	err := drv.Connect(context.Background(), model.DriverConfig{Type: model.DriverTypeSQLite})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}

func TestDriverConnectReconnect(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reconnect.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	drv := sqlite.New()
	cfg := model.DriverConfig{
		Type:   model.DriverTypeSQLite,
		SQLite: &model.SQLiteConfig{FilePath: path},
	}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	if err := drv.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestDriverConnectInvalidPath(t *testing.T) {
	drv := sqlite.New()
	err := drv.Connect(context.Background(), model.DriverConfig{
		Type:   model.DriverTypeSQLite,
		SQLite: &model.SQLiteConfig{FilePath: "/nonexistent/path/to/db.db"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_FAILED" {
		t.Fatalf("expected CONNECTION_FAILED, got %v", err)
	}
}

func TestDriverPingNotConnected(t *testing.T) {
	drv := sqlite.New()
	err := drv.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_NOT_FOUND" {
		t.Fatalf("expected CONNECTION_NOT_FOUND, got %v", err)
	}
}

func TestDriverExecInsertSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exec.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	res, err := drv.Exec(context.Background(), `INSERT INTO t (name) VALUES ('alice')`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsAffected != 1 || res.LastInsertID != 1 {
		t.Fatalf("unexpected exec result: %+v", res)
	}
}

func TestDriverGetTableSchemaTableNotFound(t *testing.T) {
	drv := openTestDBFromPath(t, filepath.Join(t.TempDir(), "schema.db"))
	defer func() { _ = drv.Close() }()

	_, err := drv.GetTableSchema(context.Background(), "missing_table")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "TABLE_NOT_FOUND" {
		t.Fatalf("expected TABLE_NOT_FOUND, got %v", err)
	}
}

func TestDriverGetTableSchemaWithDefaultValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "default.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, status TEXT DEFAULT 'active')`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	schema, err := drv.GetTableSchema(context.Background(), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(schema.Columns))
	}
	def := schema.Columns[1].DefaultValue
	if def == nil || *def != "'active'" {
		t.Fatalf("expected default value, got %+v", def)
	}
}

func TestDriverGetTableSchemaPrimaryAndUniqueIndexes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "idx2.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE users (
		a INTEGER,
		b INTEGER,
		email TEXT UNIQUE,
		PRIMARY KEY (a, b)
	)`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	schema, err := drv.GetTableSchema(context.Background(), "users")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Indexes) == 0 {
		t.Fatal("expected indexes")
	}
	var hasPrimary bool
	for _, idx := range schema.Indexes {
		if idx.Primary {
			hasPrimary = true
		}
	}
	if !hasPrimary {
		t.Fatalf("expected primary index, got %+v", schema.Indexes)
	}
}

func TestDriverGetTableSchemaViewType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "viewschema.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER); CREATE VIEW v AS SELECT id FROM t`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	schema, err := drv.GetTableSchema(context.Background(), "v")
	if err != nil {
		t.Fatal(err)
	}
	if schema.Type != model.TableTypeView {
		t.Fatalf("expected view type, got %s", schema.Type)
	}
}

func openTestDBFromPath(t *testing.T, path string) *sqlite.Driver {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	return drv
}
