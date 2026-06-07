package export_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestResolveStableRowKeyPrimaryIndex(t *testing.T) {
	schema := &model.TableSchema{
		Name: "users",
		Indexes: []model.IndexInfo{
			{Name: "sqlite_autoindex_users_1", Columns: []string{"id"}, Primary: true},
		},
		Columns: []model.ColumnInfo{
			{Name: "id", PrimaryKey: true, Position: 1},
			{Name: "name", Position: 2},
		},
	}
	key, err := export.ResolveStableRowKey(schema, model.DriverTypeSQLite, false)
	if err != nil {
		t.Fatal(err)
	}
	if key.Source != "primary_key" || len(key.Columns) != 1 || key.Columns[0] != "id" {
		t.Fatalf("unexpected key: %+v", key)
	}
}

func TestResolveStableRowKeyCompositePK(t *testing.T) {
	schema := &model.TableSchema{
		Name: "pairs",
		Columns: []model.ColumnInfo{
			{Name: "b", PrimaryKey: true, Position: 2},
			{Name: "a", PrimaryKey: true, Position: 1},
		},
	}
	key, err := export.ResolveStableRowKey(schema, model.DriverTypeSQLite, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(key.Columns) != 2 || key.Columns[0] != "a" || key.Columns[1] != "b" {
		t.Fatalf("expected composite PK order, got %+v", key.Columns)
	}
}

func TestResolveStableRowKeySQLiteRowid(t *testing.T) {
	schema := &model.TableSchema{
		Name: "no_pk",
		Columns: []model.ColumnInfo{
			{Name: "name", Position: 1},
		},
	}
	key, err := export.ResolveStableRowKey(schema, model.DriverTypeSQLite, false)
	if err != nil {
		t.Fatal(err)
	}
	if key.Source != "rowid" || key.Columns[0] != "rowid" {
		t.Fatalf("expected rowid fallback, got %+v", key)
	}
}

func TestResolveStableRowKeyWithoutRowID(t *testing.T) {
	schema := &model.TableSchema{
		Name: "wr",
		Columns: []model.ColumnInfo{
			{Name: "name", Position: 1},
		},
	}
	_, err := export.ResolveStableRowKey(schema, model.DriverTypeSQLite, true)
	if err == nil {
		t.Fatal("expected error for WITHOUT ROWID table without PK")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "EXPORT_NO_STABLE_KEY" {
		t.Fatalf("expected EXPORT_NO_STABLE_KEY, got %v", err)
	}
}

func TestResolveStableRowKeyNonSQLiteRequiresPK(t *testing.T) {
	schema := &model.TableSchema{
		Name: "no_pk",
		Columns: []model.ColumnInfo{
			{Name: "name", Position: 1},
		},
	}
	_, err := export.ResolveStableRowKey(schema, model.DriverType("postgres"), false)
	if err == nil {
		t.Fatal("expected error without PK on non-sqlite dialect")
	}
}
