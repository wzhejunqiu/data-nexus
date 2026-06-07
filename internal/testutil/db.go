package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// CreateEmptyDB returns the path to a new empty SQLite database file.
func CreateEmptyDB(tb testing.TB) string {
	tb.Helper()
	dir := tb.TempDir()
	path := filepath.Join(dir, "test.db")
	f, err := os.Create(path)
	if err != nil {
		tb.Fatal(err)
	}
	if err := f.Close(); err != nil {
		tb.Fatal(err)
	}
	return path
}

// SeedTable creates table with (id INTEGER PRIMARY KEY, payload TEXT) and inserts rowCount rows.
func SeedTable(tb testing.TB, path, table string, rowCount int) {
	tb.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		tb.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if _, err := db.Exec(fmt.Sprintf(
		`CREATE TABLE %q (id INTEGER PRIMARY KEY, payload TEXT)`, table,
	)); err != nil {
		tb.Fatal(err)
	}

	const batchSize = 1000
	if _, err := db.Exec("BEGIN"); err != nil {
		tb.Fatal(err)
	}
	stmt, err := db.Prepare(fmt.Sprintf(`INSERT INTO %q (payload) VALUES (?)`, table))
	if err != nil {
		tb.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	for i := 0; i < rowCount; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("row-%d", i)); err != nil {
			tb.Fatal(err)
		}
		if (i+1)%batchSize == 0 {
			if _, err := db.Exec("COMMIT"); err != nil {
				tb.Fatal(err)
			}
			if _, err := db.Exec("BEGIN"); err != nil {
				tb.Fatal(err)
			}
		}
	}
	if _, err := db.Exec("COMMIT"); err != nil {
		tb.Fatal(err)
	}
}
