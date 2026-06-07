package model

const CSVExportWarnRows = 10_000

// MaxCSVExportRows is deprecated for full-table export (no hard cap). Kept for reference in docs/tests.
const MaxCSVExportRows = MaxQueryRows

type CSVFormatOptions struct {
	Delimiter        string `json:"delimiter"`
	QuoteChar        string `json:"quoteChar"`
	HasHeader        bool   `json:"hasHeader"`
	NullValue        string `json:"nullValue"`
	LineEnding       string `json:"lineEnding"`
	Encoding         string `json:"encoding"`
	CommentChar      string `json:"commentChar"`
	LazyQuotes       bool   `json:"lazyQuotes"`
	TrimLeadingSpace bool   `json:"trimLeadingSpace"`
}

func DefaultCSVFormat() CSVFormatOptions {
	return CSVFormatOptions{
		Delimiter:  ",",
		QuoteChar:  `"`,
		HasHeader:  true,
		LineEnding: "crlf",
		Encoding:   "utf-8",
	}
}

func (o CSVFormatOptions) Normalized() CSVFormatOptions {
	n := DefaultCSVFormat()
	if o.Delimiter != "" {
		n.Delimiter = o.Delimiter
	}
	if o.QuoteChar != "" {
		n.QuoteChar = o.QuoteChar
	}
	n.HasHeader = o.HasHeader
	if o.LineEnding != "" {
		n.LineEnding = o.LineEnding
	}
	if o.Encoding != "" {
		n.Encoding = o.Encoding
	}
	n.NullValue = o.NullValue
	n.CommentChar = o.CommentChar
	n.LazyQuotes = o.LazyQuotes
	n.TrimLeadingSpace = o.TrimLeadingSpace
	return n
}

type ExportTableCSVRequest struct {
	ConnectionID string           `json:"connectionId"`
	TableName    string           `json:"tableName"`
	Format       CSVFormatOptions `json:"format"`
	Columns      []string         `json:"columns,omitempty"` // empty = all table columns
	DefaultPath  string           `json:"defaultPath,omitempty"`
	ExportID     string           `json:"exportId,omitempty"`
}

type ExportProgress struct {
	ExportID string `json:"exportId"`
	Exported int    `json:"exported"`
}

type ParseCSVPreviewRequest struct {
	FilePath string           `json:"filePath"`
	MaxRows  int              `json:"maxRows"`
	Format   CSVFormatOptions `json:"format"`
}

type CSVPreview struct {
	Headers   []string         `json:"headers"`
	Rows      []map[string]any `json:"rows"`
	RowCount  int              `json:"rowCount"`
	HasHeader bool             `json:"hasHeader"`
}

type ImportColumnSpec struct {
	DataType   string `json:"dataType"`
	PrimaryKey bool   `json:"primaryKey"`
}

type ImportCSVRequest struct {
	ConnectionID    string                      `json:"connectionId"`
	TargetTable     string                      `json:"targetTable"`
	NewTableName    string                      `json:"newTableName"`
	Mode            string                      `json:"mode"`
	ColumnMap       map[string]string           `json:"columnMap"`
	NewTableColumns map[string]ImportColumnSpec `json:"newTableColumns,omitempty"`
	FilePath        string                      `json:"filePath"`
	UpsertKeys      []string                    `json:"upsertKeys"`
	Format          CSVFormatOptions            `json:"format"`
}

type ImportCSVResult struct {
	RowsInserted int `json:"rowsInserted"`
	RowsUpdated  int `json:"rowsUpdated"`
}
