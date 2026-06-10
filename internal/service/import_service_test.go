package service_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
)

func TestImportServiceAppend(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE dest (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}

	csvPath := writeTempCSV(t, "name\nalice\nbob\n")
	res, err := is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		TargetTable:  "dest",
		Mode:         "append",
		ColumnMap:    map[string]string{"name": "name"},
		FilePath:     csvPath,
		Format:       model.DefaultCSVFormat(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsInserted != 2 {
		t.Fatalf("expected 2 inserts, got %d", res.RowsInserted)
	}
}

func TestImportServiceNullValue(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE dest (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}

	csvPath := writeTempCSV(t, "name\n\\N\nok\n")
	fmt := model.DefaultCSVFormat()
	fmt.NullValue = "\\N"
	res, err := is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		TargetTable:  "dest",
		Mode:         "append",
		ColumnMap:    map[string]string{"name": "name"},
		FilePath:     csvPath,
		Format:       fmt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsInserted != 2 {
		t.Fatalf("expected 2 inserts, got %d", res.RowsInserted)
	}

	data, err := qs.BrowseRows(ctx, model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "dest",
		Page:         1,
		PageSize:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Rows[0]["name"] != nil {
		t.Fatalf("expected NULL name, got %+v", data.Rows[0]["name"])
	}
}

func TestImportServiceUpdateRequiresPK(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE nopk (name TEXT); INSERT INTO nopk VALUES ('a')",
	})
	if err != nil {
		t.Fatal(err)
	}

	csvPath := writeTempCSV(t, "name\nb\n")
	_, err = is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		TargetTable:  "nopk",
		Mode:         "update",
		ColumnMap:    map[string]string{"name": "name"},
		FilePath:     csvPath,
		Format:       model.DefaultCSVFormat(),
	})
	if err == nil {
		t.Fatal("expected upsert keys error")
	}
	if !strings.Contains(err.Error(), "upsert keys") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImportServiceCreateTableWithColumnSpecs(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	csvPath := writeTempCSV(t, "id,label\n1,alpha\n2,beta\n")
	res, err := is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		NewTableName: "custom_cols",
		Mode:         "append",
		ColumnMap: map[string]string{
			"id":    "item_id",
			"label": "title",
		},
		NewTableColumns: map[string]model.ImportColumnSpec{
			"id":    {DataType: "INTEGER", PrimaryKey: true},
			"label": {DataType: "TEXT", PrimaryKey: false},
		},
		FilePath: csvPath,
		Format:   model.DefaultCSVFormat(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsInserted != 2 {
		t.Fatalf("expected 2 inserts, got %d", res.RowsInserted)
	}

	schema, err := qs.GetTableSchema(ctx, conn.ID, "custom_cols")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(schema.Columns))
	}
	byName := map[string]model.ColumnInfo{}
	for _, c := range schema.Columns {
		byName[c.Name] = c
	}
	if byName["item_id"].DataType != "INTEGER" || !byName["item_id"].PrimaryKey {
		t.Fatalf("unexpected item_id column: %+v", byName["item_id"])
	}
	if byName["title"].DataType != "TEXT" || byName["title"].PrimaryKey {
		t.Fatalf("unexpected title column: %+v", byName["title"])
	}
}

