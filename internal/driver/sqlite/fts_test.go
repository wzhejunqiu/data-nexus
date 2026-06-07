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

func openDBWithFTS(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT);
		INSERT INTO items (name) VALUES ('alpha'), ('beta'), ('alphabet');
		CREATE VIRTUAL TABLE items_fts USING fts5(name, content='items', content_rowid='id');
		INSERT INTO items_fts(rowid, name) VALUES (1, 'alpha'), (2, 'beta'), (3, 'alphabet');
	`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestDetectFTSOnMain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fts.db")
	openDBWithFTS(t, path)

	drv := sqlite.New()
	cfg := model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path}}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	ctx := context.Background()
	info, err := drv.DetectFTSTable(ctx, "items")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Enabled || info.FTSTableName != "items_fts" || info.Schema != "main" {
		t.Fatalf("unexpected fts info: %+v", info)
	}
}

func TestDetectFTSOnAttached(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.db")
	otherPath := filepath.Join(dir, "other.db")

	db, err := sql.Open("sqlite", mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE placeholder (id INTEGER PRIMARY KEY);`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	openDBWithFTS(t, otherPath)

	drv := sqlite.New()
	cfg := model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: mainPath}}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	ctx := context.Background()
	if err := drv.Attach(ctx, otherPath, "other"); err != nil {
		t.Fatal(err)
	}

	info, err := drv.DetectFTSTable(ctx, "other.items")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Enabled || info.FTSTableName != "items_fts" || info.Schema != "other" {
		t.Fatalf("unexpected attached fts info: %+v", info)
	}
}

func TestBrowseTableWithFTSSearch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fts.db")
	openDBWithFTS(t, path)

	drv := sqlite.New()
	cfg := model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path}}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	data, err := drv.BrowseTable(context.Background(), "items", model.BrowseOptions{
		Page: 1, PageSize: 50, Search: "alphabet",
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 1 {
		t.Fatalf("expected 1 fts row, got %d", data.Pagination.TotalRows)
	}
	if len(data.Rows) != 1 || data.Rows[0]["name"] != "alphabet" {
		t.Fatalf("unexpected rows: %+v", data.Rows)
	}
}

func TestBrowseAttachedTableWithFTSSearch(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.db")
	otherPath := filepath.Join(dir, "other.db")

	db, err := sql.Open("sqlite", mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE placeholder (id INTEGER PRIMARY KEY);`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	openDBWithFTS(t, otherPath)

	drv := sqlite.New()
	cfg := model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: mainPath}}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	ctx := context.Background()
	if err := drv.Attach(ctx, otherPath, "other"); err != nil {
		t.Fatal(err)
	}

	data, err := drv.BrowseTable(ctx, "other.items", model.BrowseOptions{
		Page: 1, PageSize: 50, Search: "beta",
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 1 {
		t.Fatalf("expected 1 attached fts row, got %d", data.Pagination.TotalRows)
	}
	if len(data.Rows) != 1 || data.Rows[0]["name"] != "beta" {
		t.Fatalf("unexpected rows: %+v", data.Rows)
	}
}
