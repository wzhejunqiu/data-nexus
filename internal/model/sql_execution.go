package model

import (
	"encoding/json"
	"time"
)

type SqlExecutionKind int

const (
	SqlExecutionResult SqlExecutionKind = 0
	SqlExecutionExec   SqlExecutionKind = 1
)

func SqlExecutionKindFromResponse(kind string) SqlExecutionKind {
	if kind == "exec" {
		return SqlExecutionExec
	}
	return SqlExecutionResult
}

func (k SqlExecutionKind) String() string {
	if k == SqlExecutionExec {
		return "exec"
	}
	return "result"
}

func (k SqlExecutionKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *SqlExecutionKind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*k = SqlExecutionKindFromResponse(s)
	return nil
}

type SqlExecutionRecord struct {
	ID           int64            `json:"id,omitempty"`
	ConnectionID string           `json:"connectionId"`
	SQL          string           `json:"sql"`
	Kind         SqlExecutionKind `json:"kind"`
	EffectRows   int64            `json:"effectRows"`
	DurationMs   int64            `json:"durationMs"`
	ExecutedAt   time.Time        `json:"executedAt"`
}

type SqlExecutionList struct {
	Items []SqlExecutionRecord `json:"items"`
}

func NewSqlExecutionRecord(connectionID, sqlText string, resp *QueryResponse) SqlExecutionRecord {
	kind := SqlExecutionKindFromResponse(resp.Kind)
	var effectRows int64
	if kind == SqlExecutionExec {
		effectRows = resp.RowsAffected
	} else {
		effectRows = int64(resp.RowCount)
	}
	return SqlExecutionRecord{
		ConnectionID: connectionID,
		SQL:          sqlText,
		Kind:         kind,
		EffectRows:   effectRows,
		DurationMs:   resp.DurationMs,
		ExecutedAt:   time.Now().UTC(),
	}
}
