package secrets

import (
	"errors"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func appErrCode(err error) string {
	appErr, ok := err.(*model.AppError)
	if !ok {
		return ""
	}
	return appErr.Code
}

func TestMockStoreKeychainRoundTrip(t *testing.T) {
	m := NewMockStore()
	if m.ActiveBackend() != BackendKeychain {
		t.Fatalf("expected keychain, got %q", m.Backend)
	}
	if !m.VaultInitialized() || !m.VaultUnlocked() {
		t.Fatal("expected initialized and unlocked")
	}
	if err := m.SetPassword("c1", "secret"); err != nil {
		t.Fatal(err)
	}
	pw, err := m.GetPassword("c1")
	if err != nil || pw != "secret" {
		t.Fatalf("got %q err %v", pw, err)
	}
	if err := m.DeletePassword("c1"); err != nil {
		t.Fatal(err)
	}
	pw, _ = m.GetPassword("c1")
	if pw != "" {
		t.Fatalf("expected empty after delete, got %q", pw)
	}
}

func TestMockStoreVaultLocked(t *testing.T) {
	m := NewMockVaultStore()
	if !m.VaultInitialized() || m.VaultUnlocked() {
		t.Fatal("expected initialized and locked vault store")
	}
	m.Unlocked = true
	if err := m.SetPassword("c1", "secret"); err != nil {
		t.Fatal(err)
	}
	m.Unlocked = false

	_, getErr := m.GetPassword("c1")
	if appErrCode(getErr) != "SECRETS_VAULT_LOCKED" {
		t.Fatalf("expected vault locked, got %v", getErr)
	}
	if err := m.SetPassword("c2", "x"); appErrCode(err) != "SECRETS_VAULT_LOCKED" {
		t.Fatalf("expected vault locked on set, got %v", err)
	}
}

func TestMockStoreInitAndUnlock(t *testing.T) {
	m := NewMockVaultStore()
	m.Initialized = false
	m.Unlocked = false

	unlockErr := m.UnlockVault("pass")
	if appErrCode(unlockErr) != "SECRETS_VAULT_NOT_INITIALIZED" {
		t.Fatalf("expected not initialized, got %v", unlockErr)
	}
	if err := m.InitVault("master12"); err != nil {
		t.Fatal(err)
	}
	if m.LastMasterPassword != "master12" {
		t.Fatalf("unexpected master password %q", m.LastMasterPassword)
	}
	if err := m.LockVault(); err != nil {
		t.Fatal(err)
	}
	if err := m.UnlockVault("master12"); err != nil {
		t.Fatal(err)
	}
}

func TestMockStoreChangeVaultPassword(t *testing.T) {
	m := NewMockStore()
	m.LastMasterPassword = "oldpass12"
	wrongErr := m.ChangeVaultPassword("wrong", "newpass12")
	if appErrCode(wrongErr) != "SECRETS_VAULT_WRONG_PASSWORD" {
		t.Fatalf("expected wrong password, got %v", wrongErr)
	}
	if err := m.ChangeVaultPassword("oldpass12", "newpass12"); err != nil {
		t.Fatal(err)
	}
	if m.LastMasterPassword != "newpass12" {
		t.Fatalf("expected updated password, got %q", m.LastMasterPassword)
	}
}

func TestMockStoreInjectedErrors(t *testing.T) {
	injected := errors.New("injected")
	m := NewMockStore()
	m.SetErr = injected
	m.GetErr = injected
	m.DeleteErr = injected
	m.InitErr = injected
	m.UnlockErr = injected

	for _, err := range []error{
		m.SetPassword("c", "x"),
		func() error { _, e := m.GetPassword("c"); return e }(),
		m.DeletePassword("c"),
		m.InitVault("x"),
		m.UnlockVault("x"),
	} {
		if !errors.Is(err, injected) {
			t.Fatalf("expected injected error, got %v", err)
		}
	}
}
