package wails

import (
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type ConnectionGroupService struct {
	svc *service.ConnectionGroupService
	log *zap.Logger
}

func NewConnectionGroupService(svc *service.ConnectionGroupService, log *zap.Logger) *ConnectionGroupService {
	return &ConnectionGroupService{svc: svc, log: log}
}

func (s *ConnectionGroupService) GetSidebarTree() (*model.ConnectionSidebarTree, error) {
	return call(s.log, "ConnectionGroupService.GetSidebarTree", func() (*model.ConnectionSidebarTree, error) {
		return s.svc.GetSidebarTree()
	})
}

func (s *ConnectionGroupService) CreateGroup(parentID, name string) (*model.ConnectionGroup, error) {
	return call(s.log, "ConnectionGroupService.CreateGroup", func() (*model.ConnectionGroup, error) {
		return s.svc.CreateGroup(parentID, name)
	})
}

func (s *ConnectionGroupService) RenameGroup(id, name string) (*model.ConnectionGroup, error) {
	return call(s.log, "ConnectionGroupService.RenameGroup", func() (*model.ConnectionGroup, error) {
		return s.svc.RenameGroup(id, name)
	})
}

func (s *ConnectionGroupService) CountConnectionsInGroup(id string) (int, error) {
	return call(s.log, "ConnectionGroupService.CountConnectionsInGroup", func() (int, error) {
		return s.svc.CountConnectionsInGroup(id)
	})
}

func (s *ConnectionGroupService) GetGroupDeletePreview(id string) (*model.GroupDeletePreview, error) {
	return call(s.log, "ConnectionGroupService.GetGroupDeletePreview", func() (*model.GroupDeletePreview, error) {
		return s.svc.GetGroupDeletePreview(id)
	})
}

func (s *ConnectionGroupService) DeleteGroup(req model.DeleteGroupRequest) error {
	return callVoid(s.log, "ConnectionGroupService.DeleteGroup", func() error {
		return s.svc.DeleteGroup(req)
	})
}

func (s *ConnectionGroupService) MoveGroup(req model.MoveGroupRequest) error {
	return callVoid(s.log, "ConnectionGroupService.MoveGroup", func() error {
		return s.svc.MoveGroup(req)
	})
}

func (s *ConnectionGroupService) MoveConnectionToGroup(connectionID, groupID string, sortOrder int) error {
	return callVoid(s.log, "ConnectionGroupService.MoveConnectionToGroup", func() error {
		return s.svc.MoveConnectionToGroup(connectionID, groupID, sortOrder)
	})
}

func (s *ConnectionGroupService) ReleaseConnection(connectionID string, sortOrder int) error {
	return callVoid(s.log, "ConnectionGroupService.ReleaseConnection", func() error {
		return s.svc.ReleaseConnection(connectionID, sortOrder)
	})
}

func (s *ConnectionGroupService) ReorderGroupMembers(groupID string, ordered []model.GroupMemberRef) error {
	return callVoid(s.log, "ConnectionGroupService.ReorderGroupMembers", func() error {
		return s.svc.ReorderGroupMembers(groupID, ordered)
	})
}

func (s *ConnectionGroupService) ReorderSidebarRoot(ordered []model.SidebarRootItemRef) error {
	return callVoid(s.log, "ConnectionGroupService.ReorderSidebarRoot", func() error {
		return s.svc.ReorderSidebarRoot(ordered)
	})
}
