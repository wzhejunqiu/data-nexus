package wails

import (
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type CannedQueryService struct {
	svc *service.CannedQueryService
	log *zap.Logger
}

func NewCannedQueryService(svc *service.CannedQueryService, log *zap.Logger) *CannedQueryService {
	return &CannedQueryService{svc: svc, log: log}
}

func (s *CannedQueryService) ListCannedQueries() (*model.CannedQueryList, error) {
	return call(s.log, "CannedQueryService.ListCannedQueries", func() (*model.CannedQueryList, error) {
		return s.svc.ListCannedQueries()
	})
}

func (s *CannedQueryService) SaveCannedQuery(req model.SaveCannedQueryRequest) (*model.CannedQuery, error) {
	return call(s.log, "CannedQueryService.SaveCannedQuery", func() (*model.CannedQuery, error) {
		return s.svc.SaveCannedQuery(req)
	})
}

func (s *CannedQueryService) DeleteCannedQuery(id string) error {
	return callVoid(s.log, "CannedQueryService.DeleteCannedQuery", func() error {
		return s.svc.DeleteCannedQuery(id)
	})
}
