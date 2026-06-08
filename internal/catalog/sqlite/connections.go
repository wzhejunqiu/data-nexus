package sqlite

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/wzhejunqiu/data-nexus/internal/catalog"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

const (
	keyRestoreOpenOnStartup = "restore_open_on_startup"
	keyOpenConnectionIDs    = "open_connection_ids"
)

func (s *Store) ListConnections() ([]model.SavedConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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

func (s *Store) FindByID(id string) (*model.SavedConnection, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row := s.db.QueryRow(`
		SELECT id, name, driver_type, config_json, secrets_backend, created_at, updated_at, last_used_at
		FROM connections WHERE id = ?`, id)
	item, err := scanSavedConnectionRow(row)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &item, true, nil
}

func (s *Store) FindByFilePath(filePath string) (*model.SavedConnection, bool, error) {
	items, err := s.ListConnections()
	if err != nil {
		return nil, false, err
	}
	for i := range items {
		if items[i].Config.SQLite != nil && items[i].Config.SQLite.FilePath == filePath {
			return &items[i], true, nil
		}
	}
	return nil, false, nil
}

func (s *Store) FindByFingerprint(fp string) (*model.SavedConnection, bool, error) {
	items, err := s.ListConnections()
	if err != nil {
		return nil, false, err
	}
	for i := range items {
		item := &items[i]
		var got string
		switch item.Type {
		case model.DriverTypePostgres:
			if item.Config.Postgres != nil {
				pg := item.Config.Postgres
				got = model.RemoteFingerprint(item.Type, pg.Host, pg.Port, pg.Database, pg.User)
			}
		case model.DriverTypeMySQL:
			if item.Config.MySQL != nil {
				my := item.Config.MySQL
				got = model.RemoteFingerprint(item.Type, my.Host, my.Port, my.Database, my.User)
			}
		}
		if got == fp {
			return item, true, nil
		}
	}
	return nil, false, nil
}

func (s *Store) UpsertConnection(item model.SavedConnection, placement catalog.ConnectionPlacement) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	configJSON, err := json.Marshal(item.Config)
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	secretsBackend := string(item.SecretsBackend)
	createdAt := item.CreatedAt.UTC().Format(time.RFC3339Nano)
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339Nano)
	lastUsedAt := item.LastUsedAt.UTC().Format(time.RFC3339Nano)

	var exists bool
	if err := tx.QueryRow(`SELECT 1 FROM connections WHERE id = ?`, item.ID).Scan(new(int)); err == nil {
		exists = true
	} else if err != sql.ErrNoRows {
		return model.ErrInternal(err.Error())
	}

	if exists {
		_, err = tx.Exec(`
			UPDATE connections SET name=?, driver_type=?, config_json=?, secrets_backend=?, updated_at=?, last_used_at=?
			WHERE id=?`,
			item.Name, string(item.Type), string(configJSON), nullString(secretsBackend),
			updatedAt, lastUsedAt, item.ID)
	} else {
		_, err = tx.Exec(`
			INSERT INTO connections (id, name, driver_type, config_json, secrets_backend, created_at, updated_at, last_used_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID, item.Name, string(item.Type), string(configJSON), nullString(secretsBackend),
			createdAt, updatedAt, lastUsedAt)
	}
	if err != nil {
		return model.ErrInternal(err.Error())
	}

	if !exists {
		if err := s.placeNewConnectionTx(tx, item.ID, placement); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) placeNewConnectionTx(tx *sql.Tx, connectionID string, placement catalog.ConnectionPlacement) error {
	if placement.GroupID != nil && *placement.GroupID != "" {
		sortOrder := placement.SortOrder
		if sortOrder < 0 {
			var max sql.NullInt64
			_ = tx.QueryRow(`SELECT MAX(sort_order) FROM group_members WHERE group_id = ?`, *placement.GroupID).Scan(&max)
			sortOrder = int(max.Int64) + 1
		}
		_, err := tx.Exec(`
			INSERT INTO group_members (group_id, member_type, member_id, sort_order)
			VALUES (?, 'connection', ?, ?)`,
			*placement.GroupID, connectionID, sortOrder)
		return err
	}
	sortOrder := placement.SortOrder
	if sortOrder < 0 {
		var max sql.NullInt64
		_ = tx.QueryRow(`SELECT MAX(sort_order) FROM free_connections`).Scan(&max)
		sortOrder = int(max.Int64) + 1
	}
	if _, err := tx.Exec(`INSERT INTO free_connections (connection_id, sort_order) VALUES (?, ?)`, connectionID, sortOrder); err != nil {
		return err
	}
	return s.insertSidebarRootTx(tx, "connection", connectionID, sortOrder)
}

func (s *Store) insertSidebarRootTx(tx *sql.Tx, itemType, itemID string, sortOrder int) error {
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

func (s *Store) UpdateConnection(item model.SavedConnection) error {
	return s.UpsertConnection(item, catalog.ConnectionPlacement{})
}

func (s *Store) RemoveConnection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM group_members WHERE member_type = 'connection' AND member_id = ?`, id); err != nil {
		return model.ErrInternal(err.Error())
	}
	if _, err := tx.Exec(`DELETE FROM free_connections WHERE connection_id = ?`, id); err != nil {
		return model.ErrInternal(err.Error())
	}
	if _, err := tx.Exec(`DELETE FROM sidebar_root_items WHERE item_type = 'connection' AND item_id = ?`, id); err != nil {
		return model.ErrInternal(err.Error())
	}
	res, err := tx.Exec(`DELETE FROM connections WHERE id = ?`, id)
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrSavedNotFound(id)
	}
	return tx.Commit()
}

