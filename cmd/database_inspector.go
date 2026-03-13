package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

type DatabaseInspector struct {
	path string
}

type DatabaseTable struct {
	Name     string
	RowCount int
}

type DatabaseColumn struct {
	Name         string
	Type         string
	NotNull      bool
	DefaultValue string
	PrimaryKey   bool
}

func NewDatabaseInspector(path string) *DatabaseInspector {
	return &DatabaseInspector{path: path}
}

func (i *DatabaseInspector) ListTables() ([]DatabaseTable, error) {
	db, err := i.open()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	rows, err := db.Query(`
		SELECT name, COALESCE(sql, '')
		FROM sqlite_schema
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("query database tables: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	type schemaTable struct {
		name string
		sql  string
	}

	schemaTables := make([]schemaTable, 0)
	ftsTables := make(map[string]struct{})
	for rows.Next() {
		var tableName string
		var tableSQL string
		if err := rows.Scan(&tableName, &tableSQL); err != nil {
			return nil, fmt.Errorf("scan database table name: %w", err)
		}

		schemaTables = append(schemaTables, schemaTable{
			name: tableName,
			sql:  tableSQL,
		})

		if isFTS5VirtualTableDefinition(tableSQL) {
			ftsTables[tableName] = struct{}{}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate database tables: %w", err)
	}

	tables := make([]DatabaseTable, 0, len(schemaTables))
	for _, schemaTable := range schemaTables {
		if isFTS5ImplementationTable(schemaTable.name, ftsTables) {
			continue
		}

		rowCount, err := countTableRows(db, schemaTable.name)
		if err != nil {
			return nil, err
		}

		tables = append(tables, DatabaseTable{
			Name:     schemaTable.name,
			RowCount: rowCount,
		})
	}

	return tables, nil
}

func isFTS5VirtualTableDefinition(definition string) bool {
	return strings.Contains(strings.ToLower(definition), "using fts5")
}

func isFTS5ImplementationTable(tableName string, ftsTables map[string]struct{}) bool {
	for ftsTableName := range ftsTables {
		if tableName == ftsTableName {
			return true
		}

		switch tableName {
		case ftsTableName + "_config", ftsTableName + "_content", ftsTableName + "_data", ftsTableName + "_docsize", ftsTableName + "_idx":
			return true
		}
	}

	return false
}

func (i *DatabaseInspector) DescribeTable(tableName string) ([]DatabaseColumn, error) {
	db, err := i.open()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	exists, err := tableExists(db, tableName)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("table %q does not exist", tableName)
	}

	query := fmt.Sprintf(`PRAGMA table_info("%s")`, strings.ReplaceAll(tableName, `"`, `""`))
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("describe table %q: %w", tableName, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	columns := make([]DatabaseColumn, 0)
	for rows.Next() {
		var columnID int
		var columnName string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&columnID, &columnName, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return nil, fmt.Errorf("scan column definition for table %q: %w", tableName, err)
		}

		columnDefault := "NULL"
		if defaultValue.Valid {
			columnDefault = defaultValue.String
		}

		columns = append(columns, DatabaseColumn{
			Name:         columnName,
			Type:         columnType,
			NotNull:      notNull == 1,
			DefaultValue: columnDefault,
			PrimaryKey:   primaryKey == 1,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate column definitions for table %q: %w", tableName, err)
	}

	return columns, nil
}

func countTableRows(db *sql.DB, tableName string) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, strings.ReplaceAll(tableName, `"`, `""`))

	var rowCount int
	if err := db.QueryRow(query).Scan(&rowCount); err != nil {
		return 0, fmt.Errorf("count rows for table %q: %w", tableName, err)
	}

	return rowCount, nil
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	var exists int
	if err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM sqlite_schema
			WHERE type = 'table' AND name = ? AND name NOT LIKE 'sqlite_%'
		)
	`, tableName).Scan(&exists); err != nil {
		return false, fmt.Errorf("check table %q existence: %w", tableName, err)
	}

	return exists == 1, nil
}

func (i *DatabaseInspector) open() (*sql.DB, error) {
	if err := ensureDatabaseFileExists(i.path); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", i.path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database %q: %w", i.path, err)
	}

	return db, nil
}

func ensureDatabaseFileExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("database file %q does not exist", path)
		}

		return fmt.Errorf("stat database file %q: %w", path, err)
	}

	return nil
}
