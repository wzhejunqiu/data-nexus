package wails

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

const appVersion = "0.1.0"

type AppService struct {
	log                *zap.Logger
	rt                 RuntimePort
	ctx                context.Context
	mu                 sync.RWMutex
	activeConnectionID string
}

func NewAppService(log *zap.Logger) *AppService {
	return NewAppServiceWithRuntime(log, wailsRuntime{})
}

func NewAppServiceWithRuntime(log *zap.Logger, rt RuntimePort) *AppService {
	return &AppService{log: log, rt: rt}
}

func (s *AppService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *AppService) SetActiveConnection(connectionID string) error {
	return callVoid(s.log, "AppService.SetActiveConnection", func() error {
		s.mu.Lock()
		s.activeConnectionID = connectionID
		s.mu.Unlock()
		return nil
	})
}

func (s *AppService) ActiveConnectionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeConnectionID
}

func (s *AppService) SetWindowTitle(title string) error {
	return callVoid(s.log, "AppService.SetWindowTitle", func() error {
		if s.ctx == nil {
			return model.ErrInternal("app context not ready")
		}
		s.rt.WindowSetTitle(s.ctx, title)
		return nil
	})
}

func (s *AppService) ShowAbout() error {
	return callVoid(s.log, "AppService.ShowAbout", func() error {
		if s.ctx == nil {
			return model.ErrInternal("app context not ready")
		}
		info, err := s.GetVersion()
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Data Nexus v%s\nPlatform: %s/%s", info.Version, info.Platform, info.Arch)
		_, err = s.rt.MessageDialog(s.ctx, wailsruntime.MessageDialogOptions{
			Type:    wailsruntime.InfoDialog,
			Title:   "About Data Nexus",
			Message: msg,
		})
		return err
	})
}

func (s *AppService) EmitThemeChange(mode string) error {
	return callVoid(s.log, "AppService.EmitThemeChange", func() error {
		if s.ctx == nil {
			return model.ErrInternal("app context not ready")
		}
		s.rt.EventsEmit(s.ctx, "app:theme", mode)
		return nil
	})
}

func (s *AppService) EmitLanguageChange(lang string) error {
	return callVoid(s.log, "AppService.EmitLanguageChange", func() error {
		if s.ctx == nil {
			return model.ErrInternal("app context not ready")
		}
		s.rt.EventsEmit(s.ctx, "app:language", lang)
		return nil
	})
}

func (s *AppService) GetVersion() (*model.VersionInfo, error) {
	return call(s.log, "AppService.GetVersion", func() (*model.VersionInfo, error) {
		return &model.VersionInfo{
			Version:  appVersion,
			Platform: runtime.GOOS,
			Arch:     runtime.GOARCH,
		}, nil
	})
}