func (s *Store) RestoreOpenOnStartup() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var val string
	err := s.db.QueryRow(`SELECT value FROM app_connection_state WHERE key = ?`, keyRestoreOpenOnStartup).Scan(&val)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, model.ErrInternal(err.Error())
	}
	return val == "true", nil
}

func (s *Store) SetRestoreOpenOnStartup(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := "false"
	if enabled {
		val = "true"
	}
	_, err := s.db.Exec(`
		INSERT INTO app_connection_state (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		keyRestoreOpenOnStartup, val)
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	return nil
}

func (s *Store) OpenConnectionIDs() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var val string
	err := s.db.QueryRow(`SELECT value FROM app_connection_state WHERE key = ?`, keyOpenConnectionIDs).Scan(&val)
	if err == sql.ErrNoRows {
		return []string{}, nil
	}
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	var ids []string
	if err := json.Unmarshal([]byte(val), &ids); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	return ids, nil
}

func (s *Store) SetOpenConnectionIDs(ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.Marshal(ids)
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	_, err = s.db.Exec(`
		INSERT INTO app_connection_state (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		keyOpenConnectionIDs, string(data))
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	return nil
}

func (s *Store) ensureDefaultGroup() error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM connection_groups WHERE parent_id IS NULL`).Scan(&count); err != nil {
		return model.ErrInternal(err.Error())
	}
	if count > 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id := uuid.NewString()
	_, err := s.db.Exec(`
		INSERT INTO connection_groups (id, name, parent_id, sort_order, created_at, updated_at)
		VALUES (?, ?, NULL, 0, ?, ?)`,
		id, model.DefaultGroupName, now, now)
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	_, err = s.db.Exec(`
		INSERT INTO sidebar_root_items (item_type, item_id, sort_order) VALUES ('group', ?, 0)`,
		id)
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanSavedConnection(rows scannable) (model.SavedConnection, error) {
	return scanSavedConnectionRow(rows)
}

func scanSavedConnectionRow(row scannable) (model.SavedConnection, error) {
	var item model.SavedConnection
	var driverType, configJSON string
	var secretsBackend sql.NullString
	var createdAt, updatedAt, lastUsedAt string
	if err := row.Scan(&item.ID, &item.Name, &driverType, &configJSON, &secretsBackend,
		&createdAt, &updatedAt, &lastUsedAt); err != nil {
		if err == sql.ErrNoRows {
			return item, err
		}
		return item, model.ErrInternal(err.Error())
	}
	item.Type = model.DriverType(driverType)
	if secretsBackend.Valid {
		item.SecretsBackend = model.SecretsBackend(secretsBackend.String)
	}
	if err := json.Unmarshal([]byte(configJSON), &item.Config); err != nil {
		return item, model.ErrInternal(err.Error())
	}
	var err error
	item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return item, model.ErrInternal(err.Error())
	}
	item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return item, model.ErrInternal(err.Error())
	}
	item.LastUsedAt, err = time.Parse(time.RFC3339Nano, lastUsedAt)
	if err != nil {
		return item, model.ErrInternal(err.Error())
	}
	return item, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
