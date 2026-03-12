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

func NewDatabaseInspector(path string) *DatabaseInspector {
	return &DatabaseInspector{path: path}
}

func (i *DatabaseInspector) ListTables() ([]DatabaseTable, error) {
	if err := ensureDatabaseFileExists(i.path); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", i.path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database %q: %w", i.path, err)
	}
	defer func() {
		_ = db.Close()
	}()

	rows, err := db.Query(`
		SELECT name
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

	tables := make([]DatabaseTable, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("scan database table name: %w", err)
		}

		rowCount, err := countTableRows(db, tableName)
		if err != nil {
			return nil, err
		}

		tables = append(tables, DatabaseTable{
			Name:     tableName,
			RowCount: rowCount,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate database tables: %w", err)
	}

	return tables, nil
}

func countTableRows(db *sql.DB, tableName string) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, strings.ReplaceAll(tableName, `"`, `""`))

	var rowCount int
	if err := db.QueryRow(query).Scan(&rowCount); err != nil {
		return 0, fmt.Errorf("count rows for table %q: %w", tableName, err)
	}

	return rowCount, nil
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
