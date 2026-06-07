package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/csvutil"
	"github.com/wzhejunqiu/data-nexus/internal/model"
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

	records, headers, err := readAllCSV(req.FilePath, req.Format)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &model.ImportCSVResult{}, nil
	}

	tableName := req.TargetTable
	if tableName == "" {
		tableName = req.NewTableName
	}
	if tableName == "" {
		return nil, model.ErrInvalidRequest("target table name is required")
	}
	if !identReImport.MatchString(tableName) {
		return nil, model.ErrInvalidRequest("invalid table name")
	}

	result := &model.ImportCSVResult{}
	if req.TargetTable == "" {
		if err := s.createTableFromCSV(ctx, drv, tableName, headers, records[0]); err != nil {
			return nil, err
		}
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
		if !identReImport.MatchString(dbCol) {
			return nil, model.ErrInvalidRequest("invalid mapped column: " + dbCol)
		}
		dbCols = append(dbCols, dbCol)
	}
	if len(dbCols) == 0 {
		return nil, model.ErrInvalidRequest("column map is empty")
	}

	if req.Mode == "append" {
		inserted, err := s.batchInsert(ctx, drv, tableName, headers, colMap, records, req.Format.Normalized().NullValue)
		if err != nil {
			return nil, err
		}
		result.RowsInserted = inserted
		return result, nil
	}

	upsertKeys := req.UpsertKeys
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

	inserted, updated, err := s.upsertRows(ctx, drv, tableName, headers, colMap, upsertKeys, records, req.Format.Normalized().NullValue)
	if err != nil {
		return nil, err
	}
	result.RowsInserted = inserted
	result.RowsUpdated = updated
	return result, nil
}

var identReImport = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func readAllCSV(path string, format model.CSVFormatOptions) ([][]string, []string, error) {
	data, err := csvutil.ReadFileBytes(path, format.Encoding)
	if err != nil {
		return nil, nil, err
	}
	reader, err := csvutil.NewReader(bytes.NewReader(data), format)
	if err != nil {
		return nil, nil, model.ErrInvalidRequest(err.Error())
	}
	var headers []string
	var records [][]string
	rowNum := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, model.ErrInvalidRequest("csv parse error: " + err.Error())
		}
		if rowNum == 0 && format.HasHeader {
			headers = append([]string{}, record...)
			rowNum++
			continue
		}
		if len(headers) == 0 {
			headers = make([]string, len(record))
			for i := range record {
				headers[i] = fmt.Sprintf("column_%d", i+1)
			}
		}
		records = append(records, record)
		rowNum++
	}
	return records, headers, nil
}

func (s *ImportService) createTableFromCSV(ctx context.Context, drv interface {
	Exec(ctx context.Context, sql string, params []any) (*model.ExecResult, error)
}, tableName string, headers []string, sample []string) error {
	cols := make([]string, len(headers))
	for i, h := range headers {
		colName := sanitizeColumnName(h)
		colType := inferColumnType(sample, i)
		cols[i] = fmt.Sprintf("%q %s", colName, colType)
	}
	sqlText := fmt.Sprintf("CREATE TABLE %q (%s)", tableName, strings.Join(cols, ", "))
	_, err := drv.Exec(ctx, sqlText, nil)
	return err
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

func (s *ImportService) batchInsert(ctx context.Context, drv interface {
	Exec(ctx context.Context, sql string, params []any) (*model.ExecResult, error)
}, table string, headers []string, colMap map[string]string, records [][]string, nullValue string) (int, error) {
	dbCols := mappedColumns(headers, colMap)
	if len(dbCols) == 0 {
		return 0, model.ErrInvalidRequest("no mapped columns")
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(dbCols)), ",")
	sqlText := fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s)", table, quoteJoin(dbCols), placeholders)
	inserted := 0
	for _, record := range records {
		argsOrdered := rowArgs(headers, colMap, dbCols, record, nullValue)
		if _, err := drv.Exec(ctx, sqlText, argsOrdered); err != nil {
			return inserted, err
		}
		inserted++
	}
	return inserted, nil
}

func (s *ImportService) upsertRows(ctx context.Context, drv interface {
	Exec(ctx context.Context, sql string, params []any) (*model.ExecResult, error)
	QueryRows(ctx context.Context, sql string, params []any, maxRows int) (*model.QueryResult, error)
}, table string, headers []string, colMap map[string]string, keys []string, records [][]string, nullValue string) (int, int, error) {
	dbCols := mappedColumns(headers, colMap)
	inserted := 0
	updated := 0
	for _, record := range records {
		argsOrdered := rowArgs(headers, colMap, dbCols, record, nullValue)
		keyArgs := make([]any, len(keys))
		where := make([]string, len(keys))
		for i, k := range keys {
			where[i] = fmt.Sprintf("%q = ?", k)
			keyArgs[i] = lookupMappedValue(k, headers, colMap, record, nullValue)
		}
		checkSQL := fmt.Sprintf("SELECT 1 FROM %q WHERE %s LIMIT 1", table, strings.Join(where, " AND "))
		res, err := drv.QueryRows(ctx, checkSQL, keyArgs, 1)
		if err != nil {
			return inserted, updated, err
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
				if _, err := drv.Exec(ctx, updateSQL, allArgs); err != nil {
					return inserted, updated, err
				}
				updated++
			}
		} else {
			placeholders := strings.TrimRight(strings.Repeat("?,", len(dbCols)), ",")
			insertSQL := fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s)", table, quoteJoin(dbCols), placeholders)
			if _, err := drv.Exec(ctx, insertSQL, argsOrdered); err != nil {
				return inserted, updated, err
			}
			inserted++
		}
	}
	return inserted, updated, nil
}

func mappedColumns(headers []string, colMap map[string]string) []string {
	seen := make(map[string]struct{})
	var cols []string
	for _, h := range headers {
		dc := colMap[h]
		if dc == "" {
			continue
		}
		if _, ok := seen[dc]; ok {
			continue
		}
		seen[dc] = struct{}{}
		cols = append(cols, dc)
	}
	return cols
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
