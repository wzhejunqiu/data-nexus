package sqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/driver/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	_ "modernc.org/sqlite"
)

func openExportTestDB(t *testing.T, total int) *sqlite.Driver {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "export_order.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	if total > 0 {
		_, err = db.Exec(fmt.Sprintf(`
			WITH RECURSIVE cnt(x) AS (
				SELECT 1 UNION ALL SELECT x + 1 FROM cnt WHERE x < %d
			)
			INSERT INTO items(id, name) SELECT x, 'row-' || x FROM cnt`, total))
		if err != nil {
			t.Fatal(err)
		}
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

func TestExportCursorStreaming(t *testing.T) {
	const total = 2500
	drv := openExportTestDB(t, total)
	defer func() { _ = drv.Close() }()

	cursor, err := drv.OpenTableExport(context.Background(), "items", model.TableExportOptions{BatchSize: 1000})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	seen := make(map[int64]struct{}, total)
	batches := 0
	for {
		batch, err := cursor.NextBatch(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		batches++
		for _, row := range batch.Rows {
			id := row["id"].(int64)
			if _, dup := seen[id]; dup {
				t.Fatalf("duplicate id %d", id)
			}
			seen[id] = struct{}{}
		}
		if !batch.HasMore {
			break
		}
	}
	if len(seen) != total {
		t.Fatalf("expected %d unique rows, got %d", total, len(seen))
	}
	if batches != 3 {
		t.Fatalf("expected 3 batches, got %d", batches)
	}
}

func TestExportCursorEmptyTable(t *testing.T) {
	drv := openExportTestDB(t, 0)
	defer func() { _ = drv.Close() }()

	cursor, err := drv.OpenTableExport(context.Background(), "items", model.TableExportOptions{BatchSize: 1000})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	batch, err := cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 0 || batch.HasMore {
		t.Fatalf("expected empty batch, got %+v", batch)
	}
}

func TestExportCursorNoPKStreaming(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rowid.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE logs (msg TEXT); INSERT INTO logs(msg) VALUES ('a'), ('b'), ('c')`); err != nil {
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

	cursor, err := drv.OpenTableExport(context.Background(), "logs", model.TableExportOptions{BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	total := 0
	for {
		batch, err := cursor.NextBatch(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		total += len(batch.Rows)
		for _, row := range batch.Rows {
			if _, hasRowid := row["rowid"]; hasRowid {
				t.Fatal("rowid should not appear in exported rows")
			}
			_ = row
		}
		if !batch.HasMore {
			break
		}
	}
	if total != 3 {
		t.Fatalf("expected 3 rows, got %d", total)
	}
}

func TestExportCursorWithoutRowIDWithPK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wr.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE wr (id INTEGER PRIMARY KEY, msg TEXT) WITHOUT ROWID; INSERT INTO wr VALUES (1,'x'), (2,'y')`); err != nil {
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

	cursor, err := drv.OpenTableExport(context.Background(), "wr", model.TableExportOptions{BatchSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	batch, err := cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(batch.Rows))
	}
}

func TestExportCursorCompositePK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "composite.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		CREATE TABLE pairs (a INTEGER, b INTEGER, v TEXT, PRIMARY KEY (a, b));
		INSERT INTO pairs VALUES (1,1,'x'), (1,2,'y'), (2,1,'z')`); err != nil {
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

	cursor, err := drv.OpenTableExport(context.Background(), "pairs", model.TableExportOptions{BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	var rows []map[string]any
	for {
		batch, err := cursor.NextBatch(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, batch.Rows...)
		if !batch.HasMore {
			break
		}
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
}
