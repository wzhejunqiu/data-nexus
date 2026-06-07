package secrets

import (
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type Backend string

const (
	BackendKeychain Backend = "keychain"
	BackendVault    Backend = "vault"
)

const (
	ServiceName       = "data-nexus"
	MinMasterPassword = 8
	vaultVersion      = 1
	probeKey          = "__data_nexus_probe__"
	pbkdf2Iterations  = 200_000
	rsaKeyBits        = 4096
	aesKeySize        = 32
)

type Store interface {
	SetPassword(connID, password string) error
	GetPassword(connID string) (string, error)
	DeletePassword(connID string) error
	ActiveBackend() Backend

	VaultInitialized() bool
	VaultUnlocked() bool
	InitVault(masterPassword string) error
	UnlockVault(masterPassword string) error
	LockVault() error
	ChangeVaultPassword(oldPassword, newPassword string) error
}

func NewStore(dataDir string) (Store, error) {
	return newStore(dataDir, func() error { return newKeyringBackend().Probe() })
}

func newStore(dataDir string, probe func() error) (Store, error) {
	if probe() == nil {
		return &compositeStore{backend: newKeyringBackend(), active: BackendKeychain}, nil
	}
	vault, err := newVaultBackend(dataDir)
	if err != nil {
		return nil, err
	}
	return &compositeStore{backend: vault, active: BackendVault}, nil
}

type compositeStore struct {
	backend storeBackend
	active  Backend
}

type storeBackend interface {
	SetPassword(connID, password string) error
	GetPassword(connID string) (string, error)
	DeletePassword(connID string) error
	VaultInitialized() bool
	VaultUnlocked() bool
	InitVault(masterPassword string) error
	UnlockVault(masterPassword string) error
	LockVault() error
	ChangeVaultPassword(oldPassword, newPassword string) error
}

func (s *compositeStore) SetPassword(connID, password string) error {
	return s.backend.SetPassword(connID, password)
}

func (s *compositeStore) GetPassword(connID string) (string, error) {
	return s.backend.GetPassword(connID)
}

func (s *compositeStore) DeletePassword(connID string) error {
	return s.backend.DeletePassword(connID)
}

func (s *compositeStore) ActiveBackend() Backend { return s.active }

func (s *compositeStore) VaultInitialized() bool { return s.backend.VaultInitialized() }

func (s *compositeStore) VaultUnlocked() bool { return s.backend.VaultUnlocked() }

func (s *compositeStore) InitVault(masterPassword string) error {
	return s.backend.InitVault(masterPassword)
}

func (s *compositeStore) UnlockVault(masterPassword string) error {
	return s.backend.UnlockVault(masterPassword)
}

func (s *compositeStore) LockVault() error { return s.backend.LockVault() }

func (s *compositeStore) ChangeVaultPassword(oldPassword, newPassword string) error {
	return s.backend.ChangeVaultPassword(oldPassword, newPassword)
}

func validateMasterPassword(pw string) error {
	if len(pw) < MinMasterPassword {
		return model.ErrInvalidRequest("master password must be at least 8 characters")
	}
	return nil
}
