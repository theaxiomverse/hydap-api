package agglomerator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/theaxiomverse/hydap-api/pkg/modules/base"
	"github.com/theaxiomverse/hydap-api/pkg/modules/core"
	"github.com/theaxiomverse/hydap-api/pkg/vectors"
)

type ModuleConfig struct {
	NodeID        string   `json:"nodeID"`
	VectorDims    int      `json:"vectorDims"`
	SimThreshold  float64  `json:"simThreshold"`
	EnabledChains []string `json:"enabledChains"`
	LogPath       string   `json:"logPath"`
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

type AgglomeratorModule struct {
	base.BaseModule
	agglomerator  *Agglomerator
	config        *ModuleConfig
	configManager *core.ConfigManager
	metrics       *core.MetricsExporter
	logger        *core.ModuleLogger
	txManager     *core.TransactionManager
	mu            sync.RWMutex
	moduleState   base.ModuleState // renamed from state to moduleState
	state         base.ModuleState
}

// GetAgglomerator returns the underlying agglomerator instance
func (m *AgglomeratorModule) GetAgglomerator() *Agglomerator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.agglomerator
}

// GetConfig returns the current module configuration
func (m *AgglomeratorModule) GetConfig() *ModuleConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
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
		txManager: &core.TransactionManager{
			Txns: make(map[string]*core.Transaction),
		},
		moduleState: base.StateUninitialized,
	}
}

// State returns the current module state
func (m *AgglomeratorModule) GetState() base.ModuleState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.moduleState
}

// SetState is a helper method to update module state
func (m *AgglomeratorModule) SetState(state base.ModuleState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.moduleState = state
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

	if m.GetState() != base.StateRunning {
		txn.Status = "failed"
		return fmt.Errorf("module not in running state: %s", m.GetState())
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
