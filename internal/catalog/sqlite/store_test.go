package sqlite_test

import (
	"testing"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/catalog"
	catalogsqlite "github.com/wzhejunqiu/data-nexus/internal/catalog/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestStoreFirstInitCreatesDefaultGroup(t *testing.T) {
	dir := t.TempDir()
	store, err := catalogsqlite.NewStore(dir + "/catalog.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	tree, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Groups) != 1 {
		t.Fatalf("expected 1 default group, got %d", len(tree.Groups))
	}
	if tree.Groups[0].Name != model.DefaultGroupName {
		t.Fatalf("expected default group name %q, got %q", model.DefaultGroupName, tree.Groups[0].Name)
	}
}

func TestStoreNestedGroupAndFreeConnection(t *testing.T) {
	dir := t.TempDir()
	store, err := catalogsqlite.NewStore(dir + "/catalog.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	child, err := store.CreateGroup(nil, "Production")
	if err != nil {
		t.Fatal(err)
	}
	nested, err := store.CreateGroup(&child.ID, "Nested")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	conn := model.SavedConnection{
		ID:         "conn-1",
		Name:       "app.db",
		Type:       model.DriverTypeSQLite,
		Config:     model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: "/tmp/app.db"}},
		CreatedAt:  now,
		UpdatedAt:  now,
		LastUsedAt: now,
	}
	if err := store.UpsertConnection(conn, catalog.ConnectionPlacement{GroupID: &nested.ID, SortOrder: -1}); err != nil {
		t.Fatal(err)
	}

	free := model.SavedConnection{
		ID:         "conn-2",
		Name:       "staging.db",
		Type:       model.DriverTypeSQLite,
		Config:     model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: "/tmp/staging.db"}},
		CreatedAt:  now,
		UpdatedAt:  now,
		LastUsedAt: now,
	}
	if err := store.UpsertConnection(free, catalog.ConnectionPlacement{SortOrder: -1}); err != nil {
		t.Fatal(err)
	}

	tree, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.FreeConnections) != 1 {
		t.Fatalf("expected 1 free connection, got %d", len(tree.FreeConnections))
	}
	if tree.FreeConnections[0].ID != "conn-2" {
		t.Fatalf("expected free conn-2, got %s", tree.FreeConnections[0].ID)
	}

	foundNested := false
	for _, g := range tree.Groups {
		if g.ID == child.ID {
			if len(g.ChildGroups) != 1 || g.ChildGroups[0].ID != nested.ID {
				t.Fatal("expected nested child group")
			}
			if len(g.ChildGroups[0].Connections) != 1 || g.ChildGroups[0].Connections[0].ID != "conn-1" {
				t.Fatal("expected conn-1 in nested group")
			}
			foundNested = true
		}
	}
	if !foundNested {
		t.Fatal("production group not found in tree")
	}
}

func TestStoreDeleteGroupReleasesConnections(t *testing.T) {
	dir := t.TempDir()
	store, err := catalogsqlite.NewStore(dir + "/catalog.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	grp, err := store.CreateGroup(nil, "ToDelete")
	if err != nil {
		t.Fatal(err)
	}
	conn := model.SavedConnection{
		ID:         "conn-del",
		Name:       "x.db",
		Type:       model.DriverTypeSQLite,
		Config:     model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: "/tmp/x.db"}},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		LastUsedAt: time.Now().UTC(),
	}
	if err := store.UpsertConnection(conn, catalog.ConnectionPlacement{GroupID: &grp.ID, SortOrder: -1}); err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteGroup(model.DeleteGroupRequest{ID: grp.ID, DeleteConnections: false}, func(id string) error {
		return store.RemoveConnection(id)
	}); err != nil {
		t.Fatal(err)
	}

	tree, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.FreeConnections) != 1 || tree.FreeConnections[0].ID != "conn-del" {
		t.Fatal("expected conn-del as free connection after group delete")
	}
}

func TestStoreRestoreOpenOnStartup(t *testing.T) {
	dir := t.TempDir()
	store, err := catalogsqlite.NewStore(dir + "/catalog.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	if err := store.SetRestoreOpenOnStartup(true); err != nil {
		t.Fatal(err)
	}
	v, err := store.RestoreOpenOnStartup()
	if err != nil || !v {
		t.Fatal("expected restore true")
	}
	if err := store.SetOpenConnectionIDs([]string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	ids, err := store.OpenConnectionIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != "a" {
		t.Fatalf("unexpected ids: %v", ids)
	}
}
