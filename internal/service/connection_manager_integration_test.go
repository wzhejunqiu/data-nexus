package service_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/secrets"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"github.com/wzhejunqiu/data-nexus/internal/testutil/mysqlserver"
	"github.com/wzhejunqiu/data-nexus/internal/testutil/pgserver"
)

var (
	testPostgres *pgserver.Server
	testMySQL    *mysqlserver.Server
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Short() {
		pg, err := pgserver.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "service integration: postgres harness: %v\n", err)
			os.Exit(1)
		}
		testPostgres = pg

		my, err := mysqlserver.Start()
		if err != nil {
			_ = testPostgres.Close()
			fmt.Fprintf(os.Stderr, "service integration: mysql harness: %v\n", err)
			os.Exit(1)
		}
		testMySQL = my
	}
	code := m.Run()
	if testMySQL != nil {
		_ = testMySQL.Close()
	}
	if testPostgres != nil {
		_ = testPostgres.Close()
	}
	os.Exit(code)
}

func TestIntegrationThreeDriversConcurrentOpen(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in -short mode")
	}
	if testPostgres == nil || testMySQL == nil {
		t.Fatal("integration harness not started")
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "local.db")
	f, err := os.Create(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManagerWithSecrets(store, secrets.NewMockStore())
	ctx := context.Background()

	sqliteConn, err := mgr.OpenConnectionFromFile(ctx, model.ConnectRequest{FilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}

	pgSaved, err := mgr.CreateRemoteConnection(ctx, model.RemoteConnectRequest{
		Type:     model.DriverTypePostgres,
		Name:     "integration-pg",
		Password: testPostgres.Password,
		Open:     true,
		Postgres: &model.PostgresConfig{
			Host:     testPostgres.Host,
			Port:     testPostgres.Port,
			Database: testPostgres.Database,
			User:     testPostgres.User,
			Schema:   "public",
			SSLMode:  "disable",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	mysqlSaved, err := mgr.CreateRemoteConnection(ctx, model.RemoteConnectRequest{
		Type:     model.DriverTypeMySQL,
		Name:     "integration-mysql",
		Password: "",
		Open:     true,
		MySQL: &model.MySQLConfig{
			Host:     testMySQL.Host,
			Port:     testMySQL.Port,
			Database: testMySQL.Database,
			User:     "root",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	list := mgr.ListConnections()
	openByType := map[model.DriverType]int{}
	for _, item := range list.Items {
		if item.Status != model.ConnectionStatusOpen {
			continue
		}
		openByType[item.Type]++
	}
	if openByType[model.DriverTypeSQLite] != 1 ||
		openByType[model.DriverTypePostgres] != 1 ||
		openByType[model.DriverTypeMySQL] != 1 {
		t.Fatalf("expected one open connection per driver type, got %#v", openByType)
	}

	for _, id := range []string{sqliteConn.ID, pgSaved.ID, mysqlSaved.ID} {
		drv, err := mgr.Driver(id)
		if err != nil {
			t.Fatalf("driver for %s: %v", id, err)
		}
		if _, err := drv.ListTables(ctx); err != nil {
			t.Fatalf("ListTables for %s: %v", id, err)
		}
	}

	mgr.CloseAll()
}
