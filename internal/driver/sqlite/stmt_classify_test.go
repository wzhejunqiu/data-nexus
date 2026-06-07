package sqlite_test

import (
	"context"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/driver/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestClassifySQLSelectIsQuery(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	kind, err := drv.ClassifySQL(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatal(err)
	}
	if kind != model.StatementQuery {
		t.Fatalf("got %q want query", kind)
	}
}

func TestClassifySQLWithDeleteIsWrite(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	kind, err := drv.ClassifySQL(context.Background(), "WITH c AS (SELECT 1) DELETE FROM items WHERE 1=0")
	if err != nil {
		t.Fatal(err)
	}
	if kind != model.StatementWrite {
		t.Fatalf("got %q want write", kind)
	}
}

func TestClassifySQLInsertIsWrite(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	kind, err := drv.ClassifySQL(context.Background(), "INSERT INTO items (name) VALUES ('x')")
	if err != nil {
		t.Fatal(err)
	}
	if kind != model.StatementWrite {
		t.Fatalf("got %q want write", kind)
	}
}

func TestClassifySQLWithSelectIsQuery(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	kind, err := drv.ClassifySQL(context.Background(), "WITH c AS (SELECT 1 AS n) SELECT * FROM c")
	if err != nil {
		t.Fatal(err)
	}
	if kind != model.StatementQuery {
		t.Fatalf("got %q want query", kind)
	}
}

func TestClassifySQLInvalidReturnsError(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	_, err := drv.ClassifySQL(context.Background(), "NOT VALID SQL ;;")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "SQL_ERROR" {
		t.Fatalf("expected SQL_ERROR, got %v", err)
	}
}

func TestClassifySQLEmptySQL(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	_, err := drv.ClassifySQL(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}

func TestClassifySQLNotConnected(t *testing.T) {
	drv := sqlite.New()
	_, err := drv.ClassifySQL(context.Background(), "SELECT 1")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "CONNECTION_NOT_FOUND" {
		t.Fatalf("expected CONNECTION_NOT_FOUND, got %v", err)
	}
}

func TestClassifySQLMultiStatementQuery(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	kind, err := drv.ClassifySQL(context.Background(), "SELECT 1; SELECT 2")
	if err != nil {
		t.Fatal(err)
	}
	if kind != model.StatementQuery {
		t.Fatalf("got %q want query", kind)
	}
}

func TestClassifySQLUpdateIsWrite(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	kind, err := drv.ClassifySQL(context.Background(), "UPDATE items SET name='x' WHERE id=1")
	if err != nil {
		t.Fatal(err)
	}
	if kind != model.StatementWrite {
		t.Fatalf("got %q want write", kind)
	}
}

func TestClassifySQLCreateIsWrite(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	kind, err := drv.ClassifySQL(context.Background(), "CREATE TABLE z(id INTEGER)")
	if err != nil {
		t.Fatal(err)
	}
	if kind != model.StatementWrite {
		t.Fatalf("got %q want write", kind)
	}
}

func TestClassifySQLWhitespaceOnly(t *testing.T) {
	drv := openTestDB(t)
	defer func() { _ = drv.Close() }()

	_, err := drv.ClassifySQL(context.Background(), ";;  ")
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}
