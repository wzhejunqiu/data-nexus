package service

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wzhejunqiu/data-nexus/internal/catalog"
	catalogsqlite "github.com/wzhejunqiu/data-nexus/internal/catalog/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type ConnectionStore struct {
	catalog catalog.Store
}

func NewConnectionStore(path string) (*ConnectionStore, error) {
	if path == "" {
		path = config.CatalogDBPath()
	} else if strings.HasSuffix(strings.ToLower(path), ".json") {
		// Tests may pass legacy path; use catalog.db in same dir.
		path = filepath.Join(filepath.Dir(path), "catalog.db")
	}
	cat, err := catalogsqlite.NewStore(path)
	if err != nil {
		return nil, err
	}
	return &ConnectionStore{catalog: cat}, nil
}

func (s *ConnectionStore) Close() error {
	return s.catalog.Close()
}

func (s *ConnectionStore) Save() error {
	// SQLite auto-persists; kept for API compatibility.
	return nil
}

func (s *ConnectionStore) Snapshot() model.ConnectionsFile {
	items, err := s.catalog.ListConnections()
	if err != nil {
		return model.ConnectionsFile{Version: 1, OpenConnectionIDs: []string{}, Items: []model.SavedConnection{}}
	}
	restore, _ := s.catalog.RestoreOpenOnStartup()
	openIDs, _ := s.catalog.OpenConnectionIDs()
	return model.ConnectionsFile{
		Version:              1,
		RestoreOpenOnStartup: restore,
		OpenConnectionIDs:    openIDs,
		Items:                items,
	}
}

func (s *ConnectionStore) FindByID(id string) (*model.SavedConnection, bool) {
	item, ok, err := s.catalog.FindByID(id)
	if err != nil || !ok {
		return nil, false
	}
	return item, true
}

func (s *ConnectionStore) FindByFilePath(filePath string) (*model.SavedConnection, bool) {
	item, ok, err := s.catalog.FindByFilePath(filePath)
	if err != nil || !ok {
		return nil, false
	}
	return item, true
}

