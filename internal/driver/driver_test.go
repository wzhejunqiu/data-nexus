package driver_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/driver"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestNewDriverSQLite(t *testing.T) {
	drv, err := driver.NewDriver(model.DriverTypeSQLite)
	if err != nil {
		t.Fatal(err)
	}
	if drv.Type() != model.DriverTypeSQLite {
		t.Fatalf("unexpected type %s", drv.Type())
	}
}

func TestNewDriverUnsupported(t *testing.T) {
	_, err := driver.NewDriver(model.DriverType("postgres"))
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
}
