package wails

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

type DialogService struct {
	ctx context.Context
	log *zap.Logger
	rt  RuntimePort
}

func NewDialogService(log *zap.Logger) *DialogService {
	return NewDialogServiceWithRuntime(log, wailsRuntime{})
}

func NewDialogServiceWithRuntime(log *zap.Logger, rt RuntimePort) *DialogService {
	return &DialogService{log: log, rt: rt}
}

func (s *DialogService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *DialogService) OpenDatabaseFile() (string, error) {
	return s.openFile("Open SQLite Database", []wailsruntime.FileFilter{
		{DisplayName: "SQLite Database", Pattern: "*.db;*.sqlite;*.sqlite3"},
	})
}

func (s *DialogService) OpenCSVFile() (string, error) {
	return s.openFile("Open CSV File", []wailsruntime.FileFilter{
		{DisplayName: "CSV Files", Pattern: "*.csv;*.tsv;*.txt"},
	})
}

func (s *DialogService) openFile(title string, filters []wailsruntime.FileFilter) (string, error) {
	return call(s.log, "DialogService.openFile", func() (string, error) {
		if s.ctx == nil {
			return "", model.ErrInternal("dialog context not ready")
		}
		path, err := s.rt.OpenFileDialog(s.ctx, wailsruntime.OpenDialogOptions{
			Title:   title,
			Filters: filters,
		})
		if err != nil {
			return "", model.ErrInternal(err.Error())
		}
		if path == "" {
			return "", model.ErrDialogCancelled()
		}
		return path, nil
	})
}

func (s *DialogService) SaveFile(defaultName string, filters []model.FileFilter) (string, error) {
	return call(s.log, "DialogService.SaveFile", func() (string, error) {
		if s.ctx == nil {
			return "", model.ErrInternal("dialog context not ready")
		}
		wailsFilters := make([]wailsruntime.FileFilter, len(filters))
		for i, f := range filters {
			wailsFilters[i] = wailsruntime.FileFilter{
				DisplayName: f.DisplayName,
				Pattern:     f.Pattern,
			}
		}
		if len(wailsFilters) == 0 {
			wailsFilters = []wailsruntime.FileFilter{
				{DisplayName: "CSV Files", Pattern: "*.csv"},
			}
		}
		path, err := s.rt.SaveFileDialog(s.ctx, wailsruntime.SaveDialogOptions{
			DefaultFilename: defaultName,
			Title:           "Save File",
			Filters:         wailsFilters,
		})
		if err != nil {
			return "", model.ErrInternal(err.Error())
		}
		if path == "" {
			return "", model.ErrDialogCancelled()
		}
		return path, nil
	})
}
