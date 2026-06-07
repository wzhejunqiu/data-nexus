package wails

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type SchemaService struct {
	query *service.QueryService
	log   *zap.Logger
}

func NewSchemaService(query *service.QueryService, log *zap.Logger) *SchemaService {
	return &SchemaService{query: query, log: log}
}

func (s *SchemaService) ListTables(connectionID string) (*model.TableList, error) {
	return call(s.log, "SchemaService.ListTables", func() (*model.TableList, error) {
		return s.query.ListTables(context.Background(), connectionID)
	})
}

func (s *SchemaService) GetTableSchema(connectionID string, tableName string) (*model.TableSchema, error) {
	return call(s.log, "SchemaService.GetTableSchema", func() (*model.TableSchema, error) {
		return s.query.GetTableSchema(context.Background(), connectionID, tableName)
	})
}

func (s *SchemaService) GetTableProfile(connectionID string, tableName string) (*model.TableProfile, error) {
	return call(s.log, "SchemaService.GetTableProfile", func() (*model.TableProfile, error) {
		return s.query.GetTableProfile(context.Background(), connectionID, tableName)
	})
}

func (s *SchemaService) DetectFTSTable(connectionID string, tableName string) (*model.FTSInfo, error) {
	return call(s.log, "SchemaService.DetectFTSTable", func() (*model.FTSInfo, error) {
		return s.query.DetectFTSTable(context.Background(), connectionID, tableName)
	})
}
