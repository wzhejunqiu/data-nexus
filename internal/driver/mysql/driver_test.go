package mysql

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
	return &Driver{db: db, database: "testdb"}, mock
}

func expectTableType(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT TABLE_TYPE FROM information_schema.TABLES`).
		WithArgs("testdb", "users").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_TYPE"}).AddRow("BASE TABLE"))
}

func expectSchemaColumns(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT COLUMN_NAME, DATA_TYPE`).
		WithArgs("testdb", "users").
		WillReturnRows(sqlmock.NewRows([]string{
			"COLUMN_NAME", "DATA_TYPE", "COLUMN_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT", "ORDINAL_POSITION",
		}).AddRow("id", "int", "int(11)", "NO", "PRI", nil, 1).
			AddRow("name", "varchar", "varchar(64)", "YES", "", nil, 2))
}

func expectSchemaIndexes(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT INDEX_NAME, NON_UNIQUE`).
		WithArgs("testdb", "users").
		WillReturnRows(sqlmock.NewRows([]string{"INDEX_NAME", "NON_UNIQUE", "SEQ_IN_INDEX", "COLUMN_NAME"}).
			AddRow("PRIMARY", 0, 1, "id"))
}

func expectSchemaWithLookups(mock sqlmock.Sqlmock, lookups int) {
	for range lookups {
		expectTableType(mock)
	}
	expectSchemaColumns(mock)
	expectSchemaIndexes(mock)
}

func TestConnectValidation(t *testing.T) {
	drv := New()
	if err := drv.Connect(context.Background(), model.DriverConfig{Type: model.DriverTypeMySQL}); err == nil {
		t.Fatal("expected error without mysql config")
	}
}

func TestPingNotConnected(t *testing.T) {
	drv := New()
	if err := drv.Ping(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestCloseNilDB(t *testing.T) {
	if err := New().Close(); err != nil {
		t.Fatal(err)
	}
}

func TestListNamespaces(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectQuery("SHOW DATABASES").
		WillReturnRows(sqlmock.NewRows([]string{"Database"}).
			AddRow("testdb").
			AddRow("otherdb"))

	namespaces, err := drv.ListNamespaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(namespaces) != 2 || namespaces[0].Name != "testdb" || namespaces[0].Kind != model.NamespaceKindDatabase {
		t.Fatalf("unexpected namespaces: %+v", namespaces)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListTables(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectQuery(`SELECT TABLE_NAME, TABLE_TYPE`).
		WithArgs("testdb").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE"}).
			AddRow("users", "BASE TABLE").
			AddRow("active_v", "VIEW"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM `testdb`.`users`")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	tables, err := drv.ListTables(context.Background(), model.ListTablesOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 2 || tables[0].Name != "users" || tables[0].RowCount == nil || *tables[0].RowCount != 3 {
		t.Fatalf("unexpected tables: %+v", tables)
	}
	if tables[1].Type != model.TableTypeView {
		t.Fatalf("expected view, got %s", tables[1].Type)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetTableSchema(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectTableType(mock)
	expectSchemaColumns(mock)
	expectSchemaIndexes(mock)

	schema, err := drv.GetTableSchema(context.Background(), "users")
	if err != nil {
		t.Fatal(err)
	}
	if schema.Name != "users" || len(schema.Columns) != 2 || !schema.Columns[0].PrimaryKey {
		t.Fatalf("unexpected schema: %+v", schema)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetTableSchemaInvalidName(t *testing.T) {
	drv, mock := newMockDriver(t)
	_, err := drv.GetTableSchema(context.Background(), `bad"name`)
	if err == nil {
		t.Fatal("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBrowseTable(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectSchemaWithLookups(mock, 2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM `users`")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM`).
		WithArgs(50, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "alice"))

	data, err := drv.BrowseTable(context.Background(), "users", model.BrowseOptions{Page: 1, PageSize: 50})
	if err != nil {
		t.Fatal(err)
	}
	if data.Pagination.TotalRows != 1 || len(data.Rows) != 1 {
		t.Fatalf("unexpected data: %+v", data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestQueryRowsAndClassifySQL(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectQuery(`SELECT 1`).
		WillReturnRows(sqlmock.NewRows([]string{"n"}).AddRow(1))

	res, err := drv.QueryRows(context.Background(), "SELECT 1", nil, 10)
	if err != nil || res.RowCount != 1 {
		t.Fatalf("query: %+v err=%v", res, err)
	}
	kind, err := drv.ClassifySQL(context.Background(), "INSERT INTO t VALUES (1)")
	if err != nil || kind != model.StatementWrite {
		t.Fatalf("classify: %s %v", kind, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecReadOnly(t *testing.T) {
	drv, mock := newMockDriver(t)
	drv.readOnly = true
	_, err := drv.Exec(context.Background(), "DELETE FROM users", nil)
	if err == nil {
		t.Fatal("expected read-only error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecSuccess(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectExec(`UPDATE users SET name = \?`).
		WithArgs("bob", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	res, err := drv.Exec(context.Background(), "UPDATE users SET name = ?", []any{"bob", 1})
	if err != nil || res.RowsAffected != 1 {
		t.Fatalf("exec: %+v err=%v", res, err)
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
			{Name: "id", PrimaryKey: true, DataType: "INT"},
			{Name: "name", DataType: "VARCHAR"},
		},
	}
	updated, err := drv.UpdateCells(context.Background(), "users", []model.CellChange{
		{ColumnName: "name", PrimaryKey: map[string]any{"id": int64(1)}, NewValue: "new"},
	}, schema)
	if err != nil || updated != 1 {
		t.Fatalf("updated=%d err=%v", updated, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateCellsNoPrimaryKey(t *testing.T) {
	drv, mock := newMockDriver(t)
	schema := &model.TableSchema{
		Name:    "users",
		Type:    model.TableTypeTable,
		Columns: []model.ColumnInfo{{Name: "name"}},
	}
	_, err := drv.UpdateCells(context.Background(), "users", nil, schema)
	if err == nil {
		t.Fatal("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOpenTableExport(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectQuery(`SELECT TABLE_TYPE FROM information_schema.TABLES`).
		WithArgs("testdb", "users").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_TYPE"}).AddRow("BASE TABLE"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` LIMIT 0")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "a"))

	cursor, err := drv.OpenTableExport(context.Background(), "users", model.TableExportOptions{BatchSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()
	batch, err := cursor.NextBatch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(batch.Rows))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetTableProfileEmpty(t *testing.T) {
	drv, mock := newMockDriver(t)
	expectSchemaWithLookups(mock, 1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM `users`")).
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

func TestUnsupportedSQLiteFeatures(t *testing.T) {
	drv, _ := newMockDriver(t)
	ctx := context.Background()
	for _, err := range []error{
		drv.Attach(ctx, "", ""),
		drv.Detach(ctx, "x"),
	} {
		if err == nil {
			t.Fatal("expected unsupported error")
		}
	}
	fts, err := drv.DetectFTSTable(ctx, "t")
	if err != nil || fts == nil || fts.Enabled {
		t.Fatalf("expected disabled FTS, got %+v err=%v", fts, err)
	}
}

func TestTypeAndReadOnly(t *testing.T) {
	drv := &Driver{readOnly: true}
	if drv.Type() != model.DriverTypeMySQL || !drv.ReadOnly() {
		t.Fatal("unexpected type/readonly")
	}
}

// Ensure sql.ErrNoRows path for missing table.
func TestLookupTableTypeNotFound(t *testing.T) {
	drv, mock := newMockDriver(t)
	mock.ExpectQuery(`SELECT TABLE_TYPE FROM information_schema.TABLES`).
		WithArgs("testdb", "missing").
		WillReturnError(sql.ErrNoRows)
	_, err := drv.GetTableSchema(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected not found")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
