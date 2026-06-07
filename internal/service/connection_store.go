package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type ConnectionStore struct {
	mu   sync.Mutex
	path string
	data model.ConnectionsFile
}

func NewConnectionStore(path string) (*ConnectionStore, error) {
	if path == "" {
		path = config.ConnectionsPath()
	}
	s := &ConnectionStore{
		path: path,
		data: model.ConnectionsFile{
			Version:           1,
			OpenConnectionIDs: []string{},
			Items:             []model.SavedConnection{},
		},
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *ConnectionStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.data)
}

func (s *ConnectionStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *ConnectionStore) Snapshot() model.ConnectionsFile {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

func (s *ConnectionStore) FindByID(id string) (*model.SavedConnection, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Items {
		if s.data.Items[i].ID == id {
			item := s.data.Items[i]
			return &item, true
		}
	}
	return nil, false
}

func (s *ConnectionStore) FindByFilePath(filePath string) (*model.SavedConnection, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.findByFilePathLocked(filePath)
}

func (s *ConnectionStore) UpsertSQLite(req model.ConnectRequest) (*model.SavedConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	name := filepath.Base(req.FilePath)

	if existing, ok := s.findByFilePathLocked(req.FilePath); ok {
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
	s.data.Items = append(s.data.Items, item)
	return &item, nil
}

func (s *ConnectionStore) UpsertRemote(req model.RemoteConnectRequest) (*model.SavedConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	fp := remoteFingerprint(req)
	if existing, ok := s.findByFingerprintLocked(fp); ok {
		s.applyRemoteConfig(existing, req)
		existing.UpdatedAt = now
		existing.LastUsedAt = now
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
	s.data.Items = append(s.data.Items, item)
	return &item, nil
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

func (s *ConnectionStore) findByFilePathLocked(filePath string) (*model.SavedConnection, bool) {
	for i := range s.data.Items {
		item := &s.data.Items[i]
		if item.Config.SQLite != nil && item.Config.SQLite.FilePath == filePath {
			return item, true
		}
	}
	return nil, false
}

func (s *ConnectionStore) findByFingerprintLocked(fp string) (*model.SavedConnection, bool) {
	for i := range s.data.Items {
		item := &s.data.Items[i]
		switch item.Type {
		case model.DriverTypePostgres:
			if item.Config.Postgres != nil {
				got := model.RemoteFingerprint(item.Type, item.Config.Postgres.Host, item.Config.Postgres.Port, item.Config.Postgres.Database, item.Config.Postgres.User)
				if got == fp {
					return item, true
				}
			}
		case model.DriverTypeMySQL:
			if item.Config.MySQL != nil {
				got := model.RemoteFingerprint(item.Type, item.Config.MySQL.Host, item.Config.MySQL.Port, item.Config.MySQL.Database, item.Config.MySQL.User)
				if got == fp {
					return item, true
				}
			}
		}
	}
	return nil, false
}

func (s *ConnectionStore) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, item := range s.data.Items {
		if item.ID == id {
			s.data.Items = append(s.data.Items[:i], s.data.Items[i+1:]...)
			return nil
		}
	}
	return model.ErrSavedNotFound(id)
}

func (s *ConnectionStore) UpdateSQLiteSettings(id string, update model.SQLiteSettingsUpdate) (*model.SavedConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Items {
		if s.data.Items[i].ID == id {
			if s.data.Items[i].Config.SQLite != nil {
				s.data.Items[i].Config.SQLite.ReadOnly = update.ReadOnly
				s.data.Items[i].Config.SQLite.WAL = update.WAL && !update.ReadOnly
			}
			s.data.Items[i].UpdatedAt = time.Now().UTC()
			item := s.data.Items[i]
			return &item, nil
		}
	}
	return nil, model.ErrSavedNotFound(id)
}

func (s *ConnectionStore) UpdatePostgresSettings(id string, update model.PostgresSettingsUpdate) (*model.SavedConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Items {
		if s.data.Items[i].ID == id {
			pg := s.data.Items[i].Config.Postgres
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
			s.data.Items[i].Config.Postgres = pg
			s.data.Items[i].UpdatedAt = time.Now().UTC()
			item := s.data.Items[i]
			return &item, nil
		}
	}
	return nil, model.ErrSavedNotFound(id)
}

func (s *ConnectionStore) UpdateMySQLSettings(id string, update model.MySQLSettingsUpdate) (*model.SavedConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Items {
		if s.data.Items[i].ID == id {
			my := s.data.Items[i].Config.MySQL
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
			s.data.Items[i].Config.MySQL = my
			s.data.Items[i].UpdatedAt = time.Now().UTC()
			item := s.data.Items[i]
			return &item, nil
		}
	}
	return nil, model.ErrSavedNotFound(id)
}

func (s *ConnectionStore) Rename(id, name string) (*model.SavedConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Items {
		if s.data.Items[i].ID == id {
			s.data.Items[i].Name = name
			s.data.Items[i].UpdatedAt = time.Now().UTC()
			item := s.data.Items[i]
			return &item, nil
		}
	}
	return nil, model.ErrSavedNotFound(id)
}

func (s *ConnectionStore) SetRestoreOpenOnStartup(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.RestoreOpenOnStartup = enabled
}

func (s *ConnectionStore) SetOpenConnectionIDs(ids []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.OpenConnectionIDs = ids
}

func (s *ConnectionStore) RestoreOpenOnStartup() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data.RestoreOpenOnStartup
}

func (s *ConnectionStore) OpenConnectionIDs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.data.OpenConnectionIDs...)
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
