package version_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/version"
)

func TestVersionMatchesWailsJSON(t *testing.T) {
	got := version.Version()
	if got == "" {
		t.Fatal("Version() returned empty string")
	}

	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "wails.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if got != cfg.Info.ProductVersion {
		t.Fatalf("Version() = %q, wails.json productVersion = %q", got, cfg.Info.ProductVersion)
	}

	embedded, err := os.ReadFile(filepath.Join("..", "version", "product.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got != strings.TrimSpace(string(embedded)) {
		t.Fatalf("Version() = %q, product.txt = %q", got, strings.TrimSpace(string(embedded)))
	}
}
