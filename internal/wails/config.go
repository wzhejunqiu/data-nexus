package wails

import (
	"github.com/wzhejunqiu/data-nexus/internal/config"
	"go.uber.org/zap"
)

type ConfigService struct {
	log *zap.Logger
}

func NewConfigService(log *zap.Logger) *ConfigService {
	return &ConfigService{log: log}
}

func (s *ConfigService) GetConfig() (config.Config, error) {
	return call(s.log, "ConfigService.GetConfig", func() (config.Config, error) {
		return config.Load(), nil
	})
}

func (s *ConfigService) UpdateConfig(cfg config.Config) error {
	return callVoid(s.log, "ConfigService.UpdateConfig", func() error {
		return config.Save(cfg)
	})
}

func (s *ConfigService) GetConfigPath() (string, error) {
	return config.ConfigPath(), nil
}
