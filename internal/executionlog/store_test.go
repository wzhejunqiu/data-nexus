package executionlog_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/executionlog"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestNewStoreDefaultSQLite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	store, err := executionlog.NewStore(config.DefaultExecutionLogConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	if store.Type() != model.ExecutionLogSQLite {
		t.Fatalf("got %s", store.Type())
	}
}

func TestNewStoreUnsupportedDriver(t *testing.T) {
	_, err := executionlog.NewStore(model.ExecutionLogConfig{Driver: "postgres"})
	if err == nil {
		t.Fatal("expected error")
	}
}
