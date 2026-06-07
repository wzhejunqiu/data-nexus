package model

const MaxBatchCellUpdates = 200

type CellChange struct {
	ColumnName string         `json:"columnName"`
	PrimaryKey map[string]any `json:"primaryKey"`
	NewValue   any            `json:"newValue"`
}

type UpdateCellsBatchRequest struct {
	ConnectionID string       `json:"connectionId"`
	TableName    string       `json:"tableName"`
	Changes      []CellChange `json:"changes"`
}

type UpdateCellsBatchResult struct {
	UpdatedCount int `json:"updatedCount"`
}
