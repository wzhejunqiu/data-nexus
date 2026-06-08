package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

	"github.com/wzhejunqiu/data-nexus/internal/config"
	_ "modernc.org/sqlite"
)

type Store struct {
	mu   sync.Mutex
	db   *sql.DB
	path string
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		path = config.CatalogDBPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, path: path}
	if err := migrateDB(s); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}
