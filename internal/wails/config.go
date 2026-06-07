package wails

import (
	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/logger"
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"go.uber.org/zap"
)

type ConfigService struct {
	log    *zap.Logger
	logMgr *logger.Manager
}

func NewConfigService(logMgr *logger.Manager, log *zap.Logger) *ConfigService {
	return &ConfigService{logMgr: logMgr, log: log}
}

func (s *ConfigService) GetConfig() (config.Config, error) {
	return call(s.log, "ConfigService.GetConfig", func() (config.Config, error) {
		return config.Load(), nil
	})
}

func (s *ConfigService) UpdateConfig(cfg config.Config) error {
	return callVoid(s.log, "ConfigService.UpdateConfig", func() error {
		if err := config.Validate(cfg); err != nil {
			return model.ErrInvalidRequest(err.Error())
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		if s.logMgr != nil {
			s.logMgr.SetLevel(cfg.Log.Level)
		}
		return nil
	})
}

func (s *ConfigService) GetConfigPath() (string, error) {
	return config.ConfigPath(), nil
}
