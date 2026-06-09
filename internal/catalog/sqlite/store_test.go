package sqlite_test

import (
	"strings"
	"testing"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/catalog"
	catalogsqlite "github.com/wzhejunqiu/data-nexus/internal/catalog/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func newTestStore(t *testing.T) *catalogsqlite.Store {
	dir := t.TempDir()
	store, err := catalogsqlite.NewStore(dir + "/catalog.db")
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func assertInvalidRequest(t *testing.T, err error, msgContains string) {
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "INVALID_REQUEST" {
		t.Fatalf("expected INVALID_REQUEST, got %v", err)
	}
	if msgContains != "" && !strings.Contains(appErr.Message, msgContains) {
		t.Fatalf("expected message containing %q, got %q", msgContains, appErr.Message)
	}
}

func sqliteConn(id, name, path string, now time.Time) model.SavedConnection {
	return model.SavedConnection{
		ID:         id,
		Name:       name,
		Type:       model.DriverTypeSQLite,
		Config:     model.DriverConfig{Type: model.DriverTypeSQLite, SQLite: &model.SQLiteConfig{FilePath: path}},
		CreatedAt:  now,
		UpdatedAt:  now,
		LastUsedAt: now,
	}
}

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

	preview, err := store.GetGroupDeletePreview(child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if preview.SubgroupCount != 1 {
		t.Fatalf("expected 1 subgroup, got %d", preview.SubgroupCount)
	}
	if preview.ConnCount != 1 {
		t.Fatalf("expected 1 connection in subtree, got %d", preview.ConnCount)
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

func TestStoreMoveGroupToParent(t *testing.T) {
	store := newTestStore(t)
	defer func() { _ = store.Close() }()

	target, err := store.CreateGroup(nil, "Target")
	if err != nil {
		t.Fatal(err)
	}
	movable, err := store.CreateGroup(nil, "Movable")
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MoveGroup(model.MoveGroupRequest{
		ID:          movable.ID,
		NewParentID: &target.ID,
		SortOrder:   -1,
	}); err != nil {
		t.Fatal(err)
	}

	tree, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, g := range tree.Groups {
		if g.ID == target.ID {
			if len(g.ChildGroups) != 1 || g.ChildGroups[0].ID != movable.ID {
				t.Fatalf("expected movable under target, got child groups %v", g.ChildGroups)
			}
			found = true
		}
		if g.ID == movable.ID {
			t.Fatal("movable should not remain at root")
		}
	}
	if !found {
		t.Fatal("target group not found")
	}
}

func TestStoreMoveGroupToRoot(t *testing.T) {
	store := newTestStore(t)
	defer func() { _ = store.Close() }()

	parent, err := store.CreateGroup(nil, "Parent")
	if err != nil {
		t.Fatal(err)
	}
	child, err := store.CreateGroup(&parent.ID, "Child")
	if err != nil {
		t.Fatal(err)
	}

	if err := store.MoveGroup(model.MoveGroupRequest{
		ID:          child.ID,
		NewParentID: nil,
		SortOrder:   -1,
	}); err != nil {
		t.Fatal(err)
	}

	tree, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	childAtRoot := false
	for _, ref := range tree.RootItems {
		if ref.ItemType == "group" && ref.ItemID == child.ID {
			childAtRoot = true
		}
	}
	if !childAtRoot {
		t.Fatal("expected child in root items")
	}
	for _, g := range tree.Groups {
		if g.ID == parent.ID && len(g.ChildGroups) != 0 {
			t.Fatal("parent should have no child groups")
		}
	}
}

func TestStoreReorderSidebarRoot(t *testing.T) {
	store := newTestStore(t)
	defer func() { _ = store.Close() }()

	g1, err := store.CreateGroup(nil, "G1")
	if err != nil {
		t.Fatal(err)
	}
	g2, err := store.CreateGroup(nil, "G2")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.UpsertConnection(
		sqliteConn("c1", "a.db", "/tmp/a.db", now),
		catalog.ConnectionPlacement{SortOrder: -1},
	); err != nil {
		t.Fatal(err)
	}

	tree, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.RootItems) < 2 {
		t.Fatalf("expected multiple root items, got %d", len(tree.RootItems))
	}

	reversed := make([]model.SidebarRootItemRef, len(tree.RootItems))
	for i, ref := range tree.RootItems {
		reversed[len(tree.RootItems)-1-i] = ref
	}
	if err := store.ReorderSidebarRoot(reversed); err != nil {
		t.Fatal(err)
	}

	after, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.RootItems) != len(reversed) {
		t.Fatalf("root item count: want %d got %d", len(reversed), len(after.RootItems))
	}
	for i, want := range reversed {
		if after.RootItems[i].ItemType != want.ItemType || after.RootItems[i].ItemID != want.ItemID {
			t.Fatalf("root[%d]: want %s:%s got %s:%s",
				i, want.ItemType, want.ItemID, after.RootItems[i].ItemType, after.RootItems[i].ItemID)
		}
	}
	_ = g1
	_ = g2
}

func TestStoreReorderGroupMembers(t *testing.T) {
	store := newTestStore(t)
	defer func() { _ = store.Close() }()

	grp, err := store.CreateGroup(nil, "Members")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.UpsertConnection(
		sqliteConn("c-a", "a.db", "/tmp/a.db", now),
		catalog.ConnectionPlacement{GroupID: &grp.ID, SortOrder: -1},
	); err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertConnection(
		sqliteConn("c-b", "b.db", "/tmp/b.db", now),
		catalog.ConnectionPlacement{GroupID: &grp.ID, SortOrder: -1},
	); err != nil {
		t.Fatal(err)
	}
	sub, err := store.CreateGroup(&grp.ID, "Sub")
	if err != nil {
		t.Fatal(err)
	}

	ordered := []model.GroupMemberRef{
		{MemberType: "group", MemberID: sub.ID},
		{MemberType: "connection", MemberID: "c-b"},
		{MemberType: "connection", MemberID: "c-a"},
	}
	if err := store.ReorderGroupMembers(grp.ID, ordered); err != nil {
		t.Fatal(err)
	}

	tree, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var members []model.GroupMemberRef
	for _, g := range tree.Groups {
		if g.ID == grp.ID {
			members = g.MemberItems
			break
		}
	}
	if len(members) != 3 {
		t.Fatalf("expected 3 members, got %d", len(members))
	}
	for i, want := range ordered {
		if members[i].MemberType != want.MemberType || members[i].MemberID != want.MemberID {
			t.Fatalf("member[%d]: want %s:%s got %s:%s",
				i, want.MemberType, want.MemberID, members[i].MemberType, members[i].MemberID)
		}
	}
}

func TestStoreMoveConnectionToGroupAndRelease(t *testing.T) {
	store := newTestStore(t)
	defer func() { _ = store.Close() }()

	grp, err := store.CreateGroup(nil, "Bucket")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.UpsertConnection(
		sqliteConn("free-1", "free.db", "/tmp/free.db", now),
		catalog.ConnectionPlacement{SortOrder: -1},
	); err != nil {
		t.Fatal(err)
	}

	if err := store.MoveConnectionToGroup("free-1", grp.ID, -1); err != nil {
		t.Fatal(err)
	}
	tree, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.FreeConnections) != 0 {
		t.Fatal("expected no free connections after move to group")
	}
	var inGroup bool
	for _, g := range tree.Groups {
		if g.ID == grp.ID {
			for _, c := range g.Connections {
				if c.ID == "free-1" {
					inGroup = true
				}
			}
		}
	}
	if !inGroup {
		t.Fatal("connection not found in group")
	}

	if err := store.ReleaseConnection("free-1", -1); err != nil {
		t.Fatal(err)
	}
	after, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.FreeConnections) != 1 || after.FreeConnections[0].ID != "free-1" {
		t.Fatal("expected free-1 as free connection")
	}
}

