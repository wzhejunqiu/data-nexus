package wails

import (
	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/secrets"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	"go.uber.org/zap"
)

type SecretsService struct {
	mgr *service.ConnectionManager
	log *zap.Logger
}

func NewSecretsService(mgr *service.ConnectionManager, log *zap.Logger) *SecretsService {
	return &SecretsService{mgr: mgr, log: log}
}

func (s *SecretsService) store() secrets.Store {
	return s.mgr.SecretsStore()
}

func (s *SecretsService) GetSecretsBackend() (string, error) {
	return call(s.log, "SecretsService.GetSecretsBackend", func() (string, error) {
		return string(s.store().ActiveBackend()), nil
	})
}

func (s *SecretsService) VaultInitialized() (bool, error) {
	return call(s.log, "SecretsService.VaultInitialized", func() (bool, error) {
		return s.store().VaultInitialized(), nil
	})
}

func (s *SecretsService) VaultUnlocked() (bool, error) {
	return call(s.log, "SecretsService.VaultUnlocked", func() (bool, error) {
		return s.store().VaultUnlocked(), nil
	})
}

func (s *SecretsService) InitVault(masterPassword string) error {
	err := callVoid(s.log, "SecretsService.InitVault", func() error {
		return s.store().InitVault(masterPassword)
	})
	if err == nil {
		s.log.Debug("vault initialized")
	}
	return err
}

func (s *SecretsService) UnlockVault(masterPassword string) error {
	err := callVoid(s.log, "SecretsService.UnlockVault", func() error {
		return s.store().UnlockVault(masterPassword)
	})
	if err == nil {
		s.log.Debug("vault unlocked")
	}
	return err
}

func (s *SecretsService) LockVault() error {
	err := callVoid(s.log, "SecretsService.LockVault", func() error {
		return s.store().LockVault()
	})
	if err == nil {
		s.log.Debug("vault locked")
	}
	return err
}

func (s *SecretsService) ChangeVaultPassword(oldPassword, newPassword string) error {
	err := callVoid(s.log, "SecretsService.ChangeVaultPassword", func() error {
		return s.store().ChangeVaultPassword(oldPassword, newPassword)
	})
	if err == nil {
		s.log.Debug("vault password changed")
	}
	return err
}

func (s *SecretsService) IsVaultRequiredForRemote() (bool, error) {
	return call(s.log, "SecretsService.IsVaultRequiredForRemote", func() (bool, error) {
		return s.store().ActiveBackend() == secrets.BackendVault, nil
	})
}

// Ensure secrets errors are typed for frontend mapping.
var (
	_ = model.ErrSecretsVaultLocked
	_ = model.ErrSecretsVaultNotInitialized
	_ = model.ErrSecretsVaultWrongPassword
)
