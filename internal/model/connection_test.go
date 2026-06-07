package model_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestPostgresConfigNormalized(t *testing.T) {
	pg := &model.PostgresConfig{}
	if pg.NormalizedPort() != 5432 {
		t.Fatalf("expected default port 5432, got %d", pg.NormalizedPort())
	}
	pg.Port = 5433
	if pg.NormalizedPort() != 5433 {
		t.Fatalf("expected port 5433, got %d", pg.NormalizedPort())
	}
	if pg.NormalizedSchema() != "public" {
		t.Fatalf("expected public schema, got %q", pg.NormalizedSchema())
	}
	pg.Schema = "app"
	if pg.NormalizedSchema() != "app" {
		t.Fatalf("expected app schema, got %q", pg.NormalizedSchema())
	}
	if pg.NormalizedSSLMode() != "disable" {
		t.Fatalf("expected disable ssl mode, got %q", pg.NormalizedSSLMode())
	}
	pg.SSLMode = "require"
	if pg.NormalizedSSLMode() != "require" {
		t.Fatalf("expected require ssl mode, got %q", pg.NormalizedSSLMode())
	}
}

func TestMySQLConfigNormalizedPort(t *testing.T) {
	my := &model.MySQLConfig{}
	if my.NormalizedPort() != 3306 {
		t.Fatalf("expected default port 3306, got %d", my.NormalizedPort())
	}
	my.Port = 3307
	if my.NormalizedPort() != 3307 {
		t.Fatalf("expected port 3307, got %d", my.NormalizedPort())
	}
}

func TestRemoteFingerprint(t *testing.T) {
	got := model.RemoteFingerprint(model.DriverTypePostgres, "localhost", 0, "db", "user")
	want := "postgres|localhost|5432|db|user"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = model.RemoteFingerprint(model.DriverTypeMySQL, "127.0.0.1", 0, "app", "root")
	want = "mysql|127.0.0.1|3306|app|root"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = model.RemoteFingerprint(model.DriverTypePostgres, "host", 9999, "db", "u")
	want = "postgres|host|9999|db|u"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
