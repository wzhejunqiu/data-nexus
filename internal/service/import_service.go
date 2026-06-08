package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/csvutil"
	"github.com/wzhejunqiu/data-nexus/internal/driver/sqlite"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/sqlutil"
	"go.uber.org/zap"
)

type ImportService struct {
	query *QueryService
}

func NewImportService(query *QueryService) *ImportService {
	return &ImportService{query: query}
}

func (s *ImportService) ParseCSVPreview(ctx context.Context, req model.ParseCSVPreviewRequest) (*model.CSVPreview, error) {
	_ = ctx
	if req.FilePath == "" {
		return nil, model.ErrInvalidRequest("filePath is required")
	}
	return csvutil.ReadPreview(req.FilePath, req.MaxRows, req.Format)
}

func (s *ImportService) ImportCSV(ctx context.Context, req model.ImportCSVRequest) (*model.ImportCSVResult, error) {
	if req.ConnectionID == "" || req.FilePath == "" {
		return nil, model.ErrInvalidRequest("connectionId and filePath are required")
	}
	if req.Mode != "append" && req.Mode != "update" {
		return nil, model.ErrInvalidRequest("mode must be append or update")
	}
	drv, err := s.query.mgr.Driver(req.ConnectionID)
	if err != nil {
		return nil, err
	}
	if drv.ReadOnly() {
		return nil, model.ErrReadOnly()
	}

	iter, err := csvutil.OpenIterator(req.FilePath, req.Format)
	if err != nil {
		return nil, err
	}
	defer func() { _ = iter.Close() }()

	headers := iter.Headers()
	firstBatch, err := iter.ReadBatch(1)
	if err == io.EOF {
		return &model.ImportCSVResult{}, nil
	}
	if err != nil {
		return nil, err
	}
	sample := firstBatch[0]

	tableName := req.TargetTable
	if tableName == "" {
		tableName = req.NewTableName
	}
	if tableName == "" {
		return nil, model.ErrInvalidRequest("target table name is required")
	}
	if !sqlutil.IsSafeQuotedIdentifier(tableName) {
		return nil, model.ErrInvalidRequest("invalid table name")
	}

	colMap := req.ColumnMap
	if len(colMap) == 0 {
		colMap = make(map[string]string, len(headers))
		for _, h := range headers {
			colMap[h] = sanitizeColumnName(h)
		}
	}

	dbCols := make([]string, 0, len(headers))
	for _, h := range headers {
		dbCol, ok := colMap[h]
		if !ok || dbCol == "" {
			continue
		}
		if !sqlutil.IsSafeQuotedIdentifier(dbCol) {
			return nil, model.ErrInvalidRequest("invalid mapped column: " + dbCol)
		}
		dbCols = append(dbCols, dbCol)
	}
	if len(dbCols) == 0 {
		return nil, model.ErrInvalidRequest("column map is empty")
	}

	if req.TargetTable == "" {
		if err := s.createTableFromCSV(ctx, drv, tableName, headers, sample, colMap, req.NewTableColumns); err != nil {
			return nil, err
		}
	}

	nullValue := req.Format.Normalized().NullValue
	result := &model.ImportCSVResult{}

	var upsertKeys []string
	if req.Mode == "update" {
		upsertKeys = req.UpsertKeys
		if len(upsertKeys) == 0 {
			schema, err := drv.GetTableSchema(ctx, tableName)
			if err != nil {
				return nil, err
			}
			for _, c := range schema.Columns {
				if c.PrimaryKey {
					upsertKeys = append(upsertKeys, c.Name)
				}
			}
		}
		if len(upsertKeys) == 0 {
			return nil, model.ErrInvalidRequest("upsert keys required for update mode")
		}
	}

	sqliteDrv, ok := drv.(*sqlite.Driver)
	if !ok {
		return nil, model.ErrInvalidRequest("import not supported for this driver")
	}
	tx, err := sqliteDrv.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	importSuccess := false
	defer func() {
		if !importSuccess {
			s.query.log.Debug("csv import rolled back",
				zap.String("connection_id", req.ConnectionID),
				zap.String("table", tableName),
				zap.String("mode", req.Mode),
			)
		}
		_ = tx.Rollback()
	}()

	s.query.log.Debug("csv import started",
		zap.String("connection_id", req.ConnectionID),
		zap.String("table", tableName),
		zap.String("mode", req.Mode),
		zap.Int("column_count", len(dbCols)),
	)

	if req.Mode == "append" {
		inserted, err := s.importAppendBatched(ctx, tx, tableName, headers, colMap, dbCols, firstBatch, iter, nullValue)
		if err != nil {
			return nil, err
		}
		result.RowsInserted = inserted
	} else {
		inserted, updated, err := s.importUpdateBatched(ctx, tx, tableName, headers, colMap, dbCols, upsertKeys, firstBatch, iter, nullValue)
		if err != nil {
			return nil, err
		}
		result.RowsInserted = inserted
		result.RowsUpdated = updated
	}

	if err := tx.Commit(); err != nil {
		return nil, model.ErrSQL(err.Error())
	}
	importSuccess = true
	s.query.log.Debug("csv import completed",
		zap.String("connection_id", req.ConnectionID),
		zap.String("table", tableName),
		zap.String("mode", req.Mode),
		zap.Int("rows_inserted", result.RowsInserted),
		zap.Int("rows_updated", result.RowsUpdated),
	)
	return result, nil
}

