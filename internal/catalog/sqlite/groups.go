package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func (s *Store) GetSidebarTree(
	openStatus map[string]model.ConnectionStatus,
	connectedAt map[string]*time.Time,
) (*model.ConnectionSidebarTree, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	saved, err := s.listConnectionsLocked()
	if err != nil {
		return nil, err
	}
	connMap := make(map[string]model.SavedConnection, len(saved))
	for _, c := range saved {
		connMap[c.ID] = c
	}

	rootOrder, err := s.loadSidebarRootOrderLocked()
	if err != nil {
		return nil, err
	}

	allGroups, err := s.loadAllGroupsLocked()
	if err != nil {
		return nil, err
	}
	groupMap := make(map[string]model.ConnectionGroup, len(allGroups))
	for _, g := range allGroups {
		groupMap[g.ID] = g
	}

	membersByGroup, err := s.loadGroupMembersLocked()
	if err != nil {
		return nil, err
	}

	freeConns, err := s.loadFreeConnectionsLocked()
	if err != nil {
		return nil, err
	}

	toListItem := func(id string) (model.ConnectionListItem, bool) {
		c, ok := connMap[id]
		if !ok {
			return model.ConnectionListItem{}, false
		}
		item := model.ConnectionListItem{
			ID:             c.ID,
			Name:           c.Name,
			Type:           c.Type,
			Config:         c.Config,
			SecretsBackend: c.SecretsBackend,
			Status:         model.ConnectionStatusClosed,
			LastUsedAt:     c.LastUsedAt,
		}
		if st, ok := openStatus[id]; ok {
			item.Status = st
		}
		if t, ok := connectedAt[id]; ok && t != nil {
			item.ConnectedAt = t
		}
		return item, true
	}

	var buildGroupNode func(groupID string) model.ConnectionGroupNode
	buildGroupNode = func(groupID string) model.ConnectionGroupNode {
		g := groupMap[groupID]
		node := model.ConnectionGroupNode{
			ID:          g.ID,
			Name:        g.Name,
			ChildGroups: []model.ConnectionGroupNode{},
			Connections: []model.ConnectionListItem{},
			MemberItems: []model.GroupMemberRef{},
		}
		for _, m := range membersByGroup[groupID] {
			node.MemberItems = append(node.MemberItems, model.GroupMemberRef{
				MemberType: m.MemberType,
				MemberID:   m.MemberID,
			})
			switch m.MemberType {
			case "group":
				node.ChildGroups = append(node.ChildGroups, buildGroupNode(m.MemberID))
			case "connection":
				if item, ok := toListItem(m.MemberID); ok {
					node.Connections = append(node.Connections, item)
				}
			}
		}
		return node
	}

	tree := &model.ConnectionSidebarTree{
		Groups:          []model.ConnectionGroupNode{},
		FreeConnections: []model.ConnectionListItem{},
		RootItems:       []model.SidebarRootItemRef{},
	}

	freeSet := make(map[string]int, len(freeConns))
	for _, fc := range freeConns {
		freeSet[fc.ConnectionID] = fc.SortOrder
	}

	orderedRootGroups := []string{}
	orderedFree := []string{}
	rootItems := []model.SidebarRootItemRef{}
	seenRoot := make(map[string]struct{})
	appendRootItem := func(itemType, itemID string) {
		key := itemType + ":" + itemID
		if _, ok := seenRoot[key]; ok {
			return
		}
		seenRoot[key] = struct{}{}
		rootItems = append(rootItems, model.SidebarRootItemRef{ItemType: itemType, ItemID: itemID})
		switch itemType {
		case "group":
			orderedRootGroups = append(orderedRootGroups, itemID)
		case "connection":
			orderedFree = append(orderedFree, itemID)
		}
	}
	for _, ro := range rootOrder {
		switch ro.ItemType {
		case "group":
			if g, ok := groupMap[ro.ItemID]; ok && g.ParentID == nil {
				appendRootItem("group", ro.ItemID)
			}
		case "connection":
			if _, ok := freeSet[ro.ItemID]; ok {
				appendRootItem("connection", ro.ItemID)
			}
		}
	}
	// Include root groups not in sidebar_root_items
	for _, g := range allGroups {
		if g.ParentID == nil {
			found := false
			for _, id := range orderedRootGroups {
				if id == g.ID {
					found = true
					break
				}
			}
			if !found {
				appendRootItem("group", g.ID)
			}
		}
	}
	for _, id := range orderedRootGroups {
		tree.Groups = append(tree.Groups, buildGroupNode(id))
	}

	if len(orderedFree) == 0 {
		for id := range freeSet {
			orderedFree = append(orderedFree, id)
		}
	}
	for _, id := range orderedFree {
		if _, ok := seenRoot["connection:"+id]; !ok {
			appendRootItem("connection", id)
		}
		if item, ok := toListItem(id); ok {
			tree.FreeConnections = append(tree.FreeConnections, item)
		}
	}

	tree.RootItems = rootItems

	return tree, nil
}

