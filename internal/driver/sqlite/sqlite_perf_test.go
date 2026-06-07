package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/driver/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/testutil"
)

func openSeededDriver(t *testing.T, rowCount int) *sqlite.Driver {
	t.Helper()
	path := testutil.CreateEmptyDB(t)
	testutil.SeedTable(t, path, "large_rows", rowCount)

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type:   model.DriverTypeSQLite,
		SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		t.Fatal(err)
	}
	return drv
}

func TestPerfBrowseTable100kUnder500ms(t *testing.T) {
	if testing.Short() {
		t.Skip("performance test")
	}

	drv := openSeededDriver(t, 100_000)
	defer func() { _ = drv.Close() }()

	start := time.Now()
	data, err := drv.BrowseTable(context.Background(), "large_rows", model.BrowseOptions{
		Page: 1, PageSize: 50, Sort: "id", Order: model.SortAsc,
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 100_000 {
		t.Fatalf("expected 100000 rows, got %d", data.Pagination.TotalRows)
	}
	if len(data.Rows) != 50 {
		t.Fatalf("expected 50 rows on page, got %d", len(data.Rows))
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("BrowseTable took %v, want < 500ms", elapsed)
	}
	t.Logf("BrowseTable 100k page1: %v", elapsed)
}

func TestPerfQueryRows10kUnder200ms(t *testing.T) {
	if testing.Short() {
		t.Skip("performance test")
	}

	drv := openSeededDriver(t, 10_000)
	defer func() { _ = drv.Close() }()

	start := time.Now()
	res, err := drv.QueryRows(context.Background(), "SELECT * FROM large_rows", nil, 10_000)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 10_000 {
		t.Fatalf("expected 10000 rows, got %d", res.RowCount)
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("QueryRows 10k took %v, want < 200ms", elapsed)
	}
	t.Logf("QueryRows 10k: %v", elapsed)
}

func BenchmarkBrowseTable100k(b *testing.B) {
	if testing.Short() {
		b.Skip("performance benchmark")
	}
	path := filepath.Join(b.TempDir(), "bench.db")
	testutil.SeedTable(b, path, "large_rows", 100_000)

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		b.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	opts := model.BrowseOptions{Page: 1, PageSize: 50, Sort: "id", Order: model.SortAsc}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := drv.BrowseTable(context.Background(), "large_rows", opts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueryRows10k(b *testing.B) {
	if testing.Short() {
		b.Skip("performance benchmark")
	}
	path := filepath.Join(b.TempDir(), "bench.db")
	testutil.SeedTable(b, path, "large_rows", 10_000)

	drv := sqlite.New()
	if err := drv.Connect(context.Background(), model.DriverConfig{
		Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path},
	}); err != nil {
		b.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := drv.QueryRows(context.Background(), "SELECT * FROM large_rows", nil, 10_000); err != nil {
			b.Fatal(err)
		}
	}
}
