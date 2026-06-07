package model

type TableType string

const (
	TableTypeTable TableType = "table"
	TableTypeView  TableType = "view"
)

type TableInfo struct {
	Name     string    `json:"name"`
	Type     TableType `json:"type"`
	Schema   *string   `json:"schema"`
	RowCount *int64    `json:"rowCount"`
}

type TableList struct {
	Items []TableInfo `json:"items"`
}

type ColumnInfo struct {
	Name         string  `json:"name"`
	DataType     string  `json:"dataType"`
	NativeType   *string `json:"nativeType,omitempty"`
	Nullable     bool    `json:"nullable"`
	PrimaryKey   bool    `json:"primaryKey"`
	DefaultValue *string `json:"defaultValue"`
	Position     int     `json:"position"`
}

type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Primary bool     `json:"primary"`
}

type TableSchema struct {
	Name    string       `json:"name"`
	Type    TableType    `json:"type"`
	Schema  *string      `json:"schema"`
	Columns []ColumnInfo `json:"columns"`
	Indexes []IndexInfo  `json:"indexes"`
}

type ColumnMeta struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type BrowseRowsRequest struct {
	ConnectionID string    `json:"connectionId"`
	TableName    string    `json:"tableName"`
	Page         int       `json:"page"`
	PageSize     int       `json:"pageSize"`
	Sort         string    `json:"sort"`
	Order        SortOrder `json:"order"`
}

type BrowseOptions struct {
	Page     int
	PageSize int
	Sort     string
	Order    SortOrder
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalRows  int64 `json:"totalRows"`
	TotalPages int   `json:"totalPages"`
}

type PaginatedTableData struct {
	Columns    []ColumnMeta     `json:"columns"`
	Rows       []map[string]any `json:"rows"`
	Pagination PaginationMeta   `json:"pagination"`
}
