package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	_ "modernc.org/sqlite"
)

const schemaVersion = 1

const createTableSQL = `
CREATE TABLE IF NOT EXISTS sql_executions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    connection_id TEXT    NOT NULL,
    sql_text      TEXT    NOT NULL,
    kind          INTEGER NOT NULL,
    effect_rows   INTEGER NOT NULL,
    duration_ms   INTEGER NOT NULL,
    executed_at   TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sql_executions_conn_time
    ON sql_executions (connection_id, executed_at DESC);
`

type Store struct {
	db *sql.DB
}

func NewStore(cfg *model.ExecutionLogSQLiteConfig) (*Store, error) {
	path := ""
	if cfg != nil {
		path = cfg.FilePath
	}
	if path == "" {
		path = config.SqlGlobalDBPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, model.ErrInternal(err.Error())
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	db.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return model.ErrInternal(err.Error())
	}
	if version >= schemaVersion {
		return nil
	}
	if _, err := db.Exec(createTableSQL); err != nil {
		return model.ErrInternal(err.Error())
	}
	if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, schemaVersion)); err != nil {
		return model.ErrInternal(err.Error())
	}
	return nil
}

func (s *Store) Type() model.ExecutionLogDriverType {
	return model.ExecutionLogSQLite
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Insert(ctx context.Context, record model.SqlExecutionRecord) error {
	executedAt := record.ExecutedAt
	if executedAt.IsZero() {
		executedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sql_executions (
			connection_id, sql_text, kind, effect_rows, duration_ms, executed_at
		) VALUES (?, ?, ?, ?, ?, ?)`,
		record.ConnectionID,
		record.SQL,
		int(record.Kind),
		record.EffectRows,
		record.DurationMs,
		executedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return model.ErrInternal(err.Error())
	}
	return nil
}

func (s *Store) ListQueryHistory(ctx context.Context, connectionID string, limit int) ([]string, error) {
	if connectionID == "" {
		return nil, model.ErrInvalidRequest("connectionId is required")
	}
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT sql_text
		FROM sql_executions
		WHERE connection_id = ?
		GROUP BY sql_text
		ORDER BY MAX(executed_at) DESC
		LIMIT ?`, connectionID, limit)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var out []string
	for rows.Next() {
		var sqlText string
		if err := rows.Scan(&sqlText); err != nil {
			return nil, model.ErrInternal(err.Error())
		}
		out = append(out, sqlText)
	}
	if err := rows.Err(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func (s *Store) ListExecutions(ctx context.Context, connectionID string, limit int) ([]model.SqlExecutionRecord, error) {
	if connectionID == "" {
		return nil, model.ErrInvalidRequest("connectionId is required")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, connection_id, sql_text, kind, effect_rows, duration_ms, executed_at
		FROM sql_executions
		WHERE connection_id = ?
		ORDER BY executed_at DESC
		LIMIT ?`, connectionID, limit)
	if err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	defer func() { _ = rows.Close() }()

	var out []model.SqlExecutionRecord
	for rows.Next() {
		var rec model.SqlExecutionRecord
		var kind int
		var executedAt string
		if err := rows.Scan(
			&rec.ID,
			&rec.ConnectionID,
			&rec.SQL,
			&kind,
			&rec.EffectRows,
			&rec.DurationMs,
			&executedAt,
		); err != nil {
			return nil, model.ErrInternal(err.Error())
		}
		rec.Kind = model.SqlExecutionKind(kind)
		parsed, err := time.Parse(time.RFC3339, executedAt)
		if err != nil {
			return nil, model.ErrInternal(err.Error())
		}
		rec.ExecutedAt = parsed
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, model.ErrInternal(err.Error())
	}
	if out == nil {
		out = []model.SqlExecutionRecord{}
	}
	return out, nil
}
