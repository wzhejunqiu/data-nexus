package mysql

import (
	"strings"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestBuildMySQLDSNWithoutTLS(t *testing.T) {
	dsn := buildMySQLDSN(&model.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "app",
		User:     "root",
	}, "secret")
	if strings.Contains(dsn, "tls=") {
		t.Fatalf("expected no tls param, got %q", dsn)
	}
	if !strings.Contains(dsn, "root:secret@tcp(localhost:3306)/app?") {
		t.Fatalf("unexpected dsn: %q", dsn)
	}
}

func TestBuildMySQLDSNWithTLSVerify(t *testing.T) {
	dsn := buildMySQLDSN(&model.MySQLConfig{
		Host:     "db.example.com",
		Port:     3306,
		Database: "app",
		User:     "app",
		TLS:      true,
	}, "pw")
	if !strings.Contains(dsn, "tls=true") {
		t.Fatalf("expected tls=true, got %q", dsn)
	}
	if strings.Contains(dsn, "skip-verify") {
		t.Fatalf("expected no skip-verify, got %q", dsn)
	}
}

func TestBuildMySQLDSNWithTLSSkipVerify(t *testing.T) {
	dsn := buildMySQLDSN(&model.MySQLConfig{
		Host:          "db.example.com",
		Port:          3306,
		Database:      "app",
		User:          "app",
		TLS:           true,
		TLSSkipVerify: true,
	}, "pw")
	if !strings.Contains(dsn, "tls=skip-verify") {
		t.Fatalf("expected tls=skip-verify, got %q", dsn)
	}
}
