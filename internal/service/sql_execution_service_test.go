package service_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/executionlog/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
)

func TestSqlExecutionServiceWithStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sql-global.db")
	store, err := sqlite.NewStore(&model.ExecutionLogSQLiteConfig{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	svc := service.NewSqlExecutionService(store)
	ctx := context.Background()

	if err := store.Insert(ctx, model.SqlExecutionRecord{
		ConnectionID: "conn-1",
		SQL:          "SELECT 1",
		Kind:         model.SqlExecutionResult,
		EffectRows:   1,
		DurationMs:   5,
		ExecutedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Insert(ctx, model.SqlExecutionRecord{
		ConnectionID: "conn-1",
		SQL:          "SELECT 2",
		Kind:         model.SqlExecutionResult,
		EffectRows:   1,
		DurationMs:   3,
		ExecutedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}

	history, err := svc.ListQueryHistory(ctx, "conn-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) == 0 {
		t.Fatal("expected query history entries")
	}

	executions, err := svc.ListSqlExecutions(ctx, "conn-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(executions.Items) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(executions.Items))
	}
}

func TestSqlExecutionServiceNilStore(t *testing.T) {
	svc := service.NewSqlExecutionService(nil)
	ctx := context.Background()

	history, err := svc.ListQueryHistory(ctx, "conn-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 0 {
		t.Fatalf("expected empty history, got %+v", history)
	}

	executions, err := svc.ListSqlExecutions(ctx, "conn-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(executions.Items) != 0 {
		t.Fatalf("expected empty executions, got %+v", executions.Items)
	}
}
