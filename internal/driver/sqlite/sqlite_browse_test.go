package sqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/driver/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sqlite.Driver {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT);
		INSERT INTO items (name) VALUES ('a'), ('b'), ('c');`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	cfg := model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path}}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	return drv
}

func TestBrowseTableSortAndPagination(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	page1, err := drv.BrowseTable(context.Background(), "items", model.BrowseOptions{
		Page: 1, PageSize: 2, Sort: "id", Order: model.SortAsc,
	})
	if err != nil {
		t.Fatal(err)
	}
	if page1.Pagination.TotalRows != 3 || len(page1.Rows) != 2 {
		t.Fatalf("unexpected page1: %+v", page1.Pagination)
	}

	page2, err := drv.BrowseTable(context.Background(), "items", model.BrowseOptions{
		Page: 2, PageSize: 2, Sort: "id", Order: model.SortAsc,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Rows) != 1 {
		t.Fatalf("expected 1 row on page2, got %d", len(page2.Rows))
	}
}

func TestQueryRows(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	res, err := drv.QueryRows(context.Background(), "SELECT * FROM items", nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 3 {
		t.Fatalf("expected 3 rows, got %d", res.RowCount)
	}
}

func TestInvalidTableName(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	_, err := drv.GetTableSchema(context.Background(), "bad-name")
	if err == nil {
		t.Fatal("expected error for invalid table name")
	}
}
