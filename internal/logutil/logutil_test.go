package logutil_test

import (
	"strings"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/logutil"
)

func TestSQLPreview(t *testing.T) {
	t.Parallel()
	preview := logutil.SQLPreview("  SELECT  *  FROM   users\nWHERE id = 1  ", 20)
	if preview != "SELECT * FROM users..." {
		t.Fatalf("unexpected preview: %q", preview)
	}
	if logutil.SQLPreview("short", 20) != "short" {
		t.Fatal("expected unchanged short sql")
	}
	long := strings.Repeat("a", 200)
	if !strings.HasSuffix(logutil.SQLPreview(long, 50), "...") {
		t.Fatal("expected truncated suffix")
	}
}
