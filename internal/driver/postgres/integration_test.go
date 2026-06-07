package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/driver/postgres"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func testPostgresConfig(t *testing.T) model.DriverConfig {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set")
	}
	// DSN form: postgres://user:pass@host:port/db?sslmode=disable
	return model.DriverConfig{
		Type: model.DriverTypePostgres,
		Postgres: &model.PostgresConfig{
			Host:     "127.0.0.1",
			Port:     5432,
			Database: "testdb",
			User:     "test",
			Password: "test",
			SSLMode:  "disable",
			Schema:   "public",
			ReadOnly: false,
		},
	}
}

func TestIntegrationPostgresConnectListTables(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in -short mode")
	}
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
	if testing.Short() {
		t.Skip("integration test skipped in -short mode")
	}
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
		if _, err := drv.Exec(ctx, `INSERT INTO integration_export(label) VALUES ($1)`, []any{i}); err != nil {
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