type rootItem struct {
	ItemType  string
	ItemID    string
	SortOrder int
}

type groupMember struct {
	MemberType string
	MemberID   string
	SortOrder  int
}

type freeConn struct {
	ConnectionID string
	SortOrder    int
}

func (s *Store) listConnectionsLocked() ([]model.SavedConnection, error) {
	rows, err := s.db.Query(`
		SELECT id, name, driver_type, config_json, secrets_backend, created_at, updated_at, last_used_at
		FROM connections ORDER BY last_used_at DESC`)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	defer func() { _ = rows.Close() }()
	var items []model.SavedConnection
	for rows.Next() {
		item, err := scanSavedConnection(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) loadSidebarRootOrderLocked() ([]rootItem, error) {
	rows, err := s.db.Query(`SELECT item_type, item_id, sort_order FROM sidebar_root_items ORDER BY sort_order`)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	defer func() { _ = rows.Close() }()
	var items []rootItem
	for rows.Next() {
		var ri rootItem
		if err := rows.Scan(&ri.ItemType, &ri.ItemID, &ri.SortOrder); err != nil {
			return nil, model.ErrInternal(err.Error())
		}
		items = append(items, ri)
	}
	return items, rows.Err()
}

func (s *Store) loadAllGroupsLocked() ([]model.ConnectionGroup, error) {
	rows, err := s.db.Query(`SELECT id, name, parent_id, sort_order, created_at, updated_at FROM connection_groups`)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	defer func() { _ = rows.Close() }()
	var groups []model.ConnectionGroup
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (s *Store) loadGroupMembersLocked() (map[string][]groupMember, error) {
	rows, err := s.db.Query(`SELECT group_id, member_type, member_id, sort_order FROM group_members ORDER BY sort_order`)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	defer func() { _ = rows.Close() }()
	result := make(map[string][]groupMember)
	for rows.Next() {
		var groupID string
		var m groupMember
		if err := rows.Scan(&groupID, &m.MemberType, &m.MemberID, &m.SortOrder); err != nil {
			return nil, model.ErrInternal(err.Error())
		}
		result[groupID] = append(result[groupID], m)
	}
	return result, rows.Err()
}

func (s *Store) loadFreeConnectionsLocked() ([]freeConn, error) {
	rows, err := s.db.Query(`SELECT connection_id, sort_order FROM free_connections ORDER BY sort_order`)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	defer func() { _ = rows.Close() }()
	var items []freeConn
	for rows.Next() {
		var fc freeConn
		if err := rows.Scan(&fc.ConnectionID, &fc.SortOrder); err != nil {
			return nil, model.ErrInternal(err.Error())
		}
		items = append(items, fc)
	}
	return items, rows.Err()
}

func scanGroup(row scannable) (model.ConnectionGroup, error) {
	var g model.ConnectionGroup
	var parentID sql.NullString
	var createdAt, updatedAt string
	if err := row.Scan(&g.ID, &g.Name, &parentID, &g.SortOrder, &createdAt, &updatedAt); err != nil {
		return g, model.ErrInternal(err.Error())
	}
	if parentID.Valid {
		pid := parentID.String
		g.ParentID = &pid
	}
	var err error
	g.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return g, model.ErrInternal(err.Error())
	}
	g.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return g, model.ErrInternal(err.Error())
	}
	return g, nil
}

func (s *Store) CreateGroup(parentID *string, name string) (*model.ConnectionGroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	id := uuid.NewString()
	sortOrder, err := nextGroupSortOrderTx(tx, parentID)
	if err != nil {
		return nil, err
	}
	nowStr := now.Format(time.RFC3339Nano)
	var parentVal sql.NullString
	if parentID != nil && *parentID != "" {
		parentVal = sql.NullString{String: *parentID, Valid: true}
		if _, err := tx.Exec(`INSERT INTO group_members (group_id, member_type, member_id, sort_order) VALUES (?, 'group', ?, ?)`,
			*parentID, id, sortOrder); err != nil {
			return nil, model.ErrInternal(err.Error())
		}
	} else {
		if err := insertSidebarRootItemTx(tx, "group", id, -1); err != nil {
			return nil, err
		}
	}
	_, err = tx.Exec(`
		INSERT INTO connection_groups (id, name, parent_id, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		id, name, parentVal, sortOrder, nowStr, nowStr)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	if err := tx.Commit(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	g := &model.ConnectionGroup{
		ID: id, Name: name, ParentID: parentID, SortOrder: sortOrder,
		CreatedAt: now, UpdatedAt: now,
	}
	return g, nil
}

func nextGroupSortOrderTx(tx *sql.Tx, parentID *string) (int, error) {
	if parentID != nil && *parentID != "" {
		var max sql.NullInt64
		if err := tx.QueryRow(`SELECT MAX(sort_order) FROM group_members WHERE group_id = ?`, *parentID).Scan(&max); err != nil {
			return 0, model.ErrInternal(err.Error())
		}
		return int(max.Int64) + 1, nil
	}
	var max sql.NullInt64
	if err := tx.QueryRow(`SELECT MAX(sort_order) FROM connection_groups WHERE parent_id IS NULL`).Scan(&max); err != nil {
		return 0, model.ErrInternal(err.Error())
	}
	return int(max.Int64) + 1, nil
}

func insertSidebarRootItemTx(tx *sql.Tx, itemType, itemID string, sortOrder int) error {
	if sortOrder < 0 {
		var max sql.NullInt64
		_ = tx.QueryRow(`SELECT MAX(sort_order) FROM sidebar_root_items`).Scan(&max)
		sortOrder = int(max.Int64) + 1
	}
	_, err := tx.Exec(`
		INSERT OR REPLACE INTO sidebar_root_items (item_type, item_id, sort_order) VALUES (?, ?, ?)`,
		itemType, itemID, sortOrder)
	return err
}

func (s *Store) RenameGroup(id, name string) (*model.ConnectionGroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := s.db.Exec(`UPDATE connection_groups SET name = ?, updated_at = ? WHERE id = ?`, name, now, id)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, model.ErrInvalidRequest("group not found: " + id)
	}
	row := s.db.QueryRow(`SELECT id, name, parent_id, sort_order, created_at, updated_at FROM connection_groups WHERE id = ?`, id)
	g, err := scanGroup(row)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *Store) CountConnectionsInGroup(id string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	descendants, err := s.collectGroupSubtreeIDsLocked(id)
	if err != nil {
		return 0, err
	}
	if len(descendants) == 0 {
		return 0, nil
	}
	query := `SELECT COUNT(DISTINCT member_id) FROM group_members WHERE member_type = 'connection' AND group_id IN (`
	args := make([]any, len(descendants))
	for i, gid := range descendants {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = gid
	}
	query += ")"
	var count int
	if err := s.db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, model.ErrInternal(err.Error())
	}
	return count, nil
}

func (s *Store) GetGroupDeletePreview(id string) (*model.GroupDeletePreview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	descendants, err := s.collectGroupSubtreeIDsLocked(id)
	if err != nil {
		return nil, err
	}
	subgroupCount := len(descendants) - 1
	if subgroupCount < 0 {
		subgroupCount = 0
	}
	connCount := 0
	if len(descendants) > 0 {
		query := `SELECT COUNT(DISTINCT member_id) FROM group_members WHERE member_type = 'connection' AND group_id IN (`
		args := make([]any, len(descendants))
		for i, gid := range descendants {
			if i > 0 {
				query += ","
			}
			query += "?"
			args[i] = gid
		}
		query += ")"
		if err := s.db.QueryRow(query, args...).Scan(&connCount); err != nil {
			return nil, model.ErrInternal(err.Error())
		}
	}
	return &model.GroupDeletePreview{
		SubgroupCount: subgroupCount,
		ConnCount:     connCount,
	}, nil
}

func (s *Store) collectGroupSubtreeIDsLocked(rootID string) ([]string, error) {
	all, err := s.loadAllGroupsLocked()
	if err != nil {
		return nil, err
	}
	children := make(map[string][]string)
	for _, g := range all {
		if g.ParentID != nil {
			children[*g.ParentID] = append(children[*g.ParentID], g.ID)
		}
	}
	// Also find child groups via group_members
	members, err := s.loadGroupMembersLocked()
	if err != nil {
		return nil, err
	}
	for gid, ms := range members {
		for _, m := range ms {
			if m.MemberType == "group" {
				children[gid] = append(children[gid], m.MemberID)
			}
		}
	}
	var result []string
	visited := make(map[string]struct{})
	var walk func(string)
	walk = func(id string) {
		if _, ok := visited[id]; ok {
			return
		}
		visited[id] = struct{}{}
		result = append(result, id)
		for _, child := range children[id] {
			walk(child)
		}
	}
	walk(rootID)
	return result, nil
}

func (s *Store) collectConnectionIDsInSubtreeLocked(rootID string) ([]string, error) {
	groupIDs, err := s.collectGroupSubtreeIDsLocked(rootID)
	if err != nil {
		return nil, err
	}
	members, err := s.loadGroupMembersLocked()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	var connIDs []string
	for _, gid := range groupIDs {
		for _, m := range members[gid] {
			if m.MemberType == "connection" {
				if _, ok := seen[m.MemberID]; !ok {
					seen[m.MemberID] = struct{}{}
					connIDs = append(connIDs, m.MemberID)
				}
			}
		}
	}
	return connIDs, nil
}

func (s *Store) DeleteGroup(req model.DeleteGroupRequest, removeConn func(id string) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	connIDs, err := s.collectConnectionIDsInSubtreeLocked(req.ID)
	if err != nil {
		return err
	}
	groupIDs, err := s.collectGroupSubtreeIDsLocked(req.ID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	if req.DeleteConnections {
		if err := tx.Commit(); err != nil {
			return model.ErrInternal(err.Error())
		}
		s.mu.Unlock()
		for _, cid := range connIDs {
			if err := removeConn(cid); err != nil {
				return err
			}
		}
		s.mu.Lock()
		tx, err = s.db.Begin()
		if err != nil {
			return model.ErrInternal(err.Error())
		}
		defer func() { _ = tx.Rollback() }()
	} else {
		for i, cid := range connIDs {
			if _, err := tx.Exec(`DELETE FROM group_members WHERE member_type = 'connection' AND member_id = ?`, cid); err != nil {
				return model.ErrInternal(err.Error())
			}
			var max sql.NullInt64
			_ = tx.QueryRow(`SELECT MAX(sort_order) FROM free_connections`).Scan(&max)
			sortOrder := int(max.Int64) + 1 + i
			if _, err := tx.Exec(`INSERT OR IGNORE INTO free_connections (connection_id, sort_order) VALUES (?, ?)`, cid, sortOrder); err != nil {
				return model.ErrInternal(err.Error())
			}
			if err := insertSidebarRootItemTx(tx, "connection", cid, sortOrder); err != nil {
				return err
			}
		}
	}

	for _, gid := range groupIDs {
		if _, err := tx.Exec(`DELETE FROM group_members WHERE member_type = 'group' AND member_id = ?`, gid); err != nil {
			return model.ErrInternal(err.Error())
		}
		if _, err := tx.Exec(`DELETE FROM sidebar_root_items WHERE item_type = 'group' AND item_id = ?`, gid); err != nil {
			return model.ErrInternal(err.Error())
		}
	}
	placeholders := make([]string, len(groupIDs))
	args := make([]any, len(groupIDs))
	for i, gid := range groupIDs {
		placeholders[i] = "?"
		args[i] = gid
	}
	if len(groupIDs) > 0 {
		q := fmt.Sprintf(`DELETE FROM connection_groups WHERE id IN (%s)`, joinPlaceholders(len(groupIDs)))
		if _, err := tx.Exec(q, args...); err != nil {
			return model.ErrInternal(err.Error())
		}
	}
	return tx.Commit()
}

func joinPlaceholders(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			s += ","
		}
		s += "?"
	}
	return s
}

func (s *Store) MoveGroup(req model.MoveGroupRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureNotDescendantLocked(req.ID, req.NewParentID); err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	// Remove from old parent group_members
	if _, err := tx.Exec(`DELETE FROM group_members WHERE member_type = 'group' AND member_id = ?`, req.ID); err != nil {
		return model.ErrInternal(err.Error())
	}

	var parentVal sql.NullString
	if req.NewParentID != nil && *req.NewParentID != "" {
		parentVal = sql.NullString{String: *req.NewParentID, Valid: true}
		if _, err := tx.Exec(`INSERT INTO group_members (group_id, member_type, member_id, sort_order) VALUES (?, 'group', ?, ?)`,
			*req.NewParentID, req.ID, req.SortOrder); err != nil {
			return model.ErrInternal(err.Error())
		}
		_, _ = tx.Exec(`DELETE FROM sidebar_root_items WHERE item_type = 'group' AND item_id = ?`, req.ID)
	} else {
		if err := insertSidebarRootItemTx(tx, "group", req.ID, req.SortOrder); err != nil {
			return err
		}
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = tx.Exec(`UPDATE connection_groups SET parent_id = ?, sort_order = ?, updated_at = ? WHERE id = ?`,
		parentVal, req.SortOrder, now, req.ID)
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	return tx.Commit()
}

func (s *Store) ensureNotDescendantLocked(groupID string, newParentID *string) error {
	if newParentID == nil || *newParentID == "" {
		return nil
	}
	if *newParentID == groupID {
		return model.ErrInvalidRequest("cannot move group into itself")
	}
	descendants, err := s.collectGroupSubtreeIDsLocked(groupID)
	if err != nil {
		return err
	}
	for _, d := range descendants {
		if d == *newParentID {
			return model.ErrInvalidRequest("cannot move group into its descendant")
		}
	}
	return nil
}

func (s *Store) MoveConnectionToGroup(connectionID, groupID string, sortOrder int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM group_members WHERE member_type = 'connection' AND member_id = ?`, connectionID); err != nil {
		return model.ErrInternal(err.Error())
	}
	if _, err := tx.Exec(`DELETE FROM free_connections WHERE connection_id = ?`, connectionID); err != nil {
		return model.ErrInternal(err.Error())
	}
	if _, err := tx.Exec(`DELETE FROM sidebar_root_items WHERE item_type = 'connection' AND item_id = ?`, connectionID); err != nil {
		return model.ErrInternal(err.Error())
	}
	if sortOrder < 0 {
		var max sql.NullInt64
		_ = tx.QueryRow(`SELECT MAX(sort_order) FROM group_members WHERE group_id = ?`, groupID).Scan(&max)
		sortOrder = int(max.Int64) + 1
	}
	_, err = tx.Exec(`INSERT INTO group_members (group_id, member_type, member_id, sort_order) VALUES (?, 'connection', ?, ?)`,
		groupID, connectionID, sortOrder)
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	return tx.Commit()
}

func (s *Store) ReleaseConnection(connectionID string, sortOrder int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM group_members WHERE member_type = 'connection' AND member_id = ?`, connectionID); err != nil {
		return model.ErrInternal(err.Error())
	}
	if sortOrder < 0 {
		var max sql.NullInt64
		_ = tx.QueryRow(`SELECT MAX(sort_order) FROM free_connections`).Scan(&max)
		sortOrder = int(max.Int64) + 1
	}
	if _, err := tx.Exec(`INSERT OR REPLACE INTO free_connections (connection_id, sort_order) VALUES (?, ?)`, connectionID, sortOrder); err != nil {
		return model.ErrInternal(err.Error())
	}
	if err := insertSidebarRootItemTx(tx, "connection", connectionID, sortOrder); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ReorderGroupMembers(groupID string, ordered []model.GroupMemberRef) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	defer func() { _ = tx.Rollback() }()
	for i, ref := range ordered {
		_, err := tx.Exec(`UPDATE group_members SET sort_order = ? WHERE group_id = ? AND member_type = ? AND member_id = ?`,
			i, groupID, ref.MemberType, ref.MemberID)
		if err != nil {
			return model.ErrInternal(err.Error())
		}
	}
	return tx.Commit()
}

func (s *Store) ReorderSidebarRoot(ordered []model.SidebarRootItemRef) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	defer func() { _ = tx.Rollback() }()
	for i, ref := range ordered {
		_, err := tx.Exec(`
			INSERT INTO sidebar_root_items (item_type, item_id, sort_order) VALUES (?, ?, ?)
			ON CONFLICT(item_type, item_id) DO UPDATE SET sort_order = excluded.sort_order`,
			ref.ItemType, ref.ItemID, i)
		if err != nil {
			return model.ErrInternal(err.Error())
		}
	}
	return tx.Commit()
}