const importBatchSize = 500
const sqliteMaxVars = 999

type importExec interface {
	Exec(ctx context.Context, sql string, params []any) (*model.ExecResult, error)
	QueryRows(ctx context.Context, sql string, params []any, maxRows int) (*model.QueryResult, error)
}

func (s *ImportService) importAppendBatched(
	ctx context.Context,
	tx importExec,
	table string,
	headers []string,
	colMap map[string]string,
	dbCols []string,
	firstBatch [][]string,
	iter *csvutil.Iterator,
	nullValue string,
) (int, error) {
	inserted := 0
	rowsPerStmt := importRowsPerStatement(len(dbCols))
	process := func(batch [][]string) error {
		for offset := 0; offset < len(batch); offset += rowsPerStmt {
			end := offset + rowsPerStmt
			if end > len(batch) {
				end = len(batch)
			}
			chunk := batch[offset:end]
			n, err := s.execMultiInsert(ctx, tx, table, headers, colMap, dbCols, chunk, nullValue)
			if err != nil {
				return err
			}
			inserted += n
		}
		return nil
	}
	if err := process(firstBatch); err != nil {
		return inserted, err
	}
	for {
		batch, err := iter.ReadBatch(importBatchSize)
		if err == io.EOF {
			break
		}
		if err != nil {
			return inserted, err
		}
		if err := process(batch); err != nil {
			return inserted, err
		}
	}
	return inserted, nil
}

func importRowsPerStatement(colCount int) int {
	if colCount <= 0 {
		return importBatchSize
	}
	maxRows := sqliteMaxVars / colCount
	if maxRows < 1 {
		maxRows = 1
	}
	if maxRows > importBatchSize {
		return importBatchSize
	}
	return maxRows
}

func (s *ImportService) execMultiInsert(
	ctx context.Context,
	tx importExec,
	table string,
	headers []string,
	colMap map[string]string,
	dbCols []string,
	records [][]string,
	nullValue string,
) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}
	placeholder := "(" + strings.TrimRight(strings.Repeat("?,", len(dbCols)), ",") + ")"
	placeholders := strings.TrimRight(strings.Repeat(placeholder+",", len(records)), ",")
	sqlText := fmt.Sprintf("INSERT INTO %q (%s) VALUES %s", table, quoteJoin(dbCols), placeholders)

	args := make([]any, 0, len(records)*len(dbCols))
	for _, record := range records {
		args = append(args, rowArgs(headers, colMap, dbCols, record, nullValue)...)
	}
	if _, err := tx.Exec(ctx, sqlText, args); err != nil {
		return 0, err
	}
	return len(records), nil
}

func (s *ImportService) importUpdateBatched(
	ctx context.Context,
	tx importExec,
	table string,
	headers []string,
	colMap map[string]string,
	dbCols []string,
	keys []string,
	firstBatch [][]string,
	iter *csvutil.Iterator,
	nullValue string,
) (int, int, error) {
	inserted := 0
	updated := 0
	process := func(batch [][]string) error {
		for _, record := range batch {
			ins, upd, err := s.upsertOneRow(ctx, tx, table, headers, colMap, dbCols, keys, record, nullValue)
			if err != nil {
				return err
			}
			inserted += ins
			updated += upd
		}
		return nil
	}
	if err := process(firstBatch); err != nil {
		return inserted, updated, err
	}
	for {
		batch, err := iter.ReadBatch(importBatchSize)
		if err == io.EOF {
			break
		}
		if err != nil {
			return inserted, updated, err
		}
		if err := process(batch); err != nil {
			return inserted, updated, err
		}
	}
	return inserted, updated, nil
}

