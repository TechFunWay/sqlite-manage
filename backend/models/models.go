package models

type DatabaseInfo struct {
	Path          string `json:"path"`
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	SQLiteVersion string `json:"sqliteVersion"`
	CreatedAt     string `json:"createdAt"`
	ModifiedAt    string `json:"modifiedAt"`
	TableCount    int    `json:"tableCount"`
	TotalRows     int64  `json:"totalRows"`
}

type Column struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Nullable     bool    `json:"nullable"`
	DefaultValue *string `json:"defaultValue"`
	PrimaryKey   bool    `json:"primaryKey"`
}

type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Table   string   `json:"table"`
}

type Schema struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
	Indexes []Index  `json:"indexes"`
}

type TableData struct {
	Data     []map[string]interface{} `json:"data"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
}

type RowUpdate struct {
	Data map[string]interface{} `json:"data"`
}

type CreateTableRequest struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

type CreateIndexRequest struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

type QueryRequest struct {
	SQL string `json:"sql"`
}

type QueryResult struct {
	Columns      []string                 `json:"columns"`
	Data         []map[string]interface{} `json:"data"`
	RowsAffected int64                    `json:"rowsAffected,omitempty"`
}
