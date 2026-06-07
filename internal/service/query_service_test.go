package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/executionlog"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

func TestQueryServiceReadOnlyBlocksWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ro.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{
		FilePath: path,
		ReadOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	qs := service.NewQueryService(mgr, nil, nil)
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE t(id INTEGER)",
	})
	if err == nil {
		t.Fatal("expected read-only error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "READ_ONLY" {
		t.Fatalf("expected READ_ONLY, got %v", err)
	}
}

func TestQueryServiceSelectReturnsResult(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}

	qs := service.NewQueryService(mgr, nil, nil)
	res, err := qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "SELECT 1 AS n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != "result" || res.RowCount != 1 {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestQueryServiceDescTableReturnsSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}

	qs := service.NewQueryService(mgr, nil, nil)
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE orders (id INTEGER, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "DESC orders",
	})
	if err != nil {
		t.Fatalf("DESC orders: %v", err)
	}
	if res.Kind != "result" || res.RowCount == 0 {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestQueryServiceWithDeleteUsesExecPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}

	qs := service.NewQueryService(mgr, nil, nil)
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE orders (id INTEGER)",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "WITH c AS (SELECT 1) DELETE FROM orders WHERE id = 1",
	})
	if err != nil {
		t.Fatalf("WITH DELETE: %v", err)
	}
	if res.Kind != "exec" {
		t.Fatalf("expected exec kind, got %+v", res)
	}
}

func TestQueryServiceClassifySQL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}

	qs := service.NewQueryService(mgr, nil, nil)
	kind, err := qs.ClassifySQL(context.Background(), conn.ID, "CREATE TABLE z(id INTEGER)")
	if err != nil {
		t.Fatal(err)
	}
	if kind != model.StatementWrite {
		t.Fatalf("got %q want write", kind)
	}
}

func TestQueryServiceBrowseRows(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO items (name) VALUES ('a'), ('b'), ('c')",
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := qs.BrowseRows(ctx, model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "items",
		Page:         1,
		PageSize:     2,
		Sort:         "id",
		Order:        model.SortAsc,
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 3 || len(data.Rows) != 2 {
		t.Fatalf("unexpected browse result: %+v", data.Pagination)
	}
}

func TestQueryServiceListTables(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}

	list, err := qs.ListTables(ctx, conn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].Name != "users" {
		t.Fatalf("unexpected tables: %+v", list.Items)
	}
}

func TestQueryServiceGetTableSchema(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE users (email TEXT UNIQUE)",
	})
	if err != nil {
		t.Fatal(err)
	}

	schema, err := qs.GetTableSchema(ctx, conn.ID, "users")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Columns) != 1 {
		t.Fatalf("expected 1 column, got %d", len(schema.Columns))
	}
	if len(schema.Indexes) == 0 {
		t.Fatal("expected at least one index")
	}
}

func TestQueryServiceNotConnected(t *testing.T) {
	_, qs, _ := newTestEnv(t)
	ctx := context.Background()

	_, err := qs.BrowseRows(ctx, model.BrowseRowsRequest{
		ConnectionID: "nonexistent",
		TableName:    "t",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_NOT_FOUND" {
		t.Fatalf("expected CONNECTION_NOT_FOUND, got %v", err)
	}
}

func TestQueryServiceBrowseInvalidSort(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE t (id INTEGER PRIMARY KEY)",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = qs.BrowseRows(ctx, model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "t",
		Page:         1,
		PageSize:     50,
		Sort:         `bad"name`,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}

func TestQueryServiceInvalidDescReturnsSQLError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}

	qs := service.NewQueryService(mgr, nil, nil)
	_, err = qs.Execute(context.Background(), model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "desc SELECT * from orders limit 10;",
	})
	if err == nil {
		t.Fatal("expected SQL error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "SQL_ERROR" {
		t.Fatalf("expected SQL_ERROR, got %v", err)
	}
}

func TestQueryServiceExecute_EmptyConnectionID(t *testing.T) {
	_, qs, _ := newTestEnv(t)
	_, err := qs.Execute(context.Background(), model.ExecuteQueryRequest{SQL: "SELECT 1"})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}

func TestQueryServiceExecute_EmptySQL(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	_, err := qs.Execute(context.Background(), model.ExecuteQueryRequest{ConnectionID: conn.ID})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}

func TestQueryServiceExecute_InsertUsesExecPath(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO items (name) VALUES ('alice')",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != "exec" || res.RowsAffected != 1 || res.LastInsertID != 1 {
		t.Fatalf("unexpected exec response: %+v", res)
	}
}

