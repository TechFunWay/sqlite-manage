package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Database struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
	db   *sql.DB
}

var (
	mu          sync.RWMutex
	databases   = make(map[string]*Database)
	activeDB    *Database
	connections = make(map[string]*sql.DB)
)

type Table struct {
	Name string `json:"name"`
	Rows int64  `json:"rows"`
}

type Info struct {
	ID            string  `json:"id"`
	Path          string  `json:"path"`
	Name          string  `json:"name"`
	Size          int64   `json:"size"`
	SQLiteVersion string  `json:"sqliteVersion"`
	CreatedAt     string  `json:"createdAt"`
	ModifiedAt    string  `json:"modifiedAt"`
	TableCount    int     `json:"tableCount"`
	TotalRows     int64   `json:"totalRows"`
	Active        bool    `json:"active"`
	Tables        []Table `json:"tables"`
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

func getDB() *sql.DB {
	mu.RLock()
	defer mu.RUnlock()
	if activeDB == nil {
		return nil
	}
	return activeDB.db
}

func getDBByID(id string) *sql.DB {
	mu.RLock()
	defer mu.RUnlock()
	if db, ok := connections[id]; ok {
		return db
	}
	return nil
}

func Open(dbPath string) (*Database, error) {
	mu.Lock()
	defer mu.Unlock()

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// 检查是否已存在相同路径的数据库
	for _, db := range databases {
		if db.Path == absPath {
			activeDB = db
			return db, nil
		}
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %s", absPath)
	}

	sqlDB, err := sql.Open("sqlite", absPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	database := &Database{
		ID:   id,
		Name: filepath.Base(absPath),
		Path: absPath,
	}

	stat, _ := os.Stat(absPath)
	if stat != nil {
		database.Size = stat.Size()
	}
	database.db = sqlDB

	databases[id] = database
	connections[id] = sqlDB
	activeDB = database

	return database, nil
}

func OpenOrCreate(dbPath string) (*Database, error) {
	mu.Lock()
	defer mu.Unlock()

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// 检查是否已存在相同路径的数据库
	for _, db := range databases {
		if db.Path == absPath {
			activeDB = db
			return db, nil
		}
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		file, err := os.Create(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		file.Close()
	}

	sqlDB, err := sql.Open("sqlite", absPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	database := &Database{
		ID:   id,
		Name: filepath.Base(absPath),
		Path: absPath,
	}

	stat, _ := os.Stat(absPath)
	if stat != nil {
		database.Size = stat.Size()
	}
	database.db = sqlDB

	databases[id] = database
	connections[id] = sqlDB
	activeDB = database

	return database, nil
}

func Close(id string) error {
	mu.Lock()
	defer mu.Unlock()

	db, ok := databases[id]
	if !ok {
		return fmt.Errorf("database not found: %s", id)
	}

	err := db.db.Close()
	delete(databases, id)
	delete(connections, id)

	if activeDB != nil && activeDB.ID == id {
		activeDB = nil
		for _, d := range databases {
			activeDB = d
			break
		}
	}

	return err
}

func SetActive(id string) (*Database, error) {
	mu.Lock()
	defer mu.Unlock()

	db, ok := databases[id]
	if !ok {
		return nil, fmt.Errorf("database not found: %s", id)
	}

	activeDB = db
	return db, nil
}

func GetActive() *Database {
	mu.RLock()
	defer mu.RUnlock()
	return activeDB
}

func IsOpen() bool {
	mu.RLock()
	defer mu.RUnlock()
	return activeDB != nil
}

func GetAllDatabases() []Info {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]Info, 0, len(databases))
	for _, db := range databases {
		info := getDatabaseInfo(db)
		info.Active = activeDB != nil && activeDB.ID == db.ID
		result = append(result, info)
	}
	return result
}

func GetPath() string {
	mu.RLock()
	defer mu.RUnlock()
	if activeDB == nil {
		return ""
	}
	return activeDB.Path
}

func getDatabaseInfo(db *Database) Info {
	info := Info{
		ID:   db.ID,
		Path: db.Path,
		Name: db.Name,
		Size: db.Size,
	}

	var version string
	db.db.QueryRow("SELECT sqlite_version()").Scan(&version)
	info.SQLiteVersion = version

	stat, _ := os.Stat(db.Path)
	if stat != nil {
		info.ModifiedAt = stat.ModTime().Format(time.RFC3339)
		info.CreatedAt = stat.ModTime().Format(time.RFC3339)
	}

	rows, err := db.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err == nil {
		defer rows.Close()
		var tableCount int
		var totalRows int64
		var tables []Table
		for rows.Next() {
			var name string
			rows.Scan(&name)
			tableCount++
			var count int64
			db.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", name)).Scan(&count)
			totalRows += count
			tables = append(tables, Table{Name: name, Rows: count})
		}
		info.TableCount = tableCount
		info.TotalRows = totalRows
		info.Tables = tables
	}

	return info
}

func GetInfo() (*Info, error) {
	db := getActiveDB()
	if db == nil {
		return nil, fmt.Errorf("no database connected")
	}
	info := getDatabaseInfo(db)
	return &info, nil
}

func getActiveDB() *Database {
	mu.RLock()
	defer mu.RUnlock()
	return activeDB
}

func GetTables() ([]string, error) {
	db := getDB()
	if db == nil {
		return nil, fmt.Errorf("no database connected")
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}

	return tables, nil
}

func GetTableRowCount(tableName string) (int64, error) {
	db := getDB()
	if db == nil {
		return 0, fmt.Errorf("no database connected")
	}

	var count int64
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", tableName)).Scan(&count)
	return count, err
}

func GetSchema(tableName string) (*Schema, error) {
	db := getDB()
	if db == nil {
		return nil, fmt.Errorf("no database connected")
	}

	schema := &Schema{
		Name:    tableName,
		Columns: []Column{},
		Indexes: []Index{},
	}

	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(\"%s\")", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col Column
		var cid int
		var dflt *string
		var notnull int
		var pk int
		err := rows.Scan(&cid, &col.Name, &col.Type, &notnull, &dflt, &pk)
		if err != nil {
			return nil, err
		}
		col.Nullable = notnull == 0
		col.PrimaryKey = pk > 0
		col.DefaultValue = dflt
		schema.Columns = append(schema.Columns, col)
	}

	indexRows, err := db.Query(fmt.Sprintf("PRAGMA index_list(\"%s\")", tableName))
	if err != nil {
		return nil, err
	}
	defer indexRows.Close()

	for indexRows.Next() {
		var idx Index
		var seq, unique int
		var name, origin, partial string
		err := indexRows.Scan(&seq, &name, &unique, &origin, &partial)
		if err != nil {
			return nil, err
		}
		idx.Name = name
		idx.Unique = unique == 1
		idx.Table = tableName

		idxInfoRows, err := db.Query(fmt.Sprintf("PRAGMA index_info(\"%s\")", idx.Name))
		if err != nil {
			return nil, err
		}

		for idxInfoRows.Next() {
			var seqno, cid int
			var colName string
			idxInfoRows.Scan(&seqno, &cid, &colName)
			idx.Columns = append(idx.Columns, colName)
		}
		idxInfoRows.Close()

		schema.Indexes = append(schema.Indexes, idx)
	}

	return schema, nil
}

func GetData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	db := getDB()
	if db == nil {
		return nil, 0, fmt.Errorf("no database connected")
	}

	var total int64
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", tableName)).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM \"%s\" LIMIT ? OFFSET ?", tableName)
	rows, err := db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, 0, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, total, nil
}

