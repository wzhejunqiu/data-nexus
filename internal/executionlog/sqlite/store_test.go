package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/executionlog/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func newTestStore(t *testing.T) *sqlite.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sql-global.db")
	store, err := sqlite.NewStore(&model.ExecutionLogSQLiteConfig{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestStoreInsertAndListExecutions(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.Insert(ctx, model.SqlExecutionRecord{
		ConnectionID: "c1",
		SQL:          "SELECT 1",
		Kind:         model.SqlExecutionResult,
		EffectRows:   1,
		DurationMs:   10,
		ExecutedAt:   time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Insert(ctx, model.SqlExecutionRecord{
		ConnectionID: "c1",
		SQL:          "DELETE FROM t",
		Kind:         model.SqlExecutionExec,
		EffectRows:   2,
		DurationMs:   5,
		ExecutedAt:   time.Date(2026, 6, 7, 13, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	items, err := store.ListExecutions(ctx, "c1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items", len(items))
	}
	if items[0].SQL != "DELETE FROM t" || items[0].Kind != model.SqlExecutionExec {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[1].SQL != "SELECT 1" || items[1].Kind != model.SqlExecutionResult {
		t.Fatalf("unexpected second item: %+v", items[1])
	}
}

func TestStoreListQueryHistoryDedup(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for _, at := range []time.Time{
		time.Date(2026, 6, 7, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 7, 11, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
	} {
		if err := store.Insert(ctx, model.SqlExecutionRecord{
			ConnectionID: "c1",
			SQL:          "SELECT 1",
			Kind:         model.SqlExecutionResult,
			EffectRows:   1,
			DurationMs:   1,
			ExecutedAt:   at,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Insert(ctx, model.SqlExecutionRecord{
		ConnectionID: "c1",
		SQL:          "SELECT 2",
		Kind:         model.SqlExecutionResult,
		EffectRows:   1,
		DurationMs:   1,
		ExecutedAt:   time.Date(2026, 6, 7, 9, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	history, err := store.ListQueryHistory(ctx, "c1", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 {
		t.Fatalf("got %v", history)
	}
	if history[0] != "SELECT 1" || history[1] != "SELECT 2" {
		t.Fatalf("unexpected order: %v", history)
	}
}

func TestStoreListRequiresConnectionID(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if _, err := store.ListQueryHistory(ctx, "", 50); err == nil {
		t.Fatal("expected error")
	}
	if _, err := store.ListExecutions(ctx, "", 50); err == nil {
		t.Fatal("expected error")
	}
}

func TestStoreListAllExecutions(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_ = store.Insert(ctx, model.SqlExecutionRecord{
		ConnectionID: "c1", SQL: "SELECT 1", Kind: model.SqlExecutionResult,
		EffectRows: 1, DurationMs: 1, ExecutedAt: time.Date(2026, 6, 7, 10, 0, 0, 0, time.UTC),
	})
	_ = store.Insert(ctx, model.SqlExecutionRecord{
		ConnectionID: "c2", SQL: "SELECT 2", Kind: model.SqlExecutionResult,
		EffectRows: 1, DurationMs: 2, ExecutedAt: time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
	})

	all, err := store.ListAllExecutions(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2, got %d", len(all))
	}
	if all[0].ConnectionID != "c2" {
		t.Fatalf("expected newest first, got %s", all[0].ConnectionID)
	}

	capped, err := store.ListAllExecutions(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(capped) != 1 {
		t.Fatalf("expected limit 1, got %d", len(capped))
	}
}
