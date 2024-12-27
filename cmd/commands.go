package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/theaxiomverse/hydap-api/pkg/modules/agglomerator"
	"github.com/theaxiomverse/hydap-api/pkg/modules/core"
	"github.com/theaxiomverse/hydap-api/pkg/vectors"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the agglomerator service",
	Long:  `Start the blockchain agglomerator service with the specified configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		return startService(configFile)
	},
}

var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Manage blockchain chains",
	Long:  `Add, remove, list, and manage blockchain chains in the agglomerator.`,
}

var chainAddCmd = &cobra.Command{
	Use:   "add [chain-id] [endpoint]",
	Short: "Add a new blockchain chain",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		chainID := args[0]
		endpoint := args[1]
		protocol, _ := cmd.Flags().GetString("protocol")
		return addChain(chainID, endpoint, protocol)
	},
}

var chainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered chains",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listChains()
	},
}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "Manage transactions",
	Long:  `Create, monitor, and manage cross-chain transactions.`,
}

var txCreateCmd = &cobra.Command{
	Use:   "create [from-chain] [to-chain]",
	Short: "Create a new cross-chain transaction",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fromChain := args[0]
		toChain := args[1]
		data, _ := cmd.Flags().GetString("data")
		return createTransaction(fromChain, toChain, []byte(data))
	},
}

func init() {
	// Chain command flags
	chainAddCmd.Flags().StringP("protocol", "p", "", "chain protocol (eth, sol, etc)")
	chainCmd.AddCommand(chainAddCmd)
	chainCmd.AddCommand(chainListCmd)

	// Transaction command flags
	txCreateCmd.Flags().StringP("data", "d", "", "transaction data")
	txCmd.AddCommand(txCreateCmd)
}

func startService(configFile string) error {
	// Read config file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML config
	var config struct {
		Modules struct {
			BlockchainAgglomerator map[string]interface{} `yaml:"blockchain_agglomerator"`
		} `yaml:"modules"`
	}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Initialize core components
	configManager, err := core.NewConfigManager("./data/agglomerator.db")
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	// Store initial configuration
	moduleConfig, err := json.Marshal(config.Modules.BlockchainAgglomerator)
	if err != nil {
		return fmt.Errorf("failed to marshal module config: %w", err)
	}

	if err := configManager.SetConfig("blockchain_agglomerator", moduleConfig); err != nil {
		return fmt.Errorf("failed to store initial config: %w", err)
	}

	metrics := core.NewMetricsExporter()
	logger := &core.ModuleLogger{
		Outputs: make(map[string]*os.File),
	}

	// Create and initialize module
	module := agglomerator.NewAgglomeratorModule(
		configManager,
		metrics,
		logger,
	)

	if err := module.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize module: %w", err)
	}

	// Create API router
	apiHandler := agglomerator.NewAPI(module)
	router := chi.NewRouter()
	router.Mount("/api/agglomerator", apiHandler.Routes())

	fmt.Println("Starting agglomerator service on :8088")
	return http.ListenAndServe(":8088", router)
}

func addChain(chainID, endpoint, protocol string) error {
	chain := &agglomerator.Chain{
		ID:       chainID,
		Endpoint: endpoint,
		Protocol: protocol,
		StateVector: vectors.InfiniteVector{
			Generator: func(dim int) float64 {
				return math.Exp(-float64(dim)/10.0) * math.Sin(float64(dim))
			},
		},
	}

	// Make API request to register chain
	url := "http://localhost:8088/api/agglomerator/chains"
	body, err := json.Marshal(chain)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to register chain: %s", resp.Status)
	}

	fmt.Printf("Successfully registered chain %s\n", chainID)
	return nil
}

func listChains() error {
	// Make API request to list chains
	resp, err := http.Get("http://localhost:8088/api/agglomerator/chains")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var chains []*agglomerator.Chain
	if err := json.NewDecoder(resp.Body).Decode(&chains); err != nil {
		return err
	}

	// Print chains in a formatted table
	fmt.Printf("%-20s %-40s %-10s\n", "CHAIN ID", "ENDPOINT", "PROTOCOL")
	fmt.Println(strings.Repeat("-", 70))
	for _, chain := range chains {
		fmt.Printf("%-20s %-40s %-10s\n", chain.ID, chain.Endpoint, chain.Protocol)
	}

	return nil
}

func createTransaction(fromChain, toChain string, data []byte) error {
	tx := &agglomerator.Transaction{
		ID:        uuid.NewString(),
		FromChain: fromChain,
		ToChain:   toChain,
		Data:      data,
		StateVector: vectors.InfiniteVector{
			Generator: func(dim int) float64 {
				return math.Exp(-float64(dim)/10.0) * math.Sin(float64(dim))
			},
		},
		Similarity: 0.7,
	}

	// Make API request to create transaction
	url := "http://localhost:8088/api/agglomerator/transaction"
	body, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to create transaction: %s", resp.Status)
	}

	fmt.Printf("Successfully created transaction %s\n", tx.ID)
	return nil
}
