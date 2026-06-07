package wails_test

import (
	"errors"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	wailssvc "github.com/wzhejunqiu/data-nexus/internal/wails"
	"go.uber.org/zap"
)

func TestAppServiceGetVersion(t *testing.T) {
	svc := wailssvc.NewAppService(zap.NewNop())
	info, err := svc.GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	if info.Version == "" || info.Platform == "" {
		t.Fatalf("unexpected version info: %+v", info)
	}
}

func TestConnectionFailedErrorCode(t *testing.T) {
	err := model.ErrConnectionFailed("database file does not exist")
	var appErr *model.AppError
	if !errors.As(err, &appErr) || appErr.Code != "CONNECTION_FAILED" {
		t.Fatalf("unexpected error: %v", err)
	}
}
