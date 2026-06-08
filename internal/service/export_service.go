package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/csvutil"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

type ExportService struct {
	query *QueryService
}

func NewExportService(query *QueryService) *ExportService {
	return &ExportService{query: query}
}

type ExportProgressFunc func(exported int)

const exportBatchSize = 1000

// ExportTableToFile streams table rows to path in batches without loading the full table into memory.
func (s *ExportService) ExportTableToFile(
	ctx context.Context,
	req model.ExportTableCSVRequest,
	path string,
	onProgress ExportProgressFunc,
) error {
	if req.ConnectionID == "" || req.TableName == "" {
		return model.ErrInvalidRequest("connectionId and tableName are required")
	}
	format := req.Format.Normalized()

	cursor, err := s.query.OpenTableExport(ctx, req.ConnectionID, req.TableName, model.TableExportOptions{
		BatchSize: exportBatchSize,
	})
	if err != nil {
		return err
	}
	defer func() { _ = cursor.Close() }()

	f, w, err := csvutil.OpenExportFile(path, format)
	if err != nil {
		return err
	}
	cleanPath := f.Name()
	success := false
	defer func() {
		_ = f.Close()
		if !success {
			_ = os.Remove(cleanPath)
		}
	}()

	columns := cursor.Columns()
	columnNames, err := resolveExportColumns(columns, req.Columns)
	if err != nil {
		return err
	}

	rowWriter, err := csvutil.NewRowWriter(w, columnNames, format)
	if err != nil {
		return model.ErrInternal(err.Error())
	}

	s.query.log.Debug("table export started",
		zap.String("connection_id", req.ConnectionID),
		zap.String("table", req.TableName),
		zap.String("path", path),
		zap.Int("column_count", len(columnNames)),
	)

	exported := 0
	emitProgress := func() {
		if onProgress != nil {
			onProgress(exported)
		}
	}

	for {
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				s.query.log.Debug("export canceled",
					zap.String("connection_id", req.ConnectionID),
					zap.String("table", req.TableName),
					zap.Int("rows_exported", exported),
				)
				return model.ErrExportCancelled()
			}
			return err
		}

		batch, err := cursor.NextBatch(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				s.query.log.Debug("export canceled",
					zap.String("connection_id", req.ConnectionID),
					zap.String("table", req.TableName),
					zap.Int("rows_exported", exported),
				)
				return model.ErrExportCancelled()
			}
			return err
		}
		if len(batch.Rows) > 0 {
			if err := rowWriter.WriteRows(batch.Rows); err != nil {
				return model.ErrInternal(err.Error())
			}
			exported += len(batch.Rows)
			emitProgress()
		}
		if !batch.HasMore {
			break
		}
	}

	if err := rowWriter.Close(); err != nil {
		return model.ErrInternal(err.Error())
	}
	success = true
	s.query.log.Debug("table export completed",
		zap.String("connection_id", req.ConnectionID),
		zap.String("table", req.TableName),
		zap.String("path", path),
		zap.String("format", format.Delimiter),
		zap.Int("rows_exported", exported),
	)
	return nil
}

func DefaultCSVFilename(tableName string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, tableName)
	return fmt.Sprintf("%s.csv", safe)
}

func resolveExportColumns(all []model.ColumnMeta, requested []string) ([]string, error) {
	if len(requested) == 0 {
		names := make([]string, len(all))
		for i, c := range all {
			names[i] = c.Name
		}
		return names, nil
	}
	available := make(map[string]struct{}, len(all))
	for _, c := range all {
		available[c.Name] = struct{}{}
	}
	out := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, name := range requested {
		if _, ok := available[name]; !ok {
			return nil, model.ErrInvalidRequest("invalid export column: " + name)
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil, model.ErrInvalidRequest("at least one column is required")
	}
	return out, nil
}
