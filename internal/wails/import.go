package wails

import (
	"context"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type ImportService struct {
	svc *service.ImportService
	log *zap.Logger
}

func NewImportService(svc *service.ImportService, log *zap.Logger) *ImportService {
	return &ImportService{svc: svc, log: log}
}

func (s *ImportService) ParseCSVPreview(req model.ParseCSVPreviewRequest) (*model.CSVPreview, error) {
	return call(s.log, "ImportService.ParseCSVPreview", func() (*model.CSVPreview, error) {
		return s.svc.ParseCSVPreview(context.Background(), req)
	})
}

func (s *ImportService) ImportCSV(req model.ImportCSVRequest) (*model.ImportCSVResult, error) {
	return call(s.log, "ImportService.ImportCSV", func() (*model.ImportCSVResult, error) {
		return s.svc.ImportCSV(context.Background(), req)
	})
}
