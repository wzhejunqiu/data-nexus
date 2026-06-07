package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"github.com/wzhejunqiu/data-nexus/internal/testutil"
)

func newTestEnv(t *testing.T) (*service.ConnectionManager, *service.QueryService, *model.Connection) {
	t.Helper()
	dir := t.TempDir()
	path := testutil.CreateEmptyDB(t)

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManager(store)
	conn, err := mgr.OpenConnectionFromFile(context.Background(), model.ConnectRequest{FilePath: path})
	if err != nil {
		t.Fatal(err)
	}
	return mgr, service.NewQueryService(mgr, nil, nil), conn
}
