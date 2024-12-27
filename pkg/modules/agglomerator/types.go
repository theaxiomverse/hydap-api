package agglomerator

import (
	"context"
	"errors"
	"github.com/theaxiomverse/hydap-api/pkg/vectors"
	"sync"
)

var (
	ErrNoRouteFound  = errors.New("no route found between chains")
	ErrChainNotFound = errors.New("chain not found")
)

// Agglomerator manages the cross-chain operations
type Agglomerator struct {
	chains      map[string]*Chain
	vectorIndex *vectors.InfiniteVectorIndex
	mu          sync.RWMutex
}

// AgglomeratorConfig holds initialization parameters
type AgglomeratorConfig struct {
	NodeID       string
	VectorDims   int
	SimThreshold float64
}

// Chain represents a blockchain network with vector state
// In pkg/modules/agglomerator/types.go
type Chain struct {
	ID                  string
	Endpoint            string
	Protocol            string
	StateVector         vectors.InfiniteVector
	TransactionPool     *vectors.InfiniteVectorIndex
	streamingCompressor *AdaptiveCompressor // Add this field
	compressedBlocks    []*CompressedBlock  // Add this field
}

// Transaction represents a cross-chain transaction
type Transaction struct {
	ID          string
	FromChain   string
	ToChain     string
	Data        []byte
	StateVector vectors.InfiniteVector
	Similarity  float64
}

// NewAgglomerator creates a new instance
func NewAgglomerator(config AgglomeratorConfig) *Agglomerator {
	return &Agglomerator{
		chains:      make(map[string]*Chain),
		vectorIndex: vectors.NewInfiniteVectorIndex(),
	}
}

// RegisterChain adds a new chain to the agglomerator
func (a *Agglomerator) RegisterChain(chain *Chain) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Initialize transaction pool with vector index
	chain.TransactionPool = vectors.NewInfiniteVectorIndex()

	// Store chain in local registry
	a.chains[chain.ID] = chain

	// Register chain's state vector
	record := vectors.DatabaseRecord{
		ID: chain.ID,
		Metadata: map[string]interface{}{
			"protocol": chain.Protocol,
			"endpoint": chain.Endpoint,
		},
		Vector: chain.StateVector,
	}

	return a.vectorIndex.Insert(record)
}

// ProcessTransaction handles a cross-chain transaction
func (a *Agglomerator) ProcessTransaction(ctx context.Context, tx *Transaction) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Query similar chains based on state vectors
	similarChains := a.vectorIndex.AdvancedQuery(
		tx.Similarity,
		tx.StateVector,
		50, // default dimensions to compare
	)

	if len(similarChains) == 0 {
		return ErrNoRouteFound
	}

	// Record transaction vector
	record := vectors.DatabaseRecord{
		ID: tx.ID,
		Metadata: map[string]interface{}{
			"fromChain": tx.FromChain,
			"toChain":   tx.ToChain,
		},
		Vector: tx.StateVector,
	}

	if err := a.vectorIndex.Insert(record); err != nil {
		return err
	}

	// Add to chains' transaction pools
	fromChain, exists := a.chains[tx.FromChain]
	if !exists {
		return ErrChainNotFound
	}

	toChain, exists := a.chains[tx.ToChain]
	if !exists {
		return ErrChainNotFound
	}

	// Add to transaction pools
	fromChain.TransactionPool.Insert(record)
	toChain.TransactionPool.Insert(record)

	return nil
}

func (a *Agglomerator) ListChains() []*Chain {
	a.mu.RLock()
	defer a.mu.RUnlock()

	chains := make([]*Chain, 0, len(a.chains))
	for _, chain := range a.chains {
		chains = append(chains, chain)
	}
	return chains
}

func (a *Agglomerator) GetChain(id string) (*Chain, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	chain, exists := a.chains[id]
	if !exists {
		return nil, ErrChainNotFound
	}
	return chain, nil
}
