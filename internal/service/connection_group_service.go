package service

import (
	"context"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/catalog"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type ConnectionGroupService struct {
	catalog catalog.Store
	mgr     *ConnectionManager
}

func NewConnectionGroupService(catalog catalog.Store, mgr *ConnectionManager) *ConnectionGroupService {
	return &ConnectionGroupService{catalog: catalog, mgr: mgr}
}

func (s *ConnectionGroupService) GetSidebarTree() (*model.ConnectionSidebarTree, error) {
	openStatus := make(map[string]model.ConnectionStatus)
	connectedAt := make(map[string]*time.Time)
	list := s.mgr.ListConnections()
	for _, item := range list.Items {
		openStatus[item.ID] = item.Status
		if item.ConnectedAt != nil {
			t := *item.ConnectedAt
			connectedAt[item.ID] = &t
		}
	}
	return s.catalog.GetSidebarTree(openStatus, connectedAt)
}

func (s *ConnectionGroupService) CreateGroup(parentID, name string) (*model.ConnectionGroup, error) {
	var pid *string
	if parentID != "" {
		pid = &parentID
	}
	return s.catalog.CreateGroup(pid, name)
}

func (s *ConnectionGroupService) RenameGroup(id, name string) (*model.ConnectionGroup, error) {
	return s.catalog.RenameGroup(id, name)
}

func (s *ConnectionGroupService) CountConnectionsInGroup(id string) (int, error) {
	return s.catalog.CountConnectionsInGroup(id)
}

func (s *ConnectionGroupService) GetGroupDeletePreview(id string) (*model.GroupDeletePreview, error) {
	return s.catalog.GetGroupDeletePreview(id)
}

func (s *ConnectionGroupService) DeleteGroup(req model.DeleteGroupRequest) error {
	return s.catalog.DeleteGroup(req, func(id string) error {
		return s.mgr.RemoveConnection(context.Background(), id)
	})
}

func (s *ConnectionGroupService) MoveGroup(req model.MoveGroupRequest) error {
	return s.catalog.MoveGroup(req)
}

func (s *ConnectionGroupService) MoveConnectionToGroup(connectionID, groupID string, sortOrder int) error {
	return s.catalog.MoveConnectionToGroup(connectionID, groupID, sortOrder)
}

func (s *ConnectionGroupService) ReleaseConnection(connectionID string, sortOrder int) error {
	return s.catalog.ReleaseConnection(connectionID, sortOrder)
}

func (s *ConnectionGroupService) ReorderGroupMembers(groupID string, ordered []model.GroupMemberRef) error {
	return s.catalog.ReorderGroupMembers(groupID, ordered)
}

func (s *ConnectionGroupService) ReorderSidebarRoot(ordered []model.SidebarRootItemRef) error {
	return s.catalog.ReorderSidebarRoot(ordered)
}
