package model

type StableRowKey struct {
	Columns []string // ORDER BY columns (composite PK may have multiple)
	Source  string   // "primary_key" | "unique_index" | "rowid" (SQLite only)
}

type TableExportOptions struct {
	BatchSize int // default 1000
}

func (o TableExportOptions) NormalizedBatchSize() int {
	if o.BatchSize <= 0 {
		return 1000
	}
	return o.BatchSize
}

type TableExportBatch struct {
	Rows    []map[string]any
	HasMore bool
}
