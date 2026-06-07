package service

import (
	"context"
	"strings"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type QueryService struct {
	mgr *ConnectionManager
}

func NewQueryService(mgr *ConnectionManager) *QueryService {
	return &QueryService{mgr: mgr}
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
		maxRows = 1000
	}
	if isQuerySQL(req.SQL) {
		result, err := drv.QueryRows(ctx, req.SQL, req.Params, maxRows)
		if err != nil {
			return nil, err
		}
		return &model.QueryResponse{
			Kind:       "result",
			Columns:    result.Columns,
			Rows:       result.Rows,
			RowCount:   result.RowCount,
			Truncated:  result.Truncated,
			DurationMs: result.Duration.Milliseconds(),
		}, nil
	}
	execResult, err := drv.Exec(ctx, req.SQL, req.Params)
	if err != nil {
		return nil, err
	}
	return &model.QueryResponse{
		Kind:         "exec",
		RowsAffected: execResult.RowsAffected,
		LastInsertID: execResult.LastInsertID,
		DurationMs:   execResult.Duration.Milliseconds(),
	}, nil
}

func isQuerySQL(sql string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(sql))
	switch {
	case strings.HasPrefix(trimmed, "SELECT"),
		strings.HasPrefix(trimmed, "WITH"),
		strings.HasPrefix(trimmed, "PRAGMA"),
		strings.HasPrefix(trimmed, "EXPLAIN"):
		return true
	default:
		return false
	}
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
		Page:     req.Page,
		PageSize: req.PageSize,
		Sort:     req.Sort,
		Order:    req.Order,
	}
	if opts.Order == "" {
		opts.Order = model.SortAsc
	}
	return drv.BrowseTable(ctx, req.TableName, opts)
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
