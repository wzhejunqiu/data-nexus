package service

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/executionlog"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type SqlExecutionService struct {
	store executionlog.Store
}

func NewSqlExecutionService(store executionlog.Store) *SqlExecutionService {
	return &SqlExecutionService{store: store}
}

func (s *SqlExecutionService) ListQueryHistory(ctx context.Context, connectionID string) ([]string, error) {
	if s.store == nil {
		return []string{}, nil
	}
	return s.store.ListQueryHistory(ctx, connectionID, 50)
}

func (s *SqlExecutionService) ListSqlExecutions(ctx context.Context, connectionID string, limit int) (*model.SqlExecutionList, error) {
	if s.store == nil {
		return &model.SqlExecutionList{Items: []model.SqlExecutionRecord{}}, nil
	}
	items, err := s.store.ListExecutions(ctx, connectionID, limit)
	if err != nil {
		return nil, err
	}
	return &model.SqlExecutionList{Items: items}, nil
}

func (s *SqlExecutionService) ListAllSqlExecutions(ctx context.Context, limit int) (*model.SqlExecutionList, error) {
	if s.store == nil {
		return &model.SqlExecutionList{Items: []model.SqlExecutionRecord{}}, nil
	}
	items, err := s.store.ListAllExecutions(ctx, limit)
	if err != nil {
		return nil, err
	}
	return &model.SqlExecutionList{Items: items}, nil
}
