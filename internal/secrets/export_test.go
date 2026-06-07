package secrets

// NewVaultBackendForTest exposes vault backend for unit tests.
func NewVaultBackendForTest(dataDir string) (*vaultBackend, error) {
	return newVaultBackend(dataDir)
}

// NewVaultBackendInForTest exposes vault backend with a custom default directory.
func NewVaultBackendInForTest(dataDir, defaultDir string) (*vaultBackend, error) {
	return newVaultBackendIn(dataDir, defaultDir)
}

// EncryptSecretsJSONForTest exposes secrets JSON encryption for unit tests.
func EncryptSecretsJSONForTest(dataKey []byte, secrets map[string]string) (string, error) {
	return encryptSecretsJSON(dataKey, secrets)
}
