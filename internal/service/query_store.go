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

type QueryStore struct {
	mu   sync.Mutex
	path string
	data model.CannedQueriesFile
}

func NewQueryStore(path string) (*QueryStore, error) {
	if path == "" {
		path = config.QueriesPath()
	}
	s := &QueryStore{
		path: path,
		data: model.CannedQueriesFile{Version: 1, Items: []model.CannedQuery{}},
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *QueryStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.data)
}

func (s *QueryStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *QueryStore) List() []model.CannedQuery {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]model.CannedQuery, len(s.data.Items))
	copy(out, s.data.Items)
	return out
}

func (s *QueryStore) SaveQuery(req model.SaveCannedQueryRequest) (*model.CannedQuery, error) {
	name := strings.TrimSpace(req.Name)
	sqlText := strings.TrimSpace(req.SQL)
	if name == "" {
		return nil, model.ErrInvalidRequest("name is required")
	}
	if sqlText == "" {
		return nil, model.ErrInvalidRequest("sql is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if req.ID != "" {
		for i := range s.data.Items {
			if s.data.Items[i].ID == req.ID {
				s.data.Items[i].Name = name
				s.data.Items[i].SQL = sqlText
				s.data.Items[i].ConnectionID = req.ConnectionID
				s.data.Items[i].Tags = req.Tags
				s.data.Items[i].UpdatedAt = now
				item := s.data.Items[i]
				return &item, nil
			}
		}
		return nil, model.ErrInvalidRequest("query not found")
	}
	item := model.CannedQuery{
		ID:           uuid.NewString(),
		Name:         name,
		SQL:          sqlText,
		ConnectionID: req.ConnectionID,
		Tags:         req.Tags,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.data.Items = append(s.data.Items, item)
	return &item, nil
}

func (s *QueryStore) Delete(id string) error {
	if id == "" {
		return model.ErrInvalidRequest("id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var next []model.CannedQuery
	found := false
	for _, item := range s.data.Items {
		if item.ID == id {
			found = true
			continue
		}
		next = append(next, item)
	}
	if !found {
		return model.ErrInvalidRequest("query not found")
	}
	s.data.Items = next
	return nil
}
