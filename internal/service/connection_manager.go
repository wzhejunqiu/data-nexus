package service

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/driver"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

type Session struct {
	Connection *model.Connection
	Driver     driver.Driver
}

type ConnectionManager struct {
	mu     sync.RWMutex
	store  *ConnectionStore
	active map[string]*Session
	log    *zap.Logger
}

func NewConnectionManager(store *ConnectionStore, log *zap.Logger) *ConnectionManager {
	return &ConnectionManager{
		store:  store,
		active: make(map[string]*Session),
		log:    log,
	}
}

func (m *ConnectionManager) ListConnections() *model.ConnectionListView {
	snap := m.store.Snapshot()
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]model.ConnectionListItem, 0, len(snap.Items))
	for _, saved := range snap.Items {
		item := model.ConnectionListItem{
			ID:         saved.ID,
			Name:       saved.Name,
			Type:       saved.Type,
			Config:     saved.Config,
			Status:     model.ConnectionStatusClosed,
			LastUsedAt: saved.LastUsedAt,
		}
		if sess, ok := m.active[saved.ID]; ok {
			item.Status = model.ConnectionStatusOpen
			t := sess.Connection.ConnectedAt
			item.ConnectedAt = &t
		}
		items = append(items, item)
	}
	return &model.ConnectionListView{Items: items}
}

func (m *ConnectionManager) CreateConnection(ctx context.Context, req model.ConnectRequest) (*model.SavedConnection, error) {
	req, err := normalizeConnectRequest(req)
	if err != nil {
		return nil, err
	}
	item, err := m.store.Upsert(req)
	if err != nil {
		return nil, err
	}
	if err := m.store.Save(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	return item, nil
}

func (m *ConnectionManager) OpenConnection(ctx context.Context, connectionID string) (*model.Connection, error) {
	m.mu.RLock()
	if _, ok := m.active[connectionID]; ok {
		m.mu.RUnlock()
		return nil, model.ErrConnectionAlreadyOpen(connectionID)
	}
	m.mu.RUnlock()

	saved, ok := m.store.FindByID(connectionID)
	if !ok {
		return nil, model.ErrSavedNotFound(connectionID)
	}

	if saved.Config.SQLite != nil {
		if _, err := os.Stat(saved.Config.SQLite.FilePath); err != nil {
			if os.IsNotExist(err) {
				return nil, model.ErrConnectionFailed("database file does not exist")
			}
			return nil, model.ErrConnectionFailed(err.Error())
		}
	}

	drv, err := driver.NewDriver(saved.Type)
	if err != nil {
		return nil, err
	}
	if err := drv.Connect(ctx, saved.Config); err != nil {
		return nil, err
	}

	conn := &model.Connection{
		ID:          saved.ID,
		Type:        saved.Type,
		DisplayName: saved.Name,
		Config:      saved.Config,
		ConnectedAt: time.Now().UTC(),
	}

	m.mu.Lock()
	m.active[connectionID] = &Session{Connection: conn, Driver: drv}
	m.mu.Unlock()

	_, _ = m.store.Upsert(model.ConnectRequest{
		FilePath: saved.Config.SQLite.FilePath,
		ReadOnly: saved.Config.SQLite.ReadOnly,
	})
	_ = m.store.Save()

	return conn, nil
}

func (m *ConnectionManager) OpenConnectionFromFile(ctx context.Context, req model.ConnectRequest) (*model.Connection, error) {
	req, err := normalizeConnectRequest(req)
	if err != nil {
		return nil, err
	}
	saved, err := m.CreateConnection(ctx, req)
	if err != nil {
		return nil, err
	}
	return m.OpenConnection(ctx, saved.ID)
}

func (m *ConnectionManager) CloseConnection(_ context.Context, connectionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	sess, ok := m.active[connectionID]
	if !ok {
		return model.ErrConnectionNotFound(connectionID)
	}
	if err := sess.Driver.Close(); err != nil {
		return model.ErrInternal(err.Error())
	}
	delete(m.active, connectionID)
	return nil
}

func (m *ConnectionManager) RemoveConnection(ctx context.Context, connectionID string) error {
	_ = m.CloseConnection(ctx, connectionID)
	if err := m.store.Remove(connectionID); err != nil {
		return err
	}
	return m.store.Save()
}

func (m *ConnectionManager) RenameConnection(connectionID, name string) (*model.SavedConnection, error) {
	item, err := m.store.Rename(connectionID, name)
	if err != nil {
		return nil, err
	}
	if err := m.store.Save(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	return item, nil
}

func (m *ConnectionManager) GetRestoreOpenOnStartup() bool {
	return m.store.RestoreOpenOnStartup()
}

func (m *ConnectionManager) SetRestoreOpenOnStartup(enabled bool) error {
	m.store.SetRestoreOpenOnStartup(enabled)
	return m.store.Save()
}

func (m *ConnectionManager) PersistOpenConnections() error {
	m.mu.RLock()
	ids := make([]string, 0, len(m.active))
	for id := range m.active {
		ids = append(ids, id)
	}
	m.mu.RUnlock()
	m.store.SetOpenConnectionIDs(ids)
	return m.store.Save()
}

func (m *ConnectionManager) RestoreConnectionsOnStartup(ctx context.Context) error {
	if !m.store.RestoreOpenOnStartup() {
		return nil
	}
	for _, id := range m.store.OpenConnectionIDs() {
		if _, err := m.OpenConnection(ctx, id); err != nil {
			m.log.Warn("failed to restore connection", zap.String("id", id), zap.Error(err))
		}
	}
	return nil
}

func (m *ConnectionManager) Driver(connectionID string) (driver.Driver, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, ok := m.active[connectionID]
	if !ok {
		return nil, model.ErrConnectionNotFound(connectionID)
	}
	return sess.Driver, nil
}

func (m *ConnectionManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, sess := range m.active {
		_ = sess.Driver.Close()
		delete(m.active, id)
	}
}

func normalizeConnectRequest(req model.ConnectRequest) (model.ConnectRequest, error) {
	if req.FilePath == "" {
		return req, model.ErrInvalidPath("file path is required")
	}
	if !filepath.IsAbs(req.FilePath) {
		abs, err := filepath.Abs(req.FilePath)
		if err != nil {
			return req, model.ErrInvalidPath(err.Error())
		}
		req.FilePath = abs
	}
	info, err := os.Stat(req.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return req, model.ErrInvalidPath("file does not exist")
		}
		return req, model.ErrInvalidPath(err.Error())
	}
	if info.IsDir() {
		return req, model.ErrInvalidPath("path is a directory")
	}
	return req, nil
}
