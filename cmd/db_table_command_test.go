package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDBTableCommandDisplaysTableSchema(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	paths := &AppPaths{
		developerID: defaultDeveloperID,
		appName:     defaultAppName,
		userConfigDir: func() (string, error) {
			return configDir, nil
		},
	}

	dbPath := filepath.Join(configDir, defaultDeveloperID, defaultAppName, defaultSQLiteFileName)
	createDatabaseForTablesTest(t, dbPath)

	output := executeCommand(t, newDBCommandWithPaths(paths), "table", "episodes")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 schema lines, got %d: %q", len(lines), output)
	}

	if lines[0] != "name\ttype\tnotnull\tdefault\tpk" {
		t.Fatalf("unexpected schema header %q", lines[0])
	}

	if lines[1] != "id\tINTEGER\tfalse\tNULL\ttrue" {
		t.Fatalf("unexpected first schema row %q", lines[1])
	}

	if lines[2] != "name\tTEXT\ttrue\tNULL\tfalse" {
		t.Fatalf("unexpected second schema row %q", lines[2])
	}
}

func TestDBTableCommandFailsWhenTableDoesNotExist(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	paths := &AppPaths{
		developerID: defaultDeveloperID,
		appName:     defaultAppName,
		userConfigDir: func() (string, error) {
			return configDir, nil
		},
	}

	dbPath := filepath.Join(configDir, defaultDeveloperID, defaultAppName, defaultSQLiteFileName)
	createDatabaseForTablesTest(t, dbPath)

	_, err := executeCommandErr(newDBCommandWithPaths(paths), "table", "missing_table")
	if err == nil {
		t.Fatal("expected command error")
	}

	if !strings.Contains(err.Error(), `table "missing_table" does not exist`) {
		t.Fatalf("unexpected error %q", err.Error())
	}
}
