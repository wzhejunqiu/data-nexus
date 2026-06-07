package secrets

import (
	"errors"

	"github.com/zalando/go-keyring"
)

type keyringBackend struct{}

func newKeyringBackend() *keyringBackend {
	return &keyringBackend{}
}

func (k *keyringBackend) Probe() error {
	if err := keyring.Set(ServiceName, probeKey, "probe"); err != nil {
		return err
	}
	got, err := keyring.Get(ServiceName, probeKey)
	if err != nil || got != "probe" {
		return errors.New("keyring probe failed")
	}
	_ = keyring.Delete(ServiceName, probeKey)
	return nil
}

func (k *keyringBackend) SetPassword(connID, password string) error {
	return keyring.Set(ServiceName, connID, password)
}

func (k *keyringBackend) GetPassword(connID string) (string, error) {
	pw, err := keyring.Get(ServiceName, connID)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return pw, nil
}

func (k *keyringBackend) DeletePassword(connID string) error {
	err := keyring.Delete(ServiceName, connID)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func (k *keyringBackend) VaultInitialized() bool { return true }

func (k *keyringBackend) VaultUnlocked() bool { return true }

func (k *keyringBackend) InitVault(string) error { return nil }

func (k *keyringBackend) UnlockVault(string) error { return nil }

func (k *keyringBackend) LockVault() error { return nil }

func (k *keyringBackend) ChangeVaultPassword(string, string) error { return nil }