func InsertRow(tableName string, data map[string]interface{}) (int64, error) {
	db := getDB()
	if db == nil {
		return 0, fmt.Errorf("no database connected")
	}

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		columns = append(columns, fmt.Sprintf("\"%s\"", col))
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s)",
		tableName, joinStrings(columns, ", "), joinStrings(placeholders, ", "))

	result, err := db.Exec(query, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func UpdateRow(tableName string, primaryKey string, pkValue interface{}, data map[string]interface{}) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("no database connected")
	}

	updates := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+1)

	for col, val := range data {
		updates = append(updates, fmt.Sprintf("\"%s\" = ?", col))
		values = append(values, val)
	}
	values = append(values, pkValue)

	query := fmt.Sprintf("UPDATE \"%s\" SET %s WHERE \"%s\" = ?",
		tableName, joinStrings(updates, ", "), primaryKey)

	_, err := db.Exec(query, values...)
	return err
}

func DeleteRow(tableName string, primaryKey string, pkValue interface{}) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("no database connected")
	}

	query := fmt.Sprintf("DELETE FROM \"%s\" WHERE \"%s\" = ?", tableName, primaryKey)
	_, err := db.Exec(query, pkValue)
	return err
}

func CreateTable(tableName string, columns []Column) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("no database connected")
	}

	colDefs := make([]string, 0, len(columns))
	for _, col := range columns {
		colDef := fmt.Sprintf("\"%s\" %s", col.Name, col.Type)
		if col.PrimaryKey {
			colDef += " PRIMARY KEY"
		} else if !col.Nullable {
			colDef += " NOT NULL"
		}
		if col.DefaultValue != nil {
			colDef += fmt.Sprintf(" DEFAULT %s", *col.DefaultValue)
		}
		colDefs = append(colDefs, colDef)
	}

	query := fmt.Sprintf("CREATE TABLE \"%s\" (%s)", tableName, joinStrings(colDefs, ", "))
	_, err := db.Exec(query)
	return err
}

