package service

import (
	"go.uber.org/zap"

	"github.com/wzhejunqiu/data-nexus/internal/secrets"
)

// NewTestConnectionManager builds a manager with an in-memory mock secrets store for tests.
func NewTestConnectionManager(store *ConnectionStore) *ConnectionManager {
	return NewTestConnectionManagerWithSecrets(store, secrets.NewMockStore())
}

// NewTestConnectionManagerWithSecrets allows tests to inject a custom secrets store.
func NewTestConnectionManagerWithSecrets(store *ConnectionStore, secretStore secrets.Store) *ConnectionManager {
	return NewConnectionManager(store, secretStore, zap.NewNop())
}
