package model_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestAppErrorNilReceiver(t *testing.T) {
	var err *model.AppError
	if err.Error() != "" {
		t.Fatalf("expected empty string, got %q", err.Error())
	}
}

func TestAppErrorConstructors(t *testing.T) {
	tests := []struct {
		name string
		err  *model.AppError
		code string
	}{
		{"invalid request", model.ErrInvalidRequest("bad"), "INVALID_REQUEST"},
		{"connection not found", model.ErrConnectionNotFound("id"), "CONNECTION_NOT_FOUND"},
		{"already open", model.ErrConnectionAlreadyOpen("id"), "CONNECTION_ALREADY_OPEN"},
		{"connection open for update", model.ErrConnectionOpen("id"), "CONNECTION_OPEN"},
		{"saved not found", model.ErrSavedNotFound("id"), "SAVED_NOT_FOUND"},
		{"connection failed", model.ErrConnectionFailed("msg"), "CONNECTION_FAILED"},
		{"database locked", model.ErrDatabaseLocked(), "DATABASE_LOCKED"},
		{"table not found", model.ErrTableNotFound("t"), "TABLE_NOT_FOUND"},
		{"sql error", model.ErrSQL("msg"), "SQL_ERROR"},
		{"result too large", model.ErrResultTooLarge(), "RESULT_TOO_LARGE"},
		{"read only", model.ErrReadOnly(), "READ_ONLY"},
		{"dialog cancelled", model.ErrDialogCancelled(), "DIALOG_CANCELLED"},
		{"export cancelled", model.ErrExportCancelled(), "EXPORT_CANCELLED"},
		{"export no stable key", model.ErrExportNoStableKey("t"), "EXPORT_NO_STABLE_KEY"},
		{"invalid path", model.ErrInvalidPath("msg"), "INVALID_PATH"},
		{"internal", model.ErrInternal("msg"), "INTERNAL_ERROR"},
		{"secrets vault locked", model.ErrSecretsVaultLocked(), "SECRETS_VAULT_LOCKED"},
		{"secrets vault not initialized", model.ErrSecretsVaultNotInitialized(), "SECRETS_VAULT_NOT_INITIALIZED"},
		{"secrets vault wrong password", model.ErrSecretsVaultWrongPassword(), "SECRETS_VAULT_WRONG_PASSWORD"},
		{"connection failed with reason", model.ErrConnectionFailedWithReason("failed", "timeout"), "CONNECTION_FAILED"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Code != tc.code {
				t.Fatalf("got code %s want %s", tc.err.Code, tc.code)
			}
			if tc.err.Error() == "" {
				t.Fatal("expected non-empty error string")
			}
		})
	}
}

func TestNewAppErrorWithDetails(t *testing.T) {
	err := model.NewAppError("CUSTOM", "detail", map[string]any{"k": "v"})
	if err.Code != "CUSTOM" || err.Details["k"] != "v" {
		t.Fatalf("unexpected error: %+v", err)
	}
}
