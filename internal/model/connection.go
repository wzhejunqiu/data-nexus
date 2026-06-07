package model

import "time"

type DriverType string

const DriverTypeSQLite DriverType = "sqlite"

type DriverConfig struct {
	Type   DriverType    `json:"type"`
	SQLite *SQLiteConfig `json:"sqlite,omitempty"`
}

type SQLiteConfig struct {
	FilePath string `json:"filePath"`
	ReadOnly bool   `json:"readOnly"`
}

type ConnectRequest struct {
	FilePath string `json:"filePath"`
	ReadOnly bool   `json:"readOnly"`
}

type Connection struct {
	ID          string       `json:"id"`
	Type        DriverType   `json:"type"`
	DisplayName string       `json:"displayName"`
	Config      DriverConfig `json:"config"`
	ConnectedAt time.Time    `json:"connectedAt"`
}

type SavedConnection struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Type       DriverType   `json:"type"`
	Config     DriverConfig `json:"config"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	LastUsedAt time.Time    `json:"lastUsedAt"`
}

type ConnectionStatus string

const (
	ConnectionStatusOpen   ConnectionStatus = "open"
	ConnectionStatusClosed ConnectionStatus = "closed"
)

type ConnectionListItem struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        DriverType       `json:"type"`
	Config      DriverConfig     `json:"config"`
	Status      ConnectionStatus `json:"status"`
	ConnectedAt *time.Time       `json:"connectedAt,omitempty"`
	LastUsedAt  time.Time        `json:"lastUsedAt"`
}

type ConnectionListView struct {
	Items []ConnectionListItem `json:"items"`
}

type ConnectionsFile struct {
	Version              int               `json:"version"`
	RestoreOpenOnStartup bool              `json:"restoreOpenOnStartup"`
	OpenConnectionIDs    []string          `json:"openConnectionIds"`
	Items                []SavedConnection `json:"items"`
}
