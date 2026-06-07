package secrets_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/secrets"
)

func TestVaultInitUnlockRoundTrip(t *testing.T) {
	dir := t.TempDir()
	v, err := secrets.NewVaultBackendForTest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.VaultInitialized() {
		t.Fatal("expected uninitialized")
	}
	if err := v.InitVault("testpass12"); err != nil {
		t.Fatal(err)
	}
	if !v.VaultInitialized() || !v.VaultUnlocked() {
		t.Fatal("expected initialized and unlocked")
	}
	if err := v.SetPassword("conn1", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := v.LockVault(); err != nil {
		t.Fatal(err)
	}
	if v.VaultUnlocked() {
		t.Fatal("expected locked")
	}
	_, err = v.GetPassword("conn1")
	if err == nil || err.(*model.AppError).Code != "SECRETS_VAULT_LOCKED" {
		t.Fatalf("expected locked, got %v", err)
	}
	if err := v.UnlockVault("testpass12"); err != nil {
		t.Fatal(err)
	}
	pw, err := v.GetPassword("conn1")
	if err != nil || pw != "secret" {
		t.Fatalf("got %q err %v", pw, err)
	}
	if err := v.DeletePassword("conn1"); err != nil {
		t.Fatal(err)
	}
	pw, _ = v.GetPassword("conn1")
	if pw != "" {
		t.Fatalf("expected empty after delete, got %q", pw)
	}
}

func TestVaultWrongPassword(t *testing.T) {
	dir := t.TempDir()
	v, err := secrets.NewVaultBackendForTest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.InitVault("correctpass"); err != nil {
		t.Fatal(err)
	}
	if err := v.LockVault(); err != nil {
		t.Fatal(err)
	}
	err = v.UnlockVault("wrongpass")
	if err == nil || err.(*model.AppError).Code != "SECRETS_VAULT_WRONG_PASSWORD" {
		t.Fatalf("expected wrong password, got %v", err)
	}
}

func TestVaultPersistReload(t *testing.T) {
	dir := t.TempDir()
	v1, err := secrets.NewVaultBackendForTest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v1.InitVault("persistpass"); err != nil {
		t.Fatal(err)
	}
	if err := v1.SetPassword("id-a", "pw-a"); err != nil {
		t.Fatal(err)
	}
	if err := v1.LockVault(); err != nil {
		t.Fatal(err)
	}
	v2, err := secrets.NewVaultBackendForTest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !v2.VaultInitialized() {
		t.Fatal("expected initialized on reload")
	}
	if err := v2.UnlockVault("persistpass"); err != nil {
		t.Fatal(err)
	}
	pw, err := v2.GetPassword("id-a")
	if err != nil || pw != "pw-a" {
		t.Fatalf("got %q err %v", pw, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "vault.json")); err != nil {
		t.Fatal(err)
	}
}