func TestStoreMoveGroupIntoItself(t *testing.T) {
	store := newTestStore(t)
	defer func() { _ = store.Close() }()

	grp, err := store.CreateGroup(nil, "Self")
	if err != nil {
		t.Fatal(err)
	}
	before, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = store.MoveGroup(model.MoveGroupRequest{
		ID:          grp.ID,
		NewParentID: &grp.ID,
		SortOrder:   -1,
	})
	assertInvalidRequest(t, err, "itself")

	after, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.RootItems) != len(after.RootItems) {
		t.Fatal("root items changed after rejected move")
	}
	for _, g := range after.Groups {
		if g.ID == grp.ID && len(g.ChildGroups) != 0 {
			t.Fatal("group should not have children after rejected self-move")
		}
	}
}

func TestStoreMoveGroupIntoDescendant(t *testing.T) {
	store := newTestStore(t)
	defer func() { _ = store.Close() }()

	parent, err := store.CreateGroup(nil, "Parent")
	if err != nil {
		t.Fatal(err)
	}
	child, err := store.CreateGroup(&parent.ID, "Child")
	if err != nil {
		t.Fatal(err)
	}
	grand, err := store.CreateGroup(&child.ID, "Grand")
	if err != nil {
		t.Fatal(err)
	}

	before, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = store.MoveGroup(model.MoveGroupRequest{
		ID:          parent.ID,
		NewParentID: &child.ID,
		SortOrder:   -1,
	})
	assertInvalidRequest(t, err, "descendant")

	after, err := store.GetSidebarTree(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Groups) != len(after.Groups) {
		t.Fatal("group count changed after rejected move")
	}
	for _, g := range after.Groups {
		if g.ID == parent.ID {
			if len(g.ChildGroups) != 1 || g.ChildGroups[0].ID != child.ID {
				t.Fatal("parent-child structure should remain unchanged")
			}
		}
	}
	_ = grand
}
