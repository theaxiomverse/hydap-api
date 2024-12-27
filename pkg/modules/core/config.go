// pkg/modules/core/config.go
package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

type ConfigManager struct {
	db       *sql.DB
	reloader *HotReloader
}

func NewConfigManager(dbPath string) (*ConfigManager, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := initConfigDB(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &ConfigManager{db: db}, nil
}

func initConfigDB(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS module_configs (
            module_name TEXT PRIMARY KEY,
            config JSON NOT NULL,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
	return err
}

func (cm *ConfigManager) SetConfig(module string, config json.RawMessage) error {
	if len(config) == 0 {
		return fmt.Errorf("empty configuration provided")
	}

	// Validate JSON
	var jsonCheck map[string]interface{}
	if err := json.Unmarshal(config, &jsonCheck); err != nil {
		return fmt.Errorf("invalid JSON configuration: %w", err)
	}

	_, err := cm.db.Exec(`
        INSERT OR REPLACE INTO module_configs (module_name, config, updated_at)
        VALUES (?, ?, CURRENT_TIMESTAMP)
    `, module, config)

	if err != nil {
		return fmt.Errorf("failed to store configuration: %w", err)
	}

	return nil
}

func (cm *ConfigManager) GetConfig(module string) (json.RawMessage, error) {
	var config json.RawMessage
	err := cm.db.QueryRow(`
        SELECT config FROM module_configs WHERE module_name = ?
    `, module).Scan(&config)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no configuration found for module: %s", module)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve configuration: %w", err)
	}

	return config, nil
}

// Close the database connection
func (cm *ConfigManager) Close() error {
	if cm.db != nil {
		return cm.db.Close()
	}
	return nil
}
