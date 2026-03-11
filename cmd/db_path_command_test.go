package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDBPathCommandDisplaysDefaultSQLitePath(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	paths := &AppPaths{
		developerID: defaultDeveloperID,
		appName:     defaultAppName,
		userConfigDir: func() (string, error) {
			return configDir, nil
		},
	}

	output := executeCommand(t, newDBCommandWithPaths(paths), "path")
	expectedPath := filepath.Join(configDir, defaultDeveloperID, defaultAppName, defaultSQLiteFileName)
	if strings.TrimSpace(output) != expectedPath {
		t.Fatalf("expected %q, got %q", expectedPath, output)
	}
}