func (s *ConnectionStore) UpsertSQLite(req model.ConnectRequest) (*model.SavedConnection, error) {
	now := time.Now().UTC()
	name := filepath.Base(req.FilePath)

	if existing, ok, err := s.catalog.FindByFilePath(req.FilePath); err != nil {
		return nil, err
	} else if ok {
		existing.Config = model.DriverConfig{
			Type: model.DriverTypeSQLite,
			SQLite: &model.SQLiteConfig{
				FilePath: req.FilePath,
				ReadOnly: req.ReadOnly,
				WAL:      req.WAL,
			},
		}
		existing.UpdatedAt = now
		existing.LastUsedAt = now
		if err := s.catalog.UpdateConnection(*existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	item := model.SavedConnection{
		ID:   uuid.NewString(),
		Name: name,
		Type: model.DriverTypeSQLite,
		Config: model.DriverConfig{
			Type: model.DriverTypeSQLite,
			SQLite: &model.SQLiteConfig{
				FilePath: req.FilePath,
				ReadOnly: req.ReadOnly,
				WAL:      req.WAL,
			},
		},
		CreatedAt:  now,
		UpdatedAt:  now,
		LastUsedAt: now,
	}
	if err := s.catalog.UpsertConnection(item, catalog.ConnectionPlacement{SortOrder: -1}); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *ConnectionStore) UpsertSQLiteInGroup(req model.ConnectRequest, groupID *string) (*model.SavedConnection, error) {
	item, err := s.UpsertSQLite(req)
	if err != nil {
		return nil, err
	}
	if groupID != nil && *groupID != "" {
		if err := s.catalog.MoveConnectionToGroup(item.ID, *groupID, -1); err != nil {
			return nil, err
		}
	}
	return item, nil
}

func (s *ConnectionStore) UpsertRemote(req model.RemoteConnectRequest) (*model.SavedConnection, error) {
	now := time.Now().UTC()
	fp := remoteFingerprint(req)
	if existing, ok, err := s.catalog.FindByFingerprint(fp); err != nil {
		return nil, err
	} else if ok {
		s.applyRemoteConfig(existing, req)
		existing.UpdatedAt = now
		existing.LastUsedAt = now
		if err := s.catalog.UpdateConnection(*existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	item := model.SavedConnection{
		ID:         uuid.NewString(),
		Name:       remoteDisplayName(req),
		Type:       req.Type,
		CreatedAt:  now,
		UpdatedAt:  now,
		LastUsedAt: now,
	}
	s.applyRemoteConfig(&item, req)
	if err := s.catalog.UpsertConnection(item, catalog.ConnectionPlacement{SortOrder: -1}); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *ConnectionStore) UpsertRemoteInGroup(req model.RemoteConnectRequest, groupID *string) (*model.SavedConnection, error) {
	item, err := s.UpsertRemote(req)
	if err != nil {
		return nil, err
	}
	if groupID != nil && *groupID != "" {
		if err := s.catalog.MoveConnectionToGroup(item.ID, *groupID, -1); err != nil {
			return nil, err
		}
	}
	return item, nil
}

func (s *ConnectionStore) applyRemoteConfig(item *model.SavedConnection, req model.RemoteConnectRequest) {
	switch req.Type {
	case model.DriverTypePostgres:
		if req.Postgres != nil {
			pg := *req.Postgres
			pg.Password = ""
			item.Config = model.DriverConfig{Type: model.DriverTypePostgres, Postgres: &pg}
			item.Type = model.DriverTypePostgres
		}
	case model.DriverTypeMySQL:
		if req.MySQL != nil {
			my := *req.MySQL
			my.Password = ""
			item.Config = model.DriverConfig{Type: model.DriverTypeMySQL, MySQL: &my}
			item.Type = model.DriverTypeMySQL
		}
	}
	if name := remoteDisplayName(req); name != "" {
		item.Name = name
	}
}

func (s *ConnectionStore) Remove(id string) error {
	return s.catalog.RemoveConnection(id)
}

func (s *ConnectionStore) UpdateSQLiteSettings(id string, update model.SQLiteSettingsUpdate) (*model.SavedConnection, error) {
	item, ok, err := s.catalog.FindByID(id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, model.ErrSavedNotFound(id)
	}
	if item.Config.SQLite != nil {
		item.Config.SQLite.ReadOnly = update.ReadOnly
		item.Config.SQLite.WAL = update.WAL && !update.ReadOnly
	}
	item.UpdatedAt = time.Now().UTC()
	if err := s.catalog.UpdateConnection(*item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ConnectionStore) UpdatePostgresSettings(id string, update model.PostgresSettingsUpdate) (*model.SavedConnection, error) {
	item, ok, err := s.catalog.FindByID(id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, model.ErrSavedNotFound(id)
	}
	pg := item.Config.Postgres
	if pg == nil {
		pg = &model.PostgresConfig{}
	}
	if update.Host != "" {
		pg.Host = update.Host
	}
	if update.Port > 0 {
		pg.Port = update.Port
	}
	if update.Database != "" {
		pg.Database = update.Database
	}
	if update.User != "" {
		pg.User = update.User
	}
	if update.SSLMode != "" {
		pg.SSLMode = update.SSLMode
	}
	if update.Schema != "" {
		pg.Schema = update.Schema
	}
	pg.ReadOnly = update.ReadOnly
	if update.ClientEncoding != "" {
		pg.ClientEncoding = update.ClientEncoding
	}
	item.Config.Postgres = pg
	item.UpdatedAt = time.Now().UTC()
	if err := s.catalog.UpdateConnection(*item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ConnectionStore) UpdateMySQLSettings(id string, update model.MySQLSettingsUpdate) (*model.SavedConnection, error) {
	item, ok, err := s.catalog.FindByID(id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, model.ErrSavedNotFound(id)
	}
	my := item.Config.MySQL
	if my == nil {
		my = &model.MySQLConfig{}
	}
	if update.Host != "" {
		my.Host = update.Host
	}
	if update.Port > 0 {
		my.Port = update.Port
	}
	if update.Database != "" {
		my.Database = update.Database
	}
	if update.User != "" {
		my.User = update.User
	}
	my.TLS = update.TLS
	my.TLSSkipVerify = update.TLSSkipVerify
	my.ReadOnly = update.ReadOnly
	if update.Charset != "" {
		my.Charset = update.Charset
	}
	if update.Collation != "" {
		my.Collation = update.Collation
	}
	my.DefaultStorageEngine = update.DefaultStorageEngine
	item.Config.MySQL = my
	item.UpdatedAt = time.Now().UTC()
	if err := s.catalog.UpdateConnection(*item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ConnectionStore) Rename(id, name string) (*model.SavedConnection, error) {
	item, ok, err := s.catalog.FindByID(id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, model.ErrSavedNotFound(id)
	}
	item.Name = name
	item.UpdatedAt = time.Now().UTC()
	if err := s.catalog.UpdateConnection(*item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ConnectionStore) SetRestoreOpenOnStartup(enabled bool) {
	_ = s.catalog.SetRestoreOpenOnStartup(enabled)
}

func (s *ConnectionStore) SetOpenConnectionIDs(ids []string) {
	_ = s.catalog.SetOpenConnectionIDs(ids)
}

func (s *ConnectionStore) RestoreOpenOnStartup() bool {
	v, _ := s.catalog.RestoreOpenOnStartup()
	return v
}

func (s *ConnectionStore) OpenConnectionIDs() []string {
	ids, _ := s.catalog.OpenConnectionIDs()
	return ids
}

func (s *ConnectionStore) Catalog() catalog.Store {
	return s.catalog
}

func remoteFingerprint(req model.RemoteConnectRequest) string {
	switch req.Type {
	case model.DriverTypePostgres:
		if req.Postgres != nil {
			return model.RemoteFingerprint(req.Type, req.Postgres.Host, req.Postgres.Port, req.Postgres.Database, req.Postgres.User)
		}
	case model.DriverTypeMySQL:
		if req.MySQL != nil {
			return model.RemoteFingerprint(req.Type, req.MySQL.Host, req.MySQL.Port, req.MySQL.Database, req.MySQL.User)
		}
	}
	return ""
}

func remoteDisplayName(req model.RemoteConnectRequest) string {
	if name := strings.TrimSpace(req.Name); name != "" {
		return name
	}
	switch req.Type {
	case model.DriverTypePostgres:
		if req.Postgres != nil {
			return req.Postgres.User + "@" + req.Postgres.Host + "/" + req.Postgres.Database
		}
	case model.DriverTypeMySQL:
		if req.MySQL != nil {
			return req.MySQL.User + "@" + req.MySQL.Host + "/" + req.MySQL.Database
		}
	}
	return "remote"
}
