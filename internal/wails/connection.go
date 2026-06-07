package wails

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type ConnectionService struct {
	mgr *service.ConnectionManager
	log *zap.Logger
}

func NewConnectionService(mgr *service.ConnectionManager, log *zap.Logger) *ConnectionService {
	return &ConnectionService{mgr: mgr, log: log}
}

func (s *ConnectionService) ListConnections() (*model.ConnectionListView, error) {
	return call(s.log, "ConnectionService.ListConnections", func() (*model.ConnectionListView, error) {
		return s.mgr.ListConnections(), nil
	})
}

func (s *ConnectionService) CreateConnection(req model.ConnectRequest) (*model.SavedConnection, error) {
	return call(s.log, "ConnectionService.CreateConnection", func() (*model.SavedConnection, error) {
		return s.mgr.CreateConnection(context.Background(), req)
	})
}

func (s *ConnectionService) OpenConnection(connectionID string) (*model.Connection, error) {
	return call(s.log, "ConnectionService.OpenConnection", func() (*model.Connection, error) {
		return s.mgr.OpenConnection(context.Background(), connectionID)
	})
}

func (s *ConnectionService) OpenConnectionFromFile(req model.ConnectRequest) (*model.Connection, error) {
	return call(s.log, "ConnectionService.OpenConnectionFromFile", func() (*model.Connection, error) {
		return s.mgr.OpenConnectionFromFile(context.Background(), req)
	})
}

func (s *ConnectionService) CloseConnection(connectionID string) error {
	return callVoid(s.log, "ConnectionService.CloseConnection", func() error {
		return s.mgr.CloseConnection(context.Background(), connectionID)
	})
}

func (s *ConnectionService) RemoveConnection(connectionID string) error {
	return callVoid(s.log, "ConnectionService.RemoveConnection", func() error {
		return s.mgr.RemoveConnection(context.Background(), connectionID)
	})
}

func (s *ConnectionService) RenameConnection(connectionID string, name string) (*model.SavedConnection, error) {
	return call(s.log, "ConnectionService.RenameConnection", func() (*model.SavedConnection, error) {
		return s.mgr.RenameConnection(connectionID, name)
	})
}

func (s *ConnectionService) UpdateConnectionSQLiteSettings(connectionID string, update model.SQLiteSettingsUpdate) (*model.SavedConnection, error) {
	return call(s.log, "ConnectionService.UpdateConnectionSQLiteSettings", func() (*model.SavedConnection, error) {
		return s.mgr.UpdateConnectionSQLiteSettings(context.Background(), connectionID, update)
	})
}

func (s *ConnectionService) GetRestoreOpenOnStartup() (bool, error) {
	return call(s.log, "ConnectionService.GetRestoreOpenOnStartup", func() (bool, error) {
		return s.mgr.GetRestoreOpenOnStartup(), nil
	})
}

func (s *ConnectionService) SetRestoreOpenOnStartup(enabled bool) error {
	return callVoid(s.log, "ConnectionService.SetRestoreOpenOnStartup", func() error {
		return s.mgr.SetRestoreOpenOnStartup(enabled)
	})
}

func (s *ConnectionService) Attach(connectionID string, filePath string, alias string) error {
	return callVoid(s.log, "ConnectionService.Attach", func() error {
		return s.mgr.AttachDatabase(context.Background(), connectionID, filePath, alias)
	})
}

func (s *ConnectionService) Detach(connectionID string, alias string) error {
	return callVoid(s.log, "ConnectionService.Detach", func() error {
		return s.mgr.DetachDatabase(context.Background(), connectionID, alias)
	})
}

func (s *ConnectionService) ListAttached(connectionID string) ([]model.AttachedDatabase, error) {
	return call(s.log, "ConnectionService.ListAttached", func() ([]model.AttachedDatabase, error) {
		return s.mgr.ListAttachedDatabases(context.Background(), connectionID)
	})
}
