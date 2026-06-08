package model

import "time"

const DefaultGroupName = "My Connections"

type ConnectionGroup struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parentId,omitempty"`
	SortOrder int       `json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ConnectionGroupNode struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	ChildGroups []ConnectionGroupNode `json:"childGroups"`
	Connections []ConnectionListItem  `json:"connections"`
	MemberItems []GroupMemberRef      `json:"memberItems,omitempty"`
}

type ConnectionSidebarTree struct {
	Groups          []ConnectionGroupNode `json:"groups"`
	FreeConnections []ConnectionListItem  `json:"freeConnections"`
	RootItems       []SidebarRootItemRef  `json:"rootItems,omitempty"`
}

type DeleteGroupRequest struct {
	ID                string `json:"id"`
	DeleteConnections bool   `json:"deleteConnections"`
}

type MoveGroupRequest struct {
	ID          string  `json:"id"`
	NewParentID *string `json:"newParentId,omitempty"`
	SortOrder   int     `json:"sortOrder"`
}

type GroupMemberRef struct {
	MemberType string `json:"memberType"` // "group" | "connection"
	MemberID   string `json:"memberId"`
}

type SidebarRootItemRef struct {
	ItemType string `json:"itemType"` // "group" | "connection"
	ItemID   string `json:"itemId"`
}
