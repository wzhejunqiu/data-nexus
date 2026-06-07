package wails_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	wailssvc "github.com/wzhejunqiu/data-nexus/internal/wails"
	"go.uber.org/zap"
)

func TestFileServiceWriteTextFile(t *testing.T) {
	svc := wailssvc.NewFileService(zap.NewNop())
	dir := t.TempDir()

	t.Run("csv", func(t *testing.T) {
		path := filepath.Join(dir, "out.csv")
		if err := svc.WriteTextFile(path, "a,b\n1,2", "utf-8"); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "a,b\n1,2" {
			t.Fatalf("unexpected content: %q", string(data))
		}
	})

	t.Run("tsv", func(t *testing.T) {
		path := filepath.Join(dir, "out.tsv")
		if err := svc.WriteTextFile(path, "x\ty", "utf-8"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("txt", func(t *testing.T) {
		path := filepath.Join(dir, "out.txt")
		if err := svc.WriteTextFile(path, "hello", "utf-8"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("utf-8-bom", func(t *testing.T) {
		path := filepath.Join(dir, "bom.csv")
		if err := svc.WriteTextFile(path, "data", "utf-8-bom"); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) < 3 || data[0] != 0xEF || data[1] != 0xBB || data[2] != 0xBF {
			t.Fatalf("expected UTF-8 BOM prefix, got %v", data[:min(3, len(data))])
		}
		if string(data[3:]) != "data" {
			t.Fatalf("unexpected content after BOM: %q", string(data[3:]))
		}
	})
}

func TestFileServiceWriteTextFileErrors(t *testing.T) {
	svc := wailssvc.NewFileService(zap.NewNop())
	dir := t.TempDir()

	assertCode := func(t *testing.T, err error, code string) {
		t.Helper()
		if err == nil {
			t.Fatal("expected error")
		}
		appErr, ok := err.(*model.AppError)
		if !ok || appErr.Code != code {
			t.Fatalf("expected %s, got %v", code, err)
		}
	}

	t.Run("empty path", func(t *testing.T) {
		assertCode(t, svc.WriteTextFile("", "x", "utf-8"), "INVALID_PATH")
	})

	t.Run("path traversal", func(t *testing.T) {
		assertCode(t, svc.WriteTextFile(filepath.Join(dir, "bad..name.csv"), "x", "utf-8"), "INVALID_PATH")
	})

	t.Run("invalid extension", func(t *testing.T) {
		assertCode(t, svc.WriteTextFile(filepath.Join(dir, "bad.exe"), "x", "utf-8"), "INVALID_PATH")
	})

	t.Run("unsupported encoding", func(t *testing.T) {
		assertCode(t, svc.WriteTextFile(filepath.Join(dir, "out.csv"), "x", "latin1"), "INVALID_REQUEST")
	})
}
