package secrets

import (
	"crypto/rsa"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/wzhejunqiu/data-nexus/internal/config"
	"github.com/wzhejunqiu/data-nexus/internal/model"
)

type vaultFile struct {
	Version          int    `json:"version"`
	PublicKeyPEM     string `json:"publicKeyPem"`
	WrappedDataKey   string `json:"wrappedDataKey"`
	EncryptedSecrets string `json:"encryptedSecrets"`
}

type vaultBackend struct {
	mu          sync.RWMutex
	dir         string
	metaPath    string
	privPath    string
	initialized bool
	unlocked    bool
	rsaPriv     *rsa.PrivateKey
	dataKey     []byte
	secrets     map[string]string
}

func newVaultBackend(dataDir string) (*vaultBackend, error) {
	return newVaultBackendIn(dataDir, config.VaultDir())
}

func newVaultBackendIn(dataDir, defaultDir string) (*vaultBackend, error) {
	dir := dataDir
	if dir == "" {
		dir = defaultDir
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	v := &vaultBackend{
		dir:      dir,
		metaPath: filepath.Join(dir, "vault.json"),
		privPath: filepath.Join(dir, "rsa_private.enc"),
		secrets:  map[string]string{},
	}
	if err := v.load(); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *vaultBackend) load() error {
	if _, err := os.Stat(v.metaPath); os.IsNotExist(err) {
		return nil
	}
	v.initialized = true
	return nil
}

func (v *vaultBackend) VaultInitialized() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.initialized
}

func (v *vaultBackend) VaultUnlocked() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.unlocked
}

func (v *vaultBackend) InitVault(masterPassword string) error {
	if err := validateMasterPassword(masterPassword); err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.initialized {
		return model.ErrInvalidRequest("vault already initialized")
	}
	priv, err := generateRSAKeyPair()
	if err != nil {
		return err
	}
	pubPEM, err := marshalPublicKeyPEM(&priv.PublicKey)
	if err != nil {
		return err
	}
	privPEM, err := marshalPrivateKeyPEM(priv)
	if err != nil {
		return err
	}
	encPriv, err := encryptPrivateKeyWithPassword(privPEM, masterPassword)
	if err != nil {
		return err
	}
	dataKey, err := randomAESKey()
	if err != nil {
		return err
	}
	wrapped, err := wrapDataKeyRSA(&priv.PublicKey, dataKey)
	if err != nil {
		return err
	}
	encSecrets, err := encryptSecretsJSON(dataKey, map[string]string{})
	if err != nil {
		return err
	}
	meta := vaultFile{
		Version:          vaultVersion,
		PublicKeyPEM:     pubPEM,
		WrappedDataKey:   wrapped,
		EncryptedSecrets: encSecrets,
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(v.metaPath, metaBytes, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(v.privPath, encPriv, 0o600); err != nil {
		return err
	}
	v.initialized = true
	v.unlocked = true
	v.rsaPriv = priv
	v.dataKey = dataKey
	v.secrets = map[string]string{}
	return nil
}

func (v *vaultBackend) UnlockVault(masterPassword string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.initialized {
		return model.ErrSecretsVaultNotInitialized()
	}
	if v.unlocked {
		return nil
	}
	encPriv, err := os.ReadFile(v.privPath)
	if err != nil {
		return err
	}
	privPEM, err := decryptPrivateKeyWithPassword(encPriv, masterPassword)
	if err != nil {
		return model.ErrSecretsVaultWrongPassword()
	}
	priv, err := parsePrivateKeyPEM(privPEM)
	if err != nil {
		return model.ErrSecretsVaultWrongPassword()
	}
	metaBytes, err := os.ReadFile(v.metaPath)
	if err != nil {
		return err
	}
	var meta vaultFile
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return err
	}
	dataKey, err := unwrapDataKeyRSA(priv, meta.WrappedDataKey)
	if err != nil {
		return model.ErrSecretsVaultWrongPassword()
	}
	secrets, err := decryptSecretsJSON(dataKey, meta.EncryptedSecrets)
	if err != nil {
		return err
	}
	v.unlocked = true
	v.rsaPriv = priv
	v.dataKey = dataKey
	v.secrets = secrets
	if v.secrets == nil {
		v.secrets = map[string]string{}
	}
	return nil
}

func (v *vaultBackend) LockVault() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.unlocked = false
	v.rsaPriv = nil
	v.dataKey = nil
	v.secrets = nil
	return nil
}

func (v *vaultBackend) ChangeVaultPassword(oldPassword, newPassword string) error {
	if err := validateMasterPassword(newPassword); err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.initialized {
		return model.ErrSecretsVaultNotInitialized()
	}
	if !v.unlocked {
		return model.ErrSecretsVaultLocked()
	}
	privPEM, err := marshalPrivateKeyPEM(v.rsaPriv)
	if err != nil {
		return err
	}
	encPriv, err := os.ReadFile(v.privPath)
	if err != nil {
		return err
	}
	if _, err := decryptPrivateKeyWithPassword(encPriv, oldPassword); err != nil {
		return model.ErrSecretsVaultWrongPassword()
	}
	newEnc, err := encryptPrivateKeyWithPassword(privPEM, newPassword)
	if err != nil {
		return err
	}
	return os.WriteFile(v.privPath, newEnc, 0o600)
}

func (v *vaultBackend) SetPassword(connID, password string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.initialized {
		return model.ErrSecretsVaultNotInitialized()
	}
	if !v.unlocked {
		return model.ErrSecretsVaultLocked()
	}
	if v.secrets == nil {
		v.secrets = map[string]string{}
	}
	v.secrets[connID] = password
	return v.persistLocked()
}

func (v *vaultBackend) GetPassword(connID string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if !v.initialized {
		return "", model.ErrSecretsVaultNotInitialized()
	}
	if !v.unlocked {
		return "", model.ErrSecretsVaultLocked()
	}
	return v.secrets[connID], nil
}

func (v *vaultBackend) DeletePassword(connID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.initialized {
		return model.ErrSecretsVaultNotInitialized()
	}
	if !v.unlocked {
		return model.ErrSecretsVaultLocked()
	}
	delete(v.secrets, connID)
	return v.persistLocked()
}

func (v *vaultBackend) persistLocked() error {
	enc, err := encryptSecretsJSON(v.dataKey, v.secrets)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(v.metaPath)
	if err != nil {
		return err
	}
	var meta vaultFile
	if err := json.Unmarshal(data, &meta); err != nil {
		return err
	}
	meta.EncryptedSecrets = enc
	out, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(v.metaPath, out, 0o600)
}
