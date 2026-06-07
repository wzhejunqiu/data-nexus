package service_test

import (
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
)

func TestCannedQueryServiceCRUD(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewQueryStore(filepath.Join(dir, "queries.json"))
	if err != nil {
		t.Fatal(err)
	}
	svc := service.NewCannedQueryService(store)

	list, err := svc.ListCannedQueries()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 0 {
		t.Fatalf("expected empty list, got %+v", list.Items)
	}

	saved, err := svc.SaveCannedQuery(model.SaveCannedQueryRequest{
		Name: "Count rows",
		SQL:  "SELECT COUNT(*) FROM users;",
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == "" || saved.Name != "Count rows" {
		t.Fatalf("unexpected saved query: %+v", saved)
	}

	list, err = svc.ListCannedQueries()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].SQL != "SELECT COUNT(*) FROM users;" {
		t.Fatalf("unexpected list after save: %+v", list.Items)
	}

	if err := svc.DeleteCannedQuery(saved.ID); err != nil {
		t.Fatal(err)
	}
	list, err = svc.ListCannedQueries()
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 0 {
		t.Fatalf("expected empty list after delete, got %+v", list.Items)
	}
}
