package secrets

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func probeUnavailable() error {
	return errors.New("keychain unavailable")
}

func TestNewStore(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	backend := store.ActiveBackend()
	if backend != BackendKeychain && backend != BackendVault {
		t.Fatalf("unexpected backend %q", backend)
	}

	connID := keyringTestConnID(t)
	if err := store.SetPassword(connID, "pw"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.DeletePassword(connID); err != nil {
			t.Errorf("store cleanup failed for account %q: %v", connID, err)
		}
	})

	pw, err := store.GetPassword(connID)
	if err != nil || pw != "pw" {
		t.Fatalf("got %q err %v", pw, err)
	}
	if err := store.DeletePassword(connID); err != nil {
		t.Fatal(err)
	}
}

func TestNewStoreSelectsKeychainWhenProbeSucceeds(t *testing.T) {
	dir := t.TempDir()
	store, err := newStore(dir, func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if store.ActiveBackend() != BackendKeychain {
		t.Fatalf("expected keychain backend, got %q", store.ActiveBackend())
	}
	if !store.VaultInitialized() || !store.VaultUnlocked() {
		t.Fatal("keychain store should report vault initialized and unlocked")
	}
}

func TestNewStoreFallsBackToVaultWhenProbeFails(t *testing.T) {
	dir := t.TempDir()
	store, err := newStore(dir, probeUnavailable)
	if err != nil {
		t.Fatal(err)
	}
	if store.ActiveBackend() != BackendVault {
		t.Fatalf("expected vault backend, got %q", store.ActiveBackend())
	}
	if store.VaultInitialized() {
		t.Fatal("expected uninitialized vault")
	}
}

func TestNewStoreFallbackVaultPasswordWorkflow(t *testing.T) {
	dir := t.TempDir()
	store, err := newStore(dir, probeUnavailable)
	if err != nil {
		t.Fatal(err)
	}
	if store.ActiveBackend() != BackendVault {
		t.Fatalf("expected vault backend, got %q", store.ActiveBackend())
	}
	if err := store.InitVault("testpass12"); err != nil {
		t.Fatal(err)
	}
	if !store.VaultInitialized() || !store.VaultUnlocked() {
		t.Fatal("expected initialized and unlocked")
	}
	if err := store.SetPassword("conn1", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := store.LockVault(); err != nil {
		t.Fatal(err)
	}
	_, err = store.GetPassword("conn1")
	if err == nil || appErrCode(err) != "SECRETS_VAULT_LOCKED" {
		t.Fatalf("expected vault locked, got %v", err)
	}
	if err := store.UnlockVault("testpass12"); err != nil {
		t.Fatal(err)
	}
	pw, err := store.GetPassword("conn1")
	if err != nil || pw != "secret" {
		t.Fatalf("got %q err %v", pw, err)
	}
}

func TestCompositeStoreDelegatesToKeyring(t *testing.T) {
	kc := newKeyringBackend()
	if err := kc.Probe(); err != nil {
		t.Skipf("OS keychain unavailable: %v", err)
	}
	store := &compositeStore{backend: kc, active: BackendKeychain}

	if store.ActiveBackend() != BackendKeychain {
		t.Fatalf("expected keychain, got %q", store.ActiveBackend())
	}
	if !store.VaultInitialized() || !store.VaultUnlocked() {
		t.Fatal("keychain composite should report vault ready")
	}
	for _, err := range []error{
		store.InitVault("x"),
		store.UnlockVault("x"),
		store.LockVault(),
		store.ChangeVaultPassword("old", "newpass12"),
	} {
		if err != nil {
			t.Fatalf("expected vault no-op on keychain, got %v", err)
		}
	}
}

func TestNewStoreFallbackVaultChangePassword(t *testing.T) {
	dir := t.TempDir()
	store, err := newStore(dir, probeUnavailable)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InitVault("oldpass12"); err != nil {
		t.Fatal(err)
	}
	if err := store.ChangeVaultPassword("oldpass12", "newpass12"); err != nil {
		t.Fatal(err)
	}
	if err := store.LockVault(); err != nil {
		t.Fatal(err)
	}
	if err := store.UnlockVault("newpass12"); err != nil {
		t.Fatal(err)
	}
	if !store.VaultUnlocked() {
		t.Fatal("expected unlocked with new password")
	}
}

func TestNewStoreFallbackVaultDeletePassword(t *testing.T) {
	dir := t.TempDir()
	store, err := newStore(dir, probeUnavailable)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InitVault("testpass12"); err != nil {
		t.Fatal(err)
	}
	if err := store.SetPassword("conn1", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := store.DeletePassword("conn1"); err != nil {
		t.Fatal(err)
	}
	pw, err := store.GetPassword("conn1")
	if err != nil || pw != "" {
		t.Fatalf("got %q err %v", pw, err)
	}
}

func TestNewStoreVaultBackendCreateFails(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := newStore(filepath.Join(blocker, "vault"), probeUnavailable)
	if err == nil {
		t.Fatal("expected vault backend creation error")
	}
}

func TestValidateMasterPassword(t *testing.T) {
	if err := validateMasterPassword("short"); err == nil {
		t.Fatal("expected error for short password")
	}
	if err := validateMasterPassword("longenough"); err != nil {
		t.Fatalf("expected valid password, got %v", err)
	}
}
