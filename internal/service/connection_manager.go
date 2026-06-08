package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/driver"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/secrets"
	"go.uber.org/zap"
)

type Session struct {
	Connection *model.Connection
	Driver     driver.Driver
}

type ConnectionManager struct {
	mu      sync.RWMutex
	store   *ConnectionStore
	secrets secrets.Store
	active  map[string]*Session
	log     *zap.Logger
}

func NewConnectionManager(store *ConnectionStore, secretStore secrets.Store, log *zap.Logger) *ConnectionManager {
	return &ConnectionManager{
		store:   store,
		secrets: secretStore,
		active:  make(map[string]*Session),
		log:     log,
	}
}

func (m *ConnectionManager) SecretsStore() secrets.Store {
	return m.secrets
}

func (m *ConnectionManager) ListConnections() *model.ConnectionListView {
	snap := m.store.Snapshot()
	m.mu.RLock()
	defer m.mu.RUnlock()

	backend := model.SecretsBackend(m.secrets.ActiveBackend())
	items := make([]model.ConnectionListItem, 0, len(snap.Items))
	for _, saved := range snap.Items {
		item := model.ConnectionListItem{
			ID:             saved.ID,
			Name:           saved.Name,
			Type:           saved.Type,
			Config:         saved.Config,
			SecretsBackend: saved.SecretsBackend,
			Status:         model.ConnectionStatusClosed,
			LastUsedAt:     saved.LastUsedAt,
		}
		if item.SecretsBackend == "" && saved.Type != model.DriverTypeSQLite {
			item.SecretsBackend = backend
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
	item, err := m.store.UpsertSQLite(req)
	if err != nil {
		return nil, err
	}
	if err := m.store.Save(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	return item, nil
}

func (m *ConnectionManager) CreateRemoteConnection(ctx context.Context, req model.RemoteConnectRequest) (*model.SavedConnection, error) {
	if err := validateRemoteConnectRequest(req); err != nil {
		return nil, err
	}
	saved, err := m.store.UpsertRemote(req)
	if err != nil {
		return nil, err
	}
	saved.SecretsBackend = model.SecretsBackend(m.secrets.ActiveBackend())
	if err := m.secrets.SetPassword(saved.ID, req.Password); err != nil {
		return nil, err
	}
	if err := m.store.Save(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	if req.Open {
		if _, err := m.OpenConnection(ctx, saved.ID); err != nil {
			return saved, err
		}
	}
	return saved, nil
}

func (m *ConnectionManager) TestConnection(ctx context.Context, req model.TestConnectionRequest) error {
	cfg, err := driverConfigFromTestRequest(req)
	if err != nil {
		return err
	}
	drv, err := driver.NewDriver(cfg.Type)
	if err != nil {
		return err
	}
	defer func() { _ = drv.Close() }()
	if err := drv.Connect(ctx, cfg); err != nil {
		return mapConnectionError(err)
	}
	if err := drv.Ping(ctx); err != nil {
		return mapConnectionError(err)
	}
	return nil
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

	cfg, err := m.configForOpen(saved)
	if err != nil {
		return nil, err
	}

	drv, err := driver.NewDriver(saved.Type)
	if err != nil {
		return nil, err
	}
	if err := drv.Connect(ctx, cfg); err != nil {
		_ = drv.Close()
		return nil, mapConnectionError(err)
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

	m.touchLastUsed(saved.ID)
	return conn, nil
}

func (m *ConnectionManager) configForOpen(saved *model.SavedConnection) (model.DriverConfig, error) {
	cfg := saved.Config
	switch saved.Type {
	case model.DriverTypeSQLite:
		if cfg.SQLite == nil {
			return cfg, model.ErrInvalidRequest("sqlite config required")
		}
		if _, err := os.Stat(cfg.SQLite.FilePath); err != nil {
			if os.IsNotExist(err) {
				return cfg, model.ErrConnectionFailed("database file does not exist")
			}
			return cfg, model.ErrConnectionFailed(err.Error())
		}
	case model.DriverTypePostgres, model.DriverTypeMySQL:
		pw, err := m.secrets.GetPassword(saved.ID)
		if err != nil {
			return cfg, err
		}
		if cfg.Postgres != nil {
			cfg.Postgres.Password = pw
		}
		if cfg.MySQL != nil {
			cfg.MySQL.Password = pw
		}
	default:
		return cfg, model.ErrInvalidRequest("unsupported driver type")
	}
	return cfg, nil
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
	saved, ok := m.store.FindByID(connectionID)
	if !ok {
		return model.ErrSavedNotFound(connectionID)
	}
	_ = m.CloseConnection(ctx, connectionID)
	if saved.Type != model.DriverTypeSQLite {
		if err := m.secrets.DeletePassword(connectionID); err != nil {
			return err
		}
	}
	if err := m.store.Remove(connectionID); err != nil {
		return err
	}
	return m.store.Save()
}

func (m *ConnectionManager) UpdateConnectionSQLiteSettings(ctx context.Context, connectionID string, update model.SQLiteSettingsUpdate) (*model.SavedConnection, error) {
	if err := m.ensureConnectionClosed(connectionID); err != nil {
		return nil, err
	}
	if update.ReadOnly {
		update.WAL = false
	}
	item, err := m.store.UpdateSQLiteSettings(connectionID, update)
	if err != nil {
		return nil, err
	}
	if err := m.store.Save(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	return item, nil
}

func (m *ConnectionManager) UpdateConnectionPostgresSettings(ctx context.Context, connectionID string, update model.PostgresSettingsUpdate) (*model.SavedConnection, error) {
	if err := m.ensureConnectionClosed(connectionID); err != nil {
		return nil, err
	}
	item, err := m.store.UpdatePostgresSettings(connectionID, update)
	if err != nil {
		return nil, err
	}
	if update.Password != nil && *update.Password != "" {
		if err := m.secrets.SetPassword(connectionID, *update.Password); err != nil {
			return nil, err
		}
	}
	if err := m.store.Save(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	return item, nil
}

func (m *ConnectionManager) UpdateConnectionMySQLSettings(ctx context.Context, connectionID string, update model.MySQLSettingsUpdate) (*model.SavedConnection, error) {
	if err := m.ensureConnectionClosed(connectionID); err != nil {
		return nil, err
	}
	item, err := m.store.UpdateMySQLSettings(connectionID, update)
	if err != nil {
		return nil, err
	}
	if update.Password != nil && *update.Password != "" {
		if err := m.secrets.SetPassword(connectionID, *update.Password); err != nil {
			return nil, err
		}
	}
	if err := m.store.Save(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	return item, nil
}

func (m *ConnectionManager) ensureConnectionClosed(connectionID string) error {
	m.mu.RLock()
	_, isOpen := m.active[connectionID]
	m.mu.RUnlock()
	if isOpen {
		return model.ErrConnectionOpen(connectionID)
	}
	return nil
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
		saved, ok := m.store.FindByID(id)
		if !ok {
			continue
		}
		if saved.Type != model.DriverTypeSQLite && !m.secrets.VaultUnlocked() && m.secrets.ActiveBackend() == secrets.BackendVault {
			m.log.Debug("skip restore remote connection, vault locked", zap.String("id", id))
			continue
		}
		if _, err := m.OpenConnection(ctx, id); err != nil {
			m.log.Warn("failed to restore connection", zap.String("id", id), zap.Error(err))
		}
	}
	return nil
}

func (m *ConnectionManager) AttachDatabase(ctx context.Context, connectionID, filePath, alias string) error {
	if connectionID == "" {
		return model.ErrInvalidRequest("connectionId is required")
	}
	if filePath == "" {
		return model.ErrInvalidRequest("filePath is required")
	}
	if alias == "" {
		return model.ErrInvalidRequest("alias is required")
	}
	saved, ok := m.store.FindByID(connectionID)
	if !ok || saved.Type != model.DriverTypeSQLite {
		return model.ErrInvalidRequest("attach is only supported for sqlite connections")
	}
	req, err := normalizeConnectRequest(model.ConnectRequest{FilePath: filePath})
	if err != nil {
		return err
	}
	drv, err := m.Driver(connectionID)
	if err != nil {
		return err
	}
	return drv.Attach(ctx, req.FilePath, alias)
}

func (m *ConnectionManager) DetachDatabase(ctx context.Context, connectionID, alias string) error {
	if connectionID == "" {
		return model.ErrInvalidRequest("connectionId is required")
	}
	if alias == "" {
		return model.ErrInvalidRequest("alias is required")
	}
	drv, err := m.Driver(connectionID)
	if err != nil {
		return err
	}
	return drv.Detach(ctx, alias)
}

func (m *ConnectionManager) ListAttachedDatabases(ctx context.Context, connectionID string) ([]model.AttachedDatabase, error) {
	if connectionID == "" {
		return nil, model.ErrInvalidRequest("connectionId is required")
	}
	drv, err := m.Driver(connectionID)
	if err != nil {
		return nil, err
	}
	return drv.ListAttached(ctx)
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

func (m *ConnectionManager) touchLastUsed(id string) {
	item, ok := m.store.FindByID(id)
	if !ok {
		return
	}
	item.LastUsedAt = time.Now().UTC()
	_ = m.store.Save()
}

func normalizeConnectRequest(req model.ConnectRequest) (model.ConnectRequest, error) {
	if req.ReadOnly {
		req.WAL = false
	}
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

func validateRemoteConnectRequest(req model.RemoteConnectRequest) error {
	switch req.Type {
	case model.DriverTypePostgres:
		if req.Postgres == nil {
			return model.ErrInvalidRequest("postgres config required")
		}
		if strings.TrimSpace(req.Postgres.Host) == "" {
			return model.ErrInvalidRequest("host is required")
		}
		if strings.TrimSpace(req.Postgres.Database) == "" {
			return model.ErrInvalidRequest("database is required")
		}
		if strings.TrimSpace(req.Postgres.User) == "" {
			return model.ErrInvalidRequest("user is required")
		}
	case model.DriverTypeMySQL:
		if req.MySQL == nil {
			return model.ErrInvalidRequest("mysql config required")
		}
		if strings.TrimSpace(req.MySQL.Host) == "" {
			return model.ErrInvalidRequest("host is required")
		}
		if strings.TrimSpace(req.MySQL.Database) == "" {
			return model.ErrInvalidRequest("database is required")
		}
		if strings.TrimSpace(req.MySQL.User) == "" {
			return model.ErrInvalidRequest("user is required")
		}
	default:
		return model.ErrInvalidRequest("unsupported remote driver type")
	}
	return nil
}

func driverConfigFromTestRequest(req model.TestConnectionRequest) (model.DriverConfig, error) {
	cfg := model.DriverConfig{Type: req.Type}
	switch req.Type {
	case model.DriverTypePostgres:
		if req.Postgres == nil {
			return cfg, model.ErrInvalidRequest("postgres config required")
		}
		p := *req.Postgres
		p.Password = req.Password
		cfg.Postgres = &p
	case model.DriverTypeMySQL:
		if req.MySQL == nil {
			return cfg, model.ErrInvalidRequest("mysql config required")
		}
		p := *req.MySQL
		p.Password = req.Password
		cfg.MySQL = &p
	default:
		return cfg, model.ErrInvalidRequest("unsupported driver type for test")
	}
	if err := validateRemoteConnectRequest(model.RemoteConnectRequest{Type: req.Type, Postgres: cfg.Postgres, MySQL: cfg.MySQL}); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func mapConnectionError(err error) error {
	if appErr, ok := err.(*model.AppError); ok {
		return appErr
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "password authentication failed"),
		strings.Contains(msg, "access denied"),
		strings.Contains(msg, "authentication failed"):
		return model.ErrConnectionFailedWithReason(err.Error(), "auth")
	case strings.Contains(msg, "does not exist"),
		strings.Contains(msg, "unknown database"):
		return model.ErrConnectionFailedWithReason(err.Error(), "database")
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "timeout"),
		strings.Contains(msg, "no such host"),
		strings.Contains(msg, "network"):
		return model.ErrConnectionFailedWithReason(err.Error(), "network")
	default:
		return model.ErrConnectionFailed(err.Error())
	}
}

func NewDefaultConnectionManager(log *zap.Logger) (*ConnectionManager, error) {
	store, err := NewConnectionStore("")
	if err != nil {
		return nil, err
	}
	secretStore, err := secrets.NewStore(config.VaultDir())
	if err != nil {
		return nil, err
	}
	return NewConnectionManager(store, secretStore, log), nil
}
