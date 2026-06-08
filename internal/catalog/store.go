package catalog

import (
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

// Store persists connections, groups, and app connection state in catalog.db.
type Store interface {
	Close() error

	// Connections
	ListConnections() ([]model.SavedConnection, error)
	FindByID(id string) (*model.SavedConnection, bool, error)
	FindByFilePath(filePath string) (*model.SavedConnection, bool, error)
	FindByFingerprint(fp string) (*model.SavedConnection, bool, error)
	UpsertConnection(item model.SavedConnection, placement ConnectionPlacement) error
	RemoveConnection(id string) error
	UpdateConnection(item model.SavedConnection) error

	// App state
	RestoreOpenOnStartup() (bool, error)
	SetRestoreOpenOnStartup(enabled bool) error
	OpenConnectionIDs() ([]string, error)
	SetOpenConnectionIDs(ids []string) error

	// Groups
	GetSidebarTree(openStatus map[string]model.ConnectionStatus, connectedAt map[string]*time.Time) (*model.ConnectionSidebarTree, error)
	CreateGroup(parentID *string, name string) (*model.ConnectionGroup, error)
	RenameGroup(id, name string) (*model.ConnectionGroup, error)
	CountConnectionsInGroup(id string) (int, error)
	DeleteGroup(req model.DeleteGroupRequest, removeConn func(id string) error) error
	MoveGroup(req model.MoveGroupRequest) error
	MoveConnectionToGroup(connectionID, groupID string, sortOrder int) error
	ReleaseConnection(connectionID string, sortOrder int) error
	ReorderGroupMembers(groupID string, ordered []model.GroupMemberRef) error
	ReorderSidebarRoot(ordered []model.SidebarRootItemRef) error
}

// ConnectionPlacement describes where a new/updated connection belongs in the sidebar.
type ConnectionPlacement struct {
	GroupID   *string // nil = free connection
	SortOrder int     // -1 = append at end
}
