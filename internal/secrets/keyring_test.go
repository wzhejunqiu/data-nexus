package secrets

import (
	"testing"
)

func skipWhenNoKeyring(t *testing.T) *keyringBackend {
	t.Helper()
	kc := newKeyringBackend()
	if err := kc.Probe(); err != nil {
		t.Skipf("OS keychain unavailable: %v", err)
	}
	return kc
}

func keyringTestConnID(t *testing.T) string {
	t.Helper()
	return "test-" + t.Name()
}

func cleanupKeyringPassword(t *testing.T, kc *keyringBackend, connID string) {
	t.Helper()
	t.Cleanup(func() {
		if err := kc.DeletePassword(connID); err != nil {
			t.Errorf("keyring cleanup failed for account %q: %v", connID, err)
		}
	})
}

func setKeyringTestPassword(t *testing.T, kc *keyringBackend, connID, password string) {
	t.Helper()
	if err := kc.SetPassword(connID, password); err != nil {
		t.Fatal(err)
	}
	cleanupKeyringPassword(t, kc, connID)
}

func TestKeyringBackendProbe(t *testing.T) {
	kc := newKeyringBackend()
	if err := kc.Probe(); err != nil {
		t.Skipf("OS keychain unavailable: %v", err)
	}
}

func TestKeyringBackendPasswordRoundTrip(t *testing.T) {
	kc := skipWhenNoKeyring(t)
	connID := keyringTestConnID(t)
	setKeyringTestPassword(t, kc, connID, "secret")

	pw, err := kc.GetPassword(connID)
	if err != nil || pw != "secret" {
		t.Fatalf("got %q err %v", pw, err)
	}
	if err := kc.SetPassword(connID, "updated"); err != nil {
		t.Fatal(err)
	}
	pw, err = kc.GetPassword(connID)
	if err != nil || pw != "updated" {
		t.Fatalf("got %q err %v after update", pw, err)
	}
}

func TestKeyringBackendGetPasswordMissing(t *testing.T) {
	kc := skipWhenNoKeyring(t)

	pw, err := kc.GetPassword("nonexistent-" + keyringTestConnID(t))
	if err != nil {
		t.Fatalf("expected no error for missing entry, got %v", err)
	}
	if pw != "" {
		t.Fatalf("expected empty password, got %q", pw)
	}
}

func TestKeyringBackendDeletePasswordMissing(t *testing.T) {
	kc := skipWhenNoKeyring(t)

	if err := kc.DeletePassword("nonexistent-" + keyringTestConnID(t)); err != nil {
		t.Fatalf("expected no error deleting missing entry, got %v", err)
	}
}

func TestKeyringBackendDeletePassword(t *testing.T) {
	kc := skipWhenNoKeyring(t)
	connID := keyringTestConnID(t)
	setKeyringTestPassword(t, kc, connID, "secret")

	if err := kc.DeletePassword(connID); err != nil {
		t.Fatal(err)
	}
	pw, err := kc.GetPassword(connID)
	if err != nil || pw != "" {
		t.Fatalf("expected empty after delete, got %q err %v", pw, err)
	}
}

func TestKeyringBackendVaultNoOps(t *testing.T) {
	kc := skipWhenNoKeyring(t)

	if !kc.VaultInitialized() || !kc.VaultUnlocked() {
		t.Fatal("keychain backend should report vault initialized and unlocked")
	}
	for _, err := range []error{
		kc.InitVault("ignored"),
		kc.UnlockVault("ignored"),
		kc.LockVault(),
		kc.ChangeVaultPassword("old", "newpass12"),
	} {
		if err != nil {
			t.Fatalf("expected no-op, got %v", err)
		}
	}
}
