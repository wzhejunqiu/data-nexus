package model

import "fmt"

type AppError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewAppError(code, message string, details map[string]any) *AppError {
	return &AppError{Code: code, Message: message, Details: details}
}

func ErrInvalidRequest(message string) *AppError {
	return NewAppError("INVALID_REQUEST", message, nil)
}

func ErrConnectionNotFound(id string) *AppError {
	return NewAppError("CONNECTION_NOT_FOUND", "connection not found or not open", map[string]any{"connectionId": id})
}

func ErrConnectionAlreadyOpen(id string) *AppError {
	return NewAppError("CONNECTION_ALREADY_OPEN", "connection is already open", map[string]any{"connectionId": id})
}

func ErrConnectionOpen(id string) *AppError {
	return NewAppError("CONNECTION_OPEN", "connection must be closed before updating settings", map[string]any{"connectionId": id})
}

func ErrSavedNotFound(id string) *AppError {
	return NewAppError("SAVED_NOT_FOUND", "saved connection not found", map[string]any{"connectionId": id})
}

func ErrConnectionFailed(message string) *AppError {
	return NewAppError("CONNECTION_FAILED", message, nil)
}

func ErrDatabaseLocked() *AppError {
	return NewAppError("DATABASE_LOCKED", "database is locked", nil)
}

func ErrTableNotFound(name string) *AppError {
	return NewAppError("TABLE_NOT_FOUND", "table or view not found", map[string]any{"tableName": name})
}

func ErrSQL(message string) *AppError {
	return NewAppError("SQL_ERROR", message, nil)
}

func ErrResultTooLarge() *AppError {
	return NewAppError("RESULT_TOO_LARGE", "result set exceeds maximum row limit", nil)
}

func ErrReadOnly() *AppError {
	return NewAppError("READ_ONLY", "connection is read-only", nil)
}

func ErrDialogCancelled() *AppError {
	return NewAppError("DIALOG_CANCELLED", "dialog cancelled", nil)
}

func ErrExportCancelled() *AppError {
	return NewAppError("EXPORT_CANCELLED", "export cancelled", nil)
}

func ErrExportNoStableKey(tableName string) *AppError {
	return NewAppError("EXPORT_NO_STABLE_KEY", "table has no stable sort key for export", map[string]any{"tableName": tableName})
}

func ErrInvalidPath(message string) *AppError {
	return NewAppError("INVALID_PATH", message, nil)
}

func ErrInternal(message string) *AppError {
	return NewAppError("INTERNAL_ERROR", message, nil)
}

func ErrSecretsVaultLocked() *AppError {
	return NewAppError("SECRETS_VAULT_LOCKED", "secrets vault is locked", nil)
}

func ErrSecretsVaultNotInitialized() *AppError {
	return NewAppError("SECRETS_VAULT_NOT_INITIALIZED", "secrets vault is not initialized", nil)
}

func ErrSecretsVaultWrongPassword() *AppError {
	return NewAppError("SECRETS_VAULT_WRONG_PASSWORD", "incorrect vault master password", nil)
}

func ErrConnectionFailedWithReason(message, reason string) *AppError {
	return NewAppError("CONNECTION_FAILED", message, map[string]any{"reason": reason})
}
