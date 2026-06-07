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

func openRowidTestDB(t *testing.T, ddl string) (*sqlite.Driver, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ddl); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	cfg := model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path}}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	return drv, path
}

func TestUpdateCellsRowid(t *testing.T) {
	drv, _ := openRowidTestDB(t, `CREATE TABLE items (name TEXT); INSERT INTO items VALUES ('a'), ('b')`)
	defer func() { _ = drv.Close() }()
	ctx := context.Background()

	schema, err := drv.GetTableSchema(ctx, "items")
	if err != nil {
		t.Fatal(err)
	}

	data, err := drv.BrowseTable(ctx, "items", model.BrowseOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 2 || data.Rows[0]["rowid"] == nil {
		t.Fatalf("expected rowid in browse rows, got %+v", data.Rows[0])
	}

	rowid := data.Rows[0]["rowid"]
	count, err := drv.UpdateCells(ctx, "items", []model.CellChange{
		{ColumnName: "name", PrimaryKey: map[string]any{"rowid": rowid}, NewValue: "updated"},
	}, schema)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 update, got %d", count)
	}

	data2, err := drv.BrowseTable(ctx, "items", model.BrowseOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if data2.Rows[0]["name"] != "updated" {
		t.Fatalf("expected updated, got %+v", data2.Rows[0])
	}
}

func TestUpdateCellsWithoutRowIDUsesPK(t *testing.T) {
	drv, _ := openRowidTestDB(t, `CREATE TABLE wr (id INTEGER PRIMARY KEY, msg TEXT) WITHOUT ROWID; INSERT INTO wr VALUES (1,'x')`)
	defer func() { _ = drv.Close() }()
	ctx := context.Background()

	schema, err := drv.GetTableSchema(ctx, "wr")
	if err != nil {
		t.Fatal(err)
	}

	data, err := drv.BrowseTable(ctx, "wr", model.BrowseOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range data.Columns {
		if c.Name == "rowid" {
			t.Fatal("WITHOUT ROWID table with PK should not expose rowid column")
		}
	}

	count, err := drv.UpdateCells(ctx, "wr", []model.CellChange{
		{ColumnName: "msg", PrimaryKey: map[string]any{"id": int64(1)}, NewValue: "y"},
	}, schema)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 update, got %d", count)
	}
}

func TestBrowseTableRowidColumn(t *testing.T) {
	drv, _ := openRowidTestDB(t, `CREATE TABLE t (v TEXT); INSERT INTO t VALUES ('one')`)
	defer func() { _ = drv.Close() }()
	ctx := context.Background()

	data, err := drv.BrowseTable(ctx, "t", model.BrowseOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	hasRowidCol := false
	for _, c := range data.Columns {
		if c.Name == "rowid" {
			hasRowidCol = true
		}
	}
	if !hasRowidCol {
		t.Fatal("expected rowid column in metadata")
	}
}
