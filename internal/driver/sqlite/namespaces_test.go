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

func createDBWithTable(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY);`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestListNamespaces(t *testing.T) {
	mainPath := openEmptyDB(t, "main.db")
	otherPath := openEmptyDB(t, "other.db")

	drv := sqlite.New()
	cfg := model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: mainPath}}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	ctx := context.Background()
	namespaces, err := drv.ListNamespaces(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(namespaces) != 1 || namespaces[0].Name != "main" || namespaces[0].Kind != model.NamespaceKindDatabase {
		t.Fatalf("unexpected namespaces: %+v", namespaces)
	}

	if err := drv.Attach(ctx, otherPath, "other"); err != nil {
		t.Fatal(err)
	}
	namespaces, err = drv.ListNamespaces(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(namespaces) != 2 || namespaces[1].Name != "other" || namespaces[1].Kind != model.NamespaceKindAttach {
		t.Fatalf("unexpected namespaces after attach: %+v", namespaces)
	}
	if namespaces[1].FilePath == nil || *namespaces[1].FilePath != otherPath {
		t.Fatalf("expected file path %q, got %+v", otherPath, namespaces[1].FilePath)
	}
}

func TestListTablesByNamespace(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.db")
	otherPath := filepath.Join(dir, "other.db")
	createDBWithTable(t, mainPath)
	createDBWithTable(t, otherPath)

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

	mainTables, err := drv.ListTables(ctx, model.ListTablesOptions{Database: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if len(mainTables) != 1 || mainTables[0].Name != "t" {
		t.Fatalf("unexpected main tables: %+v", mainTables)
	}

	otherTables, err := drv.ListTables(ctx, model.ListTablesOptions{Database: "other"})
	if err != nil {
		t.Fatal(err)
	}
	if len(otherTables) != 1 || otherTables[0].Name != "t" {
		t.Fatalf("unexpected other tables: %+v", otherTables)
	}
}
