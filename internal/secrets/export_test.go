package secrets

// NewVaultBackendForTest exposes vault backend for unit tests.
func NewVaultBackendForTest(dataDir string) (*vaultBackend, error) {
	return newVaultBackend(dataDir)
}
