package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func newMockDriver(t *testing.T) (*Driver, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(
		sqlmock.MonitorPingsOption(true),
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &Driver{db: db, schema: "public"}, mock
}

func expectPGTableType(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT table_type`).
		WithArgs("public", "users").
		WillReturnRows(sqlmock.NewRows([]string{"table_type"}).AddRow("BASE TABLE"))
}

func expectPGSchemaColumns(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT column_name, data_type`).
		WithArgs("public", "users").
		WillReturnRows(sqlmock.NewRows([]string{
			"column_name", "data_type", "udt_name", "is_nullable", "column_default", "ordinal_position",
		}).AddRow("id", "integer", "int4", "NO", nil, 1).
			AddRow("name", "text", "text", "YES", nil, 2))
}

func expectPGPrimaryKey(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT kcu.column_name`).
		WithArgs("public", "users").
		WillReturnRows(sqlmock.NewRows([]string{"column_name"}).AddRow("id"))
}

func expectPGIndexes(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT\s+irel.relname`).
		WithArgs("public", "users").
		WillReturnRows(sqlmock.NewRows([]string{"index_name", "indisunique", "indisprimary", "column_name", "col_position"}).
			AddRow("users_pkey", true, true, "id", 1))
}

func expectPGSchemaWithLookups(mock sqlmock.Sqlmock, lookups int) {
	for range lookups {
		expectPGTableType(mock)
	}
	expectPGSchemaColumns(mock)
	expectPGPrimaryKey(mock)
	expectPGIndexes(mock)
}

func TestConnectValidation(t *testing.T) {
	drv := New()
	if err := drv.Connect(context.Background(), model.DriverConfig{Type: model.DriverTypePostgres}); err == nil {
		t.Fatal("expected error")
	}
}

func TestClassifySQLNotConnected(t *testing.T) {
	_, err := New().ClassifySQL(context.Background(), "SELECT 1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListTables(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectQuery(`SELECT table_name, table_type`).
		WithArgs("public").
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "table_type"}).
			AddRow("users", "BASE TABLE"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "public"."users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	tables, err := drv.ListTables(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 1 || tables[0].RowCount == nil || *tables[0].RowCount != 2 {
		t.Fatalf("unexpected: %+v", tables)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetTableSchema(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectPGTableType(mock)
	expectPGSchemaColumns(mock)
	expectPGPrimaryKey(mock)
	expectPGIndexes(mock)

	schema, err := drv.GetTableSchema(context.Background(), "users")
	if err != nil {
		t.Fatal(err)
	}
	if schema.Name != "users" || !schema.Columns[0].PrimaryKey {
		t.Fatalf("unexpected schema: %+v", schema)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBrowseTable(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectPGSchemaWithLookups(mock, 2)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "public"."users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM`).
		WithArgs(50, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "alice"))

	data, err := drv.BrowseTable(context.Background(), "users", model.BrowseOptions{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 1 {
		t.Fatalf("unexpected: %+v", data.Pagination)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestQueryRowsAndExec(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectQuery(`SELECT 1`).
		WillReturnRows(sqlmock.NewRows([]string{"n"}).AddRow(1))
	mock.ExpectExec(`DELETE FROM users`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	res, err := drv.QueryRows(context.Background(), "SELECT 1", nil, 10)
	if err != nil || res.RowCount != 1 {
		t.Fatal(err)
	}
	execRes, err := drv.Exec(context.Background(), "DELETE FROM users", nil)
	if err != nil || execRes.RowsAffected != 1 {
		t.Fatalf("exec: %+v err=%v", execRes, err)
	}
	kind, err := drv.ClassifySQL(context.Background(), "SELECT 1")
	if err != nil || kind != model.StatementQuery {
		t.Fatalf("classify: %s %v", kind, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateCells(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	schema := &model.TableSchema{
		Name: "users",
		Type: model.TableTypeTable,
		Columns: []model.ColumnInfo{
			{Name: "id", PrimaryKey: true, DataType: "INTEGER"},
			{Name: "name", DataType: "TEXT"},
		},
	}
	updated, err := drv.UpdateCells(context.Background(), "users", []model.CellChange{
		{ColumnName: "name", PrimaryKey: map[string]any{"id": int64(1)}, NewValue: "x"},
	}, schema)
	if err != nil || updated != 1 {
		t.Fatalf("updated=%d err=%v", updated, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOpenTableExport(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectQuery(`SELECT table_type`).
		WithArgs("public", "users").
		WillReturnRows(sqlmock.NewRows([]string{"table_type"}).AddRow("BASE TABLE"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "public"."users" LIMIT 0`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectExec(`BEGIN READ ONLY`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "public"."users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "a"))

	cursor, err := drv.OpenTableExport(context.Background(), "users", model.TableExportOptions{BatchSize: 5})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := cursor.NextBatch(context.Background())
	if err != nil || len(batch.Rows) != 1 {
		t.Fatalf("batch: %+v err=%v", batch, err)
	}
	mock.ExpectExec(`ROLLBACK`).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := cursor.Close(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetTableProfileEmpty(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectPGSchemaWithLookups(mock, 1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "public"."users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	profile, err := drv.GetTableProfile(context.Background(), "users")
	if err != nil {
		t.Fatal(err)
	}
	if profile.TotalRows == nil || *profile.TotalRows != 0 {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestParseTableRefInvalid(t *testing.T) {
	drv, mock := newMockDriver(t)
	_, err := drv.GetTableSchema(context.Background(), `bad"name`)
	if err == nil {
		t.Fatal("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTableNotFound(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectQuery(`SELECT table_type`).
		WithArgs("public", "missing").
		WillReturnError(sql.ErrNoRows)
	_, err := drv.GetTableSchema(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
