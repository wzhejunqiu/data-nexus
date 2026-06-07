package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
)

func TestQueryStoreSaveListDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "queries.json")

	store, err := service.NewQueryStore(path)
	if err != nil {
		t.Fatal(err)
	}

	item, err := store.SaveQuery(model.SaveCannedQueryRequest{
		Name: "My Query",
		SQL:  "SELECT 1;",
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.ID == "" || item.Name != "My Query" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	reloaded, err := service.NewQueryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	items := reloaded.List()
	if len(items) != 1 || items[0].SQL != "SELECT 1;" {
		t.Fatalf("unexpected persisted items: %+v", items)
	}

	if err := reloaded.Delete(item.ID); err != nil {
		t.Fatal(err)
	}
	if err := reloaded.Save(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.List()) != 0 {
		t.Fatalf("expected empty list after delete, file: %s", string(data))
	}
}
