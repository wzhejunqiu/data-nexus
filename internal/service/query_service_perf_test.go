package service_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"github.com/wzhejunqiu/data-nexus/internal/testutil"
)

func openPerfEnv(t *testing.T, rowCount int) (*service.QueryService, *model.Connection) {
	t.Helper()
	dir := t.TempDir()
	path := testutil.CreateEmptyDB(t)
	testutil.SeedTable(t, path, "large_rows", rowCount)

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	return service.NewQueryService(mgr, nil, nil), conn
}

func TestPerfServiceBrowseRows100kUnder500ms(t *testing.T) {
	if testing.Short() {
		t.Skip("performance test")
	}

	qs, conn := openPerfEnv(t, 100_000)
	start := time.Now()
	data, err := qs.BrowseRows(context.Background(), model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "large_rows",
		Page:         1,
		PageSize:     50,
		Sort:         "id",
		Order:        model.SortAsc,
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 100_000 || len(data.Rows) != 50 {
		t.Fatalf("unexpected browse result: total=%d rows=%d", data.Pagination.TotalRows, len(data.Rows))
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("BrowseRows took %v, want < 500ms", elapsed)
	}
	t.Logf("QueryService.BrowseRows 100k page1: %v", elapsed)
}

func TestPerfServiceQuery10kUnder200ms(t *testing.T) {
	if testing.Short() {
		t.Skip("performance test")
	}

	qs, conn := openPerfEnv(t, 10_000)
	start := time.Now()
	res, err := qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "SELECT * FROM large_rows",
		MaxRows:      10_000,
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 10_000 {
		t.Fatalf("expected 10000 rows, got %d", res.RowCount)
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("Execute SELECT 10k took %v, want < 200ms", elapsed)
	}
	t.Logf("QueryService.Execute SELECT 10k: %v", elapsed)
}
