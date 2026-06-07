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

func TestBuildWhereClauseEmpty(t *testing.T) {
	schema := &model.TableSchema{Columns: []model.ColumnInfo{{Name: "id"}}}
	clause, args, err := sqlutil.BuildWhereClause(nil, schema)
	if err != nil || clause != "" || args != nil {
		t.Fatalf("expected empty, got clause=%q args=%v err=%v", clause, args, err)
	}
}

func TestBuildWhereClauseOperators(t *testing.T) {
	schema := &model.TableSchema{Columns: []model.ColumnInfo{{Name: "n"}, {Name: "s"}}}
	v := "x"
	clause, args, err := sqlutil.BuildWhereClause([]model.RowFilter{
		{Column: "n", Operator: model.FilterNe, Value: &v},
		{Column: "s", Operator: model.FilterGte, Value: ptr("1")},
		{Column: "s", Operator: model.FilterLte, Value: ptr("9")},
		{Column: "n", Operator: model.FilterLt, Value: ptr("5")},
		{Column: "n", Operator: model.FilterIsNull},
		{Column: "s", Operator: model.FilterIsNotNull},
		{Column: "n", Operator: model.FilterIn, Values: []string{"a", "b"}},
	}, schema)
	if err != nil {
		t.Fatal(err)
	}
	if clause == "" || len(args) != 6 {
		t.Fatalf("unexpected clause=%q args=%v", clause, args)
	}
}

func TestBuildWhereClauseNilValueErrors(t *testing.T) {
	schema := &model.TableSchema{Columns: []model.ColumnInfo{{Name: "id"}}}
	for _, op := range []model.FilterOperator{
		model.FilterEq, model.FilterNe, model.FilterGt, model.FilterGte,
		model.FilterLt, model.FilterLte, model.FilterLike,
	} {
		_, _, err := sqlutil.BuildWhereClause([]model.RowFilter{
			{Column: "id", Operator: op},
		}, schema)
		if err == nil {
			t.Fatalf("expected error for op %s", op)
		}
	}
}

func TestBuildWhereClauseInEmptyValues(t *testing.T) {
	schema := &model.TableSchema{Columns: []model.ColumnInfo{{Name: "id"}}}
	_, _, err := sqlutil.BuildWhereClause([]model.RowFilter{
		{Column: "id", Operator: model.FilterIn, Values: nil},
	}, schema)
	if err == nil {
		t.Fatal("expected error for empty IN values")
	}
}

func TestBuildWhereClauseUnsupportedOperator(t *testing.T) {
	schema := &model.TableSchema{Columns: []model.ColumnInfo{{Name: "id"}}}
	val := "1"
	_, _, err := sqlutil.BuildWhereClause([]model.RowFilter{
		{Column: "id", Operator: model.FilterOperator("bogus"), Value: &val},
	}, schema)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildWhereClauseLikePreservesWildcard(t *testing.T) {
	schema := &model.TableSchema{Columns: []model.ColumnInfo{{Name: "name"}}}
	val := "%a%"
	_, args, err := sqlutil.BuildWhereClause([]model.RowFilter{
		{Column: "name", Operator: model.FilterLike, Value: &val},
	}, schema)
	if err != nil {
		t.Fatal(err)
	}
	if args[0] != "%a%" {
		t.Fatalf("expected preserved wildcard, got %#v", args[0])
	}
}

func ptr(s string) *string { return &s }
