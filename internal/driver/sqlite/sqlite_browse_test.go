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

	_, err := drv.GetTableSchema(context.Background(), `bad"name`)
	if err == nil {
		t.Fatal("expected error for invalid table name")
	}
}

func TestBrowseTableSortDesc(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	data, err := drv.BrowseTable(context.Background(), "items", model.BrowseOptions{
		Page: 1, PageSize: 10, Sort: "id", Order: model.SortDesc,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) < 2 {
		t.Fatalf("expected rows, got %d", len(data.Rows))
	}
	firstID := data.Rows[0]["id"]
	lastID := data.Rows[len(data.Rows)-1]["id"]
	if firstID.(int64) <= lastID.(int64) {
		t.Fatalf("expected descending order, got first=%v last=%v", firstID, lastID)
	}
}

func TestBrowseTablePaginationDefaults(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	data, err := drv.BrowseTable(context.Background(), "items", model.BrowseOptions{
		Page: 0, PageSize: 500,
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.Page != 1 {
		t.Fatalf("expected page 1, got %d", data.Pagination.Page)
	}
	if data.Pagination.PageSize != 200 {
		t.Fatalf("expected pageSize 200, got %d", data.Pagination.PageSize)
	}
}

func TestBrowseTableEmptyTable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE empty_t (id INTEGER PRIMARY KEY)`); err != nil {
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

	data, err := drv.BrowseTable(context.Background(), "empty_t", model.BrowseOptions{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 0 || data.Pagination.TotalPages != 1 {
		t.Fatalf("unexpected pagination: %+v", data.Pagination)
	}
	if data.Rows == nil {
		t.Fatal("expected non-nil empty rows slice")
	}
}

func TestBrowseTableWithFilters(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	val := "a"
	data, err := drv.BrowseTable(context.Background(), "items", model.BrowseOptions{
		Page: 1, PageSize: 50,
		Filters: []model.RowFilter{{Column: "name", Operator: model.FilterEq, Value: &val}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 1 {
		t.Fatalf("expected 1 filtered row, got %d", data.Pagination.TotalRows)
	}
}

func TestQueryRowsZeroRows(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	res, err := drv.QueryRows(context.Background(), "SELECT * FROM items WHERE 1=0", nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 0 || res.Rows == nil {
		t.Fatalf("expected empty slice, got %+v", res)
	}
}

func TestQueryRowsMaxRowsClamping(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	res, err := drv.QueryRows(context.Background(), "SELECT id FROM items", nil, -1)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 3 {
		t.Fatalf("expected 3 rows with default max, got %d", res.RowCount)
	}

	res, err = drv.QueryRows(context.Background(), "SELECT id FROM items", nil, model.MaxQueryRows+100)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 3 {
		t.Fatalf("expected 3 rows with clamped max, got %d", res.RowCount)
	}
}
