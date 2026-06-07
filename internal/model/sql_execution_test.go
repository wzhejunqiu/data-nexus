package model_test

import (
	"encoding/json"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestSqlExecutionKindFromResponse(t *testing.T) {
	if model.SqlExecutionKindFromResponse("exec") != model.SqlExecutionExec {
		t.Fatal("expected exec")
	}
	if model.SqlExecutionKindFromResponse("result") != model.SqlExecutionResult {
		t.Fatal("expected result")
	}
}

func TestSqlExecutionKindJSON(t *testing.T) {
	data, err := json.Marshal(model.SqlExecutionExec)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"exec"` {
		t.Fatalf("got %s", data)
	}

	var k model.SqlExecutionKind
	if err := json.Unmarshal([]byte(`"result"`), &k); err != nil {
		t.Fatal(err)
	}
	if k != model.SqlExecutionResult {
		t.Fatalf("got %d", k)
	}
}

func TestNewSqlExecutionRecord(t *testing.T) {
	rec := model.NewSqlExecutionRecord("c1", "SELECT 1", &model.QueryResponse{
		Kind:       "result",
		RowCount:   3,
		DurationMs: 10,
	})
	if rec.ConnectionID != "c1" || rec.SQL != "SELECT 1" {
		t.Fatalf("unexpected record: %+v", rec)
	}
	if rec.Kind != model.SqlExecutionResult || rec.EffectRows != 3 || rec.DurationMs != 10 {
		t.Fatalf("unexpected metrics: %+v", rec)
	}

	rec = model.NewSqlExecutionRecord("c1", "DELETE FROM t", &model.QueryResponse{
		Kind:         "exec",
		RowsAffected: 5,
		DurationMs:   2,
	})
	if rec.Kind != model.SqlExecutionExec || rec.EffectRows != 5 {
		t.Fatalf("unexpected exec record: %+v", rec)
	}
}
