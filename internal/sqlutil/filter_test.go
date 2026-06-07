package sqlutil_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
)

func TestBuildWhereClause(t *testing.T) {
	schema := &model.TableSchema{
		Columns: []model.ColumnInfo{
			{Name: "id"},
			{Name: "name"},
			{Name: "score"},
		},
	}
	val := "alice"
	clause, args, err := sqlutil.BuildWhereClause([]model.RowFilter{
		{Column: "name", Operator: model.FilterEq, Value: &val},
		{Column: "score", Operator: model.FilterGt, Value: ptr("10")},
	}, schema)
	if err != nil {
		t.Fatal(err)
	}
	if clause != ` WHERE "name" = ? AND "score" > ?` {
		t.Fatalf("unexpected clause: %s", clause)
	}
	if len(args) != 2 || args[0] != "alice" || args[1] != "10" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildWhereClauseLikeWraps(t *testing.T) {
	schema := &model.TableSchema{Columns: []model.ColumnInfo{{Name: "name"}}}
	val := "bob"
	clause, args, err := sqlutil.BuildWhereClause([]model.RowFilter{
		{Column: "name", Operator: model.FilterLike, Value: &val},
	}, schema)
	if err != nil {
		t.Fatal(err)
	}
	if clause != ` WHERE "name" LIKE ?` {
		t.Fatalf("unexpected clause: %s", clause)
	}
	if args[0] != "%bob%" {
		t.Fatalf("expected wrapped like, got %#v", args)
	}
}

func TestBuildWhereClauseRejectsInvalidColumn(t *testing.T) {
	schema := &model.TableSchema{Columns: []model.ColumnInfo{{Name: "id"}}}
	val := "1"
	_, _, err := sqlutil.BuildWhereClause([]model.RowFilter{
		{Column: "evil;drop", Operator: model.FilterEq, Value: &val},
	}, schema)
	if err == nil {
		t.Fatal("expected error for invalid column")
	}
}

func ptr(s string) *string { return &s }
