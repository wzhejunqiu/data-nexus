package wails

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

type DialogService struct {
	ctx context.Context
	log *zap.Logger
}

func NewDialogService(log *zap.Logger) *DialogService {
	return &DialogService{log: log}
}

func (s *DialogService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *DialogService) OpenDatabaseFile() (string, error) {
	return call(s.log, "DialogService.OpenDatabaseFile", func() (string, error) {
		if s.ctx == nil {
			return "", model.ErrInternal("dialog context not ready")
		}
		path, err := runtime.OpenFileDialog(s.ctx, runtime.OpenDialogOptions{
			Title: "Open SQLite Database",
			Filters: []runtime.FileFilter{
				{DisplayName: "SQLite Database", Pattern: "*.db;*.sqlite;*.sqlite3"},
			},
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
