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

func TestGetTableProfileBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE items (
			id INTEGER PRIMARY KEY,
			category TEXT,
			score INTEGER,
			note TEXT
		);
		INSERT INTO items (category, score, note) VALUES ('a', 10, NULL);
		INSERT INTO items (category, score, note) VALUES ('a', 20, 'x');
		INSERT INTO items (category, score, note) VALUES ('b', 5, 'y');
	`)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	cfg := model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path}}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	profile, err := drv.GetTableProfile(context.Background(), "items")
	if err != nil {
		t.Fatal(err)
	}
	if profile.TotalRows == nil || *profile.TotalRows != 3 {
		t.Fatalf("total rows: %+v", profile.TotalRows)
	}
	if profile.IsSampled {
		t.Fatal("expected no sampling for small table")
	}
	if len(profile.Columns) != 4 {
		t.Fatalf("expected 4 column profiles, got %d", len(profile.Columns))
	}
	var catProfile *model.ColumnProfile
	for i := range profile.Columns {
		if profile.Columns[i].Name == "category" {
			cp := profile.Columns[i]
			catProfile = &cp
			break
		}
	}
	if catProfile == nil {
		t.Fatal("category profile missing")
	}
	if catProfile.DistinctCount == nil || *catProfile.DistinctCount != 2 {
		t.Fatalf("distinct: %+v", catProfile.DistinctCount)
	}
	if !catProfile.IsLowCardinality || len(catProfile.TopValues) == 0 {
		t.Fatalf("expected low cardinality top values: %+v", catProfile)
	}
}

func TestGetTableProfileEmptyTable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE empty_t (id INTEGER PRIMARY KEY);`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	drv := sqlite.New()
	cfg := model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path}}
	if err := drv.Connect(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	profile, err := drv.GetTableProfile(context.Background(), "empty_t")
	if err != nil {
		t.Fatal(err)
	}
	if profile.TotalRows == nil || *profile.TotalRows != 0 {
		t.Fatalf("expected 0 rows, got %+v", profile.TotalRows)
	}
}
