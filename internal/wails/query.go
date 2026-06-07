package wails

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type QueryService struct {
	query *service.QueryService
	log   *zap.Logger
}

func NewQueryService(query *service.QueryService, log *zap.Logger) *QueryService {
	return &QueryService{query: query, log: log}
}

func (s *QueryService) Execute(req model.ExecuteQueryRequest) (*model.QueryResponse, error) {
	return call(s.log, "QueryService.Execute", func() (*model.QueryResponse, error) {
		return s.query.Execute(context.Background(), req)
	})
}
