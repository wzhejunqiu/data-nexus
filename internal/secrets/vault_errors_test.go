package secrets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func initTestVault(t *testing.T) (*vaultBackend, string) {
	t.Helper()
	dir := t.TempDir()
	v, err := newVaultBackend(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.InitVault("testpass12"); err != nil {
		t.Fatal(err)
	}
	return v, dir
}

func readVaultMeta(t *testing.T, dir string) vaultFile {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "vault.json"))
	if err != nil {
		t.Fatal(err)
	}
	var meta vaultFile
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatal(err)
	}
	return meta
}

func writeVaultMeta(t *testing.T, dir string, meta vaultFile) {
	t.Helper()
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "vault.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNewVaultBackendEmptyDataDir(t *testing.T) {
	defaultDir := t.TempDir()
	v, err := newVaultBackendIn("", defaultDir)
	if err != nil {
		t.Fatal(err)
	}
	if v.dir != defaultDir {
		t.Fatalf("expected dir %q, got %q", defaultDir, v.dir)
	}
}

func TestNewVaultBackendMkdirFails(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := newVaultBackend(filepath.Join(blocker, "vault"))
	if err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestUnlockVaultMissingPrivateKey(t *testing.T) {
	v, dir := initTestVault(t)
	if err := v.LockVault(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "rsa_private.enc")); err != nil {
		t.Fatal(err)
	}
	if err := v.UnlockVault("testpass12"); err == nil {
		t.Fatal("expected read error")
	}
}

func TestUnlockVaultCorruptVaultJSON(t *testing.T) {
	v, dir := initTestVault(t)
	if err := v.LockVault(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "vault.json"), []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := v.UnlockVault("testpass12"); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestUnlockVaultCorruptWrappedDataKey(t *testing.T) {
	v, dir := initTestVault(t)
	if err := v.LockVault(); err != nil {
		t.Fatal(err)
	}
	meta := readVaultMeta(t, dir)
	meta.WrappedDataKey = "invalid"
	writeVaultMeta(t, dir, meta)
	err := v.UnlockVault("testpass12")
	if err == nil || appErrCode(err) != "SECRETS_VAULT_WRONG_PASSWORD" {
		t.Fatalf("expected wrong password, got %v", err)
	}
}

func TestUnlockVaultCorruptEncryptedSecrets(t *testing.T) {
	v, dir := initTestVault(t)
	if err := v.LockVault(); err != nil {
		t.Fatal(err)
	}
	meta := readVaultMeta(t, dir)
	meta.EncryptedSecrets = "not-valid"
	writeVaultMeta(t, dir, meta)
	if err := v.UnlockVault("testpass12"); err == nil {
		t.Fatal("expected decrypt error")
	}
}

func TestUnlockVaultNilSecretsThenSetPassword(t *testing.T) {
	v, dir := initTestVault(t)
	dataKey := make([]byte, len(v.dataKey))
	copy(dataKey, v.dataKey)
	if err := v.LockVault(); err != nil {
		t.Fatal(err)
	}
	meta := readVaultMeta(t, dir)
	encNull, err := encryptSecretsJSON(dataKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	meta.EncryptedSecrets = encNull
	writeVaultMeta(t, dir, meta)
	if err := v.UnlockVault("testpass12"); err != nil {
		t.Fatal(err)
	}
	if err := v.SetPassword("conn1", "secret"); err != nil {
		t.Fatal(err)
	}
	pw, err := v.GetPassword("conn1")
	if err != nil || pw != "secret" {
		t.Fatalf("got %q err %v", pw, err)
	}
}

func TestChangeVaultPasswordMissingPrivateKey(t *testing.T) {
	v, dir := initTestVault(t)
	if err := os.Remove(filepath.Join(dir, "rsa_private.enc")); err != nil {
		t.Fatal(err)
	}
	if err := v.ChangeVaultPassword("testpass12", "newpass12"); err == nil {
		t.Fatal("expected read error")
	}
}

func TestSetPasswordCorruptVaultJSON(t *testing.T) {
	v, dir := initTestVault(t)
	if err := os.WriteFile(filepath.Join(dir, "vault.json"), []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := v.SetPassword("conn1", "secret"); err == nil {
		t.Fatal("expected persist error")
	}
}

func TestSetPasswordReadOnlyVaultJSON(t *testing.T) {
	v, dir := initTestVault(t)
	metaPath := filepath.Join(dir, "vault.json")
	if err := os.Chmod(metaPath, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(metaPath, 0o600) })
	if err := v.SetPassword("conn1", "secret"); err == nil {
		t.Fatal("expected write error")
	}
}

func TestInitVaultMetaPathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "vault.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	v, err := newVaultBackend(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.InitVault("testpass12"); err == nil {
		t.Fatal("expected write error when meta path is directory")
	}
}

func TestInitVaultPrivPathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "rsa_private.enc"), 0o700); err != nil {
		t.Fatal(err)
	}
	v, err := newVaultBackend(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.InitVault("testpass12"); err == nil {
		t.Fatal("expected write error when private key path is directory")
	}
}

func TestUnlockVaultDecryptsInvalidPEM(t *testing.T) {
	v, dir := initTestVault(t)
	if err := v.LockVault(); err != nil {
		t.Fatal(err)
	}
	enc, err := encryptPrivateKeyWithPassword("not-a-valid-pem", "testpass12")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rsa_private.enc"), enc, 0o600); err != nil {
		t.Fatal(err)
	}
	err = v.UnlockVault("testpass12")
	if err == nil || appErrCode(err) != "SECRETS_VAULT_WRONG_PASSWORD" {
		t.Fatalf("expected wrong password, got %v", err)
	}
}

func TestSetPasswordMissingVaultJSON(t *testing.T) {
	v, dir := initTestVault(t)
	if err := os.Remove(filepath.Join(dir, "vault.json")); err != nil {
		t.Fatal(err)
	}
	if err := v.SetPassword("conn1", "secret"); err == nil {
		t.Fatal("expected read error")
	}
}

func TestDeletePasswordMissingVaultJSON(t *testing.T) {
	v, dir := initTestVault(t)
	if err := v.SetPassword("conn1", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "vault.json")); err != nil {
		t.Fatal(err)
	}
	if err := v.DeletePassword("conn1"); err == nil {
		t.Fatal("expected read error")
	}
}

func TestChangeVaultPasswordWrongOldPassword(t *testing.T) {
	v, _ := initTestVault(t)
	err := v.ChangeVaultPassword("wrongpass", "newpass12")
	if err == nil || appErrCode(err) != "SECRETS_VAULT_WRONG_PASSWORD" {
		t.Fatalf("expected wrong password, got %v", err)
	}
}

func TestPersistLockedEncryptFails(t *testing.T) {
	v, _ := initTestVault(t)
	v.dataKey = []byte("short")
	if err := v.SetPassword("conn1", "secret"); err == nil {
		t.Fatal("expected encrypt error")
	}
}

func TestUnlockVaultMissingVaultJSONAfterPrivateDecrypt(t *testing.T) {
	v, dir := initTestVault(t)
	if err := v.LockVault(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "vault.json")); err != nil {
		t.Fatal(err)
	}
	if err := v.UnlockVault("testpass12"); err == nil {
		t.Fatal("expected meta read error")
	}
}

func TestNewVaultBackendLoadsExistingMeta(t *testing.T) {
	dir := t.TempDir()
	v1, err := newVaultBackend(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v1.InitVault("testpass12"); err != nil {
		t.Fatal(err)
	}
	v2, err := newVaultBackend(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !v2.VaultInitialized() {
		t.Fatal("expected initialized when vault.json exists")
	}
}

func TestSetPasswordWithNilSecretsMap(t *testing.T) {
	v, _ := initTestVault(t)
	v.secrets = nil
	if err := v.SetPassword("conn1", "secret"); err != nil {
		t.Fatal(err)
	}
	pw, err := v.GetPassword("conn1")
	if err != nil || pw != "secret" {
		t.Fatalf("got %q err %v", pw, err)
	}
}
