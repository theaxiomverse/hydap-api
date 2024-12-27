package agglomerator

import (
	"math"
)

const (
	ProtocolBitcoin  = "btc"
	ProtocolEthereum = "eth"
	ProtocolSolana   = "sol"
	ProtocolPolkadot = "dot"
)

// ChainProtocol represents various blockchain protocol configurations
type ChainProtocol struct {
	ID               string
	BlockTime        float64 // Average block time in seconds
	ConfirmationTime float64 // Average confirmation time in seconds
	TPS              float64 // Transactions per second
	Finality         float64 // Time to finality in seconds
	CostWeight       float64 // Relative transaction cost weight
}

var protocolConfigs = map[string]ChainProtocol{
	ProtocolBitcoin: {
		ID:               ProtocolBitcoin,
		BlockTime:        600,  // 10 minutes
		ConfirmationTime: 3600, // 1 hour (6 confirmations)
		TPS:              7,    // Bitcoin base layer TPS
		Finality:         3600, // 1 hour
		CostWeight:       1.0,  // Base reference
	},
	ProtocolEthereum: {
		ID:               ProtocolEthereum,
		BlockTime:        12,  // ~12 seconds
		ConfirmationTime: 180, // ~3 minutes
		TPS:              15,  // Ethereum base layer TPS
		Finality:         180, // ~3 minutes
		CostWeight:       0.8,
	},
	ProtocolSolana: {
		ID:               ProtocolSolana,
		BlockTime:        0.4,   // 400ms
		ConfirmationTime: 2,     // ~2 seconds
		TPS:              65000, // Theoretical max TPS
		Finality:         2,     // ~2 seconds
		CostWeight:       0.1,
	},
	ProtocolPolkadot: {
		ID:               ProtocolPolkadot,
		BlockTime:        6,    // ~6 seconds
		ConfirmationTime: 30,   // ~30 seconds
		TPS:              1000, // Average TPS
		Finality:         30,   // ~30 seconds
		CostWeight:       0.5,
	},
}

// getProtocolConfig returns the configuration for a given protocol
func getProtocolConfig(protocol string) (ChainProtocol, bool) {
	config, exists := protocolConfigs[protocol]
	return config, exists
}

// determineProtocol gets the protocol identifier for a chain
func determineProtocol(chainID string) string {
	protocols := map[string]string{
		"bitcoin":  ProtocolBitcoin,
		"btc":      ProtocolBitcoin,
		"ethereum": ProtocolEthereum,
		"eth":      ProtocolEthereum,
		"solana":   ProtocolSolana,
		"sol":      ProtocolSolana,
		"polkadot": ProtocolPolkadot,
		"dot":      ProtocolPolkadot,
	}
	if proto, exists := protocols[chainID]; exists {
		return proto
	}
	return "unknown"
}

// getDefaultGenerator returns a protocol-specific vector generator
func getDefaultGenerator(chainID string) func(int) float64 {
	protocol := determineProtocol(chainID)
	config, exists := getProtocolConfig(protocol)
	if !exists {
		// Return default generator if protocol not found
		return func(dim int) float64 {
			return math.Exp(-float64(dim)/10.0) * math.Sin(float64(dim))
		}
	}

	// Create protocol-specific generator based on characteristics
	return func(dim int) float64 {
		// Base oscillation modified by protocol characteristics
		base := math.Exp(-float64(dim)/10.0) * math.Sin(float64(dim))

		// Modify based on protocol characteristics
		speedFactor := math.Log(1+config.TPS) / math.Log(1+65000) // Normalize TPS
		finalityFactor := 1 - (config.Finality / 3600)            // Normalize finality time
		costFactor := 1 - config.CostWeight                       // Inverse of cost weight

		// Combine factors to modify the base oscillation
		return base * (speedFactor + finalityFactor + costFactor) / 3
	}
}
