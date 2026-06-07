package secrets

import "github.com/wzhejunqiu/data-nexus/internal/model"

// MockStore is an in-memory secrets backend for tests.
type MockStore struct {
	Backend            Backend
	Passwords          map[string]string
	Initialized        bool
	Unlocked           bool
	InitErr            error
	SetErr             error
	GetErr             error
	DeleteErr          error
	UnlockErr          error
	LastMasterPassword string
}

func NewMockStore() *MockStore {
	return &MockStore{
		Backend:     BackendKeychain,
		Passwords:   map[string]string{},
		Initialized: true,
		Unlocked:    true,
	}
}

func NewMockVaultStore() *MockStore {
	return &MockStore{
		Backend:     BackendVault,
		Passwords:   map[string]string{},
		Initialized: true,
		Unlocked:    false,
	}
}

func (m *MockStore) SetPassword(connID, password string) error {
	if m.SetErr != nil {
		return m.SetErr
	}
	if m.Backend == BackendVault && !m.Unlocked {
		return model.ErrSecretsVaultLocked()
	}
	m.Passwords[connID] = password
	return nil
}

func (m *MockStore) GetPassword(connID string) (string, error) {
	if m.GetErr != nil {
		return "", m.GetErr
	}
	if m.Backend == BackendVault && !m.Unlocked {
		return "", model.ErrSecretsVaultLocked()
	}
	return m.Passwords[connID], nil
}

func (m *MockStore) DeletePassword(connID string) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.Passwords, connID)
	return nil
}

func (m *MockStore) ActiveBackend() Backend { return m.Backend }

func (m *MockStore) VaultInitialized() bool { return m.Initialized }

func (m *MockStore) VaultUnlocked() bool { return m.Unlocked }

func (m *MockStore) InitVault(masterPassword string) error {
	if m.InitErr != nil {
		return m.InitErr
	}
	m.Initialized = true
	m.Unlocked = true
	m.LastMasterPassword = masterPassword
	return nil
}

func (m *MockStore) UnlockVault(masterPassword string) error {
	if m.UnlockErr != nil {
		return m.UnlockErr
	}
	if !m.Initialized {
		return model.ErrSecretsVaultNotInitialized()
	}
	m.Unlocked = true
	m.LastMasterPassword = masterPassword
	return nil
}

func (m *MockStore) LockVault() error {
	m.Unlocked = false
	return nil
}

func (m *MockStore) ChangeVaultPassword(oldPassword, newPassword string) error {
	if oldPassword != m.LastMasterPassword && m.LastMasterPassword != "" {
		return model.ErrSecretsVaultWrongPassword()
	}
	m.LastMasterPassword = newPassword
	return nil
}
