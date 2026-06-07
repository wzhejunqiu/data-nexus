package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/driver/export"
	"github.com/wzhejunqiu/data-nexus/internal/executionlog"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

// describeSQLRe matches MySQL-style DESC/DESCRIBE with a single table identifier.
var describeSQLRe = regexp.MustCompile(`(?is)^\s*(?:DESC|DESCRIBE)\s+((?:` + "`" + `[^` + "`" + `]+` + "`" + `|\[[^\]]+\]|"[^"]+"|'[^']+'|[A-Za-z_][\w$#]*(?:\.[A-Za-z_][\w$#]*)*))\s*;?\s*$`)

func rewriteDescribeSQL(sql string) (string, bool) {
	m := describeSQLRe.FindStringSubmatch(sql)
	if m == nil {
		return sql, false
	}
	return fmt.Sprintf("PRAGMA table_info(%s)", m[1]), true
}

func normalizeSQL(sql string) string {
	if rewritten, ok := rewriteDescribeSQL(sql); ok {
		return rewritten
	}
	return sql
}

type QueryService struct {
	mgr     *ConnectionManager
	execLog executionlog.Store
	log     *zap.Logger
}

func NewQueryService(mgr *ConnectionManager, execLog executionlog.Store, log *zap.Logger) *QueryService {
	if log == nil {
		log = zap.NewNop()
	}
	return &QueryService{mgr: mgr, execLog: execLog, log: log}
}

func (s *QueryService) Execute(ctx context.Context, req model.ExecuteQueryRequest) (*model.QueryResponse, error) {
	if req.ConnectionID == "" {
		return nil, model.ErrInvalidRequest("connectionId is required")
	}
	if strings.TrimSpace(req.SQL) == "" {
		return nil, model.ErrInvalidRequest("sql is required")
	}
	drv, err := s.mgr.Driver(req.ConnectionID)
	if err != nil {
		return nil, err
	}
	maxRows := req.MaxRows
	if maxRows <= 0 {
		maxRows = model.MaxQueryRows
	}
	sqlText := normalizeSQL(req.SQL)
	kind, err := drv.ClassifySQL(ctx, sqlText)
	if err != nil {
		return nil, err
	}
	if kind == model.StatementWrite {
		if drv.ReadOnly() {
			return nil, model.ErrReadOnly()
		}
		execResult, err := drv.Exec(ctx, sqlText, req.Params)
		if err != nil {
			return nil, err
		}
		resp := &model.QueryResponse{
			Kind:         "exec",
			RowsAffected: execResult.RowsAffected,
			LastInsertID: execResult.LastInsertID,
			DurationMs:   execResult.Duration.Milliseconds(),
		}
		s.recordExecution(ctx, req.ConnectionID, sqlText, resp)
		return resp, nil
	}
	result, err := drv.QueryRows(ctx, sqlText, req.Params, maxRows)
	if err != nil {
		return nil, err
	}
	resp := &model.QueryResponse{
		Kind:       "result",
		Columns:    result.Columns,
		Rows:       result.Rows,
		RowCount:   result.RowCount,
		Truncated:  result.Truncated,
		DurationMs: result.Duration.Milliseconds(),
	}
	s.recordExecution(ctx, req.ConnectionID, sqlText, resp)
	return resp, nil
}

func (s *QueryService) recordExecution(ctx context.Context, connectionID, sqlText string, resp *model.QueryResponse) {
	if s.execLog == nil {
		return
	}
	record := model.NewSqlExecutionRecord(connectionID, sqlText, resp)
	if err := s.execLog.Insert(ctx, record); err != nil {
		s.log.Warn("execution log insert failed", zap.Error(err))
	}
}

func (s *QueryService) ClassifySQL(ctx context.Context, connectionID, sql string) (model.StatementKind, error) {
	if connectionID == "" {
		return "", model.ErrInvalidRequest("connectionId is required")
	}
	if strings.TrimSpace(sql) == "" {
		return "", model.ErrInvalidRequest("sql is required")
	}
	drv, err := s.mgr.Driver(connectionID)
	if err != nil {
		return "", err
	}
	return drv.ClassifySQL(ctx, normalizeSQL(sql))
}

