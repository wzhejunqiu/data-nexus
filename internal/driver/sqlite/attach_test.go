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

func openEmptyDB(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
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
	return path
}

func TestAttachDetachListAttached(t *testing.T) {
	mainPath := openEmptyDB(t, "main.db")
	otherPath := openEmptyDB(t, "other.db")

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

	attached, err := drv.ListAttached(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(attached) != 1 || attached[0].Alias != "other" || attached[0].FilePath != otherPath {
		t.Fatalf("unexpected attached: %+v", attached)
	}

	if err := drv.Attach(ctx, otherPath, "other"); err == nil {
		t.Fatal("expected error for duplicate alias")
	}
	if err := drv.Attach(ctx, otherPath, "main"); err == nil {
		t.Fatal("expected error for reserved alias")
	}

	if err := drv.Detach(ctx, "other"); err != nil {
		t.Fatal(err)
	}
	attached, err = drv.ListAttached(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(attached) != 0 {
		t.Fatalf("expected no attached databases, got %+v", attached)
	}
}
