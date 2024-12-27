package agglomerator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theaxiomverse/hydap-api/pkg/modules/base"
	"github.com/theaxiomverse/hydap-api/pkg/modules/core"
	"github.com/theaxiomverse/hydap-api/pkg/vectors"
	"sync"
)

type ModuleConfig struct {
	NodeID        string   `json:"nodeID"`
	VectorDims    int      `json:"vectorDims"`
	SimThreshold  float64  `json:"simThreshold"`
	EnabledChains []string `json:"enabledChains"`
	LogPath       string   `json:"logPath"`
}

func NewAgglomeratorModule(
	configManager *core.ConfigManager,
	metrics *core.MetricsExporter,
	logger *core.ModuleLogger,
) *AgglomeratorModule {
	metadata := base.NewModuleMetadata(
		"blockchain_agglomerator",
		"1.0.0",
		"Blockchain Agglomerator with Vector-based Chain Analysis",
		"HyDAP Team",
		"MIT",
	)

	baseModule := base.CreateNewModule(metadata, nil).(*base.BaseModule)

	return &AgglomeratorModule{
		BaseModule:    *baseModule,
		configManager: configManager,
		metrics:       metrics,
		logger:        logger,
		txManager:     &core.TransactionManager{},
		state:         base.StateUninitialized,
	}
}

// Initialize implements Module interface
func (m *AgglomeratorModule) Initialize() error {
	if err := m.BaseModule.Initialize(); err != nil {
		return err
	}

	// Load configuration
	configData, err := m.configManager.GetConfig(m.Name())
	if err != nil {
		m.state = base.StateError
		return fmt.Errorf("failed to load config: %w", err)
	}

	var moduleConfig ModuleConfig
	if err := json.Unmarshal(configData, &moduleConfig); err != nil {
		m.state = base.StateError
		return fmt.Errorf("failed to parse config: %w", err)
	}
	m.config = &moduleConfig

	// Initialize agglomerator
	aggConfig := AgglomeratorConfig{
		NodeID:       moduleConfig.NodeID,
		VectorDims:   moduleConfig.VectorDims,
		SimThreshold: moduleConfig.SimThreshold,
	}
	m.agglomerator = NewAgglomerator(aggConfig)

	// Register metrics
	m.metrics.RegisterModule(m.Name())

	// Initialize chains
	for _, chainID := range moduleConfig.EnabledChains {
		chain := &Chain{
			ID:       chainID,
			Protocol: determineProtocol(chainID),
			StateVector: vectors.InfiniteVector{
				Generator: getDefaultGenerator(chainID),
			},
		}
		if err := m.agglomerator.RegisterChain(chain); err != nil {
			m.logger.Log(m.Name(), "ERROR", fmt.Sprintf("Failed to register chain %s: %v", chainID, err))
			m.state = base.StateError
			return err
		}
		m.logger.Log(m.Name(), "INFO", fmt.Sprintf("Registered chain: %s", chainID))
	}

	m.state = base.StateRunning
	return nil
}

// State returns the current module state
func (m *AgglomeratorModule) State() base.ModuleState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// SetState is a helper method to update module state
func (m *AgglomeratorModule) SetState(state base.ModuleState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = state
}

// ProcessTransaction handles a cross-chain transaction
func (m *AgglomeratorModule) ProcessTransaction(tx *Transaction) error {
	// Start transaction tracking
	txn := m.txManager.Begin(m.Name(), "process_transaction")
	defer func() {
		if txn.Status == "pending" {
			txn.Status = "completed"
		}
	}()

	m.logger.Log(m.Name(), "INFO", fmt.Sprintf("Processing transaction: %s", txn.ID))

	if m.State() != base.StateRunning {
		txn.Status = "failed"
		return fmt.Errorf("module not in running state: %s", m.State())
	}

	err := m.agglomerator.ProcessTransaction(context.Background(), tx)
	if err != nil {
		txn.Status = "failed"
		m.logger.Log(m.Name(), "ERROR", fmt.Sprintf("Transaction failed: %v", err))
		return err
	}

	m.logger.Log(m.Name(), "INFO", fmt.Sprintf("Transaction completed: %s", txn.ID))
	return nil
}
