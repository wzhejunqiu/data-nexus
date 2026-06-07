package wails

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type SqlExecutionService struct {
	svc *service.SqlExecutionService
	log *zap.Logger
}

func NewSqlExecutionService(svc *service.SqlExecutionService, log *zap.Logger) *SqlExecutionService {
	return &SqlExecutionService{svc: svc, log: log}
}

func (s *SqlExecutionService) ListQueryHistory(connectionID string) ([]string, error) {
	return call(s.log, "SqlExecutionService.ListQueryHistory", func() ([]string, error) {
		return s.svc.ListQueryHistory(context.Background(), connectionID)
	})
}

func (s *SqlExecutionService) ListSqlExecutions(connectionID string, limit int) (*model.SqlExecutionList, error) {
	return call(s.log, "SqlExecutionService.ListSqlExecutions", func() (*model.SqlExecutionList, error) {
		return s.svc.ListSqlExecutions(context.Background(), connectionID, limit)
	})
}
