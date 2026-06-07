package wails_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	wailssvc "github.com/wzhejunqiu/data-nexus/internal/wails"
	"go.uber.org/zap"
)

type mockRuntime struct {
	title  string
	events []struct {
		name string
		data any
	}
	dialogMsg   string
	dialogErr   error
	filePath    string
	fileErr     error
	fileFilters int
	savePath    string
	saveErr     error
}

func (m *mockRuntime) WindowSetTitle(_ context.Context, title string) {
	m.title = title
}

func (m *mockRuntime) MessageDialog(_ context.Context, opts wailsruntime.MessageDialogOptions) (string, error) {
	m.dialogMsg = opts.Message
	if m.dialogErr != nil {
		return "", m.dialogErr
	}
	return "ok", nil
}

func (m *mockRuntime) EventsEmit(_ context.Context, event string, data ...any) {
	var payload any
	if len(data) > 0 {
		payload = data[0]
	}
	m.events = append(m.events, struct {
		name string
		data any
	}{name: event, data: payload})
}

func (m *mockRuntime) OpenFileDialog(_ context.Context, opts wailsruntime.OpenDialogOptions) (string, error) {
	m.fileFilters = len(opts.Filters)
	if m.fileErr != nil {
		return "", m.fileErr
	}
	return m.filePath, nil
}

func (m *mockRuntime) SaveFileDialog(_ context.Context, opts wailsruntime.SaveDialogOptions) (string, error) {
	m.fileFilters = len(opts.Filters)
	if m.saveErr != nil {
		return "", m.saveErr
	}
	if m.savePath != "" {
		return m.savePath, nil
	}
	return "/tmp/export.csv", nil
}

func newAppWithMock(t *testing.T) (*wailssvc.AppService, *mockRuntime) {
	t.Helper()
	rt := &mockRuntime{}
	svc := wailssvc.NewAppServiceWithRuntime(zap.NewNop(), rt)
	svc.SetContext(context.Background())
	return svc, rt
}

func TestSetWindowTitleWithContext(t *testing.T) {
	svc, rt := newAppWithMock(t)
	if err := svc.SetWindowTitle("My DB"); err != nil {
		t.Fatal(err)
	}
	if rt.title != "My DB" {
		t.Fatalf("expected title My DB, got %q", rt.title)
	}
}

func TestShowAboutWithContext(t *testing.T) {
	svc, rt := newAppWithMock(t)
	if err := svc.ShowAbout(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rt.dialogMsg, "Data Nexus v") {
		t.Fatalf("unexpected dialog message: %q", rt.dialogMsg)
	}
}

func TestShowAboutDialogError(t *testing.T) {
	svc, rt := newAppWithMock(t)
	rt.dialogErr = errors.New("dialog failed")
	err := svc.ShowAbout()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "dialog failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmitThemeChange(t *testing.T) {
	svc, rt := newAppWithMock(t)
	if err := svc.EmitThemeChange("dark"); err != nil {
		t.Fatal(err)
	}
	if len(rt.events) != 1 || rt.events[0].name != "app:theme" || rt.events[0].data != "dark" {
		t.Fatalf("unexpected events: %+v", rt.events)
	}
}

func TestEmitLanguageChange(t *testing.T) {
	svc, rt := newAppWithMock(t)
	if err := svc.EmitLanguageChange("zh-CN"); err != nil {
		t.Fatal(err)
	}
	if len(rt.events) != 1 || rt.events[0].name != "app:language" || rt.events[0].data != "zh-CN" {
		t.Fatalf("unexpected events: %+v", rt.events)
	}
}

func TestDialogServiceOpenDatabaseFileSuccess(t *testing.T) {
	rt := &mockRuntime{filePath: "/tmp/app.db"}
	svc := wailssvc.NewDialogServiceWithRuntime(zap.NewNop(), rt)
	svc.SetContext(context.Background())

	path, err := svc.OpenDatabaseFile()
	if err != nil {
		t.Fatal(err)
	}
	if path != "/tmp/app.db" {
		t.Fatalf("unexpected path %s", path)
	}
	if rt.fileFilters != 1 {
		t.Fatalf("expected file filter, got %d", rt.fileFilters)
	}
}

func TestDialogServiceOpenDatabaseFileCancelled(t *testing.T) {
	rt := &mockRuntime{filePath: ""}
	svc := wailssvc.NewDialogServiceWithRuntime(zap.NewNop(), rt)
	svc.SetContext(context.Background())

	_, err := svc.OpenDatabaseFile()
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "DIALOG_CANCELLED" {
		t.Fatalf("expected DIALOG_CANCELLED, got %v", err)
	}
}

func TestDialogServiceOpenDatabaseFileError(t *testing.T) {
	rt := &mockRuntime{fileErr: errors.New("picker failed")}
	svc := wailssvc.NewDialogServiceWithRuntime(zap.NewNop(), rt)
	svc.SetContext(context.Background())

	_, err := svc.OpenDatabaseFile()
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %v", err)
	}
}

func TestDialogServiceNilContext(t *testing.T) {
	svc := wailssvc.NewDialogServiceWithRuntime(zap.NewNop(), &mockRuntime{})
	_, err := svc.OpenDatabaseFile()
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %v", err)
	}
}
