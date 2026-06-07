package sqlite_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
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
