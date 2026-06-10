package wails

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/version"
	"go.uber.org/zap"
)

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
		s.rt.EventsEmit(s.ctx, "app:about")
		return nil
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
			Version:  version.Version(),
			Platform: runtime.GOOS,
			Arch:     runtime.GOARCH,
		}, nil
	})
}

func (s *AppService) GetPlatform() (string, error) {
	return call(s.log, "AppService.GetPlatform", func() (string, error) {
		return runtime.GOOS + "/" + runtime.GOARCH, nil
	})
}

func (s *AppService) RevealFileInExplorer(filePath string) error {
	return callVoid(s.log, "AppService.RevealFileInExplorer", func() error {
		if filePath == "" {
			return model.ErrInvalidRequest("file path is required")
		}
		if _, err := os.Stat(filePath); err != nil {
			if os.IsNotExist(err) {
				return model.ErrInvalidRequest("file not found")
			}
			return model.ErrInternal(err.Error())
		}
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", "-R", filePath)
		case "windows":
			cmd = exec.Command("explorer", "/select,", filepath.Clean(filePath))
		default:
			cmd = exec.Command("xdg-open", filepath.Dir(filePath))
		}
		if err := cmd.Start(); err != nil {
			return model.ErrInternal(err.Error())
		}
		return nil
	})
}