func DropTable(tableName string) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("no database connected")
	}

	query := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", tableName)
	_, err := db.Exec(query)
	return err
}

func RenameTable(oldName, newName string) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("no database connected")
	}

	query := fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\"", oldName, newName)
	_, err := db.Exec(query)
	return err
}

func AddColumn(tableName string, column Column) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("no database connected")
	}

	colDef := fmt.Sprintf("\"%s\" %s", column.Name, column.Type)
	if !column.Nullable {
		colDef += " NOT NULL"
	}
	if column.DefaultValue != nil {
		colDef += fmt.Sprintf(" DEFAULT %s", *column.DefaultValue)
	}

	query := fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN %s", tableName, colDef)
	_, err := db.Exec(query)
	return err
}

func DropColumn(tableName, columnName string) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("no database connected")
	}

	columns, err := GetColumnsForDrop(tableName)
	if err != nil {
		return err
	}

	newColumns := make([]Column, 0)
	for _, col := range columns {
		if col.Name != columnName {
			newColumns = append(newColumns, col)
		}
	}

	tempTable := tableName + "_temp"

	oldCols := make([]string, 0)
	newCols := make([]string, 0)
	for _, col := range columns {
		oldCols = append(oldCols, fmt.Sprintf("\"%s\"", col.Name))
	}
	for _, col := range newColumns {
		newCols = append(newCols, fmt.Sprintf("\"%s\"", col.Name))
	}

	_, err = db.Exec(fmt.Sprintf("CREATE TABLE \"%s\" AS SELECT %s FROM \"%s\"", tempTable, joinStrings(oldCols, ", "), tableName))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("DROP TABLE \"%s\"", tableName))
	if err != nil {
		return err
	}

	colDefs := make([]string, 0, len(newColumns))
	for _, col := range newColumns {
		colDef := fmt.Sprintf("\"%s\" %s", col.Name, col.Type)
		if col.PrimaryKey {
			colDef += " PRIMARY KEY"
		} else if !col.Nullable {
			colDef += " NOT NULL"
		}
		if col.DefaultValue != nil {
			colDef += fmt.Sprintf(" DEFAULT %s", *col.DefaultValue)
		}
		colDefs = append(colDefs, colDef)
	}

	_, err = db.Exec(fmt.Sprintf("CREATE TABLE \"%s\" (%s)", tableName, joinStrings(colDefs, ", ")))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("INSERT INTO \"%s\" (%s) SELECT %s FROM \"%s\"", tableName, joinStrings(newCols, ", "), joinStrings(oldCols, ", "), tempTable))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("DROP TABLE \"%s\"", tempTable))
	return err
}

func GetColumnsForDrop(tableName string) ([]Column, error) {
	db := getDB()
	if db == nil {
		return nil, fmt.Errorf("no database connected")
	}

	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(\"%s\")", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var cid int
		var dflt *string
		var notnull, pk int
		rows.Scan(&cid, &col.Name, &col.Type, &notnull, &dflt, &pk)
		col.Nullable = notnull == 0
		col.PrimaryKey = pk > 0
		col.DefaultValue = dflt
		columns = append(columns, col)
	}
	return columns, nil
}

func CreateIndex(tableName string, indexName string, columns []string, unique bool) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("no database connected")
	}

	colList := make([]string, len(columns))
	for i, col := range columns {
		colList[i] = fmt.Sprintf("\"%s\"", col)
	}

	query := fmt.Sprintf("CREATE %s INDEX \"%s\" ON \"%s\" (%s)",
		map[bool]string{true: "UNIQUE", false: ""}[unique], indexName, tableName, joinStrings(colList, ", "))

	_, err := db.Exec(query)
	return err
}

func DropIndex(indexName string) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("no database connected")
	}

	query := fmt.Sprintf("DROP INDEX IF EXISTS \"%s\"", indexName)
	_, err := db.Exec(query)
	return err
}

func ExecuteQuery(sql string) ([]map[string]interface{}, error) {
	db := getDB()
	if db == nil {
		return nil, fmt.Errorf("no database connected")
	}

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, nil
}

func ExecuteNonQuery(sql string) (int64, error) {
	db := getDB()
	if db == nil {
		return 0, fmt.Errorf("no database connected")
	}

	result, err := db.Exec(sql)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func GetPrimaryKey(tableName string) (string, error) {
	db := getDB()
	if db == nil {
		return "", fmt.Errorf("no database connected")
	}

	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(\"%s\")", tableName))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, coltype string
		var notnull, pk int
		var dflt *string
		rows.Scan(&cid, &name, &coltype, &notnull, &dflt, &pk)
		if pk > 0 {
			return name, nil
		}
	}
	return "", nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