func (s *ImportService) upsertOneRow(
	ctx context.Context,
	tx importExec,
	table string,
	headers []string,
	colMap map[string]string,
	dbCols []string,
	keys []string,
	record []string,
	nullValue string,
) (int, int, error) {
	argsOrdered := rowArgs(headers, colMap, dbCols, record, nullValue)
	keyArgs := make([]any, len(keys))
	where := make([]string, len(keys))
	for i, k := range keys {
		where[i] = fmt.Sprintf("%q = ?", k)
		keyArgs[i] = lookupMappedValue(k, headers, colMap, record, nullValue)
	}
	checkSQL := fmt.Sprintf("SELECT 1 FROM %q WHERE %s LIMIT 1", table, strings.Join(where, " AND "))
	res, err := tx.QueryRows(ctx, checkSQL, keyArgs, 1)
	if err != nil {
		return 0, 0, err
	}
	if res.RowCount > 0 {
		setParts := make([]string, 0, len(dbCols))
		setArgs := make([]any, 0, len(dbCols))
		for _, dc := range dbCols {
			isKey := false
			for _, k := range keys {
				if k == dc {
					isKey = true
					break
				}
			}
			if isKey {
				continue
			}
			setParts = append(setParts, fmt.Sprintf("%q = ?", dc))
			setArgs = append(setArgs, lookupMappedValue(dc, headers, colMap, record, nullValue))
		}
		if len(setParts) > 0 {
			updateSQL := fmt.Sprintf("UPDATE %q SET %s WHERE %s", table, strings.Join(setParts, ", "), strings.Join(where, " AND "))
			allArgs := append(setArgs, keyArgs...)
			if _, err := tx.Exec(ctx, updateSQL, allArgs); err != nil {
				return 0, 0, err
			}
			return 0, 1, nil
		}
		return 0, 0, nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(dbCols)), ",")
	insertSQL := fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s)", table, quoteJoin(dbCols), placeholders)
	if _, err := tx.Exec(ctx, insertSQL, argsOrdered); err != nil {
		return 0, 0, err
	}
	return 1, 0, nil
}

type importColEntry struct {
	name string
	typ  string
	pk   bool
}

func (s *ImportService) createTableFromCSV(ctx context.Context, drv interface {
	Exec(ctx context.Context, sql string, params []any) (*model.ExecResult, error)
}, tableName string, headers []string, sample []string, colMap map[string]string, specs map[string]model.ImportColumnSpec) error {
	var entries []importColEntry
	for i, h := range headers {
		colName := colMap[h]
		if colName == "" {
			continue
		}
		if !sqlutil.IsSafeQuotedIdentifier(colName) {
			return model.ErrInvalidRequest("invalid mapped column: " + colName)
		}
		var colType string
		isPK := false
		if spec, ok := specs[h]; ok {
			if spec.DataType != "" {
				colType = normalizeColumnType(spec.DataType)
			} else {
				colType = inferColumnType(sample, i)
			}
			isPK = spec.PrimaryKey
		} else {
			colType = inferColumnType(sample, i)
		}
		entries = append(entries, importColEntry{name: colName, typ: colType, pk: isPK})
	}
	if len(entries) == 0 {
		return model.ErrInvalidRequest("no columns for new table")
	}

	var pkCols []string
	for _, e := range entries {
		if e.pk {
			pkCols = append(pkCols, e.name)
		}
	}

	colDefs := make([]string, 0, len(entries)+1)
	for _, e := range entries {
		if len(pkCols) == 1 && e.pk {
			colDefs = append(colDefs, fmt.Sprintf("%q %s PRIMARY KEY", e.name, e.typ))
		} else {
			colDefs = append(colDefs, fmt.Sprintf("%q %s", e.name, e.typ))
		}
	}
	if len(pkCols) > 1 {
		colDefs = append(colDefs, fmt.Sprintf("PRIMARY KEY (%s)", quoteJoin(pkCols)))
	}

	sqlText := fmt.Sprintf("CREATE TABLE %q (%s)", tableName, strings.Join(colDefs, ", "))
	_, err := drv.Exec(ctx, sqlText, nil)
	return err
}

func normalizeColumnType(t string) string {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case "INTEGER", "REAL", "TEXT", "BLOB":
		return strings.ToUpper(strings.TrimSpace(t))
	default:
		return "TEXT"
	}
}

func inferColumnType(sample []string, idx int) string {
	if idx >= len(sample) {
		return "TEXT"
	}
	v := strings.TrimSpace(sample[idx])
	if v == "" {
		return "TEXT"
	}
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return "INTEGER"
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return "REAL"
	}
	return "TEXT"
}

func sanitizeColumnName(h string) string {
	s := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, h)
	if s == "" || (s[0] >= '0' && s[0] <= '9') {
		s = "col_" + s
	}
	return s
}

func quoteJoin(cols []string) string {
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = fmt.Sprintf("%q", c)
	}
	return strings.Join(parts, ", ")
}

func rowArgs(headers []string, colMap map[string]string, dbCols []string, record []string, nullValue string) []any {
	args := make([]any, len(dbCols))
	for i, dc := range dbCols {
		args[i] = lookupMappedValue(dc, headers, colMap, record, nullValue)
	}
	return args
}

func lookupMappedValue(dbCol string, headers []string, colMap map[string]string, record []string, nullValue string) any {
	for i, h := range headers {
		if colMap[h] == dbCol {
			if i < len(record) {
				return csvCellToSQL(record[i], nullValue)
			}
			return csvCellToSQL("", nullValue)
		}
	}
	return csvCellToSQL("", nullValue)
}

func csvCellToSQL(cell, nullValue string) any {
	if cell == nullValue {
		return nil
	}
	return cell
}
