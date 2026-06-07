package service

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	for i := range s.data.Items {
		item := &s.data.Items[i]
		if item.Config.SQLite != nil && item.Config.SQLite.FilePath == filePath {
			return item, true
		}
	}
	return nil, false
}

func (s *ConnectionStore) Upsert(req model.ConnectRequest) (*model.SavedConnection, error) {
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
			},
		},
		CreatedAt:  now,
		UpdatedAt:  now,
		LastUsedAt: now,
	}
	s.data.Items = append(s.data.Items, item)
	return &item, nil
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

func (s *ConnectionStore) UpdateReadOnly(id string, readOnly bool) (*model.SavedConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Items {
		if s.data.Items[i].ID == id {
			if s.data.Items[i].Config.SQLite != nil {
				s.data.Items[i].Config.SQLite.ReadOnly = readOnly
			}
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
