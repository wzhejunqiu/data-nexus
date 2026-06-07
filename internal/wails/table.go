package wails

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type TableService struct {
	query *service.QueryService
	log   *zap.Logger
}

func NewTableService(query *service.QueryService, log *zap.Logger) *TableService {
	return &TableService{query: query, log: log}
}

func (s *TableService) BrowseRows(req model.BrowseRowsRequest) (*model.PaginatedTableData, error) {
	return call(s.log, "TableService.BrowseRows", func() (*model.PaginatedTableData, error) {
		return s.query.BrowseRows(context.Background(), req)
	})
}
