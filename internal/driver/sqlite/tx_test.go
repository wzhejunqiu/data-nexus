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

func openWritableDB(t *testing.T) *sqlite.Driver {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tx.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, val TEXT)`); err != nil {
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
	return drv
}

func TestBeginTxExecQueryCommit(t *testing.T) {
	drv := openWritableDB(t)
	defer func() { _ = drv.Close() }()
	ctx := context.Background()

	tx, err := drv.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO t (val) VALUES (?)`, []any{"ok"}); err != nil {
		t.Fatal(err)
	}
	res, err := tx.QueryRows(ctx, `SELECT val FROM t`, nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 1 {
		t.Fatalf("expected 1 row, got %d", res.RowCount)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestBeginTxRollback(t *testing.T) {
	drv := openWritableDB(t)
	defer func() { _ = drv.Close() }()
	ctx := context.Background()

	tx, err := drv.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO t (val) VALUES (?)`, []any{"gone"}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	res, err := drv.QueryRows(ctx, `SELECT COUNT(*) AS n FROM t`, nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if res.Rows[0]["n"].(int64) != 0 {
		t.Fatalf("expected rollback, got %+v", res.Rows[0])
	}
}

func TestBeginTxNotConnected(t *testing.T) {
	drv := sqlite.New()
	_, err := drv.BeginTx(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBeginTxReadOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ro.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	cfg := model.DriverConfig{
		Type: model.DriverTypeSQLite,
		SQLite: &model.SQLiteConfig{
			FilePath: path,
			ReadOnly: true,
		},
	}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	_, err = drv.BeginTx(context.Background())
	if err == nil {
		t.Fatal("expected read-only error")
	}
}
