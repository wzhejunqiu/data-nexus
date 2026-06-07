package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/csvutil"
	"github.com/wzhejunqiu/data-nexus/internal/model"
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
	columnNames := make([]string, len(columns))
	for i, c := range columns {
		columnNames[i] = c.Name
	}

	rowWriter, err := csvutil.NewRowWriter(w, columnNames, format)
	if err != nil {
		return model.ErrInternal(err.Error())
	}

	exported := 0
	emitProgress := func() {
		if onProgress != nil {
			onProgress(exported)
		}
	}

	for {
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return model.ErrExportCancelled()
			}
			return err
		}

		batch, err := cursor.NextBatch(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
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
