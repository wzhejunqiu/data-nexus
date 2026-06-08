package model

import (
	"strconv"
	"time"
)

type DriverType string

const (
	DriverTypeSQLite   DriverType = "sqlite"
	DriverTypePostgres DriverType = "postgres"
	DriverTypeMySQL    DriverType = "mysql"
)

type SecretsBackend string

const (
	SecretsBackendKeychain SecretsBackend = "keychain"
	SecretsBackendVault    SecretsBackend = "vault"
)

type DriverConfig struct {
	Type     DriverType      `json:"type"`
	SQLite   *SQLiteConfig   `json:"sqlite,omitempty"`
	Postgres *PostgresConfig `json:"postgres,omitempty"`
	MySQL    *MySQLConfig    `json:"mysql,omitempty"`
}

type SQLiteConfig struct {
	FilePath string `json:"filePath"`
	ReadOnly bool   `json:"readOnly"`
	WAL      bool   `json:"wal"`
}

type PostgresConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Database       string `json:"database"`
	User           string `json:"user"`
	Password       string `json:"-"`
	SSLMode        string `json:"sslMode"`
	Schema         string `json:"schema"`
	ClientEncoding string `json:"clientEncoding,omitempty"`
	ReadOnly       bool   `json:"readOnly"`
}

type MySQLConfig struct {
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	Database             string `json:"database"`
	User                 string `json:"user"`
	Password             string `json:"-"`
	TLS                  bool   `json:"tls"`
	TLSSkipVerify        bool   `json:"tlsSkipVerify,omitempty"`
	Charset              string `json:"charset,omitempty"`
	Collation            string `json:"collation,omitempty"`
	DefaultStorageEngine string `json:"defaultStorageEngine,omitempty"`
	ReadOnly             bool   `json:"readOnly"`
}

func (c *PostgresConfig) NormalizedPort() int {
	if c.Port <= 0 {
		return 5432
	}
	return c.Port
}

func (c *PostgresConfig) NormalizedSchema() string {
	if c.Schema == "" {
		return "public"
	}
	return c.Schema
}

func (c *PostgresConfig) NormalizedSSLMode() string {
	if c.SSLMode == "" {
		return "disable"
	}
	return c.SSLMode
}

func (c *PostgresConfig) NormalizedClientEncoding() string {
	if c.ClientEncoding == "" {
		return "UTF8"
	}
	return c.ClientEncoding
}

func (c *MySQLConfig) NormalizedPort() int {
	if c.Port <= 0 {
		return 3306
	}
	return c.Port
}

func (c *MySQLConfig) NormalizedCharset() string {
	if c.Charset == "" {
		return "utf8mb4"
	}
	return c.Charset
}

func (c *MySQLConfig) NormalizedCollation() string {
	if c.Collation == "" {
		return "utf8mb4_unicode_ci"
	}
	return c.Collation
}

type ConnectRequest struct {
	FilePath string `json:"filePath"`
	ReadOnly bool   `json:"readOnly"`
	WAL      bool   `json:"wal"`
}

type RemoteConnectRequest struct {
	Type     DriverType      `json:"type"`
	Name     string          `json:"name"`
	Password string          `json:"password"`
	Postgres *PostgresConfig `json:"postgres,omitempty"`
	MySQL    *MySQLConfig    `json:"mysql,omitempty"`
	Open     bool            `json:"open"`
}

type TestConnectionRequest struct {
	Type     DriverType      `json:"type"`
	Password string          `json:"password"`
	Postgres *PostgresConfig `json:"postgres,omitempty"`
	MySQL    *MySQLConfig    `json:"mysql,omitempty"`
}

type SQLiteSettingsUpdate struct {
	ReadOnly bool `json:"readOnly"`
	WAL      bool `json:"wal"`
}

type PostgresSettingsUpdate struct {
	Host           string  `json:"host"`
	Port           int     `json:"port"`
	Database       string  `json:"database"`
	User           string  `json:"user"`
	Password       *string `json:"password,omitempty"`
	SSLMode        string  `json:"sslMode"`
	Schema         string  `json:"schema"`
	ClientEncoding string  `json:"clientEncoding,omitempty"`
	ReadOnly       bool    `json:"readOnly"`
}

type MySQLSettingsUpdate struct {
	Host                 string  `json:"host"`
	Port                 int     `json:"port"`
	Database             string  `json:"database"`
	User                 string  `json:"user"`
	Password             *string `json:"password,omitempty"`
	TLS                  bool    `json:"tls"`
	TLSSkipVerify        bool    `json:"tlsSkipVerify"`
	Charset              string  `json:"charset,omitempty"`
	Collation            string  `json:"collation,omitempty"`
	DefaultStorageEngine string  `json:"defaultStorageEngine,omitempty"`
	ReadOnly             bool    `json:"readOnly"`
}

type Connection struct {
	ID          string       `json:"id"`
	Type        DriverType   `json:"type"`
	DisplayName string       `json:"displayName"`
	Config      DriverConfig `json:"config"`
	ConnectedAt time.Time    `json:"connectedAt"`
}

type SavedConnection struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Type           DriverType     `json:"type"`
	Config         DriverConfig   `json:"config"`
	SecretsBackend SecretsBackend `json:"secretsBackend,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	LastUsedAt     time.Time      `json:"lastUsedAt"`
}

type ConnectionStatus string

const (
	ConnectionStatusOpen   ConnectionStatus = "open"
	ConnectionStatusClosed ConnectionStatus = "closed"
)

type ConnectionListItem struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Type           DriverType       `json:"type"`
	Config         DriverConfig     `json:"config"`
	SecretsBackend SecretsBackend   `json:"secretsBackend,omitempty"`
	Status         ConnectionStatus `json:"status"`
	ConnectedAt    *time.Time       `json:"connectedAt,omitempty"`
	LastUsedAt     time.Time        `json:"lastUsedAt"`
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

type AttachedDatabase struct {
	Alias    string `json:"alias"`
	FilePath string `json:"filePath"`
}

func RemoteFingerprint(t DriverType, host string, port int, database, user string) string {
	if port <= 0 {
		switch t {
		case DriverTypePostgres:
			port = 5432
		case DriverTypeMySQL:
			port = 3306
		}
	}
	return string(t) + "|" + host + "|" + strconv.Itoa(port) + "|" + database + "|" + user
}