func (s *QueryService) BrowseRows(ctx context.Context, req model.BrowseRowsRequest) (*model.PaginatedTableData, error) {
	if req.ConnectionID == "" {
		return nil, model.ErrInvalidRequest("connectionId is required")
	}
	drv, err := s.mgr.Driver(req.ConnectionID)
	if err != nil {
		return nil, err
	}
	opts := model.BrowseOptions{
		Page:           req.Page,
		PageSize:       req.PageSize,
		Sort:           req.Sort,
		Order:          req.Order,
		SkipTotalCount: req.SkipTotalCount,
		Filters:        req.Filters,
		Search:         req.Search,
	}
	if opts.Order == "" {
		opts.Order = model.SortAsc
	}
	return drv.BrowseTable(ctx, req.TableName, opts)
}

func (s *QueryService) OpenTableExport(ctx context.Context, connectionID, tableName string, opts model.TableExportOptions) (export.TableExportCursor, error) {
	if connectionID == "" {
		return nil, model.ErrInvalidRequest("connectionId is required")
	}
	if tableName == "" {
		return nil, model.ErrInvalidRequest("tableName is required")
	}
	drv, err := s.mgr.Driver(connectionID)
	if err != nil {
		return nil, err
	}
	return drv.OpenTableExport(ctx, tableName, opts)
}

func (s *QueryService) ListTables(ctx context.Context, connectionID string) (*model.TableList, error) {
	drv, err := s.mgr.Driver(connectionID)
	if err != nil {
		return nil, err
	}
	items, err := drv.ListTables(ctx)
	if err != nil {
		return nil, err
	}
	return &model.TableList{Items: items}, nil
}

func (s *QueryService) GetTableSchema(ctx context.Context, connectionID, tableName string) (*model.TableSchema, error) {
	drv, err := s.mgr.Driver(connectionID)
	if err != nil {
		return nil, err
	}
	return drv.GetTableSchema(ctx, tableName)
}

func (s *QueryService) GetTableProfile(ctx context.Context, connectionID, tableName string) (*model.TableProfile, error) {
	if connectionID == "" {
		return nil, model.ErrInvalidRequest("connectionId is required")
	}
	if tableName == "" {
		return nil, model.ErrInvalidRequest("tableName is required")
	}
	drv, err := s.mgr.Driver(connectionID)
	if err != nil {
		return nil, err
	}
	return drv.GetTableProfile(ctx, tableName)
}

func (s *QueryService) DetectFTSTable(ctx context.Context, connectionID, tableName string) (*model.FTSInfo, error) {
	if connectionID == "" {
		return nil, model.ErrInvalidRequest("connectionId is required")
	}
	drv, err := s.mgr.Driver(connectionID)
	if err != nil {
		return nil, err
	}
	return drv.DetectFTSTable(ctx, tableName)
}

func (s *QueryService) AttachDatabase(ctx context.Context, connectionID, filePath, alias string) error {
	return s.mgr.AttachDatabase(ctx, connectionID, filePath, alias)
}

func (s *QueryService) DetachDatabase(ctx context.Context, connectionID, alias string) error {
	return s.mgr.DetachDatabase(ctx, connectionID, alias)
}

func (s *QueryService) ListAttachedDatabases(ctx context.Context, connectionID string) ([]model.AttachedDatabase, error) {
	return s.mgr.ListAttachedDatabases(ctx, connectionID)
}

func (s *QueryService) UpdateCellsBatch(ctx context.Context, req model.UpdateCellsBatchRequest) (*model.UpdateCellsBatchResult, error) {
	if req.ConnectionID == "" {
		return nil, model.ErrInvalidRequest("connectionId is required")
	}
	if req.TableName == "" {
		return nil, model.ErrInvalidRequest("tableName is required")
	}
	if len(req.Changes) == 0 {
		return nil, model.ErrInvalidRequest("changes is required")
	}
	if len(req.Changes) > model.MaxBatchCellUpdates {
		return nil, model.ErrInvalidRequest(fmt.Sprintf("at most %d changes per batch", model.MaxBatchCellUpdates))
	}
	drv, err := s.mgr.Driver(req.ConnectionID)
	if err != nil {
		return nil, err
	}
	if drv.ReadOnly() {
		return nil, model.ErrReadOnly()
	}
	schema, err := drv.GetTableSchema(ctx, req.TableName)
	if err != nil {
		return nil, err
	}
	count, err := drv.UpdateCells(ctx, req.TableName, req.Changes, schema)
	if err != nil {
		return nil, err
	}
	return &model.UpdateCellsBatchResult{UpdatedCount: count}, nil
}