func TestImportServiceCreateTableChineseName(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	csvPath := writeTempCSV(t, "id,name\n1,alice\n")
	res, err := is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		NewTableName: "发送分",
		Mode:         "append",
		ColumnMap: map[string]string{
			"id":   "id",
			"name": "name",
		},
		FilePath: csvPath,
		Format:   model.DefaultCSVFormat(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsInserted != 1 {
		t.Fatalf("expected 1 insert, got %d", res.RowsInserted)
	}

	schema, err := qs.GetTableSchema(ctx, conn.ID, "发送分")
	if err != nil {
		t.Fatal(err)
	}
	if schema.Name != "发送分" || len(schema.Columns) != 2 {
		t.Fatalf("unexpected schema: %+v", schema)
	}
}

func TestImportServiceLargeCSV(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	const rowCount = 10_000
	var b strings.Builder
	b.WriteString("id,name\n")
	for i := 1; i <= rowCount; i++ {
		fmt.Fprintf(&b, "%d,row%d\n", i, i)
	}
	csvPath := writeTempCSV(t, b.String())

	res, err := is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		NewTableName: "large_import",
		Mode:         "append",
		ColumnMap: map[string]string{
			"id":   "id",
			"name": "name",
		},
		FilePath: csvPath,
		Format:   model.DefaultCSVFormat(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsInserted != rowCount {
		t.Fatalf("expected %d inserts, got %d", rowCount, res.RowsInserted)
	}
}

func TestImportServiceReadOnlyBlocksImport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ro.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	store, err := service.NewConnectionStore(filepath.Join(dir, "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{
		FilePath: path,
		ReadOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	qs := service.NewQueryService(mgr, nil, nil)
	ctx := context.Background()
	is := service.NewImportService(qs)

	csvPath := writeTempCSV(t, "name\nalice\n")
	_, err = is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		TargetTable:  "dest",
		Mode:         "append",
		ColumnMap:    map[string]string{"name": "name"},
		FilePath:     csvPath,
		Format:       model.DefaultCSVFormat(),
	})
	if err == nil {
		t.Fatal("expected read-only error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "READ_ONLY" {
		t.Fatalf("expected READ_ONLY, got %v", err)
	}
}

func TestImportServiceMalformedCSV(t *testing.T) {
	_, qs, _ := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	dir := t.TempDir()
	badPath := filepath.Join(dir, "bad.csv")
	if err := os.WriteFile(badPath, []byte("\"unclosed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := is.ParseCSVPreview(ctx, model.ParseCSVPreviewRequest{
		FilePath: badPath,
		MaxRows:  5,
		Format:   model.DefaultCSVFormat(),
	})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestImportServiceUpdateModeUpsert(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	_, err := qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = qs.Execute(ctx, model.ExecuteQueryRequest{
		ConnectionID: conn.ID,
		SQL:          "INSERT INTO users (id, name) VALUES (1, 'alice')",
	})
	if err != nil {
		t.Fatal(err)
	}

	csvPath := writeTempCSV(t, "id,name\n1,bob\n2,carol\n")
	res, err := is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		TargetTable:  "users",
		Mode:         "update",
		ColumnMap:    map[string]string{"id": "id", "name": "name"},
		FilePath:     csvPath,
		Format:       model.DefaultCSVFormat(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsUpdated != 1 || res.RowsInserted != 1 {
		t.Fatalf("expected 1 update and 1 insert, got %+v", res)
	}

	data, err := qs.BrowseRows(ctx, model.BrowseRowsRequest{
		ConnectionID: conn.ID,
		TableName:    "users",
		Page:         1,
		PageSize:     10,
		Sort:         "id",
		Order:        model.SortAsc,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(data.Rows))
	}
	if data.Rows[0]["name"] != "bob" || data.Rows[1]["name"] != "carol" {
		t.Fatalf("unexpected row data: %+v", data.Rows)
	}
}

func TestImportServiceSanitizedColumnHeaders(t *testing.T) {
	_, qs, conn := newTestEnv(t)
	ctx := context.Background()
	is := service.NewImportService(qs)

	csvPath := writeTempCSV(t, "1-id,发送\n1,alpha\n")
	res, err := is.ImportCSV(ctx, model.ImportCSVRequest{
		ConnectionID: conn.ID,
		NewTableName: "sanitized",
		Mode:         "append",
		FilePath:     csvPath,
		Format:       model.DefaultCSVFormat(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsInserted != 1 {
		t.Fatalf("expected 1 insert, got %d", res.RowsInserted)
	}

	schema, err := qs.GetTableSchema(ctx, conn.ID, "sanitized")
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, col := range schema.Columns {
		names[col.Name] = true
	}
	if !names["col_1_id"] || !names["__"] {
		t.Fatalf("expected sanitized column names, got %+v", names)
	}
}

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
