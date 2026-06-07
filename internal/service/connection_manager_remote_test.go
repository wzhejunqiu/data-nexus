package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
	"github.com/wzhejunqiu/data-nexus/internal/secrets"
	"github.com/wzhejunqiu/data-nexus/internal/service"
)

func TestConnectionManagerCreateRemoteStoresPassword(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mockSecrets := secrets.NewMockStore()
	mgr := service.NewTestConnectionManagerWithSecrets(store, mockSecrets)

	ctx := context.Background()
	saved, err := mgr.CreateRemoteConnection(ctx, model.RemoteConnectRequest{
		Type:     model.DriverTypePostgres,
		Name:     "local-pg",
		Password: "secret",
		Postgres: &model.PostgresConfig{
			Host:     "127.0.0.1",
			Port:     5432,
			Database: "testdb",
			User:     "test",
			Schema:   "public",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if mockSecrets.Passwords[saved.ID] != "secret" {
		t.Fatalf("expected password stored, got %q", mockSecrets.Passwords[saved.ID])
	}
	if saved.SecretsBackend != model.SecretsBackendKeychain {
		t.Fatalf("expected keychain backend, got %q", saved.SecretsBackend)
	}
}

func TestConnectionManagerOpenRemoteRequiresVaultUnlock(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mockSecrets := secrets.NewMockVaultStore()
	mockSecrets.Unlocked = true
	mgr := service.NewTestConnectionManagerWithSecrets(store, mockSecrets)

	ctx := context.Background()
	saved, err := mgr.CreateRemoteConnection(ctx, model.RemoteConnectRequest{
		Type:     model.DriverTypePostgres,
		Name:     "vault-pg",
		Password: "secret",
		Postgres: &model.PostgresConfig{
			Host:     "127.0.0.1",
			Port:     5432,
			Database: "testdb",
			User:     "test",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	mockSecrets.Unlocked = false
	_, err = mgr.OpenConnection(ctx, saved.ID)
	if err == nil {
		t.Fatal("expected vault locked error")
	}
	appErr, ok := err.(*model.AppError)
	if !ok || appErr.Code != "SECRETS_VAULT_LOCKED" {
		t.Fatalf("expected SECRETS_VAULT_LOCKED, got %v", err)
	}

	mockSecrets.Unlocked = true
	_, err = mgr.OpenConnection(ctx, saved.ID)
	if err == nil {
		return
	}
	// Without a live postgres instance the open may fail at connect; vault path must pass first.
	if appErr, ok := err.(*model.AppError); ok && appErr.Code == "SECRETS_VAULT_LOCKED" {
		t.Fatalf("vault should be unlocked: %v", err)
	}
}

func TestConnectionManagerRemoveRemoteDeletesPassword(t *testing.T) {
	dir := t.TempDir()
	store, err := service.NewConnectionStore(filepath.Join(dir, "connections.json"))
	if err != nil {
		t.Fatal(err)
	}
	mockSecrets := secrets.NewMockStore()
	mgr := service.NewTestConnectionManagerWithSecrets(store, mockSecrets)

	ctx := context.Background()
	saved, err := mgr.CreateRemoteConnection(ctx, model.RemoteConnectRequest{
		Type:     model.DriverTypeMySQL,
		Name:     "local-mysql",
		Password: "secret",
		MySQL: &model.MySQLConfig{
			Host:     "127.0.0.1",
			Port:     3306,
			Database: "testdb",
			User:     "test",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.RemoveConnection(ctx, saved.ID); err != nil {
		t.Fatal(err)
	}
	if _, ok := mockSecrets.Passwords[saved.ID]; ok {
		t.Fatal("expected password deleted")
	}
}
