package wails

import (
	"context"
	"sync"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type ExportService struct {
	svc     *service.ExportService
	dialog  *DialogService
	log     *zap.Logger
	rt      RuntimePort
	ctx     context.Context
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func NewExportService(svc *service.ExportService, dialog *DialogService, log *zap.Logger) *ExportService {
	return NewExportServiceWithRuntime(svc, dialog, log, wailsRuntime{})
}

func NewExportServiceWithRuntime(
	svc *service.ExportService,
	dialog *DialogService,
	log *zap.Logger,
	rt RuntimePort,
) *ExportService {
	return &ExportService{
		svc:     svc,
		dialog:  dialog,
		log:     log,
		rt:      rt,
		cancels: make(map[string]context.CancelFunc),
	}
}

func (s *ExportService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *ExportService) registerCancel(exportID string, cancel context.CancelFunc) {
	if exportID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancels[exportID] = cancel
}

func (s *ExportService) unregisterCancel(exportID string) {
	if exportID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cancels, exportID)
}

func (s *ExportService) CancelExportTableCSV(exportID string) error {
	return callVoid(s.log, "ExportService.CancelExportTableCSV", func() error {
		if exportID == "" {
			return model.ErrInvalidRequest("exportId is required")
		}
		s.mu.Lock()
		cancel := s.cancels[exportID]
		s.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		return nil
	})
}

func (s *ExportService) ExportTableCSV(req model.ExportTableCSVRequest) (string, error) {
	return call(s.log, "ExportService.ExportTableCSV", func() (string, error) {
		path := req.DefaultPath
		var err error
		if path == "" {
			defaultName := service.DefaultCSVFilename(req.TableName)
			path, err = s.dialog.SaveFile(defaultName, []model.FileFilter{
				{DisplayName: "CSV Files", Pattern: "*.csv"},
			})
			if err != nil {
				return "", err
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		if req.ExportID != "" {
			s.registerCancel(req.ExportID, cancel)
			defer s.unregisterCancel(req.ExportID)
		}
		defer cancel()

		onProgress := func(exported int) {
			if req.ExportID == "" || s.ctx == nil {
				return
			}
			s.rt.EventsEmit(s.ctx, "export:progress", model.ExportProgress{
				ExportID: req.ExportID,
				Exported: exported,
			})
		}

		if err := s.svc.ExportTableToFile(ctx, req, path, onProgress); err != nil {
			return "", err
		}
		return path, nil
	})
}
