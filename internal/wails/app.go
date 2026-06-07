package wails

import (
	"runtime"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

const appVersion = "0.1.0"

type AppService struct {
	log *zap.Logger
}

func NewAppService(log *zap.Logger) *AppService {
	return &AppService{log: log}
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
