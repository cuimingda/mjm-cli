package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	defaultDeveloperID    = "mingda.dev"
	defaultAppName        = "mjm"
	defaultSQLiteFileName = "data.sqlite"
)

type AppPaths struct {
	developerID   string
	appName       string
	userConfigDir func() (string, error)
}

func NewAppPaths(developerID string, appName string) *AppPaths {
	return &AppPaths{
		developerID:   developerID,
		appName:       appName,
		userConfigDir: os.UserConfigDir,
	}
}

func NewDefaultAppPaths() *AppPaths {
	return NewAppPaths(defaultDeveloperID, defaultAppName)
}

func (p *AppPaths) DefaultSQLitePath() (string, error) {
	configDir, err := p.userConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(configDir, p.developerID, p.appName, defaultSQLiteFileName), nil
}

func addSQLitePathFlag(command *cobra.Command, paths *AppPaths, dbPath *string) {
	defaultPath, _ := paths.DefaultSQLitePath()
	command.Flags().StringVar(dbPath, "db", defaultPath, fmt.Sprintf("SQLite database path (default %s)", defaultPath))
}

func resolveSQLitePath(paths *AppPaths, dbPath string) (string, error) {
	if dbPath != "" {
		return dbPath, nil
	}

	return paths.DefaultSQLitePath()
}
