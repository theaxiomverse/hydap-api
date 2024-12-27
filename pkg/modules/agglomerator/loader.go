package agglomerator

import (
	"encoding/json"
	"fmt"
	"github.com/theaxiomverse/hydap-api/pkg/modules/base"
	"github.com/theaxiomverse/hydap-api/pkg/modules/core"
)

type AgglomeratorLoader struct {
	configManager *core.ConfigManager
	metrics       *core.MetricsExporter
	logger        *core.ModuleLogger
}

func NewAgglomeratorLoader(
	configManager *core.ConfigManager,
	metrics *core.MetricsExporter,
	logger *core.ModuleLogger,
) *AgglomeratorLoader {
	return &AgglomeratorLoader{
		configManager: configManager,
		metrics:       metrics,
		logger:        logger,
	}
}

func (l *AgglomeratorLoader) LoadFromConfig(config base.ModuleConfig) (base.Module, error) {
	// Validate config
	if config.Name != "blockchain_agglomerator" {
		return nil, fmt.Errorf("invalid module name: %s", config.Name)
	}

	// Create module with dependencies
	module := NewAgglomeratorModule(
		l.configManager,
		l.metrics,
		l.logger,
	)

	// Store module config
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := l.configManager.SetConfig(config.Name, configJSON); err != nil {
		return nil, fmt.Errorf("failed to store config: %w", err)
	}

	return module, nil
}

func (l *AgglomeratorLoader) Load(path string) (base.Module, error) {
	// Load module configuration from file path
	// This could be implemented if needed
	return nil, fmt.Errorf("file-based loading not implemented")
}
