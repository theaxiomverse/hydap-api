// pkg/modules/core/config.go
package core

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
)

type ConfigManager struct {
	db       *sql.DB
	reloader *HotReloader
}

func NewConfigManager(dbPath string) (*ConfigManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := initConfigDB(db); err != nil {
		return nil, err
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
	_, err := cm.db.Exec(`
       INSERT OR REPLACE INTO module_configs (module_name, config)
       VALUES (?, ?)
   `, module, config)
	return err
}

func (cm *ConfigManager) GetConfig(module string) (json.RawMessage, error) {
	var config json.RawMessage
	err := cm.db.QueryRow(`
       SELECT config FROM module_configs WHERE module_name = ?
   `, module).Scan(&config)
	return config, err
}
