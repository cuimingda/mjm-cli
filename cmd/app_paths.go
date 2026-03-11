package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

const defaultSQLiteFileName = "data.sqlite"

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

func (p *AppPaths) DefaultSQLitePath() (string, error) {
	configDir, err := p.userConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(configDir, p.developerID, p.appName, defaultSQLiteFileName), nil
}
