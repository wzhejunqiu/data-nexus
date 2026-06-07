package postgres_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/driver/postgres"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/testutil/pgserver"
)

var testPostgres *pgserver.Server

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Short() {
		srv, err := pgserver.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "postgres integration harness: %v\n", err)
			os.Exit(1)
		}
		testPostgres = srv
	}
	code := m.Run()
	if testPostgres != nil {
		_ = testPostgres.Close()
	}
	os.Exit(code)
}

func testPostgresConfig(t *testing.T) model.DriverConfig {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test skipped in -short mode")
	}
	if testPostgres == nil {
		t.Fatal("postgres integration harness not started")
	}
	return model.DriverConfig{
		Type: model.DriverTypePostgres,
		Postgres: &model.PostgresConfig{
			Host:     testPostgres.Host,
			Port:     testPostgres.Port,
			Database: testPostgres.Database,
			User:     testPostgres.User,
			Password: testPostgres.Password,
			SSLMode:  "disable",
			Schema:   "public",
			ReadOnly: false,
		},
	}
}

func TestIntegrationPostgresConnectListTables(t *testing.T) {
	cfg := testPostgresConfig(t)
	ctx := context.Background()

	drv := postgres.New()
	if err := drv.Connect(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	if err := drv.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	_, err := drv.Exec(ctx, `CREATE TABLE IF NOT EXISTS integration_items (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL
	)`, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = drv.Exec(ctx, `INSERT INTO integration_items(name) SELECT 'alpha' WHERE NOT EXISTS (SELECT 1 FROM integration_items WHERE name = 'alpha')`, nil)
	if err != nil {
		t.Fatal(err)
	}

	tables, err := drv.ListTables(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tbl := range tables {
		if tbl.Name == "integration_items" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected integration_items in %v", tables)
	}

	schema, err := drv.GetTableSchema(ctx, "integration_items")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Columns) < 2 {
		t.Fatalf("expected columns, got %#v", schema.Columns)
	}
}

func TestIntegrationPostgresExportCursor(t *testing.T) {
	cfg := testPostgresConfig(t)
	ctx := context.Background()

	drv := postgres.New()
	if err := drv.Connect(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	_, err := drv.Exec(ctx, `CREATE TABLE IF NOT EXISTS integration_export (
		id SERIAL PRIMARY KEY,
		label TEXT NOT NULL
	)`, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = drv.Exec(ctx, `DELETE FROM integration_export`, nil)
	for i := 1; i <= 5; i++ {
		if _, err := drv.Exec(ctx, `INSERT INTO integration_export(label) VALUES ($1)`, []any{fmt.Sprintf("%d", i)}); err != nil {
			t.Fatal(err)
		}
	}

	cursor, err := drv.OpenTableExport(ctx, "integration_export", model.TableExportOptions{BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	total := 0
	for {
		batch, err := cursor.NextBatch(ctx)
		if err != nil {
			t.Fatal(err)
		}
		total += len(batch.Rows)
		if !batch.HasMore {
			break
		}
	}
	if total != 5 {
		t.Fatalf("expected 5 rows exported, got %d", total)
	}
}

func TestIntegrationPostgresExportCompositePK(t *testing.T) {
	cfg := testPostgresConfig(t)
	ctx := context.Background()

	drv := postgres.New()
	if err := drv.Connect(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	_, err := drv.Exec(ctx, `CREATE TABLE IF NOT EXISTS integration_composite_pk (
		a INT NOT NULL,
		b INT NOT NULL,
		label TEXT NOT NULL,
		PRIMARY KEY (a, b)
	)`, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = drv.Exec(ctx, `DELETE FROM integration_composite_pk`, nil)
	rows := [][3]any{
		{1, 1, "r11"},
		{1, 2, "r12"},
		{2, 1, "r21"},
		{2, 2, "r22"},
		{3, 1, "r31"},
	}
	for _, row := range rows {
		if _, err := drv.Exec(ctx, `INSERT INTO integration_composite_pk(a, b, label) VALUES ($1, $2, $3)`, []any{row[0], row[1], row[2]}); err != nil {
			t.Fatal(err)
		}
	}

	cursor, err := drv.OpenTableExport(ctx, "integration_composite_pk", model.TableExportOptions{BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cursor.Close() }()

	total := 0
	for {
		batch, err := cursor.NextBatch(ctx)
		if err != nil {
			t.Fatal(err)
		}
		total += len(batch.Rows)
		if !batch.HasMore {
			break
		}
	}
	if total != 5 {
		t.Fatalf("expected 5 rows exported, got %d", total)
	}
}
