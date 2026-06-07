package model

import "time"

type ExecuteQueryRequest struct {
	ConnectionID string `json:"connectionId"`
	SQL          string `json:"sql"`
	Params       []any  `json:"params"`
	MaxRows      int    `json:"maxRows"`
}

type QueryResult struct {
	Columns   []ColumnMeta     `json:"columns"`
	Rows      []map[string]any `json:"rows"`
	RowCount  int              `json:"rowCount"`
	Truncated bool             `json:"truncated"`
	Duration  time.Duration    `json:"-"`
}

type ExecResult struct {
	RowsAffected int64         `json:"rowsAffected"`
	LastInsertID int64         `json:"lastInsertId"`
	Duration     time.Duration `json:"-"`
}

type QueryResponse struct {
	Kind         string           `json:"kind"`
	Columns      []ColumnMeta     `json:"columns,omitempty"`
	Rows         []map[string]any `json:"rows,omitempty"`
	RowCount     int              `json:"rowCount,omitempty"`
	Truncated    bool             `json:"truncated,omitempty"`
	RowsAffected int64            `json:"rowsAffected,omitempty"`
	LastInsertID int64            `json:"lastInsertId,omitempty"`
	DurationMs   int64            `json:"durationMs"`
}

type VersionInfo struct {
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Arch     string `json:"arch"`
}

const MaxQueryRows = 10000

type StatementKind string

const (
	StatementQuery StatementKind = "query"
	StatementWrite StatementKind = "write"
)
