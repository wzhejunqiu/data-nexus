package wails_test

import (
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/secrets"
	"github.com/wzhejunqiu/data-nexus/internal/service"
	wailssvc "github.com/wzhejunqiu/data-nexus/internal/wails"
	"go.uber.org/zap"
)

func newSecretsTestEnv(t *testing.T, mock secrets.Store) *wailssvc.SecretsService {
	t.Helper()
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	mgr := service.NewTestConnectionManagerWithSecrets(store, mock)
	return wailssvc.NewSecretsService(mgr, zap.NewNop())
}

func TestSecretsServiceKeychainBackend(t *testing.T) {
	svc := newSecretsTestEnv(t, secrets.NewMockStore())

	backend, err := svc.GetSecretsBackend()
	if err != nil || backend != "keychain" {
		t.Fatalf("got backend %q err %v", backend, err)
	}
	init, err := svc.VaultInitialized()
	if err != nil || !init {
		t.Fatalf("VaultInitialized=%v err %v", init, err)
	}
	unlocked, err := svc.VaultUnlocked()
	if err != nil || !unlocked {
		t.Fatalf("VaultUnlocked=%v err %v", unlocked, err)
	}
	required, err := svc.IsVaultRequiredForRemote()
	if err != nil || required {
		t.Fatalf("IsVaultRequiredForRemote=%v err %v", required, err)
	}
	for _, err := range []error{
		svc.InitVault("ignored"),
		svc.UnlockVault("ignored"),
		svc.LockVault(),
	} {
		if err != nil {
			t.Fatalf("expected no-op on keychain, got %v", err)
		}
	}
	if err := svc.ChangeVaultPassword("ignored", "newpass12"); err != nil {
		t.Fatalf("expected change password no-op on keychain, got %v", err)
	}
}

func TestSecretsServiceVaultBackend(t *testing.T) {
	mock := secrets.NewMockVaultStore()
	mock.Unlocked = false
	svc := newSecretsTestEnv(t, mock)

	backend, err := svc.GetSecretsBackend()
	if err != nil || backend != "vault" {
		t.Fatalf("got backend %q err %v", backend, err)
	}
	required, err := svc.IsVaultRequiredForRemote()
	if err != nil || !required {
		t.Fatalf("IsVaultRequiredForRemote=%v err %v", required, err)
	}
	init, err := svc.VaultInitialized()
	if err != nil || !init {
		t.Fatalf("VaultInitialized=%v err %v", init, err)
	}
	unlocked, err := svc.VaultUnlocked()
	if err != nil || unlocked {
		t.Fatalf("VaultUnlocked=%v err %v", unlocked, err)
	}
	if err := svc.UnlockVault("master12"); err != nil {
		t.Fatal(err)
	}
	unlocked, err = svc.VaultUnlocked()
	if err != nil || !unlocked {
		t.Fatalf("expected unlocked after UnlockVault, got %v err %v", unlocked, err)
	}
}
