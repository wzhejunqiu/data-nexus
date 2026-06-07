package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"github.com/wzhejunqiu/data-nexus/internal/testutil"
	"go.uber.org/zap"
)

func newTestEnv(t *testing.T) (*service.ConnectionManager, *service.QueryService, *model.Connection) {
	t.Helper()
	dir := t.TempDir()
	path := testutil.CreateEmptyDB(t)

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewConnectionManager(store, zap.NewNop())
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	return mgr, service.NewQueryService(mgr), conn
}