func TestQueryServiceExecute_DefaultMaxRows(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()

	res, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "SELECT 1 AS n",
		MaxRows:      0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != "result" || res.RowCount != 1 {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestQueryServiceClassifySQL_EmptyConnectionID(t *testing.T) {
	_, qs, _ := newTestEnv(t)
	_, err := qs.ClassifySQL(context.Background(), "", "SELECT 1")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}

func TestQueryServiceClassifySQL_EmptySQL(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	_, err := qs.ClassifySQL(context.Background(), conn.ID, "   ")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}

func TestQueryServiceListTables_NotConnected(t *testing.T) {
	_, qs, _ := newTestEnv(t)
	_, err := qs.ListTables(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_NOT_FOUND" {
		t.Fatalf("expected CONNECTION_NOT_FOUND, got %v", err)
	}
}

func TestQueryServiceGetTableSchema_NotConnected(t *testing.T) {
	_, qs, _ := newTestEnv(t)
	_, err := qs.GetTableSchema(context.Background(), "nonexistent", "t")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_NOT_FOUND" {
		t.Fatalf("expected CONNECTION_NOT_FOUND, got %v", err)
	}
}

func TestQueryServiceBrowseRows_DefaultOrderAsc(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE t (id INTEGER PRIMARY KEY, v INTEGER)",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO t (v) VALUES (2), (1)",
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := qs.BrowseRows(ctx, model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "t",
		Page:         1,
		PageSize:     10,
		Sort:         "v",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(data.Rows))
	}
	if data.Rows[0]["v"] != int64(1) {
		t.Fatalf("expected ascending order, got %+v", data.Rows)
	}
}

func TestQueryServiceUpdateCellsBatch(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO users (id, email) VALUES (1, 'a@example.com'), (2, 'b@example.com')",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := qs.UpdateCellsBatch(ctx, model.UpdateCellsBatchRequest{
		ConnectionID: conn.ID,
		TableName:    "users",
		Changes: []model.CellChange{
			{ColumnName: "email", PrimaryKey: map[string]any{"id": int64(1)}, NewValue: "new@example.com"},
			{ColumnName: "email", PrimaryKey: map[string]any{"id": int64(2)}, NewValue: "b2@example.com"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.UpdatedCount != 2 {
		t.Fatalf("expected 2 updates, got %d", res.UpdatedCount)
	}

	data, err := qs.BrowseRows(ctx, model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "users",
		Page:         1,
		PageSize:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Rows[0]["email"] != "new@example.com" {
		t.Fatalf("unexpected row0: %+v", data.Rows[0])
	}
}

func TestQueryServiceUpdateCellsBatchRollback(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT NOT NULL)",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO users (id, email) VALUES (1, 'a@example.com'), (2, 'b@example.com')",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = qs.UpdateCellsBatch(ctx, model.UpdateCellsBatchRequest{
		ConnectionID: conn.ID,
		TableName:    "users",
		Changes: []model.CellChange{
			{ColumnName: "email", PrimaryKey: map[string]any{"id": int64(1)}, NewValue: "ok@example.com"},
			{ColumnName: "email", PrimaryKey: map[string]any{"id": int64(2)}, NewValue: nil},
		},
	})
	if err == nil {
		t.Fatal("expected NOT NULL constraint failure")
	}

	data, err := qs.BrowseRows(ctx, model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "users",
		Page:         1,
		PageSize:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Rows[0]["email"] != "a@example.com" {
		t.Fatalf("expected rollback, row0 email=%v", data.Rows[0]["email"])
	}
}

func TestQueryServiceExecuteRecordsExecution(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}

	execStore, err := executionlog.NewStore(model.ExecutionLogConfig{
		Driver: model.ExecutionLogSQLite,
		SQLite: &model.ExecutionLogSQLiteConfig{FilePath: filepath.Join(dir, "sql-global.db")},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = execStore.Close() }()

	qs := service.NewQueryService(mgr, execStore, zap.NewNop())
	ctx := context.Background()

	_, err = qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "SELECT 1 AS n",
	})
	if err != nil {
		t.Fatal(err)
	}

	items, err := execStore.ListExecutions(ctx, conn.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items", len(items))
	}
	if items[0].SQL != "SELECT 1 AS n" || items[0].Kind != model.SqlExecutionResult || items[0].EffectRows != 1 {
		t.Fatalf("unexpected record: %+v", items[0])
	}

	_, err = qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE t(id INTEGER)",
	})
	if err != nil {
		t.Fatal(err)
	}

	items, err = execStore.ListExecutions(ctx, conn.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items", len(items))
	}
	var hasResult, hasExec bool
	for _, item := range items {
		switch item.Kind {
		case model.SqlExecutionResult:
			hasResult = true
		case model.SqlExecutionExec:
			hasExec = true
		}
	}
	if !hasResult || !hasExec {
		t.Fatalf("expected result and exec records, got %+v", items)
	}

	_, err = qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "NOT VALID SQL",
	})
	if err == nil {
		t.Fatal("expected SQL error")
	}

	items, err = execStore.ListExecutions(ctx, conn.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("failed execution should not insert, got %d items", len(items))
	}
}
