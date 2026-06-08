package mysql_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/driver/mysql"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/testutil/mysqlserver"
)

var testMySQL *mysqlserver.Server

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Short() {
		srv, err := mysqlserver.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "mysql integration harness: %v\n", err)
			os.Exit(1)
		}
		testMySQL = srv
	}
	code := m.Run()
	if testMySQL != nil {
		_ = testMySQL.Close()
	}
	os.Exit(code)
}

func testMySQLConfig(t *testing.T) model.DriverConfig {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test skipped in -short mode")
	}
	if testMySQL == nil {
		t.Fatal("mysql integration harness not started")
	}
	return model.DriverConfig{
		Type: model.DriverTypeMySQL,
		MySQL: &model.MySQLConfig{
			Host:     testMySQL.Host,
			Port:     testMySQL.Port,
			Database: testMySQL.Database,
			User:     "root",
			Password: "",
			TLS:      false,
			ReadOnly: false,
		},
	}
}

func TestIntegrationMySQLConnectListTables(t *testing.T) {
	cfg := testMySQLConfig(t)
	ctx := context.Background()

	drv := mysql.New()
	if err := drv.Connect(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	if err := drv.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	_, err := drv.Exec(ctx, `CREATE TABLE IF NOT EXISTS integration_items (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL
	)`, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = drv.Exec(ctx, `INSERT INTO integration_items(name) SELECT 'alpha' FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM integration_items WHERE name = 'alpha')`, nil)
	if err != nil {
		t.Fatal(err)
	}

	tables, err := drv.ListTables(ctx, model.ListTablesOptions{})
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

func TestIntegrationMySQLExportCursor(t *testing.T) {
	cfg := testMySQLConfig(t)
	ctx := context.Background()

	drv := mysql.New()
	if err := drv.Connect(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	_, err := drv.Exec(ctx, `CREATE TABLE IF NOT EXISTS integration_export (
		id INT AUTO_INCREMENT PRIMARY KEY,
		label VARCHAR(64) NOT NULL
	)`, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = drv.Exec(ctx, `DELETE FROM integration_export`, nil)
	for i := 1; i <= 5; i++ {
		if _, err := drv.Exec(ctx, `INSERT INTO integration_export(label) VALUES (?)`, []any{i}); err != nil {
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

func TestIntegrationMySQLExportCompositePK(t *testing.T) {
	cfg := testMySQLConfig(t)
	ctx := context.Background()

	drv := mysql.New()
	if err := drv.Connect(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	_, err := drv.Exec(ctx, `CREATE TABLE IF NOT EXISTS integration_composite_pk (
		a INT NOT NULL,
		b INT NOT NULL,
		label VARCHAR(64) NOT NULL,
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
		if _, err := drv.Exec(ctx, `INSERT INTO integration_composite_pk(a, b, label) VALUES (?, ?, ?)`, []any{row[0], row[1], row[2]}); err != nil {
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

func TestIntegrationMySQLExportNoPKStreaming(t *testing.T) {
	cfg := testMySQLConfig(t)
	ctx := context.Background()

	drv := mysql.New()
	if err := drv.Connect(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = drv.Close() }()

	_, err := drv.Exec(ctx, `CREATE TABLE IF NOT EXISTS integration_no_pk (
		label VARCHAR(64) NOT NULL
	)`, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = drv.Exec(ctx, `DELETE FROM integration_no_pk`, nil)
	for i := 1; i <= 5; i++ {
		if _, err := drv.Exec(ctx, `INSERT INTO integration_no_pk(label) VALUES (?)`, []any{fmt.Sprintf("%d", i)}); err != nil {
			t.Fatal(err)
		}
	}

	cursor, err := drv.OpenTableExport(ctx, "integration_no_pk", model.TableExportOptions{BatchSize: 2})
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
