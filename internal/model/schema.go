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
	ConnectionID   string      `json:"connectionId"`
	TableName      string      `json:"tableName"`
	Page           int         `json:"page"`
	PageSize       int         `json:"pageSize"`
	Sort           string      `json:"sort"`
	Order          SortOrder   `json:"order"`
	SkipTotalCount bool        `json:"skipTotalCount,omitempty"`
	Filters        []RowFilter `json:"filters,omitempty"`
	Search         string      `json:"search,omitempty"`
}

type BrowseOptions struct {
	Page           int
	PageSize       int
	Sort           string
	Order          SortOrder
	SkipTotalCount bool
	Filters        []RowFilter
	Search         string
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

const (
	ProfileSampleLimit      = 10000
	LowCardinalityThreshold = 20
	ProfileTopValuesLimit   = 10
)

type ValueCount struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type ColumnProfile struct {
	Name             string       `json:"name"`
	DistinctCount    *int64       `json:"distinctCount"`
	NullPercent      *float64     `json:"nullPercent"`
	MinValue         *string      `json:"minValue"`
	MaxValue         *string      `json:"maxValue"`
	TopValues        []ValueCount `json:"topValues,omitempty"`
	IsLowCardinality bool         `json:"isLowCardinality"`
}

type TableProfile struct {
	TableName   string          `json:"tableName"`
	SampledRows int64           `json:"sampledRows"`
	TotalRows   *int64          `json:"totalRows"`
	IsSampled   bool            `json:"isSampled"`
	Columns     []ColumnProfile `json:"columns"`
}

type FilterOperator string

const (
	FilterEq        FilterOperator = "eq"
	FilterNe        FilterOperator = "ne"
	FilterGt        FilterOperator = "gt"
	FilterGte       FilterOperator = "gte"
	FilterLt        FilterOperator = "lt"
	FilterLte       FilterOperator = "lte"
	FilterLike      FilterOperator = "like"
	FilterIsNull    FilterOperator = "is_null"
	FilterIsNotNull FilterOperator = "is_not_null"
	FilterIn        FilterOperator = "in"
)

type RowFilter struct {
	Column   string         `json:"column"`
	Operator FilterOperator `json:"operator"`
	Value    *string        `json:"value,omitempty"`
	Values   []string       `json:"values,omitempty"`
}

type FTSInfo struct {
	Enabled      bool   `json:"enabled"`
	Schema       string `json:"schema,omitempty"`
	FTSTableName string `json:"ftsTableName,omitempty"`
	ContentTable string `json:"contentTable,omitempty"`
}
